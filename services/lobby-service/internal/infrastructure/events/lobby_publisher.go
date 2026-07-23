package events

import (
	"context"
	"domino/services/lobby-service/internal/domain"
	"domino/shared/contracts"
	"domino/shared/messaging"
	"encoding/json"
)

type LobbyEventPublisher struct {
	rabbitmq *messaging.RabbitMQ
}

func NewLobbyEventPublisher(rmq *messaging.RabbitMQ) *LobbyEventPublisher {
	return &LobbyEventPublisher{
		rabbitmq: rmq,
	}
}

// Publishes contracts.GameStarted with starting player ID
func (p *LobbyEventPublisher) PublishStartGame(ctx context.Context, lobby *domain.LobbyModel) error {
	payload := messaging.GameStartedData{
		StartingPlayerID: lobby.HostID,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.rabbitmq.PublishMessage(ctx, contracts.GameStarted, contracts.LobbyEvent{
		Type:     "Broadcast", // TODO: is this better to convert to EventType
		LobbyID:  lobby.ID,
		TargetID: "",
		Data:     data,
	})
}

// TODO: PublishPauseGame?
