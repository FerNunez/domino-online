package service

import (
	"context"
	"fmt"
	"rebu/services/lobby-service/internal/domain"

	"github.com/google/uuid"
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
		ID:     uuid.NewString(),
		HostID: hostID,
		Status: domain.LobbyStatusWaiting,
		Players: []*domain.PlayerModel{{
			ID:          hostID,
			Name:        hostID, // FIX:
			Slot:        1,
			IsConnected: false,
		}},
		MaxPlayers: maxPlayers,
		Settings: domain.LobbySettings{
			MaxScore:         100,
			TurnTimerSeconds: 30,
		},
	}
	return s.repo.CreateLobby(ctx, &lobby)
}

func (s *service) JoinLobby(ctx context.Context, lobbyID string, userID string) (*domain.LobbyModel, error) {
	lobby, err := s.repo.GetLobbyByID(ctx, lobbyID)
	if err != nil {
		return nil, fmt.Errorf("couldn't get lobby: %w", err)
	}

	if len(lobby.Players) >= lobby.MaxPlayers {
		return nil, fmt.Errorf("full lobby: %w", err)
	}

	player := &domain.PlayerModel{
		ID:          userID,
		Name:        userID, // FIX:
		Slot:        lobby.MaxPlayers + 1,
		IsConnected: false,
	}
	lobby.Players = append(lobby.Players, player)
	return s.repo.UpdateLobby(ctx, lobbyID, lobby)
}
