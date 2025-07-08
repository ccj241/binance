package tasks

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// FuturesMonitor 期货价格监控器（暂未使用，预留接口）
// var FuturesMonitor sync.Map

// FuturesWebSocketManager 期货WebSocket管理器
type FuturesWebSocketManager struct {
	symbol       string
	strategies   sync.Map // strategyID -> *models.FuturesStrategy
	cfg          *config.Config
	wsConn       *websocket.Conn
	stopChan     chan struct{}
	mu           sync.RWMutex
	reconnecting bool
	lastPrice    float64
}

// StartFuturesMonitoring 启动期货监控
func StartFuturesMonitoring(cfg *config.Config) {
	// 启动价格监控
	go monitorFuturesPrices(cfg)

	// 启动持仓监控
	go monitorFuturesPositions(cfg)

	// 启动订单状态检查
	go checkFuturesOrders(cfg)
}

// monitorFuturesPrices 监控期货价格
func monitorFuturesPrices(cfg *config.Config) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	wsManagers := make(map[string]*FuturesWebSocketManager)

	log.Println("期货价格监控已启动")

	for range ticker.C {
		// 获取所有等待中的策略
		var strategies []models.FuturesStrategy
		if err := cfg.DB.Where("enabled = ? AND status = ? AND deleted_at IS NULL",
			true, "waiting").Find(&strategies).Error; err != nil {
			log.Printf("查询策略失败: %v", err)
			continue
		}

		// 只在策略数量变化时输出日志
		// log.Printf("找到 %d 个等待中的策略", len(strategies))

		// 按交易对分组
		symbolStrategies := make(map[string][]models.FuturesStrategy)
		for _, strategy := range strategies {
			symbolStrategies[strategy.Symbol] = append(symbolStrategies[strategy.Symbol], strategy)
		}

		// 为每个交易对创建或更新WebSocket连接
		for symbol, strats := range symbolStrategies {
			if manager, exists := wsManagers[symbol]; exists {
				// 更新策略列表
				for _, s := range strats {
					manager.strategies.Store(s.ID, &s)
				}
			} else {
				// 创建新的WebSocket连接
				manager := &FuturesWebSocketManager{
					symbol:   symbol,
					cfg:      cfg,
					stopChan: make(chan struct{}),
				}
				for _, s := range strats {
					manager.strategies.Store(s.ID, &s)
				}
				wsManagers[symbol] = manager
				go manager.start()
			}
		}

		// 清理不再需要的连接
		for symbol, manager := range wsManagers {
			if _, exists := symbolStrategies[symbol]; !exists {
				close(manager.stopChan)
				delete(wsManagers, symbol)
			}
		}
	}
}

// start 启动WebSocket连接
func (m *FuturesWebSocketManager) start() {
	// 将交易对转换为小写
	wsURL := fmt.Sprintf("wss://fstream.binance.com/ws/%s@markPrice@1s", strings.ToLower(m.symbol))

	// 减少日志输出
	// log.Printf("准备连接 %s 的 WebSocket", m.symbol)

	for {
		select {
		case <-m.stopChan:
			return
		default:
			m.connect(wsURL)
			if m.wsConn != nil {
				if err := m.wsConn.Close(); err != nil {
					log.Printf("关闭WebSocket连接失败: %v", err)
				}
			}
			time.Sleep(5 * time.Second) // 重连间隔
		}
	}
}

// connect 建立WebSocket连接
func (m *FuturesWebSocketManager) connect(wsURL string) {
	// 减少连接日志
	// log.Printf("正在连接 WebSocket: %s", wsURL)

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		log.Printf("期货WebSocket连接失败 %s: %v", m.symbol, err)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("关闭WebSocket连接失败: %v", err)
		}
	}()

	m.wsConn = conn
	// 减少连接成功日志
	// log.Printf("WebSocket 连接成功: %s", m.symbol)

	for {
		select {
		case <-m.stopChan:
			return
		default:
			var msg map[string]interface{}
			if err := conn.ReadJSON(&msg); err != nil {
				log.Printf("期货WebSocket读取错误 %s: %v", m.symbol, err)
				return
			}

			// 解析标记价格
			if markPriceStr, ok := msg["p"].(string); ok {
				if markPrice, err := strconv.ParseFloat(markPriceStr, 64); err == nil {
					// 移除价格更新日志
					// log.Printf("%s 收到价格更新: %.8f", m.symbol, markPrice)
					m.mu.Lock()
					m.lastPrice = markPrice
					m.mu.Unlock()

					// 检查策略触发
					m.checkStrategies(markPrice)
				}
			}
		}
	}
}

// checkStrategies 检查策略是否触发
func (m *FuturesWebSocketManager) checkStrategies(currentPrice float64) {
	// 移除调试日志
	// log.Printf("检查 %s 策略，当前价格: %.8f", m.symbol, currentPrice)

	m.strategies.Range(func(key, value interface{}) bool {
		strategy := value.(*models.FuturesStrategy)

		// 移除策略详情日志
		// log.Printf("策略 %d: %s %s, 基准价格: %.8f, 当前价格: %.8f, 状态: %s, 启用: %v",
		// 	strategy.ID, strategy.Side, strategy.Symbol, strategy.BasePrice, currentPrice,
		// 	strategy.Status, strategy.Enabled)

		// 检查是否触发
		shouldTrigger := false
		if strategy.Side == "LONG" && currentPrice <= strategy.BasePrice {
			shouldTrigger = true
			// 减少触发条件日志
			// log.Printf("策略 %d 满足做多触发条件", strategy.ID)
		} else if strategy.Side == "SHORT" && currentPrice >= strategy.BasePrice {
			shouldTrigger = true
			// 减少触发条件日志
			// log.Printf("策略 %d 满足做空触发条件", strategy.ID)
		}

		if shouldTrigger {
			// 使用事务确保并发安全
			err := m.cfg.DB.Transaction(func(tx *gorm.DB) error {
				// 重新查询策略状态
				var currentStrategy models.FuturesStrategy
				if err := tx.Set("gorm:query_option", "FOR UPDATE").
					First(&currentStrategy, strategy.ID).Error; err != nil {
					return err
				}

				// 双重检查状态
				if currentStrategy.Status != "waiting" || !currentStrategy.Enabled {
					return fmt.Errorf("策略状态已变更")
				}

				// 更新状态为已触发
				currentStrategy.Status = "triggered"
				now := time.Now()
				currentStrategy.TriggeredAt = &now

				if err := tx.Save(&currentStrategy).Error; err != nil {
					return err
				}

				// 保留策略触发的关键日志
				log.Printf("期货策略 %d 触发: %s %s @ %.8f",
					strategy.ID, strategy.Side, strategy.Symbol, currentPrice)

				// 异步执行开仓
				go m.executeStrategy(&currentStrategy)

				return nil
			})

			if err != nil && err.Error() != "策略状态已变更" {
				log.Printf("更新策略状态失败: %v", err)
			}

			// 从监控中移除已触发的策略
			m.strategies.Delete(strategy.ID)
		}

		return true
	})
}

// executeStrategy 执行策略开仓
func (m *FuturesWebSocketManager) executeStrategy(strategy *models.FuturesStrategy) {
	// 获取用户信息
	var user models.User
	if err := m.cfg.DB.First(&user, strategy.UserID).Error; err != nil {
		log.Printf("获取用户信息失败: %v", err)
		return
	}

	// 解密API密钥
	apiKey, err := user.GetDecryptedAPIKey()
	if err != nil {
		log.Printf("解密API Key失败: %v", err)
		return
	}
	secretKey, err := user.GetDecryptedSecretKey()
	if err != nil {
		log.Printf("解密Secret Key失败: %v", err)
		return
	}

	// 创建期货客户端
	client := binance.NewFuturesClient(apiKey, secretKey)

	// 根据策略类型执行不同的开仓逻辑
	switch strategy.StrategyType {
	case "iceberg":
		// 执行冰山策略
		m.executeIcebergStrategy(strategy, client)
	case "slow_iceberg":
		// 执行慢冰山策略
		m.executeSlowIcebergStrategy(strategy, client)
	default:
		// 执行简单策略
		m.executeSimpleStrategy(strategy, client)
	}
}

// executeSimpleStrategy 执行简单策略
func (m *FuturesWebSocketManager) executeSimpleStrategy(strategy *models.FuturesStrategy, client *futures.Client) {
	// 设置杠杆
	if err := setLeverage(client, strategy.Symbol, strategy.Leverage); err != nil {
		log.Printf("设置杠杆失败: %v", err)
		updateStrategyStatus(m.cfg.DB, strategy, "cancelled", err.Error())
		return
	}

	// 设置保证金模式（忽略已存在的错误）
	if err := setMarginType(client, strategy.Symbol, strategy.MarginType); err != nil {
		// 检查是否是"不需要更改"的错误
		if !strings.Contains(err.Error(), "No need to change margin type") {
			log.Printf("设置保证金模式失败: %v", err)
			// 其他错误继续执行，不取消策略
		}
	}

	// 获取交易规则（精度信息）
	exchangeInfo, err := client.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		log.Printf("获取交易规则失败: %v", err)
		updateStrategyStatus(m.cfg.DB, strategy, "cancelled", err.Error())
		return
	}

	// 查找当前交易对的规则
	var symbolInfo *futures.Symbol
	for _, s := range exchangeInfo.Symbols {
		if s.Symbol == strategy.Symbol {
			symbolInfo = &s
			break
		}
	}

	if symbolInfo == nil {
		log.Printf("未找到交易对 %s 的规则", strategy.Symbol)
		updateStrategyStatus(m.cfg.DB, strategy, "cancelled", "未找到交易对规则")
		return
	}

	// 获取价格和数量精度
	var pricePrecision int
	var quantityPrecision int
	var tickSize float64
	var stepSize float64
	var minQty float64

	// 直接使用 Symbol 结构体中的精度信息
	pricePrecision = symbolInfo.PricePrecision
	quantityPrecision = symbolInfo.QuantityPrecision

	// 从过滤器中获取 tick size、step size 和最小数量
	for _, filter := range symbolInfo.Filters {
		if filterType, ok := filter["filterType"].(string); ok {
			switch filterType {
			case "PRICE_FILTER":
				if tickSizeStr, ok := filter["tickSize"].(string); ok {
					tickSize, _ = strconv.ParseFloat(tickSizeStr, 64)
				}
			case "LOT_SIZE":
				if stepSizeStr, ok := filter["stepSize"].(string); ok {
					stepSize, _ = strconv.ParseFloat(stepSizeStr, 64)
				}
				if minQtyStr, ok := filter["minQty"].(string); ok {
					minQty, _ = strconv.ParseFloat(minQtyStr, 64)
				}
			}
		}
	}

	// 减少规则日志
	// log.Printf("交易对 %s 规则 - 价格精度: %d, 数量精度: %d, TickSize: %f, StepSize: %f, MinQty: %f",
	// 	strategy.Symbol, pricePrecision, quantityPrecision, tickSize, stepSize, minQty)

	// 获取深度数据以计算开仓价格
	depth, err := client.NewDepthService().
		Symbol(strategy.Symbol).
		Limit(20). // 增加深度层级以便更好地避免吃单
		Do(context.Background())
	if err != nil {
		log.Printf("获取深度失败: %v", err)
		updateStrategyStatus(m.cfg.DB, strategy, "cancelled", err.Error())
		return
	}

	// 计算开仓价格（优化挂单逻辑）
	var entryPrice float64
	if strategy.Side == "LONG" {
		// 做多时使用卖一价
		if len(depth.Asks) > 0 {
			askPrice, _ := strconv.ParseFloat(depth.Asks[0].Price, 64)

			// 如果设置了浮动，按浮动计算
			if strategy.EntryPriceFloat > 0 {
				entryPrice = askPrice * (1 - strategy.EntryPriceFloat/10000) // 万分比
			} else {
				// 没有设置浮动时，稍微低于卖一价挂单（避免立即吃单）
				// 默认低于卖一价1个tick
				entryPrice = askPrice - tickSize
			}

			// 确保价格不会高于卖一价（避免吃单）
			if entryPrice >= askPrice {
				entryPrice = askPrice - tickSize
			}
		}
	} else {
		// 做空时使用买一价
		if len(depth.Bids) > 0 {
			bidPrice, _ := strconv.ParseFloat(depth.Bids[0].Price, 64)

			// 如果设置了浮动，按浮动计算
			if strategy.EntryPriceFloat > 0 {
				entryPrice = bidPrice * (1 + strategy.EntryPriceFloat/10000) // 万分比
			} else {
				// 没有设置浮动时，稍微高于买一价挂单（避免立即吃单）
				// 默认高于买一价1个tick
				entryPrice = bidPrice + tickSize
			}

			// 确保价格不会低于买一价（避免吃单）
			if entryPrice <= bidPrice {
				entryPrice = bidPrice + tickSize
			}
		}
	}

	if entryPrice == 0 {
		log.Printf("无法获取开仓价格")
		updateStrategyStatus(m.cfg.DB, strategy, "cancelled", "无法获取价格")
		return
	}

	// 将价格调整为 tick size 的整数倍
	if tickSize > 0 {
		entryPrice = math.Round(entryPrice/tickSize) * tickSize
	}

	// 计算合约数量（使用本金×杠杆计算实际开仓价值）
	actualOrderValue := strategy.Quantity * float64(strategy.Leverage) // 本金×杠杆=实际开仓价值
	contractQuantity := actualOrderValue / entryPrice                  // 实际开仓价值÷价格=合约数量

	// 减少开仓计算日志
	// log.Printf("开仓计算 - 本金: %.2f USDT, 杠杆: %dx, 实际开仓价值: %.2f USDT, 合约数量: %.8f",
	// 	strategy.Quantity, strategy.Leverage, actualOrderValue, contractQuantity)

	// 将数量调整为 step size 的整数倍
	if stepSize > 0 {
		contractQuantity = math.Floor(contractQuantity/stepSize) * stepSize
	}

	// 检查数量是否小于最小数量
	if contractQuantity < minQty {
		// 如果小于最小数量，使用最小数量
		contractQuantity = minQty
		requiredValue := contractQuantity * entryPrice
		requiredMargin := requiredValue / float64(strategy.Leverage)
		log.Printf("警告：计算的合约数量小于最小数量 %.8f，将使用最小数量。需要本金 %.2f USDT",
			minQty, requiredMargin)
	}

	// 再次检查数量是否为0
	if contractQuantity <= 0 {
		errMsg := fmt.Sprintf("计算后的合约数量为0。本金: %.2f USDT, 杠杆: %dx, 价格: %.2f, 最小数量: %.8f",
			strategy.Quantity, strategy.Leverage, entryPrice, minQty)
		log.Printf(errMsg)
		updateStrategyStatus(m.cfg.DB, strategy, "cancelled", errMsg)
		return
	}

	// 格式化数量和价格，确保不超过允许的精度
	quantityFormat := fmt.Sprintf("%%.%df", quantityPrecision)
	priceFormat := fmt.Sprintf("%%.%df", pricePrecision)

	formattedQuantity := fmt.Sprintf(quantityFormat, contractQuantity)
	formattedPrice := fmt.Sprintf(priceFormat, entryPrice)

	// 保留关键的开仓参数日志
	log.Printf("开仓参数 - 策略ID: %d, 交易对: %s, 方向: %s, 数量: %s, 价格: %s",
		strategy.ID, strategy.Symbol, strategy.Side, formattedQuantity, formattedPrice)

	// 更新策略的实际开仓价格
	strategy.EntryPrice = entryPrice
	strategy.CalculateTakeProfitPrice()
	strategy.CalculateStopLossPrice()
	m.cfg.DB.Save(strategy)

	// 创建开仓订单
	side := futures.SideTypeBuy
	if strategy.Side == "SHORT" {
		side = futures.SideTypeSell
	}

	// 使用期货客户端创建订单
	orderService := client.NewCreateOrderService().
		Symbol(strategy.Symbol).
		Side(side).
		PositionSide(futures.PositionSideType(strategy.Side)).
		Type(futures.OrderTypeLimit).
		TimeInForce(futures.TimeInForceTypeGTC).
		Quantity(formattedQuantity).
		Price(formattedPrice)

	order, err := orderService.Do(context.Background())
	if err != nil {
		log.Printf("创建开仓订单失败: %v", err)
		updateStrategyStatus(m.cfg.DB, strategy, "cancelled", err.Error())
		return
	}

	// 保存订单记录
	dbOrder := models.FuturesOrder{
		UserID:       strategy.UserID,
		StrategyID:   strategy.ID,
		Symbol:       strategy.Symbol,
		Side:         string(side),
		PositionSide: strategy.Side,
		Type:         "LIMIT",
		Price:        entryPrice,
		Quantity:     contractQuantity,
		OrderID:      order.OrderID,
		Status:       string(order.Status),
		OrderPurpose: "entry",
	}

	if err := m.cfg.DB.Create(&dbOrder).Error; err != nil {
		log.Printf("保存订单记录失败: %v", err)
	}

	log.Printf("期货策略 %d 开仓订单创建成功: OrderID=%d", strategy.ID, order.OrderID)

	// 启动订单监控
	go monitorEntryOrder(m.cfg, strategy, order.OrderID)
}

// executeSlowIcebergStrategy 执行慢冰山策略
func (m *FuturesWebSocketManager) executeSlowIcebergStrategy(strategy *models.FuturesStrategy, client *futures.Client) {
	// 设置杠杆
	if err := setLeverage(client, strategy.Symbol, strategy.Leverage); err != nil {
		log.Printf("设置杠杆失败: %v", err)
		updateStrategyStatus(m.cfg.DB, strategy, "cancelled", err.Error())
		return
	}

	// 设置保证金模式（忽略已存在的错误）
	if err := setMarginType(client, strategy.Symbol, strategy.MarginType); err != nil {
		if !strings.Contains(err.Error(), "No need to change margin type") {
			log.Printf("设置保证金模式失败: %v", err)
		}
	}

	// 获取交易规则（精度信息）
	exchangeInfo, err := client.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		log.Printf("获取交易规则失败: %v", err)
		updateStrategyStatus(m.cfg.DB, strategy, "cancelled", err.Error())
		return
	}

	// 查找当前交易对的规则
	var symbolInfo *futures.Symbol
	for _, s := range exchangeInfo.Symbols {
		if s.Symbol == strategy.Symbol {
			symbolInfo = &s
			break
		}
	}

	if symbolInfo == nil {
		log.Printf("未找到交易对 %s 的规则", strategy.Symbol)
		updateStrategyStatus(m.cfg.DB, strategy, "cancelled", "未找到交易对规则")
		return
	}

	// 获取价格和数量精度
	var pricePrecision int
	var quantityPrecision int
	var tickSize float64
	var stepSize float64
	var minQty float64

	// 直接使用 Symbol 结构体中的精度信息
	pricePrecision = symbolInfo.PricePrecision
	quantityPrecision = symbolInfo.QuantityPrecision

	// 从过滤器中获取 tick size、step size 和最小数量
	for _, filter := range symbolInfo.Filters {
		if filterType, ok := filter["filterType"].(string); ok {
			switch filterType {
			case "PRICE_FILTER":
				if tickSizeStr, ok := filter["tickSize"].(string); ok {
					tickSize, _ = strconv.ParseFloat(tickSizeStr, 64)
				}
			case "LOT_SIZE":
				if stepSizeStr, ok := filter["stepSize"].(string); ok {
					stepSize, _ = strconv.ParseFloat(stepSizeStr, 64)
				}
				if minQtyStr, ok := filter["minQty"].(string); ok {
					minQty, _ = strconv.ParseFloat(minQtyStr, 64)
				}
			}
		}
	}

	// 获取当前市场深度
	depth, err := client.NewDepthService().
		Symbol(strategy.Symbol).
		Limit(20).
		Do(context.Background())
	if err != nil {
		log.Printf("获取深度失败: %v", err)
		updateStrategyStatus(m.cfg.DB, strategy, "cancelled", err.Error())
		return
	}

	// 解析冰山策略配置
	quantities := parseQuantities(strategy.IcebergQuantities)
	priceGaps := parsePriceGaps(strategy.IcebergPriceGaps, strategy.Side)

	if len(quantities) != len(priceGaps) {
		log.Printf("慢冰山策略配置错误：数量和价格间隔数量不匹配")
		updateStrategyStatus(m.cfg.DB, strategy, "cancelled", "配置错误")
		return
	}

	// 获取基准价格（第一层）
	var basePrice float64
	if strategy.Side == "LONG" {
		// 做多时使用卖一价作为基准
		if len(depth.Asks) > 0 {
			basePrice, _ = strconv.ParseFloat(depth.Asks[0].Price, 64)
		}
	} else {
		// 做空时使用买一价作为基准
		if len(depth.Bids) > 0 {
			basePrice, _ = strconv.ParseFloat(depth.Bids[0].Price, 64)
		}
	}

	if basePrice == 0 {
		log.Printf("无法获取基准价格")
		updateStrategyStatus(m.cfg.DB, strategy, "cancelled", "无法获取价格")
		return
	}

	// 计算第一层的价格
	firstLayerPrice := basePrice * (1 + priceGaps[0]/10000)

	// 应用开仓价格浮动（万分比）并避免吃单
	if strategy.Side == "LONG" {
		if strategy.EntryPriceFloat > 0 {
			firstLayerPrice = firstLayerPrice * (1 - strategy.EntryPriceFloat/10000)
		} else {
			// 默认低于基准价1个tick避免吃单
			firstLayerPrice = firstLayerPrice - tickSize
		}
		// 确保不会高于卖一价
		if firstLayerPrice >= basePrice {
			firstLayerPrice = basePrice - tickSize
		}
	} else {
		if strategy.EntryPriceFloat > 0 {
			firstLayerPrice = firstLayerPrice * (1 + strategy.EntryPriceFloat/10000)
		} else {
			// 默认高于基准价1个tick避免吃单
			firstLayerPrice = firstLayerPrice + tickSize
		}
		// 确保不会低于买一价
		if firstLayerPrice <= basePrice {
			firstLayerPrice = basePrice + tickSize
		}
	}

	// 将价格调整为 tick size 的整数倍
	if tickSize > 0 {
		firstLayerPrice = math.Round(firstLayerPrice/tickSize) * tickSize
	}

	// 计算第一层的价值和数量
	totalOrderValue := strategy.Quantity * float64(strategy.Leverage)
	firstLayerValue := totalOrderValue * quantities[0]
	firstLayerQuantity := firstLayerValue / firstLayerPrice

	// 将数量调整为 step size 的整数倍
	if stepSize > 0 {
		firstLayerQuantity = math.Floor(firstLayerQuantity/stepSize) * stepSize
	}

	// 检查第一层数量是否满足最小要求
	if firstLayerQuantity < minQty {
		log.Printf("慢冰山策略第一层数量 %.8f 小于最小数量 %.8f", firstLayerQuantity, minQty)
		updateStrategyStatus(m.cfg.DB, strategy, "cancelled", "第一层数量太小")
		return
	}

	// 格式化数量和价格的格式字符串
	quantityFormat := fmt.Sprintf("%%.%df", quantityPrecision)
	priceFormat := fmt.Sprintf("%%.%df", pricePrecision)

	// 格式化第一层数量和价格
	formattedQuantity := fmt.Sprintf(quantityFormat, firstLayerQuantity)
	formattedPrice := fmt.Sprintf(priceFormat, firstLayerPrice)

	// 计算第一层使用的本金
	firstLayerMargin := firstLayerValue / float64(strategy.Leverage)

	log.Printf("慢冰山策略 %d 第1层 - 价格: %s, 数量: %s (本金: %.2f USDT)",
		strategy.ID, formattedPrice, formattedQuantity, firstLayerMargin)

	// 创建开仓方向
	side := futures.SideTypeBuy
	if strategy.Side == "SHORT" {
		side = futures.SideTypeSell
	}

	// 创建第一层限价订单
	orderService := client.NewCreateOrderService().
		Symbol(strategy.Symbol).
		Side(side).
		PositionSide(futures.PositionSideType(strategy.Side)).
		Type(futures.OrderTypeLimit).
		TimeInForce(futures.TimeInForceTypeGTC).
		Quantity(formattedQuantity).
		Price(formattedPrice)

	order, err := orderService.Do(context.Background())
	if err != nil {
		log.Printf("创建慢冰山第1层订单失败: %v", err)
		updateStrategyStatus(m.cfg.DB, strategy, "cancelled", err.Error())
		return
	}

	// 保存订单记录
	dbOrder := models.FuturesOrder{
		UserID:       strategy.UserID,
		StrategyID:   strategy.ID,
		Symbol:       strategy.Symbol,
		Side:         string(side),
		PositionSide: strategy.Side,
		Type:         "LIMIT",
		Price:        firstLayerPrice,
		Quantity:     firstLayerQuantity,
		OrderID:      order.OrderID,
		Status:       string(order.Status),
		OrderPurpose: "entry",
	}

	if err := m.cfg.DB.Create(&dbOrder).Error; err != nil {
		log.Printf("保存订单记录失败: %v", err)
	}

	log.Printf("慢冰山策略 %d 第1层订单创建成功: OrderID=%d", strategy.ID, order.OrderID)

	// 启动慢冰山订单监控（传递必要的参数）
	go monitorSlowIcebergOrders(m.cfg, strategy, order.OrderID, 0, quantities, priceGaps,
		pricePrecision, quantityPrecision, tickSize, stepSize, minQty)
}

// executeIcebergStrategy 执行冰山策略
func (m *FuturesWebSocketManager) executeIcebergStrategy(strategy *models.FuturesStrategy, client *futures.Client) {
	// 设置杠杆
	if err := setLeverage(client, strategy.Symbol, strategy.Leverage); err != nil {
		log.Printf("设置杠杆失败: %v", err)
		updateStrategyStatus(m.cfg.DB, strategy, "cancelled", err.Error())
		return
	}

	// 设置保证金模式（忽略已存在的错误）
	if err := setMarginType(client, strategy.Symbol, strategy.MarginType); err != nil {
		if !strings.Contains(err.Error(), "No need to change margin type") {
			log.Printf("设置保证金模式失败: %v", err)
		}
	}

	// 获取交易规则（精度信息）
	exchangeInfo, err := client.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		log.Printf("获取交易规则失败: %v", err)
		updateStrategyStatus(m.cfg.DB, strategy, "cancelled", err.Error())
		return
	}

	// 查找当前交易对的规则
	var symbolInfo *futures.Symbol
	for _, s := range exchangeInfo.Symbols {
		if s.Symbol == strategy.Symbol {
			symbolInfo = &s
			break
		}
	}

	if symbolInfo == nil {
		log.Printf("未找到交易对 %s 的规则", strategy.Symbol)
		updateStrategyStatus(m.cfg.DB, strategy, "cancelled", "未找到交易对规则")
		return
	}

	// 获取价格和数量精度
	var pricePrecision int
	var quantityPrecision int
	var tickSize float64
	var stepSize float64
	var minQty float64

	// 直接使用 Symbol 结构体中的精度信息
	pricePrecision = symbolInfo.PricePrecision
	quantityPrecision = symbolInfo.QuantityPrecision

	// 从过滤器中获取 tick size、step size 和最小数量
	for _, filter := range symbolInfo.Filters {
		if filterType, ok := filter["filterType"].(string); ok {
			switch filterType {
			case "PRICE_FILTER":
				if tickSizeStr, ok := filter["tickSize"].(string); ok {
					tickSize, _ = strconv.ParseFloat(tickSizeStr, 64)
				}
			case "LOT_SIZE":
				if stepSizeStr, ok := filter["stepSize"].(string); ok {
					stepSize, _ = strconv.ParseFloat(stepSizeStr, 64)
				}
				if minQtyStr, ok := filter["minQty"].(string); ok {
					minQty, _ = strconv.ParseFloat(minQtyStr, 64)
				}
			}
		}
	}

	// 获取当前市场深度
	depth, err := client.NewDepthService().
		Symbol(strategy.Symbol).
		Limit(20).
		Do(context.Background())
	if err != nil {
		log.Printf("获取深度失败: %v", err)
		updateStrategyStatus(m.cfg.DB, strategy, "cancelled", err.Error())
		return
	}

	// 解析冰山策略配置
	quantities := parseQuantities(strategy.IcebergQuantities)
	priceGaps := parsePriceGaps(strategy.IcebergPriceGaps, strategy.Side)

	if len(quantities) != len(priceGaps) {
		log.Printf("冰山策略配置错误：数量和价格间隔数量不匹配")
		updateStrategyStatus(m.cfg.DB, strategy, "cancelled", "配置错误")
		return
	}

	// 获取基准价格
	var basePrice float64
	if strategy.Side == "LONG" {
		// 做多时使用卖一价作为基准
		if len(depth.Asks) > 0 {
			basePrice, _ = strconv.ParseFloat(depth.Asks[0].Price, 64)
		}
	} else {
		// 做空时使用买一价作为基准
		if len(depth.Bids) > 0 {
			basePrice, _ = strconv.ParseFloat(depth.Bids[0].Price, 64)
		}
	}

	if basePrice == 0 {
		log.Printf("无法获取基准价格")
		updateStrategyStatus(m.cfg.DB, strategy, "cancelled", "无法获取价格")
		return
	}

	// 创建冰山订单
	side := futures.SideTypeBuy
	if strategy.Side == "SHORT" {
		side = futures.SideTypeSell
	}

	var successfulOrders []int64
	var totalExecutedQuantity float64
	var weightedPriceSum float64

	// 计算实际开仓价值（本金×杠杆）
	totalOrderValue := strategy.Quantity * float64(strategy.Leverage)

	// 格式化数量和价格的格式字符串
	quantityFormat := fmt.Sprintf("%%.%df", quantityPrecision)
	priceFormat := fmt.Sprintf("%%.%df", pricePrecision)

	// 先计算所有层的价格和检查哪些层的数量太小
	type layerInfo struct {
		price    float64
		quantity float64
		value    float64 // 该层的价值（USDT）
		skip     bool
	}

	layers := make([]layerInfo, len(quantities))
	skippedValue := 0.0

	// 第一遍：计算每层信息并标记需要跳过的层
	for i := 0; i < len(quantities); i++ {
		// 计算每层的价格
		layerPrice := basePrice * (1 + priceGaps[i]/10000)

		// 应用开仓价格浮动（万分比）并避免吃单
		if i == 0 { // 只对第一层应用浮动和避免吃单逻辑
			if strategy.Side == "LONG" {
				if strategy.EntryPriceFloat > 0 {
					layerPrice = layerPrice * (1 - strategy.EntryPriceFloat/10000)
				} else if priceGaps[i] == 0 {
					// 第一层价格间隔为0且没有设置浮动时，避免吃单
					layerPrice = layerPrice - tickSize
				}
				// 确保不会高于卖一价
				if layerPrice >= basePrice {
					layerPrice = basePrice - tickSize
				}
			} else {
				if strategy.EntryPriceFloat > 0 {
					layerPrice = layerPrice * (1 + strategy.EntryPriceFloat/10000)
				} else if priceGaps[i] == 0 {
					// 第一层价格间隔为0且没有设置浮动时，避免吃单
					layerPrice = layerPrice + tickSize
				}
				// 确保不会低于买一价
				if layerPrice <= basePrice {
					layerPrice = basePrice + tickSize
				}
			}
		}

		// 将价格调整为 tick size 的整数倍
		if tickSize > 0 {
			layerPrice = math.Round(layerPrice/tickSize) * tickSize
		}

		// 计算每层的价值（按比例分配总价值）
		layerValue := totalOrderValue * quantities[i]

		// 转换为合约数量
		layerContractQuantity := layerValue / layerPrice

		// 将数量调整为 step size 的整数倍
		if stepSize > 0 {
			layerContractQuantity = math.Floor(layerContractQuantity/stepSize) * stepSize
		}

		layers[i] = layerInfo{
			price:    layerPrice,
			quantity: layerContractQuantity,
			value:    layerValue,
			skip:     layerContractQuantity < minQty,
		}

		if layers[i].skip {
			skippedValue += layerValue
			// 减少跳过层的日志
			// log.Printf("冰山第%d层将被跳过 - 价值: %.2f USDT, 数量: %.8f < 最小数量: %.8f",
			// 	i+1, layerValue, layerContractQuantity, minQty)
		}
	}

	// 如果有被跳过的金额，重新分配到有效层
	if skippedValue > 0 {
		validLayers := 0
		for i := 0; i < len(layers); i++ {
			if !layers[i].skip {
				validLayers++
			}
		}

		if validLayers > 0 {
			// 将跳过的价值平均分配到有效层
			additionalValuePerLayer := skippedValue / float64(validLayers)
			// 减少重新分配日志
			// log.Printf("将 %.2f USDT 价值重新分配到 %d 个有效层，每层增加 %.2f USDT",
			// 	skippedValue, validLayers, additionalValuePerLayer)

			for i := 0; i < len(layers); i++ {
				if !layers[i].skip {
					// 增加价值
					newValue := layers[i].value + additionalValuePerLayer
					// 重新计算合约数量
					newContractQty := newValue / layers[i].price
					// 调整为 step size 的整数倍
					if stepSize > 0 {
						newContractQty = math.Floor(newContractQty/stepSize) * stepSize
					}
					layers[i].quantity = newContractQty
					layers[i].value = newValue
				}
			}
		}
	}

	// 第二遍：创建订单
	for i := 0; i < len(layers); i++ {
		if layers[i].skip {
			continue
		}

		// 格式化数量和价格
		formattedQuantity := fmt.Sprintf(quantityFormat, layers[i].quantity)
		formattedPrice := fmt.Sprintf(priceFormat, layers[i].price)

		// 计算该层使用的本金
		layerMargin := layers[i].value / float64(strategy.Leverage)

		// 保留关键层信息日志
		log.Printf("冰山策略 %d 第%d层 - 价格: %s, 数量: %s (本金: %.2f USDT)",
			strategy.ID, i+1, formattedPrice, formattedQuantity, layerMargin)

		// 创建限价订单
		orderService := client.NewCreateOrderService().
			Symbol(strategy.Symbol).
			Side(side).
			PositionSide(futures.PositionSideType(strategy.Side)).
			Type(futures.OrderTypeLimit).
			TimeInForce(futures.TimeInForceTypeGTC).
			Quantity(formattedQuantity).
			Price(formattedPrice)

		order, err := orderService.Do(context.Background())
		if err != nil {
			log.Printf("创建第%d层订单失败: %v", i+1, err)
			// 如果是第一个有效订单就失败，取消整个策略
			if len(successfulOrders) == 0 {
				updateStrategyStatus(m.cfg.DB, strategy, "cancelled", err.Error())
				return
			}
			// 否则继续尝试下一层
			continue
		}

		successfulOrders = append(successfulOrders, order.OrderID)

		// 累计加权价格用于计算平均开仓价
		weightedPriceSum += layers[i].price * layers[i].quantity
		totalExecutedQuantity += layers[i].quantity

		// 保存订单记录
		dbOrder := models.FuturesOrder{
			UserID:       strategy.UserID,
			StrategyID:   strategy.ID,
			Symbol:       strategy.Symbol,
			Side:         string(side),
			PositionSide: strategy.Side,
			Type:         "LIMIT",
			Price:        layers[i].price,
			Quantity:     layers[i].quantity,
			OrderID:      order.OrderID,
			Status:       string(order.Status),
			OrderPurpose: "entry",
		}

		if err := m.cfg.DB.Create(&dbOrder).Error; err != nil {
			log.Printf("保存订单记录失败: %v", err)
		}
	}

	if len(successfulOrders) == 0 {
		updateStrategyStatus(m.cfg.DB, strategy, "cancelled", "所有订单创建失败：金额太小")
		return
	}

	// 计算并更新预估的平均开仓价格
	if totalExecutedQuantity > 0 {
		avgEntryPrice := weightedPriceSum / totalExecutedQuantity
		strategy.EntryPrice = avgEntryPrice
		strategy.CalculateTakeProfitPrice()
		strategy.CalculateStopLossPrice()
		m.cfg.DB.Save(strategy)
	}

	log.Printf("期货冰山策略 %d 开仓订单创建完成，共%d层成功", strategy.ID, len(successfulOrders))

	// 启动订单监控
	go monitorIcebergOrders(m.cfg, strategy, successfulOrders)
}

// monitorSlowIcebergOrders 监控慢冰山订单
func monitorSlowIcebergOrders(cfg *config.Config, strategy *models.FuturesStrategy,
	currentOrderID int64, currentLayer int, quantities []float64, priceGaps []float64,
	pricePrecision int, quantityPrecision int, tickSize float64, stepSize float64, minQty float64) {

	// 获取用户信息
	var user models.User
	if err := cfg.DB.First(&user, strategy.UserID).Error; err != nil {
		return
	}

	// 解密API密钥
	apiKey, err := user.GetDecryptedAPIKey()
	if err != nil {
		return
	}
	secretKey, err := user.GetDecryptedSecretKey()
	if err != nil {
		return
	}

	client := binance.NewFuturesClient(apiKey, secretKey)

	// 定期检查订单状态
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(10 * time.Minute) // 10分钟超时

	// 用于追踪所有层的成交情况
	var totalFilledQuantity float64
	var weightedPriceSum float64
	var allOrderIDs []int64

	// 将当前订单加入列表
	allOrderIDs = append(allOrderIDs, currentOrderID)

	for {
		select {
		case <-ticker.C:
			order, err := client.NewGetOrderService().
				Symbol(strategy.Symbol).
				OrderID(currentOrderID).
				Do(context.Background())

			if err != nil {
				log.Printf("查询慢冰山订单状态失败: %v", err)
				continue
			}

			// 更新订单状态
			execQty, _ := strconv.ParseFloat(order.ExecutedQuantity, 64)
			avgPrice, _ := strconv.ParseFloat(order.AvgPrice, 64)

			cfg.DB.Model(&models.FuturesOrder{}).
				Where("order_id = ?", currentOrderID).
				Updates(map[string]interface{}{
					"status":       string(order.Status),
					"executed_qty": execQty,
					"avg_price":    avgPrice,
				})

			// 检查订单是否成交
			if order.Status == futures.OrderStatusTypeFilled {
				log.Printf("慢冰山第%d层订单成交: 策略ID=%d, OrderID=%d, AvgPrice=%.8f",
					currentLayer+1, strategy.ID, currentOrderID, avgPrice)

				// 更新成交统计
				totalFilledQuantity += execQty
				weightedPriceSum += avgPrice * execQty

				// 检查是否还有下一层
				if currentLayer+1 < len(quantities) {
					// 获取最新的市场深度
					depth, depthErr := client.NewDepthService().
						Symbol(strategy.Symbol).
						Limit(20).
						Do(context.Background())

					if depthErr != nil {
						log.Printf("获取深度失败: %v", depthErr)
						break
					}

					// 根据最新的买卖1价计算下一层价格
					var basePrice float64
					if strategy.Side == "LONG" {
						if len(depth.Asks) > 0 {
							basePrice, _ = strconv.ParseFloat(depth.Asks[0].Price, 64)
						}
					} else {
						if len(depth.Bids) > 0 {
							basePrice, _ = strconv.ParseFloat(depth.Bids[0].Price, 64)
						}
					}

					if basePrice == 0 {
						log.Printf("无法获取下一层基准价格")
						break
					}

					// 计算下一层的价格（基于新的买卖1价）
					nextLayerPrice := basePrice * (1 + priceGaps[currentLayer+1]/10000)

					// 应用开仓价格浮动并避免吃单
					if strategy.Side == "LONG" {
						if strategy.EntryPriceFloat > 0 {
							nextLayerPrice = nextLayerPrice * (1 - strategy.EntryPriceFloat/10000)
						} else {
							// 避免吃单
							nextLayerPrice = nextLayerPrice - tickSize
						}
						// 确保不会高于卖一价
						if nextLayerPrice >= basePrice {
							nextLayerPrice = basePrice - tickSize
						}
					} else {
						if strategy.EntryPriceFloat > 0 {
							nextLayerPrice = nextLayerPrice * (1 + strategy.EntryPriceFloat/10000)
						} else {
							// 避免吃单
							nextLayerPrice = nextLayerPrice + tickSize
						}
						// 确保不会低于买一价
						if nextLayerPrice <= basePrice {
							nextLayerPrice = basePrice + tickSize
						}
					}

					// 将价格调整为 tick size 的整数倍
					if tickSize > 0 {
						nextLayerPrice = math.Round(nextLayerPrice/tickSize) * tickSize
					}

					// 计算下一层的数量
					totalOrderValue := strategy.Quantity * float64(strategy.Leverage)
					nextLayerValue := totalOrderValue * quantities[currentLayer+1]
					nextLayerQuantity := nextLayerValue / nextLayerPrice

					// 将数量调整为 step size 的整数倍
					if stepSize > 0 {
						nextLayerQuantity = math.Floor(nextLayerQuantity/stepSize) * stepSize
					}

					// 检查数量是否满足最小要求
					if nextLayerQuantity < minQty {
						// 减少跳过层的日志
						// log.Printf("慢冰山第%d层数量太小，跳过", currentLayer+2)
						// 递归调用处理下一层
						if currentLayer+2 < len(quantities) {
							go monitorSlowIcebergOrders(cfg, strategy, currentOrderID, currentLayer+1,
								quantities, priceGaps, pricePrecision, quantityPrecision,
								tickSize, stepSize, minQty)
						}
						break
					}

					// 格式化数量和价格
					quantityFormat := fmt.Sprintf("%%.%df", quantityPrecision)
					priceFormat := fmt.Sprintf("%%.%df", pricePrecision)
					formattedQuantity := fmt.Sprintf(quantityFormat, nextLayerQuantity)
					formattedPrice := fmt.Sprintf(priceFormat, nextLayerPrice)

					log.Printf("慢冰山策略 %d 第%d层 - 价格: %s (基于新的%s价), 数量: %s",
						strategy.ID, currentLayer+2, formattedPrice,
						map[bool]string{true: "卖1", false: "买1"}[strategy.Side == "LONG"],
						formattedQuantity)

					// 创建下一层订单
					side := futures.SideTypeBuy
					if strategy.Side == "SHORT" {
						side = futures.SideTypeSell
					}

					nextOrder, nextErr := client.NewCreateOrderService().
						Symbol(strategy.Symbol).
						Side(side).
						PositionSide(futures.PositionSideType(strategy.Side)).
						Type(futures.OrderTypeLimit).
						TimeInForce(futures.TimeInForceTypeGTC).
						Quantity(formattedQuantity).
						Price(formattedPrice).
						Do(context.Background())

					if nextErr != nil {
						log.Printf("创建慢冰山第%d层订单失败: %v", currentLayer+2, nextErr)
						break
					}

					// 保存订单记录
					dbOrder := models.FuturesOrder{
						UserID:       strategy.UserID,
						StrategyID:   strategy.ID,
						Symbol:       strategy.Symbol,
						Side:         string(side),
						PositionSide: strategy.Side,
						Type:         "LIMIT",
						Price:        nextLayerPrice,
						Quantity:     nextLayerQuantity,
						OrderID:      nextOrder.OrderID,
						Status:       string(nextOrder.Status),
						OrderPurpose: "entry",
					}

					if err := cfg.DB.Create(&dbOrder).Error; err != nil {
						log.Printf("保存订单记录失败: %v", err)
					}

					// 将新订单加入列表
					allOrderIDs = append(allOrderIDs, nextOrder.OrderID)

					// 递归监控下一层
					go monitorSlowIcebergOrders(cfg, strategy, nextOrder.OrderID, currentLayer+1,
						quantities, priceGaps, pricePrecision, quantityPrecision,
						tickSize, stepSize, minQty)
				} else {
					// 所有层都已完成，创建持仓和止盈止损订单
					avgEntryPrice := weightedPriceSum / totalFilledQuantity

					// 创建持仓记录
					position := models.FuturesPosition{
						UserID:       strategy.UserID,
						StrategyID:   strategy.ID,
						Symbol:       strategy.Symbol,
						PositionSide: strategy.Side,
						EntryPrice:   avgEntryPrice,
						Quantity:     totalFilledQuantity,
						Leverage:     strategy.Leverage,
						MarginType:   strategy.MarginType,
						Status:       "open",
						OpenedAt:     time.Now(),
					}

					cfg.DB.Create(&position)

					// 更新策略状态和实际开仓价格
					strategy.Status = "position_opened"
					strategy.CurrentPositionId = allOrderIDs[0] // 使用第一个订单ID作为标识
					strategy.EntryPrice = avgEntryPrice
					strategy.CalculateTakeProfitPrice()
					strategy.CalculateStopLossPrice()
					cfg.DB.Save(strategy)

					// 创建止盈订单
					createTakeProfitOrder(cfg, client, strategy, totalFilledQuantity)

					// 如果设置了止损，创建止损订单
					if strategy.StopLossRate > 0 {
						createStopLossOrder(cfg, client, strategy, totalFilledQuantity)
					}

					log.Printf("慢冰山策略 %d 所有层完成，平均开仓价: %.8f", strategy.ID, avgEntryPrice)
				}

				return

			} else if order.Status == futures.OrderStatusTypeCanceled ||
				order.Status == futures.OrderStatusTypeExpired ||
				order.Status == futures.OrderStatusTypeRejected {
				log.Printf("慢冰山第%d层订单失败: OrderID=%d, Status=%s",
					currentLayer+1, currentOrderID, order.Status)

				// 如果是第一层就失败，取消策略
				if currentLayer == 0 && totalFilledQuantity == 0 {
					updateStrategyStatus(cfg.DB, strategy, "cancelled", string(order.Status))
				} else if totalFilledQuantity > 0 {
					// 如果已有部分成交，创建持仓
					avgEntryPrice := weightedPriceSum / totalFilledQuantity

					position := models.FuturesPosition{
						UserID:       strategy.UserID,
						StrategyID:   strategy.ID,
						Symbol:       strategy.Symbol,
						PositionSide: strategy.Side,
						EntryPrice:   avgEntryPrice,
						Quantity:     totalFilledQuantity,
						Leverage:     strategy.Leverage,
						MarginType:   strategy.MarginType,
						Status:       "open",
						OpenedAt:     time.Now(),
					}

					cfg.DB.Create(&position)

					strategy.Status = "position_opened"
					strategy.CurrentPositionId = allOrderIDs[0]
					strategy.EntryPrice = avgEntryPrice
					strategy.CalculateTakeProfitPrice()
					strategy.CalculateStopLossPrice()
					cfg.DB.Save(strategy)

					createTakeProfitOrder(cfg, client, strategy, totalFilledQuantity)
					if strategy.StopLossRate > 0 {
						createStopLossOrder(cfg, client, strategy, totalFilledQuantity)
					}
				}

				return
			}

		case <-timeout:
			// 超时取消当前订单
			log.Printf("慢冰山第%d层订单超时，取消订单: OrderID=%d", currentLayer+1, currentOrderID)
			_, cancelErr := client.NewCancelOrderService().
				Symbol(strategy.Symbol).
				OrderID(currentOrderID).
				Do(context.Background())
			if cancelErr != nil {
				log.Printf("取消订单失败: %v", cancelErr)
			}

			// 如果有部分成交，仍然创建持仓
			if totalFilledQuantity > 0 {
				avgEntryPrice := weightedPriceSum / totalFilledQuantity

				position := models.FuturesPosition{
					UserID:       strategy.UserID,
					StrategyID:   strategy.ID,
					Symbol:       strategy.Symbol,
					PositionSide: strategy.Side,
					EntryPrice:   avgEntryPrice,
					Quantity:     totalFilledQuantity,
					Leverage:     strategy.Leverage,
					MarginType:   strategy.MarginType,
					Status:       "open",
					OpenedAt:     time.Now(),
				}

				cfg.DB.Create(&position)

				strategy.Status = "position_opened"
				strategy.CurrentPositionId = allOrderIDs[0]
				strategy.EntryPrice = avgEntryPrice
				strategy.CalculateTakeProfitPrice()
				strategy.CalculateStopLossPrice()
				cfg.DB.Save(strategy)

				createTakeProfitOrder(cfg, client, strategy, totalFilledQuantity)
				if strategy.StopLossRate > 0 {
					createStopLossOrder(cfg, client, strategy, totalFilledQuantity)
				}
			} else {
				updateStrategyStatus(cfg.DB, strategy, "cancelled", "timeout")
			}

			return
		}
	}
}

// monitorIcebergOrders 监控冰山订单
func monitorIcebergOrders(cfg *config.Config, strategy *models.FuturesStrategy, orderIDs []int64) {
	// 获取用户信息
	var user models.User
	if err := cfg.DB.First(&user, strategy.UserID).Error; err != nil {
		return
	}

	// 解密API密钥
	apiKey, err := user.GetDecryptedAPIKey()
	if err != nil {
		return
	}
	secretKey, err := user.GetDecryptedSecretKey()
	if err != nil {
		return
	}

	client := binance.NewFuturesClient(apiKey, secretKey)

	// 定期检查订单状态
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(10 * time.Minute) // 10分钟超时

	filledOrders := make(map[int64]bool)
	var totalFilledQuantity float64
	var weightedPriceSum float64

	for {
		select {
		case <-ticker.C:
			allFilled := true

			for _, orderID := range orderIDs {
				if filledOrders[orderID] {
					continue // 已成交的订单跳过
				}

				order, err := client.NewGetOrderService().
					Symbol(strategy.Symbol).
					OrderID(orderID).
					Do(context.Background())

				if err != nil {
					log.Printf("查询订单状态失败: %v", err)
					continue
				}

				// 更新订单状态
				execQty, _ := strconv.ParseFloat(order.ExecutedQuantity, 64)
				avgPrice, _ := strconv.ParseFloat(order.AvgPrice, 64)

				cfg.DB.Model(&models.FuturesOrder{}).
					Where("order_id = ?", orderID).
					Updates(map[string]interface{}{
						"status":       string(order.Status),
						"executed_qty": execQty,
						"avg_price":    avgPrice,
					})

				// 检查订单是否成交
				if order.Status == futures.OrderStatusTypeFilled {
					filledOrders[orderID] = true
					totalFilledQuantity += execQty
					weightedPriceSum += avgPrice * execQty
					log.Printf("冰山订单成交: 策略ID=%d, OrderID=%d, AvgPrice=%.8f", strategy.ID, orderID, avgPrice)
				} else if order.Status == futures.OrderStatusTypeCanceled ||
					order.Status == futures.OrderStatusTypeExpired ||
					order.Status == futures.OrderStatusTypeRejected {
					filledOrders[orderID] = true // 标记为已处理
					log.Printf("冰山订单失败: OrderID=%d, Status=%s", orderID, order.Status)
				} else {
					allFilled = false
				}
			}

			// 如果有订单成交且所有订单都已处理，创建持仓和止盈止损订单
			if totalFilledQuantity > 0 && allFilled {
				avgEntryPrice := weightedPriceSum / totalFilledQuantity

				// 创建持仓记录
				position := models.FuturesPosition{
					UserID:       strategy.UserID,
					StrategyID:   strategy.ID,
					Symbol:       strategy.Symbol,
					PositionSide: strategy.Side,
					EntryPrice:   avgEntryPrice,
					Quantity:     totalFilledQuantity,
					Leverage:     strategy.Leverage,
					MarginType:   strategy.MarginType,
					Status:       "open",
					OpenedAt:     time.Now(),
				}

				cfg.DB.Create(&position)

				// 更新策略状态和实际开仓价格
				strategy.Status = "position_opened"
				strategy.CurrentPositionId = orderIDs[0] // 使用第一个订单ID作为标识
				strategy.EntryPrice = avgEntryPrice
				strategy.CalculateTakeProfitPrice()
				strategy.CalculateStopLossPrice()
				cfg.DB.Save(strategy)

				// 创建止盈订单
				if strategy.StrategyType == "iceberg" {
					// 冰山策略的止盈也使用冰山方式
					createIcebergTakeProfitOrders(cfg, client, strategy, totalFilledQuantity)
				} else {
					createTakeProfitOrder(cfg, client, strategy, totalFilledQuantity)
				}

				// 如果设置了止损，创建止损订单
				if strategy.StopLossRate > 0 {
					createStopLossOrder(cfg, client, strategy, totalFilledQuantity)
				}

				return
			}

		case <-timeout:
			// 超时取消未成交订单
			log.Printf("冰山订单超时，取消未成交订单")
			for _, orderID := range orderIDs {
				if !filledOrders[orderID] {
					_, cancelErr := client.NewCancelOrderService().
						Symbol(strategy.Symbol).
						OrderID(orderID).
						Do(context.Background())
					if cancelErr != nil {
						log.Printf("取消订单失败: %v", cancelErr)
					}
				}
			}

			// 如果有部分成交，仍然创建持仓
			if totalFilledQuantity > 0 {
				avgEntryPrice := weightedPriceSum / totalFilledQuantity

				// 创建持仓记录
				position := models.FuturesPosition{
					UserID:       strategy.UserID,
					StrategyID:   strategy.ID,
					Symbol:       strategy.Symbol,
					PositionSide: strategy.Side,
					EntryPrice:   avgEntryPrice,
					Quantity:     totalFilledQuantity,
					Leverage:     strategy.Leverage,
					MarginType:   strategy.MarginType,
					Status:       "open",
					OpenedAt:     time.Now(),
				}

				cfg.DB.Create(&position)

				// 更新策略状态
				strategy.Status = "position_opened"
				strategy.CurrentPositionId = orderIDs[0]
				strategy.EntryPrice = avgEntryPrice
				strategy.CalculateTakeProfitPrice()
				strategy.CalculateStopLossPrice()
				cfg.DB.Save(strategy)

				// 创建止盈止损订单
				if strategy.StrategyType == "iceberg" {
					createIcebergTakeProfitOrders(cfg, client, strategy, totalFilledQuantity)
				} else {
					createTakeProfitOrder(cfg, client, strategy, totalFilledQuantity)
				}

				if strategy.StopLossRate > 0 {
					createStopLossOrder(cfg, client, strategy, totalFilledQuantity)
				}
			} else {
				updateStrategyStatus(cfg.DB, strategy, "cancelled", "timeout")
			}

			return
		}
	}
}

// createIcebergTakeProfitOrders 创建冰山止盈订单
func createIcebergTakeProfitOrders(cfg *config.Config, client *futures.Client,
	strategy *models.FuturesStrategy, totalQuantity float64) {

	// 解析冰山配置
	quantities := parseQuantities(strategy.IcebergQuantities)

	// 确定止盈方向
	side := futures.SideTypeSell
	if strategy.Side == "SHORT" {
		side = futures.SideTypeBuy
	}

	// 计算基准止盈价格
	baseTakeProfitPrice := strategy.TakeProfitPrice

	// 为止盈创建反向的价格间隔
	// 做多时：止盈价格递增（更高的价格）
	// 做空时：止盈价格递减（更低的价格）
	takeProfitGaps := make([]float64, len(quantities))
	if strategy.Side == "LONG" {
		// 做多止盈：价格递增
		takeProfitGaps[0] = 0
		for i := 1; i < len(takeProfitGaps); i++ {
			takeProfitGaps[i] = float64(i) * 20 // 每层增加20万分比（0.2%）
		}
	} else {
		// 做空止盈：价格递减
		takeProfitGaps[0] = 0
		for i := 1; i < len(takeProfitGaps); i++ {
			takeProfitGaps[i] = float64(i) * -20 // 每层减少20万分比（0.2%）
		}
	}

	successCount := 0
	for i := 0; i < len(quantities); i++ {
		// 计算每层的止盈价格（万分比）
		layerPrice := baseTakeProfitPrice * (1 + takeProfitGaps[i]/10000)
		layerQuantity := totalQuantity * quantities[i]

		order, err := client.NewCreateOrderService().
			Symbol(strategy.Symbol).
			Side(side).
			PositionSide(futures.PositionSideType(strategy.Side)).
			Type(futures.OrderTypeLimit).
			TimeInForce(futures.TimeInForceTypeGTC).
			Quantity(fmt.Sprintf("%.8f", layerQuantity)).
			Price(fmt.Sprintf("%.8f", layerPrice)).
			Do(context.Background())

		if err != nil {
			log.Printf("创建第%d层止盈订单失败: %v", i+1, err)
			continue
		}

		// 保存订单记录
		dbOrder := models.FuturesOrder{
			UserID:       strategy.UserID,
			StrategyID:   strategy.ID,
			Symbol:       strategy.Symbol,
			Side:         string(side),
			PositionSide: strategy.Side,
			Type:         "LIMIT",
			Price:        layerPrice,
			Quantity:     layerQuantity,
			OrderID:      order.OrderID,
			Status:       string(order.Status),
			OrderPurpose: "take_profit",
		}

		if err := cfg.DB.Create(&dbOrder).Error; err != nil {
			log.Printf("保存止盈订单失败: %v", err)
		}

		successCount++
		// 减少止盈订单创建日志
		// log.Printf("冰山止盈第%d层订单创建成功: OrderID=%d, Price=%.8f, Quantity=%.8f",
		// 	i+1, order.OrderID, layerPrice, layerQuantity)
	}

	if successCount > 0 {
		log.Printf("冰山止盈订单创建完成，共%d层", successCount)
	}
}

// 辅助函数：解析数量配置
func parseQuantities(quantitiesStr string) []float64 {
	if quantitiesStr == "" {
		return []float64{0.35, 0.25, 0.2, 0.1, 0.1} // 默认值
	}

	var quantities []float64
	for _, q := range strings.Split(quantitiesStr, ",") {
		if val, err := strconv.ParseFloat(strings.TrimSpace(q), 64); err == nil {
			quantities = append(quantities, val)
		}
	}
	return quantities
}

// 辅助函数：解析价格间隔配置
func parsePriceGaps(gapsStr string, side string) []float64 {
	if gapsStr == "" {
		// 根据方向返回默认值（万分比）
		if side == "LONG" {
			return []float64{0, -10, -30, -50, -70} // 做多默认值（万分比）
		} else {
			return []float64{0, 10, 30, 50, 70} // 做空默认值（万分比）
		}
	}

	var gaps []float64
	for _, g := range strings.Split(gapsStr, ",") {
		if val, err := strconv.ParseFloat(strings.TrimSpace(g), 64); err == nil {
			gaps = append(gaps, val)
		}
	}
	return gaps
}

// monitorEntryOrder 监控开仓订单
func monitorEntryOrder(cfg *config.Config, strategy *models.FuturesStrategy, orderID int64) {
	// 获取用户信息
	var user models.User
	if err := cfg.DB.First(&user, strategy.UserID).Error; err != nil {
		return
	}

	// 解密API密钥
	apiKey, err := user.GetDecryptedAPIKey()
	if err != nil {
		return
	}
	secretKey, err := user.GetDecryptedSecretKey()
	if err != nil {
		return
	}

	client := binance.NewFuturesClient(apiKey, secretKey)

	// 定期检查订单状态
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(10 * time.Minute) // 10分钟超时

	for {
		select {
		case <-ticker.C:
			order, err := client.NewGetOrderService().
				Symbol(strategy.Symbol).
				OrderID(orderID).
				Do(context.Background())

			if err != nil {
				log.Printf("查询订单状态失败: %v", err)
				continue
			}

			// 更新订单状态
			cfg.DB.Model(&models.FuturesOrder{}).
				Where("order_id = ?", orderID).
				Updates(map[string]interface{}{
					"status":       string(order.Status),
					"executed_qty": order.ExecutedQuantity,
					"avg_price":    order.AvgPrice,
				})

			// 检查订单是否成交
			if order.Status == futures.OrderStatusTypeFilled {
				log.Printf("开仓订单成交: 策略ID=%d, OrderID=%d", strategy.ID, orderID)

				// 创建持仓记录
				avgPrice, _ := strconv.ParseFloat(order.AvgPrice, 64)
				execQty, _ := strconv.ParseFloat(order.ExecutedQuantity, 64)

				position := models.FuturesPosition{
					UserID:       strategy.UserID,
					StrategyID:   strategy.ID,
					Symbol:       strategy.Symbol,
					PositionSide: strategy.Side,
					EntryPrice:   avgPrice,
					Quantity:     execQty,
					Leverage:     strategy.Leverage,
					MarginType:   strategy.MarginType,
					Status:       "open",
					OpenedAt:     time.Now(),
				}

				cfg.DB.Create(&position)

				// 更新策略状态
				strategy.Status = "position_opened"
				strategy.CurrentPositionId = orderID
				cfg.DB.Save(strategy)

				// 创建止盈订单
				createTakeProfitOrder(cfg, client, strategy, execQty)

				// 如果设置了止损，创建止损订单
				if strategy.StopLossRate > 0 {
					createStopLossOrder(cfg, client, strategy, execQty)
				}

				return
			} else if order.Status == futures.OrderStatusTypeCanceled ||
				order.Status == futures.OrderStatusTypeExpired ||
				order.Status == futures.OrderStatusTypeRejected {
				log.Printf("开仓订单失败: OrderID=%d, Status=%s", orderID, order.Status)
				updateStrategyStatus(cfg.DB, strategy, "cancelled", string(order.Status))
				return
			}

		case <-timeout:
			// 超时取消订单
			log.Printf("开仓订单超时，取消订单: OrderID=%d", orderID)
			_, cancelErr := client.NewCancelOrderService().
				Symbol(strategy.Symbol).
				OrderID(orderID).
				Do(context.Background())
			if cancelErr != nil {
				log.Printf("取消订单失败: %v", cancelErr)
			}

			updateStrategyStatus(cfg.DB, strategy, "cancelled", "timeout")
			return
		}
	}
}

// createTakeProfitOrder 创建止盈订单（优化避免吃单）
func createTakeProfitOrder(cfg *config.Config, client *futures.Client,
	strategy *models.FuturesStrategy, quantity float64) {

	// 获取当前深度
	depth, err := client.NewDepthService().
		Symbol(strategy.Symbol).
		Limit(5).
		Do(context.Background())
	if err != nil {
		log.Printf("获取深度失败，使用策略预设止盈价: %v", err)
		// 如果获取深度失败，使用策略中的止盈价格
	} else {
		// 检查止盈价格是否会立即吃单
		if strategy.Side == "LONG" {
			// 做多止盈是卖出，检查买一价
			if len(depth.Bids) > 0 {
				bidPrice, _ := strconv.ParseFloat(depth.Bids[0].Price, 64)
				if strategy.TakeProfitPrice <= bidPrice {
					// 如果止盈价格低于或等于买一价，会立即吃单
					log.Printf("警告：止盈价格 %.8f 低于买一价 %.8f，可能立即成交",
						strategy.TakeProfitPrice, bidPrice)
				}
			}
		} else {
			// 做空止盈是买入，检查卖一价
			if len(depth.Asks) > 0 {
				askPrice, _ := strconv.ParseFloat(depth.Asks[0].Price, 64)
				if strategy.TakeProfitPrice >= askPrice {
					// 如果止盈价格高于或等于卖一价，会立即吃单
					log.Printf("警告：止盈价格 %.8f 高于卖一价 %.8f，可能立即成交",
						strategy.TakeProfitPrice, askPrice)
				}
			}
		}
	}

	// 确定止盈方向
	side := futures.SideTypeSell
	if strategy.Side == "SHORT" {
		side = futures.SideTypeBuy
	}

	order, err := client.NewCreateOrderService().
		Symbol(strategy.Symbol).
		Side(side).
		PositionSide(futures.PositionSideType(strategy.Side)).
		Type(futures.OrderTypeLimit).
		TimeInForce(futures.TimeInForceTypeGTC).
		Quantity(fmt.Sprintf("%.8f", quantity)).
		Price(fmt.Sprintf("%.8f", strategy.TakeProfitPrice)).
		Do(context.Background())

	if err != nil {
		log.Printf("创建止盈订单失败: %v", err)
		return
	}

	// 保存订单记录
	dbOrder := models.FuturesOrder{
		UserID:       strategy.UserID,
		StrategyID:   strategy.ID,
		Symbol:       strategy.Symbol,
		Side:         string(side),
		PositionSide: strategy.Side,
		Type:         "LIMIT",
		Price:        strategy.TakeProfitPrice,
		Quantity:     quantity,
		OrderID:      order.OrderID,
		Status:       string(order.Status),
		OrderPurpose: "take_profit",
	}

	if err := cfg.DB.Create(&dbOrder).Error; err != nil {
		log.Printf("保存止盈订单失败: %v", err)
	}

	log.Printf("止盈订单创建成功: 策略ID=%d, OrderID=%d, Price=%.8f",
		strategy.ID, order.OrderID, strategy.TakeProfitPrice)
}

// createStopLossOrder 创建止损订单
func createStopLossOrder(cfg *config.Config, client *futures.Client,
	strategy *models.FuturesStrategy, quantity float64) {

	// 确定止损方向
	side := futures.SideTypeSell
	if strategy.Side == "SHORT" {
		side = futures.SideTypeBuy
	}

	// 使用止损市价单
	order, err := client.NewCreateOrderService().
		Symbol(strategy.Symbol).
		Side(side).
		PositionSide(futures.PositionSideType(strategy.Side)).
		Type(futures.OrderTypeStopMarket).
		StopPrice(fmt.Sprintf("%.8f", strategy.StopLossPrice)).
		Quantity(fmt.Sprintf("%.8f", quantity)).
		Do(context.Background())

	if err != nil {
		log.Printf("创建止损订单失败: %v", err)
		return
	}

	// 保存订单记录
	dbOrder := models.FuturesOrder{
		UserID:       strategy.UserID,
		StrategyID:   strategy.ID,
		Symbol:       strategy.Symbol,
		Side:         string(side),
		PositionSide: strategy.Side,
		Type:         "STOP_MARKET",
		Price:        strategy.StopLossPrice,
		Quantity:     quantity,
		OrderID:      order.OrderID,
		Status:       string(order.Status),
		OrderPurpose: "stop_loss",
	}

	if err := cfg.DB.Create(&dbOrder).Error; err != nil {
		log.Printf("保存止损订单失败: %v", err)
	}

	log.Printf("止损订单创建成功: 策略ID=%d, OrderID=%d, StopPrice=%.8f",
		strategy.ID, order.OrderID, strategy.StopLossPrice)
}

// monitorFuturesPositions 监控期货持仓
func monitorFuturesPositions(cfg *config.Config) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 获取所有开仓中的持仓
		var positions []models.FuturesPosition
		if err := cfg.DB.Where("status = ?", "open").Find(&positions).Error; err != nil {
			continue
		}

		// 按用户分组
		userPositions := make(map[uint][]models.FuturesPosition)
		for _, pos := range positions {
			userPositions[pos.UserID] = append(userPositions[pos.UserID], pos)
		}

		// 更新每个用户的持仓
		for userID, userPos := range userPositions {
			go updateUserPositions(cfg, userID, userPos)
		}
	}
}

// updateUserPositions 更新用户持仓
func updateUserPositions(cfg *config.Config, userID uint, positions []models.FuturesPosition) {
	// 获取用户信息
	var user models.User
	if err := cfg.DB.First(&user, userID).Error; err != nil {
		return
	}

	// 解密API密钥
	apiKey, err := user.GetDecryptedAPIKey()
	if err != nil {
		return
	}
	secretKey, err := user.GetDecryptedSecretKey()
	if err != nil {
		return
	}

	client := binance.NewFuturesClient(apiKey, secretKey)

	// 获取账户信息
	account, err := client.NewGetAccountService().Do(context.Background())
	if err != nil {
		return
	}

	// 创建持仓映射
	positionMap := make(map[string]*futures.AccountPosition)
	for _, pos := range account.Positions {
		key := pos.Symbol + "_" + string(pos.PositionSide)
		positionMap[key] = pos // pos 已经是指针类型
	}

	// 更新本地持仓
	for _, pos := range positions {
		key := pos.Symbol + "_" + pos.PositionSide
		if accPos, exists := positionMap[key]; exists {
			// 更新持仓信息
			unrealizedPnl, _ := strconv.ParseFloat(accPos.UnrealizedProfit, 64)

			updates := map[string]interface{}{
				"unrealized_pnl": unrealizedPnl,
				"updated_at":     time.Now(),
			}

			cfg.DB.Model(&pos).Updates(updates)

			// 检查是否已平仓
			posAmt, _ := strconv.ParseFloat(accPos.PositionAmt, 64)
			if posAmt == 0 {
				// 持仓已平，更新状态
				pos.Status = "closed"
				now := time.Now()
				pos.ClosedAt = &now
				cfg.DB.Save(&pos)

				// 更新策略状态
				var strategy models.FuturesStrategy
				if err := cfg.DB.First(&strategy, pos.StrategyID).Error; err == nil {
					strategy.Status = "completed"
					strategy.CompletedAt = &now
					cfg.DB.Save(&strategy)
				}
			}
		}
	}
}

// checkFuturesOrders 检查期货订单状态
func checkFuturesOrders(cfg *config.Config) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 获取所有未完成的订单
		var orders []models.FuturesOrder
		if err := cfg.DB.Where("status IN ?", []string{"NEW", "PARTIALLY_FILLED"}).
			Find(&orders).Error; err != nil {
			continue
		}

		// 按用户分组
		userOrders := make(map[uint][]models.FuturesOrder)
		for _, order := range orders {
			userOrders[order.UserID] = append(userOrders[order.UserID], order)
		}

		// 处理每个用户的订单
		for userID, userOrderList := range userOrders {
			go checkFuturesUserOrders(cfg, userID, userOrderList)
		}
	}
}

// checkFuturesUserOrders 检查用户订单（重命名以避免冲突）
func checkFuturesUserOrders(cfg *config.Config, userID uint, orders []models.FuturesOrder) {
	// 获取用户信息
	var user models.User
	if err := cfg.DB.First(&user, userID).Error; err != nil {
		return
	}

	// 解密API密钥
	apiKey, err := user.GetDecryptedAPIKey()
	if err != nil {
		return
	}
	secretKey, err := user.GetDecryptedSecretKey()
	if err != nil {
		return
	}

	client := binance.NewFuturesClient(apiKey, secretKey)

	// 批量查询订单
	for _, order := range orders {
		futuresOrder, err := client.NewGetOrderService().
			Symbol(order.Symbol).
			OrderID(order.OrderID).
			Do(context.Background())

		if err != nil {
			continue
		}

		// 更新订单状态
		execQty, _ := strconv.ParseFloat(futuresOrder.ExecutedQuantity, 64)
		avgPrice, _ := strconv.ParseFloat(futuresOrder.AvgPrice, 64)

		updates := map[string]interface{}{
			"status":       string(futuresOrder.Status),
			"executed_qty": execQty,
			"avg_price":    avgPrice,
			"updated_at":   time.Now(),
		}

		cfg.DB.Model(&order).Updates(updates)

		// 如果是止盈或止损订单成交，更新相关记录
		if futuresOrder.Status == futures.OrderStatusTypeFilled &&
			(order.OrderPurpose == "take_profit" || order.OrderPurpose == "stop_loss") {

			// 计算盈亏
			var position models.FuturesPosition
			if err := cfg.DB.Where("strategy_id = ? AND status = ?",
				order.StrategyID, "open").First(&position).Error; err == nil {

				// 计算已实现盈亏
				var realizedPnl float64
				if order.PositionSide == "LONG" {
					realizedPnl = (avgPrice - position.EntryPrice) * execQty
				} else {
					realizedPnl = (position.EntryPrice - avgPrice) * execQty
				}

				// 更新持仓状态
				position.RealizedPnl = realizedPnl
				position.Status = "closed"
				now := time.Now()
				position.ClosedAt = &now
				cfg.DB.Save(&position)

				// 更新策略状态
				var strategy models.FuturesStrategy
				if err := cfg.DB.First(&strategy, order.StrategyID).Error; err == nil {
					strategy.Status = "completed"
					strategy.CompletedAt = &now
					cfg.DB.Save(&strategy)

					log.Printf("策略 %d 完成，盈亏: %.8f", strategy.ID, realizedPnl)

					// 检查是否需要自动重启
					if strategy.AutoRestart && strategy.Enabled {
						log.Printf("策略 %d 设置了自动重启，正在创建新策略...", strategy.ID)

						// 创建新的策略（复制原策略配置）
						newStrategy := models.FuturesStrategy{
							UserID:            strategy.UserID,
							StrategyName:      strategy.StrategyName,
							Symbol:            strategy.Symbol,
							Side:              strategy.Side,
							StrategyType:      strategy.StrategyType,
							BasePrice:         strategy.BasePrice,
							EntryPrice:        0, // 重置为0
							EntryPriceFloat:   strategy.EntryPriceFloat,
							Leverage:          strategy.Leverage,
							Quantity:          strategy.Quantity,
							TakeProfitRate:    strategy.TakeProfitRate,
							TakeProfitPrice:   0, // 重置为0
							StopLossRate:      strategy.StopLossRate,
							StopLossPrice:     0, // 重置为0
							MarginType:        strategy.MarginType,
							IcebergLevels:     strategy.IcebergLevels,
							IcebergQuantities: strategy.IcebergQuantities,
							IcebergPriceGaps:  strategy.IcebergPriceGaps,
							AutoRestart:       strategy.AutoRestart, // 保持自动重启设置
							Enabled:           true,
							Status:            "waiting",
						}

						if err := cfg.DB.Create(&newStrategy).Error; err != nil {
							log.Printf("自动重启策略失败: %v", err)
						} else {
							log.Printf("策略 %d 已自动重启，新策略ID: %d", strategy.ID, newStrategy.ID)
						}
					}
				}
			}
		}
	}
}

// Helper functions

// setLeverage 设置杠杆
func setLeverage(client *futures.Client, symbol string, leverage int) error {
	_, err := client.NewChangeLeverageService().
		Symbol(symbol).
		Leverage(leverage).
		Do(context.Background())
	return err
}

// setMarginType 设置保证金模式
func setMarginType(client *futures.Client, symbol string, marginType string) error {
	err := client.NewChangeMarginTypeService().
		Symbol(symbol).
		MarginType(futures.MarginType(marginType)).
		Do(context.Background())
	return err
}

// updateStrategyStatus 更新策略状态
func updateStrategyStatus(db *gorm.DB, strategy *models.FuturesStrategy, status string, reason string) {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if status == "completed" || status == "cancelled" {
		now := time.Now()
		updates["completed_at"] = &now
	}

	db.Model(strategy).Updates(updates)

	if reason != "" {
		log.Printf("策略 %d 状态更新为 %s: %s", strategy.ID, status, reason)
	}
}
