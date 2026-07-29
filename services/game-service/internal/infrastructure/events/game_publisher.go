package events

import (
	"context"
	"domino/services/game-service/internal/domain"
	"domino/shared/contracts"
	"domino/shared/messaging"
	"encoding/json"
)

type GameEventPublisher struct {
	rabbitmq *messaging.RabbitMQ
}

func NewGameEventPublisher(rmq *messaging.RabbitMQ) *GameEventPublisher {
	return &GameEventPublisher{
		rabbitmq: rmq,
	}
}

func (p *GameEventPublisher) PublishMoveMade(ctx context.Context, lobbyID string, userID string, tile domain.Tile, side string) error {
	// create event
	payload := messaging.MoveChangedData{
		UserID: userID,
		Tile:   tile.String(),
		Side:   side,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Publish to MoveMade
	return p.rabbitmq.PublishMessage(ctx, contracts.PlayerMoveMade, contracts.DominoEvent{
		LobbyID:  lobbyID,
		TargetID: "",
		Data:     data,
	})
}

func (p *GameEventPublisher) PublishTurnChanged(ctx context.Context, lobbyID, userID string) error {
	// create event
	payload := messaging.TurnChangedData{UserID: userID}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Publish to MoveMade
	return p.rabbitmq.PublishMessage(ctx, contracts.PlayerPassed, contracts.DominoEvent{
		LobbyID:  lobbyID,
		TargetID: "",
		Data:     data,
	})
}
