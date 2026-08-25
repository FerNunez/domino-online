package grpc

import (
	"context"

	"domino/services/game-service/internal/domain"
	"domino/services/game-service/internal/infrastructure/events"

	pbg "domino/shared/proto/game"
	"domino/shared/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct {
	pbg.UnimplementedGameServiceServer
	service   domain.GameService
	publisher *events.GameEventPublisher
}

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
	side := types.Side(req.Side)
	if side != types.Left && side != types.Right {
		return nil, status.Error(codes.Internal, "failed to parse side")
	}
	tile := types.Tile{Left: int(req.Tile.Left), Right: int(req.Tile.Right)}

	// Play tile
	game, roundResult, err := h.service.PlayTile(ctx, req.LobbyId, req.UserId, tile, side)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to play tile: %v", err)
	}

	round := game.CurrentRound

	// publish move made
	if err := h.publisher.PublishMoveMade(ctx, req, round, roundResult); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish move made: %v", err)
	}

	if round.Status == domain.RoundStatusRoundOver && roundResult != nil {
		if err := h.publisher.PublishRoundOver(ctx, game, roundResult); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to publish GameStarted: %v", err)
		}
		if game.Status == domain.GameStatusGameOver {
			if err := h.publisher.PublishGameOver(ctx, game); err != nil {
				return nil, status.Errorf(codes.Internal, "failed to publish GameOver: %v", err)
			}
		}
	}

	// return grpc
	return &pbg.PlayTileResponse{
		Board:       toProtoTiles(round.Board.Tiles),
		Hand:        toProtoTiles(round.Hands[req.UserId]),
		RoundResult: toProtoRoundResult(roundResult),
	}, nil
}

func (h *gRPCHandler) PassTurn(ctx context.Context, req *pbg.PassTurnRequest) (*pbg.PassTurnResponse, error) {
	// Pass turn
	game, roundResult, err := h.service.PassTurn(ctx, req.LobbyId, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to pass turn: %v", err)
	}

	round := game.CurrentRound
	// publish turn passed
	if err := h.publisher.PublishTurnPassed(ctx, req, round, roundResult); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish GameStarted: %v", err)
	}

	if round.Status == domain.RoundStatusRoundOver && roundResult != nil {
		if err := h.publisher.PublishRoundOver(ctx, game, roundResult); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to publish GameStarted: %v", err)
		}
		if game.Status == domain.GameStatusGameOver {
			if err := h.publisher.PublishGameOver(ctx, game); err != nil {
				return nil, status.Errorf(codes.Internal, "failed to publish GameOver: %v", err)
			}
		}
	}

	return &pbg.PassTurnResponse{
		RoundResult: toProtoRoundResult(roundResult),
	}, nil
}

func (h *gRPCHandler) NextRound(ctx context.Context, req *pbg.NextRoundRequest) (*pbg.NextRoundResponse, error) {
	game, err := h.service.NextRound(ctx, req.LobbyId, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to start next round: %v", err)
	}

	round := game.CurrentRound
	if err := h.publisher.PublishRoundStarted(ctx, round); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish RoundStarted: %v", err)
	}
	if err := h.publisher.PublishHandsDealt(ctx, round); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to publish HandsDealt: %v", err)
	}

	return &pbg.NextRoundResponse{
		RoundNumber: int32(round.RoundNumber),
	}, nil
}

func toProtoRoundResult(rr *types.RoundResult) *pbg.RoundResult {
	if rr == nil {
		return nil
	}
	return &pbg.RoundResult{
		WinnerId: string(rr.WinnerTeamID),
		Reason:   string(rr.Reason),
		Scores:   toProtoScores(rr.Scores),
	}
}

func toProtoTiles(tiles []types.Tile) []*pbg.Tile {
	out := make([]*pbg.Tile, len(tiles))
	for i, t := range tiles {
		out[i] = &pbg.Tile{Left: int32(t.Left), Right: int32(t.Right)}
	}
	return out
}

func toProtoScores(score map[types.TeamID]int) map[string]int32 {
	protoScore := make(map[string]int32)
	for team, s := range score {
		protoScore[string(team)] = int32(s)
	}
	return protoScore
}
