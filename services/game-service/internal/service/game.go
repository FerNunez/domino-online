package service

import (
	"context"
	"crypto/rand"
	"domino/services/game-service/internal/domain"
	"fmt"
	"math/big"
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

func (s *service) CreateGame(ctx context.Context, lobbyID string) (*domain.GameModel, error) {

	pile := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"}
	if err := secureShuffle(pile); err != nil {
		return nil, fmt.Errorf("couldnt shuffle tiles")
	}

	numPlayers := 4
	playerTiles := make([][]string, 0, numPlayers)
	for idx := range numPlayers {
		size := len(pile) / numPlayers
		playerTiles = append(playerTiles, pile[idx*size:idx*size+size])
	}

	game := &domain.GameModel{
		LobbyID:     lobbyID,
		PlayerTiles: playerTiles,
		Head:        make([]string, 0, len(pile)),
		Tail:        make([]string, 0, len(pile)),
	}
	return game, nil
}

// --- Helper ---
func secureShuffle(s []string) error {
	for i := len(s) - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return err
		}
		j := jBig.Int64()
		s[i], s[j] = s[j], s[i]
	}
	return nil
}
