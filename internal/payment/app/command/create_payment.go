package command

import (
	"context"

	"github.com/peiyouyao/gorder/common/constants"
	"github.com/peiyouyao/gorder/common/convert"
	"github.com/peiyouyao/gorder/common/decorator"
	"github.com/peiyouyao/gorder/common/entity"
	"github.com/peiyouyao/gorder/common/metrics"
	"github.com/peiyouyao/gorder/payment/domain"
	"github.com/sirupsen/logrus"
)

type CreatePayment struct {
	Order *entity.Order
}

type CreatePaymentHandler decorator.CommandHandler[CreatePayment, string]

type createPaymentHandler struct {
	processor domain.Processor
	orderGRPC OrderService
}

func (c createPaymentHandler) Handle(ctx context.Context, cmd CreatePayment) (link string, err error) {
	if link, err = c.processor.CreatePaymentLink(ctx, cmd.Order); err != nil {
		return
	}

	newOrder, err := entity.NewValidOrder(
		cmd.Order.ID,
		cmd.Order.CustomerID,
		constants.OrderStatusWaitingForPayment,
		link,
		cmd.Order.Items,
	)
	if err != nil {
		return
	}
	err = c.orderGRPC.UpdateOrder(ctx, convert.OrderEntityToProto(newOrder))
	return link, err
}

func NewCreatePaymentHandler(
	processor domain.Processor,
	orderGRPC OrderService,
	logger *logrus.Entry,
	metrics metrics.MetricsClient,
) decorator.CommandHandler[CreatePayment, string] {

	return decorator.ApplyCommandDecorators[CreatePayment, string](
		createPaymentHandler{
			processor: processor,
			orderGRPC: orderGRPC,
		},
		logger,
		metrics,
	)
}
