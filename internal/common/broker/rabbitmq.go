package broker

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

func Connect(user, password, host, port string) (*amqp.Channel, func() error) {
	address := fmt.Sprintf("amqp://%s:%s@%s:%s", user, password, host, port)
	conn, e := amqp.Dial(address)
	if e != nil {
		logrus.Fatal(e)
	}
	ch, e1 := conn.Channel()
	if e1 != nil {
		logrus.Fatal(e1)
	}
	e2 := ch.ExchangeDeclare(EventOrderCreated, "direct", true, false, false, false, nil)
	if e2 != nil {
		logrus.Fatal(e2)
	}
	e3 := ch.ExchangeDeclare(EventOrderPaid, "fanout", true, false, false, false, nil)
	if e3 != nil {
		logrus.Fatal(e3)
	}
	return ch, conn.Close
}

type RabbitMQHeaderCarrier map[string]interface{}

/*
func (r *RabbitMQHeaderCarrier) Get(key string) string {
	value, ok := (*r)[key]
	if !ok {
		return ""
	}
	return value.(string)
}

func (r *RabbitMQHeaderCarrier) Set(key, value string) {
	(*r)[key] = value
}

func (r *RabbitMQHeaderCarrier) Keys() []string {
	keys := make([]string, len(*r))
	i := 0
	for key := range *r {
		keys[i] = key
		i++
	}
	return keys
}

func InjectRannitMQHeaders(ctx context.Context) map[string]interface{} {
	carrier := make(RabbitMQHeaderCarrier)
	otel.GetTextMapPropagator().Inject(ctx, &carrier)
	return carrier
}
*/

func (r RabbitMQHeaderCarrier) Get(key string) string {
	value, ok := r[key]
	if !ok {
		return ""
	}
	return value.(string)
}

func (r RabbitMQHeaderCarrier) Set(key, value string) {
	r[key] = value
}

func (r RabbitMQHeaderCarrier) Keys() []string {
	keys := make([]string, len(r))
	i := 0
	for key := range r {
		keys[i] = key
		i++
	}
	return keys
}

func InjectRannitMQHeaders(ctx context.Context) map[string]interface{} {
	carrier := make(RabbitMQHeaderCarrier)
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return carrier
}

func ExtractRabbitMQHeaders(ctx context.Context, headrs map[string]interface{}) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, RabbitMQHeaderCarrier(headrs))
}
