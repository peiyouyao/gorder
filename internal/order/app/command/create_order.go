package command

import (
	"context"
	"fmt"

	"github.com/peiyouyao/gorder/common/broker"
	"github.com/peiyouyao/gorder/common/convert"
	"github.com/peiyouyao/gorder/common/decorator"
	"github.com/peiyouyao/gorder/common/entity"
	"github.com/peiyouyao/gorder/common/metrics"
	"github.com/peiyouyao/gorder/order/app/query"
	domain "github.com/peiyouyao/gorder/order/domain/order"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

type CreateOrder struct {
	CustomerID string
	Items      []*entity.ItemWithQuantity
}

type CreateOrderResult struct {
	OrderID string
}

type CreateOrderHandler decorator.CommandHandler[CreateOrder, *CreateOrderResult]

type createOrderHandler struct {
	orderRepo domain.Repository
	stockGRPC query.StockService
	channel   *amqp.Channel
}

func NewCreateOrderHandler(
	orderRepo domain.Repository,
	stockGRPC query.StockService,
	channel *amqp.Channel,
	logger *logrus.Entry,
	metricClient metrics.MetricsClient,
) CreateOrderHandler {
	if orderRepo == nil {
		panic("nil orderRepo")
	}
	if stockGRPC == nil {
		panic("nil stockGRPC")
	}
	if channel == nil {
		panic("nil channel ")
	}
	return decorator.ApplyCommandDecorators[CreateOrder, *CreateOrderResult](
		createOrderHandler{
			orderRepo: orderRepo,
			stockGRPC: stockGRPC,
			channel:   channel,
		},
		logger,
		metricClient,
	)
}

func (c createOrderHandler) Handle(ctx context.Context, cmd CreateOrder) (*CreateOrderResult, error) {
	q, err := c.channel.QueueDeclare(broker.EventOrderCreated, true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	t := otel.Tracer("rabbitmq")
	ctx, span := t.Start(ctx, fmt.Sprintf("rabbitmq.%s.publish", q.Name))
	defer span.End()

	validItems, err := c.validate(ctx, cmd.Items)
	if err != nil {
		return nil, err
	}

	pendingOrder, err := domain.NewPendingOrder(cmd.CustomerID, validItems)
	if err != nil {
		return nil, err
	}

	logrus.Debug("orderRepo.Create_start")
	o, err := c.orderRepo.Create(ctx, pendingOrder)
	if err != nil {
		logrus.Debugf("orderRepo.Create_fail || err=%v", err)
		return nil, err
	}
	logrus.Debugf("orderRepo.Create_success || order=%v", *o)

	logrus.Debug("broker.PublishEvent_start")
	r := &broker.PublishEventReq{
		Channel:  c.channel,
		Routing:  broker.Direct,
		Queue:    q.Name,
		Exchange: "",
		Body:     *o,
	}
	err = broker.PublishEvent(ctx, r)
	if err != nil {
		logrus.Debugf("broker.PublishEvent_fail || err=%v", err)
		return nil, errors.Wrap(err, "failed to publish order created event")
	}
	logrus.Debugf("broker.PublishEvent_success || broker.PublishEventReq=%v", *r)

	return &CreateOrderResult{OrderID: o.ID}, nil
}

func (c createOrderHandler) validate(ctx context.Context, items []*entity.ItemWithQuantity) ([]*entity.Item, error) {
	if len(items) == 0 {
		return nil, errors.New("must have at least one item")
	}
	items = packItems(items)
	resp, err := c.stockGRPC.CheckIfItemsInStock(ctx, convert.ItemWithQuantityEntitiesToProtos(items))
	if err != nil {
		return nil, err
	}
	return convert.ItemProtosToEntities(resp.Items), nil
}

func packItems(items []*entity.ItemWithQuantity) []*entity.ItemWithQuantity {
	merged := make(map[string]int32)
	for _, item := range items {
		merged[item.ID] += item.Quantity
	}
	var res []*entity.ItemWithQuantity
	for id, quantity := range merged {
		res = append(res, &entity.ItemWithQuantity{
			ID:       id,
			Quantity: quantity,
		})
	}
	return res
}
