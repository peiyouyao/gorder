package app

import (
	"context"

	"github.com/peiyouyao/gorder/common/metrics"
	"github.com/spf13/viper"

	"github.com/peiyouyao/gorder/stock/adapters"
	"github.com/peiyouyao/gorder/stock/app/query"
	"github.com/peiyouyao/gorder/stock/infrastructure/intergration"
	"github.com/peiyouyao/gorder/stock/infrastructure/persistent"
	"github.com/sirupsen/logrus"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct{}

type Queries struct {
	CheckIfItemsInStock query.CheckIfItemsInStockHandler
	GetItems            query.GetItemsHandler
}

func NewApplication(ctx context.Context) Application {
	db := persistent.NewMySQL()
	stockRepo := adapters.NewStockRepositoryMySQL(db)
	stripeAPI := intergration.NewStripeAPI()
	logger := logrus.NewEntry(logrus.StandardLogger())
	metrics := metrics.NewPrometheusMetricsClient(&metrics.PrometheusMetricsClientConfig{
		Host:        viper.GetString("stock.metrics-addr"),
		ServiceName: viper.GetString("stock.service-name"),
	})
	return Application{
		Commands: Commands{},
		Queries: Queries{
			CheckIfItemsInStock: query.NewCheckIfItemsInStockHandler(stockRepo, stripeAPI, logger, metrics),
			GetItems:            query.NewGetItemsHandler(stockRepo, logger, metrics),
		},
	}
}
