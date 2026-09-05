package db

import (
	"log"

	"domino/shared/env"

	"github.com/redis/go-redis/extra/redisotel/v9"
	redis "github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	URI string
	// Database string
}

// Reads REDISDB_URI from the environment
// Returns empty URI if unset. // The caller must validate/check
func NewRedisDefaultConfig() *RedisConfig {
	return &RedisConfig{
		URI: env.GetString("REDIS_URI", "localhost:6379"),
	}
}

// NewRedisClient creates a new Redis client with the given config.
// Every command issued on the returned client is automatically instrumented with otlp span
func NewRedisClient(config *RedisConfig) *redis.Client {
	opt, _ := redis.ParseURL(config.URI)
	client := redis.NewClient(opt)
	if err := redisotel.InstrumentTracing(client); err != nil {
		log.Printf("redis: failed to instrument tracing: %v", err)
	}
	return client
}
