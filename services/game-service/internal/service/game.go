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
func (s *service) PlayTile(ctx context.Context, lobbyID, userID string, tile types.Tile, side types.Side) (*domain.RoundModel, *types.RoundResult, error) {
	game, err := s.repo.GetCurrentGame(ctx, lobbyID)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't get game: %w", err)
	}

	round := game.CurrentRound
	if round == nil {
		return nil, nil, fmt.Errorf("couldn't find current round for game")
	}

	if err := round.PlayTile(userID, tile, side); err != nil {
		return nil, nil, err
	}

	if _, err = s.repo.UpdateGame(ctx, game); err != nil {
		return nil, nil, fmt.Errorf("couldn't update game: %w", err)
	}

	if round.Status == domain.RoundStatusRoundOver {
		result := round.ResolveResult()
		return round, &result, nil
	}
	return round, nil, nil
}

// Sync called:
// Check if pass move is valid, then pass to next user and updates game.
// Round can end if all 4 players pass in a row (blocked board).
func (s *service) PassTurn(ctx context.Context, lobbyID, userID string) (*domain.RoundModel, *types.RoundResult, error) {
	game, err := s.repo.GetCurrentGame(ctx, lobbyID)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't get game: %w", err)
	}

	round := game.CurrentRound
	if round == nil {
		return nil, nil, fmt.Errorf("couldn't find current round for game")
	}

	if err := round.PassTurn(userID); err != nil {
		return nil, nil, err
	}

	if _, err = s.repo.UpdateGame(ctx, game); err != nil {
		return nil, nil, fmt.Errorf("couldn't update game: %w", err)
	}

	if round.Status == domain.RoundStatusRoundOver {
		result := round.ResolveResult()
		return round, &result, nil
	}
	return round, nil, nil
}

// Sync called:
// Resolves the just-finished round into the match score, then either ends
// the match (a team reached GoalScore) or deals the next round.
func (s *service) NextRound(ctx context.Context, lobbyID string) (*domain.RoundModel, *types.RoundResult, error) {
	game, err := s.repo.GetCurrentGame(ctx, lobbyID)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't get game: %w", err)
	}

	round := game.CurrentRound
	if round == nil {
		return nil, nil, fmt.Errorf("couldn't find current round for game")
	}
	if round.Status != domain.RoundStatusRoundOver {
		return nil, nil, fmt.Errorf("current round is not over")
	}

	result := round.ResolveResult()
	for teamID, pips := range result.Scores {
		game.TeamScores[teamID] += pips
	}

	if game.TeamScores[types.TeamA] >= game.GoalScore || game.TeamScores[types.TeamB] >= game.GoalScore {
		game.Status = domain.GameStatusGameOver
		if _, err = s.repo.UpdateGame(ctx, game); err != nil {
			return nil, nil, fmt.Errorf("couldn't update game: %w", err)
		}
		return round, &result, nil
	}

	nextRound, err := domain.NewRound(lobbyID, game.ID, round.ID+1, round.PlayerOrder, round.CurrentTurn)
	if err != nil {
		return nil, nil, fmt.Errorf("couldn't deal next round: %w", err)
	}
	game.CurrentRound = nextRound

	if _, err = s.repo.UpdateGame(ctx, game); err != nil {
		return nil, nil, fmt.Errorf("couldn't update game: %w", err)
	}
	return nextRound, &result, nil
}
