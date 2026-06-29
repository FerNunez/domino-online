package types

import "time"

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusSuccess   PaymentStatus = "success"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusCancelled PaymentStatus = "cancelled"
)

// Payment is a placeholder for a future persistencemodel
// NOTE: Not yet used. Defined here to indicate the intedended direction
type Payment struct {
	ID              string        `json:"id"`
	TripID          string        `json:"trip_id"`
	UserID          string        `json:"user_id"`
	Amount          int64         `json:"amount"` // in cents
	Currency        string        `json:"currency"`
	Status          PaymentStatus `json:"status"`
	StripeSessionID string        `json:"stipe_session_id"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}

// PaymentIntent is the result fo a successful Stripe session Creation
type PaymentIntent struct {
	ID              string    `json:"id"`
	TripID          string    `json:"trip_id"`
	UserID          string    `json:"user_id"`
	Amount          int64     `json:"amount"`
	Currency        string    `json:"currency"`
	StripeSessionID string    `json:"stipe_session_id"`
	CreatedAt       time.Time `json:"created_at"`
}

// PaymentConfig holds the Stripe credentials and redirect URLs
type PaymentConfig struct {
	StripeSecretKey     string
	StripeWebhookSecret string
	Currency            string
	SuccessURL          string
	CancelURL           string
}
