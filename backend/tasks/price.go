// backend/tasks/price.go - 完整版本（修复并发问题）

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
var wsConnections sync.Map // 管理WebSocket连接

// StrategyExecutionManager 策略执行管理器
type StrategyExecutionManager struct {
	locks          sync.Map // strategyID -> *StrategyLock
	executionTimes sync.Map // strategyID -> time.Time
	mu             sync.Mutex
}

// StrategyLock 策略锁
type StrategyLock struct {
	mu              sync.Mutex
	executing       bool
	lastExecuteTime time.Time
	minInterval     time.Duration
}

var strategyManager = &StrategyExecutionManager{}

// TryExecuteStrategy 尝试执行策略（带并发控制）
func (m *StrategyExecutionManager) TryExecuteStrategy(strategyID uint, minInterval time.Duration) (unlock func(), canExecute bool) {
	// 获取或创建策略锁
	lockInterface, _ := m.locks.LoadOrStore(strategyID, &StrategyLock{
		minInterval: minInterval,
	})
	lock := lockInterface.(*StrategyLock)

	lock.mu.Lock()

	// 检查是否正在执行
	if lock.executing {
		lock.mu.Unlock()
		return nil, false
	}

	// 检查执行间隔
	if time.Since(lock.lastExecuteTime) < lock.minInterval {
		lock.mu.Unlock()
		return nil, false
	}

	// 标记为正在执行
	lock.executing = true
	lock.lastExecuteTime = time.Now()

	// 返回解锁函数
	unlock = func() {
		lock.mu.Lock()
		lock.executing = false
		lock.mu.Unlock()
	}

	lock.mu.Unlock()
	return unlock, true
}

// WebSocketManager 管理WebSocket连接
type WebSocketManager struct {
	symbol       string
	users        sync.Map // userID -> true
	stopChan     chan struct{}
	doneC        chan struct{}
	cfg          *config.Config
	lastPrice    float64
	mu           sync.RWMutex
	reconnecting bool
	reconnectMu  sync.Mutex
}

// DBUpdateManager 管理数据库更新频率
type DBUpdateManager struct {
	lastUpdate map[string]time.Time
	mu         sync.Mutex
}

var dbUpdateManager = &DBUpdateManager{
	lastUpdate: make(map[string]time.Time),
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
			// 防止重复重连
			m.reconnectMu.Lock()
			if m.reconnecting {
				m.reconnectMu.Unlock()
				time.Sleep(1 * time.Second)
				continue
			}
			m.reconnecting = true
			m.reconnectMu.Unlock()

			m.connect()

			m.reconnectMu.Lock()
			m.reconnecting = false
			m.reconnectMu.Unlock()

			// 如果连接断开，等待5秒后重连
			select {
			case <-m.stopChan:
				return
			case <-time.After(5 * time.Second):
				log.Printf("准备重连 %s WebSocket", m.symbol)
			}
		}
	}
}

// connect 建立WebSocket连接
// connect 建立WebSocket连接
func (m *WebSocketManager) connect() {
	// 移除连接日志，只在错误时记录

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

			// 异步检查并执行策略
			go m.checkStrategies(uid, price)
			return true
		})

		// 更新数据库价格（限流）
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
	// 移除连接关闭日志
}

// updatePriceInDB 更新数据库中的价格（限流）
func (m *WebSocketManager) updatePriceInDB(price float64) {
	dbUpdateManager.mu.Lock()
	defer dbUpdateManager.mu.Unlock()

	lastUpdate, exists := dbUpdateManager.lastUpdate[m.symbol]
	if exists && time.Since(lastUpdate) < time.Second {
		return
	}

	dbUpdateManager.lastUpdate[m.symbol] = time.Now()

	go func() {
		priceModel := models.Price{
			Symbol:    m.symbol,
			Price:     fmt.Sprintf("%.8f", price),
			UpdatedAt: time.Now(),
		}

		if err := m.cfg.DB.Where("symbol = ?", m.symbol).Assign(priceModel).FirstOrCreate(&priceModel).Error; err != nil {
			// 只在错误时记录
			log.Printf("保存 %s 价格失败: %v", m.symbol, err)
		}
	}()
}

// 添加用户缓存
var userCache sync.Map // userID -> *models.User

// checkStrategies 检查并执行策略（使用新的并发控制）
func (m *WebSocketManager) checkStrategies(userID uint, currentPrice float64) {
	// 尝试从缓存获取用户信息
	var user *models.User
	if cached, ok := userCache.Load(userID); ok {
		user = cached.(*models.User)
		// 检查缓存是否过期（5分钟）
		if user != nil && time.Since(user.UpdatedAt) > 5*time.Minute {
			userCache.Delete(userID)
			user = nil
		}
	}

	// 如果缓存中没有，从数据库获取
	if user == nil {
		var dbUser models.User
		if err := m.cfg.DB.First(&dbUser, userID).Error; err != nil {
			// 只在错误时记录
			if err != gorm.ErrRecordNotFound {
				log.Printf("用户未找到: ID=%d, error=%v", userID, err)
			}
			return
		}
		user = &dbUser
		// 存入缓存
		userCache.Store(userID, user)
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

	if apiKey == "" || secretKey == "" {
		return
	}

	// 查询用户的活跃策略 - 使用更精确的查询
	var strategies []models.Strategy
	if err := m.cfg.DB.
		Select("id", "symbol", "side", "price", "enabled", "pending_batch", "strategy_type", "total_quantity",
			"buy_quantities", "sell_quantities", "buy_depth_levels", "sell_depth_levels",
			"buy_basis_points", "sell_basis_points", "cancel_after_minutes").
		Where("user_id = ? AND symbol = ? AND status = ? AND enabled = ? AND pending_batch = ?",
			userID, m.symbol, "active", true, false).
		Where("deleted_at IS NULL").
		Find(&strategies).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Printf("获取用户 %d 的 %s 策略失败: %v", userID, m.symbol, err)
		}
		return
	}

	if len(strategies) == 0 {
		return
	}

	client := binance.NewClient(apiKey, secretKey)

	for _, strategy := range strategies {
		// 使用新的并发控制机制
		unlock, canExecute := strategyManager.TryExecuteStrategy(strategy.ID, 30*time.Second)
		if !canExecute {
			continue
		}

		// 在goroutine中执行策略
		go func(s models.Strategy) {
			defer unlock()
			m.executeStrategy(client, s, userID, currentPrice)
		}(strategy)
	}
}

// executeStrategy 执行策略 - 修复版本
func (m *WebSocketManager) executeStrategy(client *binance.Client, strategy models.Strategy, userID uint, currentPrice float64) {
	// 检查策略触发条件
	shouldExecute := false
	if strategy.Side == "SELL" && currentPrice >= strategy.Price {
		shouldExecute = true
	} else if strategy.Side == "BUY" && currentPrice <= strategy.Price {
		shouldExecute = true
	}

	if !shouldExecute {
		return
	}

	// 只记录策略触发
	log.Printf("策略 %d 触发: %s %s @ %.8f", strategy.ID, strategy.Side, strategy.Symbol, currentPrice)

	// 双重检查策略状态（使用事务）
	tx := m.cfg.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("策略执行恐慌: %v", r)
		}
	}()

	var currentStrategy models.Strategy
	if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&currentStrategy, strategy.ID).Error; err != nil {
		tx.Rollback()
		log.Printf("查询策略 %d 失败: %v", strategy.ID, err)
		return
	}

	if currentStrategy.PendingBatch || !currentStrategy.Enabled {
		tx.Rollback()
		return
	}

	// 标记策略正在执行
	if err := tx.Model(&currentStrategy).Update("pending_batch", true).Error; err != nil {
		tx.Rollback()
		log.Printf("更新策略 %d pending_batch 失败: %v", strategy.ID, err)
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		log.Printf("提交事务失败: %v", err)
		return
	}

	// 获取市场深度
	depth, err := client.NewDepthService().Symbol(strategy.Symbol).Limit(20).Do(context.Background())
	if err != nil {
		log.Printf("获取 %s 深度失败: %v", strategy.Symbol, err)
		m.cfg.DB.Model(&strategy).Update("pending_batch", false)
		return
	}

	// 执行下单
	err = placeOrders(client, strategy, userID, currentPrice, depth, strategy.Side, m.cfg)
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

	// 使用批量查询，减少数据库访问
	var symbols []models.CustomSymbol
	if err := cfg.DB.
		Select("DISTINCT symbol, user_id").
		Where("deleted_at IS NULL").
		Find(&symbols).Error; err != nil {
		log.Printf("获取自定义交易对失败: %v", err)
		return
	}

	// 预加载用户信息到缓存
	userIDs := make([]uint, 0)
	for _, s := range symbols {
		userIDs = append(userIDs, s.UserID)
	}

	// 批量查询用户
	var users []models.User
	if err := cfg.DB.Where("id IN ?", userIDs).Find(&users).Error; err == nil {
		for _, user := range users {
			userCache.Store(user.ID, &user)
		}
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
				uid := userID.(uint)
				// 检查用户是否还有活跃的策略
				var count int64
				cfg.DB.Model(&models.Strategy{}).Where(
					"user_id = ? AND symbol = ? AND enabled = ? AND deleted_at IS NULL",
					uid, symbol, true,
				).Count(&count)

				if count == 0 {
					manager.users.Delete(userID)
					MonitoredSymbols.Delete(fmt.Sprintf("%s|%d", symbol, uid))
					log.Printf("移除用户 %d 对 %s 的监控", uid, symbol)
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

// placeOrders 下单函数 - 支持自定义取消时间（使用解密后的API密钥）
func placeOrders(client *binance.Client, strategy models.Strategy, userID uint, currentPrice float64, depth *binance.DepthResponse, side string, cfg *config.Config) error {
	var quantities []float64
	var depthLevels []float64
	var placedOrders []models.Order

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

	// 验证数量分配
	if err := validateQuantities(quantities); err != nil {
		return fmt.Errorf("数量配置错误: %v", err)
	}

	// 如果没有配置，使用默认值
	if len(quantities) == 0 {
		quantities = []float64{1.0}
		depthLevels = []float64{1.0}
	}

	// 计算各层订单价格
	priceLevels, err := calculatePriceLevels(strategy, side, depth, quantities, depthLevels, pricePrecision)
	if err != nil {
		return err
	}

	// 计算订单取消时间
	cancelAfterMinutes := strategy.CancelAfterMinutes
	if cancelAfterMinutes <= 0 {
		cancelAfterMinutes = 120 // 默认120分钟
	}
	cancelAfterDuration := time.Duration(cancelAfterMinutes) * time.Minute

	// 执行下单
	successCount := 0
	failCount := 0

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

		order, err := client.NewCreateOrderService().
			Symbol(strategy.Symbol).
			Side(binance.SideType(side)).
			Type(binance.OrderTypeLimit).
			TimeInForce(binance.TimeInForceTypeGTC).
			Quantity(quantityStr).
			Price(priceStr).
			Do(context.Background())

		if err != nil {
			failCount++
			// 如果是第一个订单就失败，回滚所有
			if successCount == 0 {
				return fmt.Errorf("首个订单下单失败: %v", err)
			}
			// 否则继续尝试下一个订单
			continue
		}

		// 保存订单记录，使用策略的自定义取消时间
		dbOrder := models.Order{
			StrategyID:  strategy.ID,
			UserID:      userID,
			Symbol:      strategy.Symbol,
			Side:        side,
			Price:       price,
			Quantity:    quantity,
			OrderID:     order.OrderID,
			Status:      "pending",
			CancelAfter: time.Now().Add(cancelAfterDuration),
		}

		if err := cfg.DB.Create(&dbOrder).Error; err != nil {
			log.Printf("保存订单失败: %v", err)
			// 取消刚下的订单
			client.NewCancelOrderService().Symbol(strategy.Symbol).OrderID(order.OrderID).Do(context.Background())
			continue
		}

		placedOrders = append(placedOrders, dbOrder)
		successCount++
	}

	if successCount == 0 {
		return fmt.Errorf("所有订单都失败了")
	}

	// 只记录最终结果
	if failCount > 0 {
		log.Printf("策略 %d: 成功 %d 笔，失败 %d 笔", strategy.ID, successCount, failCount)
	} else {
		log.Printf("策略 %d: 成功下单 %d 笔", strategy.ID, successCount)
	}

	return nil
}

// PriceLevel 价格级别
type PriceLevel struct {
	Price float64
}

// calculatePriceLevels 根据市场深度计算价格级别 - 支持自定义万分比
func calculatePriceLevels(strategy models.Strategy, side string, depth *binance.DepthResponse,
	quantities []float64, depthLevels []float64, pricePrecision int) ([]PriceLevel, error) {

	var priceLevels []PriceLevel
	var depthData []binance.Ask

	// 选择正确的深度数据
	if side == "SELL" {
		depthData = depth.Bids // 卖单看买盘深度
	} else {
		depthData = depth.Asks // 买单看卖盘深度
	}

	if len(depthData) == 0 {
		return nil, fmt.Errorf("没有%s深度数据", side)
	}

	// 获取基准价格
	basePrice, err := strconv.ParseFloat(depthData[0].Price, 64)
	if err != nil {
		return nil, fmt.Errorf("解析基准价格失败: %v", err)
	}

	switch strategy.StrategyType {
	case "simple":
		// 简单策略：直接使用对手盘第一档价格
		priceLevels = append(priceLevels, PriceLevel{Price: basePrice})

	case "iceberg":
		// 冰山策略：在对手盘深度中分层下单
		// 使用固定的万分比偏移
		var basisPoints []float64
		if side == "SELL" {
			basisPoints = []float64{0, 1, 3, 5, 7} // 卖单价格递增
		} else {
			basisPoints = []float64{0, -1, -3, -5, -7} // 买单价格递减
		}

		for i := 0; i < len(quantities) && i < len(basisPoints); i++ {
			multiplier := 1 + (basisPoints[i] / 10000)
			price := basePrice * multiplier
			priceLevels = append(priceLevels, PriceLevel{Price: price})
		}

	case "custom":
		// 自定义策略：使用用户指定的万分比
		var basisPoints []float64

		// 解析万分比配置
		if side == "BUY" && strategy.BuyBasisPoints != "" {
			for _, bp := range strings.Split(strategy.BuyBasisPoints, ",") {
				if basisPoint, err := strconv.ParseFloat(strings.TrimSpace(bp), 64); err == nil {
					basisPoints = append(basisPoints, basisPoint)
				}
			}
		} else if side == "SELL" && strategy.SellBasisPoints != "" {
			for _, bp := range strings.Split(strategy.SellBasisPoints, ",") {
				if basisPoint, err := strconv.ParseFloat(strings.TrimSpace(bp), 64); err == nil {
					basisPoints = append(basisPoints, basisPoint)
				}
			}
		}

		// 如果没有万分比配置，使用深度级别
		if len(basisPoints) == 0 {
			// 兼容旧版本：使用深度级别
			for i := 0; i < len(quantities) && i < len(depthLevels); i++ {
				level := int(depthLevels[i]) - 1
				if level < 0 {
					level = 0
				}
				if level >= len(depthData) {
					level = len(depthData) - 1
				}

				price, err := strconv.ParseFloat(depthData[level].Price, 64)
				if err != nil {
					continue
				}

				priceLevels = append(priceLevels, PriceLevel{Price: price})
			}
		} else {
			// 使用万分比计算价格
			for i := 0; i < len(quantities) && i < len(basisPoints); i++ {
				multiplier := 1 + (basisPoints[i] / 10000)
				price := basePrice * multiplier

				// 确保价格为正数
				if price <= 0 {
					price = basePrice * 0.9999
				}

				priceLevels = append(priceLevels, PriceLevel{Price: price})
			}
		}

	default:
		return nil, fmt.Errorf("未知策略类型: %s", strategy.StrategyType)
	}

	// 对价格进行精度处理
	for i := range priceLevels {
		priceLevels[i].Price = math.Round(priceLevels[i].Price*math.Pow(10, float64(pricePrecision))) /
			math.Pow(10, float64(pricePrecision))
	}

	return priceLevels, nil
}

// parseSymbolInfo 解析交易对信息
func parseSymbolInfo(symbol binance.Symbol) (pricePrecision, quantityPrecision int, minNotional float64) {
	// 设置默认值
	pricePrecision = 8
	quantityPrecision = 8
	minNotional = 10.0

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
		case "MIN_NOTIONAL", "NOTIONAL":
			if minNot, ok := filter["minNotional"].(string); ok {
				if mn, err := strconv.ParseFloat(minNot, 64); err == nil {
					minNotional = mn
				}
			}
		}
	}

	// 限制精度范围
	if pricePrecision > 8 {
		pricePrecision = 8
	}
	if quantityPrecision > 8 {
		quantityPrecision = 8
	}

	return
}

// parseQuantitiesAndDepthLevels 解析数量和深度级别配置
func parseQuantitiesAndDepthLevels(quantitiesStr, depthLevelsStr string, strategyID uint) ([]float64, []float64, error) {
	var quantities, depthLevels []float64

	// 去除可能的方括号和空格
	quantitiesStr = strings.Trim(strings.TrimSpace(quantitiesStr), "[]")
	depthLevelsStr = strings.Trim(strings.TrimSpace(depthLevelsStr), "[]")

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
			if lvl, err := strconv.ParseFloat(d, 64); err == nil && lvl > 0 {
				depthLevels = append(depthLevels, lvl)
			}
		}
	}

	// 验证长度匹配
	if len(quantities) > 0 && len(depthLevels) > 0 && len(quantities) != len(depthLevels) {
		return nil, nil, fmt.Errorf("数量和深度级别数量不匹配: %d vs %d", len(quantities), len(depthLevels))
	}

	// 如果没有深度级别，使用默认值
	if len(quantities) > 0 && len(depthLevels) == 0 {
		depthLevels = make([]float64, len(quantities))
		for i := range depthLevels {
			depthLevels[i] = float64(i + 1)
		}
	}

	return quantities, depthLevels, nil
}

// validateQuantities 验证数量分配
func validateQuantities(quantities []float64) error {
	if len(quantities) == 0 {
		return nil // 空配置使用默认值
	}

	sum := 0.0
	for _, q := range quantities {
		if q <= 0 {
			return fmt.Errorf("数量必须大于0")
		}
		sum += q
	}

	// 允许一定的浮点误差
	if math.Abs(sum-1.0) > 0.001 {
		return fmt.Errorf("数量总和必须为1.0，当前为%.4f", sum)
	}

	return nil
}

// StopSymbolMonitoring 停止对特定用户的交易对监控
func StopSymbolMonitoring(symbol string, userID uint) {
	key := fmt.Sprintf("%s|%d", symbol, userID)

	// 从监控列表中移除
	MonitoredSymbols.Delete(key)
	PriceMonitor.Delete(key)

	// 检查是否还有其他用户在监控这个交易对
	if manager, ok := wsConnections.Load(symbol); ok {
		wsManager := manager.(*WebSocketManager)
		wsManager.users.Delete(userID)

		// 检查是否还有其他用户
		activeUsers := 0
		wsManager.users.Range(func(_, _ interface{}) bool {
			activeUsers++
			return false // 只需要计数
		})

		// 如果没有其他用户，关闭WebSocket连接
		if activeUsers == 0 {
			select {
			case <-wsManager.stopChan:
				// 已经关闭
			default:
				close(wsManager.stopChan)
			}
			wsConnections.Delete(symbol)
			log.Printf("停止 %s WebSocket 连接（用户 %d 移除后无其他用户）", symbol, userID)
		} else {
			log.Printf("用户 %d 停止监控 %s，还有 %d 个其他用户在监控", userID, symbol, activeUsers)
		}
	}

	log.Printf("用户 %d 停止监控交易对 %s", userID, symbol)
}
