package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
	"github.com/gin-gonic/gin"
)

type FuturesController struct {
	Config *config.Config
}

// CreateStrategy 创建永续期货策略
func (ctrl *FuturesController) CreateStrategy(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req struct {
		StrategyName   string  `json:"strategyName" binding:"required"`
		Symbol         string  `json:"symbol" binding:"required"`
		Side           string  `json:"side" binding:"required,oneof=LONG SHORT"`
		BasePrice      float64 `json:"basePrice" binding:"required,gt=0"`
		EntryPrice     float64 `json:"entryPrice" binding:"required,gt=0"`
		Leverage       int     `json:"leverage" binding:"required,min=1,max=125"`
		Quantity       float64 `json:"quantity" binding:"required,gt=0"`
		TakeProfitRate float64 `json:"takeProfitRate" binding:"required,gt=0"`
		StopLossRate   float64 `json:"stopLossRate" binding:"min=0"` // 可选
		MarginType     string  `json:"marginType" binding:"omitempty,oneof=ISOLATED CROSSED"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据", "details": err.Error()})
		return
	}

	// 设置默认保证金类型
	if req.MarginType == "" {
		req.MarginType = "CROSSED" // 默认改为全仓
	}

	// 创建策略
	strategy := models.FuturesStrategy{
		UserID:         userID.(uint),
		StrategyName:   req.StrategyName,
		Symbol:         req.Symbol,
		Side:           req.Side,
		BasePrice:      req.BasePrice,
		EntryPrice:     req.EntryPrice,
		Leverage:       req.Leverage,
		Quantity:       req.Quantity,
		TakeProfitRate: req.TakeProfitRate,
		StopLossRate:   req.StopLossRate,
		MarginType:     req.MarginType,
		Enabled:        true,
		Status:         "waiting",
	}

	// 计算止盈止损价格
	strategy.CalculateTakeProfitPrice()
	strategy.CalculateStopLossPrice()

	if err := ctrl.Config.DB.Create(&strategy).Error; err != nil {
		log.Printf("创建永续期货策略失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建策略失败"})
		return
	}

	log.Printf("用户 %d 创建永续期货策略: %s", userID.(uint), strategy.StrategyName)
	c.JSON(http.StatusOK, gin.H{
		"message":  "策略创建成功",
		"strategy": strategy,
	})
}

// GetStrategies 获取用户的永续期货策略列表
func (ctrl *FuturesController) GetStrategies(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var strategies []models.FuturesStrategy
	if err := ctrl.Config.DB.Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at desc").
		Find(&strategies).Error; err != nil {
		log.Printf("获取永续期货策略失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取策略列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"strategies": strategies})
}

// UpdateStrategy 更新策略
func (ctrl *FuturesController) UpdateStrategy(c *gin.Context) {
	userID, _ := c.Get("user_id")
	strategyID := c.Param("id")

	var strategy models.FuturesStrategy
	// 验证策略所有权
	if err := ctrl.Config.DB.Where("id = ? AND user_id = ? AND deleted_at IS NULL", strategyID, userID).
		First(&strategy).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "策略未找到或无权访问"})
		return
	}

	// 检查策略状态
	if strategy.Status != "waiting" && strategy.Status != "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "只能修改等待中或已取消的策略"})
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
		return
	}

	// 允许更新的字段
	allowedFields := map[string]bool{
		"enabled":        true,
		"basePrice":      true,
		"entryPrice":     true,
		"quantity":       true,
		"takeProfitRate": true,
		"stopLossRate":   true,
	}

	updates := make(map[string]interface{})
	for field, value := range updateData {
		if allowedFields[field] {
			updates[field] = value
		}
	}

	// 如果更新了价格相关字段，重新计算止盈止损价格
	needRecalculate := false
	if _, ok := updates["entryPrice"]; ok {
		needRecalculate = true
	}
	if _, ok := updates["takeProfitRate"]; ok {
		needRecalculate = true
	}
	if _, ok := updates["stopLossRate"]; ok {
		needRecalculate = true
	}

	if needRecalculate {
		// 先更新字段
		if err := ctrl.Config.DB.Model(&strategy).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新策略失败"})
			return
		}

		// 重新加载策略
		ctrl.Config.DB.First(&strategy, strategyID)

		// 重新计算价格
		strategy.CalculateTakeProfitPrice()
		strategy.CalculateStopLossPrice()

		// 保存计算后的价格
		ctrl.Config.DB.Model(&strategy).Updates(map[string]interface{}{
			"take_profit_price": strategy.TakeProfitPrice,
			"stop_loss_price":   strategy.StopLossPrice,
			"updated_at":        time.Now(),
		})
	} else {
		// 直接更新
		updates["updated_at"] = time.Now()
		if err := ctrl.Config.DB.Model(&strategy).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新策略失败"})
			return
		}
	}

	// 重新查询更新后的策略
	ctrl.Config.DB.First(&strategy, strategyID)

	c.JSON(http.StatusOK, gin.H{"message": "策略更新成功", "strategy": strategy})
}

// DeleteStrategy 删除策略
func (ctrl *FuturesController) DeleteStrategy(c *gin.Context) {
	userID, _ := c.Get("user_id")
	strategyID := c.Param("id")

	var strategy models.FuturesStrategy
	// 验证策略所有权
	if err := ctrl.Config.DB.Where("id = ? AND user_id = ? AND deleted_at IS NULL", strategyID, userID).
		First(&strategy).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "策略未找到或无权访问"})
		return
	}

	// 如果策略正在执行中，先尝试平仓
	if strategy.Status == "position_opened" && strategy.CurrentPositionId > 0 {
		// 获取用户信息
		var user models.User
		if err := ctrl.Config.DB.First(&user, userID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
			return
		}

		// 尝试平仓
		if err := ctrl.closePosition(user, &strategy); err != nil {
			log.Printf("平仓失败: %v", err)
			// 即使平仓失败也允许删除策略，但要警告用户
			c.JSON(http.StatusOK, gin.H{
				"message": "策略已删除，但平仓失败，请手动检查持仓",
				"warning": err.Error(),
			})
		}
	}

	// 软删除策略
	if err := ctrl.Config.DB.Delete(&strategy).Error; err != nil {
		log.Printf("删除策略失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除策略失败"})
		return
	}

	log.Printf("策略 %s 已删除", strategyID)
	c.JSON(http.StatusOK, gin.H{"message": "策略删除成功"})
}

// GetOrders 获取策略相关订单
func (ctrl *FuturesController) GetOrders(c *gin.Context) {
	userID, _ := c.Get("user_id")
	strategyID := c.Query("strategyId")
	status := c.Query("status")

	query := ctrl.Config.DB.Model(&models.FuturesOrder{}).Where("user_id = ?", userID)

	if strategyID != "" {
		// 验证策略所有权
		var strategy models.FuturesStrategy
		if err := ctrl.Config.DB.Where("id = ? AND user_id = ?", strategyID, userID).
			First(&strategy).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "策略未找到或无权访问"})
			return
		}
		query = query.Where("strategy_id = ?", strategyID)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var orders []models.FuturesOrder
	if err := query.Order("created_at desc").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取订单列表失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

// GetPositions 获取持仓信息
func (ctrl *FuturesController) GetPositions(c *gin.Context) {
	userID, _ := c.Get("user_id")
	status := c.Query("status") // open/closed

	query := ctrl.Config.DB.Model(&models.FuturesPosition{}).Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var positions []models.FuturesPosition
	if err := query.Order("created_at desc").Find(&positions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取持仓列表失败"})
		return
	}

	// 如果查询开仓中的持仓，更新实时数据
	if status == "open" || status == "" {
		ctrl.updatePositionsRealtime(userID.(uint), positions)
	}

	c.JSON(http.StatusOK, gin.H{"positions": positions})
}

// GetStats 获取统计信息
func (ctrl *FuturesController) GetStats(c *gin.Context) {
	userID, _ := c.Get("user_id")

	stats := models.FuturesStats{
		UserID: userID.(uint),
	}

	// 获取总交易次数
	var totalTrades int64
	ctrl.Config.DB.Model(&models.FuturesPosition{}).
		Where("user_id = ? AND status = ?", userID, "closed").
		Count(&totalTrades)
	stats.TotalTrades = int(totalTrades)

	// 获取盈亏统计
	var positions []models.FuturesPosition
	ctrl.Config.DB.Where("user_id = ? AND status = ?", userID, "closed").
		Find(&positions)

	for _, pos := range positions {
		if pos.RealizedPnl > 0 {
			stats.WinTrades++
			if pos.RealizedPnl > stats.MaxWin {
				stats.MaxWin = pos.RealizedPnl
			}
		} else if pos.RealizedPnl < 0 {
			stats.LossTrades++
			if pos.RealizedPnl < stats.MaxLoss {
				stats.MaxLoss = pos.RealizedPnl
			}
		}
		stats.TotalPnl += pos.RealizedPnl
	}

	// 计算胜率
	if stats.TotalTrades > 0 {
		stats.WinRate = float64(stats.WinTrades) / float64(stats.TotalTrades) * 100
		stats.AveragePnl = stats.TotalPnl / float64(stats.TotalTrades)
	}

	// 获取总手续费
	ctrl.Config.DB.Model(&models.FuturesOrder{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(commission), 0)").
		Scan(&stats.TotalCommission)

	// 计算净盈亏
	stats.NetPnl = stats.TotalPnl - stats.TotalCommission

	// 获取活跃持仓数
	var activePositions int64
	ctrl.Config.DB.Model(&models.FuturesPosition{}).
		Where("user_id = ? AND status = ?", userID, "open").
		Count(&activePositions)
	stats.ActivePositions = int(activePositions)

	// 获取活跃策略数
	var activeStrategies int64
	ctrl.Config.DB.Model(&models.FuturesStrategy{}).
		Where("user_id = ? AND enabled = ? AND status IN ? AND deleted_at IS NULL",
			userID, true, []string{"waiting", "triggered", "position_opened"}).
		Count(&activeStrategies)
	stats.ActiveStrategies = int(activeStrategies)

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// closePosition 平仓辅助函数
func (ctrl *FuturesController) closePosition(user models.User, strategy *models.FuturesStrategy) error {
	// 解密API密钥
	apiKey, err := user.GetDecryptedAPIKey()
	if err != nil {
		return fmt.Errorf("解密API Key失败: %v", err)
	}
	secretKey, err := user.GetDecryptedSecretKey()
	if err != nil {
		return fmt.Errorf("解密Secret Key失败: %v", err)
	}

	// 创建期货客户端
	client := binance.NewFuturesClient(apiKey, secretKey)

	// 获取当前持仓
	positions, err := client.NewGetPositionRiskService().
		Symbol(strategy.Symbol).
		Do(context.Background())
	if err != nil {
		return fmt.Errorf("获取持仓信息失败: %v", err)
	}

	// 查找对应的持仓
	var position *futures.PositionRisk
	for _, pos := range positions {
		if string(pos.PositionSide) == strategy.Side {
			position = pos // pos 已经是指针类型
			break
		}
	}

	if position == nil {
		return fmt.Errorf("未找到对应持仓")
	}

	// 获取持仓数量
	positionAmt, _ := strconv.ParseFloat(position.PositionAmt, 64)
	if positionAmt == 0 {
		return nil // 已经没有持仓了
	}

	// 确定平仓方向
	side := futures.SideTypeBuy
	if strategy.Side == "LONG" {
		side = futures.SideTypeSell
	}

	// 创建市价平仓订单
	order, err := client.NewCreateOrderService().
		Symbol(strategy.Symbol).
		Side(side).
		PositionSide(futures.PositionSideType(strategy.Side)).
		Type(futures.OrderTypeMarket).
		Quantity(fmt.Sprintf("%.8f", abs(positionAmt))).
		Do(context.Background())

	if err != nil {
		return fmt.Errorf("创建平仓订单失败: %v", err)
	}

	log.Printf("策略 %d 平仓成功，订单ID: %d", strategy.ID, order.OrderID)

	// 更新策略状态
	strategy.Status = "completed"
	now := time.Now()
	strategy.CompletedAt = &now
	ctrl.Config.DB.Save(strategy)

	return nil
}

// abs 取绝对值
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// updatePositionsRealtime 更新持仓实时数据
func (ctrl *FuturesController) updatePositionsRealtime(userID uint, positions []models.FuturesPosition) {
	// 获取用户信息
	var user models.User
	if err := ctrl.Config.DB.First(&user, userID).Error; err != nil {
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

	if apiKey == "" || secretKey == "" {
		return
	}

	// 创建期货客户端
	client := binance.NewFuturesClient(apiKey, secretKey)

	// 获取所有持仓
	riskPositions, err := client.NewGetPositionRiskService().Do(context.Background())
	if err != nil {
		return
	}

	// 创建映射
	positionMap := make(map[string]*futures.PositionRisk)
	for _, pos := range riskPositions {
		key := pos.Symbol + "_" + string(pos.PositionSide)
		positionMap[key] = pos // pos 已经是指针类型
	}

	// 更新本地持仓数据
	for i := range positions {
		key := positions[i].Symbol + "_" + positions[i].PositionSide
		if riskPos, exists := positionMap[key]; exists {
			// 更新实时数据
			positions[i].UnrealizedPnl, _ = strconv.ParseFloat(riskPos.UnRealizedProfit, 64)
			positions[i].MarkPrice, _ = strconv.ParseFloat(riskPos.MarkPrice, 64)
			positions[i].LiquidationPrice, _ = strconv.ParseFloat(riskPos.LiquidationPrice, 64)
		}
	}
}
