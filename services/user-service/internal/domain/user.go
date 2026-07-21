package domain

import (
	"context"
	pbu "domino/shared/proto/user"

	"github.com/google/uuid"
)

type User struct {
	ID          string
	DisplayName string
}

func (u User) ToProto() *pbu.User {
	return &pbu.User{
		Id:          u.ID,
		DisplayName: u.DisplayName,
	}
}

type UserRepository interface {
	GetUserByID(ctx context.Context, userUUID uuid.UUID) (*User, error)
	CreateUser(ctx context.Context, userUUID uuid.UUID, email, hashedPassword, displayName string) (*User, error)
}

type UserService interface {
	GetUserByID(ctx context.Context, userID string) (*User, error)
	CreateUser(ctx context.Context, userID, email, password, displayName string) (*User, error)
}
