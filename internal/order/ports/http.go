package ports

import (
	"fmt"

	"github.com/peiyouyao/gorder/common/constants"
	myerrors "github.com/peiyouyao/gorder/common/handler/errors"

	"errors"

	"github.com/gin-gonic/gin"
	client "github.com/peiyouyao/gorder/common/client/order"
	"github.com/peiyouyao/gorder/common/convert"
	common "github.com/peiyouyao/gorder/common/response"
	"github.com/peiyouyao/gorder/order/app"
	"github.com/peiyouyao/gorder/order/app/command"
	"github.com/peiyouyao/gorder/order/app/dto"
	"github.com/peiyouyao/gorder/order/app/query"
)

type HTTPServer struct {
	common.BaseResponse // 继承
	App                 app.Application
}

func (s *HTTPServer) PostCustomerCustomerIdOrders(c *gin.Context, customerID string) {
	var (
		req  client.CreateOrderRequest
		err  error
		resp dto.CreateOrderResponse
	)
	defer func() {
		s.Response(c, err, &resp)
	}()

	if err = c.ShouldBind(&req); err != nil {
		err = myerrors.NewWithError(constants.ErrnoBindRequest, err)
		return
	}

	if err = s.validate(&req); err != nil {
		err = myerrors.NewWithError(constants.ErrnoInvalidParams, err)
		return
	}
	r, err := s.App.Commands.CreateOrder.Handle(c.Request.Context(), command.CreateOrder{
		CustomerID: req.CustomerId,
		Items:      convert.ItemWithQuantityClientsToEntities(req.Items),
	})
	if err != nil {
		return
	}

	resp.CustomerID = req.CustomerId
	resp.OrderID = r.OrderID
	resp.RedirectURL = fmt.Sprintf("http://localhost:8282/success?customerID=%s&orderID=%s", req.CustomerId, r.OrderID)
}

func (s *HTTPServer) GetCustomerCustomerIdOrdersOrderId(c *gin.Context, customerID string, orderID string) {
	var (
		err  error
		resp struct {
			Order *client.Order `json:"order"`
		}
	)
	defer func() {
		s.Response(c, err, resp)
	}()

	o, err := s.App.Queries.GetCustomerOrder.Handle(c.Request.Context(), query.GetCustomerOrder{
		CustomerID: customerID,
		OrderID:    orderID,
	})
	if err != nil {
		return
	}
	resp.Order = &client.Order{
		Id:          o.ID,
		CustomerId:  o.CustomerID,
		Status:      o.Status,
		Items:       convert.ItemEntitiesToClients(o.Items),
		PaymentLink: o.PaymentLink,
	}
}

func (s *HTTPServer) validate(req *client.CreateOrderRequest) error {
	if req == nil || req.Items == nil {
		return errors.New("nil req or nil items")
	}
	for _, iq := range req.Items {
		if iq.Quantity <= 0 {
			return errors.New("negative quantity")
		}
	}
	return nil
}
