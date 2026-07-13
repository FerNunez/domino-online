package domain

import (
	"context"
	pbu "rebu/shared/proto/user"

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
	CreateGuest(ctx context.Context) (*User, error)
	GetUserByID(ctx context.Context, userUUID uuid.UUID) (*User, error)
	CreateUser(ctx context.Context, userUUID uuid.UUID, email, hashedPassword, displayName string) (*User, error)
}

type UserService interface {
	CreateGuest(ctx context.Context) (*User, error)
	GetUserByID(ctx context.Context, userID string) (*User, error)
	CreateUser(ctx context.Context, userID, email, hashedPassword, displayName string) (*User, error)
}
