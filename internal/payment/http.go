package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type PaymentHander struct{}

func NewPaymentHandler() *PaymentHander {
	return &PaymentHander{}
}

func (h *PaymentHander) RegisterRoutes(c *gin.Engine) {
	c.POST("/api/webhook", h.handleWebhook)
}

func (h *PaymentHander) handleWebhook(ctx *gin.Context) {
	logrus.Info("got webhook from stripe")
}
