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
消费 mq 中 order.paid 消息
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
	logrus.Infof("receive_order.paid || from=%s || msg_id=%s", q.Name, msg.MessageId)
	logrus.Debug("receive_order.paid")

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
			logrus.WithContext(ctx).WithFields(fs).Warn("mq_consume_failed")
			_ = msg.Nack(false, false)
		} else {
			logrus.WithContext(ctx).Info("mq_consume_success")
			_ = msg.Ack(false)
		}
	}()

	o := &domain.Order{}
	if err = json.Unmarshal(msg.Body, o); err != nil {
		logrus.Warnf("unmarshal_fail || err=%s", err.Error())
		return
	}
	logrus.Debugf("paid_order=%v", *o)

	logrus.Debug("app.Commands.UpdateOrder.Handle_start")
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
		logrus.WithContext(ctx).WithFields(fs).Error("update_order_fail")
		// retry
		if err = broker.HandleRetry(ctx, ch, &msg); err != nil {
			logrus.Warnf("retry_fail || message_id=%s || err=%v", msg.MessageId, err)
		}
		return
	}

	span.AddEvent("order.update")
	logrus.Info("consume_success")
}
