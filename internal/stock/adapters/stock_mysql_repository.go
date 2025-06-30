package adapters

import (
	"context"

	"github.com/pkg/errors"

	"github.com/peiyouyao/gorder/stock/entity"
	"github.com/peiyouyao/gorder/stock/infrastructure/persistent"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
		return nil, errors.Wrap(err, "failed to get stock by ID")
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

func (m MySQLStockRepository) UpdateStock(
	ctx context.Context,
	data []*entity.ItemWithQuantity,
	updateFn func(
		ctx context.Context,
		existing []*entity.ItemWithQuantity,
		query []*entity.ItemWithQuantity,
	) ([]*entity.ItemWithQuantity, error),
) error {
	return m.db.StartTransaction(func(tx *gorm.DB) (err error) {
		defer func() {
			if err != nil {
				logrus.Warnf("transaction failed, err = %v", err)
			}
		}()
		err = m.updateWithPessimisticLock(ctx, tx, data, updateFn)
		// err = m.updateWithOptimisticLock(ctx, tx, data, updateFn)
		return err
	})
}

// 悲观锁 (排他锁) SELECT * FROM o_stock WHERE product_id IN ? FOR UPDATE
func (m MySQLStockRepository) updateWithPessimisticLock(
	ctx context.Context,
	tx *gorm.DB,
	data []*entity.ItemWithQuantity,
	updateFn func(context.Context, []*entity.ItemWithQuantity, []*entity.ItemWithQuantity) ([]*entity.ItemWithQuantity, error)) (err error) {
	var dest []*persistent.StockModel
	if err = tx.Model(&persistent.StockModel{}).
		Clauses(clause.Locking{Strength: "UPDATE"}). // 悲观锁
		Where("product_id IN ?", getIDFromEntities(data)).Find(&dest).Error; err != nil {
		return errors.Wrap(err, "failed to get existing stock")
	}
	existing := m.unmarshalFromDatabase(dest)

	updated, err := updateFn(ctx, existing, data)
	if err != nil {
		return err
	}

	for _, upd := range updated {
		if err = tx.Model(&persistent.StockModel{}).Where("product_id = ?", upd.ID).Update("quantity", upd.Quantity).Error; err != nil {
			return errors.Wrap(err, "unable to update"+upd.ID)
		}
	}
	return nil
}

// 乐观锁
func (m MySQLStockRepository) updateWithOptimisticLock(
	ctx context.Context,
	tx *gorm.DB,
	data []*entity.ItemWithQuantity,
	updateFn func(context.Context, []*entity.ItemWithQuantity, []*entity.ItemWithQuantity) ([]*entity.ItemWithQuantity, error)) (err error) {
	var dest []*persistent.StockModel
	if err := tx.Model(&persistent.StockModel{}).
		Where("product_id IN (?)", getIDFromEntities(data)).
		Find(&dest).Error; err != nil {
		return errors.Wrap(err, "failed to find data")
	}

	for _, queryData := range data {
		var newest persistent.StockModel
		if err := tx.Model(&persistent.StockModel{}).Where("product_id = ?", queryData.ID).
			First(&newest).Error; err != nil {
			return err
		}
		if err := tx.Model(&persistent.StockModel{}).
			Where("product_id = ? AND version = ? AND quantity - ? >= 0", queryData.ID, newest.Version, queryData.Quantity).
			Updates(map[string]any{
				"quantity": gorm.Expr("quantity - ?", queryData.Quantity),
				"version":  newest.Version + 1,
			}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (m MySQLStockRepository) unmarshalFromDatabase(dest []*persistent.StockModel) []*entity.ItemWithQuantity {
	var result []*entity.ItemWithQuantity
	for _, i := range dest {
		result = append(result, &entity.ItemWithQuantity{
			ID:       i.ProductID,
			Quantity: i.Quantity,
		})
	}
	return result
}

func getIDFromEntities(items []*entity.ItemWithQuantity) []string {
	var ids []string
	for _, i := range items {
		ids = append(ids, i.ID)
	}
	return ids
}
