package repository

// import (
// 	"context"
// 	"domino/services/game-service/internal/domain"
// 	"encoding/json"
// 	"fmt"
//
// 	redis "github.com/redis/go-redis/v9"
// )

//
// type redisRepository struct {
// 	client *redis.Client
// }
//
// func NewRedisRepository(rdb *redis.Client) *redisRepository {
// 	return &redisRepository{
// 		client: rdb,
// 	}
// }
//
// func (rdb *redisRepository) CreateGame(ctx context.Context, l *domain.GameModel) (*domain.GameModel, error) {
// 	key := generateKey("game", l.ID)
// 	gameData, err := json.Marshal(l)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if err := rdb.client.Set(ctx, key, gameData, 0).Err(); err != nil {
// 		return nil, err
// 	}
// 	return l, nil
// }
//
// func (rdb *redisRepository) GetGameByID(ctx context.Context, id string) (*domain.GameModel, error) {
// 	key := generateKey("game", id)
// 	var game domain.GameModel
// 	val, err := rdb.client.Get(ctx, key).Result()
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = json.Unmarshal([]byte(val), &game)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &game, nil
// }
//
// func (rdb *redisRepository) UpdateGame(ctx context.Context, id string, l *domain.GameModel) (*domain.GameModel, error) {
// 	key := generateKey("game", l.ID)
// 	gameData, err := json.Marshal(l)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if err := rdb.client.Set(ctx, key, gameData, 0).Err(); err != nil {
// 		return nil, err
// 	}
// 	return l, nil
// }
//
// // func (rdb *redisRepository) StartGame(ctx context.Context, id string) error {
// // }
//
// // ---- helper function
//
// func generateKey(key, id string) string {
// 	return fmt.Sprintf("%s:%s", key, id)
// }
