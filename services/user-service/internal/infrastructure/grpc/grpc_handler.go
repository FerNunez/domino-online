package grpc_handler

import (
	"context"
	"rebu/services/user-service/internal/domain"
	pbu "rebu/shared/proto/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct {
	pbu.UnimplementedUserServiceServer
	service domain.UserService
}

func NewGRPCHandler(server *grpc.Server, service domain.UserService) *gRPCHandler {
	h := &gRPCHandler{
		UnimplementedUserServiceServer: pbu.UnimplementedUserServiceServer{},
		service:                        service,
	}
	pbu.RegisterUserServiceServer(server, h)
	return h
}

func (h *gRPCHandler) GetUser(ctx context.Context, req *pbu.GetUserRequest) (*pbu.AuthResponse, error) {
	guestUser, err := h.service.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user : %v", err)
	}

	return &pbu.AuthResponse{
		User: guestUser.ToProto(),
	}, nil
}

func (h *gRPCHandler) Register(ctx context.Context, req *pbu.RegisterRequest) (*pbu.AuthResponse, error) {
	guestUser, err := h.service.CreateUser(ctx, req.UserID, req.Email, req.Password, req.DisplayName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user : %v", err)
	}

	return &pbu.AuthResponse{
		User: guestUser.ToProto(),
	}, nil
}
