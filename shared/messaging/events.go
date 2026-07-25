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
	PlayersID []string `json:"playerID"`
}

type HandsDeltData struct {
	PlayerTiles      map[string][]string `json:"playerTiles"`
	StartingPlayerID string              `json:"StartingPlayerID"`
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
