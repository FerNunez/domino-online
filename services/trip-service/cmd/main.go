package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"rebu/services/trip-service/internal/infrastructure/events"
	grpchandler "rebu/services/trip-service/internal/infrastructure/grpc"
	"rebu/services/trip-service/internal/infrastructure/repository"
	"rebu/services/trip-service/internal/service"
	"rebu/shared/db"
	"rebu/shared/env"
	"rebu/shared/messaging"
	"rebu/shared/tracing"

	grpcserver "google.golang.org/grpc"
)

const grpcAddr = ":9093"

func main() {
	// 1. Initialice Tracer
	sh, err := tracing.InitTracer(tracing.Config{
		ServiceName:    "trip-service",
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
	mongoClient, err := db.NewMongoClient(ctx, db.NewMongoDefaultConfig())
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)
	mongoDb := db.GetDatabase(mongoClient, db.NewMongoDefaultConfig())

	// 3. Connect to RabbitMQ:
	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq.5672/")
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitmq.Close()

	// 4.  Wire the dependency graph (innermost first)
	repo := repository.NewMongoRepository(mongoDb)
	svc := service.NewService(repo)
	publisher := events.NewTripEventPublisher(rabbitmq)

	go func() {
		if err := events.NewDriverConsumer(rabbitmq, svc).Listen(); err != nil {
			log.Printf("Driver consumer error: %v", err)
			cancel()
		}
	}()

	go func() {
		if err := events.NewPaymentConsumer(rabbitmq, svc).Listen(); err != nil {
			log.Printf("Payment consumer error: %v", err)
			cancel()
		}
	}()

	// 6. Start the Grpc Server. Listen
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", grpcAddr, err)
	}
	server := grpcserver.NewServer(tracing.WithTracingInterceptors()...)
	grpchandler.NewGRPCHandler(server, svc, publisher)

	log.Printf("Trip service listening on %s", grpcAddr)
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
