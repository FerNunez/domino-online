package service

import (
	"context"
	"domino/services/user-service/internal/domain"
	"domino/shared/hash"

	"github.com/google/uuid"
)

type Service struct {
	repo domain.UserRepository
}

func NewService(repo domain.UserRepository) *Service {
	return &Service{
		repo: repo,
	}
}

func (u *Service) CreateUser(ctx context.Context, userID, email, password, displayName string) (*domain.User, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := hash.HashPassword(password)
	if err != nil {
		return nil, err
	}

	return u.repo.CreateUser(ctx, userUUID, email, hashedPassword, displayName)
}

func (u *Service) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	user, err := u.repo.GetUserByID(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	return user, nil
}
