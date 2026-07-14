package grpc

import (
	"context"
	"rebu/services/lobby-service/internal/domain"
	pbl "rebu/shared/proto/lobby"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct {
	pbl.UnimplementedLobbyServiceServer
	service domain.LobbyService
}

func NewGRPCHandler(server *grpc.Server, service domain.LobbyService) *gRPCHandler {
	h := &gRPCHandler{
		UnimplementedLobbyServiceServer: pbl.UnimplementedLobbyServiceServer{},
		service:                         service,
	}
	pbl.RegisterLobbyServiceServer(server, h)
	return h
}

func (h *gRPCHandler) CreateLobby(ctx context.Context, req *pbl.CreateLobbyRequest) (*pbl.CreateLobbyResponse, error) {
	lobby, err := h.service.CreateLobby(ctx, req.UserID, int(req.MaxPlayers))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create lobby : %v", err)
	}

	return &pbl.CreateLobbyResponse{
		Lobby: lobby.ToProto(),
	}, nil
}

func (h *gRPCHandler) JoinLobby(ctx context.Context, req *pbl.JoinLobbyRequest) (*pbl.JoinLobbyResponse, error) {
	lobby, err := h.service.JoinLobby(ctx, req.LobbyID, req.SecretCode, &domain.PlayerModel{
		ID:      primitive.NewObjectID(), // FIX: fix me please, use player ID
		Name:    req.UserID,              // FIX:
		Slot:    0,                       // FIX:
		IsReady: true,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create lobby : %v", err)
	}

	return &pbl.JoinLobbyResponse{
		Lobby: lobby.ToProto(),
	}, nil
}
