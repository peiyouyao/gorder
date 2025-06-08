package service

import (
	"context"

	"github.com/PerryYao-GitHub/gorder/common/broker"
	grpcClient "github.com/PerryYao-GitHub/gorder/common/client"
	"github.com/PerryYao-GitHub/gorder/common/metrics"
	"github.com/PerryYao-GitHub/gorder/order/adapters"
	"github.com/PerryYao-GitHub/gorder/order/adapters/grpc"
	"github.com/PerryYao-GitHub/gorder/order/app"
	"github.com/PerryYao-GitHub/gorder/order/app/command"
	"github.com/PerryYao-GitHub/gorder/order/app/query"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewApplication(ctx context.Context) (app.Application, func()) {
	stockClient, closeStockClient, err := grpcClient.NewStockGRPCClient(ctx)
	if err != nil {
		panic(err)
	}
	ch, closeConn := broker.Connect(
		viper.GetString("rabbitmq.user"),
		viper.GetString("rabbitmq.password"),
		viper.GetString("rabbitmq.host"),
		viper.GetString("rabbitmq.port"),
	)
	stockGRPC := grpc.NewStockGRPC(stockClient)
	return newAppliction(ctx, stockGRPC, ch), func() {
		_ = closeStockClient()
		_ = ch.Close()
		_ = closeConn()
	}
}

func newAppliction(
	_ context.Context,
	stockGRPC query.StockService,
	ch *amqp.Channel,
) app.Application {
	orderRepo := adapters.NewMemoryOrderRepository()
	logger := logrus.NewEntry(logrus.StandardLogger())
	metricsClient := metrics.TodoMetrics{}

	return app.Application{
		Commands: app.Commands{
			CreateOrder: command.NewCreateOrderHandler(orderRepo, stockGRPC, ch, logger, metricsClient),
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
