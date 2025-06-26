package tasks

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
	"gorm.io/gorm"
)

// StartDualInvestmentTasks 启动双币投资相关任务
func StartDualInvestmentTasks(cfg *config.Config) {
	// 产品同步任务 - 每5分钟执行一次
	go syncDualInvestmentProducts(cfg)

	// 策略执行任务 - 每分钟检查一次
	go executeDualInvestmentStrategies(cfg)

	// 订单结算监控 - 每10分钟检查一次
	go monitorDualInvestmentSettlement(cfg)
}

// syncDualInvestmentProducts 同步双币投资产品
func syncDualInvestmentProducts(cfg *config.Config) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// 立即执行一次
	doSyncProducts(cfg)

	for range ticker.C {
		doSyncProducts(cfg)
	}
}

// doSyncProducts 执行产品同步
func doSyncProducts(cfg *config.Config) {
	log.Println("开始同步双币投资产品...")

	// 获取所有用户的API密钥（这里简化处理，实际应该选择一个有效的API）
	var user models.User
	if err := cfg.DB.Where("api_key != ? AND secret_key != ?", "", "").First(&user).Error; err != nil {
		log.Printf("没有找到有效的API密钥用于同步产品")
		return
	}

	client := binance.NewClient(user.APIKey, user.SecretKey)

	// 注意：币安的双币投资API可能需要特殊的接口，这里使用模拟数据
	// 实际实现时需要查看币安的具体API文档

	// 模拟产品数据
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"}
	directions := []string{"UP", "DOWN"}

	for _, symbol := range symbols {
		// 获取当前价格
		prices, err := client.NewListPricesService().Symbol(symbol).Do(context.Background())
		if err != nil {
			log.Printf("获取 %s 价格失败: %v", symbol, err)
			continue
		}

		if len(prices) == 0 {
			continue
		}

		currentPrice, _ := strconv.ParseFloat(prices[0].Price, 64)

		// 分离基础资产和计价资产
		baseAsset := strings.TrimSuffix(symbol, "USDT")
		quoteAsset := "USDT"

		for _, direction := range directions {
			// 生成不同执行价格的产品
			for i := -5; i <= 5; i++ {
				if i == 0 {
					continue
				}

				// 计算执行价格（偏离当前价格的百分比）
				offset := float64(i) * 0.01 // 1%的间隔
				strikePrice := currentPrice * (1 + offset)

				// 根据偏离度计算年化收益率（简化模型）
				apy := math.Abs(offset) * 100 * 2 // 偏离越大，收益率越高
				if direction == "DOWN" && i > 0 {
					apy *= 1.2 // 看跌且执行价高于现价，收益更高
				} else if direction == "UP" && i < 0 {
					apy *= 1.2 // 看涨且执行价低于现价，收益更高
				}

				// 创建或更新产品
				product := models.DualInvestmentProduct{
					Symbol:         symbol,
					Direction:      direction,
					StrikePrice:    strikePrice,
					APY:            apy,
					Duration:       7, // 7天期
					MinAmount:      100,
					MaxAmount:      10000,
					SettlementTime: time.Now().Add(7 * 24 * time.Hour),
					ProductID:      fmt.Sprintf("%s_%s_%d_%d", symbol, direction, int(strikePrice), 7),
					Status:         "active",
					BaseAsset:      baseAsset,
					QuoteAsset:     quoteAsset,
					CurrentPrice:   currentPrice,
				}

				// 使用 ProductID 作为唯一标识
				if err := cfg.DB.Where("product_id = ?", product.ProductID).
					Assign(product).
					FirstOrCreate(&product).Error; err != nil {
					log.Printf("保存产品失败: %v", err)
				}
			}
		}
	}

	// 清理过期产品
	if err := cfg.DB.Model(&models.DualInvestmentProduct{}).
		Where("settlement_time < ? AND status = ?", time.Now(), "active").
		Update("status", "expired").Error; err != nil {
		log.Printf("更新过期产品失败: %v", err)
	}

	log.Println("双币投资产品同步完成")
}

// executeDualInvestmentStrategies 执行双币投资策略
func executeDualInvestmentStrategies(cfg *config.Config) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// 查询所有启用的策略
		var strategies []models.DualInvestmentStrategy
		if err := cfg.DB.Where("enabled = ? AND status = ?", true, "active").
			Find(&strategies).Error; err != nil {
			log.Printf("查询双币投资策略失败: %v", err)
			continue
		}

		for _, strategy := range strategies {
			go executeStrategy(cfg, strategy)
		}
	}
}

// executeStrategy 执行单个策略
func executeStrategy(cfg *config.Config, strategy models.DualInvestmentStrategy) {
	// 获取用户信息
	var user models.User
	if err := cfg.DB.First(&user, strategy.UserID).Error; err != nil {
		log.Printf("策略用户未找到: %v", err)
		return
	}

	if user.APIKey == "" || user.SecretKey == "" {
		return
	}

	// 检查投资限额
	if strategy.CurrentInvested >= strategy.TotalInvestmentLimit {
		log.Printf("策略 %d 已达到投资限额", strategy.ID)
		return
	}

	symbol := strategy.BaseAsset + strategy.QuoteAsset

	switch strategy.StrategyType {
	case "single":
		executeSingleStrategy(cfg, strategy, user, symbol)
	case "auto_reinvest":
		executeAutoReinvestStrategy(cfg, strategy, user, symbol)
	case "ladder":
		executeLadderStrategy(cfg, strategy, user, symbol)
	case "price_trigger":
		executePriceTriggerStrategy(cfg, strategy, user, symbol)
	}
}

// executeSingleStrategy 执行单次投资策略
func executeSingleStrategy(cfg *config.Config, strategy models.DualInvestmentStrategy, user models.User, symbol string) {
	// 查找符合条件的产品
	product := findBestProduct(cfg, strategy, symbol)
	if product == nil {
		return
	}

	// 计算投资金额
	investAmount := calculateInvestAmount(strategy, product)
	if investAmount <= 0 {
		return
	}

	// 创建订单
	createDualInvestmentOrder(cfg, user, strategy, product, investAmount)
}

// executeAutoReinvestStrategy 执行自动复投策略
func executeAutoReinvestStrategy(cfg *config.Config, strategy models.DualInvestmentStrategy, user models.User, symbol string) {
	// 检查是否有已结算的订单需要复投
	var settledOrders []models.DualInvestmentOrder
	if err := cfg.DB.Where("strategy_id = ? AND status = ? AND created_at > ?",
		strategy.ID, "settled", time.Now().Add(-24*time.Hour)).
		Find(&settledOrders).Error; err != nil {
		return
	}

	// 对每个已结算订单进行复投
	for _, order := range settledOrders {
		// 查找类似的产品
		product := findBestProduct(cfg, strategy, symbol)
		if product == nil {
			continue
		}

		// 使用结算金额进行复投
		investAmount := order.SettlementAmount
		if investAmount > strategy.MaxSingleAmount {
			investAmount = strategy.MaxSingleAmount
		}

		createDualInvestmentOrder(cfg, user, strategy, product, investAmount)
	}
}

// executeLadderStrategy 执行梯度投资策略
func executeLadderStrategy(cfg *config.Config, strategy models.DualInvestmentStrategy, user models.User, symbol string) {
	// 获取当前价格
	client := binance.NewClient(user.APIKey, user.SecretKey)
	prices, err := client.NewListPricesService().Symbol(symbol).Do(context.Background())
	if err != nil || len(prices) == 0 {
		return
	}

	currentPrice, _ := strconv.ParseFloat(prices[0].Price, 64)

	// 检查是否已经在当前价格区间有订单
	var existingOrders int64
	cfg.DB.Model(&models.DualInvestmentOrder{}).
		Where("strategy_id = ? AND status IN ? AND created_at > ?",
			strategy.ID, []string{"pending", "active"}, time.Now().Add(-24*time.Hour)).
		Count(&existingOrders)

	if existingOrders >= int64(strategy.LadderSteps) {
		return // 已经有足够的梯度订单
	}

	// 计算每层投资金额
	totalAvailable := strategy.TotalInvestmentLimit - strategy.CurrentInvested
	if totalAvailable <= 0 {
		return
	}

	amountPerStep := totalAvailable / float64(strategy.LadderSteps-int(existingOrders))
	if amountPerStep > strategy.MaxSingleAmount {
		amountPerStep = strategy.MaxSingleAmount
	}

	// 为每个梯度层级寻找产品
	for i := 0; i < strategy.LadderSteps-int(existingOrders); i++ {
		// 计算目标执行价格
		priceOffset := float64(i) * strategy.LadderStepPercent / 100
		var targetStrikePrice float64

		if strategy.DirectionPreference == "UP" || strategy.DirectionPreference == "BOTH" {
			targetStrikePrice = currentPrice * (1 + priceOffset)
			product := findProductByStrikePrice(cfg, symbol, "UP", targetStrikePrice, strategy)
			if product != nil {
				createDualInvestmentOrder(cfg, user, strategy, product, amountPerStep)
			}
		}

		if strategy.DirectionPreference == "DOWN" || strategy.DirectionPreference == "BOTH" {
			targetStrikePrice = currentPrice * (1 - priceOffset)
			product := findProductByStrikePrice(cfg, symbol, "DOWN", targetStrikePrice, strategy)
			if product != nil {
				createDualInvestmentOrder(cfg, user, strategy, product, amountPerStep)
			}
		}
	}
}

// executePriceTriggerStrategy 执行价格触发策略
func executePriceTriggerStrategy(cfg *config.Config, strategy models.DualInvestmentStrategy, user models.User, symbol string) {
	// 获取当前价格
	client := binance.NewClient(user.APIKey, user.SecretKey)
	prices, err := client.NewListPricesService().Symbol(symbol).Do(context.Background())
	if err != nil || len(prices) == 0 {
		return
	}

	currentPrice, _ := strconv.ParseFloat(prices[0].Price, 64)

	// 检查是否触发
	triggered := false
	if strategy.TriggerType == "above" && currentPrice >= strategy.TriggerPrice {
		triggered = true
	} else if strategy.TriggerType == "below" && currentPrice <= strategy.TriggerPrice {
		triggered = true
	}

	if !triggered {
		return
	}

	// 触发后执行投资
	product := findBestProduct(cfg, strategy, symbol)
	if product == nil {
		return
	}

	investAmount := calculateInvestAmount(strategy, product)
	if investAmount <= 0 {
		return
	}

	// 创建订单
	if createDualInvestmentOrder(cfg, user, strategy, product, investAmount) {
		// 更新策略状态为已完成
		cfg.DB.Model(&strategy).Updates(map[string]interface{}{
			"status":           "completed",
			"last_executed_at": time.Now(),
		})
	}
}

// findBestProduct 查找最佳产品
func findBestProduct(cfg *config.Config, strategy models.DualInvestmentStrategy, symbol string) *models.DualInvestmentProduct {
	query := cfg.DB.Model(&models.DualInvestmentProduct{}).
		Where("symbol = ? AND status = ?", symbol, "active")

	// 方向筛选
	if strategy.DirectionPreference != "BOTH" {
		query = query.Where("direction = ?", strategy.DirectionPreference)
	}

	// APY筛选
	if strategy.TargetAPYMin > 0 {
		query = query.Where("apy >= ?", strategy.TargetAPYMin)
	}
	if strategy.TargetAPYMax > 0 {
		query = query.Where("apy <= ?", strategy.TargetAPYMax)
	}

	// 期限筛选
	if strategy.MinDuration > 0 {
		query = query.Where("duration >= ?", strategy.MinDuration)
	}
	if strategy.MaxDuration > 0 {
		query = query.Where("duration <= ?", strategy.MaxDuration)
	}

	// 执行价格偏离度筛选
	if strategy.MaxStrikePriceOffset > 0 {
		query = query.Where("ABS((strike_price - current_price) / current_price * 100) <= ?",
			strategy.MaxStrikePriceOffset)
	}

	var product models.DualInvestmentProduct
	if err := query.Order("apy desc").First(&product).Error; err != nil {
		return nil
	}

	return &product
}

// findProductByStrikePrice 根据执行价格查找产品
func findProductByStrikePrice(cfg *config.Config, symbol, direction string, targetPrice float64,
	strategy models.DualInvestmentStrategy) *models.DualInvestmentProduct {

	var product models.DualInvestmentProduct
	err := cfg.DB.Where("symbol = ? AND direction = ? AND status = ?", symbol, direction, "active").
		Where("ABS(strike_price - ?) / ? < 0.01", targetPrice, targetPrice). // 1%的误差范围
		Where("apy >= ?", strategy.TargetAPYMin).
		Order("apy desc").
		First(&product).Error

	if err != nil {
		return nil
	}

	return &product
}

// calculateInvestAmount 计算投资金额
func calculateInvestAmount(strategy models.DualInvestmentStrategy, product *models.DualInvestmentProduct) float64 {
	// 可用额度
	available := strategy.TotalInvestmentLimit - strategy.CurrentInvested
	if available <= 0 {
		return 0
	}

	// 单笔限额
	amount := strategy.MaxSingleAmount
	if amount > available {
		amount = available
	}

	// 产品限额
	if amount > product.MaxAmount {
		amount = product.MaxAmount
	}
	if amount < product.MinAmount {
		return 0 // 低于最小限额
	}

	return amount
}

// createDualInvestmentOrder 创建双币投资订单
func createDualInvestmentOrder(cfg *config.Config, user models.User, strategy models.DualInvestmentStrategy,
	product *models.DualInvestmentProduct, investAmount float64) bool {

	// TODO: 调用币安API创建实际订单
	// client := binance.NewClient(user.APIKey, user.SecretKey)
	// 实际的API调用...

	// 创建订单记录
	order := models.DualInvestmentOrder{
		UserID:         user.ID,
		StrategyID:     &strategy.ID,
		ProductID:      product.ID,
		OrderID:        fmt.Sprintf("DUAL_%d_%d", strategy.ID, time.Now().Unix()),
		Symbol:         product.Symbol,
		InvestAsset:    product.BaseAsset, // 简化处理
		InvestAmount:   investAmount,
		StrikePrice:    product.StrikePrice,
		APY:            product.APY,
		Direction:      product.Direction,
		Duration:       product.Duration,
		SettlementTime: product.SettlementTime,
		Status:         "active",
	}

	// 使用事务
	err := cfg.DB.Transaction(func(tx *gorm.DB) error {
		// 创建订单
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		// 更新策略已投资金额
		if err := tx.Model(&strategy).Updates(map[string]interface{}{
			"current_invested": gorm.Expr("current_invested + ?", investAmount),
			"last_executed_at": time.Now(),
		}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Printf("创建双币投资订单失败: %v", err)
		return false
	}

	log.Printf("创建双币投资订单成功: 策略=%d, 产品=%s, 金额=%.2f",
		strategy.ID, product.Symbol, investAmount)
	return true
}

// monitorDualInvestmentSettlement 监控双币投资结算
func monitorDualInvestmentSettlement(cfg *config.Config) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// 查询即将到期的订单
		var orders []models.DualInvestmentOrder
		if err := cfg.DB.Where("status = ? AND settlement_time <= ?",
			"active", time.Now().Add(1*time.Hour)).
			Find(&orders).Error; err != nil {
			log.Printf("查询待结算订单失败: %v", err)
			continue
		}

		for _, order := range orders {
			go settleOrder(cfg, order)
		}
	}
}

// settleOrder 结算订单
func settleOrder(cfg *config.Config, order models.DualInvestmentOrder) {
	// 获取用户信息
	var user models.User
	if err := cfg.DB.First(&user, order.UserID).Error; err != nil {
		return
	}

	if user.APIKey == "" || user.SecretKey == "" {
		return
	}

	// TODO: 调用币安API获取实际结算结果
	// 这里使用模拟结算

	// 获取当前价格
	client := binance.NewClient(user.APIKey, user.SecretKey)
	prices, err := client.NewListPricesService().Symbol(order.Symbol).Do(context.Background())
	if err != nil || len(prices) == 0 {
		return
	}

	currentPrice, _ := strconv.ParseFloat(prices[0].Price, 64)

	// 模拟结算逻辑
	var settlementAsset string
	var settlementAmount float64
	touched := false

	if order.Direction == "UP" && currentPrice >= order.StrikePrice {
		touched = true
		settlementAsset = "USDT"
		settlementAmount = order.InvestAmount * order.StrikePrice
	} else if order.Direction == "DOWN" && currentPrice <= order.StrikePrice {
		touched = true
		settlementAsset = order.InvestAsset
		settlementAmount = order.InvestAmount / order.StrikePrice
	} else {
		// 未触及执行价，返还本金+利息
		settlementAsset = order.InvestAsset
		interestRate := order.APY / 100.0 / 365.0 * float64(order.Duration)
		settlementAmount = order.InvestAmount * (1 + interestRate)
	}

	// 计算盈亏
	var pnl, pnlPercent float64
	if settlementAsset == order.InvestAsset {
		pnl = settlementAmount - order.InvestAmount
		pnlPercent = pnl / order.InvestAmount * 100
	} else {
		// 需要转换为同一币种计算，这里简化处理
		if touched && order.Direction == "UP" {
			// 卖出获得USDT
			pnl = settlementAmount - order.InvestAmount*currentPrice
			pnlPercent = pnl / (order.InvestAmount * currentPrice) * 100
		}
	}

	// 更新订单
	updates := map[string]interface{}{
		"status":            "settled",
		"settlement_asset":  settlementAsset,
		"settlement_amount": settlementAmount,
		"actual_apy":        order.APY, // 简化处理
		"settled_at":        time.Now(),
		"pn_l":              pnl,
		"pn_l_percent":      pnlPercent,
	}

	err = cfg.DB.Transaction(func(tx *gorm.DB) error {
		// 更新订单
		if err := tx.Model(&order).Updates(updates).Error; err != nil {
			return err
		}

		// 如果有关联策略，更新策略的已投资金额
		if order.StrategyID != nil {
			if err := tx.Model(&models.DualInvestmentStrategy{}).
				Where("id = ?", *order.StrategyID).
				Update("current_invested", gorm.Expr("current_invested - ?", order.InvestAmount)).
				Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("结算订单失败: %v", err)
		return
	}

	log.Printf("订单 %s 结算完成: %s %.4f -> %s %.4f",
		order.OrderID, order.InvestAsset, order.InvestAmount,
		settlementAsset, settlementAmount)
}
