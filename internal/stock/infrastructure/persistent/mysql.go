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

func (d MySQL) StartTransaction(f func(tx *gorm.DB) error) error {
	return d.db.Transaction(f)
}

func (d MySQL) BatchGetStockByID(ctx context.Context, query *builder.Stock) (res []StockModel, err error) {
	_, logFn := logging.WhenMySQL(ctx, "BatchGetStockByID", query)
	tx := query.Fill(d.db.WithContext(ctx)).Find(&res)
	err = tx.Error
	defer logFn(res, &err)
	return res, err
}

func NewMySQLWithDB(db *gorm.DB) *MySQL {
	if db == nil {
		logrus.Panic("db is nil")
	}
	return &MySQL{db: db}
}

func (d MySQL) Create(ctx context.Context, create *StockModel) (err error) {
	_, logFn := logging.WhenMySQL(ctx, "Create", create)
	err = d.db.WithContext(ctx).Create(create).Error
	defer logFn(create, &err)
	return err
}
