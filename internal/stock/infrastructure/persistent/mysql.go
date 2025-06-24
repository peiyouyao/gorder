package persistent

import (
	"context"
	"fmt"
	"time"

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
}

func (m StockModel) TableName() string {
	return "o_stock"
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

func (d MySQL) BatchGetStockByID(ctx context.Context, productIDs []string) ([]StockModel, error) {
	var res []StockModel
	tx := d.db.WithContext(ctx).Where("product_id IN ?", productIDs).Find(&res)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return res, nil
}
