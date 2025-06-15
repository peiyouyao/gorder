package broker

import (
	"context"
	"fmt"
	"time"

	_ "github.com/peiyouyao/gorder/common/config"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
)

const (
	DLX                = "dlx"
	DLQ                = "dlq"
	amqpRetryHeaderKey = "x-retry-count"
)

var (
	maxRetryCnt = viper.GetInt64("rabbitmq.max-retry")
)

func Connect(user, password, host, port string) (*amqp.Channel, func() error) {
	address := fmt.Sprintf("amqp://%s:%s@%s:%s", user, password, host, port)
	conn, err := amqp.Dial(address)
	if err != nil {
		logrus.Fatal(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		logrus.Fatal(err)
	}

	err = ch.ExchangeDeclare(EventOrderCreated, "direct", true, false, false, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}

	err = ch.ExchangeDeclare(EventOrderPaid, "fanout", true, false, false, false, nil)
	if err != nil {
		logrus.Fatal(err)
	}

	err = createDLX(ch)
	if err != nil {
		logrus.Fatal(err)
	}
	return ch, conn.Close
}

func createDLX(ch *amqp.Channel) (err error) {
	q, err := ch.QueueDeclare("share_mq", true, false, false, false, nil)
	if err != nil {
		return
	}

	err = ch.ExchangeDeclare(DLX, "fanout", true, false, false, false, nil)
	if err != nil {
		return
	}

	err = ch.QueueBind(q.Name, "", DLX, false, nil)
	if err != nil {
		return
	}

	_, err = ch.QueueDeclare(DLQ, true, false, false, false, nil)
	return
}

func HandleRetry(ctx context.Context, ch *amqp.Channel, d *amqp.Delivery) error {
	if d.Headers == nil {
		d.Headers = amqp.Table{}
	}
	retryCnt, ok := d.Headers[amqpRetryHeaderKey].(int64)
	if !ok {
		retryCnt = 0
	}
	retryCnt++
	d.Headers[amqpRetryHeaderKey] = retryCnt

	publishing := amqp.Publishing{
		Headers:      d.Headers,
		ContentType:  "application/json",
		Body:         d.Body,
		DeliveryMode: amqp.Persistent,
	}

	if retryCnt >= maxRetryCnt {
		logrus.Infof("moveing message %s to dlq", d.MessageId)
		return ch.PublishWithContext(ctx, "", DLQ, false, false, publishing)
	}

	logrus.Infof("retrying message %s, count=%d", d.MessageId, retryCnt)
	time.Sleep(time.Second * time.Duration(retryCnt))
	return ch.PublishWithContext(ctx, d.Exchange, d.RoutingKey, false, false, publishing)
}

type RabbitMQHeaderCarrier map[string]interface{}

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

func InjectRabbitMQHeaders(ctx context.Context) map[string]interface{} {
	carrier := make(RabbitMQHeaderCarrier)
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return carrier
}

func ExtractRabbitMQHeaders(ctx context.Context, headrs map[string]interface{}) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, RabbitMQHeaderCarrier(headrs))
}
