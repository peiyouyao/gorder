package service

import (
	"context"

	grpcClient "github.com/peiyouyao/gorder/common/client"
	"github.com/peiyouyao/gorder/common/metrics"
	"github.com/peiyouyao/gorder/payment/adapters"
	"github.com/peiyouyao/gorder/payment/app"
	"github.com/peiyouyao/gorder/payment/app/command"
	"github.com/peiyouyao/gorder/payment/domain"
	"github.com/peiyouyao/gorder/payment/infrastructure/processor"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func NewApplication(ctx context.Context) (app.Application, func()) {

	orderClinet, closeOrderClient, err := grpcClient.NewOrderGRPCClient(ctx)
	if err != nil {
		panic(err)
	}
	orderGRPC := adapters.NewOrderGRPC(orderClinet)
	// memoryProcessor := processor.NewInmemProcess()
	stripeProcessor := processor.NewStripeProcessor(viper.GetString("stripe-key"))

	return newApplication(orderGRPC, stripeProcessor), func() {
		_ = closeOrderClient()
	}
}

func newApplication(
	orderGRPC command.OrderService,
	processor domain.Processor,
) app.Application {

	logger := logrus.NewEntry(logrus.StandardLogger())
	metricsClient := metrics.TodoMetrics{}
	return app.Application{
		Commands: app.Commands{
			CreatePayment: command.NewCreatePaymentHandler(processor, orderGRPC, logger, metricsClient),
		},
	}
}
