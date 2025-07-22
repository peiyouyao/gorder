package main

import (
	"context"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/peiyouyao/gorder/common/broker"
	_ "github.com/peiyouyao/gorder/common/config"
	"github.com/peiyouyao/gorder/common/discovery"
	"github.com/peiyouyao/gorder/common/genproto/orderpb"
	"github.com/peiyouyao/gorder/common/logging"
	"github.com/peiyouyao/gorder/common/server"
	"github.com/peiyouyao/gorder/common/tracing"
	"github.com/peiyouyao/gorder/order/app"
	"github.com/peiyouyao/gorder/order/infrastructure/consumer"
	"github.com/peiyouyao/gorder/order/ports"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func init() {
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard
	logging.Init()
}

func main() {
	serviceName := viper.GetString("order.service-name")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown, err := tracing.InitJaegerProvider(viper.GetString("jaeger.url"), serviceName)
	if err != nil {
		logrus.Fatal(err)
	}
	defer shutdown(ctx)

	application, cleanup := app.NewApplication(ctx)
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
		ports.RegisterHandlersWithOptions(router, &ports.HTTPServer{App: application}, ports.GinServerOptions{
			BaseURL:      "/api",
			Middlewares:  nil,
			ErrorHandler: nil,
		})
	})
}
