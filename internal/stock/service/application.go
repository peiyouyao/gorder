package service

import (
	"context"

	"github.com/peiyouyao/gorder/common/metrics"

	"github.com/peiyouyao/gorder/stock/adapters"
	"github.com/peiyouyao/gorder/stock/app"
	"github.com/peiyouyao/gorder/stock/app/query"
	"github.com/peiyouyao/gorder/stock/infrastructure/intergration"
	"github.com/peiyouyao/gorder/stock/infrastructure/persistent"
	"github.com/sirupsen/logrus"
)

func NewApplication(ctx context.Context) app.Application {
	db := persistent.NewMySQL()
	stockRepo := adapters.NewMySQLStockRepository(db)
	stripeAPI := intergration.NewStripeAPI()
	logger := logrus.NewEntry(logrus.StandardLogger())
	metricsClient := metrics.NoMetrics{}
	return app.Application{
		Commands: app.Commands{},
		Queries: app.Queries{
			CheckIfItemsInStock: query.NewCheckIfItemsInStockHandler(stockRepo, stripeAPI, logger, metricsClient),
			GetItems:            query.NewGetItemsHandler(stockRepo, logger, metricsClient),
		},
	}
}
