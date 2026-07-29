package grpc

import (
	"context"
	"domino/services/game-service/internal/domain"
	"domino/services/game-service/internal/infrastructure/events"

	pbg "domino/shared/proto/game"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct {
	pbg.UnimplementedGameServiceServer
	service   domain.GameService
	publisher *events.GameEventPublisher
}

// publisher *events.GameEventPublisher
func NewGRPCHandler(server *grpc.Server, service domain.GameService, publisher *events.GameEventPublisher) *gRPCHandler {
	h := &gRPCHandler{
		UnimplementedGameServiceServer: pbg.UnimplementedGameServiceServer{},
		service:                        service,
		publisher:                      publisher,
	}
	pbg.RegisterGameServiceServer(server, h)
	return h
}

func (h *gRPCHandler) PlayTile(ctx context.Context, req *pbg.PlayTileRequest) (*pbg.PlayTileResponse, error) {
	// check req validity
	if req.Side != "left" && req.Side != "right" {
		return nil, status.Error(codes.Internal, "failed to parse side")
	}
	tile := domain.Tile{Left: int(req.Tile.Left), Right: int(req.Tile.Right)}

	// Play tile
	game, roundResult, err := h.service.PlayTile(ctx, req.LobbyId, req.UserId, tile, req.Side)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to play tile: %v", err)
	}

	// publish move
	if err := h.publisher.PublishMoveMade(ctx, game.LobbyID, req.UserId, tile, req.Side); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish move made: %v", err)
	}

	// return grpc
	return &pbg.PlayTileResponse{
		Board:       toProtoTiles(game.Board.Tiles),
		Hand:        toProtoTiles(game.Hands[req.UserId]),
		RoundResult: toProtoRoundResult(roundResult),
	}, nil
}

func (h *gRPCHandler) PassTurn(ctx context.Context, req *pbg.PassTurnRequest) (*pbg.PassTurnResponse, error) {
	// Pass turn
	_, roundResult, err := h.service.PassTurn(ctx, req.LobbyId, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to pass turn: %v", err)
	}

	if err := h.publisher.PublishTurnChanged(ctx, req.LobbyId, req.UserId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish GameStarted: %v", err)
	}

	return &pbg.PassTurnResponse{
		RoundResult: toProtoRoundResult(roundResult),
	}, nil
}

func toProtoRoundResult(rr *domain.RoundResult) *pbg.RoundResult {
	if rr == nil {
		return nil
	}
	return &pbg.RoundResult{
		WinnerId: rr.WinnerID,
		Reason:   rr.Reason,
	}
}

func toProtoTiles(tiles []domain.Tile) []*pbg.Tile {
	out := make([]*pbg.Tile, len(tiles))
	for i, t := range tiles {
		out[i] = &pbg.Tile{Left: int32(t.Left), Right: int32(t.Right)}
	}
	return out
}
