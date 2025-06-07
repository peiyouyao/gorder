package main

import (
	"context"
	"log"

	"github.com/PerryYao-GitHub/gorder/common/config"
	"github.com/PerryYao-GitHub/gorder/common/discovery"
	"github.com/PerryYao-GitHub/gorder/common/genproto/stockpb"
	"github.com/PerryYao-GitHub/gorder/common/server"
	"github.com/PerryYao-GitHub/gorder/stock/ports"
	"github.com/PerryYao-GitHub/gorder/stock/service"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func init() {
	if err := config.NewViperConfig(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	serviceName := viper.GetString("stock.service-name")
	serverType := viper.GetString("stock.server-to-run")

	ctx, cancal := context.WithCancel(context.Background())
	defer cancal()
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
