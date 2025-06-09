package processor

import (
	"context"
	"encoding/json"

	"github.com/PerryYao-GitHub/gorder/common/genproto/orderpb"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
)

type StripeProcessor struct {
	apiKey string
}

func NewStripeProcessor(apiKey string) *StripeProcessor {
	if apiKey == "" {
		panic("empty api key")
	}
	stripe.Key = apiKey
	return &StripeProcessor{apiKey: apiKey}
}

var (
	sucessURL = "http://localhost:8282"
)

func (s StripeProcessor) CreatePaymentLink(ctx context.Context, order *orderpb.Order) (string, error) {

	var items []*stripe.CheckoutSessionLineItemParams
	for _, item := range order.Items {
		items = append(items, &stripe.CheckoutSessionLineItemParams{
			Price: stripe.String("price_1RXLrqPqGUzmzBMUyWDWprnO"),
			// Price:    stripe.String(string(item.PriceID)),
			Quantity: stripe.Int64(int64(item.Quantity)),
		})
	}

	marshalledItems, _ := json.Marshal(order.Items)
	metadata := map[string]string{
		"orderID":    order.ID,
		"customerID": order.CustomerID,
		"status":     order.Status,
		"items":      string(marshalledItems),
	}

	params := &stripe.CheckoutSessionParams{
		Metadata:   metadata,
		LineItems:  items,
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(string(sucessURL)),
	}

	res, err := session.New(params)
	if err != nil {
		return "", err
	}
	return res.URL, nil
}
