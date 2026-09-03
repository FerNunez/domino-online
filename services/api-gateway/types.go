package main

import (
	pbl "domino/shared/proto/lobby"
)

// Create Lobby
type createLobbyRequest struct {
	MaxPlayers  int    `json:"maxPlayers"`
	DisplayName string `json:"displayName"`
}

func newProtoCreateLobbyRequest(req *createLobbyRequest, userID string) *pbl.CreateLobbyRequest {
	return &pbl.CreateLobbyRequest{
		UserID:      userID,
		MaxPlayers:  int32(req.MaxPlayers),
		DisplayName: req.DisplayName,
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

type joinLobbyRequest struct {
	DisplayName string `json:"displayName"`
}

func newProtoJoinLobbyRequest(userID, lobbyID, displayName string) *pbl.JoinLobbyRequest {
	return &pbl.JoinLobbyRequest{
		UserID:      userID,
		LobbyID:     lobbyID,
		DisplayName: displayName,
	}
}

func newProtoReconnectLobbyRequest(userID, lobbyID string) *pbl.ReconnectLobbyRequest {
	return &pbl.ReconnectLobbyRequest{
		UserID:  userID,
		LobbyID: lobbyID,
	}
}

// Start Game
type startLobbyRequest struct {
	LobbyID string `json:"lobbyID"`
	UserID  string `json:"userID"`
}

func (c *startLobbyRequest) toProto() *pbl.StartLobbyRequest {
	return &pbl.StartLobbyRequest{
		LobbyID: c.LobbyID,
		UserID:  c.UserID,
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

type playTileCmd struct {
	Tile struct {
		Left  int32 `json:"left"`
		Right int32 `json:"right"`
	} `json:"tile"`
	Side string `json:"side"`
}
