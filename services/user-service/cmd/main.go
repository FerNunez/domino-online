package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpchandler "domino/services/user-service/internal/infrastructure/grpc"
	"domino/services/user-service/internal/infrastructure/repository"
	"domino/services/user-service/internal/service"
	"domino/shared/db"
	"domino/shared/env"
	"domino/shared/tracing"

	grpcserver "google.golang.org/grpc"
)

const grpcAddr = ":9091"

func main() {
	// 1. Initialice Tracer
	sh, err := tracing.InitTracer(tracing.Config{
		ServiceName:    "user-service",
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

	// 2. Connect to MongoDB
	sqlQueries, err := db.NewSQLQueries(ctx, db.NewSQLDefaultConfig())
	if err != nil {
		log.Fatalf("Failed to connect to Queries: %v", err)
	}
	// TODO: Add disconnect?

	// 4.  Wire the dependency graph (innermost first)
	repo := repository.NewSqlRepository(sqlQueries)
	svc := service.NewService(repo)

	// 6. Start the Grpc Server. Listen
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", grpcAddr, err)
	}
	server := grpcserver.NewServer(tracing.WithTracingInterceptors()...)
	grpchandler.NewGRPCHandler(server, svc)

	log.Printf("User service listening on %s", grpcAddr)
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

	log.Println("Shutting down user services...")
	server.GracefulStop() // wait gor in-flight RPCs to complete
}
