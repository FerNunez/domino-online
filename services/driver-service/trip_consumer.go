package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand/v2"
	"rebu/shared/contracts"
	"rebu/shared/messaging"

	amqp "github.com/rabbitmq/amqp091-go"
)

type tripConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  *Service
}

func NewTripConsumer(rabbitmq *messaging.RabbitMQ, service *Service) *tripConsumer {
	return &tripConsumer{
		rabbitmq: rabbitmq,
		service:  service,
	}
}

// Listen consumes messages from FindAvailableDriversQueue and applies handle which
// then notifies user NoDriverFound or Notifies driver for a trip request command
func (c *tripConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessages(messaging.FindAvailableDriversQueue, c.handle)
}

func (c *tripConsumer) handle(ctx context.Context, msg amqp.Delivery) error {
	var envelope contracts.AmqpMessage
	if err := json.Unmarshal(msg.Body, &envelope); err != nil {
		return err
	}

	var payload messaging.TripEventData
	if err := json.Unmarshal(envelope.Data, &payload); err != nil {
		return err
	}

	switch msg.RoutingKey {
	case contracts.TripEventCreated, contracts.TripEventDriversNotInterested:
		return c.handleFindAndNotify(ctx, payload)
	default:
		log.Printf("driver-service: unknown routing key %s", msg.RoutingKey)
		return nil
	}
}

// handleFindAndNotify  find available drivers with seletected fare slug
// if no available => publish NoDriverFound with trip userID
// if avaialable => publish DriverCmdTripRequest with tripData
func (c *tripConsumer) handleFindAndNotify(ctx context.Context, payload messaging.TripEventData) error {
	candidates := c.service.FindAvailableDrivers(payload.Trip.SelectedFare.PackageSlug)
	log.Printf("driver-service: found %d candidates for %s", len(candidates), payload.Trip.SelectedFare.PackageSlug)

	if len(candidates) == 0 {
		// No drivers available — notify the rider
		return c.rabbitmq.PublishMessage(ctx, contracts.TripEventNoDriversFound, contracts.AmqpMessage{
			OwnerID: payload.Trip.UserID,
		})
	}

	// Pick one driver at random
	driverID := candidates[rand.IntN(len(candidates))]
	tripData, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	// OwnerID is the driver ID
	return c.rabbitmq.PublishMessage(ctx, contracts.DriverCmdTripRequest, contracts.AmqpMessage{
		OwnerID: driverID, // driver ID
		Data:    tripData,
	})
}
