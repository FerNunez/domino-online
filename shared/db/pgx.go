package db

import (
	"context"
	"fmt"
	"time"

	"domino/shared/env"

	"domino/shared/db/sql"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	UsersCollection = "users"
)

type SQLConfig struct {
	URI      string
	Database string
}

func NewSQLDefaultConfig() *SQLConfig {
	return &SQLConfig{
		URI:      env.GetString("POSTGRESQL_URI", "postgres://domino:domino@localhost:5433/domino"),
		Database: "ride-sharing",
	}
}

func NewSQLQueries(ctx context.Context, cfg *SQLConfig) (*sql.Queries, error) {
	if cfg.URI == "" {
		return nil, fmt.Errorf("mongodb URI is required")
	}

	connCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	pool, err := pgxpool.New(connCtx, cfg.URI)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(connCtx); err != nil {
		return nil, err
	}
	q := sql.New(pool)
	return q, nil
}
