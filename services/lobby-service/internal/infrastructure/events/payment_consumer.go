package events

import (
	"context"
	"encoding/json"
	"rebu/services/trip-service/internal/domain"
	"rebu/shared/contracts"
	"rebu/shared/messaging"

	amqp "github.com/rabbitmq/amqp091-go"
)

type paymentConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  domain.TripService
}

func NewPaymentConsumer(rmq *messaging.RabbitMQ, s domain.TripService) *paymentConsumer {
	return &paymentConsumer{
		rabbitmq: rmq,
		service:  s,
	}
}

// Listen consumes messages NotifyPaymentSuccessQueue and update trip state to payed
func (c *paymentConsumer) Listen() error {
	// paymentSucessHandle is the handle func
	return c.rabbitmq.ConsumeMessages(messaging.NotifyPaymentSuccessQueue, c.paymentSucessHandle)
}

// paymentSucessHandle updates the state to paid fore the trip
func (c *paymentConsumer) paymentSucessHandle(ctx context.Context, msg amqp.Delivery) error {
	var envelope contracts.AmqpMessage
	if err := json.Unmarshal(msg.Body, &envelope); err != nil {
		return err
	}
	var payload messaging.PaymentStatusUpdateData
	if err := json.Unmarshal(envelope.Data, &payload); err != nil {
		return err
	}
	return c.service.UpdateTrip(ctx, payload.TripID, "paid", nil)
}
