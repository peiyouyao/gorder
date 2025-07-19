package order

import "errors"

type Identity struct {
	OrderID    string
	CustomerID string
}

type AggregateRoot struct {
	Identity Identity
	Order    *Order
}

func NewAggregateRoot(identity Identity, order *Order) *AggregateRoot {
	return &AggregateRoot{Identity: identity, Order: order}
}

func (r *AggregateRoot) BusinessIdentity() Identity {
	return Identity{
		OrderID:    r.Order.ID,
		CustomerID: r.Order.CustomerID,
	}
}

func (r *AggregateRoot) Validate() error {
	if r.Identity.OrderID == "" || r.Identity.CustomerID == "" {
		return errors.New("invalid identity")
	}
	if r.Order == nil {
		return errors.New("empty order")
	}
	return nil
}
