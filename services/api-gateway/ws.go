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
	gpb "domino/shared/proto/game"
)

// connManager is a package-level singleton shared by all WebSocket handlers
// All handlers in the same process share one connection map, enablig cross-handler message delivery(eg. a RabbitMQ consumer pushing to a rider whose connection was registered by handleRidersWebsocket

func handleLobbyWebsocket(w http.ResponseWriter, r *http.Request, rb *messaging.RabbitMQ) {
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

	conn, err := connManager.Upgrade(w, r)

	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	connManager.AddToLobby(lobbyID, claims.UserID)
	defer connManager.RemoveFromLobby(lobbyID, claims.UserID)
	connManager.Add(claims.UserID, conn)
	defer connManager.Remove(claims.UserID)

	for _, q := range []string{
		messaging.NotifyLobby,
	} {
		fmt.Println("consumming ", q)
		consumer := messaging.NewQueueConsumer(rb, connManager, q)
		if err := consumer.Start(); err != nil {
			log.Printf("Failed to start consumer for %s: %v", q, err)
		}
	}

	gameSvc, err := grpc_clients.NewGameServiceClient()
	if err != nil {
		fmt.Printf("couldnt reach game service: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Read loop
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		type playerMessage struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}

		var playerMsg playerMessage
		if err := json.Unmarshal(message, &playerMsg); err != nil {
			log.Printf("Error unmarshaling player message: %v", err)
			continue
		}
		fmt.Printf("got message %v\n", playerMsg.Data)
		fmt.Printf("got message type %v\n", playerMsg.Type)

		// Handle the different message type from websocket client
		switch playerMsg.Type {
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
				log.Printf("Error unmarshaling driver message: %v", err)
				continue
			}
			WSMsg := contracts.WSMessage{
				Type: "response",
				Data: response,
			}

			if err := connManager.SendMessage(claims.UserID, WSMsg); err != nil {
				log.Printf("couldt send message: %v", err)
			}

		// case contracts.DriverCmdLocation:
		// 	// Handle driver location update in the future
		// 	continue
		// case contracts.DriverCmdTripAccept, contracts.DriverCmdTripDecline:
		// 	// Forward the message to RabbitMQ
		// 	if err := rb.PublishMessage(ctx, driverMsg.Type, contracts.AmqpMessage{
		// 		OwnerID: userID,
		// 		Data:    driverMsg.Data,
		// 	}); err != nil {
		// 		log.Printf("Error publishing message to RabbitMQ: %v", err)
		// 	}
		default:
			log.Printf("Unknown message type: %s", playerMsg.Type)
		}
	}

}
