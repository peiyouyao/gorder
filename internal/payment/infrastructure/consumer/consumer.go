package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/peiyouyao/gorder/common/broker"
	"github.com/peiyouyao/gorder/common/entity"
	"github.com/peiyouyao/gorder/payment/app"
	"github.com/peiyouyao/gorder/payment/app/command"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

/*
处理 order 通过 mq 传递的消息, 获取消息并创建链接
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
	q, err := ch.QueueDeclare(broker.EventOrderCreated, true, false, false, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		logrus.Warnf("consume_fail || queue=%s || err=%v", q.Name, err)
	}

	for msg := range msgs {
		c.handleMessage(ch, msg, q)
	}
}

func (c *Consumer) handleMessage(ch *amqp.Channel, msg amqp.Delivery, q amqp.Queue) {
	logrus.Infof("receive_order.create || from=%s || msg_id=%s", q.Name, msg.MessageId)

	ctx := broker.ExtractRabbitMQHeaders(context.Background(), msg.Headers)
	tr := otel.Tracer("rabbitmq")
	_, span := tr.Start(ctx, fmt.Sprintf("rabbitmq.%s.consume", q.Name))
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

	o := entity.Order{}
	if err = json.Unmarshal(msg.Body, &o); err != nil {
		logrus.Infof("unmarshal_fail || err=%s", err.Error())
		return
	}
	logrus.Debugf("created_order=%v", o)

	logrus.Debug("app.Commands.CreatePayment.Handle_start")
	if _, err = c.app.Commands.CreatePayment.Handle(ctx, command.CreatePayment{Order: &o}); err != nil {
		logrus.Warnf("create_payment_fail || order_id=%s || err=%s", o.ID, err.Error())
		// retry
		if err = broker.HandleRetry(ctx, ch, &msg); err != nil {
			logrus.Warnf("retry_fail || message_id=%s || err=%v", msg.MessageId, err)
		}
		return
	}
	logrus.Debug("app.Commands.CreatePayment.Handle_success")

	span.AddEvent("payment.craeted")
	logrus.Info("consume_success")
}
