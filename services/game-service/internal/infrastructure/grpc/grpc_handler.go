package grpc

import (
	"context"
	"domino/services/game-service/internal/domain"

	pbg "domino/shared/proto/game"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct {
	pbg.UnimplementedGameServiceServer
	service domain.GameService
	//publisher *events.GameEventPublisher
}

// publisher *events.GameEventPublisher
func NewGRPCHandler(server *grpc.Server, service domain.GameService) *gRPCHandler {
	h := &gRPCHandler{
		UnimplementedGameServiceServer: pbg.UnimplementedGameServiceServer{},
		service:                        service,
		//publisher:                      publisher,
	}
	pbg.RegisterGameServiceServer(server, h)
	return h
}

func (h *gRPCHandler) PlayTile(ctx context.Context, req *pbg.PlayTileRequest) (*pbg.PlayTileResponse, error) {
	if req.Side != "left" && req.Side != "right" {
		return nil, status.Error(codes.Internal, "failed to parse side")
	}
	tile := domain.Tile{Left: int(req.Tile.Left), Right: int(req.Tile.Right)}

	game, _, err := h.service.PlayTile(ctx, req.LobbyId, req.UserId, tile, req.Side)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to play tile : %v", err)
	}

	// TODO: Publish game played righ?
	// if err := h.publisher.PublishStartGame(ctx, game); err != nil {
	// 	return nil, status.Errorf(codes.Internal, "failed to publish GameStarted: %v", err)
	// }
	//

	return &pbg.PlayTileResponse{
		Board: toProtoTiles(game.Board.Tiles),
		Hand:  toProtoTiles(game.Hands[req.UserId]),
	}, nil
}

func toProtoTiles(tiles []domain.Tile) []*pbg.Tile {
	out := make([]*pbg.Tile, len(tiles))
	for i, t := range tiles {
		out[i] = &pbg.Tile{Left: int32(t.Left), Right: int32(t.Right)}
	}
	return out
}
