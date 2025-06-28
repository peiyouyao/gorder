package main

import (
	"fmt"

	"github.com/peiyouyao/gorder/common/consts"
	"github.com/peiyouyao/gorder/common/handler/errors"

	"github.com/gin-gonic/gin"
	client "github.com/peiyouyao/gorder/common/client/order"
	common "github.com/peiyouyao/gorder/common/response"
	"github.com/peiyouyao/gorder/order/app"
	"github.com/peiyouyao/gorder/order/app/command"
	"github.com/peiyouyao/gorder/order/app/dto"
	"github.com/peiyouyao/gorder/order/app/query"
	"github.com/peiyouyao/gorder/order/convertor"
)

type HTTPServer struct {
	common.BaseResponse
	app app.Application
}

func (H *HTTPServer) PostCustomerCustomerIdOrders(c *gin.Context, customerID string) {
	var (
		req  client.CreateOrderRequest
		err  error
		resp dto.CreateOrderResponse
	)
	defer func() {
		H.Response(c, err, &resp)
	}()

	if err = c.ShouldBind(&req); err != nil {
		err = errors.NewWithError(consts.ErrnoBindRequest, err)
		return
	}
	if !H.validate(&req) {
		err = errors.NewWithError(consts.ErrnoInvalidParams, err)
		return
	}
	r, err := H.app.Commands.CreateOrder.Handle(c.Request.Context(), command.CreateOrder{
		CustomerID: req.CustomerId,
		Items:      convertor.NewItemWithQuantityConvertor().ClientsToEntities(req.Items),
	})
	if err != nil {
		return
	}

	resp.CustomerID = req.CustomerId
	resp.OrderID = r.OrderID
	resp.RedirectURL = fmt.Sprintf("http://localhost:8282/success?customerID=%s&orderID=%s", req.CustomerId, r.OrderID)
}

func (H *HTTPServer) GetCustomerCustomerIdOrdersOrderId(c *gin.Context, customerID string, orderID string) {
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

func (H *HTTPServer) validate(req *client.CreateOrderRequest) bool {
	for _, iq := range req.Items {
		if iq.Quantity <= 0 {
			return false
		}
	}
	return true
}
