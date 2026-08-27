// Package service containt the game service
package service

import (
	"context"
	"errors"
	"fmt"

	"domino/services/game-service/internal/domain"
	"domino/shared/types"
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
func (s *service) CreateGameWithID(ctx context.Context, gameID, lobbyID string, playerIDs []string) (*domain.GameModel, error) {
	// game next number obtained from lobby ID
	gameNumber, err := s.repo.NextGameNumber(ctx, lobbyID)
	if err != nil {
		return nil, fmt.Errorf("couldn't reserve game number: %w", err)
	}

	game, err := domain.NewGame(lobbyID, gameID, gameNumber, playerIDs)
	if err != nil {
		return nil, fmt.Errorf("couldn't create game: %w", err)
	}
	game, err = s.repo.CreateGame(ctx, game)
	if err != nil {
		return nil, fmt.Errorf("couldn't persist game: %w", err)
	}
	return game, nil
}

// Sync called
// PlayTile if valid, update game and send result if over
func (s *service) PlayTile(ctx context.Context, lobbyID, userID string, tile types.Tile, side types.Side) (*domain.GameModel, *types.RoundResult, error) {
	var result *types.RoundResult

	game, err := s.repo.UpdateCurrentGame(ctx, lobbyID, func(g *domain.GameModel) error {
		// Reset in case a previous attempt got this far before losing the
		// CAS race — mutate can run more than once per call.
		result = nil

		round := g.CurrentRound
		if round == nil {
			return fmt.Errorf("couldn't find current round for game")
		}

		if err := round.PlayTile(userID, tile, side); err != nil {
			return err
		}

		if round.Status == domain.RoundStatusRoundOver {
			r := round.ResolveResult()
			if err := g.UpdateScore(&r); err != nil {
				return err
			}
			result = &r
		}
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't update game: %w", err)
	}
	return game, result, nil
}

// Sync called:
// Check if pass move is valid, then pass to next user and updates game.
// Round can end if all 4 players pass in a row (blocked board).
func (s *service) PassTurn(ctx context.Context, lobbyID, userID string) (*domain.GameModel, *types.RoundResult, error) {
	var result *types.RoundResult

	game, err := s.repo.UpdateCurrentGame(ctx, lobbyID, func(g *domain.GameModel) error {
		result = nil

		round := g.CurrentRound
		if round == nil {
			return fmt.Errorf("couldn't find current round for game")
		}

		if err := round.PassTurn(userID); err != nil {
			return err
		}

		if round.Status == domain.RoundStatusRoundOver {
			r := round.ResolveResult()
			if err := g.UpdateScore(&r); err != nil {
				return err
			}
			result = &r
		}
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't update game: %w", err)
	}
	return game, result, nil
}

// NextRound creates a new round
func (s *service) NextRound(ctx context.Context, lobbyID, userID string) (*domain.GameModel, error) {
	game, err := s.repo.UpdateCurrentGame(ctx, lobbyID, func(g *domain.GameModel) error {
		if g.Status == domain.GameStatusGameOver {
			return errors.New("game already over")
		}

		round := g.CurrentRound
		if round == nil {
			return fmt.Errorf("couldn't find current round for game")
		}
		if round.Status != domain.RoundStatusRoundOver {
			return fmt.Errorf("current round is not over")
		}
		if userID != round.PlayerOrder[0] {
			return fmt.Errorf("user is not the owner of the game")
		}

		nextRound, err := domain.NewRound(lobbyID, g.ID, round.RoundNumber+1, round.PlayerOrder, round.StartingPlayer)
		if err != nil {
			return fmt.Errorf("couldn't create next round: %w", err)
		}
		g.CurrentRound = nextRound
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("couldn't update game: %w", err)
	}
	return game, nil
}
