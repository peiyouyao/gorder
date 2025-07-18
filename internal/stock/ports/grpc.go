package ports

import (
	context "context"

	"github.com/peiyouyao/gorder/common/convert"
	"github.com/peiyouyao/gorder/common/genproto/stockpb"
	"github.com/peiyouyao/gorder/common/tracing"
	"github.com/peiyouyao/gorder/stock/app"
	"github.com/peiyouyao/gorder/stock/app/query"
	"google.golang.org/grpc/status"
)

// impl stockpb.StockServiceServer
type GRPCServer struct {
	app app.Application
}

func NewGRPCServer(app app.Application) *GRPCServer {
	return &GRPCServer{app: app}
}

func (G GRPCServer) GetItems(ctx context.Context, request *stockpb.GetItemsRequest) (*stockpb.GetItemsResponse, error) {
	_, span := tracing.Start(ctx, "GetItems")
	defer span.End()

	items, err := G.app.Queries.GetItems.Handle(ctx, query.GetItems{ItemIDs: request.ItemIDs})
	if err != nil {
		return nil, err
	}
	return &stockpb.GetItemsResponse{Items: convert.ItemEntitiesToProtos(items)}, nil
}

func (G GRPCServer) CheckIfItemsInStock(ctx context.Context, request *stockpb.CheckIfItemsInStockRequest) (*stockpb.CheckIfItemsInStockResponse, error) {
	_, span := tracing.Start(ctx, "CheckIfItemsInStock")
	defer span.End()

	items, err := G.app.Queries.CheckIfItemsInStock.Handle(ctx, query.CheckIfItemsInStock{
		Items: convert.ItemWithQuantityProtosToEntities(request.Items)})
	if err != nil {
		return nil, status.Error(status.Code(err), err.Error())
	}
	return &stockpb.CheckIfItemsInStockResponse{
		InStock: 1,
		Items:   convert.ItemEntitiesToProtos(items),
	}, nil
}
