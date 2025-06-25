package models

import (
	"time"

	"github.com/adshao/go-binance/v2"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"type:varchar(255);uniqueIndex" json:"username"`
	Password  string    `gorm:"type:varchar(255)" json:"-"` // 不序列化
	APIKey    string    `gorm:"type:varchar(500)" json:"apiKey"`
	SecretKey string    `gorm:"type:varchar(500)" json:"secretKey"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Symbol struct {
	gorm.Model
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"index"`
	Symbol    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Price struct {
	gorm.Model
	ID        uint      `gorm:"primaryKey" json:"id"`
	Symbol    string    `gorm:"unique" json:"symbol"`
	Price     string    `json:"price"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Trade struct {
	gorm.Model
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"index"`
	Symbol    string
	Price     float64
	Qty       float64
	Time      int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Strategy struct {
	gorm.Model
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserID          uint      `gorm:"index" json:"userId"`
	Symbol          string    `gorm:"type:varchar(50)" json:"symbol"`
	StrategyType    string    `gorm:"type:varchar(20)" json:"strategyType"` // simple, iceberg, custom
	Side            string    `gorm:"type:varchar(10)" json:"side"`         // BUY, SELL
	Price           float64   `json:"price"`
	TotalQuantity   float64   `json:"totalQuantity"`
	Status          string    `gorm:"type:varchar(20);default:'active'" json:"status"`
	Enabled         bool      `gorm:"default:true" json:"enabled"`
	BuyQuantities   string    `gorm:"type:text" json:"buyQuantities"` // 逗号分隔的百分比
	SellQuantities  string    `gorm:"type:text" json:"sellQuantities"`
	BuyDepthLevels  string    `gorm:"type:text" json:"buyDepthLevels"` // 逗号分隔的深度级别
	SellDepthLevels string    `gorm:"type:text" json:"sellDepthLevels"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	PendingBatch    bool      `gorm:"default:false" json:"pendingBatch"` // 标记是否有活跃订单批次
}

type Order struct {
	gorm.Model
	ID          uint      `gorm:"primaryKey" json:"id"`
	StrategyID  uint      `gorm:"index" json:"strategyId"`
	UserID      uint      `gorm:"index" json:"userId"`
	Symbol      string    `gorm:"type:varchar(50)" json:"symbol"`
	Side        string    `gorm:"type:varchar(10)" json:"side"`
	Price       float64   `json:"price"`
	Quantity    float64   `json:"quantity"`
	OrderID     int64     `gorm:"index" json:"orderId"`
	Status      string    `gorm:"type:varchar(20)" json:"status"`
	CancelAfter time.Time `json:"cancelAfter"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Withdrawal struct {
	gorm.Model
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"userId"`
	Asset     string    `gorm:"type:varchar(20)" json:"asset"`
	Amount    float64   `json:"amount"`
	Address   string    `gorm:"type:varchar(500)" json:"address"`
	Threshold float64   `json:"threshold"`
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	Status    string    `gorm:"type:varchar(20)" json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type WithdrawalHistory struct {
	gorm.Model
	ID           uint      `gorm:"primaryKey" json:"id"`
	UserID       uint      `gorm:"index" json:"userId"`
	Asset        string    `gorm:"type:varchar(20)" json:"asset"`
	Amount       float64   `json:"amount"`
	Address      string    `gorm:"type:varchar(500)" json:"address"`
	WithdrawalID string    `gorm:"type:varchar(100)" json:"withdrawalId"`
	TxID         string    `gorm:"type:varchar(100)" json:"txId"`
	Status       string    `gorm:"type:varchar(20)" json:"status"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type CustomSymbol struct {
	gorm.Model
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"userId"`
	Symbol    string    `gorm:"type:varchar(50)" json:"symbol"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Balance struct {
	binance.Balance
}

// MigrateDB 自动迁移数据库表
func MigrateDB(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Symbol{},
		&Price{},
		&Trade{},
		&Strategy{},
		&Order{},
		&Withdrawal{},
		&WithdrawalHistory{},
		&CustomSymbol{},
	)
}
