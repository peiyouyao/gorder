package main

import (
	"github.com/PerryYao-GitHub/gorder/order/app"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	app app.Application
}

func (s *HTTPServer) PostCustomerCustomerIDOrder(c *gin.Context, customerID string) {

}

func (s *HTTPServer) GetCustomerCustomerIDOrderOrderID(c *gin.Context, customerID string, orderID string) {

}
