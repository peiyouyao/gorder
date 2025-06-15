package main

import (
	"context"

	"github.com/PerryYao-GitHub/gorder/common/broker"
	_ "github.com/PerryYao-GitHub/gorder/common/config"
	"github.com/PerryYao-GitHub/gorder/common/discovery"
	"github.com/PerryYao-GitHub/gorder/common/genproto/orderpb"
	"github.com/PerryYao-GitHub/gorder/common/logging"
	"github.com/PerryYao-GitHub/gorder/common/server"
	"github.com/PerryYao-GitHub/gorder/common/tracing"
	"github.com/PerryYao-GitHub/gorder/order/infrastructure/consumer"
	"github.com/PerryYao-GitHub/gorder/order/ports"
	"github.com/PerryYao-GitHub/gorder/order/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func init() {
	logging.Init()
}

func main() {
	serviceName := viper.GetString("order.service-name")
	serverType := viper.GetString("order.server-to-run")
	if serverType != "http" && serverType != "grpc" {
		panic("unexpected server type: " + serverType)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown, err := tracing.InitJaegerProvider(viper.GetString("jaeger.url"), serviceName)
	if err != nil {
		logrus.Fatal(err)
	}
	defer shutdown(ctx)

	application, cleanup := service.NewApplication(ctx)
	defer cleanup()

	deregisterFn, err := discovery.RegisterToConsul(ctx, serviceName)
	if err != nil {
		logrus.Fatal(err)
	}
	defer func() {
		_ = deregisterFn()
	}()

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

	go server.RunGRPCServer(serviceName, func(server *grpc.Server) {
		svc := ports.NewGRPCServer(application)
		orderpb.RegisterOrderServiceServer(server, svc)
	})

	server.RunHTTPServer(serviceName, func(router *gin.Engine) {
		router.StaticFile("/success", "../../public/success.html")
		ports.RegisterHandlersWithOptions(router, &HTTPServer{app: application}, ports.GinServerOptions{
			BaseURL:      "/api",
			Middlewares:  nil,
			ErrorHandler: nil,
		})
	})
}
