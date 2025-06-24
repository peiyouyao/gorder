package adapters

import (
	"context"

	"github.com/peiyouyao/gorder/stock/entity"
	"github.com/peiyouyao/gorder/stock/infrastructure/persistent"
)

type MySQLStockRepository struct {
	db *persistent.MySQL
}

func NewMySQLStockRepository(db *persistent.MySQL) *MySQLStockRepository {
	return &MySQLStockRepository{db: db}
}

func (m *MySQLStockRepository) GetItems(ctx context.Context, ids []string) ([]*entity.Item, error) {
	panic("")
}
func (m *MySQLStockRepository) GetStock(ctx context.Context, ids []string) ([]*entity.ItemWithQuantity, error) {
	data, err := m.db.BatchGetStockByID(ctx, ids)
	if err != nil {
		return nil, err
	}
	var res []*entity.ItemWithQuantity
	for _, d := range data {
		res = append(res, &entity.ItemWithQuantity{
			ID:       d.ProductID,
			Quantity: d.Quantity,
		})
	}
	return res, nil
}
