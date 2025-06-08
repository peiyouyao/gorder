package command

import (
	"context"

	"github.com/PerryYao-GitHub/gorder/common/genproto/orderpb"
)

type OrderService interface {
	UpdateOrder(ctx context.Context, order *orderpb.Order) error
}
