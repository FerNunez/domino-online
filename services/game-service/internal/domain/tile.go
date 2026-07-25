package domain

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

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

// FullSet generates the 28 unique tiles of a double-six domino set.
func FullSet() []Tile {
	tiles := make([]Tile, 0, 28)
	for i := 0; i <= 6; i++ {
		for j := i; j <= 6; j++ {
			tiles = append(tiles, Tile{Left: i, Right: j})
		}
	}
	return tiles
}

// shuffleTiles performs an in-place Fisher-Yates shuffle using crypto/rand.
func shuffleTiles(tiles []Tile) error {
	for i := len(tiles) - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return err
		}
		j := jBig.Int64()
		tiles[i], tiles[j] = tiles[j], tiles[i]
	}
	return nil
}
