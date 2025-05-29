package main

import (
	"context"
	"log"

	"github.com/PerryYao-GitHub/gorder/common/config"
	"github.com/PerryYao-GitHub/gorder/common/genproto/orderpb"
	"github.com/PerryYao-GitHub/gorder/common/server"
	"github.com/PerryYao-GitHub/gorder/order/ports"
	"github.com/PerryYao-GitHub/gorder/order/service"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func init() {
	if err := config.NewViperConfig(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	serviceName := viper.GetString("order.service-name")
	serverType := viper.GetString("order.server-to-run")
	if serverType != "http" && serverType != "grpc" {
		panic("unexpected server type: " + serverType)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	application := service.NewApplication(ctx)

	go server.RunGRPCServer(serviceName, func(server *grpc.Server) {
		svc := ports.NewGRPCServer(application)
		orderpb.RegisterOrderServiceServer(server, svc)
	})

	server.RunHTTPServer(serviceName, func(router *gin.Engine) {
		ports.RegisterHandlersWithOptions(router, &HTTPServer{app: application}, ports.GinServerOptions{
			BaseURL:      "/api",
			Middlewares:  nil,
			ErrorHandler: nil,
		})
	})
}
