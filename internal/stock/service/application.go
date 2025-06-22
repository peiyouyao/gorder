package service

import (
	"context"

	"github.com/peiyouyao/gorder/common/metrics"

	"github.com/peiyouyao/gorder/stock/adapters"
	"github.com/peiyouyao/gorder/stock/app"
	"github.com/peiyouyao/gorder/stock/app/query"
	"github.com/peiyouyao/gorder/stock/infrastructure/intergration"
	"github.com/sirupsen/logrus"
)

func NewApplication(ctx context.Context) app.Application {
	stockRepo := adapters.NewMemoryStockRepository()
	stripeAPI := intergration.NewStripeAPI()
	logger := logrus.NewEntry(logrus.StandardLogger())
	metricsClient := metrics.TodoMetrics{}
	return app.Application{
		Commands: app.Commands{},
		Queries: app.Queries{
			CheckIfItemsInStock: query.NewCheckIfItemsInStockHandler(stockRepo, stripeAPI, logger, metricsClient),
			GetItems:            query.NewGetItemsHandler(stockRepo, logger, metricsClient),
		},
	}
}
