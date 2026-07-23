package messaging

// Queue name constants: single source of truth
const (
	DeaedLetterQueue = "deaed_letter_queue"
	NotifyLobby      = "notify_lobby_queue"
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

type GameStartedData struct {
	StartingPlayerID string `json:"StartingPlayerID"`
}

type HandDeltData struct {
	Tiles []string `json:"tiles"`
}

type TurnChangedData struct {
	UserID string `json:"userID"`
}

type MoveChangedData struct {
	UserID string `json:"userID"`
	Tile   string `json:"tile"`
	Side   string `json:"side"`
}

type PlayerPassedData struct {
	UserID string `json:"userID"`
}
