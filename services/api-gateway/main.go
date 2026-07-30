package main

import (
	"context"
	"domino/shared/env"
	"domino/shared/messaging"
	"domino/shared/tracing"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	httpAddr    = env.GetString("HTTP_ADDR", ":8081")
	rabbitMqURI = env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
)
var connManager = messaging.NewConnectionManager()

func main() {
	// RabbitMQ connection
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitmq.Close()

	// Declare and bind new queue and consomme it
	queueName, err := rabbitmq.DeclareGatewayQueue([]string{"lobby.*", "game.*"}, messaging.DominoExchange)
	if err != nil {
		log.Fatal(err)
	}
	consumer := messaging.NewQueueConsumer(rabbitmq, connManager, queueName)
	if err := consumer.Start(); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("POST /auth/guest", tracing.WrapHandlerFunc(enableCORS(handleAuthGuest), "/auth/guest"))
	mux.Handle("POST /auth/register", tracing.WrapHandlerFunc(enableCORS(authMiddleware(handleAuthRegiter)), "/auth/register"))
	mux.Handle("GET /users/{id}", tracing.WrapHandlerFunc(enableCORS(handleGetUser), "/users"))
	mux.Handle("POST /lobbies", tracing.WrapHandlerFunc(enableCORS(authMiddleware(handleCreateLobby)), "/lobby"))
	mux.Handle("POST /lobbies/{id}/join", tracing.WrapHandlerFunc(enableCORS(authMiddleware(handleJoinLobby)), "/lobby/join"))
	mux.Handle("POST /lobbies/{id}/start", tracing.WrapHandlerFunc(enableCORS(authMiddleware(handleStartGame)), "/lobby/start"))

	mux.Handle("/lobbies/{id}/ws", tracing.WrapHandlerFunc(handleLobbyWebsocket, "/lobby/ws"))

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
