package query

import (
	"context"

	"github.com/PerryYao-GitHub/gorder/common/decorator"
	"github.com/PerryYao-GitHub/gorder/common/genproto/orderpb"
	domain "github.com/PerryYao-GitHub/gorder/stock/domain/stock"
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

func (c checkIfItemsInStock) Handle(ctx context.Context, query CheckIfItemsInStock) ([]*orderpb.Item, error) {
	var res []*orderpb.Item
	for _, i := range query.Items {
		res = append(res, &orderpb.Item{
			ID:       i.ID,
			Quantity: i.Quantity,
		})
	}
	return res, nil
	// var (
	// 	ids          []string
	// 	idToQuantity = make(map[string]int32)
	// )
	// for _, q := range query.Items {
	// 	ids = append(ids, q.ID)
	// 	idToQuantity[q.ID] = q.Quantity
	// }

	// itemsInStock, err := c.stockRepo.GetItems(ctx, ids)
	// if err != nil {
	// 	return nil, err
	// }

	// var res []*orderpb.Item
	// for _, item := range itemsInStock {
	// 	need, ok := idToQuantity[item.ID]
	// 	if !ok {
	// 		continue
	// 	}
	// 	if item.Quantity >= need {
	// 		res = append(res, &orderpb.Item{
	// 			ID:       item.ID,
	// 			Name:     item.Name,
	// 			Quantity: item.Quantity,
	// 			PriceID:  item.PriceID,
	// 		})
	// 	}
	// }
	// return res, nil
}
