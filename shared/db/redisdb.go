package db

import (
	"rebu/shared/env"

	redis "github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	URI string
	//Database string
}

// Reads REDISDB_URI from the environment
// Returns empty URI if unset. // The caller must validate/check
func NewRedisDefaultConfig() *RedisConfig {
	return &RedisConfig{
		URI: env.GetString("REDIS_URI", "localhost:6379"),
	}
}

// NewRedisClient creates a new Redis client with the given config
func NewRedisClient(config *RedisConfig) *redis.Client {
	opt, _ := redis.ParseURL(config.URI)
	client := redis.NewClient(opt)
	return client
}
