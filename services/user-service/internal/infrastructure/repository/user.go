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
	dbUser, err := s.queries.CreateUser(ctx, sql.CreateUserParams{
		ID:             pgtype.UUID{Bytes: uuid.New(), Valid: true},
		DisplayName:    "guest",
		Email:          "guest",
		HashedPassword: "guest",
	})
	if err != nil {
		return nil, err
	}
	return &domain.User{
		ID:          uuid.UUID(dbUser.ID.Bytes).String(),
		DisplayName: dbUser.DisplayName,
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
	}, nil
}

func (s *SqlRepository) CreateUser(ctx context.Context, userUUID uuid.UUID, email, hashedPassword, displayName string) (*domain.User, error) {
	dbUser, err := s.queries.CreateUser(ctx, sql.CreateUserParams{
		ID:             pgtype.UUID{Bytes: userUUID, Valid: true},
		Email:          email,
		HashedPassword: hashedPassword,
		DisplayName:    displayName,
	})
	if err != nil {
		return nil, err
	}
	return &domain.User{
		ID:          uuid.UUID(dbUser.ID.Bytes).String(),
		DisplayName: dbUser.DisplayName,
	}, nil
}
