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
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
)

var PriceMonitor sync.Map
var MonitoredSymbols sync.Map
var strategyLocks sync.Map
var wsConnections sync.Map // 管理WebSocket连接

// WebSocketManager 管理WebSocket连接
type WebSocketManager struct {
	symbol    string
	users     sync.Map // userID -> true
	stopChan  chan struct{}
	doneC     chan struct{}
	cfg       *config.Config
	lastPrice float64
	mu        sync.RWMutex
}

// MonitorNewSymbol 启动对新交易对的监控
func MonitorNewSymbol(symbol string, userID uint, cfg *config.Config) {
	key := fmt.Sprintf("%s|%d", symbol, userID)
	if _, loaded := MonitoredSymbols.LoadOrStore(key, true); !loaded {
		log.Printf("为用户 %d 启动 %s 价格监控", userID, symbol)

		// 检查是否已有该交易对的WebSocket连接
		if manager, ok := wsConnections.Load(symbol); ok {
			// 将用户添加到现有连接
			wsManager := manager.(*WebSocketManager)
			wsManager.users.Store(userID, true)
			log.Printf("用户 %d 加入现有 %s WebSocket 连接", userID, symbol)
		} else {
			// 创建新的WebSocket连接
			wsManager := &WebSocketManager{
				symbol:   symbol,
				stopChan: make(chan struct{}),
				cfg:      cfg,
			}
			wsManager.users.Store(userID, true)
			wsConnections.Store(symbol, wsManager)
			go wsManager.start()
		}
	}
}

// start 启动WebSocket连接
func (m *WebSocketManager) start() {
	for {
		select {
		case <-m.stopChan:
			log.Printf("停止 %s WebSocket 连接", m.symbol)
			return
		default:
			m.connect()
			// 如果连接断开，等待5秒后重连
			time.Sleep(5 * time.Second)
		}
	}
}

// connect 建立WebSocket连接
func (m *WebSocketManager) connect() {
	log.Printf("建立 %s WebSocket 连接", m.symbol)

	// 使用交易流而不是深度流
	wsTradeHandler := func(event *binance.WsTradeEvent) {
		price, err := strconv.ParseFloat(event.Price, 64)
		if err != nil {
			log.Printf("解析 %s 价格错误: %v", m.symbol, err)
			return
		}

		m.mu.Lock()
		m.lastPrice = price
		m.mu.Unlock()

		// 更新所有用户的价格
		m.users.Range(func(userID, _ interface{}) bool {
			uid := userID.(uint)
			key := fmt.Sprintf("%s|%d", m.symbol, uid)
			PriceMonitor.Store(key, price)

			// 检查并执行策略
			go m.checkStrategies(uid, price)
			return true
		})

		// 更新数据库价格（限流，每秒最多更新一次）
		m.updatePriceInDB(price)
	}

	wsErrHandler := func(err error) {
		log.Printf("%s WebSocket 错误: %v", m.symbol, err)
	}

	// 使用交易流获取实时成交价
	doneC, _, err := binance.WsTradeServe(m.symbol, wsTradeHandler, wsErrHandler)
	if err != nil {
		log.Printf("启动 %s WebSocket 失败: %v", m.symbol, err)
		return
	}

	m.doneC = doneC
	<-doneC
	log.Printf("%s WebSocket 连接已关闭", m.symbol)
}

var lastDBUpdate = make(map[string]time.Time)
var dbUpdateMutex sync.Mutex

// updatePriceInDB 更新数据库中的价格（限流）
func (m *WebSocketManager) updatePriceInDB(price float64) {
	dbUpdateMutex.Lock()
	defer dbUpdateMutex.Unlock()

	lastUpdate, exists := lastDBUpdate[m.symbol]
	if exists && time.Since(lastUpdate) < time.Second {
		return
	}

	lastDBUpdate[m.symbol] = time.Now()

	priceModel := models.Price{
		Symbol:    m.symbol,
		Price:     fmt.Sprintf("%.8f", price),
		UpdatedAt: time.Now(),
	}

	if err := m.cfg.DB.Where("symbol = ?", m.symbol).Assign(priceModel).FirstOrCreate(&priceModel).Error; err != nil {
		log.Printf("保存 %s 价格失败: %v", m.symbol, err)
	}
}

// checkStrategies 检查并执行策略
func (m *WebSocketManager) checkStrategies(userID uint, currentPrice float64) {
	// 获取用户信息
	var user models.User
	if err := m.cfg.DB.First(&user, userID).Error; err != nil {
		log.Printf("用户未找到: ID=%d", userID)
		return
	}

	if user.APIKey == "" || user.SecretKey == "" {
		return
	}

	// 查询用户的活跃策略
	var strategies []models.Strategy
	if err := m.cfg.DB.Where(
		"user_id = ? AND symbol = ? AND status = ? AND enabled = ? AND deleted_at IS NULL AND pending_batch = ?",
		userID, m.symbol, "active", true, false,
	).Find(&strategies).Error; err != nil {
		log.Printf("获取用户 %d 的 %s 策略失败: %v", userID, m.symbol, err)
		return
	}

	client := binance.NewClient(user.APIKey, user.SecretKey)

	for _, strategy := range strategies {
		// 使用策略锁防止重复执行
		lockKey := fmt.Sprintf("%d", strategy.ID)
		lock, _ := strategyLocks.LoadOrStore(lockKey, &sync.Mutex{})
		mutex := lock.(*sync.Mutex)

		if !mutex.TryLock() {
			continue
		}

		go func(s models.Strategy) {
			defer mutex.Unlock()
			m.executeStrategy(client, s, userID, currentPrice)
		}(strategy)
	}
}

// executeStrategy 执行策略
func (m *WebSocketManager) executeStrategy(client *binance.Client, strategy models.Strategy, userID uint, currentPrice float64) {
	// 检查策略触发条件
	shouldExecute := false
	if strategy.Side == "SELL" && currentPrice >= strategy.Price {
		shouldExecute = true
		log.Printf("策略 %d 触发: 当前价格 %.2f >= 目标价格 %.2f (卖出)", strategy.ID, currentPrice, strategy.Price)
	} else if strategy.Side == "BUY" && currentPrice <= strategy.Price {
		shouldExecute = true
		log.Printf("策略 %d 触发: 当前价格 %.2f <= 目标价格 %.2f (买入)", strategy.ID, currentPrice, strategy.Price)
	}

	if !shouldExecute {
		return
	}

	// 标记策略正在执行
	if err := m.cfg.DB.Model(&strategy).Update("pending_batch", true).Error; err != nil {
		log.Printf("更新策略 %d pending_batch 失败: %v", strategy.ID, err)
		return
	}

	// 获取市场深度用于计算下单价格
	depth, err := client.NewDepthService().Symbol(strategy.Symbol).Limit(20).Do(context.Background())
	if err != nil {
		log.Printf("获取 %s 深度失败: %v", strategy.Symbol, err)
		m.cfg.DB.Model(&strategy).Update("pending_batch", false)
		return
	}

	// 执行下单
	if strategy.Side == "SELL" {
		err = placeOrders(client, strategy, userID, depth.Asks, "SELL", m.cfg)
	} else {
		err = placeOrders(client, strategy, userID, depth.Bids, "BUY", m.cfg)
	}

	if err != nil {
		log.Printf("策略 %d 下单失败: %v", strategy.ID, err)
		m.cfg.DB.Model(&strategy).Update("pending_batch", false)
	}
}

// StartPriceMonitoring 开始监控价格
func StartPriceMonitoring(cfg *config.Config) {
	if cfg.DB == nil {
		log.Println("数据库未初始化，跳过价格监控")
		return
	}

	// 查询所有需要监控的交易对
	var symbols []models.CustomSymbol
	if err := cfg.DB.Select("DISTINCT symbol, user_id").Where("deleted_at IS NULL").Find(&symbols).Error; err != nil {
		log.Printf("获取自定义交易对失败: %v", err)
		return
	}

	// 按交易对分组
	symbolUsers := make(map[string][]uint)
	for _, s := range symbols {
		symbolUsers[s.Symbol] = append(symbolUsers[s.Symbol], s.UserID)
	}

	// 为每个交易对创建一个WebSocket连接
	for symbol, users := range symbolUsers {
		wsManager := &WebSocketManager{
			symbol:   symbol,
			stopChan: make(chan struct{}),
			cfg:      cfg,
		}

		// 添加所有用户
		for _, userID := range users {
			wsManager.users.Store(userID, true)
			MonitoredSymbols.Store(fmt.Sprintf("%s|%d", symbol, userID), true)
		}

		wsConnections.Store(symbol, wsManager)
		go wsManager.start()

		log.Printf("启动 %s 价格监控，共 %d 个用户", symbol, len(users))
	}

	// 启动清理任务
	go cleanupInactiveConnections(cfg)
}

// cleanupInactiveConnections 清理不活跃的连接
func cleanupInactiveConnections(cfg *config.Config) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		wsConnections.Range(func(symbol, value interface{}) bool {
			manager := value.(*WebSocketManager)
			activeUsers := 0

			manager.users.Range(func(userID, _ interface{}) bool {
				// 检查用户是否还有活跃的策略
				var count int64
				cfg.DB.Model(&models.Strategy{}).Where(
					"user_id = ? AND symbol = ? AND enabled = ? AND deleted_at IS NULL",
					userID, symbol,
					true,
				).Count(&count)

				if count == 0 {
					manager.users.Delete(userID)
					MonitoredSymbols.Delete(fmt.Sprintf("%s|%d", symbol, userID))
					log.Printf("移除用户 %d 对 %s 的监控", userID, symbol)
				} else {
					activeUsers++
				}
				return true
			})

			// 如果没有活跃用户，关闭连接
			if activeUsers == 0 {
				close(manager.stopChan)
				wsConnections.Delete(symbol)
				log.Printf("关闭 %s WebSocket 连接（无活跃用户）", symbol)
			}

			return true
		})
	}
}

// placeOrders 下单函数（保持原有逻辑，但优化价格计算）
func placeOrders(client *binance.Client, strategy models.Strategy, userID uint, levels interface{}, side string, cfg *config.Config) error {
	var quantities []float64
	var depthLevels []float64
	var placedOrders []models.Order

	log.Printf("执行 %s 策略: ID=%d, 类型=%s", side, strategy.ID, strategy.StrategyType)

	// 获取交易所信息
	exchangeInfo, err := client.NewExchangeInfoService().Symbol(strategy.Symbol).Do(context.Background())
	if err != nil {
		return fmt.Errorf("获取交易所信息失败: %v", err)
	}

	var symbolInfo binance.Symbol
	for _, s := range exchangeInfo.Symbols {
		if s.Symbol == strategy.Symbol {
			symbolInfo = s
			break
		}
	}

	// 解析精度信息
	pricePrecision, quantityPrecision, minNotional := parseSymbolInfo(symbolInfo)
	log.Printf("市场 %s: 价格精度=%d, 数量精度=%d, 最小名义价值=%.2f",
		strategy.Symbol, pricePrecision, quantityPrecision, minNotional)

	// 解析数量和深度配置
	if side == "SELL" {
		quantities, depthLevels, err = parseQuantitiesAndDepthLevels(
			strategy.SellQuantities, strategy.SellDepthLevels, strategy.ID)
	} else {
		quantities, depthLevels, err = parseQuantitiesAndDepthLevels(
			strategy.BuyQuantities, strategy.BuyDepthLevels, strategy.ID)
	}

	if err != nil {
		return err
	}

	// 如果没有配置，使用默认值
	if len(quantities) == 0 {
		quantities = []float64{1.0}
		depthLevels = []float64{0.0}
	}

	// 根据策略类型计算价格
	priceLevels, err := calculatePriceLevels(strategy, side, levels, quantities, depthLevels, pricePrecision)
	if err != nil {
		return err
	}

	// 执行下单
	for i, priceLevel := range priceLevels {
		if i >= len(quantities) {
			break
		}

		price := priceLevel.Price
		quantity := strategy.TotalQuantity * quantities[i]

		// 确保满足最小名义价值
		if price*quantity < minNotional {
			quantity = math.Ceil(minNotional/price*math.Pow(10, float64(quantityPrecision))) /
				math.Pow(10, float64(quantityPrecision))
		}

		// 格式化价格和数量
		priceStr := fmt.Sprintf("%.*f", pricePrecision, price)
		quantityStr := fmt.Sprintf("%.*f", quantityPrecision, quantity)

		log.Printf("下单: 策略ID=%d, %s %s, 价格=%s, 数量=%s",
			strategy.ID, side, strategy.Symbol, priceStr, quantityStr)

		order, err := client.NewCreateOrderService().
			Symbol(strategy.Symbol).
			Side(binance.SideType(side)).
			Type(binance.OrderTypeLimit).
			TimeInForce(binance.TimeInForceTypeGTC).
			Quantity(quantityStr).
			Price(priceStr).
			Do(context.Background())

		if err != nil {
			// 回滚已下的订单
			for _, po := range placedOrders {
				client.NewCancelOrderService().Symbol(strategy.Symbol).OrderID(po.OrderID).Do(context.Background())
			}
			return fmt.Errorf("下单失败: %v", err)
		}

		// 保存订单记录
		dbOrder := models.Order{
			StrategyID:  strategy.ID,
			UserID:      userID,
			Symbol:      strategy.Symbol,
			Side:        side,
			Price:       price,
			Quantity:    quantity,
			OrderID:     order.OrderID,
			Status:      "pending",
			CancelAfter: time.Now().Add(2 * time.Hour),
		}

		if err := cfg.DB.Create(&dbOrder).Error; err != nil {
			// 取消所有订单
			client.NewCancelOrderService().Symbol(strategy.Symbol).OrderID(order.OrderID).Do(context.Background())
			for _, po := range placedOrders {
				client.NewCancelOrderService().Symbol(strategy.Symbol).OrderID(po.OrderID).Do(context.Background())
			}
			return fmt.Errorf("保存订单失败: %v", err)
		}

		placedOrders = append(placedOrders, dbOrder)
	}

	log.Printf("策略 %d 成功下单 %d 笔", strategy.ID, len(placedOrders))
	return nil
}

// PriceLevel 价格级别
type PriceLevel struct {
	Price float64
}

// calculatePriceLevels 计算价格级别
func calculatePriceLevels(strategy models.Strategy, side string, levels interface{},
	quantities []float64, depthLevels []float64, pricePrecision int) ([]PriceLevel, error) {

	var priceLevels []PriceLevel

	switch strategy.StrategyType {
	case "simple":
		// 简单策略：使用策略设定的价格
		priceLevels = append(priceLevels, PriceLevel{Price: strategy.Price})

	case "iceberg":
		// 冰山策略：在市场价附近分层下单
		basePrice := strategy.Price
		if side == "SELL" {
			// 卖单：在目标价格及以上分层
			factors := []float64{1.0, 1.0001, 1.0003, 1.0005, 1.0007}
			for i := 0; i < len(quantities) && i < len(factors); i++ {
				priceLevels = append(priceLevels, PriceLevel{
					Price: basePrice * factors[i],
				})
			}
		} else {
			// 买单：在目标价格及以下分层
			factors := []float64{1.0, 0.9999, 0.9997, 0.9995, 0.9993}
			for i := 0; i < len(quantities) && i < len(factors); i++ {
				priceLevels = append(priceLevels, PriceLevel{
					Price: basePrice * factors[i],
				})
			}
		}

	case "custom":
		// 自定义策略：根据深度级别计算价格
		if len(depthLevels) == 0 {
			return nil, fmt.Errorf("自定义策略需要深度级别配置")
		}

		// 从市场深度获取参考价格
		var refPrices []float64
		if side == "SELL" {
			asks := levels.([]binance.Ask)
			for i := 0; i < len(depthLevels) && i < len(asks); i++ {
				idx := int(depthLevels[i]) - 1
				if idx >= 0 && idx < len(asks) {
					price, _ := strconv.ParseFloat(asks[idx].Price, 64)
					refPrices = append(refPrices, price)
				}
			}
		} else {
			bids := levels.([]binance.Bid)
			for i := 0; i < len(depthLevels) && i < len(bids); i++ {
				idx := int(depthLevels[i]) - 1
				if idx >= 0 && idx < len(bids) {
					price, _ := strconv.ParseFloat(bids[idx].Price, 64)
					refPrices = append(refPrices, price)
				}
			}
		}

		// 使用参考价格或策略价格
		for i := 0; i < len(quantities); i++ {
			if i < len(refPrices) {
				priceLevels = append(priceLevels, PriceLevel{Price: refPrices[i]})
			} else {
				// 如果深度不够，使用策略价格
				priceLevels = append(priceLevels, PriceLevel{Price: strategy.Price})
			}
		}
	}

	return priceLevels, nil
}

// parseSymbolInfo 解析交易对信息
func parseSymbolInfo(symbol binance.Symbol) (pricePrecision, quantityPrecision int, minNotional float64) {
	for _, filter := range symbol.Filters {
		switch filter["filterType"] {
		case "PRICE_FILTER":
			if tickSize, ok := filter["tickSize"].(string); ok {
				if ts, err := strconv.ParseFloat(tickSize, 64); err == nil && ts > 0 {
					pricePrecision = int(-math.Log10(ts))
				}
			}
		case "LOT_SIZE":
			if stepSize, ok := filter["stepSize"].(string); ok {
				if ss, err := strconv.ParseFloat(stepSize, 64); err == nil && ss > 0 {
					quantityPrecision = int(-math.Log10(ss))
				}
			}
		case "NOTIONAL":
			if minNot, ok := filter["minNotional"].(string); ok {
				minNotional, _ = strconv.ParseFloat(minNot, 64)
			}
		}
	}

	// 设置默认值
	if pricePrecision == 0 {
		pricePrecision = 8
	}
	if quantityPrecision == 0 {
		quantityPrecision = 8
	}
	if minNotional == 0 {
		minNotional = 10.0
	}

	return
}

// parseQuantitiesAndDepthLevels 解析数量和深度级别配置
func parseQuantitiesAndDepthLevels(quantitiesStr, depthLevelsStr string, strategyID uint) ([]float64, []float64, error) {
	var quantities, depthLevels []float64

	// 去除可能的方括号
	quantitiesStr = strings.Trim(quantitiesStr, "[]")
	depthLevelsStr = strings.Trim(depthLevelsStr, "[]")

	// 解析数量
	if quantitiesStr != "" {
		for _, q := range strings.Split(quantitiesStr, ",") {
			q = strings.TrimSpace(q)
			if q == "" {
				continue
			}
			if qty, err := strconv.ParseFloat(q, 64); err == nil && qty > 0 {
				quantities = append(quantities, qty)
			}
		}
	}

	// 解析深度级别
	if depthLevelsStr != "" {
		for _, d := range strings.Split(depthLevelsStr, ",") {
			d = strings.TrimSpace(d)
			if d == "" {
				continue
			}
			if lvl, err := strconv.ParseFloat(d, 64); err == nil && lvl >= 0 {
				depthLevels = append(depthLevels, lvl)
			}
		}
	}

	// 验证长度匹配
	if len(quantities) > 0 && len(depthLevels) > 0 && len(quantities) != len(depthLevels) {
		return nil, nil, fmt.Errorf("数量和深度级别数量不匹配: %d vs %d", len(quantities), len(depthLevels))
	}

	return quantities, depthLevels, nil
}
