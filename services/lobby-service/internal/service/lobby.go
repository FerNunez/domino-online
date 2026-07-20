package service

import (
	"context"
	"fmt"
	"rebu/services/lobby-service/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type service struct {
	repo domain.LobbyRepository
}

// NewService creates the service layer wired to the given repository
func NewService(repo domain.LobbyRepository) *service {
	return &service{
		repo: repo,
	}
}

func (s *service) CreateLobby(ctx context.Context, hostID string, maxPlayers int) (*domain.LobbyModel, error) {
	lobby := domain.LobbyModel{
		ID:         primitive.NewObjectID(),
		HostID:     hostID,
		Status:     domain.LobbyStatusWaiting,
		Players:    []*domain.PlayerModel{},
		MaxPlayers: maxPlayers,
		Settings: domain.LobbySettings{
			MaxScore:         100,
			TurnTimerSeconds: 30,
		},
	}
	return s.repo.CreateLobby(ctx, &lobby)
}

func (s *service) JoinLobby(ctx context.Context, id string, secretToken string, player *domain.PlayerModel) (*domain.LobbyModel, error) {
	lobby, err := s.repo.GetLobbyByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting lobby: %w", err)
	}
	if len(lobby.Players) >= lobby.MaxPlayers {
		return nil, fmt.Errorf("full lobby: %w", err)
	}
	lobby.Players = append(lobby.Players, player)

	return s.repo.UpdateLobby(ctx, id, lobby)
}
