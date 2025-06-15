package main

import (
	"context"

	_ "github.com/peiyouyao/gorder/common/config"
	"github.com/peiyouyao/gorder/common/discovery"
	"github.com/peiyouyao/gorder/common/genproto/stockpb"
	"github.com/peiyouyao/gorder/common/logging"
	"github.com/peiyouyao/gorder/common/server"
	"github.com/peiyouyao/gorder/common/tracing"
	"github.com/peiyouyao/gorder/stock/ports"
	"github.com/peiyouyao/gorder/stock/service"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func init() {
	logging.Init()
}

func main() {
	serviceName := viper.GetString("stock.service-name")
	serverType := viper.GetString("stock.server-to-run")

	ctx, cancal := context.WithCancel(context.Background())
	defer cancal()

	shutdown, err := tracing.InitJaegerProvider(viper.GetString("jaeger.url"), serviceName)
	if err != nil {
		logrus.Fatal(err)
	}
	defer shutdown(ctx)

	application := service.NewApplication(ctx)

	deregisterFn, err := discovery.RegisterToConsul(ctx, serviceName)
	if err != nil {
		logrus.Fatal(err)
	}
	defer func() {
		_ = deregisterFn()
	}()

	switch serverType {
	case "grpc":
		server.RunGRPCServer(serviceName, func(server *grpc.Server) {
			svc := ports.NewGRPCServer(application)
			stockpb.RegisterStockServiceServer(server, svc)
		})
	case "http":
		// todo
	default:
		panic("unexpected server type: " + serverType)
	}

}
