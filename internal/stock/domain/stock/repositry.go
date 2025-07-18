package stock

import (
	"context"
	"fmt"
	"strings"

	"github.com/peiyouyao/gorder/common/entity"
)

type Repository interface {
	GetItems(ctx context.Context, ids []string) ([]*entity.Item, error)
	GetStock(ctx context.Context, ids []string) ([]*entity.ItemWithQuantity, error)
	UpdateStock(
		ctx context.Context,
		data []*entity.ItemWithQuantity,
		updateFn func(
			ctx context.Context,
			existing []*entity.ItemWithQuantity,
			query []*entity.ItemWithQuantity,
		) ([]*entity.ItemWithQuantity, error),
	) error
}

type NotFoundError struct {
	Missing []string
}

func (e NotFoundError) Error() string {
	return "these items not found in stock: " + strings.Join(e.Missing, ",")
}

type ExceedStockError struct {
	FailedOn []struct {
		ID   string
		Want int32
		Have int32
	}
}

func (e ExceedStockError) Error() string {
	return fmt.Sprintf("not enough stock for %v", e.FailedOn)
}
