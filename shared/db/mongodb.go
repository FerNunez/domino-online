package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	TripsCollection     = "trips"
	RideFaresCollection = "ride_fares"
)

type MongoConfig struct {
	URI      string
	Database string
}

// Reads MONGODB_URI from the environment
// Returns empty URI if unset. // The caller must validate/check
func NewMongoDefaultConfig() *MongoConfig {
	return &MongoConfig{
		URI:      os.Getenv("MONGODB_URI"),
		Database: "ride-sharing",
	}
}

// NewMongoClient creates a connected, verified MongoDB client
// Calls Ping before returning
func NewMongoClient(ctx context.Context, cfg *MongoConfig) (*mongo.Client, error) {
	if cfg.URI == "" {
		return nil, fmt.Errorf("mongodb URI is required")
	}
	if cfg.Database == "" {
		return nil, fmt.Errorf("mongodb databese is required")
	}

	connCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client, err := mongo.Connect(connCtx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(connCtx, readpref.Primary()); err != nil {
		return nil, err
	}

	log.Printf("Successfully connected to MongoDB at %s", cfg.URI)
	return client, nil
}

// GetDatabase returns a *mongo.Database handle from a connected client
func GetDatabase(client *mongo.Client, cfg *MongoConfig) *mongo.Database {
	return client.Database(cfg.Database)
}
