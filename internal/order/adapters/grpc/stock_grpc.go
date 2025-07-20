package grpc

import (
	"context"

	"github.com/peiyouyao/gorder/common/genproto/orderpb"
	"github.com/peiyouyao/gorder/common/genproto/stockpb"
)

type StockGRPC struct {
	client stockpb.StockServiceClient
}

func NewStockGRPC(client stockpb.StockServiceClient) *StockGRPC {
	return &StockGRPC{client: client}
}

func (s StockGRPC) CheckIfItemsInStock(ctx context.Context, items []*orderpb.ItemWithQuantity) (*stockpb.CheckIfItemsInStockResponse, error) {
	resp, err := s.client.CheckIfItemsInStock(ctx, &stockpb.CheckIfItemsInStockRequest{Items: items})
	return resp, err
}

func (s StockGRPC) GetItems(ctx context.Context, itemsIDs []string) ([]*orderpb.Item, error) {
	resp, err := s.client.GetItems(ctx, &stockpb.GetItemsRequest{ItemIDs: itemsIDs})
	if err != nil {
		return nil, err
	}
	return resp.Items, nil
}
