package main

import (
	"github.com/PerryYao-GitHub/gorder/common/broker"
	"github.com/PerryYao-GitHub/gorder/common/config"
	"github.com/PerryYao-GitHub/gorder/common/logging"
	"github.com/PerryYao-GitHub/gorder/common/server"
	"github.com/PerryYao-GitHub/gorder/payment/infrastructure/consumer"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	logging.Init()
	if err := config.NewViperConfig(); err != nil {
		logrus.Fatal(err)
	}
}

func main() {
	serviceType := viper.GetString("payment.server-to-run")

	ch, closeConn := broker.Connect(
		viper.GetString("rabbitmq.user"),
		viper.GetString("rabbitmq.password"),
		viper.GetString("rabbitmq.host"),
		viper.GetString("rabbitmq.port"),
	)
	defer func() {
		_ = ch.Close()
		_ = closeConn()
	}()

	go consumer.NewConsumer().Listen(ch)

	paymentHandler := NewPaymentHandler()
	switch serviceType {
	case "http":
		server.RunHTTPServer(viper.GetString("payment.service-name"), paymentHandler.RegisterRoutes)
	case "grpc":
		logrus.Panic("unsupported server type: grpc")
	default:
		logrus.Panic("unreachable code")
	}
}
