package command

import (
	"context"

	"github.com/peiyouyao/gorder/common/decorator"
	"github.com/peiyouyao/gorder/common/metrics"
	domain "github.com/peiyouyao/gorder/order/domain/order"
	"github.com/sirupsen/logrus"
)

type UpdateOrder struct {
	Order    *domain.Order
	UpdateFn func(context.Context, *domain.Order) (*domain.Order, error)
}

type UpdateOrderHandler decorator.CommandHandler[UpdateOrder, interface{}]

type updateOrderHandler struct {
	orderRepo domain.Repository
	// stockGRPC
}

func NewUpdateOrderHandler(
	orderRepo domain.Repository,
	logger *logrus.Entry,
	metricsClient metrics.MetricsClient,
) UpdateOrderHandler {
	if orderRepo == nil {
		panic("nil orderRepo")
	}
	return decorator.ApplyCommandDecorators[UpdateOrder, interface{}](
		updateOrderHandler{orderRepo: orderRepo},
		logger,
		metricsClient,
	)
}

func (u updateOrderHandler) Handle(ctx context.Context, cmd UpdateOrder) (interface{}, error) {
	if cmd.UpdateFn == nil {
		logrus.Warnf("nil_UpdateFn || order=%v", cmd.Order)
		cmd.UpdateFn = func(ctx context.Context, o *domain.Order) (*domain.Order, error) {
			return o, nil
		}
	}
	logrus.Debugf("orderRepo.Update_start || order=%v", *cmd.Order)
	if err := u.orderRepo.Update(ctx, cmd.Order, cmd.UpdateFn); err != nil {
		logrus.Debugf("orderRepo.Update_fail || err=%v", err)
		return nil, err
	}
	logrus.Debugf("orderRepo.Update_success")
	return nil, nil
}
