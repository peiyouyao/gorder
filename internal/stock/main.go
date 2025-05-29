package main

import (
	"github.com/PerryYao-GitHub/gorder/common/genproto/stockpb"
	"github.com/PerryYao-GitHub/gorder/common/server"
	"github.com/PerryYao-GitHub/gorder/stock/ports"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func main() {
	serviceName := viper.GetString("stock.service-name")
	serverType := viper.GetString("stock.server-to-run")

	switch serverType {
	case "grpc":
		server.RunGRPCServer(serviceName, func(server *grpc.Server) {
			svc := ports.NewGRPCServer()
			stockpb.RegisterStockServiceServer(server, svc)
		})
	case "http":
		// todo
	default:
		panic("unexpected server type: " + serverType)
	}

}
