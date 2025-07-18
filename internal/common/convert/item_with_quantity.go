package convert

import (
	client "github.com/peiyouyao/gorder/common/client/order"
	"github.com/peiyouyao/gorder/common/entity"
	"github.com/peiyouyao/gorder/common/genproto/orderpb"
)

func ItemWithQuantityEntitiesToProtos(items []*entity.ItemWithQuantity) (res []*orderpb.ItemWithQuantity) {
	for _, i := range items {
		res = append(res, ItemWithQuantityEntityToProto(i))
	}
	return
}

func ItemWithQuantityEntityToProto(i *entity.ItemWithQuantity) *orderpb.ItemWithQuantity {
	return &orderpb.ItemWithQuantity{
		ID:       i.ID,
		Quantity: i.Quantity,
	}
}

func ItemWithQuantityProtosToEntities(items []*orderpb.ItemWithQuantity) (res []*entity.ItemWithQuantity) {
	for _, i := range items {
		res = append(res, ItemWithQuantityProtoToEntity(i))
	}
	return
}

func ItemWithQuantityProtoToEntity(i *orderpb.ItemWithQuantity) *entity.ItemWithQuantity {
	return &entity.ItemWithQuantity{
		ID:       i.ID,
		Quantity: i.Quantity,
	}
}

func ItemWithQuantityClientsToEntities(items []client.ItemWithQuantity) (res []*entity.ItemWithQuantity) {
	for _, i := range items {
		res = append(res, ItemWithQuantityClientToEntity(i))
	}
	return
}

func ItemWithQuantityClientToEntity(i client.ItemWithQuantity) *entity.ItemWithQuantity {
	return &entity.ItemWithQuantity{
		ID:       i.Id,
		Quantity: i.Quantity,
	}
}
