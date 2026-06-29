package events

import (
	"context"
	"encoding/json"
	"log"
	"rebu/services/payment-service/internal/domain"
	"rebu/shared/contracts"
	"rebu/shared/messaging"

	amqp "github.com/rabbitmq/amqp091-go"
)

type TripConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  domain.Service
}

func NewTripConsumer(rabbitmq *messaging.RabbitMQ, service domain.Service) *TripConsumer {
	return &TripConsumer{
		rabbitmq: rabbitmq,
		service:  service,
	}
}

func (c *TripConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessages(messaging.PaymentTripResponseQueue, c.handle)
}

func (c *TripConsumer) handle(ctx context.Context, msg amqp.Delivery) error {
	var envelope contracts.AmqpMessage
	if err := json.Unmarshal(msg.Body, &envelope); err != nil {
		return err
	}

	var payload messaging.PaymentTripResponseData
	if err := json.Unmarshal(envelope.Data, &payload); err != nil {
		return err
	}

	switch msg.RoutingKey {
	case contracts.PaymentCmdCreateSession:
		return c.handleCreateSession(ctx, &payload)
	default:
		log.Printf("payment-service: unknown routing key %s", msg.RoutingKey)
		return nil
	}
}

func (c *TripConsumer) handleCreateSession(ctx context.Context, payload *messaging.PaymentTripResponseData) error {
	log.Printf("Creating payment session for trip %s", payload.TripID)

	intent, err := c.service.CreatePaymentSession(ctx, payload.TripID, payload.UserID, payload.DriverID, int64(payload.Amount), payload.Currency)
	if err != nil {
		return err
	}

	// Publish the session ID so the API gateway can push it to the ride's Websocket
	data, err := json.Marshal(messaging.PaymentEventSessionCreatedData{
		TripID:    payload.TripID,
		SessionID: intent.StripeSessionID,
		Amount:    float64(intent.Amount) / 100.0, // cents -> dollars for display
		Currency:  intent.Currency,
	})
	if err != nil {
		return err
	}
	return c.rabbitmq.PublishMessage(ctx, contracts.PaymentEventSessionCreated, contracts.AmqpMessage{
		OwnerID: payload.UserID, // routes to the rider's Websocket in the gateway
		Data:    data,
	})
}
