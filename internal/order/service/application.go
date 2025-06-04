package service

import (
	"context"

	"github.com/PerryYao-GitHub/gorder/common/metrics"
	"github.com/PerryYao-GitHub/gorder/order/adapters"
	"github.com/PerryYao-GitHub/gorder/order/app"
	"github.com/PerryYao-GitHub/gorder/order/app/command"
	"github.com/PerryYao-GitHub/gorder/order/app/query"
	"github.com/sirupsen/logrus"
)

func NewApplication(ctx context.Context) app.Application {
	orderRepo := adapters.NewMemoryOrderRepository()
	logger := logrus.NewEntry(logrus.StandardLogger())
	metricsClient := metrics.TodoMetrics{}
	return app.Application{
		Commands: app.Commands{
			CreateOrder: command.NewCreateOrderHandler(orderRepo, logger, metricsClient),
			UpdateOrder: command.NewUpdateOrderHandler(orderRepo, logger, metricsClient),
		},
		Queries: app.Queries{
			GetCustomerOrder: query.NewGetCustomerOrderHandler(
				orderRepo,
				logger,
				metricsClient,
			),
		},
	}
}
