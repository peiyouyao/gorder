package broker

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/peiyouyao/gorder/common/logging"
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

func PublishEvent(ctx context.Context, p PublishEventReq) (err error) {
	_, dlog := logging.WhenEventPublish(ctx, p)
	defer dlog(nil, &err)

	if err = check(p); err != nil {
		return err
	}

	switch p.Routing {
	case Fanout:
		return fout(ctx, p)
	case Direct:
		return direct(ctx, p)
	default:
		logrus.WithContext(ctx).Panicf("unsupported routing type: %s", string(p.Routing))
	}
	return nil
}

func direct(ctx context.Context, p PublishEventReq) (err error) {
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

func fout(ctx context.Context, p PublishEventReq) (err error) {
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
		logging.Warnf(ctx, nil, "_publish_event_failed||exchange=%s||key=%s||msg=%v", exchange, key, msg)
		return err
	}
	return nil
}

func check(p PublishEventReq) error {
	if p.Channel == nil {
		return errors.New("nil channel")
	}
	return nil
}
