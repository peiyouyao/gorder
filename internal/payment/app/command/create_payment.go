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
	logrus.Debug("create_link_from_stripe_success")

	newOrder, err := entity.NewValidOrder(
		cmd.Order.ID,
		cmd.Order.CustomerID,
		constants.OrderStatusWaitingForPayment, // 生成了 link, 改状态为 waiting_for_pay
		link,
		cmd.Order.Items,
	)
	if err != nil {
		return
	}
	logrus.Debugf("NewValidOrder_success || newOrder=%v", *newOrder)

	logrus.Debug("orderGRPC.UpdateOrder_start")
	err = c.orderGRPC.UpdateOrder(ctx, convert.OrderEntityToProto(newOrder)) // 发送 grpc 给 order
	if err != nil {
		logrus.Debug("orderGRPC.UpdateOrder_fail")
	}
	logrus.Debug("orderGRPC.UpdateOrder_success")
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
