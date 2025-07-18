package query

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/peiyouyao/gorder/common/decorator"
	"github.com/peiyouyao/gorder/common/handler/redis"
	"github.com/peiyouyao/gorder/common/metrics"
	domain "github.com/peiyouyao/gorder/stock/domain/stock"
	"github.com/peiyouyao/gorder/stock/entity"
	"github.com/peiyouyao/gorder/stock/infrastructure/intergration"
	"github.com/sirupsen/logrus"
)

const (
	redisLockPrefix = "gorder:stock:check_if_items_in_stock:"
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
	metricsClient metrics.MetricsClient,
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

func (h checkIfItemsInStockHandler) Handle(ctx context.Context, query CheckIfItemsInStock) ([]*entity.Item, error) {
	if err := lock(ctx, getLockKey(query)); err != nil {
		return nil, errors.Wrap(err, "redis lock error")
	}
	defer func() {
		if err := unlock(ctx, getLockKey(query)); err != nil {
			logrus.Warnf("unlock fail, err = %v", err)
		}
	}()

	var res []*entity.Item
	for _, i := range query.Items {
		priceID, err := h.stripeAPI.GetPriceByProductID(ctx, i.ID)
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
	if err := h.checkStock(ctx, query.Items); err != nil {
		return nil, err
	}
	return res, nil
}

func lock(ctx context.Context, key string) error {
	return redis.SetNX(ctx, redis.LocaClient(), key, "1", 5*time.Minute)
}

func unlock(ctx context.Context, key string) error {
	return redis.Del(ctx, redis.LocaClient(), key)
}

func getLockKey(query CheckIfItemsInStock) string {
	var ids []string
	for _, i := range query.Items {
		ids = append(ids, i.ID)
	}
	return redisLockPrefix + "items:" + strings.Join(ids, ",")
}

func (h checkIfItemsInStockHandler) checkStock(ctx context.Context, query []*entity.ItemWithQuantity) error {
	var ids []string
	for _, i := range query {
		ids = append(ids, i.ID)
	}
	records, err := h.stockRepo.GetStock(ctx, ids)
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
	for _, item := range query {
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
		return h.stockRepo.UpdateStock(
			ctx,
			query,
			func(
				ctx context.Context,
				existing []*entity.ItemWithQuantity,
				query []*entity.ItemWithQuantity) ([]*entity.ItemWithQuantity, error) {
				var newItems []*entity.ItemWithQuantity
				for _, e := range existing {
					for _, q := range query {
						if e.ID == q.ID {
							newItems = append(newItems, &entity.ItemWithQuantity{
								ID:       e.ID,
								Quantity: e.Quantity - q.Quantity,
							})
						}
					}
				}
				return newItems, nil
			},
		)
	}
	return domain.ExceedStockError{FailedOn: failedOn}
}
