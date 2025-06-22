package query

import (
	"context"

	"github.com/peiyouyao/gorder/common/decorator"
	domain "github.com/peiyouyao/gorder/stock/domain/stock"
	"github.com/peiyouyao/gorder/stock/entity"
	"github.com/peiyouyao/gorder/stock/infrastructure/intergration"
	"github.com/sirupsen/logrus"
)

type CheckIfItemsInStock struct {
	Items []*entity.ItemWithQuantity
}

type CheckIfItemsInStockHandler decorator.QueryHandler[CheckIfItemsInStock, []*entity.Item]

type checkIfItemsInStock struct {
	stockRepo domain.Repository
	stripeAPI *intergration.StripeAPI
}

func NewCheckIfItemsInStockHandler(
	stockRepo domain.Repository,
	stripeAPI *intergration.StripeAPI,
	logger *logrus.Entry,
	metricsClient decorator.MetricsClient,
) CheckIfItemsInStockHandler {
	if stockRepo == nil {
		panic("nil stockRepo")
	}
	if stripeAPI == nil {
		panic("nil stripeAPI")
	}
	return decorator.ApplyQueryDecorators[CheckIfItemsInStock, []*entity.Item](
		checkIfItemsInStock{stockRepo: stockRepo, stripeAPI: stripeAPI},
		logger,
		metricsClient,
	)
}

func (c checkIfItemsInStock) Handle(ctx context.Context, query CheckIfItemsInStock) ([]*entity.Item, error) {
	var res []*entity.Item
	for _, i := range query.Items {
		priceID, err := c.stripeAPI.GetPriceByProductID(ctx, i.ID)
		if err != nil {
			logrus.Warnf("GetPriceByProductID error, item ID = %s, err = %v", i.ID, err)
			continue
		}

		res = append(res, &entity.Item{
			ID:       i.ID,
			Quantity: i.Quantity,
			PriceID:  priceID,
		})
	}
	return res, nil
}
