package domain

import (
	"testing"
)

func fourPlayers() []string {
	return []string{"p1", "p2", "p3", "p4"}
}

func TestNewGame_DealsFullSetWithNoDuplicates(t *testing.T) {
	game, err := NewGame("lobby-1", fourPlayers())
	if err != nil {
		t.Fatalf("NewGame returned error: %v", err)
	}

	seen := make(map[Tile]bool)
	total := 0
	for _, playerID := range fourPlayers() {
		hand := game.Hands[playerID]
		if len(hand) != 7 {
			t.Fatalf("player %s: expected 7 tiles, got %d", playerID, len(hand))
		}
		for _, tile := range hand {
			if seen[tile] {
				t.Fatalf("tile %v dealt more than once", tile)
			}
			seen[tile] = true
			total++
		}
	}
	if total != 28 {
		t.Fatalf("expected 28 tiles dealt total, got %d", total)
	}
}

func TestNewGame_StartingPlayerHoldsDoubleSix(t *testing.T) {
	game, err := NewGame("lobby-1", fourPlayers())
	if err != nil {
		t.Fatalf("NewGame returned error: %v", err)
	}

	hand := game.Hands[game.CurrentTurn]
	found := false
	for _, tile := range hand {
		if tile == (Tile{Left: 6, Right: 6}) {
			found = true
		}
	}
	if !found {
		t.Fatalf("starting player %s does not hold 6-6, hand=%v", game.CurrentTurn, hand)
	}
	if game.Status != GameStatusDealt {
		t.Fatalf("expected status %v, got %v", GameStatusDealt, game.Status)
	}
}

func TestNewGame_RejectsWrongPlayerCount(t *testing.T) {
	if _, err := NewGame("lobby-1", []string{"p1", "p2"}); err != ErrWrongPlayerCnt {
		t.Fatalf("expected ErrWrongPlayerCnt, got %v", err)
	}
}

// newTestGame builds a deterministic game with hand-picked hands so rule
// enforcement can be exercised without relying on the shuffle.
func newTestGame() *GameModel {
	return &GameModel{
		LobbyID:     "lobby-1",
		Status:      GameStatusDealt,
		PlayerOrder: fourPlayers(),
		Hands: map[string][]Tile{
			"p1": {{6, 6}, {6, 5}, {1, 2}},
			"p2": {{5, 4}, {0, 0}, {3, 3}},
			"p3": {{4, 4}, {2, 2}, {1, 1}},
			"p4": {{3, 4}, {2, 3}, {6, 0}},
		},
		Board:       emptyBoard(),
		CurrentTurn: "p1",
	}
}

func TestPlayTile_RejectsWrongTurn(t *testing.T) {
	game := newTestGame()
	err := game.PlayTile("p2", Tile{5, 4}, SideLeft)
	if err != ErrNotYourTurn {
		t.Fatalf("expected ErrNotYourTurn, got %v", err)
	}
}

func TestPlayTile_RejectsTileNotInHand(t *testing.T) {
	game := newTestGame()
	err := game.PlayTile("p1", Tile{2, 6}, SideLeft)
	if err != ErrTileNotInHand {
		t.Fatalf("expected ErrTileNotInHand, got %v", err)
	}
}

func TestPlayTile_RejectsNonMatchingPip(t *testing.T) {
	game := newTestGame()
	// p1 opens with 6-6, board ends become 6/6.
	if err := game.PlayTile("p1", Tile{6, 6}, SideLeft); err != nil {
		t.Fatalf("unexpected error opening board: %v", err)
	}

	// p2 tries to play 3-3, which matches neither open end.
	err := game.PlayTile("p2", Tile{3, 3}, SideLeft)
	if err != ErrIllegalMove {
		t.Fatalf("expected ErrIllegalMove, got %v", err)
	}
}

func TestPassTurn_RejectsWhenLegalMoveExists(t *testing.T) {
	game := newTestGame()
	if err := game.PlayTile("p1", Tile{6, 6}, SideLeft); err != nil {
		t.Fatalf("unexpected error opening board: %v", err)
	}

	// p2 holds 0-0/... none match 6, but p3 (next after p2 passes) holds 4-4/... unrelated.
	// Force a scenario where the current player DOES have a legal move: p4 holds 6-0.
	game.CurrentTurn = "p4"
	if err := game.PassTurn("p4"); err != ErrHasLegalMove {
		t.Fatalf("expected ErrHasLegalMove, got %v", err)
	}
}

func TestPassTurn_AllowedWhenNoLegalMove(t *testing.T) {
	game := newTestGame()
	if err := game.PlayTile("p1", Tile{6, 6}, SideLeft); err != nil {
		t.Fatalf("unexpected error opening board: %v", err)
	}

	// p2's hand (5-4, 0-0, 3-3) has no 6, so passing must succeed.
	if err := game.PassTurn("p2"); err != nil {
		t.Fatalf("expected pass to succeed, got %v", err)
	}
	if game.CurrentTurn != "p3" {
		t.Fatalf("expected turn to advance to p3, got %s", game.CurrentTurn)
	}
	if game.PassStreak != 1 {
		t.Fatalf("expected pass streak 1, got %d", game.PassStreak)
	}
}

func TestRound_EndsInDominoWithCorrectScores(t *testing.T) {
	// Hands rigged so p1 empties their hand in three plays while it stays their
	// turn again each time (others pass with genuinely no legal move).
	game := &GameModel{
		LobbyID:     "lobby-1",
		Status:      GameStatusDealt,
		PlayerOrder: fourPlayers(),
		Hands: map[string][]Tile{
			"p1": {{6, 6}, {6, 5}, {5, 5}},
			"p2": {{0, 0}, {1, 1}},
			"p3": {{2, 2}, {3, 3}},
			"p4": {{1, 2}},
		},
		Board:       emptyBoard(),
		CurrentTurn: "p1",
	}

	mustPlay := func(userID string, tile Tile, side string) {
		t.Helper()
		if err := game.PlayTile(userID, tile, side); err != nil {
			t.Fatalf("PlayTile(%s, %v, %s) failed: %v", userID, tile, side, err)
		}
	}
	mustPass := func(userID string) {
		t.Helper()
		if err := game.PassTurn(userID); err != nil {
			t.Fatalf("PassTurn(%s) failed: %v", userID, err)
		}
	}

	mustPlay("p1", Tile{6, 6}, SideLeft) // board: 6-6, ends 6/6
	mustPass("p2")                       // no 6 in hand
	mustPass("p3")                       // no 6 in hand
	mustPass("p4")                       // no 6 in hand

	mustPlay("p1", Tile{6, 5}, SideLeft) // board ends become 5/6
	mustPass("p2")
	mustPass("p3")
	mustPass("p4")

	mustPlay("p1", Tile{5, 5}, SideLeft) // p1's hand now empty -> round over

	if game.Status != GameStatusRoundOver {
		t.Fatalf("expected round over, got status %v", game.Status)
	}

	result := game.ResolveRoundResult()
	if result.Reason != ReasonDomino {
		t.Fatalf("expected reason %v, got %v", ReasonDomino, result.Reason)
	}
	if result.WinnerID != "p1" {
		t.Fatalf("expected winner p1, got %s", result.WinnerID)
	}
	want := map[string]int{"p2": 2, "p3": 10, "p4": 3}
	for playerID, wantScore := range want {
		if got := result.Scores[playerID]; got != wantScore {
			t.Errorf("scores[%s] = %d, want %d", playerID, got, wantScore)
		}
	}
}

func TestRound_EndsBlockedWithLowestHandWinner(t *testing.T) {
	game := &GameModel{
		LobbyID:     "lobby-1",
		Status:      GameStatusInProgress,
		PlayerOrder: fourPlayers(),
		Hands: map[string][]Tile{
			"p1": {{6, 6}}, // pips: 12
			"p2": {{1, 1}}, // pips: 2 (lowest)
			"p3": {{3, 3}}, // pips: 6
			"p4": {{4, 4}}, // pips: 8
		},
		Board:       Board{Tiles: []Tile{{0, 0}}, LeftEnd: 0, RightEnd: 0},
		CurrentTurn: "p1",
		PassStreak:  3, // p2, p3, p4 already passed
	}

	if err := game.PassTurn("p1"); err != nil {
		t.Fatalf("expected final pass to succeed, got %v", err)
	}
	if game.Status != GameStatusRoundOver {
		t.Fatalf("expected round over after full pass streak, got %v", game.Status)
	}

	result := game.ResolveRoundResult()
	if result.Reason != ReasonBlocked {
		t.Fatalf("expected reason %v, got %v", ReasonBlocked, result.Reason)
	}
	if result.WinnerID != "p2" {
		t.Fatalf("expected winner p2 (lowest pips), got %s", result.WinnerID)
	}
	want := map[string]int{"p1": 12, "p3": 6, "p4": 8}
	for playerID, wantScore := range want {
		if got := result.Scores[playerID]; got != wantScore {
			t.Errorf("scores[%s] = %d, want %d", playerID, got, wantScore)
		}
	}
}

func TestRound_BlockedTieHasNoWinner(t *testing.T) {
	game := &GameModel{
		LobbyID:     "lobby-1",
		Status:      GameStatusInProgress,
		PlayerOrder: fourPlayers(),
		Hands: map[string][]Tile{
			"p1": {{1, 1}}, // pips: 2
			"p2": {{2, 0}}, // pips: 2 (tie)
			"p3": {{3, 3}}, // pips: 6
			"p4": {{4, 4}}, // pips: 8
		},
		Board:       Board{Tiles: []Tile{{5, 5}}, LeftEnd: 5, RightEnd: 5},
		CurrentTurn: "p1",
		PassStreak:  3,
	}

	if err := game.PassTurn("p1"); err != nil {
		t.Fatalf("expected final pass to succeed, got %v", err)
	}

	result := game.ResolveRoundResult()
	if result.Reason != ReasonBlocked {
		t.Fatalf("expected reason %v, got %v", ReasonBlocked, result.Reason)
	}
	if result.WinnerID != "" {
		t.Fatalf("expected no winner on tie, got %s", result.WinnerID)
	}
	if len(result.Scores) != 0 {
		t.Fatalf("expected no scores on tie, got %v", result.Scores)
	}
}

func TestValidMoves_EmptyBoardAllowsAnyTile(t *testing.T) {
	hand := []Tile{{1, 2}, {3, 4}}
	moves := ValidMoves(hand, emptyBoard())
	if len(moves) != len(hand) {
		t.Fatalf("expected %d moves on empty board, got %d", len(hand), len(moves))
	}
}

func TestValidMoves_MatchesOpenEnds(t *testing.T) {
	board := Board{Tiles: []Tile{{6, 6}}, LeftEnd: 6, RightEnd: 6}
	hand := []Tile{{6, 5}, {1, 2}, {6, 3}}
	moves := ValidMoves(hand, board)

	matchesTile := func(tile Tile) int {
		count := 0
		for _, m := range moves {
			if m.Tile == tile {
				count++
			}
		}
		return count
	}
	if matchesTile(Tile{6, 5}) != 2 { // matches both ends since board is 6/6
		t.Errorf("expected 6-5 to match both ends, got %d matches", matchesTile(Tile{6, 5}))
	}
	if matchesTile(Tile{1, 2}) != 0 {
		t.Errorf("expected 1-2 to have no valid moves, got %d", matchesTile(Tile{1, 2}))
	}
}
