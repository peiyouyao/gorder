package main

import (
	"context"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/peiyouyao/gorder/common/broker"
	_ "github.com/peiyouyao/gorder/common/config"
	"github.com/peiyouyao/gorder/common/logging"
	"github.com/peiyouyao/gorder/common/server"
	"github.com/peiyouyao/gorder/common/tracing"
	"github.com/peiyouyao/gorder/payment/app"
	"github.com/peiyouyao/gorder/payment/infrastructure/consumer"
	"github.com/peiyouyao/gorder/payment/ports"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard
	logging.Init()
}

func main() {
	serviceName := viper.GetString("payment.service-name")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown, err := tracing.InitJaegerProvider(viper.GetString("jaeger.url"), serviceName)
	if err != nil {
		logrus.Fatal(err)
	}
	defer shutdown(ctx)

	application, cleanup := app.NewApplication(ctx)
	defer cleanup()

	ch, closeCh := broker.Connect(
		viper.GetString("rabbitmq.user"),
		viper.GetString("rabbitmq.password"),
		viper.GetString("rabbitmq.host"),
		viper.GetString("rabbitmq.port"),
	)
	defer func() {
		_ = ch.Close()
		_ = closeCh()
	}()

	go consumer.NewConsumer(application).Listen(ch)

	paymentHandler := ports.NewPaymentHandler(ch)
	server.RunHTTPServer(serviceName, paymentHandler.RegisterRoutes)
}
