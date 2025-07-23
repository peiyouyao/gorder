package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/peiyouyao/gorder/common/broker"
	grpcClient "github.com/peiyouyao/gorder/common/client"
	_ "github.com/peiyouyao/gorder/common/config"
	"github.com/peiyouyao/gorder/common/logging"
	"github.com/peiyouyao/gorder/common/tracing"
	"github.com/peiyouyao/gorder/kitchen/adapters"
	"github.com/peiyouyao/gorder/kitchen/infrastructure/consumer"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	logging.Init()
}

func main() {
	serviceName := viper.GetString("kitchen.service-name")

	ctx, cancal := context.WithCancel(context.Background())
	defer cancal()

	shutdown, err := tracing.InitJaegerProvider(viper.GetString("jaeger.url"), serviceName)
	if err != nil {
		logrus.Fatal(err)
	}
	defer shutdown(ctx)

	orderGRPCCli, closeFn, err := grpcClient.NewOrderGRPCClient(ctx)
	if err != nil {
		logrus.Fatal(err)
	}
	defer closeFn()

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

	orderGRPC := adapters.NewOrderGRPC(orderGRPCCli)
	go consumer.NewConsumer(orderGRPC).Listen(ch)

	// ^C 退出 kitchen
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-sigs
		logrus.Info("Receive signal, exiting ...")
		os.Exit(0)
	}()
	logrus.Println("To exit, press Ctrl+C")
	select {}
}
