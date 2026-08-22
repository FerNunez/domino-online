package domain

import "domino/shared/types"

type RoundStatus string

const (
	RoundStatusDealt      RoundStatus = "ROUND_STATUS_DEALT"
	RoundStatusInProgress RoundStatus = "ROUND_STATUS_IN_PROGRESS"
	RoundStatusRoundOver  RoundStatus = "ROUND_STATUS_ROUND_OVER"
)

// noEnd is the sentinel value for Board.LeftEnd/RightEnd when the board is empty.
const noEnd = -1

// Board tracks the tiles played, in play order, and the two pip values
// currently open for matching.
type Board struct {
	Tiles    []types.Tile
	LeftEnd  int
	RightEnd int
}

func emptyBoard() Board {
	return Board{Tiles: []types.Tile{}, LeftEnd: noEnd, RightEnd: noEnd}
}

func (b Board) IsEmpty() bool {
	return len(b.Tiles) == 0
}

type RoundModel struct {
	LobbyID     string
	GameID      string
	ID          int // RoundID. From 1, sequential within a game.
	Status      RoundStatus
	PlayerOrder []string                // turn rotation order
	Hands       map[string][]types.Tile // map of PlayerId -> []Tiles{}
	Board       Board
	CurrentTurn string
	PassStreak  uint

	// roundOverReason is set internally by PlayTile/PassTurn when they conclude
	// the round, so ResolveResult knows which outcome to compute.
	roundOverReason types.Reason
}

// Move describes a candidate legal play: a tile from hand and which open end it matches.
type Move struct {
	Tile types.Tile
	Side types.Side
}

// NewRound deals the next round of an in-progress game: fresh hands, empty
// board, and the seat after previousStarter leading (simple dealer rotation).
func NewRound(lobbyID, gameID string, roundNumber int, playerIDs []string, previousStarter string) (*RoundModel, error) {
	if len(playerIDs) != 4 {
		return nil, ErrWrongPlayerCnt
	}

	hands, err := dealHands(playerIDs)
	if err != nil {
		return nil, err
	}

	// Find starting player, if not previous starter => first game
	startingPlayer := playerIDs[0]
	if previousStarter == "" || roundNumber == 1 {
		startingPlayer = holderOfDoubleSix(hands)
	} else {
		// not first game => keep rotating to next player
		for idx, playerID := range playerIDs {
			if playerID == previousStarter {
				startingPlayer = playerIDs[(idx+1)%len(playerIDs)]
				break
			}
		}
	}

	return &RoundModel{
		LobbyID:     lobbyID,
		GameID:      gameID,
		ID:          roundNumber,
		Status:      RoundStatusDealt,
		PlayerOrder: playerIDs,
		Hands:       hands,
		Board:       emptyBoard(),
		CurrentTurn: startingPlayer,
	}, nil
}

func (g *RoundModel) nextTurn() string {
	for idx, playerID := range g.PlayerOrder {
		if playerID == g.CurrentTurn {
			return g.PlayerOrder[(idx+1)%len(g.PlayerOrder)]
		}
	}
	panic("current turn is not part of the player order")
}

// PlayTile validates and applies userID playing tile against the open end on side.
func (g *RoundModel) PlayTile(userID string, tile types.Tile, side types.Side) error {
	if g.Status == RoundStatusRoundOver {
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
		g.Board.Tiles = append([]types.Tile{oriented}, g.Board.Tiles...)
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
		g.Status = RoundStatusRoundOver
		g.roundOverReason = types.ReasonDomino
		return nil
	}

	g.Status = RoundStatusInProgress
	g.CurrentTurn = g.nextTurn()
	return nil
}

// PassTurn validates that userID genuinely has no legal move and advances the turn.
// A pass streak covering every player means the board is blocked and the round ends.
func (g *RoundModel) PassTurn(userID string) error {
	if g.Status == RoundStatusRoundOver {
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
	if int(g.PassStreak) >= len(g.PlayerOrder) {
		g.Status = RoundStatusRoundOver
		g.roundOverReason = types.ReasonBlocked
		return nil
	}

	g.Status = RoundStatusInProgress
	g.CurrentTurn = g.nextTurn()
	return nil
}

// TODO: Change to proper 2 v 2 scoring system
// ResolveRoundResult computes the winner and score once the round has ended.
// It must only be called when Status == GameStatusRoundOver.
func (g *RoundModel) ResolveResult() types.RoundResult {
	pipSum := func(hand []types.Tile) int {
		total := 0
		for _, t := range hand {
			total += t.Pips()
		}
		return total
	}

	switch g.roundOverReason {
	case types.ReasonDomino:
		var winnerTeamID types.TeamID

		for idx, playerID := range g.PlayerOrder {
			hand := g.Hands[playerID]
			slot := idx + 1 // slot goes from 1
			if len(hand) == 0 {
				winnerTeamID = types.SlotToTeamID(slot)
				break
			}
		}
		scores := make(map[types.TeamID]int)

		for idx, playerID := range g.PlayerOrder {
			hand := g.Hands[playerID]
			slot := idx + 1 // slot goes from 1
			if types.SlotToTeamID(slot) == winnerTeamID {
				continue
			}
			scores[winnerTeamID] += pipSum(hand)
		}

		return types.RoundResult{WinnerTeamID: winnerTeamID, Reason: types.ReasonDomino, Scores: scores}

	case types.ReasonBlocked:
		scores := make(map[types.TeamID]int)
		for idx, playerID := range g.PlayerOrder {
			hand := g.Hands[playerID]
			slot := idx + 1 // slot goes from 1
			teamID := types.SlotToTeamID(slot)
			scores[teamID] += pipSum(hand)
		}
		if scores[types.TeamA] == scores[types.TeamB] {
			return types.RoundResult{Reason: types.ReasonBlocked, Scores: map[types.TeamID]int{}}
		}
		// Lowest pip count wins a blocked round.
		var winnerID types.TeamID
		if scores[types.TeamA] < scores[types.TeamB] {
			winnerID = types.TeamA
		} else {
			winnerID = types.TeamB
		}
		return types.RoundResult{WinnerTeamID: winnerID, Reason: types.ReasonBlocked, Scores: scores}

	default:
		panic("ResolveResult called before the round ended")
	}
}

// ValidMoves returns every legal (tile, side) combination playable from hand
// given the current board state.
func ValidMoves(hand []types.Tile, b Board) []Move {
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
