package repository

import (
	"context"
	"domino/services/game-service/internal/domain"
	"encoding/json"
	"fmt"

	redis "github.com/redis/go-redis/v9"
)

type redisRepository struct {
	client *redis.Client
}

func NewRedisRepository(rdb *redis.Client) *redisRepository {
	return &redisRepository{
		client: rdb,
	}
}

func (rdb *redisRepository) CreateGame(ctx context.Context, g *domain.GameModel) (*domain.GameModel, error) {
	gameData, err := json.Marshal(g)
	if err != nil {
		return nil, err
	}
	if err := rdb.client.Set(ctx, generateGameKey(g.ID), gameData, 0).Err(); err != nil {
		return nil, err
	}
	if err := rdb.client.Set(ctx, generateCurrentGameKey(g.LobbyID), g.ID, 0).Err(); err != nil {
		return nil, err
	}
	return g, nil
}

// GetCurrentGame resolves the lobby's in-progress game via its pointer key.
func (rdb *redisRepository) GetCurrentGame(ctx context.Context, lobbyID string) (*domain.GameModel, error) {
	gameID, err := rdb.client.Get(ctx, generateCurrentGameKey(lobbyID)).Result()
	if err != nil {
		return nil, err
	}
	return rdb.GetGameByID(ctx, gameID)
}

func (rdb *redisRepository) GetGameByID(ctx context.Context, gameID string) (*domain.GameModel, error) {
	val, err := rdb.client.Get(ctx, generateGameKey(gameID)).Result()
	if err != nil {
		return nil, err
	}
	var game domain.GameModel
	if err := json.Unmarshal([]byte(val), &game); err != nil {
		return nil, err
	}
	return &game, nil
}

func (rdb *redisRepository) UpdateGame(ctx context.Context, g *domain.GameModel) (*domain.GameModel, error) {
	gameData, err := json.Marshal(g)
	if err != nil {
		return nil, err
	}
	if err := rdb.client.Set(ctx, generateGameKey(g.ID), gameData, 0).Err(); err != nil {
		return nil, err
	}
	return g, nil
}

// NextGameNumber atomically reserves the next sequential game number for
// lobbyID (1, 2, 3...), so games can be ordered/displayed without depending
// on the uuid primary key.
func (rdb *redisRepository) NextGameNumber(ctx context.Context, lobbyID string) (int, error) {
	n, err := rdb.client.Incr(ctx, generateGameSeqKey(lobbyID)).Result()
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

// ---- helper functions

func generateKey(key, id string) string {
	return fmt.Sprintf("%s:%s", key, id)
}

// generateGameKey is the permanent record of a single game, keyed by its uuid.
func generateGameKey(gameID string) string {
	return generateKey("game", gameID)
}

// generateCurrentGameKey points a lobby at whichever game is currently in progress.
func generateCurrentGameKey(lobbyID string) string {
	return generateKey("lobby-current-game", lobbyID)
}

// generateGameSeqKey backs the atomic counter used to assign GameNumber.
func generateGameSeqKey(lobbyID string) string {
	return generateKey("lobby-game-seq", lobbyID)
}
