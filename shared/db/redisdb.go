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
		//URI: os.Getenv("REDISDB_URI"),
		URI: env.GetString("REDISDB_URI", "localhost:6379"),
		//Database: "ride-sharing",
	}
}

// NewRedisClient creates a new Redis client with the given config
func NewRedisClient(config *RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     config.URI,
		Password: "", // FIX: NO Password
		DB:       0,  // FIX:
		Protocol: 2,  // FIX: Connection protocol

	})
	return client
}
