package query

import (
	"context"

	"github.com/peiyouyao/gorder/common/decorator"
	"github.com/peiyouyao/gorder/common/genproto/orderpb"
	domain "github.com/peiyouyao/gorder/stock/domain/stock"
	"github.com/sirupsen/logrus"
)

type CheckIfItemsInStock struct {
	Items []*orderpb.ItemWithQuantity
}

type CheckIfItemsInStockHandler decorator.QueryHandler[CheckIfItemsInStock, []*orderpb.Item]

type checkIfItemsInStock struct {
	stockRepo domain.Repository
}

func NewCheckIfItemsInStockHandler(
	stockRepo domain.Repository,
	logger *logrus.Entry,
	metricsClient decorator.MetricsClient,
) CheckIfItemsInStockHandler {
	if stockRepo == nil {
		panic("nil stockRepo")
	}
	return decorator.ApplyQueryDecorators[CheckIfItemsInStock, []*orderpb.Item](
		checkIfItemsInStock{stockRepo: stockRepo},
		logger,
		metricsClient,
	)
}

// TODO: del
var stub = map[string]string{
	"1": "price_1RXLrqPqGUzmzBMUyWDWprnO",
	"2": "price_1RY15bPqGUzmzBMU2sfOn6gf",
	"3": "price_1RY18LPqGUzmzBMUicg0gEVS",
}

func (c checkIfItemsInStock) Handle(ctx context.Context, query CheckIfItemsInStock) ([]*orderpb.Item, error) {
	var res []*orderpb.Item
	for _, i := range query.Items {
		// TODO: get from db or stripe
		priceID, ok := stub[i.ID]
		if !ok {
			priceID = stub["1"] // default priceID
		}

		res = append(res, &orderpb.Item{
			ID:       i.ID,
			Quantity: i.Quantity,
			PriceID:  priceID,
		})
	}
	return res, nil
}
