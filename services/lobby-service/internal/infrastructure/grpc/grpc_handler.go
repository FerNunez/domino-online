package grpc

import (
	"context"
	"domino/services/lobby-service/internal/domain"
	"domino/services/lobby-service/internal/infrastructure/events"

	pbl "domino/shared/proto/lobby"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct {
	pbl.UnimplementedLobbyServiceServer
	service   domain.LobbyService
	publisher *events.LobbyEventPublisher
}

func NewGRPCHandler(server *grpc.Server, service domain.LobbyService, publisher *events.LobbyEventPublisher) *gRPCHandler {
	h := &gRPCHandler{
		UnimplementedLobbyServiceServer: pbl.UnimplementedLobbyServiceServer{},
		service:                         service,
		publisher:                       publisher,
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
	lobby, err := h.service.JoinLobby(ctx, req.LobbyID, req.UserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create lobby : %v", err)
	}

	return &pbl.JoinLobbyResponse{
		Lobby: lobby.ToProto(),
	}, nil
}

func (h *gRPCHandler) StartLobby(ctx context.Context, req *pbl.StartLobbyRequest) (*pbl.StartLobbyResponse, error) {
	lobby, err := h.service.StartLobby(ctx, req.LobbyID, req.UserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to start lobby : %v", err)
	}

	if err := h.publisher.PublishStartGame(ctx, lobby); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish GameStarted: %v", err)
	}

	return &pbl.StartLobbyResponse{}, nil
}
