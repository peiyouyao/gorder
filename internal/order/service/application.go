package service

import (
	"context"

	"github.com/PerryYao-GitHub/gorder/order/adapters"
	"github.com/PerryYao-GitHub/gorder/order/app"
)

func NewApplication(ctx context.Context) app.Application {
	orderRepo := adapters.NewMemoryOrderRepository()
	return app.Application{}
}
