package messaging

import (
	"context"
	"domino/shared/contracts"
	"domino/shared/retry"
	"domino/shared/tracing"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	TripExchange       = "trip"
	DeadLetterExchange = "dlx"
)

type RabbitMQ struct {
	conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewRabbitMQ(uri string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create to channel: %v", err)
	}

	rmq := &RabbitMQ{
		conn:    conn,
		Channel: ch,
	}

	if err := rmq.setupExchangesAndQueues(); err != nil {
		rmq.Close()
		return nil, fmt.Errorf("failed to setup exchanges and queues: %v", err)
	}

	return rmq, nil
}

// Messages handle is the function signature every service consumer must implement
type MessageHandler func(context.Context, amqp.Delivery) error

// ConsumeMessages starts a consumer on queueName
// Each message is process by handler with expontential backoff retry
// On exhausted retries, the message is rejedted to the dead letter exchange
func (r *RabbitMQ) ConsumeMessages(queueName string, handler MessageHandler) error {
	// Prefetch 1: RabbitMQ delivers the next message only after the current one is acknowledge
	// This gives faire dispatch across multiple service instances
	if err := r.Channel.Qos(1, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS: %v", err)
	}

	msgs, err := r.Channel.Consume(queueName,
		"", // consumer tag: auto generated
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			// NOTE: Here our TracedConsumer passes a delivery that is not processed
			if err := tracing.TracedConsumer(msg, func(ctx context.Context, d amqp.Delivery) error {
				log.Printf("Received messgae on %s: %s", queueName, msg.Body)

				cfg := retry.DefualtConfig()
				err := retry.WithBackoff(ctx, cfg, func() error {
					return handler(ctx, d)
				})

				// Handler never succeeded after retries -> sent to DeadLetterQueue
				if err != nil {
					log.Printf("Message processing failed after %d retries: %v", cfg.MaxRetries, err)

					// Enrich headers with failure context before sending to DeadLetterQueue
					headers := amqp.Table{}
					if d.Headers != nil {
						headers = d.Headers
					}
					headers["x-death-reason"] = err.Error()
					headers["x-origin-exchange"] = d.Exchange
					headers["x-original-routing-key"] = d.RoutingKey
					headers["x-retry-count"] = cfg.MaxRetries
					d.Headers = headers

					// requee = false -> message goes to x-dead-letter-exchange (dlx)
					_ = d.Reject(false)
					return err
				}

				// handler executed correctly -> Acknowledge
				if ackErr := d.Ack(false); ackErr != nil {
					log.Printf("Failed to ack message: %v", ackErr)
				}
				return nil
			}); err != nil {
				log.Printf("Error in traced consumer: %v", err)
			}
		}
	}()

	return nil
}

// PublishMessage marshals message as JSON and publishes it to the trip exchange with the routing key.
// The msg is persistent
func (r *RabbitMQ) PublishMessage(ctx context.Context, routingKey string, message contracts.AmqpMessage) error {
	log.Printf("Publishing message with routing key: %s", routingKey)
	jsonMsg, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}
	msg := amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent, // survives broker strart
		Body:         jsonMsg,
	}
	return tracing.TracedPublisher(ctx, TripExchange, routingKey, msg, r.publish)
}

func (r *RabbitMQ) publish(ctx context.Context, exchange, routingKey string, msg amqp.Publishing) error {
	return r.Channel.PublishWithContext(ctx,
		exchange,
		routingKey,
		false, //mandatory: dont return the message if no queue matches
		false, //immediate: dont require an active consumer
		msg,
	)

}

// / privates:
func (r *RabbitMQ) setupExchangesAndQueues() any {
	// DeadLetter Exchange
	if err := r.setupDeadLetterExchange(); err != nil {
		return err
	}

	// Trip Exchange declaration
	if err := r.Channel.ExchangeDeclare(TripExchange, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("failed to declare exchange %s: %v", TripExchange, err)
	}

	// Each queue binds to one or more routing keys on the TripExchange
	type queueDef struct {
		name        string
		routingKeys []string
	}
	queues := []queueDef{}
	for _, q := range queues {
		if err := r.declareAndBindQueue(q.name, q.routingKeys, TripExchange); err != nil {
			return err
		}
	}
	return nil
}

func (r *RabbitMQ) setupDeadLetterExchange() any {
	if err := r.Channel.ExchangeDeclare(DeadLetterExchange, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("failed to declare exchange %s: %v", DeadLetterExchange, err)
	}
	q, err := r.Channel.QueueDeclare(DeaedLetterQueue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %v", DeaedLetterQueue, err)
	}
	// "#" matches all routing keys so every rejected message lands here
	if err := r.Channel.QueueBind(q.Name, "#", DeadLetterExchange, false, nil); err != nil {
		return fmt.Errorf("failed to bind DLQ: %v", err)
	}
	return nil
}

// Declare and bind Queues with dead letter exchange setted
func (r *RabbitMQ) declareAndBindQueue(name string, routineKeys []string, exchange string) error {
	// x-dead-letter-exchange routes rejected messgaes to the DLX
	args := amqp.Table{
		"x-dead-letter-exchange": DeadLetterExchange,
	}

	q, err := r.Channel.QueueDeclare(name, true, false, false, false, args)
	if err != nil {
		return fmt.Errorf("failed to declare queue %s: %v", name, err)
	}

	for _, key := range routineKeys {
		if err := r.Channel.QueueBind(q.Name, key, exchange, false, nil); err != nil {
			return fmt.Errorf("failed to bind queue %s to key %s: %v", name, key, err)
		}
	}
	return nil
}

func (r *RabbitMQ) Close() {
	if r.conn != nil {
		r.conn.Close()
	}
	if r.Channel != nil {
		r.Channel.Close()
	}
}
