package types

import "fmt"

// RoundResult is the outcome of a finished round (hand emptied or blocked board).
type RoundResult struct {
	WinnerID string
	Reason   string // ReasonDomino | ReasonBlocked
	Scores   map[string]int
}

// Tile is a domino tile identified by its two values.
type Tile struct {
	Left  int
	Right int
}

func (t Tile) String() string {
	return fmt.Sprintf("%d-%d", t.Left, t.Right)
}

// ParseTile parses the tuple format "L-R" back into a Tile.
func ParseTile(s string) (Tile, error) {
	var l, r int
	if _, err := fmt.Sscanf(s, "%d-%d", &l, &r); err != nil {
		return Tile{}, fmt.Errorf("invalid tile %q: %w", s, err)
	}
	return Tile{Left: l, Right: r}, nil
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
