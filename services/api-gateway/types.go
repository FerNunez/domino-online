package main

import (
	pbl "rebu/shared/proto/lobby"
	pb "rebu/shared/proto/trip"
	"rebu/shared/types"
)

type previewTripRequest struct {
	UserID      string           `json:"userID"`
	Pickup      types.Coordinate `json:"pickup"`
	Destination types.Coordinate `json:"destination"`
}

func (p *previewTripRequest) toProto() *pb.PreviewTripRequest {
	return &pb.PreviewTripRequest{
		UserID: p.UserID,
		StartLocation: &pb.Coordinate{
			Latitude:  p.Pickup.Latitude,
			Longitude: p.Pickup.Longitude,
		},
		EndLocation: &pb.Coordinate{
			Latitude:  p.Destination.Latitude,
			Longitude: p.Destination.Longitude,
		},
	}
}

type startTripRequest struct {
	RideFareID string `json:"rideFareID"`
	UserID     string `json:"userID"`
}

func (c *startTripRequest) toProto() *pb.CreateTripRequest {
	return &pb.CreateTripRequest{
		RideFareID: c.RideFareID,
		UserID:     c.UserID,
	}
}

// Create Lobby
type createLobbyRequest struct {
	MaxPlayers int `json:"maxPlayers"`
}

func newProtoCreateLobbyRequest(req *createLobbyRequest, userID string) *pbl.CreateLobbyRequest {
	return &pbl.CreateLobbyRequest{
		UserID:     userID,
		MaxPlayers: int32(req.MaxPlayers),
	}
}

type createLobbyResponse struct {
	LobbyID string `json:"lobbyID"`
	WsToken string `json:"wsToken"`
}

func newProtoCreateLobbyResponse(lobbyID, WsToken string) createLobbyResponse {
	return createLobbyResponse{
		LobbyID: lobbyID,
		WsToken: WsToken,
	}
}

func newProtoJoinLobbyRequest(userID, lobbyID string) *pbl.JoinLobbyRequest {
	return &pbl.JoinLobbyRequest{
		UserID:  userID,
		LobbyID: lobbyID,
	}
}

// Start Game
type startGameRequest struct {
	LobbyID string `json:"lobbyID"`
	HostID  string `json:"hostID"`
}

func (c *startGameRequest) toProto() *pbl.StartGameRequest {
	return &pbl.StartGameRequest{
		LobbyID: c.LobbyID,
		HostID:  c.HostID,
	}
}

// Create Lobby
type AuthResponse struct {
	UserID string `json:"userID"`
	Type   string `json:"type"`
}

type RegisterRequest struct {
	DisplayName string `json:"displayName"`
	Email       string `json:"email"`
	Password    string `json:"password"`
}
