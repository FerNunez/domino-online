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

// Publishes contracts.GameStartCmd with starting player ID
func (p *LobbyEventPublisher) PublishStartGame(ctx context.Context, lobby *domain.LobbyModel) error {
	playersID := make([]string, len(lobby.Players))
	for idx, p := range lobby.Players {
		playersID[idx] = p.ID
	}
	payload := messaging.GameStartCmd{
		PlayersID: playersID,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.rabbitmq.PublishMessage(ctx, contracts.GameStartCmd, contracts.DominoEvent{
		LobbyID:  lobby.ID,
		TargetID: "",
		Data:     data,
	})
}

// TODO: PublishPauseGame?
