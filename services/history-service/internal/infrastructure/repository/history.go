package repository

import (
	"context"
	"encoding/json"
	"errors"

	"domino/services/history-service/internal/domain"
	"domino/shared/db/sql"
	"domino/shared/types"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type SQLRepository struct {
	queries *sql.Queries
}

func NewSQLRepository(queries *sql.Queries) *SQLRepository {
	return &SQLRepository{queries: queries}
}

func (s *SQLRepository) UpsertGame(ctx context.Context, game domain.Game) error {
	gameUUID, err := uuid.Parse(game.ID)
	if err != nil {
		return err
	}
	lobbyUUID, err := uuid.Parse(game.LobbyID)
	if err != nil {
		return err
	}
	finalScores, err := json.Marshal(game.FinalScores)
	if err != nil {
		return err
	}

	return s.queries.UpsertGame(ctx, sql.UpsertGameParams{
		GameID:      pgtype.UUID{Bytes: gameUUID, Valid: true},
		LobbyID:     pgtype.UUID{Bytes: lobbyUUID, Valid: true},
		FinalScores: finalScores,
		TeamWinner:  game.TeamWinner,
		GameState:   game.GameState,
	})
}

func (s *SQLRepository) UpsertRound(ctx context.Context, round domain.Round) error {
	roundUUID, err := uuid.Parse(round.ID)
	if err != nil {
		return err
	}
	gameUUID, err := uuid.Parse(round.GameID)
	if err != nil {
		return err
	}
	scores, err := json.Marshal(round.Scores)
	if err != nil {
		return err
	}

	return s.queries.UpsertRound(ctx, sql.UpsertRoundParams{
		RoundID:          pgtype.UUID{Bytes: roundUUID, Valid: true},
		GameID:           pgtype.UUID{Bytes: gameUUID, Valid: true},
		RoundNumber:      int32(round.RoundNumber),
		StartingPlayerID: round.StartingPlayerID,
		PlayerOrder:      round.PlayerOrder,
		WinnerTeamID:     string(round.WinnerTeamID),
		Reason:           string(round.Reason),
		Scores:           scores,
		ActionCount:      int32(round.ActionCount),
	})
}

func (s *SQLRepository) InsertAction(ctx context.Context, action domain.Action) error {
	roundUUID, err := uuid.Parse(action.RoundID)
	if err != nil {
		return err
	}

	return s.queries.InsertAction(ctx, sql.InsertActionParams{
		RoundID:           pgtype.UUID{Bytes: roundUUID, Valid: true},
		ActionNumber:      int32(action.ActionNumber),
		PlayerID:          action.PlayerID,
		ActionType:        string(action.ActionType),
		TileLeft:          intPtrToPgInt4(action.TileLeft),
		TileRight:         intPtrToPgInt4(action.TileRight),
		Side:              stringToPgText(string(action.Side)),
		ResultingLeftEnd:  intPtrToPgInt4(action.ResultingLeftEnd),
		ResultingRightEnd: intPtrToPgInt4(action.ResultingRightEnd),
	})
}

func (s *SQLRepository) GetActionsByRoundID(ctx context.Context, roundID string) ([]domain.Action, error) {
	roundUUID, err := uuid.Parse(roundID)
	if err != nil {
		return nil, err
	}

	rows, err := s.queries.GetActionsByRoundID(ctx, pgtype.UUID{Bytes: roundUUID, Valid: true})
	if err != nil {
		return nil, err
	}

	actions := make([]domain.Action, len(rows))
	for i, row := range rows {
		actions[i] = domain.Action{
			RoundID:           uuid.UUID(row.RoundID.Bytes).String(),
			ActionNumber:      int(row.ActionNumber),
			PlayerID:          row.PlayerID,
			ActionType:        types.ActionType(row.ActionType),
			TileLeft:          pgInt4ToIntPtr(row.TileLeft),
			TileRight:         pgInt4ToIntPtr(row.TileRight),
			Side:              types.Side(pgTextToString(row.Side)),
			ResultingLeftEnd:  pgInt4ToIntPtr(row.ResultingLeftEnd),
			ResultingRightEnd: pgInt4ToIntPtr(row.ResultingRightEnd),
		}
	}
	return actions, nil
}

func (s *SQLRepository) UpsertHand(ctx context.Context, hand domain.Hand) error {
	roundUUID, err := uuid.Parse(hand.RoundID)
	if err != nil {
		return err
	}
	tiles, err := json.Marshal(hand.Tiles)
	if err != nil {
		return err
	}

	return s.queries.UpsertHand(ctx, sql.UpsertHandParams{
		RoundID:  pgtype.UUID{Bytes: roundUUID, Valid: true},
		PlayerID: hand.PlayerID,
		Tiles:    tiles,
	})
}

func (s *SQLRepository) UpsertGamePlayer(ctx context.Context, gamePlayer domain.GamePlayer) error {
	gameUUID, err := uuid.Parse(gamePlayer.GameID)
	if err != nil {
		return err
	}

	return s.queries.UpsertGamePlayer(ctx, sql.UpsertGamePlayerParams{
		GameID:   pgtype.UUID{Bytes: gameUUID, Valid: true},
		PlayerID: gamePlayer.PlayerID,
		TeamID:   string(gamePlayer.TeamID),
	})
}

func (s *SQLRepository) GetHandsByRoundID(ctx context.Context, roundID string) ([]domain.Hand, error) {
	roundUUID, err := uuid.Parse(roundID)
	if err != nil {
		return nil, err
	}

	rows, err := s.queries.GetHandsByRoundID(ctx, pgtype.UUID{Bytes: roundUUID, Valid: true})
	if err != nil {
		return nil, err
	}

	hands := make([]domain.Hand, len(rows))
	for i, row := range rows {
		var tiles []types.Tile
		if err := json.Unmarshal(row.Tiles, &tiles); err != nil {
			return nil, err
		}
		hands[i] = domain.Hand{
			RoundID:  uuid.UUID(row.RoundID.Bytes).String(),
			PlayerID: row.PlayerID,
			Tiles:    tiles,
		}
	}
	return hands, nil
}

func (s *SQLRepository) GetGamesByPlayerID(ctx context.Context, playerID string, limit, offset int) ([]domain.Game, error) {
	rows, err := s.queries.GetGamesByPlayerID(ctx, sql.GetGamesByPlayerIDParams{
		PlayerID: playerID,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		return nil, err
	}

	games := make([]domain.Game, len(rows))
	for i, row := range rows {
		var finalScores map[string]int
		if err := json.Unmarshal(row.FinalScores, &finalScores); err != nil {
			return nil, err
		}
		games[i] = domain.Game{
			ID:          uuid.UUID(row.GameID.Bytes).String(),
			LobbyID:     uuid.UUID(row.LobbyID.Bytes).String(),
			FinalScores: finalScores,
			TeamWinner:  row.TeamWinner,
			GameState:   row.GameState,
			CreatedAt:   row.CreatedAt.Time,
		}
	}
	return games, nil
}

func (s *SQLRepository) GetRoundsByGameID(ctx context.Context, gameID string) ([]domain.Round, error) {
	gameUUID, err := uuid.Parse(gameID)
	if err != nil {
		return nil, err
	}

	rows, err := s.queries.GetRoundsByGameID(ctx, pgtype.UUID{Bytes: gameUUID, Valid: true})
	if err != nil {
		return nil, err
	}

	rounds := make([]domain.Round, len(rows))
	for i, row := range rows {
		var scores map[string]int
		if err := json.Unmarshal(row.Scores, &scores); err != nil {
			return nil, err
		}
		rounds[i] = domain.Round{
			ID:               uuid.UUID(row.RoundID.Bytes).String(),
			GameID:           uuid.UUID(row.GameID.Bytes).String(),
			RoundNumber:      int(row.RoundNumber),
			StartingPlayerID: row.StartingPlayerID,
			PlayerOrder:      row.PlayerOrder,
			WinnerTeamID:     types.TeamID(row.WinnerTeamID),
			Reason:           types.Reason(row.Reason),
			Scores:           scores,
			ActionCount:      int(row.ActionCount),
		}
	}
	return rounds, nil
}

func (s *SQLRepository) GetRoundByID(ctx context.Context, roundID string) (domain.Round, error) {
	roundUUID, err := uuid.Parse(roundID)
	if err != nil {
		return domain.Round{}, err
	}

	row, err := s.queries.GetRoundByID(ctx, pgtype.UUID{Bytes: roundUUID, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Round{}, domain.ErrRoundNotFound
		}
		return domain.Round{}, err
	}

	var scores map[string]int
	if err := json.Unmarshal(row.Scores, &scores); err != nil {
		return domain.Round{}, err
	}
	return domain.Round{
		ID:               uuid.UUID(row.RoundID.Bytes).String(),
		GameID:           uuid.UUID(row.GameID.Bytes).String(),
		RoundNumber:      int(row.RoundNumber),
		StartingPlayerID: row.StartingPlayerID,
		PlayerOrder:      row.PlayerOrder,
		WinnerTeamID:     types.TeamID(row.WinnerTeamID),
		Reason:           types.Reason(row.Reason),
		Scores:           scores,
		ActionCount:      int(row.ActionCount),
	}, nil
}

func intPtrToPgInt4(i *int) pgtype.Int4 {
	if i == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: int32(*i), Valid: true}
}

func pgInt4ToIntPtr(v pgtype.Int4) *int {
	if !v.Valid {
		return nil
	}
	i := int(v.Int32)
	return &i
}

func stringToPgText(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func pgTextToString(v pgtype.Text) string {
	if !v.Valid {
		return ""
	}
	return v.String
}
