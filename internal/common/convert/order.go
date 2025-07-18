package convert

import (
	client "github.com/peiyouyao/gorder/common/client/order"
	"github.com/peiyouyao/gorder/common/entity"
	"github.com/peiyouyao/gorder/common/genproto/orderpb"
)

func OrderEntityToProto(o *entity.Order) *orderpb.Order {
	check(o)
	return &orderpb.Order{
		ID:          o.ID,
		CustomerID:  o.CustomerID,
		Status:      o.Status,
		Items:       ItemEntitiesToProtos(o.Items),
		PaymentLink: o.PaymentLink,
	}
}

func OrderProtoToEntity(o *orderpb.Order) *entity.Order {
	check(o)
	return &entity.Order{
		ID:          o.ID,
		CustomerID:  o.CustomerID,
		Status:      o.Status,
		PaymentLink: o.PaymentLink,
		Items:       ItemProtosToEntities(o.Items),
	}
}

func OrderClientToEntity(o *client.Order) *entity.Order {
	check(o)
	return &entity.Order{
		ID:          o.Id,
		CustomerID:  o.CustomerId,
		Status:      o.Status,
		PaymentLink: o.PaymentLink,
		Items:       ItemClientsToEntities(o.Items),
	}
}

func OrderEntityToClient(o *entity.Order) *client.Order {
	check(o)
	return &client.Order{
		Id:          o.ID,
		CustomerId:  o.CustomerID,
		Status:      o.Status,
		PaymentLink: o.PaymentLink,
		Items:       ItemEntitiesToClients(o.Items),
	}
}

func check(o interface{}) {
	if o == nil {
		panic("connot convert nil order")
	}
}
