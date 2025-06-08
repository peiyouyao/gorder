package domain

import (
	"context"

	"github.com/PerryYao-GitHub/gorder/common/genproto/orderpb"
)

type Processor interface {
	CreatePaymentLink(context.Context, *orderpb.Order) (string, error)
}
