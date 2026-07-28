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
	key := generateKey("game", g.LobbyID)
	gameData, err := json.Marshal(g)
	if err != nil {
		return nil, err
	}
	if err := rdb.client.Set(ctx, key, gameData, 0).Err(); err != nil {
		return nil, err
	}
	return g, nil
}

func (rdb *redisRepository) GetGameByLobbyID(ctx context.Context, lobbyID string) (*domain.GameModel, error) {
	key := generateKey("game", lobbyID)
	var game domain.GameModel
	val, err := rdb.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(val), &game); err != nil {
		return nil, err
	}
	return &game, nil
}

func (rdb *redisRepository) UpdateGame(ctx context.Context, lobbyID string, g *domain.GameModel) (*domain.GameModel, error) {
	key := generateKey("game", lobbyID)
	gameData, err := json.Marshal(g)
	if err != nil {
		return nil, err
	}
	if err := rdb.client.Set(ctx, key, gameData, 0).Err(); err != nil {
		return nil, err
	}
	return g, nil
}

// ---- helper function

func generateKey(key, id string) string {
	return fmt.Sprintf("%s:%s", key, id)
}
