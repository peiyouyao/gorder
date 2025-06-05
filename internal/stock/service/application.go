package service

import (
	"context"

	"github.com/PerryYao-GitHub/gorder/common/metrics"

	"github.com/PerryYao-GitHub/gorder/stock/adapters"
	"github.com/PerryYao-GitHub/gorder/stock/app"
	"github.com/PerryYao-GitHub/gorder/stock/app/query"
	"github.com/sirupsen/logrus"
)

func NewApplication(ctx context.Context) app.Application {
	stockRepo := adapters.NewMemoryStockRepository()
	logger := logrus.NewEntry(logrus.StandardLogger())
	metricsClient := metrics.TodoMetrics{}
	return app.Application{
		Commands: app.Commands{},
		Queries: app.Queries{
			CheckIfItemsInStock: query.NewCheckIfItemsInStockHandler(stockRepo, logger, metricsClient),
			GetItems:            query.NewGetItemsHandler(stockRepo, logger, metricsClient),
		},
	}
}
