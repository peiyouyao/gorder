package app

import (
	"context"

	grpcClient "github.com/peiyouyao/gorder/common/client"
	"github.com/peiyouyao/gorder/common/metrics"
	"github.com/peiyouyao/gorder/payment/adapters"
	"github.com/peiyouyao/gorder/payment/app/command"
	"github.com/peiyouyao/gorder/payment/domain"
	"github.com/peiyouyao/gorder/payment/infrastructure/processor"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Application struct {
	Commands Commands
}

type Commands struct {
	CreatePayment command.CreatePaymentHandler
}

func NewApplication(ctx context.Context) (Application, func()) {

	orderClinet, closeOrderClient, err := grpcClient.NewOrderGRPCClient(ctx)
	if err != nil {
		panic(err)
	}
	orderGRPC := adapters.NewOrderGRPC(orderClinet)
	p := processor.NewStripeProcessor(viper.GetString("stripe-key"))

	return newApplication(orderGRPC, p), func() {
		_ = closeOrderClient()
	}
}

func newApplication(
	orderGRPC command.OrderService,
	processor domain.Processor,
) Application {

	logger := logrus.NewEntry(logrus.StandardLogger())
	metricsClient := metrics.NoMetrics{}
	return Application{
		Commands: Commands{
			CreatePayment: command.NewCreatePaymentHandler(processor, orderGRPC, logger, metricsClient),
		},
	}
}
