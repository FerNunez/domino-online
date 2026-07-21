package db

import (
	"context"
	"domino/shared/env"
	"fmt"
	"time"

	"domino/shared/db/sql"

	"github.com/jackc/pgx/v5"
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

func NewSqlQueries(ctx context.Context, cfg *SQLConfig) (*sql.Queries, error) {

	if cfg.URI == "" {
		return nil, fmt.Errorf("mongodb URI is required")
	}

	connCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()

	conn, err := pgx.Connect(connCtx, cfg.URI)
	if err != nil {
		return nil, err
	}
	q := sql.New(conn)
	return q, nil
}
