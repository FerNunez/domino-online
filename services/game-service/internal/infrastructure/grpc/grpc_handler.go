package grpc

// import (
// 	"context"
// 	"domino/services/game-service/internal/domain"
// 	"domino/services/game-service/internal/infrastructure/events"
//
// 	pbl "domino/shared/proto/game"
//
// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/status"
// )
//
// type gRPCHandler struct {
// 	pbl.UnimplementedGameServiceServer
// 	service   domain.GameService
// 	publisher *events.GameEventPublisher
// }
//
// func NewGRPCHandler(server *grpc.Server, service domain.GameService, publisher *events.GameEventPublisher) *gRPCHandler {
// 	h := &gRPCHandler{
// 		UnimplementedGameServiceServer: pbl.UnimplementedGameServiceServer{},
// 		service:                        service,
// 		publisher:                      publisher,
// 	}
// 	pbl.RegisterGameServiceServer(server, h)
// 	return h
// }
//
// func (h *gRPCHandler) CreateGame(ctx context.Context, req *pbl.CreateGameRequest) (*pbl.CreateGameResponse, error) {
// 	game, err := h.service.CreateGame(ctx, req.UserID, int(req.MaxPlayers))
// 	if err != nil {
// 		return nil, status.Errorf(codes.Internal, "failed to create game : %v", err)
// 	}
//
// 	return &pbl.CreateGameResponse{
// 		Game: game.ToProto(),
// 	}, nil
// }
//
// func (h *gRPCHandler) JoinGame(ctx context.Context, req *pbl.JoinGameRequest) (*pbl.JoinGameResponse, error) {
// 	game, err := h.service.JoinGame(ctx, req.GameID, req.UserID)
// 	if err != nil {
// 		return nil, status.Errorf(codes.Internal, "failed to create game : %v", err)
// 	}
//
// 	return &pbl.JoinGameResponse{
// 		Game: game.ToProto(),
// 	}, nil
// }
//
// func (h *gRPCHandler) StartGame(ctx context.Context, req *pbl.StartGameRequest) (*pbl.StartGameResponse, error) {
// 	game, err := h.service.StartGame(ctx, req.GameID, req.UserID)
// 	if err != nil {
// 		return nil, status.Errorf(codes.Internal, "failed to start game : %v", err)
// 	}
//
// 	if err := h.publisher.PublishStartGame(ctx, game); err != nil {
// 		return nil, status.Errorf(codes.Internal, "failed to publish GameStarted: %v", err)
// 	}
//
// 	return &pbl.StartGameResponse{}, nil
// }
