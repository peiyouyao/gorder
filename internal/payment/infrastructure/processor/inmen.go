package processor

import (
	"context"

	"github.com/peiyouyao/gorder/common/entity"
)

// stub
// impl domain.Processor
type InmemProcessor struct {
}

func NewInmemProcess() *InmemProcessor {

	return &InmemProcessor{}
}

func (i InmemProcessor) CreatePaymentLink(ctx context.Context, order *entity.Order) (string, error) {

	return "inmem-payment-link", nil
}
