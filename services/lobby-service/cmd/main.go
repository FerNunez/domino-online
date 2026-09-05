package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"domino/services/lobby-service/internal/infrastructure/events"
	grpchandler "domino/services/lobby-service/internal/infrastructure/grpc"
	"domino/services/lobby-service/internal/infrastructure/repository"
	"domino/services/lobby-service/internal/service"
	"domino/shared/db"
	"domino/shared/env"
	"domino/shared/messaging"
	"domino/shared/tracing"

	grpcserver "google.golang.org/grpc"
)

const grpcAddr = ":9092"

func main() {
	// 1. Initialice Tracer
	sh, err := tracing.InitTracer(tracing.Config{
		ServiceName:    "lobby-service",
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

	// 2. Connect to RabbitMQ:
	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq.5672/")
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitmq.Close()

	pub := events.NewLobbyEventPublisher(rabbitmq)

	// 3. Redis Client
	redisClient := db.NewRedisClient(db.NewRedisDefaultConfig())

	// 4. Wire the dependency graph (innermost first)
	repo := repository.NewRedisRepository(redisClient)
	svc := service.NewService(repo, pub)

	// 5. Consume PlayerConnected/PlayerDisconnected
	go func() {
		consumer, err := events.NewLobbyConsumer(rabbitmq, svc)
		if err != nil {
			log.Printf("Failed to create lobby consumer: %v", err)
			cancel()
			return
		}
		if err := consumer.Listen(); err != nil {
			log.Printf("consumer error: %v", err)
			cancel()
		}
	}()

	// 6. Start the Grpc Server. Listen
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", grpcAddr, err)
	}
	server := grpcserver.NewServer(tracing.WithTracingInterceptors()...)
	grpchandler.NewGRPCHandler(server, svc, pub)

	log.Printf("Lobby service listening on %s", grpcAddr)
	go func() {
		if err := server.Serve(lis); err != nil {
			log.Printf("gRPC server error: %v", err)
			cancel()
		}
	}()

	// 7. Wait for shutfown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sigCh:
		log.Println("Received shutdown signal")
	case <-ctx.Done():
	}

	log.Println("Shutting down lobby services...")
	server.GracefulStop() // wait gor in-flight RPCs to complete
}
