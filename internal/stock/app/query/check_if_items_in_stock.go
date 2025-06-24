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

type checkIfItemsInStockHandler struct {
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
		checkIfItemsInStockHandler{stockRepo: stockRepo, stripeAPI: stripeAPI},
		logger,
		metricsClient,
	)
}

func (c checkIfItemsInStockHandler) Handle(ctx context.Context, query CheckIfItemsInStock) ([]*entity.Item, error) {
	if err := c.checkStock(ctx, query.Items); err != nil {
		return nil, err
	}

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

func (c checkIfItemsInStockHandler) checkStock(ctx context.Context, items []*entity.ItemWithQuantity) error {
	var ids []string
	for _, i := range items {
		ids = append(ids, i.ID)
	}
	records, err := c.stockRepo.GetStock(ctx, ids)
	if err != nil {
		return err
	}
	idQuantityMap := make(map[string]int32)
	for _, r := range records {
		idQuantityMap[r.ID] += r.Quantity
	}

	var (
		ok       = true
		failedOn []struct {
			ID   string
			Want int32
			Have int32
		}
	)
	for _, item := range items {
		if item.Quantity > idQuantityMap[item.ID] {
			ok = false
			failedOn = append(failedOn, struct {
				ID   string
				Want int32
				Have int32
			}{ID: item.ID, Want: item.Quantity, Have: idQuantityMap[item.ID]})
		}
	}
	if ok {
		return nil
	}
	return domain.ExceedStockError{FailedOn: failedOn}
}
