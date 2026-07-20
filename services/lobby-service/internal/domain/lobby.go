package domain

import (
	"context"
	pbl "rebu/shared/proto/lobby"

	"go.mongodb.org/mongo-driver/bson/primitive"
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
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	HostID     string             `bson:"hostID"`
	Status     LobbyStatus        `bson:"status"`
	Players    []*PlayerModel     `bson:"players"`
	MaxPlayers int                `bson:"maxPlayers"`
	Settings   LobbySettings      `bson:"settings"`
}

type PlayerModel struct {
	ID      primitive.ObjectID `bson:"_id,omitempty"`
	Name    string             `bson:"name"`
	Slot    int                `bson:"slot"`
	IsReady bool               `bson:"isReady"`
}

type LobbySettings struct {
	MaxScore         int `bson:"maxScore"`
	TurnTimerSeconds int `bson:"TurnTimerSeconds"`
}

func (l *LobbyModel) ToProto() *pbl.Lobby {
	protoPlayers := make([]*pbl.Player, len(l.Players))
	for i, p := range l.Players {
		protoPlayers[i] = p.ToProto()
	}
	return &pbl.Lobby{
		Id:         l.ID.Hex(),
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
		Id:   p.ID.String(),
		Name: p.Name,
		Slot: int32(p.Slot),
	}
}

type LobbyRepository interface {
	CreateLobby(ctx context.Context, l *LobbyModel) (*LobbyModel, error)
	//StartLobby(ctx context.Context, id string) error

	GetLobbyByID(ctx context.Context, id string) (*LobbyModel, error)
	// TODO: UpdateLobby(ctx context.Context, lobbyID string, status string, driver *pbd.Driver) error

	UpdateLobby(ctx context.Context, id string, l *LobbyModel) (*LobbyModel, error)
	//AddPlayer(ctx context.Context, id string, p *PlayerModel) (*LobbyModel, error)
}

type LobbyService interface {
	CreateLobby(ctx context.Context, hostID string, maxPlayers int) (*LobbyModel, error)
	JoinLobby(ctx context.Context, id string, secretToken string, player *PlayerModel) (*LobbyModel, error)
	//StartLobby(ctx context.Context, id string, host_id string) error
}
