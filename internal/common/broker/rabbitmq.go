package broker

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
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
