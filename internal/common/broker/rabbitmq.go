package broker

import (
	"context"
	"fmt"
	"time"

	_ "github.com/peiyouyao/gorder/common/config"
	"github.com/peiyouyao/gorder/common/logging"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
)

const (
	dlx                = "dlx"
	dlq                = "dlq"
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

	err = ch.ExchangeDeclare(dlx, "fanout", true, false, false, false, nil)
	if err != nil {
		return
	}

	err = ch.QueueBind(q.Name, "", dlx, false, nil)
	if err != nil {
		return
	}

	_, err = ch.QueueDeclare(dlq, true, false, false, false, nil)
	return
}

// confirm consumer **receive** a message
func HandleRetry(ctx context.Context, ch *amqp.Channel, d *amqp.Delivery) (err error) {
	fs := logrus.Fields{
		"q_msg_id": d.MessageId,
	}
	dlog := logging.LoggingWithCost(ctx, "mq_retry", fs)
	defer dlog(nil, err)
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
		err = ch.PublishWithContext(ctx, "", dlq, false, false, publishing)
		return
	}

	logrus.Infof("retrying message %s, count=%d", d.MessageId, retryCnt)
	time.Sleep(time.Second * time.Duration(retryCnt))
	err = ch.PublishWithContext(ctx, d.Exchange, d.RoutingKey, false, false, publishing)
	return
}

// impl propagation.TextMapCarrier
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
