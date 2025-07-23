package broker

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/peiyouyao/gorder/common/util"
	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

const (
	EventOrderCreated = "order.created"
	EventOrderPaid    = "order.paid"
)

type RoutingType string

const (
	Fanout RoutingType = "fanout"
	Direct RoutingType = "direct"
)

type PublishEventReq struct {
	Channel  *amqp091.Channel
	Routing  RoutingType
	Queue    string
	Exchange string
	Body     any
}

func PublishEvent(ctx context.Context, p *PublishEventReq) (err error) {
	_, dlog := logPublishing(ctx, p)
	defer dlog(&err)

	if err = check(p); err != nil {
		return err
	}

	switch p.Routing {
	case Fanout:
		return fout(ctx, p)
	case Direct:
		return direct(ctx, p)
	default:
		logrus.
			WithContext(ctx).
			WithField("routing_type", string(p.Routing)).
			Panicf("Unsupported routing type")
	}
	return nil
}

func direct(ctx context.Context, p *PublishEventReq) (err error) {
	if _, err = p.Channel.QueueDeclare(p.Queue, true, false, false, false, nil); err != nil {
		return
	}

	jsonBody, err := json.Marshal(p.Body)
	if err != nil {
		return
	}

	return doPublish(ctx, p.Channel, p.Exchange, p.Queue, false, false, amqp091.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp091.Persistent,
		Body:         jsonBody,
		Headers:      InjectRabbitMQHeaders(ctx),
	})
}

func fout(ctx context.Context, p *PublishEventReq) (err error) {
	jsonBody, err := json.Marshal(p.Body)
	if err != nil {
		return
	}

	return doPublish(ctx, p.Channel, p.Exchange, "", false, false, amqp091.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp091.Persistent,
		Body:         jsonBody,
		Headers:      InjectRabbitMQHeaders(ctx),
	})
}

func doPublish(ctx context.Context, ch *amqp091.Channel, exchange, key string, mandatory, immediate bool, msg amqp091.Publishing) error {
	if err := ch.PublishWithContext(ctx, exchange, key, mandatory, immediate, msg); err != nil {
		logrus.WithContext(ctx).WithFields(logrus.Fields{
			"exchange": exchange,
			"key":      key,
			"q_msg":    msg,
		}).Warn("Publish event fail")
		return err
	}
	return nil
}

func check(p *PublishEventReq) error {
	if p.Channel == nil {
		return errors.New("nil channel")
	}
	return nil
}

func logPublishing(ctx context.Context, p *PublishEventReq) (logrus.Fields, func(*error)) {
	fields := logrus.Fields{
		"queue":    p.Queue,
		"routing":  p.Routing,
		"exchange": p.Exchange,
		"body":     util.MarshalStringWithoutErr(p.Body),
	}
	start := time.Now()
	return fields, func(err *error) {
		level, msg := logrus.InfoLevel, "MQ publish ok"
		fields["publish_time_cost"] = time.Since(start).Milliseconds()

		if err != nil && (*err != nil) {
			level, msg = logrus.ErrorLevel, "MQ publish fail"
			fields["publish_error"] = (*err).Error()
		}

		logrus.WithContext(ctx).WithFields(fields).Log(level, msg)
	}
}
