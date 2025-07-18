package domain

import (
	"context"

	"github.com/peiyouyao/gorder/common/entity"
)

type Processor interface {
	CreatePaymentLink(context.Context, *entity.Order) (string, error)
}

type Order struct {
	ID          string
	CustomerID  string
	Status      string
	PaymentLink string
	Items       []*entity.Item
}
