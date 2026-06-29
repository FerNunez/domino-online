package events

import (
	"context"
	"encoding/json"
	"rebu/services/trip-service/internal/domain"
	"rebu/shared/contracts"
	"rebu/shared/messaging"
)

type TripEventPublisher struct {
	rabbitmq *messaging.RabbitMQ
}

func NewTripEventPublisher(rmq *messaging.RabbitMQ) *TripEventPublisher {
	return &TripEventPublisher{
		rabbitmq: rmq,
	}
}

// PublishTripCreated publish a rabbitmq message with routing 	TripEventCreated = "trip.event.created"
// OwnerID is set to the rider's userID so the API gatway can route any doewnstream notification to the correct WebSocket Connection
func (p *TripEventPublisher) PublishTripCreated(ctx context.Context, trip *domain.TripModel) error {
	payload := messaging.TripEventData{Trip: trip.ToProto()}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.rabbitmq.PublishMessage(ctx, contracts.TripEventCreated, contracts.AmqpMessage{
		OwnerID: trip.UserID,
		Data:    data,
	})
}
