package domain

import (
	"context"
	pbu "rebu/shared/proto/user"

	"github.com/google/uuid"
)

type User struct {
	ID          string
	DisplayName string
	Type        string
}

func (u User) ToProto() *pbu.User {
	return &pbu.User{
		Id:          u.ID,
		DisplayName: u.DisplayName,
		Type:        u.Type,
	}
}

type UserRepository interface {
	CreateGuest(ctx context.Context) (*User, error)
	GetUserByID(ctx context.Context, userUUID uuid.UUID) (*User, error)
	//UpdateUser(ctx context.Context, user User)
}

type UserService interface {
	CreateGuest(ctx context.Context) (*User, error)
	GetUserByID(ctx context.Context, userID string) (*User, error)

	// RegisterUser(ctx context.Context)
	// Login(ctx context.Context)

	//ValidateToken(ctx context.Context)
}
