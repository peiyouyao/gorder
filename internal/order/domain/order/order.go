package order

import (
	"errors"
	"fmt"
	"slices"

	"github.com/peiyouyao/gorder/common/constants"
	"github.com/peiyouyao/gorder/common/entity"
	"github.com/stripe/stripe-go/v82"
)

type Order struct {
	ID          string
	CustomerID  string
	Status      string
	PaymentLink string
	Items       []*entity.Item
}

func NewOrder(id, customerID, status, paymentLink string, items []*entity.Item) (*Order, error) {
	if id == "" {
		return nil, errors.New("empty id")
	}
	if customerID == "" {
		return nil, errors.New("empty customerID")
	}
	if status == "" {
		return nil, errors.New("empty status")
	}
	if items == nil {
		return nil, errors.New("empty items")
	}
	return &Order{
		ID:          id,
		CustomerID:  customerID,
		Status:      status,
		PaymentLink: paymentLink,
		Items:       items,
	}, nil
}

func NewPendingOrder(customerID string, items []*entity.Item) (*Order, error) {
	if customerID == "" {
		return nil, errors.New("empty customerID")
	}

	if items == nil {
		return nil, errors.New("empty items")
	}

	return &Order{
		CustomerID: customerID,
		Status:     "pending",
		Items:      items,
	}, nil
}

func (o *Order) IsPaid() error {
	if o.Status == string(stripe.CheckoutSessionPaymentStatusPaid) {
		return nil
	}
	return fmt.Errorf("order status not paid, order id = %s, status = %s", o.ID, o.Status)
}

func (o *Order) UpdatePaymentLink(link string) error {
	if link == "" {
		return errors.New("cannot update empty paymentLink")
	}
	o.PaymentLink = link
	return nil
}

func (o *Order) UpdateItems(items []*entity.Item) error {
	o.Items = items
	return nil
}

func (o *Order) UpdateStatus(to string) error {
	if !o.isValidStatusTransition(to) {
		return fmt.Errorf("cannot transit from '%s' to '%s'", o.Status, to)
	}
	o.Status = to
	return nil
}

func (o *Order) isValidStatusTransition(to string) bool {
	switch o.Status {
	default:
		return false
	case constants.OrderStatusPending:
		return slices.Contains([]string{constants.OrderStatusWaitingForPayment}, to)
	case constants.OrderStatusWaitingForPayment:
		return slices.Contains([]string{constants.OrderStatusPaid}, to)
	case constants.OrderStatusPaid:
		return slices.Contains([]string{constants.OrderStatusReady}, to)
	}
}
