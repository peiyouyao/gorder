package processor

import (
	"context"

	"github.com/peiyouyao/gorder/common/genproto/orderpb"
)

// stub
type InmemProcessor struct {
}

func NewInmemProcess() *InmemProcessor {

	return &InmemProcessor{}
}

func (i InmemProcessor) CreatePaymentLink(ctx context.Context, order *orderpb.Order) (string, error) {

	return "inmem-payment-link", nil
}
