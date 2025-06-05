package adapters

import (
	"context"
	"sync"

	"github.com/PerryYao-GitHub/gorder/common/genproto/orderpb"
	domain "github.com/PerryYao-GitHub/gorder/stock/domain/stock"
)

type MemoryStockRepository struct {
	lock  *sync.RWMutex
	store map[string]*orderpb.Item
}

var stub = map[string]*orderpb.Item{
	"item-1": {
		ID:       "item-1",
		Name:     "bar-1",
		Quantity: 10000,
		PriceID:  "pid-1",
	},
	"item-2": {
		ID:       "item-2",
		Name:     "bar-2",
		Quantity: 10000,
		PriceID:  "pid-2",
	},
	"item-3": {
		ID:       "item-3",
		Name:     "bar-3",
		Quantity: 10000,
		PriceID:  "pid-3",
	},
	"item-4": {
		ID:       "item-4",
		Name:     "bar-4",
		Quantity: 10000,
		PriceID:  "pid-4",
	},
}

func NewMemoryStockRepository() *MemoryStockRepository {
	return &MemoryStockRepository{
		lock:  &sync.RWMutex{},
		store: stub,
	}
}

func (m MemoryStockRepository) GetItems(ctx context.Context, ids []string) ([]*orderpb.Item, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	var (
		res     []*orderpb.Item
		missing []string
	)
	for _, id := range ids {
		if item, exist := m.store[id]; exist {
			res = append(res, item)
		} else {
			missing = append(missing, id)
		}
	}
	if len(res) == len(ids) {
		return res, nil
	}
	return res, domain.NotFoundError{Missing: missing}
}
