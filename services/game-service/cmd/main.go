package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"domino/services/game-service/internal/infrastructure/events"
	"domino/services/game-service/internal/service"
	"domino/shared/env"
	"domino/shared/messaging"
	"domino/shared/tracing"
	//grpcserver "google.golang.org/grpc"
)

//const grpcAddr = ":9092"

func main() {
	// 1. Initialice Tracer
	sh, err := tracing.InitTracer(tracing.Config{
		ServiceName:    "game-service",
		Environment:    env.GetString("ENVIRONMENT", "development"),
		JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces"),
	})
	if err != nil {
		log.Fatalf("Failed to init tracer: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer sh(ctx)

	// 2. Connect to RabbitMQ:
	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq.5672/")
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitmq.Close()

	// 3. Redis Client
	//redisClient := db.NewRedisClient(db.NewRedisDefaultConfig())

	// 4.  Wire the dependency graph (innermost first)
	//repo := repository.NewRedisRepository(redisClient)
	svc := service.NewService(nil) // TODO: wire real repo

	go func() {
		if err := events.NewGameConsumer(rabbitmq, svc).Listen(); err != nil {
			log.Printf(" consumer error: %v", err)
			cancel()
		}
	}()

	// 6. Start the Grpc Server. Listen
	// lis, err := net.Listen("tcp", grpcAddr)
	// if err != nil {
	// 	log.Fatalf("Failed to listen on %s: %v", grpcAddr, err)
	// }
	//server := grpcserver.NewServer(tracing.WithTracingInterceptors()...)
	//grpchandler.NewGRPCHandler(server, svc, pub)
	// log.Printf("Game service listening on %s", grpcAddr)
	// go func() {
	// 	if err := server.Serve(lis); err != nil {
	// 		log.Printf("gRPC server error: %v", err)
	// 		cancel()
	// 	}
	// }()

	// 7. Wait for shutfown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sigCh:
		log.Println("Received shutdown signal")
	case <-ctx.Done():
	}

	log.Println("Shutting down game services...")
	//server.GracefulStop() // wait gor in-flight RPCs to complete
}
