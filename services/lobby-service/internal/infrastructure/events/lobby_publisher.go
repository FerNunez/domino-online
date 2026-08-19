package events

import (
	"context"
	"domino/services/lobby-service/internal/domain"
	"domino/shared/contracts"
	"domino/shared/messaging"
	"encoding/json"

	"github.com/google/uuid"
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
	sorted := lobby.SortedPlayers()
	playersID := make([]string, len(sorted))
	for idx, p := range sorted {
		playersID[idx] = p.ID
	}
	payload := messaging.GameStartCmd{
		PlayersID: playersID,
		GameID:    uuid.NewString(), // ID of the new game Lobby commands //NOTE: todo if make sense
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.rabbitmq.PublishMessage(ctx, contracts.GameStartCmd, &contracts.DominoEvent{
		LobbyID:  lobby.ID,
		TargetID: "",
		Data:     data,
	})
}

// Publishes contracts.PlayerJoinedLobby so lobby members already connected
// learn about a new player
func (p *LobbyEventPublisher) PublishPlayerJoined(ctx context.Context, lobby *domain.LobbyModel, player *domain.PlayerModel) error {
	payload := messaging.PlayerJoinedData{
		UserID:      player.ID,
		DisplayName: player.Name,
		PlayerCount: len(lobby.Players),
		MaxPlayers:  lobby.MaxPlayers,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.rabbitmq.PublishMessage(ctx, contracts.PlayerJoinedLobby, &contracts.DominoEvent{
		LobbyID:  lobby.ID,
		TargetID: "",
		Data:     data,
	})
}

// TODO: PublishPauseGame?
