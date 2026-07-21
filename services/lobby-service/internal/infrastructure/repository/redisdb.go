package repository

import (
	"context"
	"domino/services/lobby-service/internal/domain"
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

func (rdb *redisRepository) CreateLobby(ctx context.Context, l *domain.LobbyModel) (*domain.LobbyModel, error) {
	key := generateKey("lobby", l.ID)
	lobbyData, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}
	if err := rdb.client.Set(ctx, key, lobbyData, 0).Err(); err != nil {
		return nil, err
	}
	return l, nil
}

func (rdb *redisRepository) GetLobbyByID(ctx context.Context, id string) (*domain.LobbyModel, error) {
	key := generateKey("lobby", id)
	var lobby domain.LobbyModel
	val, err := rdb.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(val), &lobby)
	if err != nil {
		return nil, err
	}
	return &lobby, nil
}

func (rdb *redisRepository) UpdateLobby(ctx context.Context, id string, l *domain.LobbyModel) (*domain.LobbyModel, error) {
	key := generateKey("lobby", l.ID)
	lobbyData, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}
	if err := rdb.client.Set(ctx, key, lobbyData, 0).Err(); err != nil {
		return nil, err
	}
	return l, nil
}

// func (rdb *redisRepository) StartLobby(ctx context.Context, id string) error {
// }

// ---- helper function

func generateKey(key, id string) string {
	return fmt.Sprintf("%s:%s", key, id)
}
