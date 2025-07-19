package mq

import (
	"context"

	"github.com/peiyouyao/gorder/common/broker"
	domain "github.com/peiyouyao/gorder/order/domain/order"

	"github.com/rabbitmq/amqp091-go"
)

// impl domain.EventPublisher interface
type RabbitMQEventPublisher struct {
	Channel *amqp091.Channel
}

func (p *RabbitMQEventPublisher) Publish(ctx context.Context, event domain.DomainEvent) error {
	return broker.PublishEvent(ctx, broker.PublishEventReq{
		Channel:  p.Channel,
		Routing:  broker.Direct,
		Queue:    event.Dest,
		Exchange: "",
		Body:     event.Data,
	})
}

func (p *RabbitMQEventPublisher) Broadcast(ctx context.Context, event domain.DomainEvent) error {
	return broker.PublishEvent(ctx, broker.PublishEventReq{
		Channel:  p.Channel,
		Routing:  broker.Fanout,
		Queue:    event.Dest,
		Exchange: "",
		Body:     event.Data,
	})
}
