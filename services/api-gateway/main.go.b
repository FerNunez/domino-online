package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"rebu/shared/env"
	"rebu/shared/messaging"
	"rebu/shared/tracing"
	"syscall"
	"time"
)

var (
	httpAddr    = env.GetString("HTTP_ADDR", ":8081")
	rabbitMqURI = env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")
)

func main() {
	sh, err := tracing.InitTracer(tracing.Config{
		ServiceName:    "api-gateway",
		Environment:    env.GetString("ENVIRONMENT", "development"),
		JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces"),
	})
	if err != nil {
		log.Fatalf("Failed to init tracer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer sh(ctx)

	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitmq.Close()

	mux := http.NewServeMux()

	mux.Handle("POST /trip/preview", tracing.WrapHandlerFunc(enableCORS(handleTripPreview), "/trip/preview"))
	mux.Handle("POST /trip/start", tracing.WrapHandlerFunc(enableCORS(handleTripStart), "/trip/start"))
	mux.Handle("POST /lobbies", tracing.WrapHandlerFunc(enableCORS(handleCreateLobby), "/lobby"))
	mux.Handle("POST /lobbies/{id}/join", tracing.WrapHandlerFunc(enableCORS(handleJoinLobby), "/lobby/join"))
	mux.Handle("POST /lobbies/{id}/start", tracing.WrapHandlerFunc(enableCORS(handleStartGame), "/lobby/start"))

	mux.Handle("/ws/drivers", tracing.WrapHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleDriversWebsocket(w, r, rabbitmq)
	}, "/trip/start"))
	mux.Handle("/ws/riders", tracing.WrapHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleRidersWebsocket(w, r, rabbitmq)
	}, "/trip/start"))
	mux.Handle("POST /webhook/stripe", tracing.WrapHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleStripeWebhook(w, r, rabbitmq)
	}, "/webhook/strip"))

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
