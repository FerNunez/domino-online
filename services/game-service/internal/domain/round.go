package domain

import (
	"domino/shared/types"

	"github.com/google/uuid"
)

type RoundStatus string

const (
	RoundStatusDealt      RoundStatus = "ROUND_STATUS_DEALT"
	RoundStatusInProgress RoundStatus = "ROUND_STATUS_IN_PROGRESS"
	RoundStatusRoundOver  RoundStatus = "ROUND_STATUS_OVER"
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
	LobbyID        string
	GameID         string
	ID             string
	RoundNumber    int
	Status         RoundStatus             // Dealt, in progress or over
	PlayerOrder    []string                // turn rotation order
	Hands          map[string][]types.Tile // map of PlayerId -> []Tiles{}
	Board          Board
	StartingPlayer string
	CurrentTurn    string
	PassStreak     uint
	ActionCount    int                // count of accepted plays/passes this round, used for history ordering
	Result         *types.RoundResult // nil while round in progress then set when gam efinished
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
	startingPlayer := previousStarter
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
		LobbyID:        lobbyID,
		GameID:         gameID,
		ID:             uuid.NewString(),
		Status:         RoundStatusDealt,
		PlayerOrder:    playerIDs,
		Hands:          hands,
		Board:          emptyBoard(),
		StartingPlayer: startingPlayer,
		CurrentTurn:    startingPlayer,
		PassStreak:     0,
		RoundNumber:    roundNumber,
	}, nil
}

func (r *RoundModel) nextTurn() string {
	for idx, playerID := range r.PlayerOrder {
		if playerID == r.CurrentTurn {
			return r.PlayerOrder[(idx+1)%len(r.PlayerOrder)]
		}
	}
	panic("current turn is not part of the player order")
}

// PlayTile validates and applies userID playing tile against the open end on side.
func (r *RoundModel) PlayTile(userID string, tile types.Tile, side types.Side) error {
	if r.Status == RoundStatusRoundOver {
		return ErrRoundOver
	}
	if r.CurrentTurn != userID {
		return ErrNotYourTurn
	}
	if side != SideLeft && side != SideRight {
		return ErrInvalidSide
	}

	// Check if player has the tile it wants to play
	hand := r.Hands[userID]
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
	if r.Board.IsEmpty() {
		//	Place first tile
		r.Board.Tiles = append(r.Board.Tiles, placed)
		r.Board.LeftEnd = placed.Left
		r.Board.RightEnd = placed.Right
	} else if side == SideLeft {
		// Check if can be placed left
		if !placed.Has(r.Board.LeftEnd) {
			return ErrIllegalMove
		}
		// Check if needs to flip
		oriented := placed
		if oriented.Right != r.Board.LeftEnd {
			oriented = oriented.Flip()
		}
		r.Board.Tiles = append([]types.Tile{oriented}, r.Board.Tiles...)
		r.Board.LeftEnd = oriented.Left
	} else {
		// Check if can be placed left
		if !placed.Has(r.Board.RightEnd) {
			return ErrIllegalMove
		}
		// Check if needs to flip
		oriented := placed
		if oriented.Left != r.Board.RightEnd {
			oriented = oriented.Flip()
		}
		r.Board.Tiles = append(r.Board.Tiles, oriented)
		r.Board.RightEnd = oriented.Right
	}

	// remove user place tile
	r.Hands[userID] = append(hand[:handIdx], hand[handIdx+1:]...)
	// reset pass counter for block state
	r.PassStreak = 0
	r.ActionCount++

	// Check if current player won by placing
	if len(r.Hands[userID]) == 0 {
		r.Status = RoundStatusRoundOver
		result := r.resolveRound(types.ReasonDomino)
		r.Result = &result
		return nil
	}

	r.Status = RoundStatusInProgress
	r.CurrentTurn = r.nextTurn()
	return nil
}

// PassTurn validates that userID genuinely has no legal move and advances the turn.
// A pass streak covering every player means the board is blocked and the round ends.
func (r *RoundModel) PassTurn(userID string) error {
	if r.Status == RoundStatusRoundOver {
		return ErrRoundOver
	}
	if r.CurrentTurn != userID {
		return ErrNotYourTurn
	}
	// Check incorrect passing!! mmeaning no valid moves in its hands
	if len(ValidMoves(r.Hands[userID], r.Board)) > 0 {
		return ErrHasLegalMove
	}

	// keep memory of how many consecutives players passed, if reaches 4 => game over
	r.ActionCount++
	r.PassStreak++
	if int(r.PassStreak) >= len(r.PlayerOrder) {
		r.Status = RoundStatusRoundOver
		result := r.resolveRound(types.ReasonBlocked)
		r.Result = &result
		return nil
	}

	r.Status = RoundStatusInProgress
	r.CurrentTurn = r.nextTurn()
	return nil
}

// resolveRound computes the result of round: winner, score and sets reason
func (r *RoundModel) resolveRound(reason types.Reason) types.RoundResult {
	// Compute pip count by team
	pipCountByTeam := make(map[types.TeamID]int)
	for idx, playerID := range r.PlayerOrder {
		hand := r.Hands[playerID]
		slot := idx + 1 // slot goes from 1
		teamID := types.SlotToTeamID(slot)
		pipCountByTeam[teamID] += sumTiles(hand)
	}

	switch reason {
	case types.ReasonDomino:
		// find winner = user with hand len = 0
		var winnerTeamID types.TeamID
		for idx, playerID := range r.PlayerOrder {
			hand := r.Hands[playerID]
			slot := idx + 1 // slot goes from 1
			if len(hand) == 0 {
				winnerTeamID = types.SlotToTeamID(slot)
				break
			}
		}
		return types.RoundResult{
			WinnerTeamID: winnerTeamID,
			Reason:       types.ReasonDomino,
			PipCounts:    pipCountByTeam,
		}

	case types.ReasonBlocked:
		// add score by team
		// Draw:
		if pipCountByTeam[types.TeamA] == pipCountByTeam[types.TeamB] {
			return types.RoundResult{
				Reason:       types.ReasonBlocked,
				WinnerTeamID: "",
				PipCounts:    pipCountByTeam,
			}
		}

		// Lowest pip count wins a blocked round.
		var winnerID types.TeamID
		if pipCountByTeam[types.TeamA] < pipCountByTeam[types.TeamB] {
			winnerID = types.TeamA
		} else {
			winnerID = types.TeamB
		}
		return types.RoundResult{
			WinnerTeamID: winnerID,
			Reason:       types.ReasonBlocked,
			PipCounts:    pipCountByTeam,
		}
	default:
		panic("resolveResult called before the round ended")
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

// SumTiles adds all pipes of a hand
func sumTiles(hand []Tile) int {
	total := 0
	for _, t := range hand {
		total += t.Pips()
	}
	return total
}
