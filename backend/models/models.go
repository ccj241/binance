// backend/models/models.go
package models

import (
	"github.com/adshao/go-binance/v2"
	"github.com/ccj241/binance/utils"
	"gorm.io/gorm"
	"time"
)

type User struct {
	gorm.Model
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"type:varchar(255);uniqueIndex" json:"username"`
	Password  string    `gorm:"type:varchar(255)" json:"-"`                       // 不序列化
	APIKey    string    `gorm:"type:varchar(500)" json:"-"`                       // 加密存储，不序列化
	SecretKey string    `gorm:"type:varchar(500)" json:"-"`                       // 加密存储，不序列化
	Role      string    `gorm:"type:varchar(20);default:'user'" json:"role"`      // admin, user
	Status    string    `gorm:"type:varchar(20);default:'pending'" json:"status"` // pending, active, disabled
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// BeforeSave 保存前加密API密钥
func (u *User) BeforeSave(tx *gorm.DB) error {
	// 如果APIKey不为空且看起来不是加密的（base64编码通常更长）
	if u.APIKey != "" && len(u.APIKey) < 100 {
		encrypted, err := utils.Encrypt(u.APIKey)
		if err != nil {
			return err
		}
		u.APIKey = encrypted
	}

	// 如果SecretKey不为空且看起来不是加密的
	if u.SecretKey != "" && len(u.SecretKey) < 100 {
		encrypted, err := utils.Encrypt(u.SecretKey)
		if err != nil {
			return err
		}
		u.SecretKey = encrypted
	}

	return nil
}

// GetDecryptedAPIKey 获取解密后的API Key
func (u *User) GetDecryptedAPIKey() (string, error) {
	if u.APIKey == "" {
		return "", nil
	}
	return utils.Decrypt(u.APIKey)
}

// GetDecryptedSecretKey 获取解密后的Secret Key
func (u *User) GetDecryptedSecretKey() (string, error) {
	if u.SecretKey == "" {
		return "", nil
	}
	return utils.Decrypt(u.SecretKey)
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
	ID              uint    `gorm:"primaryKey" json:"id"`
	UserID          uint    `gorm:"index" json:"userId"`
	Symbol          string  `gorm:"type:varchar(50)" json:"symbol"`
	StrategyType    string  `gorm:"type:varchar(20)" json:"strategyType"` // simple, iceberg, custom
	Side            string  `gorm:"type:varchar(10)" json:"side"`         // BUY, SELL
	Price           float64 `json:"price" gorm:"comment:触发价格"`        // 触发价格：买入策略在价格<=此值时触发，卖出策略在价格>=此值时触发
	TotalQuantity   float64 `json:"totalQuantity" gorm:"comment:总数量"`  // 策略的总交易数量
	Status          string  `gorm:"type:varchar(20);default:'active'" json:"status"`
	Enabled         bool    `gorm:"default:true" json:"enabled"`
	BuyQuantities   string  `gorm:"type:text;comment:买入数量分配(逗号分隔的比例)" json:"buyQuantities"`  // 逗号分隔的百分比，总和应为1.0
	SellQuantities  string  `gorm:"type:text;comment:卖出数量分配(逗号分隔的比例)" json:"sellQuantities"` // 逗号分隔的百分比，总和应为1.0
	BuyDepthLevels  string  `gorm:"type:text;comment:买入深度级别(逗号分隔)" json:"buyDepthLevels"`       // 逗号分隔的深度级别(1,2,3...)
	SellDepthLevels string  `gorm:"type:text;comment:卖出深度级别(逗号分隔)" json:"sellDepthLevels"`      // 逗号分隔的深度级别(1,2,3...)
	// 新增字段：万分比配置（预留）
	BuyBasisPoints     string    `gorm:"type:text;comment:买入价格偏移(万分比)" json:"buyBasisPoints"`         // 逗号分隔的万分比 (如: -10,-5,0,5,10)
	SellBasisPoints    string    `gorm:"type:text;comment:卖出价格偏移(万分比)" json:"sellBasisPoints"`        // 逗号分隔的万分比
	CancelAfterMinutes int       `gorm:"default:120;comment:订单自动取消时间(分钟)" json:"cancelAfterMinutes"` // 订单自动取消时间（分钟），默认120分钟
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
	PendingBatch       bool      `gorm:"default:false;comment:是否有待处理订单批次" json:"pendingBatch"` // 标记是否有活跃订单批次
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
	Status      string    `gorm:"type:varchar(20)" json:"status"` // pending, filled, cancelled, expired, rejected
	CancelAfter time.Time `json:"cancelAfter" gorm:"comment:自动取消时间"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Withdrawal struct {
	gorm.Model
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"userId"`
	Asset     string    `gorm:"type:varchar(20)" json:"asset"`
	Amount    float64   `json:"amount" gorm:"comment:提币金额，0表示提取全部"` // 0表示提取最大可用金额
	Address   string    `gorm:"type:varchar(500)" json:"address"`
	Threshold float64   `json:"threshold" gorm:"comment:触发阈值"` // 余额达到此值时触发提币
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	Status    string    `gorm:"type:varchar(20)" json:"status"` // active, paused
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
	Status       string    `gorm:"type:varchar(20)" json:"status"` // processing, completed, failed
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
