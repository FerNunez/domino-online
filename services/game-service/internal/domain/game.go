package domain

import (
	"context"
	//pbl "domino/shared/proto/game"
)

type GameStatus string

// const (
// 	GameStatusWaiting  GameStatus = "GAME_STATUS_WAITING"
// 	GameStatusInGame   GameStatus = "GAME_STATUS_IN_GAME"
// 	GameStatusFinished GameStatus = "GAME_STATUS_FINISHED"
// )
//
// func (ls GameStatus) ToProto() pbl.GameStatus {
// 	switch ls {
// 	case GameStatusWaiting:
// 		return pbl.GameStatus_GAME_STATUS_WAITING
// 	case GameStatusInGame:
// 		return pbl.GameStatus_GAME_STATUS_IN_GAME
// 	case GameStatusFinished:
// 		return pbl.GameStatus_GAME_STATUS_FINISHED
// 	default:
// 		panic("unknow game status")
// 	}
// }

type Tile string

type GameModel struct {
	LobbyID     string
	PlayerTiles [][]string // [UserID][tiles]
	Head        []string
	Tail        []string
}

type GameRepository interface {
	// 	CreateGame(ctx context.Context) (*GameModel, error)
}

type GameService interface {
	CreateGame(ctx context.Context, lobbyID string) (*GameModel, error)
}
