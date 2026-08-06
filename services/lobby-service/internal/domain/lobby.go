package domain

import (
	"context"
	pbl "domino/shared/proto/lobby"
	"sort"
)

type LobbyStatus string

const (
	LobbyStatusWaiting  LobbyStatus = "LOBBY_STATUS_WAITING"
	LobbyStatusInGame   LobbyStatus = "LOBBY_STATUS_IN_GAME"
	LobbyStatusFinished LobbyStatus = "LOBBY_STATUS_FINISHED"
)

func (ls LobbyStatus) ToProto() pbl.LobbyStatus {
	switch ls {
	case LobbyStatusWaiting:
		return pbl.LobbyStatus_LOBBY_STATUS_WAITING
	case LobbyStatusInGame:
		return pbl.LobbyStatus_LOBBY_STATUS_IN_GAME
	case LobbyStatusFinished:
		return pbl.LobbyStatus_LOBBY_STATUS_FINISHED
	default:
		panic("unknow lobby status")
	}
}

type LobbyModel struct {
	ID         string                  `bson:"id"`
	HostID     string                  `bson:"hostID"`
	Status     LobbyStatus             `bson:"status"`
	Players    map[string]*PlayerModel `bson:"players"` // userID -> PlayerModel
	MaxPlayers int                     `bson:"maxPlayers"`
	Settings   LobbySettings           `bson:"settings"`
}

type PlayerModel struct {
	ID          string `bson:"id"`
	Name        string `bson:"name"`
	Slot        int    `bson:"slot"` // goes from 1 to MaxPlayers
	IsConnected bool   `bson:"isReady"`
}

type LobbySettings struct {
	MaxScore         int `bson:"maxScore"`
	TurnTimerSeconds int `bson:"TurnTimerSeconds"`
}

// SortedPlayers returns players ordered by Slot
func (l *LobbyModel) SortedPlayers() []*PlayerModel {
	players := make([]*PlayerModel, 0, len(l.Players))
	for _, p := range l.Players {
		players = append(players, p)
	}
	sort.Slice(players, func(i, j int) bool { return players[i].Slot < players[j].Slot })
	return players
}

func (l *LobbyModel) ToProto() *pbl.Lobby {
	sorted := l.SortedPlayers()
	protoPlayers := make([]*pbl.Player, len(sorted))
	for i, p := range sorted {
		protoPlayers[i] = p.ToProto()
	}
	return &pbl.Lobby{
		Id:         l.ID,
		HostId:     l.HostID,
		Players:    protoPlayers,
		MaxPlayers: int32(l.MaxPlayers),
		Status:     l.Status.ToProto(),
		Settings: &pbl.LobbySettings{
			MaxScore:         int32(l.Settings.MaxScore),
			TurnTimerSeconds: int32(l.Settings.TurnTimerSeconds),
		},
	}
}

func (p *PlayerModel) ToProto() *pbl.Player {
	return &pbl.Player{
		Id:          p.ID,
		Name:        p.Name,
		Slot:        int32(p.Slot),
		IsConnected: p.IsConnected,
	}
}

type LobbyEventPublisher interface {
	PublishPlayerJoined(ctx context.Context, lobby *LobbyModel, player *PlayerModel) error
}

type LobbyRepository interface {
	CreateLobby(ctx context.Context, l *LobbyModel) (*LobbyModel, error)
	GetLobbyByID(ctx context.Context, id string) (*LobbyModel, error)

	SetStatus(ctx context.Context, lobbyID string, status LobbyStatus) error // used for StartLobby for example

	AddPlayer(ctx context.Context, lobbyID string, p *PlayerModel) error
	SetPlayerConnection(ctx context.Context, lobbyID string, userID string, connected bool) error
}

type LobbyService interface {
	CreateLobby(ctx context.Context, hostID string, maxPlayers int) (*LobbyModel, error)
	JoinLobby(ctx context.Context, lobbyID string, userID string) (*LobbyModel, error)
	StartLobby(ctx context.Context, lobbyID string, userID string) (*LobbyModel, error)
	GetLobby(ctx context.Context, lobbyID string) (*LobbyModel, error)

	SetPlayerConnected(ctx context.Context, lobbyID string, userID string) error
	SetPlayerDisconnected(ctx context.Context, lobbyID string, userID string) error
}
