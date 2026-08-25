package domain

import (
	"context"
	"errors"
	"time"

	"domino/shared/types"
)

// ErrRoundNotFound means a round's history isn't fully readable yet for a
// given round ID,  either no `rounds` row exists at all, or it exists but
// its actions/hands haven't all landed yet
var ErrRoundNotFound = errors.New("round not found")

// Game is the final record of a finished match.
type Game struct {
	ID          string
	LobbyID     string
	FinalScores map[string]int
	TeamWinner  string
	GameState   string
	CreatedAt   time.Time
}

// Round within a game.
type Round struct {
	ID               string
	GameID           string
	RoundNumber      int
	StartingPlayerID string
	PlayerOrder      []string
	WinnerTeamID     types.TeamID
	Reason           types.Reason
	Scores           map[string]int
	ActionCount      int // total accepted moves produced
}

// Action is one accepted play or pass within a round, in the order it happened.
// TileLeft/TileRight/Side/ResultingLeftEnd/ResultingRightEnd are nil for a pass action.
type Action struct {
	RoundID           string
	ActionNumber      int
	PlayerID          string
	ActionType        types.ActionType
	TileLeft          *int
	TileRight         *int
	Side              types.Side
	ResultingLeftEnd  *int
	ResultingRightEnd *int
}

// Hand is the tiles dealt to a player at the start of a round.
// Combined with the round's Actions, this is enough to fully replay the round.
type Hand struct {
	RoundID  string
	PlayerID string
	Tiles    []types.Tile
}

// GamePlayer links a player to a game they took part in, so a players
// match history can be looked up without scanning rounds.
type GamePlayer struct {
	GameID   string
	PlayerID string
	TeamID   types.TeamID
}

// HistoryRepository persists and reads durable game history. Writes are
// idempotent (safe to reprocess a redelivered RabbitMQ message).
type HistoryRepository interface {
	UpsertGame(ctx context.Context, game Game) error
	UpsertRound(ctx context.Context, round Round) error
	InsertAction(ctx context.Context, action Action) error
	UpsertHand(ctx context.Context, hand Hand) error
	UpsertGamePlayer(ctx context.Context, gamePlayer GamePlayer) error
	GetActionsByRoundID(ctx context.Context, roundID string) ([]Action, error)
	GetHandsByRoundID(ctx context.Context, roundID string) ([]Hand, error)
	GetRoundsByGameID(ctx context.Context, gameID string) ([]Round, error)
	GetRoundByID(ctx context.Context, roundID string) (Round, error)
	GetGamesByPlayerID(ctx context.Context, playerID string, limit, offset int) ([]Game, error)
}
