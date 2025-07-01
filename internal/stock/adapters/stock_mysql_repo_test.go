package adapters

import (
	"context"
	"fmt"
	"sync"
	"testing"

	_ "github.com/peiyouyao/gorder/common/config"
	"github.com/peiyouyao/gorder/stock/entity"
	"github.com/peiyouyao/gorder/stock/infrastructure/persistent"
	"github.com/peiyouyao/gorder/stock/infrastructure/persistent/builder"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestMySQLStockRepo_UpdateStock_Race(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)

	var (
		ctx                = context.Background()
		testItem           = "test-race-item"
		initialStock int32 = 100
	)

	err := db.Create(ctx, nil, &persistent.StockModel{
		ProductID: testItem,
		Quantity:  initialStock,
	})
	assert.NoError(t, err)

	repo := NewMySQLStockRepository(db)

	var wg sync.WaitGroup
	goroutines := 10
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := repo.UpdateStock(
				ctx,
				[]*entity.ItemWithQuantity{{ID: testItem, Quantity: 1}},
				func(
					ctx context.Context,
					existing []*entity.ItemWithQuantity,
					query []*entity.ItemWithQuantity) ([]*entity.ItemWithQuantity, error) {
					var newItems []*entity.ItemWithQuantity
					for _, e := range existing {
						for _, q := range query {
							if e.ID == q.ID && e.Quantity >= q.Quantity {
								newItems = append(newItems, &entity.ItemWithQuantity{
									ID:       e.ID,
									Quantity: e.Quantity - q.Quantity,
								})
							}
						}
					}
					return newItems, nil
				},
			)
			assert.NoError(t, err, "UpdateStock failed in goroutine")
		}()
	}
	wg.Wait()

	query := builder.NewStock().ProductIDs(testItem)
	res, err := db.GetBatchByID(ctx, query)
	assert.NoError(t, err, "BatchGetStockByID failed")
	assert.NotEmpty(t, res, "Expected stock record to exist after updates")

	expected := initialStock - int32(goroutines)
	assert.Equal(t, expected, res[0].Quantity)
}

func TestMySQLStockRepo_UpdateStock_OverSell(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)

	var (
		ctx                = context.Background()
		testItem           = "test-oversell-item"
		initialStock int32 = 5
	)

	err := db.Create(ctx, nil, &persistent.StockModel{
		ProductID: testItem,
		Quantity:  initialStock,
	})
	assert.NoError(t, err)

	repo := NewMySQLStockRepository(db)

	var wg sync.WaitGroup
	goroutines := 10
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := repo.UpdateStock(
				ctx,
				[]*entity.ItemWithQuantity{{ID: testItem, Quantity: 1}},
				func(
					ctx context.Context,
					existing []*entity.ItemWithQuantity,
					query []*entity.ItemWithQuantity) ([]*entity.ItemWithQuantity, error) {
					var newItems []*entity.ItemWithQuantity
					for _, e := range existing {
						for _, q := range query {
							if e.ID == q.ID && e.Quantity >= q.Quantity {
								newItems = append(newItems, &entity.ItemWithQuantity{
									ID:       e.ID,
									Quantity: e.Quantity - q.Quantity,
								})
							}
						}
					}
					return newItems, nil
				},
			)
			assert.NoError(t, err, "UpdateStock failed in goroutine")
		}()
	}
	wg.Wait()

	query := builder.NewStock().ProductIDs(testItem)
	res, err := db.GetBatchByID(ctx, query)
	assert.NoError(t, err, "BatchGetStockByID failed")
	assert.NotEmpty(t, res, "Expected stock record to exist after updates")

	fmt.Println(res[0])
	assert.GreaterOrEqual(t, res[0].Quantity, int32(0))
}

func setupTestDB(t *testing.T) *persistent.MySQL {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		viper.GetString("mysql.user"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetString("mysql.port"),
		"",
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)

	testDB := viper.GetString("mysql.db-name") + "_shadow"

	assert.NoError(t, db.Exec("DROP DATABASE IF EXISTS "+testDB).Error)
	assert.NoError(t, db.Exec("CREATE DATABASE "+testDB).Error)

	dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		viper.GetString("mysql.user"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetString("mysql.port"),
		testDB,
	)
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	})
	assert.NoError(t, err)
	assert.NoError(t, db.AutoMigrate(&persistent.StockModel{}), "Failed to migrate StockModel")

	return persistent.NewMySQLWithDB(db)
}
