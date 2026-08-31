package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"domino/shared/env"
	"domino/shared/messaging"
	"domino/shared/tracing"
)

var (
	httpAddr    = env.GetString("HTTP_ADDR", ":8081")
	rabbitMqURI = env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
)
var connManager = NewConnectionManager()

func main() {
	if historySvcErr != nil {
		log.Fatal(historySvcErr)
	}
	defer historySvc.Close()

	// RabbitMQ connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()

	// Declare and bind new queue and consomme it
	queueName, err := rabbitmq.DeclareAndBindExclusiveQueue([]string{"lobby.*", "game.*"}, messaging.DominoExchange)
	if err != nil {
		log.Fatal(err)
	}
	consumer := NewWebsocketEventConsumer(rabbitmq, connManager, queueName)
	if err := consumer.Start(); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("POST /auth/guest", tracing.WrapHandlerFunc(enableCORS(handleAuthGuest), "/auth/guest"))
	mux.Handle("POST /auth/register", tracing.WrapHandlerFunc(enableCORS(authMiddleware(handleAuthRegiter)), "/auth/register"))
	mux.Handle("GET /users/{id}", tracing.WrapHandlerFunc(enableCORS(handleGetUser), "/users"))
	mux.Handle("POST /lobbies", tracing.WrapHandlerFunc(enableCORS(authMiddleware(handleCreateLobby)), "/lobby"))
	mux.Handle("GET /lobbies/{id}", tracing.WrapHandlerFunc(enableCORS(authMiddleware(handleGetLobby)), "/lobby/get"))
	mux.Handle("POST /lobbies/{id}/join", tracing.WrapHandlerFunc(enableCORS(authMiddleware(handleJoinLobby)), "/lobby/join"))
	mux.Handle("POST /lobbies/{id}/start", tracing.WrapHandlerFunc(enableCORS(authMiddleware(handleStartGame)), "/lobby/start"))
	mux.Handle("GET /rounds/{id}/actions", tracing.WrapHandlerFunc(enableCORS(authMiddleware(handleGetRoundActions)), "/rounds/actions"))
	mux.Handle("GET /games/{id}/history", tracing.WrapHandlerFunc(enableCORS(authMiddleware(handleGetGameHistory)), "/games/history"))
	mux.Handle("GET /players/me/games", tracing.WrapHandlerFunc(enableCORS(authMiddleware(handleGetPlayerGames)), "/players/games"))

	// enables OPTIONS 'path' for each /path that does preflight: those using authMiddleware
	corsPreflight := enableCORS(func(w http.ResponseWriter, r *http.Request) {})
	for _, path := range []string{"/auth/guest", "/auth/register", "/users/{id}", "/lobbies", "/lobbies/{id}", "/lobbies/{id}/join", "/lobbies/{id}/start", "/rounds/{id}/actions", "/games/{id}/history", "/players/me/games"} {
		mux.Handle("OPTIONS "+path, corsPreflight)
	}

	mux.Handle("/lobbies/{id}/ws", tracing.WrapHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleLobbyWebsocket(w, r, rabbitmq)
	}, "/lobby/ws"))

	server := &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}

	serverErrs := make(chan error, 1)
	go func() {
		log.Printf("API gateway listening on %s", httpAddr)
		serverErrs <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrs:
		log.Printf("Server error: %v", err)
	case sig := <-shutdown:
		log.Printf("Received %v, shutting down...", sig)
		shutdownCtx, sshutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer sshutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			server.Close()
		}
	}
}
