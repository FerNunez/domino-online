package messaging

import "domino/shared/types"

// Queue name constants: single source of truth
const (
	DeaedLetterQueue = "deaed_letter_queue"
	NotifyLobby      = "notify_lobby_queue"
	NotifyGame       = "notify_game_queue"
)

type PlayerJoinedData struct {
	UserID      string `json:"userID"`
	DisplayName string `json:"displayName"`
	PlayerCount int    `json:"playerCount"`
	MaxPlayers  int    `json:"maxPlayers"`
}

type PlayerLeftData struct {
	UserID      string `json:"userID"`
	PlayerCount int    `json:"playerCount"`
}

type GameStartCmd struct {
	PlayersID []string `json:"playerID"`
}

type GameStartedData struct {
	PlayerOrder []string       `json:"playerOrder"` // userIDs, order: 0, 1, 2 ,3
	HandsSize   map[string]int `json:"handsSize"`
	CurrentTurn string         `json:"currentTurn"` // an userID
}

type HandDeltData struct {
	PlayerID    string   `json:"playerID"`
	PlayerTiles []string `json:"playerTiles"`
}

type MoveMadeData struct {
	UserID      string             `json:"userID"`
	Tile        types.Tile         `json:"tile"`
	Side        string             `json:"side"`
	NextTurn    string             `json:"next_turn"`
	RoundResult *types.RoundResult `json:"round_result,omitempty"`
}

type PlayerPassedData struct {
	UserID      string             `json:"userID"`
	NextTurn    string             `json:"next_turn"`
	RoundResult *types.RoundResult `json:"round_result,omitempty"`
}

type GameEndedData struct {
	WinnerID string         `json:"winnerID"`
	Reason   string         `json:"reason"` // "domino" | "blocked"
	Scores   map[string]int `json:"scores"`
}

type PlayTileResponseData struct {
	Board       []types.Tile       `json:"board"`
	Hand        []types.Tile       `json:"hand"`
	RoundResult *types.RoundResult `json:"round_result,omitempty"`
}

type PassTurnResponseData struct {
	RoundResult *types.RoundResult `json:"round_result,omitempty"`
}
