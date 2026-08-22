package domain

import (
	"context"
	"domino/shared/types"
	"errors"
)

// alias so domain code say Tile instead of types.Tile
type Tile = types.Tile

type GameStatus string

const (
	GameStatusInProgress GameStatus = "GAME_STATUS_IN_PROGRESS"
	GameStatusGameOver   GameStatus = "GAME_STATUS_FINISHED"
)

const (
	SideLeft  = types.Left
	SideRight = types.Right
)

var (
	ErrNotYourTurn    = errors.New("not your turn")
	ErrTileNotInHand  = errors.New("tile not in hand")
	ErrIllegalMove    = errors.New("illegal move: tile does not match the open end")
	ErrInvalidSide    = errors.New("side must be \"left\" or \"right\"")
	ErrRoundOver      = errors.New("round is already over")
	ErrHasLegalMove   = errors.New("cannot pass: a legal move exists")
	ErrWrongPlayerCnt = errors.New("standard block domino requires exactly 4 players")
)

type GameModel struct {
	LobbyID      string
	ID           string // GameID: uuid, unique across all games ever played
	GameNumber   int    // 1, 2, 3... sequential within LobbyID, for ordering/display
	Status       GameStatus
	TeamScores   map[types.TeamID]int
	CurrentRound *RoundModel
	GoalScore    int // MaxGame
}

// TODO: Add this as config sent from LobbyService
// GoalScore is the default team score a match plays to.
const GoalScore = 100

// NewGame: creates game with a first round, and 0 scores
func NewGame(lobbyID, gameID string, gameNumber int, playerIDs []string) (*GameModel, error) {
	if len(playerIDs) != 4 {
		return nil, ErrWrongPlayerCnt
	}

	// crete first round
	firstRound, err := NewRound(lobbyID, gameID, 1, playerIDs, "")
	if err != nil {
		return nil, err
	}

	return &GameModel{
		LobbyID:      lobbyID,
		ID:           gameID,
		GameNumber:   gameNumber,
		Status:       GameStatusInProgress,
		TeamScores:   map[types.TeamID]int{types.TeamA: 0, types.TeamB: 0},
		CurrentRound: firstRound,
		GoalScore:    GoalScore,
	}, nil
}

type GameService interface {
	CreateGameWithID(ctx context.Context, gameID, lobbyID string, playerIDs []string) (*GameModel, error)
	PlayTile(ctx context.Context, lobbyID, userID string, tile types.Tile, side types.Side) (*RoundModel, *types.RoundResult, error)
	PassTurn(ctx context.Context, lobbyID, userID string) (*RoundModel, *types.RoundResult, error)
	// NextRound resolves the finished current round, applies its score to the
	// match, and deals the next round (or ends the match if GoalScore was reached).
	NextRound(ctx context.Context, lobbyID string) (*RoundModel, *types.RoundResult, error)
}

type GameRepository interface {
	CreateGame(ctx context.Context, g *GameModel) (*GameModel, error)
	// GetCurrentGame fetches the lobby's in-progress game (the normal path for gameplay RPCs).
	GetCurrentGame(ctx context.Context, lobbyID string) (*GameModel, error)
	// GetGameByID fetches any game ever played, by its uuid, e.g. for history/replay.
	GetGameByID(ctx context.Context, gameID string) (*GameModel, error)
	UpdateGame(ctx context.Context, g *GameModel) (*GameModel, error)
	// NextGameNumber atomically reserves the next sequential game number for lobbyID.
	NextGameNumber(ctx context.Context, lobbyID string) (int, error)
}
