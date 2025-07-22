package persistent

import (
	"context"
	"fmt"
	"time"

	"github.com/peiyouyao/gorder/common/util"
	"github.com/peiyouyao/gorder/stock/infrastructure/persistent/builder"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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

type MySQL struct {
	db *gorm.DB
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
		panic("db is nil")
	}
	return &MySQL{db: db}
}

func (d MySQL) GetByID(ctx context.Context, query *builder.Stock) (res *StockModel, err error) {
	_, dlog := logMySQL(ctx, "GetByID", query)
	defer dlog(res, &err)

	err = query.Fill(d.db.WithContext(ctx)).First(&res).Error
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (d MySQL) GetBatchByID(ctx context.Context, query *builder.Stock) (res []StockModel, err error) {
	_, dlog := logMySQL(ctx, "GetBatchByID", query)
	defer dlog(res, &err)

	err = query.Fill(d.db.WithContext(ctx)).Find(&res).Error
	if err != nil {
		return nil, err
	}
	return
}

func (d MySQL) Update(ctx context.Context, tx *gorm.DB, cond *builder.Stock, update map[string]any) (err error) {
	var returning StockModel
	_, dlog := logMySQL(ctx, "UpdateBatch", cond)
	defer dlog(returning, &err)

	res := cond.Fill(d.useTransaction(tx).WithContext(ctx).Model(&returning).Clauses(clause.Returning{})).Updates(update)
	return res.Error
}

func (d MySQL) Create(ctx context.Context, tx *gorm.DB, create *StockModel) (err error) {
	var returning StockModel
	_, dlog := logMySQL(ctx, "Create", create)
	defer dlog(returning, &err)
	return d.useTransaction(tx).WithContext(ctx).Model(&returning).Clauses(clause.Returning{}).Create(create).Error
}

func (d MySQL) StartTransaction(f func(tx *gorm.DB) error) error {
	return d.db.Transaction(f)
}

func (d *MySQL) useTransaction(tx *gorm.DB) *gorm.DB {
	if tx == nil {
		return d.db
	}
	return tx
}

func logMySQL(ctx context.Context, cmd string, args ...any) (logrus.Fields, func(any, *error)) {
	fields := logrus.Fields{
		"mysql_cmd":  cmd,
		"mysql_args": util.FormatArgs(args),
	}
	start := time.Now()
	return fields, func(resp any, err *error) {
		level, msg := logrus.InfoLevel, "mysql_success"
		fields["mysql_cost"] = time.Since(start).Milliseconds()
		fields["mysql_resp"] = resp

		if err != nil && (*err != nil) {
			level, msg = logrus.ErrorLevel, "mysql_error"
			fields["mysql_err"] = (*err).Error()
		}

		logrus.WithContext(ctx).WithFields(fields).Log(level, msg)
	}
}
