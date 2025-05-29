package ports

import (
	context "context"

	"github.com/PerryYao-GitHub/gorder/common/genproto/stockpb"
)

type GRPCServer struct{}

func NewGRPCServer() *GRPCServer {
	return &GRPCServer{}
}

// impl stockpb.StockServiceServer
func (G GRPCServer) GetItems(ctx context.Context, request *stockpb.GetItemsRequest) (*stockpb.GetItemsResponse, error) {
	panic("implement me")
}

func (G GRPCServer) CheckIfItemsInStock(ctx context.Context, request *stockpb.CheckIfItemsInStockRequest) (*stockpb.CheckIfItemsInStockResponse, error) {
	panic("implement me")
}
