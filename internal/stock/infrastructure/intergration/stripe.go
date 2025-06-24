package intergration

import (
	"context"
	"fmt"

	_ "github.com/peiyouyao/gorder/common/logging"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/product"
)

type StripeAPI struct {
	apiKey string
}

func NewStripeAPI() *StripeAPI {
	key := viper.GetString("stripe-key")
	if key == "" {
		logrus.Fatal("empty key")
	}
	return &StripeAPI{apiKey: key}
}

func (s *StripeAPI) GetPriceByProductID(ctx context.Context, pid string) (string, error) {
	stripe.Key = s.apiKey
	result, err := product.Get(pid, &stripe.ProductParams{})
	if err != nil {
		return "", err
	}
	if result.DefaultPrice == nil {
		return "", fmt.Errorf("default price not found for product %s", pid)
	}
	return result.DefaultPrice.ID, nil
}
