package domain

import (
	"crypto/rand"
	"domino/shared/types"
	"fmt"
	"math/big"
	"slices"
)

// dealHands shuffles&deal a full domino set (28 tiles, 7 for each player)
func dealHands(playerIDs []string) (map[string][]types.Tile, error) {
	pile := FullSet()
	if err := shuffleTiles(pile); err != nil {
		return nil, fmt.Errorf("couldn't shuffle tiles: %w", err)
	}

	const handSize = 7
	hands := make(map[string][]types.Tile, len(playerIDs))
	for idx, playerID := range playerIDs {
		hands[playerID] = pile[idx*handSize : (idx+1)*handSize]
	}
	return hands, nil
}

// holderOfDoubleSix returns whichever player in hands holds the 6-6 double.
func holderOfDoubleSix(hands map[string][]types.Tile) string {
	for playerID, hand := range hands {
		if slices.Contains(hand, (types.Tile{Left: 6, Right: 6})) {
			return playerID
		}
	}
	panic("expected a player to hold the 6-6 double")
}

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

// sameTile compares tiles ignoring orientation (2-5 == 5-2).
func sameTile(a, b types.Tile) bool {
	return a == b || a == b.Flip()
}
