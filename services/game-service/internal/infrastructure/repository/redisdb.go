package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"domino/services/game-service/internal/domain"

	redis "github.com/redis/go-redis/v9"
)

// maxCASRetries bounds how many times UpdateCurrentGame will re-run mutate
// against a fresh read after losing a race to another writer.
const maxCASRetries = 5

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

// UpdateCurrentGame resolves lobbyID's current game and atomically applies mutate to it using Redis's WATCH/MULTI/EXEC
func (rdb *redisRepository) UpdateCurrentGame(ctx context.Context, lobbyID string, mutate func(*domain.GameModel) error) (*domain.GameModel, error) {
	gameID, err := rdb.client.Get(ctx, generateCurrentGameKey(lobbyID)).Result()
	if err != nil {
		return nil, fmt.Errorf("couldn't resolve current game for lobby %s: %w", lobbyID, err)
	}
	key := generateGameKey(gameID)

	var game domain.GameModel
	for range maxCASRetries {
		txErr := rdb.client.Watch(ctx, func(tx *redis.Tx) error {
			val, err := tx.Get(ctx, key).Result()
			if err != nil {
				return err
			}
			game = domain.GameModel{}
			if err := json.Unmarshal([]byte(val), &game); err != nil {
				return err
			}

			if err := mutate(&game); err != nil {
				return err // domain rejection: WATCH is released, nothing is queued/written
			}

			data, err := json.Marshal(game)
			if err != nil {
				return err
			}
			_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.Set(ctx, key, data, 0)
				return nil
			})
			return err
		}, key)

		if txErr == nil {
			return &game, nil
		}
		if !errors.Is(txErr, redis.TxFailedErr) {
			// Domain error or infra failure: retrying would fail the same way.
			return nil, txErr
		}
		// Lost the race: `key` changed between our GET and EXEC. Loop and
		// retry mutate against a fresh read.
	}
	return nil, fmt.Errorf("couldn't update game %s after %d attempts: too much contention", gameID, maxCASRetries)
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
