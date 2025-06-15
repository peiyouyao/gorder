package query

import (
	"context"

	"github.com/peiyouyao/gorder/common/genproto/orderpb"
	"github.com/peiyouyao/gorder/common/genproto/stockpb"
)

type StockService interface {
	CheckIfItemsInStock(ctx context.Context, items []*orderpb.ItemWithQuantity) (*stockpb.CheckIfItemsInStockResponse, error)
	GetItems(ctx context.Context, itemsIDs []string) ([]*orderpb.Item, error)
}
