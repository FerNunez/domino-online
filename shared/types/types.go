package types

type ActionType string

const (
	Play ActionType = "play"
	Pass ActionType = "pass"
)

type Reason string

const (
	ReasonDomino  Reason = "REASON_DOMINO"
	ReasonBlocked Reason = "REASON_BLOCKED"
)

// RoundResult is the outcome of a finished round (hand emptied or blocked board).
type RoundResult struct {
	WinnerTeamID TeamID         `json:"winnerTeamID"`
	Reason       Reason         `json:"reason"` // ReasonDomino | ReasonBlocked
	Scores       map[TeamID]int `json:"scores"`
}

// Tile is a domino tile identified by its two values.
type Tile struct {
	Left  int `json:"left"`
	Right int `json:"right"`
}

func (t Tile) IsDouble() bool {
	return t.Left == t.Right
}

// Flip returns the tile with its pips swapped.
func (t Tile) Flip() Tile {
	return Tile{Left: t.Right, Right: t.Left}
}

// Has reports whether pip is one of the tile's two values.
func (t Tile) Has(pip int) bool {
	return t.Left == pip || t.Right == pip
}

func (t Tile) Pips() int {
	return t.Left + t.Right
}

// Side is which open end of the board a tile is played against.
type Side string

const (
	Left  Side = "left"
	Right Side = "right"
)

type TeamID string

const (
	TeamA TeamID = "TEAM_A"
	TeamB TeamID = "TEAM_B"
)

// SlotToTeamID Define player team: 1,3 vs 2,4
func SlotToTeamID(slot int) TeamID {
	switch slot {
	case 1, 3:
		return TeamA
	case 2, 4:
		return TeamB
	default:
		panic("slot should be between 1 to 4")

	}
}
