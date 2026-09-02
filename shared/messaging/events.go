// Package messaging contains all datatypes for  message transmitted as events
package messaging

import "domino/shared/types"

// Queue name constants: single source of truth
const (
	DeaedLetterQueue = "deaed_letter_queue"
	// "" // Each gateway has its unique queue
	GameQueue      = "game_queue"       // game service consomes this queue
	LobbyQueue     = "lobby_queue"      // lobby service consomes this queue
	GameStoreQueue = "game_store_queue" // history service consumes this queue
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

type PlayerConnectedData struct {
	UserID  string `json:"userID"`
	LobbyID string `json:"lobbyID"`
}

type PlayerDisconnectedData struct {
	UserID  string `json:"userID"`
	LobbyID string `json:"lobbyID"`
}

type GameStartCmd struct {
	PlayersID []string `json:"playerID"`
	GameID    string   `json:"gameID"`
}

type GameStartedData struct {
	GameID      string               `json:"gameID"`
	PlayerOrder []string             `json:"playerOrder"` // userIDs, order: 0, 1, 2 ,3
	HandsSize   map[string]int       `json:"handsSize"`
	CurrentTurn string               `json:"currentTurn"` // an userID
	Scores      map[types.TeamID]int `json:"scores"`
}
type GameOverData struct {
	LobbyID    string               `json:"lobbyID"`
	GameID     string               `json:"gameID"`
	GameState  string               `json:"gameState"`
	GameScore  map[types.TeamID]int `json:"gameScores"`
	TeamWinner types.TeamID         `json:"teamWinner"`
}

type HandDeltData struct {
	RoundID     string       `json:"roundID"`
	PlayerID    string       `json:"playerID"`
	PlayerTiles []types.Tile `json:"playerTiles"`
}

type MoveMadeData struct {
	UserID            string             `json:"userID"`
	Tile              types.Tile         `json:"tile"`
	Side              types.Side         `json:"side"`
	NextTurn          string             `json:"nextTurn"`
	RoundResult       *types.RoundResult `json:"roundResult,omitempty"`
	RoundID           string             `json:"roundID"`
	ActionNumber      int                `json:"actionNumber"`
	ResultingLeftEnd  int                `json:"resultingLeftEnd"`
	ResultingRightEnd int                `json:"resultingRightEnd"`
}

type PlayerPassedData struct {
	UserID       string             `json:"userID"`
	NextTurn     string             `json:"nextTurn"`
	RoundResult  *types.RoundResult `json:"roundResult,omitempty"`
	RoundID      string             `json:"roundID"`
	ActionNumber int                `json:"actionNumber"`
}

type RoundStartedData struct {
	LobbyID        string  `json:"lobbyID"`
	RoundID        string  `json:"roundID"`
	StartingPlayer *string `json:"nextStartingPlayer"`
	RoundNumber    int     `json:"roundNumber"`
}

type RoundOverData struct {
	LobbyID        string               `json:"lobbyID"`
	GameID         string               `json:"gameID"`
	RoundID        string               `json:"roundID"`
	RoundNumber    int                  `json:"roundNumber"`
	StartingPlayer string               `json:"startingPlayer"`
	PlayerOrder    []string             `json:"playerOrder"`
	RoundResult    *types.RoundResult   `json:"roundResult"`
	ActionCount    int                  `json:"actionCount"`
	RoundWinner    types.TeamID         `json:"roundWinner"`
	GameScore      map[types.TeamID]int `json:"gameScores"`
	GameState      string               `json:"gameState"`
}

type PlayTileResponseData struct {
	Board       []types.Tile       `json:"board"`
	Hand        []types.Tile       `json:"hand"`
	RoundResult *types.RoundResult `json:"roundResult,omitempty"`
}

type PassTurnResponseData struct {
	RoundResult *types.RoundResult `json:"roundResult,omitempty"`
}

type NextRoundResponseData struct {
	RoundNumber int `json:"roundNumber"`
}

// BoardData mirrors protos Board
type BoardData struct {
	Tiles    []types.Tile `json:"tiles"`
	LeftEnd  int          `json:"leftEnd"`
	RightEnd int          `json:"rightEnd"`
}

// GameStateSnapshotData mirrors protos GetGameStateResponse
type GameStateSnapshotData struct {
	GameID      string               `json:"gameID"`
	GameNumber  int                  `json:"gameNumber"`
	Status      string               `json:"status"`
	TeamScores  map[types.TeamID]int `json:"teamScores"`
	TeamWinner  types.TeamID         `json:"teamWinner"`
	GoalScore   int                  `json:"goalScore"`
	RoundID     string               `json:"roundID"`
	RoundNumber int                  `json:"roundNumber"`
	RoundStatus string               `json:"roundStatus"`
	PlayerOrder []string             `json:"playerOrder"`
	CurrentTurn string               `json:"currentTurn"`
	Board       BoardData            `json:"board"`
	Hand        []types.Tile         `json:"hand"`
	HandSizes   map[string]int       `json:"handSizes"`
	RoundResult *types.RoundResult   `json:"roundResult,omitempty"`
}
