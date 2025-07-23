package query

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/peiyouyao/gorder/common/decorator"
	"github.com/peiyouyao/gorder/common/entity"
	"github.com/peiyouyao/gorder/common/handler/redis"
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
	var lkerr error
	lockKey := getLockKey(query)
	if lkerr = lock(ctx, lockKey); lkerr != nil {
		return nil, errors.Wrapf(err, "Redis lock error key=%s", lockKey)
	}
	defer func() {
		if lkerr = unlock(ctx, lockKey); lkerr != nil {
			logrus.WithContext(ctx).Warnf("Redis unlock fail err=%v", lkerr)
		}
	}()

	for _, it := range query.Items {
		priceID, err := h.stripeAPI.GetPriceByProductID(ctx, it.ID)
		if err != nil {
			logrus.WithContext(ctx).Warnf("GetPriceByProductID from stripe fail item_id=%s err=%v", it.ID, err)
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

	fs := logrus.Fields{
		"query": query,
		"res":   res,
	}
	if err != nil {
		logrus.WithContext(ctx).WithFields(fs).Errorf("checkIfItemsInStock fail err=%v", err)
	} else {
		logrus.WithContext(ctx).WithFields(fs).Info("checkIfItemsInStock ok")
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
	idQuantityMap := h.tidyItems(records)

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
	if !ok {
		return domain.ExceedStockError{FailedOn: failedOn}
	}
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

func (h checkIfItemsInStockHandler) tidyItems(items []*entity.ItemWithQuantity) (res map[string]int32) {
	res = make(map[string]int32)
	for _, it := range items {
		res[it.ID] += it.Quantity
	}
	return
}
