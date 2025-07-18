package adapters

import (
	"context"

	"github.com/pkg/errors"

	"github.com/peiyouyao/gorder/common/entity"
	"github.com/peiyouyao/gorder/stock/infrastructure/persistent"
	"github.com/peiyouyao/gorder/stock/infrastructure/persistent/builder"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type StockRepositoryMySQL struct {
	db *persistent.MySQL
}

func NewStockRepositoryMySQL(db *persistent.MySQL) *StockRepositoryMySQL {
	return &StockRepositoryMySQL{db: db}
}

func (m *StockRepositoryMySQL) GetItems(ctx context.Context, ids []string) ([]*entity.Item, error) {
	panic("")
}

func (m *StockRepositoryMySQL) GetStock(ctx context.Context, ids []string) ([]*entity.ItemWithQuantity, error) {
	data, err := m.db.GetBatchByID(ctx, builder.NewStock().ProductIDs(ids...))
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

func (m StockRepositoryMySQL) UpdateStock(
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
func (m StockRepositoryMySQL) updateWithPessimisticLock(
	ctx context.Context,
	tx *gorm.DB,
	data []*entity.ItemWithQuantity,
	updateFn func(context.Context, []*entity.ItemWithQuantity, []*entity.ItemWithQuantity) ([]*entity.ItemWithQuantity, error)) (err error) {
	var dest []persistent.StockModel

	dest, err = m.db.GetBatchByID(ctx, builder.NewStock().ProductIDs(getIDFromEntities(data)...).ForUpdate())
	if err != nil {
		return errors.Wrap(err, "failed to get existing stock")
	}

	existing := m.unmarshalFromDatabase(dest)
	updated, err := updateFn(ctx, existing, data)
	if err != nil {
		return err
	}

	for _, u := range updated {
		for _, query := range data {
			if u.ID != query.ID {
				continue
			}

			if err = m.db.Update(ctx, tx,
				builder.NewStock().ProductIDs(u.ID).QuantityGT(query.Quantity),
				map[string]any{"quantity": gorm.Expr("quantity - ?", query.Quantity)},
			); err != nil {
				return errors.Wrapf(err, "unable to update stock for product %s", u.ID)
			}
		}
	}
	return nil
}

// 乐观锁
func (m StockRepositoryMySQL) updateWithOptimisticLock(
	ctx context.Context,
	tx *gorm.DB,
	data []*entity.ItemWithQuantity,
	updateFn func(context.Context, []*entity.ItemWithQuantity, []*entity.ItemWithQuantity) ([]*entity.ItemWithQuantity, error)) (err error) {
	for _, query := range data {
		var newest *persistent.StockModel

		newest, err = m.db.GetByID(ctx, builder.NewStock().ProductIDs(query.ID))
		if err != nil {
			return err
		}

		if err = m.db.Update(
			ctx,
			tx,
			builder.NewStock().ProductIDs(query.ID).Versions(newest.Version).QuantityGT(query.Quantity),
			map[string]any{
				"quantity": gorm.Expr("quantity - ?", query.Quantity),
				"version":  newest.Version + 1,
			},
		); err != nil {
			return err
		}
	}
	return nil
}

func (m StockRepositoryMySQL) unmarshalFromDatabase(dest []persistent.StockModel) []*entity.ItemWithQuantity {
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
