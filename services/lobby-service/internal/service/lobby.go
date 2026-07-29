package service

import (
	"context"
	"domino/services/lobby-service/internal/domain"
	"fmt"

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
		Slot:        len(lobby.Players) + 1,
		IsConnected: false,
	}
	lobby.Players = append(lobby.Players, player)
	return s.repo.UpdateLobby(ctx, lobbyID, lobby)
}

func (s *service) StartLobby(ctx context.Context, lobbyID string, userID string) (*domain.LobbyModel, error) {
	lobby, err := s.repo.GetLobbyByID(ctx, lobbyID)
	if err != nil {
		return nil, fmt.Errorf("couldn't get lobby: %w", err)
	}
	if lobby.HostID != userID {
		return nil, fmt.Errorf("start request needs to be from host")
	}

	if lobby.MaxPlayers != len(lobby.Players) {
		return nil, fmt.Errorf("needs full players")
	}

	// state machine
	switch lobby.Status {
	case domain.LobbyStatusWaiting:
		lobby.Status = domain.LobbyStatusInGame
		lobby, err = s.repo.UpdateLobby(ctx, lobbyID, lobby)
		if err != nil {
			return nil, fmt.Errorf("couldn't update lobby state")
		}
	default:
		return nil, fmt.Errorf("wrong lobby state transition")
	}
	return lobby, nil
}
