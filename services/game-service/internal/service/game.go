package service

import (
	"context"
	"domino/services/game-service/internal/domain"
	"domino/shared/types"
	"fmt"
)

type service struct {
	repo domain.GameRepository
}

// NewService creates the service layer wired to the given repository
func NewService(repo domain.GameRepository) *service {
	return &service{
		repo: repo,
	}
}

// Async called
func (s *service) CreateGame(ctx context.Context, lobbyID string, playerIDs []string) (*domain.GameModel, error) {
	game, err := domain.NewGame(lobbyID, playerIDs)
	if err != nil {
		return nil, fmt.Errorf("couldn't create game: %w", err)
	}
	return s.repo.CreateGame(ctx, game)
}

// Sync called
// PlayTile if valid, update game and send result if over
func (s *service) PlayTile(ctx context.Context, lobbyID, userID string, tile types.Tile, side string) (*domain.GameModel, *types.RoundResult, error) {
	game, err := s.repo.GetGameByLobbyID(ctx, lobbyID)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't get game: %w", err)
	}

	if err := game.PlayTile(userID, tile, side); err != nil {
		return nil, nil, err
	}

	game, err = s.repo.UpdateGame(ctx, lobbyID, game)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't update game: %w", err)
	}

	if game.Status == domain.GameStatusRoundOver {
		result := game.ResolveRoundResult()
		return game, &result, nil
	}
	return game, nil, nil
}

// Sync called:
// Check if pass move is valid, then pass to next user and updates game.
// Game can end if 4 players already passed!!
func (s *service) PassTurn(ctx context.Context, lobbyID, userID string) (*domain.GameModel, *types.RoundResult, error) {
	game, err := s.repo.GetGameByLobbyID(ctx, lobbyID)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't get game: %w", err)
	}

	if err := game.PassTurn(userID); err != nil {
		return nil, nil, err
	}

	game, err = s.repo.UpdateGame(ctx, lobbyID, game)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't update game: %w", err)
	}

	// Return result of play when game round over
	if game.Status == domain.GameStatusRoundOver {
		result := game.ResolveRoundResult()
		return game, &result, nil
	}
	return game, nil, nil
}
