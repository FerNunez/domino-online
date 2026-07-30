package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"domino/services/api-gateway/grpc_clients"
	"domino/shared/contracts"
	"domino/shared/jwt"
	gpb "domino/shared/proto/game"
)

// NOTE: connManager is a package-level singleton shared by all WebSocket handlers
// All handlers in the same process share one connection map, enablig cross-handler message delivery(eg. a RabbitMQ consumer pushing to UserID whose connection was registered by handleLobbyWebsocket

func handleLobbyWebsocket(w http.ResponseWriter, r *http.Request) {
	// Inputs check: lobbyID, claimID
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
	connManager.AddToLobby(lobbyID, claims.UserID)
	defer connManager.RemoveFromLobby(lobbyID, claims.UserID)
	connManager.Add(claims.UserID, conn)
	defer connManager.Remove(claims.UserID)

	// NOTE: broadcast/target events (lobby.*, game.*) reach connManager via the
	// single gateway-wide consumer started once in main.go, not per connection here.

	// Create Game GRPC Client
	gameSvc, err := grpc_clients.NewGameServiceClient()
	if err != nil {
		fmt.Printf("couldnt reach game service: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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

		// Handle by Msg Type
		switch playerMsg.Type {

		// Play a tile
		case contracts.PlayTileCmd:
			var cmd playTileCmd
			if err := json.Unmarshal(playerMsg.Data, &cmd); err != nil {
				log.Printf("Error unmarshaling play tile command: %v", err)
				continue
			}
			req := &gpb.PlayTileRequest{
				LobbyId: lobbyID,
				UserId:  claims.UserID,
				Tile:    &gpb.Tile{Left: cmd.Tile.Left, Right: cmd.Tile.Right},
				Side:    cmd.Side,
			}
			response, err := gameSvc.Client.PlayTile(r.Context(), req)
			if err != nil {
				log.Printf("couldn't play tile to server: %v", err)
				continue
			}

			// Send response to WebSocket
			WSMsg := contracts.WSMessage{
				Type: "TODO",
				Data: response,
			}
			if err := connManager.SendMessage(claims.UserID, WSMsg); err != nil {
				log.Printf("couldt send message: %v", err)
			}

		case contracts.PassTurnCmd:
			// Pass cmd doesnt pass anything
			req := &gpb.PassTurnRequest{
				LobbyId: lobbyID,
				UserId:  claims.UserID,
			}
			response, err := gameSvc.Client.PassTurn(r.Context(), req)
			if err != nil {
				log.Printf("couldn't pass turn to server: %v", err)
				continue
			}
			// Send response to WebSocket
			WSMsg := contracts.WSMessage{
				Type: "TODO",
				Data: response,
			}

			if err := connManager.SendMessage(claims.UserID, WSMsg); err != nil {
				log.Printf("couldt send message: %v", err)
			}

			log.Printf("TODO: Handle of %s", playerMsg.Type)

		default:
			log.Printf("Unknown message type: %s", playerMsg.Type)
		}
	}
}
