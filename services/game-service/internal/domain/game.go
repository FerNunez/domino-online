package domain

import (
	"context"
	"errors"
	"fmt"
)

type GameStatus string

const (
	GameStatusDealt      GameStatus = "GAME_STATUS_DEALT"
	GameStatusInProgress GameStatus = "GAME_STATUS_IN_PROGRESS"
	GameStatusRoundOver  GameStatus = "GAME_STATUS_ROUND_OVER"
)

const (
	SideLeft  = "left"
	SideRight = "right"
)

const (
	ReasonDomino  = "domino"
	ReasonBlocked = "blocked"
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

// noEnd is the sentinel value for Board.LeftEnd/RightEnd when the board is empty.
const noEnd = -1

// Board tracks the tiles played, in play order, and the two pip values
// currently open for matching.
type Board struct {
	Tiles    []Tile
	LeftEnd  int
	RightEnd int
}

func emptyBoard() Board {
	return Board{Tiles: []Tile{}, LeftEnd: noEnd, RightEnd: noEnd}
}

func (b Board) IsEmpty() bool {
	return len(b.Tiles) == 0
}

// Move describes a candidate legal play: a tile from hand and which open end it matches.
type Move struct {
	Tile Tile
	Side string
}

// RoundResult is the outcome of a finished round (hand emptied or blocked board).
type RoundResult struct {
	WinnerID string
	Reason   string // ReasonDomino | ReasonBlocked
	Scores   map[string]int
}

type GameModel struct {
	LobbyID     string
	Status      GameStatus
	PlayerOrder []string          // turn rotation order
	Hands       map[string][]Tile // map of PlayerId -> []Tiles{}
	Board       Board
	CurrentTurn string
	PassStreak  int

	// roundOverReason is set internally by PlayTile/PassTurn when they conclude
	// the round, so ResolveRoundResult knows which outcome to compute.
	roundOverReason string
}

type GameRepository interface {
	CreateGame(ctx context.Context, g *GameModel) (*GameModel, error)
	GetGameByLobbyID(ctx context.Context, lobbyID string) (*GameModel, error)
	UpdateGame(ctx context.Context, lobbyID string, g *GameModel) (*GameModel, error)
}

type GameService interface {
	CreateGame(ctx context.Context, lobbyID string, playerIDs []string) (*GameModel, error)
	PlayTile(ctx context.Context, lobbyID, userID string, tile Tile, side string) (*GameModel, *RoundResult, error)
	PassTurn(ctx context.Context, lobbyID, userID string) (*GameModel, *RoundResult, error)
}

// NewGame builds a fresh standard-block domino game: shuffles a double-six
// set, deals 7 tiles to each of the 4 players, and sets the starting player
// to whoever holds the 6-6 double (guaranteed present since all 28 tiles are dealt).
func NewGame(lobbyID string, playerIDs []string) (*GameModel, error) {
	if len(playerIDs) != 4 {
		return nil, ErrWrongPlayerCnt
	}

	pile := FullSet()
	if err := shuffleTiles(pile); err != nil {
		return nil, fmt.Errorf("couldn't shuffle tiles: %w", err)
	}

	const handSize = 7
	hands := make(map[string][]Tile, len(playerIDs))
	for idx, playerID := range playerIDs {
		hands[playerID] = pile[idx*handSize : (idx+1)*handSize]
	}

	startingPlayer := ""
	for playerID, hand := range hands {
		for _, t := range hand {
			if t == (Tile{Left: 6, Right: 6}) {
				startingPlayer = playerID
			}
		}
	}
	if startingPlayer == "" {
		panic("expected a player to hold the 6-6 double")
	}

	return &GameModel{
		LobbyID:     lobbyID,
		Status:      GameStatusDealt,
		PlayerOrder: playerIDs,
		Hands:       hands,
		Board:       emptyBoard(),
		CurrentTurn: startingPlayer,
	}, nil
}

// sameTile compares tiles ignoring orientation (2-5 == 5-2).
func sameTile(a, b Tile) bool {
	return a == b || a == b.Flip()
}

// ValidMoves returns every legal (tile, side) combination playable from hand
// given the current board state.
func ValidMoves(hand []Tile, b Board) []Move {
	moves := make([]Move, 0)
	if b.IsEmpty() {
		for _, t := range hand {
			moves = append(moves, Move{Tile: t, Side: SideLeft})
		}
		return moves
	}

	for _, t := range hand {
		if t.Has(b.LeftEnd) {
			moves = append(moves, Move{Tile: t, Side: SideLeft})
		}
		if t.Has(b.RightEnd) {
			moves = append(moves, Move{Tile: t, Side: SideRight})
		}
	}
	return moves
}

func (g *GameModel) nextTurn() string {
	for idx, playerID := range g.PlayerOrder {
		if playerID == g.CurrentTurn {
			return g.PlayerOrder[(idx+1)%len(g.PlayerOrder)]
		}
	}
	panic("current turn is not part of the player order")
}

// PlayTile validates and applies userID playing tile against the open end on side.
func (g *GameModel) PlayTile(userID string, tile Tile, side string) error {
	if g.Status == GameStatusRoundOver {
		return ErrRoundOver
	}
	if g.CurrentTurn != userID {
		return ErrNotYourTurn
	}
	if side != SideLeft && side != SideRight {
		return ErrInvalidSide
	}

	// Check if player has the tile it wants to play
	hand := g.Hands[userID]
	handIdx := -1
	for idx, t := range hand {
		if sameTile(t, tile) {
			handIdx = idx
			break
		}
	}
	if handIdx == -1 {
		return ErrTileNotInHand
	}

	// Check if move is approuved
	placed := hand[handIdx]
	if g.Board.IsEmpty() {
		//	Place first tile
		g.Board.Tiles = append(g.Board.Tiles, placed)
		g.Board.LeftEnd = placed.Left
		g.Board.RightEnd = placed.Right
	} else if side == SideLeft {
		// Check if can be placed left
		if !placed.Has(g.Board.LeftEnd) {
			return ErrIllegalMove
		}
		// Check if needs to flip
		oriented := placed
		if oriented.Right != g.Board.LeftEnd {
			oriented = oriented.Flip()
		}
		g.Board.Tiles = append([]Tile{oriented}, g.Board.Tiles...)
		g.Board.LeftEnd = oriented.Left
	} else {
		// Check if can be placed left
		if !placed.Has(g.Board.RightEnd) {
			return ErrIllegalMove
		}
		// Check if needs to flip
		oriented := placed
		if oriented.Left != g.Board.RightEnd {
			oriented = oriented.Flip()
		}
		g.Board.Tiles = append(g.Board.Tiles, oriented)
		g.Board.RightEnd = oriented.Right
	}

	// remove user place tile
	g.Hands[userID] = append(hand[:handIdx], hand[handIdx+1:]...)
	// reset pass counter for block state
	g.PassStreak = 0

	// Check if current player won by placing
	if len(g.Hands[userID]) == 0 {
		g.Status = GameStatusRoundOver
		g.roundOverReason = ReasonDomino
		return nil
	}

	g.Status = GameStatusInProgress
	g.CurrentTurn = g.nextTurn()
	return nil
}

// PassTurn validates that userID genuinely has no legal move and advances the turn.
// A pass streak covering every player means the board is blocked and the round ends.
func (g *GameModel) PassTurn(userID string) error {
	if g.Status == GameStatusRoundOver {
		return ErrRoundOver
	}
	if g.CurrentTurn != userID {
		return ErrNotYourTurn
	}
	// Check incorrect passing!! mmeaning no valid moves in its hands
	if len(ValidMoves(g.Hands[userID], g.Board)) > 0 {
		return ErrHasLegalMove
	}

	// keep memory of how many consecutives players passed, if reaches 4 => game over
	g.PassStreak++
	if g.PassStreak >= len(g.PlayerOrder) {
		g.Status = GameStatusRoundOver
		g.roundOverReason = ReasonBlocked
		return nil
	}

	g.Status = GameStatusInProgress
	g.CurrentTurn = g.nextTurn()
	return nil
}

// TODO: Change to proper 2 v 2 scoring system
// ResolveRoundResult computes the winner and score once the round has ended.
// It must only be called when Status == GameStatusRoundOver.
func (g *GameModel) ResolveRoundResult() RoundResult {
	pipSum := func(hand []Tile) int {
		total := 0
		for _, t := range hand {
			total += t.Pips()
		}
		return total
	}

	switch g.roundOverReason {
	case ReasonDomino:
		winnerID := ""
		for playerID, hand := range g.Hands {
			if len(hand) == 0 {
				winnerID = playerID
				break
			}
		}
		scores := make(map[string]int)
		for playerID, hand := range g.Hands {
			if playerID == winnerID {
				continue
			}
			scores[playerID] = pipSum(hand)
		}
		return RoundResult{WinnerID: winnerID, Reason: ReasonDomino, Scores: scores}

	case ReasonBlocked:
		lowest := -1
		winnerID := ""
		tie := false
		for playerID, hand := range g.Hands {
			total := pipSum(hand)
			switch {
			case lowest == -1 || total < lowest:
				lowest = total
				winnerID = playerID
				tie = false
			case total == lowest:
				tie = true
			}
		}
		if tie {
			return RoundResult{Reason: ReasonBlocked, Scores: map[string]int{}}
		}
		scores := make(map[string]int)
		for playerID, hand := range g.Hands {
			if playerID == winnerID {
				continue
			}
			scores[playerID] = pipSum(hand)
		}
		return RoundResult{WinnerID: winnerID, Reason: ReasonBlocked, Scores: scores}

	default:
		panic("ResolveRoundResult called before the round ended")
	}
}
