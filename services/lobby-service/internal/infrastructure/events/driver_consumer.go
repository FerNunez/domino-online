package events

import (
	"context"
	"encoding/json"
	"fmt"

	"rebu/services/trip-service/internal/domain"
	"rebu/shared/contracts"
	"rebu/shared/messaging"
	pbd "rebu/shared/proto/driver"

	amqp "github.com/rabbitmq/amqp091-go"
)

type driverConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  domain.TripService
}

func NewDriverConsumer(r *messaging.RabbitMQ, s domain.TripService) *driverConsumer {
	return &driverConsumer{
		rabbitmq: r,
		service:  s,
	}
}

// Listen consumes messages from the DriverTripResponseQueue and applies handle function
func (c *driverConsumer) Listen() error {
	return c.rabbitmq.ConsumeMessages(messaging.DriverTripResponseQueue, c.handleFunc)
}

// handleFunc decodes the msg and handles succeeded or declined driver command
func (c *driverConsumer) handleFunc(ctx context.Context, msg amqp.Delivery) error {
	var envelope contracts.AmqpMessage
	if err := json.Unmarshal(msg.Body, &envelope); err != nil {
		return err
	}

	var payload messaging.DriverTripResponseData
	if err := json.Unmarshal(envelope.Data, &payload); err != nil {
		return err
	}

	switch msg.RoutingKey {
	case contracts.DriverCmdTripAccept:
		// event TripEventDriverAssigned
		// command payment
		return c.handleAccepted(ctx, payload.TripID, payload.Driver)
	case contracts.DriverCmdTripDecline:
		return c.handleDeclined(ctx, payload.TripID)
	}

	return nil
}

// handleAccepted gets trip and updates trip to status=accepted and attach driver
// publishes event TripEventDriverAssigned with trip data
// publishes command with event 'PaymentCmdCreateSession" with payment data from tripID, riderID and driver_accepted data
func (c *driverConsumer) handleAccepted(ctx context.Context, tripID string, driver *pbd.Driver) error {
	trip, err := c.service.GetTripByID(ctx, tripID)
	if err != nil {
		return err
	}
	if trip == nil {
		return fmt.Errorf("trip not found: %s", tripID)
	}

	// 1. Update Trip status and attache Driver information
	if err := c.service.UpdateTrip(ctx, tripID, "accepted", driver); err != nil {
		return err
	}

	// Re-fetch tho get the updated trip with driver info
	trip, err = c.service.GetTripByID(ctx, tripID)
	if err != nil {
		return err
	}
	if trip == nil {
		return fmt.Errorf("trip not found: %s", tripID)
	}

	// 2. Notify the rider that a driver has been assigned
	tripData, err := json.Marshal(trip)
	if err != nil {
		return err
	}
	// TripEventDriverAssigned w/ trip data)
	if err := c.rabbitmq.PublishMessage(ctx, contracts.TripEventDriverAssigned, contracts.AmqpMessage{
		OwnerID: trip.UserID,
		Data:    tripData,
	}); err != nil {
		return err
	}
	// NOTE: Why this paymentData is called PaymentTripResponseData
	// 3. Instruct the pament service to create a Stripe session and publish event
	paymentData, err := json.Marshal(messaging.PaymentTripResponseData{
		TripID:   tripID,
		UserID:   trip.UserID,
		DriverID: driver.Id,
		Amount:   trip.RideFare.TotalPriceInCents,
		Currency: "USD",
	})
	if err := c.rabbitmq.PublishMessage(ctx, contracts.PaymentCmdCreateSession, contracts.AmqpMessage{
		OwnerID: trip.UserID,
		Data:    paymentData,
	}); err != nil {
		return err
	}
	return nil
}

// handleDeclined publishes Event TripEventDriversNotInterested with trip information
func (c *driverConsumer) handleDeclined(ctx context.Context, tripID string) error {
	trip, err := c.service.GetTripByID(ctx, tripID)
	if err != nil {
		return err
	}
	if trip == nil {
		return fmt.Errorf("trip not found: %s", tripID)
	}
	data, err := json.Marshal(trip)
	if err != nil {
		return err
	}
	c.rabbitmq.PublishMessage(ctx, contracts.TripEventDriversNotInterested, contracts.AmqpMessage{
		OwnerID: trip.UserID,
		Data:    data,
	})
	return nil
}
