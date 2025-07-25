package adapters

import (
	"context"
	"strconv"
	"sync"
	"time"

	domain "github.com/peiyouyao/gorder/order/domain/order"
	"github.com/sirupsen/logrus"
)

type OrderRepositoryInmem struct {
	lock  *sync.RWMutex
	store []*domain.Order
}

func NewOrderRepositoryInmem() *OrderRepositoryInmem {
	s := make([]*domain.Order, 0)
	return &OrderRepositoryInmem{
		lock:  &sync.RWMutex{},
		store: s,
	}
}

// impl domain.Repository
func (m *OrderRepositoryInmem) Create(_ context.Context, order *domain.Order) (*domain.Order, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	res := &domain.Order{
		ID:          strconv.FormatInt(time.Now().Unix(), 10),
		CustomerID:  order.CustomerID,
		Status:      order.Status,
		PaymentLink: order.PaymentLink,
		Items:       order.Items,
	}
	m.store = append(m.store, res)
	logrus.WithFields(logrus.Fields{
		"input_order":        order,
		"store_after_create": m.store,
	}).Debug("OrderRepositoryInmem.Create")
	return res, nil
}

func (m *OrderRepositoryInmem) Get(ctx context.Context, id, customerID string) (*domain.Order, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	for _, o := range m.store {
		if o.ID == id && o.CustomerID == customerID {
			logrus.Debugf("OrderRepositoryInmem.Get id=%s customerID=%s res=%+v", id, customerID, *o)
			return o, nil
		}
	}
	return nil, domain.NotFoundError{OrderID: id}
}

func (m *OrderRepositoryInmem) Update(
	ctx context.Context,
	order *domain.Order,
	updateFn func(context.Context, *domain.Order) (*domain.Order, error),
) error {

	m.lock.Lock()
	defer m.lock.Unlock()
	for i, o := range m.store {
		if o.ID == order.ID && o.CustomerID == order.CustomerID {
			updatedOrder, err := updateFn(ctx, order)
			if err != nil {
				return err
			}
			m.store[i] = updatedOrder
			return nil
		}
	}
	return domain.NotFoundError{OrderID: order.ID}
}
