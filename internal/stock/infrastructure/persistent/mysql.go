package persistent

import (
	"context"
	"fmt"
	"time"

	"github.com/peiyouyao/gorder/common/logging"
	"github.com/peiyouyao/gorder/stock/infrastructure/persistent/builder"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MySQL struct {
	db *gorm.DB
}

type StockModel struct {
	ID        int64     `gorm:"column:id"`
	ProductID string    `gorm:"column:product_id"`
	Quantity  int32     `gorm:"column:quantity"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
	Version   int64     `gorm:"column:version"` // 乐观锁版本号
}

func (m StockModel) TableName() string {
	return "o_stock"
}

func (m *StockModel) BeforeCreate(tx *gorm.DB) (err error) {
	m.UpdatedAt = time.Now()
	return nil
}

func NewMySQL() *MySQL {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		viper.GetString("mysql.user"),
		viper.GetString("mysql.password"),
		viper.GetString("mysql.host"),
		viper.GetString("mysql.port"),
		viper.GetString("mysql.db-name"),
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logrus.Panicf("init mysql wrong %v", err)
	}
	return &MySQL{db: db}
}

func NewMySQLWithDB(db *gorm.DB) *MySQL {
	if db == nil {
		logrus.Panic("db is nil")
	}
	return &MySQL{db: db}
}

func (d MySQL) StartTransaction(f func(tx *gorm.DB) error) error {
	return d.db.Transaction(f)
}

func (d MySQL) GetBatchByID(ctx context.Context, query *builder.Stock) (res []StockModel, err error) {
	_, logFn := logging.WhenMySQL(ctx, "GetBatchByID", query)
	tx := query.Fill(d.db.WithContext(ctx)).Find(&res)
	err = tx.Error
	defer logFn(res, &err)
	return res, err
}

func (d MySQL) GetByID(ctx context.Context, query *builder.Stock) (*StockModel, error) {
	_, deferLog := logging.WhenMySQL(ctx, "GetByID", query)
	var result StockModel
	tx := query.Fill(d.db.WithContext(ctx)).First(&result)
	defer deferLog(result, &tx.Error)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &result, nil
}

func (d MySQL) Update(ctx context.Context, tx *gorm.DB, cond *builder.Stock, update map[string]any) error {
	_, deferLog := logging.WhenMySQL(ctx, "BatchUpdateStock", cond)
	var returning StockModel
	res := cond.Fill(d.UseTransaction(tx).WithContext(ctx).Model(&returning).Clauses(clause.Returning{})).Updates(update)
	defer deferLog(returning, &res.Error)
	return res.Error
}

func (d *MySQL) UseTransaction(tx *gorm.DB) *gorm.DB {
	if tx == nil {
		return d.db
	}
	return tx
}

func (d MySQL) Create(ctx context.Context, tx *gorm.DB, create *StockModel) error {
	_, deferLog := logging.WhenMySQL(ctx, "Create", create)
	var returning StockModel
	err := d.UseTransaction(tx).WithContext(ctx).Model(&returning).Clauses(clause.Returning{}).Create(create).Error
	defer deferLog(returning, &err)
	return err
}
