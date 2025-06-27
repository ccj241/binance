package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"context"
	"github.com/adshao/go-binance/v2"
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DualInvestmentController struct {
	Config *config.Config
}

// GetProducts 获取可投资的双币产品列表
func (ctrl *DualInvestmentController) GetProducts(c *gin.Context) {
	// 获取查询参数
	symbol := c.Query("symbol")
	direction := c.Query("direction")
	minAPY := c.Query("minApy")

	// 简化日志输出
	log.Printf("获取双币产品: symbol=%s, direction=%s, minAPY=%s", symbol, direction, minAPY)

	query := ctrl.Config.DB.Model(&models.DualInvestmentProduct{}).
		Where("status = ?", "active")

	if symbol != "" {
		query = query.Where("symbol = ?", symbol)
	}
	if direction != "" {
		query = query.Where("direction = ?", direction)
	}
	if minAPY != "" {
		if apy, err := strconv.ParseFloat(minAPY, 64); err == nil {
			query = query.Where("apy >= ?", apy)
		}
	}

	var products []models.DualInvestmentProduct
	if err := query.Order("apy desc").Find(&products).Error; err != nil {
		log.Printf("获取双币产品失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取产品列表失败"})
		return
	}

	// 只在调试模式下打印详细信息
	if symbol == "SOLUSDT" && len(products) > 0 {
		log.Printf("SOLUSDT产品数量: %d, 最高APY: %.2f%%, 最低APY: %.2f%%",
			len(products),
			products[0].APY,
			products[len(products)-1].APY)
	}

	c.JSON(http.StatusOK, gin.H{"products": products})
}

// CreateStrategy 创建双币投资策略 - 修复版本
func (ctrl *DualInvestmentController) CreateStrategy(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req struct {
		StrategyName         string  `json:"strategyName" binding:"required"`
		StrategyType         string  `json:"strategyType" binding:"required,oneof=single auto_reinvest ladder price_trigger"`
		BaseAsset            string  `json:"baseAsset" binding:"required"`
		QuoteAsset           string  `json:"quoteAsset" binding:"required"`
		DirectionPreference  string  `json:"directionPreference" binding:"required,oneof=UP DOWN BOTH"`
		TargetAPYMin         float64 `json:"targetApyMin" binding:"min=0"`
		TargetAPYMax         float64 `json:"targetApyMax" binding:"min=0"`
		MaxSingleAmount      float64 `json:"maxSingleAmount" binding:"required,gt=0"`
		TotalInvestmentLimit float64 `json:"totalInvestmentLimit" binding:"required,gt=0"`
		MaxStrikePriceOffset float64 `json:"maxStrikePriceOffset" binding:"min=0,max=100"`
		MinDuration          int     `json:"minDuration" binding:"min=1"`
		MaxDuration          int     `json:"maxDuration" binding:"min=1"`
		MaxPositionRatio     float64 `json:"maxPositionRatio" binding:"min=0,max=100"`
		AutoReinvest         bool    `json:"autoReinvest"`
		BasePrice            float64 `json:"basePrice"` // 基准价格
		// 价格触发策略参数
		TriggerPrice float64 `json:"triggerPrice"`
		TriggerType  string  `json:"triggerType"`
		// 梯度策略参数 - 直接接收数组
		LadderConfig []models.LadderConfigItem `json:"ladderConfig"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据", "details": err.Error()})
		return
	}

	// 验证逻辑
	if req.TargetAPYMax > 0 && req.TargetAPYMax < req.TargetAPYMin {
		c.JSON(http.StatusBadRequest, gin.H{"error": "最大年化收益率不能小于最小年化收益率"})
		return
	}

	if req.MaxDuration > 0 && req.MaxDuration < req.MinDuration {
		c.JSON(http.StatusBadRequest, gin.H{"error": "最大期限不能小于最小期限"})
		return
	}

	// 价格触发策略验证
	if req.StrategyType == "price_trigger" {
		if req.TriggerPrice <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "价格触发策略需要设置触发价格"})
			return
		}
		if req.TriggerType != "above" && req.TriggerType != "below" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "触发类型必须是 above 或 below"})
			return
		}
	}

	// 梯度策略验证
	if req.StrategyType == "ladder" {
		if len(req.LadderConfig) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "梯度策略需要配置梯度参数"})
			return
		}

		// 验证梯度配置
		totalPercentage := 0.0
		for _, config := range req.LadderConfig {
			if config.MinDepth <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "深度级别必须大于0"})
				return
			}
			if config.Percentage <= 0 || config.Percentage > 100 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "投资百分比必须在0-100之间"})
				return
			}
			totalPercentage += config.Percentage
		}

		if totalPercentage > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "总投资百分比不能超过100%"})
			return
		}

		// 梯度策略必须设置基准价格
		if req.BasePrice <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "梯度策略必须设置基准价格"})
			return
		}
	}

	// 将梯度配置转换为JSON字符串存储
	var ladderConfigJSON string
	if req.StrategyType == "ladder" && len(req.LadderConfig) > 0 {
		configBytes, err := json.Marshal(req.LadderConfig)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "处理梯度配置失败"})
			return
		}
		ladderConfigJSON = string(configBytes)
	}

	// 创建策略
	strategy := models.DualInvestmentStrategy{
		UserID:               userID.(uint),
		StrategyName:         req.StrategyName,
		StrategyType:         req.StrategyType,
		BaseAsset:            req.BaseAsset,
		QuoteAsset:           req.QuoteAsset,
		DirectionPreference:  req.DirectionPreference,
		TargetAPYMin:         req.TargetAPYMin,
		TargetAPYMax:         req.TargetAPYMax,
		MaxSingleAmount:      req.MaxSingleAmount,
		TotalInvestmentLimit: req.TotalInvestmentLimit,
		MaxStrikePriceOffset: req.MaxStrikePriceOffset,
		MinDuration:          req.MinDuration,
		MaxDuration:          req.MaxDuration,
		MaxPositionRatio:     req.MaxPositionRatio,
		AutoReinvest:         req.AutoReinvest,
		TriggerPrice:         req.TriggerPrice,
		TriggerType:          req.TriggerType,
		LadderConfig:         ladderConfigJSON, // 存储JSON字符串
		BasePrice:            req.BasePrice,
		Enabled:              true,
		Status:               "active",
	}

	if err := ctrl.Config.DB.Create(&strategy).Error; err != nil {
		log.Printf("创建双币投资策略失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建策略失败"})
		return
	}

	log.Printf("用户 %d 创建双币投资策略: %s", userID.(uint), strategy.StrategyName)
	c.JSON(http.StatusOK, gin.H{
		"message":  "策略创建成功",
		"strategy": strategy,
	})
}

// GetStrategies 获取用户的双币投资策略列表 - 修复版本
func (ctrl *DualInvestmentController) GetStrategies(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var strategies []models.DualInvestmentStrategy
	if err := ctrl.Config.DB.Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at desc").
		Find(&strategies).Error; err != nil {
		log.Printf("获取双币投资策略失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取策略列表失败"})
		return
	}

	// 格式化返回的策略数据
	type StrategyResponse struct {
		models.DualInvestmentStrategy
		ParsedLadderConfig []models.LadderConfigItem `json:"parsedLadderConfig,omitempty"`
	}

	responseStrategies := make([]StrategyResponse, 0, len(strategies))
	for _, strategy := range strategies {
		resp := StrategyResponse{
			DualInvestmentStrategy: strategy,
		}

		// 解析梯度配置
		if strategy.StrategyType == "ladder" && strategy.LadderConfig != "" {
			var config []models.LadderConfigItem
			if err := json.Unmarshal([]byte(strategy.LadderConfig), &config); err == nil {
				resp.ParsedLadderConfig = config
			}
		}

		responseStrategies = append(responseStrategies, resp)
	}

	c.JSON(http.StatusOK, gin.H{"strategies": responseStrategies})
}

// UpdateStrategy 更新策略 - 修复版本（添加权限验证）
func (ctrl *DualInvestmentController) UpdateStrategy(c *gin.Context) {
	userID, _ := c.Get("user_id")
	strategyID := c.Param("id")

	var strategy models.DualInvestmentStrategy
	// 添加用户权限验证
	if err := ctrl.Config.DB.Where("id = ? AND user_id = ? AND deleted_at IS NULL", strategyID, userID).
		First(&strategy).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "策略未找到或无权访问"})
		return
	}

	var req struct {
		Enabled              *bool                     `json:"enabled"`
		TargetAPYMin         *float64                  `json:"targetApyMin"`
		TargetAPYMax         *float64                  `json:"targetApyMax"`
		MaxSingleAmount      *float64                  `json:"maxSingleAmount"`
		TotalInvestmentLimit *float64                  `json:"totalInvestmentLimit"`
		AutoReinvest         *bool                     `json:"autoReinvest"`
		BasePrice            *float64                  `json:"basePrice"`
		LadderConfig         []models.LadderConfigItem `json:"ladderConfig"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.TargetAPYMin != nil {
		updates["target_apy_min"] = *req.TargetAPYMin
	}
	if req.TargetAPYMax != nil {
		updates["target_apy_max"] = *req.TargetAPYMax
	}
	if req.MaxSingleAmount != nil {
		updates["max_single_amount"] = *req.MaxSingleAmount
	}
	if req.TotalInvestmentLimit != nil {
		updates["total_investment_limit"] = *req.TotalInvestmentLimit
	}
	if req.AutoReinvest != nil {
		updates["auto_reinvest"] = *req.AutoReinvest
	}
	if req.BasePrice != nil {
		updates["base_price"] = *req.BasePrice
	}

	// 更新梯度配置
	if strategy.StrategyType == "ladder" && len(req.LadderConfig) > 0 {
		// 验证梯度配置
		totalPercentage := 0.0
		for _, config := range req.LadderConfig {
			if config.MinDepth <= 0 || config.Percentage <= 0 || config.Percentage > 100 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "无效的梯度配置"})
				return
			}
			totalPercentage += config.Percentage
		}

		if totalPercentage > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "总投资百分比不能超过100%"})
			return
		}

		configBytes, err := json.Marshal(req.LadderConfig)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "处理梯度配置失败"})
			return
		}
		updates["ladder_config"] = string(configBytes)
	}

	if err := ctrl.Config.DB.Model(&strategy).Updates(updates).Error; err != nil {
		log.Printf("更新策略失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新策略失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "策略更新成功"})
}

// DeleteStrategy 删除策略（添加权限验证）
func (ctrl *DualInvestmentController) DeleteStrategy(c *gin.Context) {
	userID, _ := c.Get("user_id")
	strategyID := c.Param("id")

	var strategy models.DualInvestmentStrategy
	// 添加用户权限验证
	if err := ctrl.Config.DB.Where("id = ? AND user_id = ? AND deleted_at IS NULL", strategyID, userID).
		First(&strategy).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "策略未找到或无权访问"})
		return
	}

	// 检查是否有活跃订单
	var activeOrders int64
	ctrl.Config.DB.Model(&models.DualInvestmentOrder{}).
		Where("strategy_id = ? AND status IN ?", strategyID, []string{"pending", "active"}).
		Count(&activeOrders)

	if activeOrders > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "策略有活跃订单，无法删除"})
		return
	}

	if err := ctrl.Config.DB.Delete(&strategy).Error; err != nil {
		log.Printf("删除策略失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除策略失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "策略删除成功"})
}

// CreateOrder 创建双币投资订单
func (ctrl *DualInvestmentController) CreateOrder(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req struct {
		ProductID    uint    `json:"productId" binding:"required"`
		InvestAmount float64 `json:"investAmount" binding:"required,gt=0"`
		StrategyID   *uint   `json:"strategyId"` // 可选，手动下单时为空
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 获取用户信息
	var user models.User
	if err := ctrl.Config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
		return
	}

	if user.APIKey == "" || user.SecretKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请先设置API密钥"})
		return
	}

	// 获取产品信息
	var product models.DualInvestmentProduct
	if err := ctrl.Config.DB.First(&product, req.ProductID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "产品未找到"})
		return
	}

	// 验证投资金额
	if req.InvestAmount < product.MinAmount {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("投资金额不能小于 %.2f", product.MinAmount),
		})
		return
	}
	if req.InvestAmount > product.MaxAmount {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("投资金额不能大于 %.2f", product.MaxAmount),
		})
		return
	}

	// 如果关联策略，验证策略
	if req.StrategyID != nil {
		var strategy models.DualInvestmentStrategy
		if err := ctrl.Config.DB.Where("id = ? AND user_id = ?", *req.StrategyID, userID).
			First(&strategy).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "策略未找到"})
			return
		}

		// 检查策略限额
		if strategy.CurrentInvested+req.InvestAmount > strategy.TotalInvestmentLimit {
			c.JSON(http.StatusBadRequest, gin.H{"error": "超出策略总投资限额"})
			return
		}
	}

	// TODO: 调用币安API创建订单
	// 这里需要根据币安实际的双币投资API进行调整
	/*
		client := binance.NewClient(user.APIKey, user.SecretKey)
		// 调用双币投资下单接口
	*/

	// 创建订单记录
	order := models.DualInvestmentOrder{
		UserID:         userID.(uint),
		StrategyID:     req.StrategyID,
		ProductID:      req.ProductID,
		OrderID:        fmt.Sprintf("DUAL_%d_%d", userID, time.Now().Unix()), // 临时订单号
		Symbol:         product.Symbol,
		InvestAsset:    product.BaseAsset, // 根据产品类型确定
		InvestAmount:   req.InvestAmount,
		StrikePrice:    product.StrikePrice,
		APY:            product.APY,
		Direction:      product.Direction,
		Duration:       product.Duration,
		SettlementTime: product.SettlementTime,
		Status:         "active",
	}

	// 开启事务
	err := ctrl.Config.DB.Transaction(func(tx *gorm.DB) error {
		// 创建订单
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		// 更新策略已投资金额
		if req.StrategyID != nil {
			if err := tx.Model(&models.DualInvestmentStrategy{}).
				Where("id = ?", *req.StrategyID).
				Update("current_invested", gorm.Expr("current_invested + ?", req.InvestAmount)).
				Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("创建双币投资订单失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建订单失败"})
		return
	}

	log.Printf("用户 %d 创建双币投资订单: %s, 金额: %.2f", userID, order.Symbol, order.InvestAmount)
	c.JSON(http.StatusOK, gin.H{
		"message": "订单创建成功",
		"order":   order,
	})
}

// GetOrders 获取用户的双币投资订单（添加策略权限验证）
func (ctrl *DualInvestmentController) GetOrders(c *gin.Context) {
	userID, _ := c.Get("user_id")
	status := c.Query("status")
	strategyID := c.Query("strategyId")

	query := ctrl.Config.DB.Model(&models.DualInvestmentOrder{}).
		Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 如果指定了策略ID，验证策略所有权
	if strategyID != "" {
		// 先验证策略是否属于当前用户
		var strategy models.DualInvestmentStrategy
		if err := ctrl.Config.DB.Where("id = ? AND user_id = ?", strategyID, userID).
			First(&strategy).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "策略未找到或无权访问"})
			return
		}
		query = query.Where("strategy_id = ?", strategyID)
	}

	var orders []models.DualInvestmentOrder
	if err := query.Order("created_at desc").Find(&orders).Error; err != nil {
		log.Printf("获取双币投资订单失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取订单列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

// GetStats 获取双币投资统计信息 - 使用币安API
func (ctrl *DualInvestmentController) GetStats(c *gin.Context) {
	userID, _ := c.Get("user_id")

	// 获取用户信息
	var user models.User
	if err := ctrl.Config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
		return
	}

	// 如果用户没有设置API密钥，返回空统计
	if user.APIKey == "" || user.SecretKey == "" {
		c.JSON(http.StatusOK, gin.H{
			"stats": models.DualInvestmentStats{
				UserID: userID.(uint),
			},
		})
		return
	}

	// 从币安获取双币投资统计数据
	client := binance.NewClient(user.APIKey, user.SecretKey)

	// 获取账户总览信息
	account, err := client.NewGetAccountService().Do(context.Background())
	if err != nil {
		log.Printf("获取币安账户信息失败: %v", err)
		// 如果API失败，使用本地数据
		stats := ctrl.getLocalStats(userID.(uint))
		c.JSON(http.StatusOK, gin.H{"stats": stats})
		return
	}

	// 计算总资产价值（以USDT计价）
	totalAssetValue := 0.0
	for _, balance := range account.Balances {
		free, _ := strconv.ParseFloat(balance.Free, 64)
		locked, _ := strconv.ParseFloat(balance.Locked, 64)
		total := free + locked

		if total > 0 {
			// 如果是USDT，直接计入
			if balance.Asset == "USDT" {
				totalAssetValue += total
			} else {
				// 其他资产需要转换为USDT价值
				// 这里简化处理，实际应该获取实时价格
				symbol := balance.Asset + "USDT"
				prices, err := client.NewListPricesService().Symbol(symbol).Do(context.Background())
				if err == nil && len(prices) > 0 {
					price, _ := strconv.ParseFloat(prices[0].Price, 64)
					totalAssetValue += total * price
				}
			}
		}
	}

	// TODO: 调用币安的双币投资API获取实际统计数据
	// 由于币安API文档中双币投资的接口不是公开的，这里使用本地数据和账户余额结合

	// 获取本地统计数据
	localStats := ctrl.getLocalStats(userID.(uint))

	// 结合币安账户数据和本地数据
	stats := models.DualInvestmentStats{
		UserID:          userID.(uint),
		TotalInvested:   localStats.TotalInvested,
		TotalSettled:    localStats.TotalSettled,
		TotalPnL:        localStats.TotalPnL,
		TotalPnLPercent: localStats.TotalPnLPercent,
		WinCount:        localStats.WinCount,
		LossCount:       localStats.LossCount,
		WinRate:         localStats.WinRate,
		AverageAPY:      localStats.AverageAPY,
		ActiveOrders:    localStats.ActiveOrders,
		ActiveAmount:    localStats.ActiveAmount,
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// getLocalStats 获取本地统计数据
func (ctrl *DualInvestmentController) getLocalStats(userID uint) models.DualInvestmentStats {
	stats := models.DualInvestmentStats{
		UserID: userID,
	}

	// 获取总投资和结算信息
	ctrl.Config.DB.Model(&models.DualInvestmentOrder{}).
		Where("user_id = ? AND status = ?", userID, "settled").
		Select("COALESCE(SUM(invest_amount), 0) as total_invested, " +
			"COALESCE(SUM(settlement_amount), 0) as total_settled, " +
			"COALESCE(SUM(pn_l), 0) as total_pn_l").
		Scan(&stats)

	// 计算总盈亏百分比
	if stats.TotalInvested > 0 {
		stats.TotalPnLPercent = (stats.TotalPnL / stats.TotalInvested) * 100
	}

	// 获取盈亏统计
	var winCount int64
	ctrl.Config.DB.Model(&models.DualInvestmentOrder{}).
		Where("user_id = ? AND status = ? AND pn_l > 0", userID, "settled").
		Count(&winCount)
	stats.WinCount = int(winCount)

	var lossCount int64
	ctrl.Config.DB.Model(&models.DualInvestmentOrder{}).
		Where("user_id = ? AND status = ? AND pn_l < 0", userID, "settled").
		Count(&lossCount)
	stats.LossCount = int(lossCount)

	// 计算胜率
	totalSettled := int64(stats.WinCount + stats.LossCount)
	if totalSettled > 0 {
		stats.WinRate = float64(stats.WinCount) / float64(totalSettled) * 100
	}

	// 获取平均年化收益率
	var avgAPY float64
	ctrl.Config.DB.Model(&models.DualInvestmentOrder{}).
		Where("user_id = ? AND status = ?", userID, "settled").
		Select("COALESCE(AVG(actual_apy), 0)").
		Scan(&avgAPY)
	stats.AverageAPY = avgAPY

	// 获取活跃订单信息
	var activeStats struct {
		ActiveOrders int64
		ActiveAmount float64
	}
	ctrl.Config.DB.Model(&models.DualInvestmentOrder{}).
		Where("user_id = ? AND status = ?", userID, "active").
		Select("COUNT(*) as active_orders, COALESCE(SUM(invest_amount), 0) as active_amount").
		Scan(&activeStats)

	stats.ActiveOrders = int(activeStats.ActiveOrders)
	stats.ActiveAmount = activeStats.ActiveAmount

	return stats
}

// SimulateInvestment 模拟投资计算
func (ctrl *DualInvestmentController) SimulateInvestment(c *gin.Context) {
	var req struct {
		InvestAmount float64 `json:"investAmount" binding:"required,gt=0"`
		StrikePrice  float64 `json:"strikePrice" binding:"required,gt=0"`
		CurrentPrice float64 `json:"currentPrice" binding:"required,gt=0"`
		APY          float64 `json:"apy" binding:"required,gt=0"`
		Duration     int     `json:"duration" binding:"required,gt=0"`
		Direction    string  `json:"direction" binding:"required,oneof=UP DOWN"`
		InvestAsset  string  `json:"investAsset" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 计算到期收益
	daysInYear := 365.0
	interestRate := req.APY / 100.0 / daysInYear * float64(req.Duration)
	interest := req.InvestAmount * interestRate

	// 模拟不同情况下的结算结果
	results := make(map[string]interface{})

	// 情况1：价格未触及执行价（获得利息）
	results["noTouch"] = map[string]interface{}{
		"settlementAsset":  req.InvestAsset,
		"settlementAmount": req.InvestAmount + interest,
		"profit":           interest,
		"profitPercent":    interestRate * 100,
		"description":      "价格未触及执行价，获得利息收益",
	}

	// 情况2：价格触及执行价（币种转换）
	var convertedAsset string
	var convertedAmount float64

	if req.Direction == "UP" {
		// 看涨：如果价格上涨超过执行价，以执行价卖出基础资产
		convertedAsset = "USDT" // 假设计价资产是USDT
		convertedAmount = req.InvestAmount * req.StrikePrice * (1 + interestRate)
		results["touched"] = map[string]interface{}{
			"settlementAsset":  convertedAsset,
			"settlementAmount": convertedAmount,
			"description":      fmt.Sprintf("价格上涨超过执行价，以 %.2f 的价格卖出", req.StrikePrice),
		}
	} else {
		// 看跌：如果价格下跌低于执行价，以执行价买入基础资产
		convertedAsset = "BTC" // 假设基础资产是BTC
		convertedAmount = (req.InvestAmount / req.StrikePrice) * (1 + interestRate)
		results["touched"] = map[string]interface{}{
			"settlementAsset":  convertedAsset,
			"settlementAmount": convertedAmount,
			"description":      fmt.Sprintf("价格下跌低于执行价，以 %.2f 的价格买入", req.StrikePrice),
		}
	}

	// 风险提示
	results["risks"] = []string{
		"双币投资不保本，可能产生本金损失",
		"结算币种取决于到期时的市场价格",
		"实际收益率可能与预期不符",
		"需要承担汇率风险",
	}

	c.JSON(http.StatusOK, gin.H{"simulation": results})
}
