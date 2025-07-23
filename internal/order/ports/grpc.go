package ports

import (
	context "context"

	"github.com/peiyouyao/gorder/common/convert"
	"github.com/peiyouyao/gorder/common/genproto/orderpb"
	"github.com/peiyouyao/gorder/order/app"
	"github.com/peiyouyao/gorder/order/app/command"
	"github.com/peiyouyao/gorder/order/app/query"
	domain "github.com/peiyouyao/gorder/order/domain/order"
	"github.com/sirupsen/logrus"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// impl orderpb.OrderServiceServer
type GRPCServer struct {
	app app.Application
}

func NewGRPCServer(app app.Application) *GRPCServer {
	return &GRPCServer{app: app}
}

func (s *GRPCServer) CreateOrder(ctx context.Context, request *orderpb.CreateOrderRequest) (*emptypb.Empty, error) {
	_, err := s.app.Commands.CreateOrder.Handle(ctx, command.CreateOrder{
		CustomerID: request.CustomerID,
		Items:      convert.ItemWithQuantityProtosToEntities(request.Items),
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *GRPCServer) GetOrder(ctx context.Context, request *orderpb.GetOrderRequest) (*orderpb.Order, error) {
	o, err := s.app.Queries.GetCustomerOrder.Handle(ctx, query.GetCustomerOrder{
		CustomerID: request.CustomerID,
		OrderID:    request.OrderID,
	})
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &orderpb.Order{
		ID:          o.ID,
		CustomerID:  o.CustomerID,
		Status:      o.Status,
		Items:       convert.ItemEntitiesToProtos(o.Items),
		PaymentLink: o.PaymentLink,
	}, nil
}

func (s *GRPCServer) UpdateOrder(ctx context.Context, request *orderpb.Order) (_ *emptypb.Empty, err error) {
	logrus.Trace("GRPCServer.UpdateOrder")
	order, err := domain.NewOrder(
		request.ID,
		request.CustomerID,
		request.Status,
		request.PaymentLink,
		convert.ItemProtosToEntities(request.Items),
	)
	if err != nil {
		err = status.Error(codes.Internal, err.Error())
		return
	}
	logrus.Tracef("domain.NewOrder order=%v", *order)

	logrus.Trace("app.Commands.UpdateOrder.Handle start")
	_, err = s.app.Commands.UpdateOrder.Handle(ctx, command.UpdateOrder{
		Order: order,
		UpdateFn: func(ctx context.Context, o *domain.Order) (*domain.Order, error) {
			return o, nil
		},
	})
	if err != nil {
		logrus.Trace("app.Commands.UpdateOrder.Handle fail")
	}
	logrus.Trace("app.Commands.UpdateOrder.Handle ok")
	return
}
