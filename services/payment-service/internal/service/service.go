package service

import (
	"context"
	"fmt"
	"rebu/services/payment-service/internal/domain"
	"rebu/services/payment-service/pkg/types"
	"time"

	"github.com/google/uuid"
)

// The service layer converts domain concepts to payment-provider concepts
type paymentService struct {
	processor domain.PaymentProcessor
}

func NewPaymentService(p domain.PaymentProcessor) *paymentService {
	return &paymentService{
		processor: p,
	}
}

func (s *paymentService) CreatePaymentSession(ctx context.Context, tripID, userID, driverID string, amount int64, currency string) (*types.PaymentIntent, error) {
	// Metadata is stored in the Stripe session and returned in the webhook payload.
	// This allows the API gateway webhook hgandler to reconstruct context (which trip, which user) without a database lookup
	metadata := map[string]string{
		"trip_id":   tripID,
		"user_id":   userID,
		"driver_id": driverID,
	}

	sessionID, err := s.processor.CreatePaymentSession(ctx, amount, currency, &metadata)
	if err != nil {
		return nil, fmt.Errorf("creating payment session: %w", err)
	}

	return &types.PaymentIntent{
		ID:              uuid.New().String(),
		TripID:          tripID,
		UserID:          userID,
		Amount:          amount,
		Currency:        currency,
		StripeSessionID: sessionID,
		CreatedAt:       time.Now(),
	}, nil
}
