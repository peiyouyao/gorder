package main

import (
	"fmt"

	client "github.com/PerryYao-GitHub/gorder/common/client/order"
	common "github.com/PerryYao-GitHub/gorder/common/response"
	"github.com/PerryYao-GitHub/gorder/order/app"
	"github.com/PerryYao-GitHub/gorder/order/app/command"
	"github.com/PerryYao-GitHub/gorder/order/app/query"
	"github.com/PerryYao-GitHub/gorder/order/convertor"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	common.BaseResponse
	app app.Application
}

func (H *HTTPServer) PostCustomerCustomerIDOrders(c *gin.Context, customerID string) {
	var (
		req  client.CreateOrderRequest
		err  error
		resp struct {
			CustomerID  string `json:"customer_id"`
			OrderID     string `json:"order_id"`
			RedirectURL string `json:"redirect_url"`
		}
	)
	defer func() {
		H.Response(c, err, &resp)
	}()

	if err = c.ShouldBind(&req); err != nil {
		return
	}
	r, err := H.app.Commands.CreateOrder.Handle(c.Request.Context(), command.CreateOrder{
		CustomerID: req.CustomerID,
		Items:      convertor.NewItemWithQuantityConvertor().ClientsToEntities(req.Items),
	})
	if err != nil {
		return
	}

	resp.CustomerID = req.CustomerID
	resp.OrderID = r.OrderID
	resp.RedirectURL = fmt.Sprintf("http://localhost:8282/success?customerID=%s&orderID=%s", req.CustomerID, r.OrderID)
}

func (H *HTTPServer) GetCustomerCustomerIDOrdersOrderID(c *gin.Context, customerID string, orderID string) {
	var (
		err  error
		resp struct {
			Order *client.Order `json:"order"`
		}
	)
	defer func() {
		H.Response(c, err, resp)
	}()

	o, err := H.app.Queries.GetCustomerOrder.Handle(c.Request.Context(), query.GetCustomerOrder{
		CustomerID: customerID,
		OrderID:    orderID,
	})
	if err != nil {
		return
	}
	resp.Order = convertor.NewOrderConvertor().EntityToClient(o)
}
