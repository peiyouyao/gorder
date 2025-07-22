package ports

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/peiyouyao/gorder/common/broker"
	"github.com/peiyouyao/gorder/common/constants"
	"github.com/peiyouyao/gorder/common/entity"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
	"go.opentelemetry.io/otel"
)

/*
暴露.../api/webhook 接口, 供stripe调用(POST)
*/
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
	logrus.WithContext(c.Request.Context()).Info("receive_webhook_from_stripe")

	var err error
	defer func() {
		if err != nil {
			logrus.WithContext(c.Request.Context()).Warnf("handle_webhook_fail || err=%s", err.Error())
		} else {
			logrus.WithContext(c.Request.Context()).Info("handle_webhook_success")
		}
	}()

	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logrus.Infof("read_request_body_fail || err=%v", err)
		c.JSON(http.StatusServiceUnavailable, err.Error())
		return
	}

	event, err := webhook.ConstructEvent(payload, c.Request.Header.Get("Stripe-Signature"),
		viper.GetString("ENDPOINT_STRIPE_SECRET"))

	if err != nil {
		logrus.Infof("verifying_webhook_signature_fail || err=%v", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	switch event.Type {
	case stripe.EventTypeCheckoutSessionCompleted:
		logrus.Debug("user_paid")
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			logrus.Infof("unmarshal_event.Data.Raw_fail || err = %v", err)
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		if session.PaymentStatus == stripe.CheckoutSessionPaymentStatusPaid {
			tr := otel.Tracer("rabbitmq")
			ctx, span := tr.Start(
				c.Request.Context(),
				fmt.Sprintf("rabbitmq.%s.publish", broker.EventOrderPaid),
			)
			defer span.End()

			var items []*entity.Item
			_ = json.Unmarshal([]byte(session.Metadata["items"]), &items)

			o := entity.NewOrder(
				session.Metadata["orderID"],
				session.Metadata["customerID"],
				constants.OrderStatusPaid,
				session.Metadata["paymentLink"], // 没有取到
				items,
			)
			logrus.Debugf("receive_user_paid || order=%v", o)

			logrus.Debug("broker.PublishEvent")
			err = broker.PublishEvent(ctx, &broker.PublishEventReq{
				Channel:  h.channel,
				Routing:  broker.Fanout,
				Exchange: broker.EventOrderPaid,
				Queue:    "",
				Body:     *o,
			})
			if err != nil {
				logrus.Debug("broker.PublishEvent_fail")
			}
			logrus.Debug("broker.PublishEvent_success")
		}
	}
	c.JSON(http.StatusOK, nil)
}
