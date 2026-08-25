package grpc

import (
	"context"
	"errors"
	"time"

	"domino/services/history-service/internal/domain"
	pbh "domino/shared/proto/history"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type gRPCHandler struct {
	pbh.UnimplementedHistoryServiceServer
	repo domain.HistoryRepository
}

func NewGRPCHandler(server *grpc.Server, repo domain.HistoryRepository) *gRPCHandler {
	h := &gRPCHandler{
		UnimplementedHistoryServiceServer: pbh.UnimplementedHistoryServiceServer{},
		repo:                              repo,
	}
	pbh.RegisterHistoryServiceServer(server, h)
	return h
}

func (h *gRPCHandler) GetRoundActions(ctx context.Context, req *pbh.GetRoundActionsRequest) (*pbh.GetRoundActionsResponse, error) {
	round, err := h.repo.GetRoundByID(ctx, req.RoundId)
	if err != nil {
		if errors.Is(err, domain.ErrRoundNotFound) {
			return nil, status.Errorf(codes.NotFound, "round %s not found or not yet persisted", req.RoundId)
		}
		return nil, status.Errorf(codes.Internal, "failed to look up round: %v", err)
	}

	actions, err := h.repo.GetActionsByRoundID(ctx, req.RoundId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get round actions: %v", err)
	}
	hands, err := h.repo.GetHandsByRoundID(ctx, req.RoundId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get round hands: %v", err)
	}

	// Check if received number actions and round number of action is the same, meaning data being stored
	if len(actions) < round.ActionCount || len(hands) < len(round.PlayerOrder) {
		return nil, status.Errorf(codes.NotFound, "round %s not fully persisted yet", req.RoundId)
	}

	return &pbh.GetRoundActionsResponse{
		Actions: toProtoActions(actions),
		Hands:   toProtoHands(hands),
	}, nil
}

func (h *gRPCHandler) GetPlayerGames(ctx context.Context, req *pbh.GetPlayerGamesRequest) (*pbh.GetPlayerGamesResponse, error) {
	// TODO: add proper paging
	limit := int(req.Limit)
	if limit <= 0 {
		limit = 20
	}
	games, err := h.repo.GetGamesByPlayerID(ctx, req.PlayerId, limit, int(req.Offset))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get player games: %v", err)
	}

	return &pbh.GetPlayerGamesResponse{
		Games: toProtoGameSummaries(games),
	}, nil
}

func (h *gRPCHandler) GetGameHistory(ctx context.Context, req *pbh.GetGameHistoryRequest) (*pbh.GetGameHistoryResponse, error) {
	rounds, err := h.repo.GetRoundsByGameID(ctx, req.GameId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get game history: %v", err)
	}

	return &pbh.GetGameHistoryResponse{
		Rounds: toProtoRoundSummaries(rounds),
	}, nil
}

func toProtoActions(actions []domain.Action) []*pbh.Action {
	out := make([]*pbh.Action, len(actions))
	for i, a := range actions {
		pa := &pbh.Action{
			ActionNumber: int32(a.ActionNumber),
			PlayerId:     a.PlayerID,
			ActionType:   string(a.ActionType),
			Side:         string(a.Side),
		}
		if a.TileLeft != nil && a.TileRight != nil {
			pa.Tile = &pbh.Tile{Left: int32(*a.TileLeft), Right: int32(*a.TileRight)}
		}
		if a.ResultingLeftEnd != nil {
			pa.ResultingLeftEnd = int32(*a.ResultingLeftEnd)
		}
		if a.ResultingRightEnd != nil {
			pa.ResultingRightEnd = int32(*a.ResultingRightEnd)
		}
		out[i] = pa
	}
	return out
}

func toProtoHands(hands []domain.Hand) []*pbh.Hand {
	out := make([]*pbh.Hand, len(hands))
	for i, hand := range hands {
		tiles := make([]*pbh.Tile, len(hand.Tiles))
		for j, t := range hand.Tiles {
			tiles[j] = &pbh.Tile{Left: int32(t.Left), Right: int32(t.Right)}
		}
		out[i] = &pbh.Hand{
			PlayerId: hand.PlayerID,
			Tiles:    tiles,
		}
	}
	return out
}

func toProtoGameSummaries(games []domain.Game) []*pbh.GameSummary {
	out := make([]*pbh.GameSummary, len(games))
	for i, g := range games {
		finalScores := make(map[string]int32, len(g.FinalScores))
		for team, score := range g.FinalScores {
			finalScores[team] = int32(score)
		}
		out[i] = &pbh.GameSummary{
			GameId:      g.ID,
			LobbyId:     g.LobbyID,
			FinalScores: finalScores,
			TeamWinner:  g.TeamWinner,
			GameState:   g.GameState,
			CreatedAt:   g.CreatedAt.Format(time.RFC3339),
		}
	}
	return out
}

func toProtoRoundSummaries(rounds []domain.Round) []*pbh.RoundSummary {
	out := make([]*pbh.RoundSummary, len(rounds))
	for i, r := range rounds {
		scores := make(map[string]int32, len(r.Scores))
		for team, score := range r.Scores {
			scores[team] = int32(score)
		}
		out[i] = &pbh.RoundSummary{
			RoundId:          r.ID,
			RoundNumber:      int32(r.RoundNumber),
			StartingPlayerId: r.StartingPlayerID,
			WinnerTeamId:     string(r.WinnerTeamID),
			Reason:           string(r.Reason),
			Scores:           scores,
		}
	}
	return out
}
