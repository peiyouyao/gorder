package adapters

import (
	"context"

	"github.com/PerryYao-GitHub/gorder/common/genproto/orderpb"
	"github.com/sirupsen/logrus"
)

type OrderGRPC struct {
	client orderpb.OrderServiceClient
}

func NewOrderGRPC(client orderpb.OrderServiceClient) *OrderGRPC {

	return &OrderGRPC{client: client}
}

func (o OrderGRPC) UpdateOrder(ctx context.Context, order *orderpb.Order) error {

	_, err := o.client.UpdateOrder(ctx, order)
	logrus.Infof("payment_adapter || update_order, err=%v", err)
	return err
}
