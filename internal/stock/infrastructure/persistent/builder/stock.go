package builder

import (
	"encoding/json"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Stock struct {
	ID_        []int64  `json:"id,omitempty"`
	ProductID_ []string `json:"product_id,omitempty"`
	Quantity_  []int32  `json:"quantity,omitempty"`
	Version_   []int64  `json:"version,omitempty"`

	// extend fields
	OrderBy_   string `json:"order_by,omitempty"`
	ForUpdate_ bool   `json:"for_update,omitempty"`
}

func NewStock() *Stock {
	return &Stock{}
}

// implement ArgFormatter interface
func (s *Stock) FormatArg() (string, error) {
	bytes, err := json.Marshal(s)
	return string(bytes), err
}

func (s *Stock) Fill(db *gorm.DB) *gorm.DB {
	db = s.fillWhere(db)
	if s.OrderBy_ != "" {
		db = db.Order(s.OrderBy_)
	}
	return db
}
func (s *Stock) fillWhere(db *gorm.DB) *gorm.DB {
	if len(s.ID_) > 0 {
		db = db.Where("id in (?)", s.ID_)
	}
	if len(s.ProductID_) > 0 {
		db = db.Where("product_id in (?)", s.ProductID_)
	}
	if len(s.Version_) > 0 {
		db = db.Where("version in (?)", s.Version_)
	}
	if len(s.Quantity_) > 0 {
		db = s.fillQuantityGT(db)
	}

	if s.ForUpdate_ {
		db = db.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate})
	}
	return db
}
func (s *Stock) fillQuantityGT(db *gorm.DB) *gorm.DB {
	db = db.Where("quantity >= ?", s.Quantity_)
	return db
}

func (s *Stock) IDs(v ...int64) *Stock {
	s.ID_ = v
	return s
}

func (s *Stock) ProductIDs(v ...string) *Stock {
	s.ProductID_ = v
	return s
}

func (s *Stock) OrderBy(v string) *Stock {
	s.OrderBy_ = v
	return s
}

func (s *Stock) Versions(v ...int64) *Stock {
	s.Version_ = v
	return s
}

func (s *Stock) QuantityGT(v ...int32) *Stock {
	s.Quantity_ = v
	return s
}

func (s *Stock) ForUpdate() *Stock {
	s.ForUpdate_ = true
	return s
}
