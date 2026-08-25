package domain

import (
	"domino/shared/types"
	"testing"
)

func newTestRound() *RoundModel {
	return &RoundModel{
		LobbyID:     "lobby-1",
		GameID:      "game-1",
		ID:          "round-1",
		RoundNumber: 1,
		Status:      RoundStatusInProgress,
		PlayerOrder: []string{"p1", "p2", "p3", "p4"},
		Hands: map[string][]types.Tile{
			"p1": {{Left: 6, Right: 6}, {Left: 5, Right: 5}},
			"p2": {{Left: 1, Right: 2}},
			"p3": {},
			"p4": {},
		},
		Board:          emptyBoard(),
		StartingPlayer: "p1",
		CurrentTurn:    "p1",
	}
}

func TestPlayTile_ActionCount(t *testing.T) {
	t.Run("accepted move increments ActionCount and advances turn", func(t *testing.T) {
		r := newTestRound()
		if err := r.PlayTile("p1", types.Tile{Left: 6, Right: 6}, SideLeft); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if r.ActionCount != 1 {
			t.Errorf("ActionCount = %d, want 1", r.ActionCount)
		}
		if r.CurrentTurn != "p2" {
			t.Errorf("CurrentTurn = %q, want p2", r.CurrentTurn)
		}
	})

	t.Run("wrong turn does not increment ActionCount", func(t *testing.T) {
		r := newTestRound()
		err := r.PlayTile("p2", types.Tile{Left: 1, Right: 2}, SideLeft)
		if err != ErrNotYourTurn {
			t.Fatalf("err = %v, want ErrNotYourTurn", err)
		}
		if r.ActionCount != 0 {
			t.Errorf("ActionCount = %d, want 0", r.ActionCount)
		}
	})

	t.Run("tile not in hand does not increment ActionCount", func(t *testing.T) {
		r := newTestRound()
		err := r.PlayTile("p1", types.Tile{Left: 3, Right: 4}, SideLeft)
		if err != ErrTileNotInHand {
			t.Fatalf("err = %v, want ErrTileNotInHand", err)
		}
		if r.ActionCount != 0 {
			t.Errorf("ActionCount = %d, want 0", r.ActionCount)
		}
	})

	t.Run("illegal move does not increment ActionCount", func(t *testing.T) {
		r := newTestRound()
		if err := r.PlayTile("p1", types.Tile{Left: 6, Right: 6}, SideLeft); err != nil {
			t.Fatalf("setup move failed: %v", err)
		}
		// board's open ends are now 6/6; p2's {1,2} doesn't match either end.
		err := r.PlayTile("p2", types.Tile{Left: 1, Right: 2}, SideLeft)
		if err != ErrIllegalMove {
			t.Fatalf("err = %v, want ErrIllegalMove", err)
		}
		if r.ActionCount != 1 {
			t.Errorf("ActionCount = %d, want 1 (unchanged from setup move)", r.ActionCount)
		}
	})

	t.Run("round already over does not increment ActionCount", func(t *testing.T) {
		r := newTestRound()
		r.Status = RoundStatusRoundOver
		err := r.PlayTile("p1", types.Tile{Left: 6, Right: 6}, SideLeft)
		if err != ErrRoundOver {
			t.Fatalf("err = %v, want ErrRoundOver", err)
		}
		if r.ActionCount != 0 {
			t.Errorf("ActionCount = %d, want 0", r.ActionCount)
		}
	})
}

func TestPassTurn_ActionCount(t *testing.T) {
	t.Run("accepted pass increments ActionCount", func(t *testing.T) {
		r := newTestRound()
		// give the board an open end p1's hand can't match, so passing is legal.
		r.Board = Board{Tiles: []types.Tile{{Left: 3, Right: 3}}, LeftEnd: 3, RightEnd: 3}
		err := r.PassTurn("p1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if r.ActionCount != 1 {
			t.Errorf("ActionCount = %d, want 1", r.ActionCount)
		}
		if r.PassStreak != 1 {
			t.Errorf("PassStreak = %d, want 1", r.PassStreak)
		}
	})

	t.Run("has legal move does not increment ActionCount", func(t *testing.T) {
		r := newTestRound()
		// board's open ends are 6/6, matching p1's {6,6} hand — a legal move exists.
		r.Board = Board{Tiles: []types.Tile{{Left: 6, Right: 6}}, LeftEnd: 6, RightEnd: 6}
		err := r.PassTurn("p1")
		if err != ErrHasLegalMove {
			t.Fatalf("err = %v, want ErrHasLegalMove", err)
		}
		if r.ActionCount != 0 {
			t.Errorf("ActionCount = %d, want 0", r.ActionCount)
		}
	})

	t.Run("wrong turn does not increment ActionCount", func(t *testing.T) {
		r := newTestRound()
		err := r.PassTurn("p2")
		if err != ErrNotYourTurn {
			t.Fatalf("err = %v, want ErrNotYourTurn", err)
		}
		if r.ActionCount != 0 {
			t.Errorf("ActionCount = %d, want 0", r.ActionCount)
		}
	})

	t.Run("round already over does not increment ActionCount", func(t *testing.T) {
		r := newTestRound()
		r.Status = RoundStatusRoundOver
		err := r.PassTurn("p1")
		if err != ErrRoundOver {
			t.Fatalf("err = %v, want ErrRoundOver", err)
		}
		if r.ActionCount != 0 {
			t.Errorf("ActionCount = %d, want 0", r.ActionCount)
		}
	})
}
