package repository

import (
	"context"
	"rebu/services/user-service/internal/domain"
	"rebu/shared/db/sql"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type SqlRepository struct {
	queries *sql.Queries
}

func NewSqlRepository(queries *sql.Queries) *SqlRepository {
	return &SqlRepository{queries: queries}
}

func (s *SqlRepository) CreateGuest(ctx context.Context) (*domain.User, error) {
	dbUser, err := s.queries.CreateGuest(ctx, sql.CreateGuestParams{
		ID:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
		DisplayName: "guest",
		Type:        "guest",
	})
	if err != nil {
		return nil, err
	}
	return &domain.User{
		ID:          uuid.UUID(dbUser.ID.Bytes).String(),
		DisplayName: dbUser.DisplayName,
		Type:        dbUser.Type,
	}, nil
}

func (s *SqlRepository) GetUserByID(ctx context.Context, userUUID uuid.UUID) (*domain.User, error) {
	dbUser, err := s.queries.GetUserByID(ctx, pgtype.UUID{
		Bytes: userUUID,
		Valid: true,
	})
	if err != nil {
		return nil, err
	}
	return &domain.User{
		ID:          uuid.UUID(dbUser.ID.Bytes).String(),
		DisplayName: dbUser.DisplayName,
		Type:        dbUser.Type,
	}, nil
}
