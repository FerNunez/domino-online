package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"domino/services/api-gateway/grpc_clients"
	"domino/shared/contracts"
	"domino/shared/jwt"
	"domino/shared/messaging"
	pbg "domino/shared/proto/game"
	"domino/shared/types"
)

// NOTE: connManager is a package-level singleton shared by all WebSocket handlers
// All handlers in the same process share one connection map, enablig cross-handler message delivery(eg. a RabbitMQ consumer pushing to UserID whose connection was registered by handleLobbyWebsocket

func handleLobbyWebsocket(w http.ResponseWriter, r *http.Request, rmq *messaging.RabbitMQ) {
	// check inputs: lobbyID, claimID
	lobbyID := r.PathValue("id")
	if lobbyID == "" {
		http.Error(w, "lobby id is required", http.StatusBadRequest)
		return
	}
	claims, err := jwt.ParseLobbyTicket(r.URL.Query().Get("wsToken"))
	if err != nil {
		fmt.Printf("couldnt approuve ws token: %v", err)
		http.Error(w, "couldnt approuve ws token", http.StatusUnauthorized)
		return
	}
	if claims.LobbyID != lobbyID {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	// Upgrade
	conn, err := connManager.Upgrade(w, r)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Add in Manager
	connManager.AddToLobby(claims.LobbyID, claims.UserID)
	connManager.Add(claims.UserID, conn)

	// NOTE: broadcast/target events (lobby.*, game.*) reach connManager via the
	// single gateway-wide consumer started once in main.go, not per connection here.
	// Create Game GRPC Client
	gameSvc, err := grpc_clients.NewGameServiceClient()
	if err != nil {
		fmt.Printf("couldnt reach game service: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Prepare msg to publish PlayerConnected
	playerConnectedMsg := messaging.PlayerConnectedData{
		UserID:  claims.UserID,
		LobbyID: claims.LobbyID,
	}
	data, err := json.Marshal(playerConnectedMsg)
	if err != nil {
		fmt.Printf("couldn't marshal joined player: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Publish player conenctedbefore connect
	if err := rmq.PublishMessage(r.Context(), contracts.PlayerConnected, &contracts.DominoEvent{
		LobbyID:  claims.LobbyID,
		TargetID: "",
		Data:     data,
	}); err != nil {
		fmt.Printf("couldn't notify player: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// This defer function runs when the websocket is dropt. Need to clean up and inform
	defer func() {
		// clean up:
		connManager.RemoveFromLobby(claims.LobbyID, claims.UserID)
		connManager.Remove(claims.UserID)
		// Prepare msg to publish Disconnect
		playerDisonnectedMsg := messaging.PlayerDisconnectedData{
			UserID:  claims.UserID,
			LobbyID: claims.LobbyID,
		}
		data, err := json.Marshal(playerDisonnectedMsg)
		if err != nil {
			log.Printf("couldn't marshal disconnected player: %v", err)
			return
		}
		// Publish player disconnected
		if err := rmq.PublishMessage(r.Context(), contracts.PlayerDisconnected, &contracts.DominoEvent{
			LobbyID:  claims.LobbyID,
			TargetID: "",
			Data:     data,
		}); err != nil {
			log.Printf("couldn't notify player disconnected: %v", err)
			return
		}
	}()

	// Read loop => read each user msg => check type and handle it accordingly
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		// Unmarshal into {Type and Data}
		type playerMessage struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}
		var playerMsg playerMessage
		if err := json.Unmarshal(message, &playerMsg); err != nil {
			log.Printf("Error unmarshaling player message: %v", err)
			continue
		}

		log.Printf("%v: got message %v", claims.UserID, playerMsg)

		// Handle by Msg Type
		switch playerMsg.Type {

		// -- Play a tile
		case contracts.PlayTileCmd:
			var cmd playTileCmd
			if err := json.Unmarshal(playerMsg.Data, &cmd); err != nil {
				log.Printf("Error unmarshaling play tile command: %v", err)
				continue
			}
			req := &pbg.PlayTileRequest{
				LobbyId: lobbyID,
				UserId:  claims.UserID,
				Tile:    &pbg.Tile{Left: cmd.Tile.Left, Right: cmd.Tile.Right},
				Side:    cmd.Side,
			}
			responseGrpc, err := gameSvc.Client.PlayTile(r.Context(), req)
			if err != nil {
				log.Printf("couldn't play tile to server: %v", err)
				continue
			}

			response := messaging.PlayTileResponseData{
				Board:       toTiles(responseGrpc.Board),
				Hand:        toTiles(responseGrpc.Hand),
				RoundResult: toRoundResult(responseGrpc.RoundResult),
			}

			// Send response to WebSocket
			WSMsg := contracts.WSMessage{
				Type: contracts.PlayTileResponse,
				Data: response,
			}
			if err := connManager.SendMessage(claims.UserID, WSMsg); err != nil {
				log.Printf("couldt send message: %v", err)
			}

		// -- PassTurn
		case contracts.PassTurnCmd:
			// Pass cmd doesnt pass anything
			req := &pbg.PassTurnRequest{
				LobbyId: lobbyID,
				UserId:  claims.UserID,
			}
			responseGrpc, err := gameSvc.Client.PassTurn(r.Context(), req)
			if err != nil {
				log.Printf("couldn't pass turn to server: %v", err)
				continue
			}
			response := messaging.PassTurnResponseData{
				RoundResult: toRoundResult(responseGrpc.RoundResult),
			}

			// Send response to WebSocket
			WSMsg := contracts.WSMessage{
				Type: contracts.PassTurnResponse,
				Data: response,
			}
			if err := connManager.SendMessage(claims.UserID, WSMsg); err != nil {
				log.Printf("couldt send message: %v", err)
			}

		default:
			log.Printf("Unknown message type: %s", playerMsg.Type)
		}
	}
}

func toTiles(tiles []*pbg.Tile) []types.Tile {
	out := make([]types.Tile, len(tiles))
	for idx, t := range tiles {
		out[idx] = types.Tile{
			Left:  int(t.Left),
			Right: int(t.Right),
		}
	}
	return out
}

func toRoundResult(roundResult *pbg.RoundResult) *types.RoundResult {
	if roundResult == nil {
		return nil
	}

	scores := make(map[types.TeamID]int, len(roundResult.Scores))
	for key, val := range roundResult.Scores {
		scores[types.TeamID(key)] = int(val)
	}

	return &types.RoundResult{
		WinnerTeamID: types.TeamID(roundResult.WinnerId),
		Reason:       types.Reason(roundResult.Reason),
		Scores:       scores,
	}
}
