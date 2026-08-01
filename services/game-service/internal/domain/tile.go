package domain

import (
	"crypto/rand"
	"domino/shared/types"
	"math/big"
)

// FullSet generates the 28 unique tiles of a double-six domino set.
func FullSet() []types.Tile {
	tiles := make([]types.Tile, 0, 28)
	for i := 0; i <= 6; i++ {
		for j := i; j <= 6; j++ {
			tiles = append(tiles, types.Tile{Left: i, Right: j})
		}
	}
	return tiles
}

// shuffleTiles performs an in-place Fisher-Yates shuffle using crypto/rand.
func shuffleTiles(tiles []types.Tile) error {
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
