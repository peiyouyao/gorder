package service

import (
	"context"

	"github.com/peiyouyao/gorder/common/broker"
	"github.com/peiyouyao/gorder/common/entity"
	domain "github.com/peiyouyao/gorder/order/domain/order"
	"github.com/pkg/errors"
)

type OrderDomainService struct {
	Repo           domain.Repository
	EventPublisher domain.EventPublisher
}

func NewOrderDomainService(repo domain.Repository, eventPublisher domain.EventPublisher) *OrderDomainService {
	return &OrderDomainService{Repo: repo, EventPublisher: eventPublisher}
}

func (s *OrderDomainService) CreateOrder(ctx context.Context, order domain.Order) (res *entity.Order, err error) {
	root := domain.NewAggregateRoot(
		domain.Identity{CustomerID: order.CustomerID, OrderID: order.ID},
		&order,
	)
	o, err := s.Repo.Create(ctx, root.Order)
	if err != nil {
		return
	}

	if err = s.EventPublisher.Publish(ctx, domain.DomainEvent{
		Dest: broker.EventOrderCreated,
		Data: o,
	}); err != nil {
		return nil, errors.Wrapf(err, "publish event error||q.Name=%s", broker.EventOrderCreated)
	}

	return &entity.Order{
		ID:          o.ID,
		CustomerID:  o.CustomerID,
		Status:      o.Status,
		PaymentLink: o.PaymentLink,
		Items:       o.Items,
	}, nil
}
