package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	grpchandler "domino/services/lobby-service/internal/infrastructure/grpc"
	"domino/services/lobby-service/internal/infrastructure/repository"
	"domino/services/lobby-service/internal/service"
	"domino/shared/db"
	"domino/shared/env"
	"domino/shared/tracing"

	grpcserver "google.golang.org/grpc"
)

const grpcAddr = ":9092"

func main() {
	// 1. Initialice Tracer
	sh, err := tracing.InitTracer(tracing.Config{
		ServiceName:    "lobby-service",
		Environment:    env.GetString("ENVIRONMENT", "development"),
		JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces"),
	})
	if err != nil {
		log.Fatalf("Failed to init tracer: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer sh(ctx)

	// 3. Redis Client
	redisClient := db.NewRedisClient(db.NewRedisDefaultConfig())

	// 4.  Wire the dependency graph (innermost first)
	repo := repository.NewRedisRepository(redisClient)
	svc := service.NewService(repo)

	// 6. Start the Grpc Server. Listen
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", grpcAddr, err)
	}
	server := grpcserver.NewServer(tracing.WithTracingInterceptors()...)
	grpchandler.NewGRPCHandler(server, svc)

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

	log.Println("Shutting down trip services...")
	server.GracefulStop() // wait gor in-flight RPCs to complete

}
