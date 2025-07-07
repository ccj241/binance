package models

import (
	"gorm.io/gorm"
	"time"
)

// FuturesStrategy 永续期货策略
type FuturesStrategy struct {
	gorm.Model
	ID                uint       `gorm:"primaryKey" json:"id"`
	UserID            uint       `gorm:"index" json:"userId"`
	StrategyName      string     `gorm:"type:varchar(100)" json:"strategyName"`                     // 策略名称
	Symbol            string     `gorm:"type:varchar(50)" json:"symbol"`                            // 交易对，如BTCUSDT
	Side              string     `gorm:"type:varchar(10)" json:"side"`                              // LONG/SHORT
	StrategyType      string     `gorm:"type:varchar(20);default:'simple'" json:"strategyType"`     // simple/iceberg
	BasePrice         float64    `json:"basePrice" gorm:"comment:基准价格"`                             // 触发价格
	EntryPrice        float64    `json:"entryPrice" gorm:"comment:开仓价格"`                            // 限价单价格（将在触发时计算）
	EntryPriceFloat   float64    `json:"entryPriceFloat" gorm:"comment:开仓价格浮动千分比"`                  // 开仓价格浮动千分比
	Leverage          int        `json:"leverage" gorm:"comment:杠杆倍数"`                              // 杠杆倍数 1-125
	Quantity          float64    `json:"quantity" gorm:"comment:开仓数量"`                              // 开仓数量（USDT）
	TakeProfitRate    float64    `json:"takeProfitRate" gorm:"comment:止盈百分比"`                       // 止盈百分比（扣除手续费后）
	TakeProfitPrice   float64    `json:"takeProfitPrice" gorm:"comment:止盈价格"`                       // 计算后的止盈价格
	StopLossRate      float64    `json:"stopLossRate" gorm:"comment:止损百分比"`                         // 止损百分比（可选）
	StopLossPrice     float64    `json:"stopLossPrice" gorm:"comment:止损价格"`                         // 计算后的止损价格
	MarginType        string     `gorm:"type:varchar(20);default:'CROSSED'" json:"marginType"`      // ISOLATED/CROSSED
	IcebergLevels     int        `json:"icebergLevels" gorm:"default:5;comment:冰山层数"`               // 冰山策略层数
	IcebergQuantities string     `gorm:"type:text;comment:冰山策略各层数量比例" json:"icebergQuantities"`     // 逗号分隔的比例
	IcebergPriceGaps  string     `gorm:"type:text;comment:冰山策略各层价格间隔(万分比)" json:"icebergPriceGaps"` // 逗号分隔的万分比
	Enabled           bool       `gorm:"default:true" json:"enabled"`                               // 是否启用
	Status            string     `gorm:"type:varchar(20);default:'waiting'" json:"status"`          // waiting/triggered/position_opened/completed/cancelled
	TriggeredAt       *time.Time `json:"triggeredAt" gorm:"comment:触发时间"`                           // 触发时间
	CompletedAt       *time.Time `json:"completedAt" gorm:"comment:完成时间"`                           // 完成时间
	CurrentPositionId int64      `json:"currentPositionId" gorm:"comment:当前持仓ID"`                   // 币安持仓ID
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

// FuturesOrder 永续期货订单
type FuturesOrder struct {
	gorm.Model
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserID          uint      `gorm:"index" json:"userId"`
	StrategyID      uint      `gorm:"index" json:"strategyId"`                 // 关联策略
	Symbol          string    `gorm:"type:varchar(50)" json:"symbol"`          // 交易对
	Side            string    `gorm:"type:varchar(10)" json:"side"`            // BUY/SELL
	PositionSide    string    `gorm:"type:varchar(10)" json:"positionSide"`    // LONG/SHORT
	Type            string    `gorm:"type:varchar(20)" json:"type"`            // LIMIT/MARKET/STOP_MARKET等
	Price           float64   `json:"price"`                                   // 价格
	Quantity        float64   `json:"quantity"`                                // 数量
	OrderID         int64     `gorm:"index" json:"orderId"`                    // 币安订单ID
	Status          string    `gorm:"type:varchar(20)" json:"status"`          // NEW/FILLED/CANCELED等
	OrderPurpose    string    `gorm:"type:varchar(20)" json:"orderPurpose"`    // entry/take_profit/stop_loss
	ExecutedQty     float64   `json:"executedQty" gorm:"comment:已成交数量"`        // 已成交数量
	AvgPrice        float64   `json:"avgPrice" gorm:"comment:平均成交价"`           // 平均成交价
	Commission      float64   `json:"commission" gorm:"comment:手续费"`           // 手续费
	CommissionAsset string    `gorm:"type:varchar(20)" json:"commissionAsset"` // 手续费资产
	RealizedPnl     float64   `json:"realizedPnl" gorm:"comment:已实现盈亏"`        // 已实现盈亏
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// FuturesPosition 永续期货持仓记录
type FuturesPosition struct {
	gorm.Model
	ID               uint       `gorm:"primaryKey" json:"id"`
	UserID           uint       `gorm:"index" json:"userId"`
	StrategyID       uint       `gorm:"index" json:"strategyId"`              // 关联策略
	Symbol           string     `gorm:"type:varchar(50)" json:"symbol"`       // 交易对
	PositionSide     string     `gorm:"type:varchar(10)" json:"positionSide"` // LONG/SHORT
	EntryPrice       float64    `json:"entryPrice" gorm:"comment:开仓均价"`       // 开仓均价
	Quantity         float64    `json:"quantity" gorm:"comment:持仓数量"`         // 持仓数量
	UnrealizedPnl    float64    `json:"unrealizedPnl" gorm:"comment:未实现盈亏"`   // 未实现盈亏
	RealizedPnl      float64    `json:"realizedPnl" gorm:"comment:已实现盈亏"`     // 已实现盈亏
	Leverage         int        `json:"leverage" gorm:"comment:杠杆倍数"`         // 杠杆倍数
	MarginType       string     `gorm:"type:varchar(20)" json:"marginType"`   // ISOLATED/CROSSED
	IsolatedMargin   float64    `json:"isolatedMargin" gorm:"comment:逐仓保证金"`  // 逐仓保证金
	MarkPrice        float64    `json:"markPrice" gorm:"comment:标记价格"`        // 标记价格
	LiquidationPrice float64    `json:"liquidationPrice" gorm:"comment:强平价格"` // 强平价格
	Status           string     `gorm:"type:varchar(20)" json:"status"`       // open/closed
	OpenedAt         time.Time  `json:"openedAt" gorm:"comment:开仓时间"`         // 开仓时间
	ClosedAt         *time.Time `json:"closedAt" gorm:"comment:平仓时间"`         // 平仓时间
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

// FuturesStats 永续期货统计
type FuturesStats struct {
	UserID           uint    `json:"userId"`
	TotalTrades      int     `json:"totalTrades"`      // 总交易次数
	WinTrades        int     `json:"winTrades"`        // 盈利次数
	LossTrades       int     `json:"lossTrades"`       // 亏损次数
	WinRate          float64 `json:"winRate"`          // 胜率
	TotalPnl         float64 `json:"totalPnl"`         // 总盈亏
	TotalCommission  float64 `json:"totalCommission"`  // 总手续费
	NetPnl           float64 `json:"netPnl"`           // 净盈亏（扣除手续费）
	AveragePnl       float64 `json:"averagePnl"`       // 平均盈亏
	MaxWin           float64 `json:"maxWin"`           // 最大盈利
	MaxLoss          float64 `json:"maxLoss"`          // 最大亏损
	ActivePositions  int     `json:"activePositions"`  // 当前持仓数
	ActiveStrategies int     `json:"activeStrategies"` // 活跃策略数
}

// CalculateTakeProfitPrice 计算止盈价格
func (s *FuturesStrategy) CalculateTakeProfitPrice() {
	// 手续费率（币安期货默认）
	takerFeeRate := 0.0004 // 0.04%

	// 计算净利润率（扣除开仓和平仓手续费）
	netProfitRate := s.TakeProfitRate / 100.0
	totalFeeRate := takerFeeRate * 2 // 开仓+平仓手续费

	if s.Side == "LONG" {
		// 多头：止盈价格 = 开仓价格 * (1 + 净利润率 + 手续费率)
		s.TakeProfitPrice = s.EntryPrice * (1 + netProfitRate + totalFeeRate)
	} else {
		// 空头：止盈价格 = 开仓价格 * (1 - 净利润率 - 手续费率)
		s.TakeProfitPrice = s.EntryPrice * (1 - netProfitRate - totalFeeRate)
	}
}

// CalculateStopLossPrice 计算止损价格
func (s *FuturesStrategy) CalculateStopLossPrice() {
	if s.StopLossRate <= 0 {
		return
	}

	stopLossRate := s.StopLossRate / 100.0

	if s.Side == "LONG" {
		// 多头：止损价格 = 开仓价格 * (1 - 止损率)
		s.StopLossPrice = s.EntryPrice * (1 - stopLossRate)
	} else {
		// 空头：止损价格 = 开仓价格 * (1 + 止损率)
		s.StopLossPrice = s.EntryPrice * (1 + stopLossRate)
	}
}

// TableName 指定表名
func (FuturesStrategy) TableName() string {
	return "futures_strategies"
}

func (FuturesOrder) TableName() string {
	return "futures_orders"
}

func (FuturesPosition) TableName() string {
	return "futures_positions"
}

// MigrateFuturesTables 迁移期货相关表
func MigrateFuturesTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&FuturesStrategy{},
		&FuturesOrder{},
		&FuturesPosition{},
	)
}
