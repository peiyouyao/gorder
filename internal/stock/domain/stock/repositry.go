package stock

import (
	"context"
	"strings"

	"github.com/peiyouyao/gorder/common/genproto/orderpb"
)

type Repository interface {
	GetItems(ctx context.Context, ids []string) ([]*orderpb.Item, error)
}

type NotFoundError struct {
	Missing []string
}

func (e NotFoundError) Error() string {
	return "these items not found in stock: " + strings.Join(e.Missing, ",")
}
