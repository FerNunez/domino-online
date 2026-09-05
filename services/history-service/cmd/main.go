package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	historyevents "domino/services/history-service/internal/infrastructure/events"
	grpchandler "domino/services/history-service/internal/infrastructure/grpc"
	"domino/services/history-service/internal/infrastructure/repository"
	"domino/shared/db"
	"domino/shared/env"
	"domino/shared/messaging"
	"domino/shared/tracing"

	grpcserver "google.golang.org/grpc"
)

const grpcAddr = ":9094"

func main() {
	// 1. Initialize Tracer
	sh, err := tracing.InitTracer(tracing.Config{
		ServiceName:    "history-service",
		Environment:    env.GetString("ENVIRONMENT", "development"),
		OTLPEndpoint: env.GetString("OTLP_ENDPOINT", "jaeger:4317"),
	})
	if err != nil {
		log.Fatalf("Failed to init tracer: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if err := sh(shutdownCtx); err != nil {
			log.Printf("tracer shutdown error: %v", err)
		}
	}()

	// 2. Connect to Postgres
	sqlQueries, err := db.NewSQLQueries(ctx, db.NewSQLDefaultConfig())
	if err != nil {
		log.Fatalf("Failed to connect to Queries: %v", err)
	}

	// 3. Connect to RabbitMQ
	rmq, err := messaging.NewRabbitMQ(env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq.5672/"))
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rmq.Close()

	// 4. Wire the dependency graph (innermost first)
	repo := repository.NewSQLRepository(sqlQueries)

	consumer, err := historyevents.NewHistoryConsumer(rmq, repo)
	if err != nil {
		log.Fatalf("Failed to create history consumer: %v", err)
	}
	go func() {
		if err := consumer.Listen(); err != nil {
			log.Fatalf("Failed to listen for history events: %v", err)
		}
	}()

	// 5. Start the Grpc Server. Listen
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", grpcAddr, err)
	}
	server := grpcserver.NewServer(tracing.WithTracingInterceptors()...)
	grpchandler.NewGRPCHandler(server, repo)

	log.Printf("History service listening on %s", grpcAddr)
	go func() {
		if err := server.Serve(lis); err != nil {
			log.Printf("gRPC server error: %v", err)
			cancel()
		}
	}()

	// 6. Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sigCh:
		log.Println("Received shutdown signal")
	case <-ctx.Done():
	}

	log.Println("Shutting down history service...")
	server.GracefulStop() // wait for in-flight RPCs to complete
}
