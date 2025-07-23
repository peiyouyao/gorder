package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/peiyouyao/gorder/common/broker"
	"github.com/peiyouyao/gorder/order/app"
	"github.com/peiyouyao/gorder/order/app/command"
	domain "github.com/peiyouyao/gorder/order/domain/order"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

/*
消费 mq 中 order.paid 消息, 更新 order状态为 paid
*/
type Consumer struct {
	app app.Application
}

func NewConsumer(app app.Application) *Consumer {
	return &Consumer{
		app: app,
	}
}

func (c *Consumer) Listen(ch *amqp.Channel) {
	q, err := ch.QueueDeclare(broker.EventOrderPaid, true, false, true, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}
	err = ch.QueueBind(q.Name, "", broker.EventOrderPaid, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}
	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}

	for msg := range msgs {
		c.handleMessage(ch, msg, q)
	}
}

func (c *Consumer) handleMessage(ch *amqp.Channel, msg amqp.Delivery, q amqp.Queue) { // order的consume只执行更新订单
	logrus.WithFields(logrus.Fields{
		"from_q": q.Name,
		"msg_id": msg.MessageId,
	}).Info("Receive order.paid msg")
	logrus.Trace("Receive order.paid")

	tr := otel.Tracer("rabbitmq")
	ctx, span := tr.Start(
		broker.ExtractRabbitMQHeaders(context.Background(), msg.Headers),
		fmt.Sprintf("rabbitmq.%s.consume", q.Name),
	)
	defer span.End()

	var err error
	defer func() {
		if err != nil {
			fs := logrus.Fields{
				"q_name": q.Name,
				"q_msg":  msg,
				"err":    err.Error(),
			}
			logrus.WithContext(ctx).WithFields(fs).Warn("MQ consume fail")
			_ = msg.Nack(false, false)
		} else {
			logrus.WithContext(ctx).Info("MQ consume ok")
			_ = msg.Ack(false)
		}
	}()

	o := &domain.Order{}
	if err = json.Unmarshal(msg.Body, o); err != nil {
		logrus.Warnf("unmarshal_fail || err=%s", err.Error())
		return
	}
	logrus.Tracef("paid.order=%v", *o)

	logrus.Trace("app.Commands.UpdateOrder.Handle start")
	_, err = c.app.Commands.UpdateOrder.Handle(ctx, command.UpdateOrder{
		Order: o,
		UpdateFn: func(ctx context.Context, order *domain.Order) (*domain.Order, error) {
			if err := order.IsPaid(); err != nil {
				return nil, err
			}
			return order, nil
		},
	})
	if err != nil {
		fs := logrus.Fields{
			"order_id": o.ID,
			"q_name":   q.Name,
			"q_msg":    msg,
			"err":      err.Error(),
		}
		logrus.WithContext(ctx).WithFields(fs).Error("Update order fail")
		// retry
		if err = broker.HandleRetry(ctx, ch, &msg); err != nil {
			logrus.WithFields(logrus.Fields{
				"msg_id": msg.MessageId,
				"err":    err.Error(),
			}).Warn("Retry fail")
		}
		return
	}

	span.AddEvent("order.update")
	logrus.Info("Consume ok")
}
