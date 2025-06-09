package service

import (
	"context"

	grpcClient "github.com/PerryYao-GitHub/gorder/common/client"
	"github.com/PerryYao-GitHub/gorder/common/metrics"
	"github.com/PerryYao-GitHub/gorder/payment/adapters"
	"github.com/PerryYao-GitHub/gorder/payment/app"
	"github.com/PerryYao-GitHub/gorder/payment/app/command"
	"github.com/PerryYao-GitHub/gorder/payment/domain"
	"github.com/PerryYao-GitHub/gorder/payment/infrastructure/processor"
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

	return newApplication(ctx, orderGRPC, stripeProcessor), func() {
		_ = closeOrderClient()
	}
}

func newApplication(
	ctx context.Context,
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
