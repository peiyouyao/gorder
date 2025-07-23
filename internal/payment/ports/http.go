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

// 获取 paid.order 信息, 发布 paid.order
func (h *PaymentHandler) handleWebhook(c *gin.Context) {
	logrus.WithContext(c.Request.Context()).Info("Receive webhook from stripe")

	var err error
	defer func() {
		if err != nil {
			logrus.WithContext(c.Request.Context()).Warnf("Handle webhook fail err=%s", err.Error())
		} else {
			logrus.WithContext(c.Request.Context()).Info("Handle webhook ok")
		}
	}()

	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logrus.Infof("Read request body fail err=%v", err)
		c.JSON(http.StatusServiceUnavailable, err.Error())
		return
	}

	event, err := webhook.ConstructEvent(payload, c.Request.Header.Get("Stripe-Signature"),
		viper.GetString("ENDPOINT_STRIPE_SECRET"))

	if err != nil {
		logrus.Infof("Verifying webhook signature fail err=%v", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	switch event.Type {
	case stripe.EventTypeCheckoutSessionCompleted:
		logrus.Trace("User paid")
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			logrus.Errorf("Unmarshal event.Data.Raw fail err = %v", err)
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
			logrus.Tracef("Upate paymentLink and Status order=%v", o)

			logrus.Trace("broker.PublishEvent")
			err = broker.PublishEvent(ctx, &broker.PublishEventReq{
				Channel:  h.channel,
				Routing:  broker.Fanout,
				Exchange: broker.EventOrderPaid,
				Queue:    "",
				Body:     *o,
			})
			if err != nil {
				logrus.Trace("broker.PublishEvent fail")
			}
			logrus.Trace("broker.PublishEvent success")
		}
	}
	c.JSON(http.StatusOK, nil)
}
