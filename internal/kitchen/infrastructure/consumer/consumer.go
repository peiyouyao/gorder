package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/peiyouyao/gorder/common/broker"
	"github.com/peiyouyao/gorder/common/constants"
	"github.com/peiyouyao/gorder/common/convert"
	"github.com/peiyouyao/gorder/common/entity"
	"github.com/peiyouyao/gorder/common/genproto/orderpb"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

type OrderService interface {
	UpdateOrder(ctx context.Context, request *orderpb.Order) error
}

/*
消费 mq 中 order.paid 消息
*/
type Consumer struct {
	orderGRPC OrderService
}

func NewConsumer(orderGRPC OrderService) *Consumer {
	return &Consumer{orderGRPC: orderGRPC}
}

func (c *Consumer) Listen(ch *amqp.Channel) {
	q, err := ch.QueueDeclare("", true, false, true, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}

	if err = ch.QueueBind(q.Name, "", broker.EventOrderPaid, false, nil); err != nil {
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

func (c *Consumer) handleMessage(ch *amqp.Channel, msg amqp.Delivery, q amqp.Queue) {
	logrus.WithFields(logrus.Fields{
		"from_q": q.Name,
		"msg_id": msg.MessageId,
	}).Info("Receive order.create msg")

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

	o := &entity.Order{}
	if err = json.Unmarshal(msg.Body, o); err != nil {
		logrus.WithField("err", err.Error()).Warn("Unmarshal fail")
		return
	}

	if o.Status != constants.OrderStatusPaid {
		err = errors.New("order not paid can not cook")
		return
	}
	cook(ctx, o)

	span.AddEvent(fmt.Sprintf("order.cooked.%v", o))
	if err := c.orderGRPC.UpdateOrder(ctx, &orderpb.Order{
		ID:          o.ID,
		CustomerID:  o.CustomerID,
		Status:      constants.OrderStatusReady,
		Items:       convert.ItemEntitiesToProtos(o.Items),
		PaymentLink: o.PaymentLink,
	}); err != nil {
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

	span.AddEvent("kitchen.order.finished.updated")
	logrus.Info("Consume create.order ok")
}

func cook(ctx context.Context, o *entity.Order) {
	logrus.WithContext(ctx).Infof("Cooking order_id=%s", o.ID)
	time.Sleep(5 * time.Second)
	logrus.WithContext(ctx).Infof("Cooking done order_id=%s", o.ID)
}
