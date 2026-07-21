package messaging

import (
	"encoding/json"
	"log"
	"rebu/shared/contracts"
)

// QueueConsumer reads from a rabbitMQ queue and forwards msg to the Websocket client indentified by
// AmqpMessage.OwnerID
type LobbyQueueConsumer struct {
	rb        *RabbitMQ
	connMgr   *ConnectionManager
	queueName string
}

func NewLobbyQueueConsumer(rb *RabbitMQ, connMgr *ConnectionManager, queueName string) *QueueConsumer {
	return &QueueConsumer{
		rb:        rb,
		connMgr:   connMgr,
		queueName: queueName,
	}
}

func (qc *QueueConsumer) Start() error {
	msgs, err := qc.rb.Channel.Consume(qc.queueName,
		"",
		true, // auto-ack: message is acknowledged immediately on delivery
		false, false, false, nil)
	if err != nil {
		return nil
	}

	go func() {
		for msg := range msgs {
			var envelope contracts.AmqpMessage
			if err = json.Unmarshal(msg.Body, &envelope); err != nil {
				log.Println("QueueConsumer: failed to unmarshal envelope: ", err)
				continue
			}

			// Deserialize the Data field so the frontend receives a proper JSON object,
			// Not a base64-enconded string (what a raw []byte would produce)
			var payload any
			if envelope.Data != nil {
				if err := json.Unmarshal(envelope.Data, &payload); err != nil {
					log.Println("QueueConsumer: failed to unmarshal payload", err)
					continue
				}
			}

			wsMsg := contracts.WSMessage{
				Type: msg.RoutingKey,
				Data: payload,
			}
			if err := qc.connMgr.SendMessage(envelope.OwnerID, wsMsg); err != nil {
				log.Printf("QueueConsumer: failed to send to user %s: %v", envelope.OwnerID, err)
			}
		}
	}()
	return nil
}
