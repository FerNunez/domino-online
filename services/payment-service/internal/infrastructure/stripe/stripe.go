package stripe

import (
	"context"
	"fmt"
	"rebu/services/payment-service/pkg/types"

	stripelib "github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
)

type stripeClient struct {
	config *types.PaymentConfig
}

// NewStripeClient sets the global Stripe API key and returns a PaymentProcessor
// The global key assignment is safe because stripe.key is read-only afgter initialisation.
func NewStripeClient(config *types.PaymentConfig) *stripeClient {
	stripelib.Key = config.StripeSecretKey
	return &stripeClient{config: config}
}

func (s *stripeClient) CreatePaymentSession(
	_ context.Context,
	amount int64,
	currency string,
	metadata *map[string]string,
) (string, error) {
	params := &stripelib.CheckoutSessionParams{
		CancelURL: stripelib.String(s.config.CancelURL),
		LineItems: []*stripelib.CheckoutSessionLineItemParams{
			{
				PriceData: &stripelib.CheckoutSessionLineItemPriceDataParams{
					Currency: stripelib.String(currency),
					ProductData: &stripelib.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripelib.String("Ride Payment"),
					},
					UnitAmount: stripelib.Int64(amount),
				},
				Quantity: stripelib.Int64(1),
			},
		},
		Metadata:   *metadata,
		Mode:       stripelib.String(string(stripelib.CheckoutSessionModePayment)),
		SuccessURL: stripelib.String(s.config.SuccessURL),
	}

	result, err := session.New(params)
	if err != nil {
		return "", fmt.Errorf("stripe session creation failed: %w", err)
	}

	return result.ID, nil
}
