package query

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/peiyouyao/gorder/common/decorator"
	"github.com/peiyouyao/gorder/common/entity"
	"github.com/peiyouyao/gorder/common/handler/redis"
	"github.com/peiyouyao/gorder/common/logging"
	"github.com/peiyouyao/gorder/common/metrics"
	domain "github.com/peiyouyao/gorder/stock/domain/stock"
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
	metrics metrics.MetricsClient,
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
		metrics,
	)
}

func (h checkIfItemsInStockHandler) Handle(ctx context.Context, query CheckIfItemsInStock) (res []*entity.Item, err error) {
	lockKey := getLockKey(query)
	if err = lock(ctx, lockKey); err != nil {
		return nil, errors.Wrapf(err, "redis lock error||key=%s", lockKey)
	}
	defer func() {
		f := logrus.Fields{
			"query": query,
			"res":   res,
		}
		if err != nil {
			logging.Errorf(ctx, f, "checkIfItemsInStock error||err=%v", err)
		} else {
			logging.Infof(ctx, f, "checkIfItemsInStock success")
		}

		if err = unlock(ctx, lockKey); err != nil {
			logging.Warnf(ctx, nil, "redis unlock error||err=%v", err)
		}
	}()

	for _, it := range query.Items {
		priceID, err := h.stripeAPI.GetPriceByProductID(ctx, it.ID)
		if err != nil {
			logging.Warnf(ctx, nil, "GetPriceByProductID error, item ID = %s, err = %v", it.ID, err)
			return nil, err
		}

		res = append(res, &entity.Item{
			ID:       it.ID,
			Quantity: it.Quantity,
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

func (h checkIfItemsInStockHandler) checkStock(ctx context.Context, queryItems []*entity.ItemWithQuantity) error {
	var ids []string
	for _, it := range queryItems {
		ids = append(ids, it.ID)
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
	for _, item := range queryItems {
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
			queryItems,
			func(
				ctx context.Context,
				existing []*entity.ItemWithQuantity,
				query []*entity.ItemWithQuantity,
			) ([]*entity.ItemWithQuantity, error) {
				var newItems []*entity.ItemWithQuantity
				for _, e := range existing {
					for _, q := range query {
						if e.ID == q.ID {
							itq, err := entity.NewValidItemWithQuantity(e.ID, e.Quantity-q.Quantity)
							if err != nil {
								return nil, err
							}
							newItems = append(newItems, itq)
						}
					}
				}
				return newItems, nil
			},
		)
	}
	return domain.ExceedStockError{FailedOn: failedOn}
}
