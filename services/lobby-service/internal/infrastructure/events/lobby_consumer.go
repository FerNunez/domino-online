package events

import (
	"context"
	"domino/services/lobby-service/internal/domain"
	"domino/shared/contracts"
	"domino/shared/messaging"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type lobbyConsumer struct {
	rabbitmq *messaging.RabbitMQ
	service  domain.LobbyService
}

func NewLobbyConsumer(rabbitmq *messaging.RabbitMQ, service domain.LobbyService) (*lobbyConsumer, error) {
	_, err := rabbitmq.DeclareQueueAndBind(messaging.LobbyQueue, []string{contracts.PlayerConnected, contracts.PlayerDisconnected}, messaging.DominoExchange)
	if err != nil {
		return nil, fmt.Errorf("couldn't bind to lobby queue: %w", err)
	}

	return &lobbyConsumer{
		rabbitmq: rabbitmq,
		service:  service,
	}, nil
}

// Listen consumes lobby broadcast events (GameStarted) and player commands
// (play tile / pass) directed at this service.
func (c *lobbyConsumer) Listen() error {
	if err := c.rabbitmq.ConsumeMessages(messaging.LobbyQueue, c.handleLobbyMessages); err != nil {
		return err
	}
	// TODO: To consume game? or rather to publish into game?
	//return c.rabbitmq.ConsumeMessages(messaging.NotifyGame, c.handleGameCmd)
	return nil
}

func (c *lobbyConsumer) handleLobbyMessages(ctx context.Context, msg amqp.Delivery) error {
	var envelope contracts.DominoEvent
	if err := json.Unmarshal(msg.Body, &envelope); err != nil {
		return err
	}

	switch msg.RoutingKey {
	case contracts.PlayerConnected:
		var payload messaging.PlayerConnectedData
		if err := json.Unmarshal(envelope.Data, &payload); err != nil {
			return err
		}
		return c.service.SetPlayerConnected(ctx, payload.LobbyID, payload.UserID)
	case contracts.PlayerDisconnected:
		var payload messaging.PlayerDisconnectedData
		if err := json.Unmarshal(envelope.Data, &payload); err != nil {
			return err
		}
		return c.service.SetPlayerDisconnected(ctx, payload.LobbyID, payload.UserID)
	default:
		log.Printf("game-service: unknown routing key %s", msg.RoutingKey)
		return nil
	}
}
