package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"rebu/shared/env"
	"rebu/shared/tracing"
	"syscall"
	"time"
)

var (
	httpAddr = env.GetString("HTTP_ADDR", ":8081")
)

func main() {

	mux := http.NewServeMux()

	mux.Handle("POST /lobbies", tracing.WrapHandlerFunc(enableCORS(handleCreateLobby), "/lobby"))
	mux.Handle("POST /lobbies/{id}/join", tracing.WrapHandlerFunc(enableCORS(handleJoinLobby), "/lobby/join"))
	mux.Handle("POST /lobbies/{id}/start", tracing.WrapHandlerFunc(enableCORS(handleStartGame), "/lobby/start"))

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
