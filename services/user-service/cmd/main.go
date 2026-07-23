package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

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
		JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces"),
	})
	if err != nil {
		log.Fatalf("Failed to init tracer: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer sh(ctx)

	//2. Connect to MongoDB
	sqlQueries, err := db.NewSqlQueries(ctx, db.NewSQLDefaultConfig())
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
