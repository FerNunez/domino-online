package service

import (
	"context"
	"fmt"

	"domino/services/lobby-service/internal/domain"

	"github.com/google/uuid"
)

type service struct {
	repo      domain.LobbyRepository
	publisher domain.LobbyEventPublisher
}

// NewService creates the service layer wired to the given repository
func NewService(repo domain.LobbyRepository, publisher domain.LobbyEventPublisher) *service {
	return &service{
		repo:      repo,
		publisher: publisher,
	}
}

// defaultPlayerName is used when the caller doesn't supply a display name.
func defaultPlayerName(userID string) string {
	if len(userID) < 2 {
		return "Guest-" + userID
	}
	return "Guest-" + userID[:2]
}

func (s *service) CreateLobby(ctx context.Context, hostID, hostName string, maxPlayers int) (*domain.LobbyModel, error) {
	if hostName == "" {
		hostName = defaultPlayerName(hostID)
	}

	host := &domain.PlayerModel{
		ID:          hostID,
		Name:        hostName,
		Slot:        1,
		IsConnected: false,
	}
	lobby := domain.LobbyModel{
		ID:         uuid.NewString(),
		HostID:     hostID,
		Status:     domain.LobbyStatusWaiting,
		Players:    map[string]*domain.PlayerModel{hostID: host},
		MaxPlayers: maxPlayers,
		Settings: domain.LobbySettings{
			MaxScore:         100,
			TurnTimerSeconds: 30,
		},
	}
	return s.repo.CreateLobby(ctx, &lobby)
}

func (s *service) JoinLobby(ctx context.Context, lobbyID string, userID string, displayName string) (*domain.LobbyModel, error) {
	lobby, err := s.repo.GetLobbyByID(ctx, lobbyID)
	if err != nil {
		return nil, fmt.Errorf("couldn't get lobby: %w", err)
	}

	if _, ok := lobby.Players[userID]; ok {
		return nil, domain.ErrAlreadyMember
	}

	if len(lobby.Players) >= lobby.MaxPlayers {
		return nil, fmt.Errorf("full lobby")
	}

	if displayName == "" {
		displayName = defaultPlayerName(userID)
	}

	// Add player to lobby
	player := &domain.PlayerModel{
		ID:          userID,
		Name:        displayName,
		Slot:        len(lobby.Players) + 1,
		IsConnected: false,
	}
	if err := s.repo.AddPlayer(ctx, lobbyID, player); err != nil {
		return nil, fmt.Errorf("couldn't add player: %w", err)
	}
	lobby.Players[userID] = player

	// Publish that a player joined lobby
	if err := s.publisher.PublishPlayerJoined(ctx, lobby, player); err != nil {
		return nil, fmt.Errorf("couldn't notify player joined: %w", err)
	}

	return lobby, nil
}

func (s *service) ReconnectLobby(ctx context.Context, lobbyID string, userID string) (*domain.LobbyModel, error) {
	lobby, err := s.repo.GetLobbyByID(ctx, lobbyID)
	if err != nil {
		return nil, fmt.Errorf("couldn't get lobby: %w", err)
	}

	if _, ok := lobby.Players[userID]; !ok {
		return nil, domain.ErrNotMember
	}

	return lobby, nil
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
		if err := s.repo.SetStatus(ctx, lobbyID, domain.LobbyStatusInGame); err != nil {
			return nil, fmt.Errorf("couldn't update lobby state: %w", err)
		}
		lobby.Status = domain.LobbyStatusInGame
	default:
		return nil, fmt.Errorf("wrong lobby state transition")
	}
	return lobby, nil
}

func (s *service) GetLobby(ctx context.Context, lobbyID string) (*domain.LobbyModel, error) {
	lobby, err := s.repo.GetLobbyByID(ctx, lobbyID)
	if err != nil {
		return nil, fmt.Errorf("couldn't get lobby: %w", err)
	}
	return lobby, nil
}

func (s *service) SetPlayerConnected(ctx context.Context, lobbyID string, userID string) error {
	if err := s.repo.SetPlayerConnection(ctx, lobbyID, userID, true); err != nil {
		return fmt.Errorf("couldn't set player connected: %w", err)
	}
	return nil
}

func (s *service) SetPlayerDisconnected(ctx context.Context, lobbyID string, userID string) error {
	if err := s.repo.SetPlayerConnection(ctx, lobbyID, userID, false); err != nil {
		return fmt.Errorf("couldn't set player disconnected: %w", err)
	}
	return nil
}
