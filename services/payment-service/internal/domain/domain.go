package domain

import (
	"context"
	"rebu/services/payment-service/pkg/types"
)

// Service is the business logic interface.
// The event consumer holds a Service and calls a CreatePaymentSession
type Service interface {
	// Converts the domain logic: tripID, userID, driverID into metadata that then calls the specific payment session
	// It is infrastructure agnostic
	CreatePaymentSession(ctx context.Context, tripID, userID, driverID string, amount int64, currency string) (*types.PaymentIntent, error)
}

// PaymentProcessor is the external payment provider interface
// stripClient implements this; any other provider would implement it too
type PaymentProcessor interface {
	// Creates a payment session by probably creating a client for stripe or paypal or what ever
	CreatePaymentSession(ctx context.Context, amount int64, currency string, metadata *map[string]string) (string, error)
}

//NOTE:
// Why two interfaces? Service speaks the domain language (tripID, userID, driverID). PaymentProcessor speaks the payment provider language (amount, currency, metadata). The service layer converts domain concepts to payment-provider concepts. If you swap Stripe for PayPal, you write a new PaymentProcessor implementation — Service and the consumer never change.
