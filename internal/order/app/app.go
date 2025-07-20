package app

import (
	"context"
	"fmt"
	"time"

	"github.com/peiyouyao/gorder/common/broker"
	grpcClient "github.com/peiyouyao/gorder/common/client"
	"github.com/peiyouyao/gorder/common/metrics"
	"github.com/peiyouyao/gorder/order/adapters"
	"github.com/peiyouyao/gorder/order/adapters/grpc"
	"github.com/peiyouyao/gorder/order/app/command"
	"github.com/peiyouyao/gorder/order/app/query"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct {
	CreateOrder command.CreateOrderHandler
	UpdateOrder command.UpdateOrderHandler
}

type Queries struct {
	GetCustomerOrder query.GetCustomerOrderHandler
}

func NewApplication(ctx context.Context) (Application, func()) {
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
) Application {
	mongoCli := newMongoClient()
	orderRepo := adapters.NewOrderRepositoryMongo(mongoCli)
	logger := logrus.NewEntry(logrus.StandardLogger())
	metrics := metrics.NewPrometheusMetricsClient(&metrics.PrometheusMetricsClientConfig{
		Host:        viper.GetString("order.metrics-addr"),
		ServiceName: viper.GetString("order.service-name"),
	})

	return Application{
		Commands: Commands{
			CreateOrder: command.NewCreateOrderHandler(orderRepo, stockGRPC, ch, logger, metrics),
			UpdateOrder: command.NewUpdateOrderHandler(orderRepo, logger, metrics),
		},
		Queries: Queries{
			GetCustomerOrder: query.NewGetCustomerOrderHandler(orderRepo, logger, metrics),
		},
	}
}

func newMongoClient() *mongo.Client {
	uri := fmt.Sprintf(
		"mongodb://%s:%s@%s:%s",
		viper.GetString("mongo.user"),
		viper.GetString("mongo.password"),
		viper.GetString("mongo.host"),
		viper.GetString("mongo.port"),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	c, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	if err = c.Ping(ctx, readpref.Primary()); err != nil {
		panic(err)
	}
	return c
}
