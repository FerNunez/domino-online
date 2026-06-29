package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"rebu/shared/env"
	"rebu/shared/messaging"
	"rebu/shared/tracing"
	"syscall"

	grpcserver "google.golang.org/grpc"
)

const grpcAddr = ":9092"

func main() {
	sh, err := tracing.InitTracer(tracing.Config{
		ServiceName:    "driver-service",
		Environment:    env.GetString("ENVIRONMENT", "development"),
		JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces"),
	})
	if err != nil {
		log.Fatalf("Failed to init tracer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer sh(ctx)

	svc := NewService()
	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/")

	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitmq.Close()

	// Start trip Consumer
	//consuming in find_available_drivers_queue -> request drivers a trip, or tell user no drivers
	go func() {
		if err := NewTripConsumer(rabbitmq, svc).Listen(); err != nil {
			log.Printf("Trip consumer error: %v", err)
			cancel()
		}
	}()

	// Grpc Server & Handler
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", grpcAddr, err)
	}
	server := grpcserver.NewServer(tracing.WithTracingInterceptors()...)
	NewGrpcHandler(server, svc)
	log.Printf("Driver service listening on %s", grpcAddr)
	go func() {
		if err := server.Serve(lis); err != nil {
			log.Printf("grpc server error: %v", err)
			cancel()
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sigCh:
	case <-ctx.Done():
	}

	log.Println("Shutting down driver service...")
	server.GracefulStop()
}
