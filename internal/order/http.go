package main

import (
	"fmt"
	"net/http"

	client "github.com/PerryYao-GitHub/gorder/common/client/order"
	"github.com/PerryYao-GitHub/gorder/common/tracing"
	"github.com/PerryYao-GitHub/gorder/order/app"
	"github.com/PerryYao-GitHub/gorder/order/app/command"
	"github.com/PerryYao-GitHub/gorder/order/app/query"
	"github.com/PerryYao-GitHub/gorder/order/convertor"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	app app.Application
}

func (H *HTTPServer) PostCustomerCustomerIDOrders(c *gin.Context, customerID string) {
	ctx, span := tracing.Start(c, "PostCustomerCustomerIDOrders")
	defer span.End()

	var req client.CreateOrderRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	r, err := H.app.Commands.CreateOrder.Handle(ctx, command.CreateOrder{
		CustomerID: req.CustomerID,
		Items:      convertor.NewItemWithQuantityConvertor().ClientsToEntities(req.Items),
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":      "sucess",
		"customer_id":  req.CustomerID,
		"order_id":     r.OrderID,
		"redirect_url": fmt.Sprintf("http://localhost:8282/success?customerID=%s&orderID=%s", req.CustomerID, r.OrderID),
		"trace_id":     tracing.TraceID(ctx),
	})
}

func (H *HTTPServer) GetCustomerCustomerIDOrdersOrderID(c *gin.Context, customerID string, orderID string) {
	ctx, span := tracing.Start(c, "PostCustomerCustomerIDOrders")
	defer span.End()

	o, err := H.app.Queries.GetCustomerOrder.Handle(ctx, query.GetCustomerOrder{
		CustomerID: customerID,
		OrderID:    orderID,
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"Order": o,
		},
		"trace_id": tracing.TraceID(ctx),
	})
}
