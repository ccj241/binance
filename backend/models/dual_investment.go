package models

import (
	"gorm.io/gorm"
	"time"
)

// DualInvestmentProduct 双币投资产品
type DualInvestmentProduct struct {
	gorm.Model
	ID             uint      `gorm:"primaryKey" json:"id"`
	Symbol         string    `gorm:"type:varchar(50)" json:"symbol"`                  // 币种对，如BTCUSDT
	Direction      string    `gorm:"type:varchar(10)" json:"direction"`               // UP(看涨)/DOWN(看跌)
	StrikePrice    float64   `json:"strikePrice" gorm:"comment:执行价格"`             // 执行价格
	APY            float64   `json:"apy" gorm:"comment:年化收益率"`                   // 年化收益率(百分比)
	Duration       int       `json:"duration" gorm:"comment:期限(天)"`                // 投资期限（天）
	MinAmount      float64   `json:"minAmount" gorm:"comment:最小投资额"`             // 最小投资额
	MaxAmount      float64   `json:"maxAmount" gorm:"comment:最大投资额"`             // 最大投资额
	SettlementTime time.Time `json:"settlementTime" gorm:"comment:结算时间"`          // 结算时间
	ProductID      string    `gorm:"type:varchar(100)" json:"productId"`              // 币安产品ID
	Status         string    `gorm:"type:varchar(20);default:'active'" json:"status"` // active/expired/sold_out
	BaseAsset      string    `gorm:"type:varchar(20)" json:"baseAsset"`               // 基础资产
	QuoteAsset     string    `gorm:"type:varchar(20)" json:"quoteAsset"`              // 计价资产
	CurrentPrice   float64   `json:"currentPrice" gorm:"comment:当前价格"`            // 当前市场价格
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// DualInvestmentStrategy 双币投资策略
type DualInvestmentStrategy struct {
	gorm.Model
	ID                   uint    `gorm:"primaryKey" json:"id"`
	UserID               uint    `gorm:"index" json:"userId"`
	StrategyName         string  `gorm:"type:varchar(100)" json:"strategyName"`          // 策略名称
	StrategyType         string  `gorm:"type:varchar(50)" json:"strategyType"`           // single/auto_reinvest/ladder/price_trigger
	BaseAsset            string  `gorm:"type:varchar(20)" json:"baseAsset"`              // 基础资产
	QuoteAsset           string  `gorm:"type:varchar(20)" json:"quoteAsset"`             // 计价资产
	DirectionPreference  string  `gorm:"type:varchar(20)" json:"directionPreference"`    // UP/DOWN/BOTH
	TargetAPYMin         float64 `json:"targetApyMin" gorm:"comment:目标最小年化收益率"` // 目标最小年化收益率
	TargetAPYMax         float64 `json:"targetApyMax" gorm:"comment:目标最大年化收益率"` // 目标最大年化收益率
	MaxSingleAmount      float64 `json:"maxSingleAmount" gorm:"comment:单笔最大投资额"`  // 单笔最大投资额
	TotalInvestmentLimit float64 `json:"totalInvestmentLimit" gorm:"comment:总投资限额"` // 总投资限额
	CurrentInvested      float64 `json:"currentInvested" gorm:"comment:当前已投资金额"`  // 当前已投资金额
	// 风险参数JSON字段
	MaxStrikePriceOffset float64 `json:"maxStrikePriceOffset" gorm:"comment:最大执行价格偏离度(%)"` // 最大执行价格偏离度
	MinDuration          int     `json:"minDuration" gorm:"comment:最小投资期限(天)"`               // 最小投资期限
	MaxDuration          int     `json:"maxDuration" gorm:"comment:最大投资期限(天)"`               // 最大投资期限
	MaxPositionRatio     float64 `json:"maxPositionRatio" gorm:"comment:最大仓位比例(%)"`           // 占总资产最大比例
	AutoReinvest         bool    `gorm:"default:false" json:"autoReinvest" gorm:"comment:自动复投"` // 是否自动复投
	// 价格触发策略参数
	TriggerPrice float64 `json:"triggerPrice" gorm:"comment:触发价格"` // 触发价格
	TriggerType  string  `gorm:"type:varchar(20)" json:"triggerType"`  // above/below
	// 梯度策略参数 - 修改：不再使用百分比
	LadderSteps       int        `json:"ladderSteps" gorm:"comment:梯度层数"`             // 梯度层数
	LadderStepPercent float64    `json:"ladderStepPercent" gorm:"comment:已弃用"`         // 已弃用，保留字段兼容性
	BasePrice         float64    `json:"basePrice" gorm:"comment:基准价格"`               // 梯度投资的基准价格
	Enabled           bool       `gorm:"default:true" json:"enabled"`                     // 是否启用
	Status            string     `gorm:"type:varchar(20);default:'active'" json:"status"` // active/paused/completed
	LastExecutedAt    *time.Time `json:"lastExecutedAt" gorm:"comment:最后执行时间"`      // 最后执行时间
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

// DualInvestmentOrder 双币投资订单
type DualInvestmentOrder struct {
	gorm.Model
	ID               uint       `gorm:"primaryKey" json:"id"`
	UserID           uint       `gorm:"index" json:"userId"`
	StrategyID       *uint      `gorm:"index" json:"strategyId"`                  // 可能为空（手动下单）
	ProductID        uint       `gorm:"index" json:"productId"`                   // 关联产品ID
	OrderID          string     `gorm:"type:varchar(100);index" json:"orderId"`   // 币安订单ID
	Symbol           string     `gorm:"type:varchar(50)" json:"symbol"`           // 币种对
	InvestAsset      string     `gorm:"type:varchar(20)" json:"investAsset"`      // 投资币种
	InvestAmount     float64    `json:"investAmount" gorm:"comment:投资金额"`     // 投资金额
	StrikePrice      float64    `json:"strikePrice" gorm:"comment:执行价格"`      // 执行价格
	APY              float64    `json:"apy" gorm:"comment:年化收益率"`            // 年化收益率
	Direction        string     `gorm:"type:varchar(10)" json:"direction"`        // UP/DOWN
	Duration         int        `json:"duration" gorm:"comment:期限(天)"`         // 期限（天）
	SettlementTime   time.Time  `json:"settlementTime" gorm:"comment:结算时间"`   // 结算时间
	SettlementAsset  string     `gorm:"type:varchar(20)" json:"settlementAsset"`  // 结算币种
	SettlementAmount float64    `json:"settlementAmount" gorm:"comment:结算金额"` // 结算金额
	ActualAPY        float64    `json:"actualApy" gorm:"comment:实际年化收益率"`  // 实际年化收益率
	Status           string     `gorm:"type:varchar(20)" json:"status"`           // pending/active/settled/cancelled
	SettledAt        *time.Time `json:"settledAt" gorm:"comment:实际结算时间"`    // 实际结算时间
	PnL              float64    `json:"pnl" gorm:"comment:盈亏金额"`              // 盈亏金额
	PnLPercent       float64    `json:"pnlPercent" gorm:"comment:盈亏百分比"`     // 盈亏百分比
	Notes            string     `gorm:"type:text" json:"notes"`                   // 备注
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

// DualInvestmentStats 双币投资统计
type DualInvestmentStats struct {
	UserID          uint    `json:"userId"`
	TotalInvested   float64 `json:"totalInvested"`   // 总投资金额
	TotalSettled    float64 `json:"totalSettled"`    // 总结算金额
	TotalPnL        float64 `json:"totalPnl"`        // 总盈亏
	TotalPnLPercent float64 `json:"totalPnlPercent"` // 总盈亏百分比
	WinCount        int     `json:"winCount"`        // 盈利次数
	LossCount       int     `json:"lossCount"`       // 亏损次数
	WinRate         float64 `json:"winRate"`         // 胜率
	AverageAPY      float64 `json:"averageApy"`      // 平均年化收益率
	ActiveOrders    int     `json:"activeOrders"`    // 活跃订单数
	ActiveAmount    float64 `json:"activeAmount"`    // 活跃投资金额
}

// TableName 指定表名
func (DualInvestmentProduct) TableName() string {
	return "dual_investment_products"
}

func (DualInvestmentStrategy) TableName() string {
	return "dual_investment_strategies"
}

func (DualInvestmentOrder) TableName() string {
	return "dual_investment_orders"
}

// 添加到数据库迁移
func MigrateDualInvestmentTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&DualInvestmentProduct{},
		&DualInvestmentStrategy{},
		&DualInvestmentOrder{},
	)
}
