package query

import (
	"context"

	"github.com/peiyouyao/gorder/common/decorator"
	"github.com/peiyouyao/gorder/common/entity"
	"github.com/peiyouyao/gorder/common/metrics"
	domain "github.com/peiyouyao/gorder/stock/domain/stock"
	"github.com/sirupsen/logrus"
)

type GetItems struct {
	ItemIDs []string
}

type GetItemsHandler decorator.QueryHandler[GetItems, []*entity.Item]

type getItemsHandler struct {
	stockRepo domain.Repository
}

func NewGetItemsHandler(
	stockRepo domain.Repository,
	logger *logrus.Entry,
	metrics metrics.MetricsClient,
) GetItemsHandler {
	if stockRepo == nil {
		panic("nil stockRepo")
	}
	return decorator.ApplyQueryDecorators[GetItems, []*entity.Item](
		getItemsHandler{stockRepo: stockRepo},
		logger,
		metrics,
	)
}

func (g getItemsHandler) Handle(ctx context.Context, query GetItems) ([]*entity.Item, error) {
	items, err := g.stockRepo.GetItems(ctx, query.ItemIDs)
	if err != nil {
		return nil, err
	}
	return items, nil
}
