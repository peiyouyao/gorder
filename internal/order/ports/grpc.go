package ports

import (
	context "context"

	"github.com/PerryYao-GitHub/gorder/common/genproto/orderpb"
	"github.com/PerryYao-GitHub/gorder/order/app"

	"google.golang.org/protobuf/types/known/emptypb"
)

type GRPCServer struct {
	app app.Application
}

func NewGRPCServer(app app.Application) *GRPCServer {
	return &GRPCServer{app: app}
}

// impl orderpb.OrderServiceServer
func (G GRPCServer) CreateOrder(_ context.Context, _ *orderpb.CreateOrderRequest) (_ *emptypb.Empty, _ error) {
	panic("not implemented") // TODO: Implement
}

func (G GRPCServer) GetOrder(_ context.Context, _ *orderpb.GetOrderRequest) (_ *orderpb.Order, _ error) {
	panic("not implemented") // TODO: Implement
}

func (G GRPCServer) UpdateOrder(_ context.Context, _ *orderpb.Order) (_ *emptypb.Empty, _ error) {
	panic("not implemented") // TODO: Implement
}
