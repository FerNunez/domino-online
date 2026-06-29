package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	//"rebu/services/payment-service/internal/service"
	"rebu/services/payment-service/internal/events"
	"rebu/services/payment-service/internal/infrastructure/stripe"
	"rebu/services/payment-service/internal/service"
	"rebu/services/payment-service/pkg/types"
	"rebu/shared/env"
	"rebu/shared/messaging"
	"rebu/shared/tracing"
)

func main() {
	sh, err := tracing.InitTracer(tracing.Config{
		ServiceName:    "payment-service",
		Environment:    env.GetString("ENVIRONMENT", "development"),
		JaegerEndpoint: env.GetString("JAEGER_ENDPOINT", "http://jaeger:14268/api/traces"),
	})
	if err != nil {
		log.Fatalf("Failed to init tracer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer sh(ctx)

	stripeKey := env.GetString("STRIPE_SECRET_KEY", "")
	if stripeKey == "" {
		log.Fatalf("STRIPE_SECRET_KEY is not set")
	}
	config := &types.PaymentConfig{
		StripeSecretKey: stripeKey,
		SuccessURL:      env.GetString("STRIPE_SUCCESS_URL", "http://localhost:3000?payment=success"),
		CancelURL:       env.GetString("STRIPE_CANCEL_URL", "http://localhost:3000?payment=cancel"),
	}

	// Processor
	stripclient := stripe.NewStripeClient(config)

	rabbitMqURI := env.GetString("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672")
	rabbitmq, err := messaging.NewRabbitMQ(rabbitMqURI)
	if err != nil {
		log.Fatalf("Failed to connect to rabbitmq: %v", err)
	}

	svc := service.NewPaymentService(stripclient)
	go func() {
		if err := events.NewTripConsumer(rabbitmq, svc).Listen(); err != nil {
			fmt.Println("Failed to consume trip")
			cancel()
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sigCh:
	case <-ctx.Done():
	}
	log.Println("Shutting down payment service...")

}
