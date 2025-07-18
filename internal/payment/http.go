package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/peiyouyao/gorder/common/broker"
	"github.com/peiyouyao/gorder/common/constants"
	"github.com/peiyouyao/gorder/common/entity"
	"github.com/peiyouyao/gorder/common/logging"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
	"go.opentelemetry.io/otel"
)

type PaymentHandler struct {
	channel *amqp.Channel
}

func NewPaymentHandler(ch *amqp.Channel) *PaymentHandler {
	return &PaymentHandler{channel: ch}
}

// stripe listen --forward-to localhost:8284/api/webhook
func (h *PaymentHandler) RegisterRoutes(c *gin.Engine) {
	c.POST("/api/webhook", h.handleWebhook)
}

func (h *PaymentHandler) handleWebhook(c *gin.Context) {
	logging.Infof(c.Request.Context(), nil, "receive webhook from stripe")

	var err error
	defer func() {
		if err != nil {
			logging.Warnf(c.Request.Context(), nil, "handleWebhook err=%v", err)
		} else {
			logging.Infof(c.Request.Context(), nil, "handleWebhook success")
		}
	}()

	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logrus.Infof("Error reading request body: %v\n", err)
		c.JSON(http.StatusServiceUnavailable, err.Error())
		return
	}

	event, err := webhook.ConstructEvent(payload, c.Request.Header.Get("Stripe-Signature"),
		viper.GetString("ENDPOINT_STRIPE_SECRET"))

	if err != nil {
		logrus.Infof("Error verifying webhook signature: %v\n", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	switch event.Type {
	case stripe.EventTypeCheckoutSessionCompleted:
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			logrus.Infof("error unmarshal event.data.raw into session, err = %v", err)
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		if session.PaymentStatus == stripe.CheckoutSessionPaymentStatusPaid {
			var items []*entity.Item
			_ = json.Unmarshal([]byte(session.Metadata["items"]), &items)

			tr := otel.Tracer("rabbitmq")
			ctx, span := tr.Start(
				c.Request.Context(),
				fmt.Sprintf("rabbitmq.%s.publish", broker.EventOrderPaid),
			)
			defer span.End()

			err = broker.PublishEvent(ctx, broker.PublishEventReq{
				Channel: h.channel,
				Routing: broker.Fanout,
				Queue:   "",
				Body: entity.NewOrder(
					session.Metadata["orderID"],
					session.Metadata["customerID"],
					constants.OrderStatusPaid,
					session.Metadata["paymentLink"],
					items,
				),
			})
		}
	}
	c.JSON(http.StatusOK, nil)
}
