package app

import "github.com/PerryYao-GitHub/gorder/order/app/query"

type Application struct {
	Commands Commands
	Queries  Queries
}

type Commands struct{}

type Queries struct {
	GetCustomerOrder query.GetCustomerOrderHandler
}
