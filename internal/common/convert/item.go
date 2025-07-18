package convert

import (
	client "github.com/peiyouyao/gorder/common/client/order"
	"github.com/peiyouyao/gorder/common/entity"
	"github.com/peiyouyao/gorder/common/genproto/orderpb"
)

func ItemEntitiesToProtos(items []*entity.Item) (res []*orderpb.Item) {
	for _, i := range items {
		res = append(res, ItemEntityToProto(i))
	}
	return
}

func ItemProtosToEntities(items []*orderpb.Item) (res []*entity.Item) {
	for _, i := range items {
		res = append(res, ItemProtoToEntity(i))
	}
	return
}

func ItemClientsToEntities(items []client.Item) (res []*entity.Item) {
	for _, i := range items {
		res = append(res, ItemClientToEntity(i))
	}
	return
}

func ItemEntitiesToClients(items []*entity.Item) (res []client.Item) {
	for _, i := range items {
		res = append(res, ItemEntityToClient(i))
	}
	return
}

func ItemEntityToProto(i *entity.Item) *orderpb.Item {
	return &orderpb.Item{
		ID:       i.ID,
		Name:     i.Name,
		Quantity: i.Quantity,
		PriceID:  i.PriceID,
	}
}

func ItemProtoToEntity(i *orderpb.Item) *entity.Item {
	return &entity.Item{
		ID:       i.ID,
		Name:     i.Name,
		Quantity: i.Quantity,
		PriceID:  i.PriceID,
	}
}

func ItemClientToEntity(i client.Item) *entity.Item {
	return &entity.Item{
		ID:       i.Id,
		Name:     i.Name,
		Quantity: i.Quantity,
		PriceID:  i.PriceId,
	}
}

func ItemEntityToClient(i *entity.Item) client.Item {
	return client.Item{
		Id:       i.ID,
		Name:     i.Name,
		Quantity: i.Quantity,
		PriceId:  i.PriceID,
	}
}
