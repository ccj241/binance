package handlers

import (
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// GinStrategyStatsHandler 获取策略统计信息（带权限验证）
func GinStrategyStatsHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		strategyIDStr := c.Param("id")
		strategyID, err := strconv.ParseUint(strategyIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的策略ID"})
			return
		}

		// 查询策略并验证所有权
		var strategy models.Strategy
		if err := cfg.DB.Where("id = ? AND user_id = ? AND deleted_at IS NULL", strategyID, user.ID).
			First(&strategy).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "策略未找到或无权访问"})
			return
		}

		// 统计订单信息
		var stats struct {
			TotalOrders     int64   `json:"totalOrders"`
			PendingOrders   int64   `json:"pendingOrders"`
			FilledOrders    int64   `json:"filledOrders"`
			CancelledOrders int64   `json:"cancelledOrders"`
			TotalVolume     float64 `json:"totalVolume"`
			FilledVolume    float64 `json:"filledVolume"`
		}

		// 总订单数
		cfg.DB.Model(&models.Order{}).
			Where("strategy_id = ?", strategy.ID).
			Count(&stats.TotalOrders)

		// 待处理订单数
		cfg.DB.Model(&models.Order{}).
			Where("strategy_id = ? AND status = ?", strategy.ID, "pending").
			Count(&stats.PendingOrders)

		// 已成交订单数
		cfg.DB.Model(&models.Order{}).
			Where("strategy_id = ? AND status = ?", strategy.ID, "filled").
			Count(&stats.FilledOrders)

		// 已取消订单数
		cfg.DB.Model(&models.Order{}).
			Where("strategy_id = ? AND status IN (?)", strategy.ID, []string{"cancelled", "expired", "rejected"}).
			Count(&stats.CancelledOrders)

		// 计算交易量
		var orders []models.Order
		cfg.DB.Where("strategy_id = ?", strategy.ID).Find(&orders)

		for _, order := range orders {
			volume := order.Price * order.Quantity
			stats.TotalVolume += volume
			if order.Status == "filled" {
				stats.FilledVolume += volume
			}
		}

		// 获取最近的订单
		var recentOrders []models.Order
		cfg.DB.Where("strategy_id = ?", strategy.ID).
			Order("created_at desc").
			Limit(10).
			Find(&recentOrders)

		// 格式化最近订单
		formattedOrders := make([]map[string]interface{}, 0, len(recentOrders))
		for _, order := range recentOrders {
			formattedOrders = append(formattedOrders, map[string]interface{}{
				"id":        order.ID,
				"orderId":   order.OrderID,
				"side":      order.Side,
				"price":     order.Price,
				"quantity":  order.Quantity,
				"status":    order.Status,
				"createdAt": order.CreatedAt,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"stats":        stats,
			"recentOrders": formattedOrders,
			"strategy": map[string]interface{}{
				"id":            strategy.ID,
				"symbol":        strategy.Symbol,
				"strategyType":  strategy.StrategyType,
				"side":          strategy.Side,
				"price":         strategy.Price,
				"totalQuantity": strategy.TotalQuantity,
				"enabled":       strategy.Enabled,
				"status":        strategy.Status,
				"pendingBatch":  strategy.PendingBatch,
			},
		})
	}
}

// GinStrategyOrdersHandler 获取策略的所有订单（带权限验证）
func GinStrategyOrdersHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		strategyIDStr := c.Param("id")
		strategyID, err := strconv.ParseUint(strategyIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的策略ID"})
			return
		}

		// 验证策略所有权
		var strategy models.Strategy
		if err := cfg.DB.Where("id = ? AND user_id = ? AND deleted_at IS NULL", strategyID, user.ID).
			First(&strategy).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "策略未找到或无权访问"})
			return
		}

		// 获取所有订单
		var orders []models.Order
		query := cfg.DB.Where("strategy_id = ?", strategyID)

		// 支持状态筛选
		if status := c.Query("status"); status != "" {
			query = query.Where("status = ?", status)
		}

		// 按创建时间倒序
		if err := query.Order("created_at desc").Find(&orders).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取订单失败"})
			return
		}

		// 格式化订单数据
		formattedOrders := make([]map[string]interface{}, 0, len(orders))
		for _, order := range orders {
			formattedOrders = append(formattedOrders, map[string]interface{}{
				"id":          order.ID,
				"orderId":     order.OrderID,
				"symbol":      order.Symbol,
				"side":        order.Side,
				"price":       order.Price,
				"quantity":    order.Quantity,
				"status":      order.Status,
				"cancelAfter": order.CancelAfter,
				"createdAt":   order.CreatedAt,
				"updatedAt":   order.UpdatedAt,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"orders": formattedOrders,
			"total":  len(formattedOrders),
		})
	}
}
