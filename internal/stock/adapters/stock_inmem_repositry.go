package adapters

import (
	"context"
	"sync"

	"github.com/peiyouyao/gorder/common/genproto/orderpb"
	domain "github.com/peiyouyao/gorder/stock/domain/stock"
)

type MemoryStockRepository struct {
	lock  *sync.RWMutex
	store map[string]*orderpb.Item
}

var stub = map[string]*orderpb.Item{
	"1": {
		ID:       "1",
		Name:     "Fires",
		Quantity: 5000,
		PriceID:  "price_1RXLrqPqGUzmzBMUyWDWprnO",
	},
	"2": {
		ID:       "2",
		Name:     "Cookie",
		Quantity: 1600,
		PriceID:  "price_1RY15bPqGUzmzBMU2sfOn6gf",
	},
	"3": {
		ID:       "3",
		Name:     "AnimeBook",
		Quantity: 400,
		PriceID:  "price_1RY18LPqGUzmzBMUicg0gEVS",
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
