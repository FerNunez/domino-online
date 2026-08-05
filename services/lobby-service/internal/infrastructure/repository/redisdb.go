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

// lobbyMeta is contain lobby's non-player fields. Players are
// stored separately (see playersKey) so that connecting/disconnecting/joining
type lobbyMeta struct {
	ID         string               `json:"id"`
	HostID     string               `json:"hostID"`
	Status     domain.LobbyStatus   `json:"status"`
	MaxPlayers int                  `json:"maxPlayers"`
	Settings   domain.LobbySettings `json:"settings"`
}

func (rdb *redisRepository) CreateLobby(ctx context.Context, l *domain.LobbyModel) (*domain.LobbyModel, error) {
	if err := rdb.saveMeta(ctx, l); err != nil {
		return nil, err
	}
	for _, p := range l.Players {
		if err := rdb.AddPlayer(ctx, l.ID, p); err != nil {
			return nil, err
		}
	}
	return l, nil
}

func (rdb *redisRepository) GetLobbyByID(ctx context.Context, id string) (*domain.LobbyModel, error) {
	// Read LobbyMeta
	val, err := rdb.client.Get(ctx, metaKey(id)).Result()
	if err != nil {
		return nil, err
	}
	var meta lobbyMeta
	if err := json.Unmarshal([]byte(val), &meta); err != nil {
		return nil, err
	}

	// Read lobby's player data
	playerVals, err := rdb.client.HGetAll(ctx, playersKey(id)).Result()
	if err != nil {
		return nil, err
	}
	players := make(map[string]*domain.PlayerModel, len(playerVals))
	for userID, raw := range playerVals {
		var p domain.PlayerModel
		if err := json.Unmarshal([]byte(raw), &p); err != nil {
			return nil, err
		}
		players[userID] = &p
	}

	return &domain.LobbyModel{
		ID:         meta.ID,
		HostID:     meta.HostID,
		Status:     meta.Status,
		Players:    players,
		MaxPlayers: meta.MaxPlayers,
		Settings:   meta.Settings,
	}, nil
}

func (rdb *redisRepository) SetStatus(ctx context.Context, lobbyID string, status domain.LobbyStatus) error {
	key := metaKey(lobbyID)
	val, err := rdb.client.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	var meta lobbyMeta
	if err := json.Unmarshal([]byte(val), &meta); err != nil {
		return err
	}
	meta.Status = status

	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return rdb.client.Set(ctx, key, data, 0).Err()
}

func (rdb *redisRepository) AddPlayer(ctx context.Context, lobbyID string, p *domain.PlayerModel) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return rdb.client.HSet(ctx, playersKey(lobbyID), p.ID, data).Err()
}

func (rdb *redisRepository) SetPlayerConnection(ctx context.Context, lobbyID string, userID string, connected bool) error {
	key := playersKey(lobbyID)
	val, err := rdb.client.HGet(ctx, key, userID).Result()
	if err == redis.Nil {
		return fmt.Errorf("player %s not found in lobby %s", userID, lobbyID)
	}
	if err != nil {
		return err
	}

	var player domain.PlayerModel
	if err := json.Unmarshal([]byte(val), &player); err != nil {
		return err
	}
	player.IsConnected = connected

	data, err := json.Marshal(player)
	if err != nil {
		return err
	}
	return rdb.client.HSet(ctx, key, userID, data).Err()
}

func (rdb *redisRepository) saveMeta(ctx context.Context, l *domain.LobbyModel) error {
	meta := lobbyMeta{
		ID:         l.ID,
		HostID:     l.HostID,
		Status:     l.Status,
		MaxPlayers: l.MaxPlayers,
		Settings:   l.Settings,
	}
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return rdb.client.Set(ctx, metaKey(l.ID), data, 0).Err()
}

// ---- helper functions

func generateKey(key, id string) string {
	return fmt.Sprintf("%s:%s", key, id)
}

func metaKey(id string) string {
	return generateKey("lobby", id)
}

func playersKey(id string) string {
	return generateKey("lobby", id) + ":players"
}
