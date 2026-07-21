package main

import (
	"domino/shared/jwt"
	"domino/shared/messaging"
	"fmt"
	"log"
	"net/http"
)

// connManager is a package-level singleton shared by all WebSocket handlers
// All handlers in the same process share one connection map, enablig cross-handler message delivery(eg. a RabbitMQ consumer pushing to a rider whose connection was registered by handleRidersWebsocket
var connManager = messaging.NewConnectionManager()

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
		consumer := messaging.NewQueueConsumer(rb, connManager, q)
		if err := consumer.Start(); err != nil {
			log.Printf("Failed to start consumer for %s: %v", q, err)
		}
	}

	// Read loop
	// Keeps the handler goroutine alive and drains any client messages
	// Riders do not currently send messages; the loop exists on WebSocket close
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}

}
