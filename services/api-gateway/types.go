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
	UserID string `json:"userID"`
}

func (c *createLobbyRequest) toProto() *pbl.CreateLobbyRequest {
	return &pbl.CreateLobbyRequest{UserID: c.UserID}
}

// Join Lobby
type joinLobbyRequest struct {
	UserID     string `json:"userID"`
	LobbyID    string `json:"lobbyID"`
	SecretCode string `json:"secretCode"`
}

func (c *joinLobbyRequest) toProto() *pbl.JoinLobbyRequest {
	return &pbl.JoinLobbyRequest{
		UserID:     c.UserID,
		LobbyID:    c.LobbyID,
		SecretCode: c.SecretCode,
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
