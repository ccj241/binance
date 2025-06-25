package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/adshao/go-binance/v2"
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
	"github.com/ccj241/binance/tasks"
	"github.com/gin-gonic/gin"
)

// getUserFromGinContext 从Gin上下文中获取用户信息
func getUserFromGinContext(c *gin.Context, cfg *config.Config) (*models.User, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return nil, fmt.Errorf("用户ID不存在")
	}

	uid, ok := userID.(uint)
	if !ok {
		return nil, fmt.Errorf("用户ID类型错误")
	}

	var user models.User
	if err := cfg.DB.First(&user, uid).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// GinPricesHandler Gin版本的价格处理器
func GinPricesHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		priceMap := make(map[string]float64)
		tasks.PriceMonitor.Range(func(key, value any) bool {
			symbolUser, ok := key.(string)
			if !ok {
				return true
			}
			price, ok := value.(float64)
			if !ok {
				return true
			}
			parts := strings.Split(symbolUser, "|")
			if len(parts) != 2 {
				return true
			}
			symbol, userIDStr := parts[0], parts[1]
			userID, err := strconv.ParseUint(userIDStr, 10, 32)
			if err != nil {
				return true
			}
			if uint(userID) == user.ID {
				priceMap[symbol] = price
			}
			return true
		})

		c.JSON(http.StatusOK, gin.H{"prices": priceMap})
	}
}

// GinBalanceHandler Gin版本的余额处理器
func GinBalanceHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		if user.APIKey == "" || user.SecretKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "API 密钥未设置"})
			return
		}

		client := binance.NewClient(user.APIKey, user.SecretKey)
		account, err := client.NewGetAccountService().Do(context.Background())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取余额失败"})
			return
		}

		balances := make([]map[string]interface{}, 0)
		for _, b := range account.Balances {
			free, _ := strconv.ParseFloat(b.Free, 64)
			locked, _ := strconv.ParseFloat(b.Locked, 64)
			if free > 0 || locked > 0 {
				balances = append(balances, map[string]interface{}{
					"asset":  b.Asset,
					"free":   free,
					"locked": locked,
				})
			}
		}

		c.JSON(http.StatusOK, gin.H{"balances": balances})
	}
}

// GinTradesHandler Gin版本的交易记录处理器
func GinTradesHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		var trades []models.Trade
		if err := cfg.DB.Where("user_id = ?", user.ID).Find(&trades).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取交易记录失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"trades": trades})
	}
}

// GinAddSymbolHandler Gin版本的添加交易对处理器
func GinAddSymbolHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		var request struct {
			Symbol string `json:"symbol" binding:"required"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求体"})
			return
		}

		// 检查是否已存在
		var existingSymbol models.CustomSymbol
		if err := cfg.DB.Where("user_id = ? AND symbol = ?", user.ID, request.Symbol).First(&existingSymbol).Error; err == nil {
			c.JSON(http.StatusOK, gin.H{"message": "Symbol 已存在"})
			return
		}

		symbol := models.CustomSymbol{
			UserID: user.ID,
			Symbol: request.Symbol,
		}
		if err := cfg.DB.Create(&symbol).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "添加 symbol 失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Symbol 添加成功"})
	}
}

// GinWithdrawalHistoryHandler Gin版本的取款历史处理器
func GinWithdrawalHistoryHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		var history []models.WithdrawalHistory
		if err := cfg.DB.Where("user_id = ?", user.ID).Find(&history).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取取款历史失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"history": history})
	}
}

// GinCreateWithdrawalRuleHandler Gin版本的创建取款规则处理器
func GinCreateWithdrawalRuleHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		var rule models.Withdrawal
		if err := c.ShouldBindJSON(&rule); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求体"})
			return
		}

		rule.UserID = user.ID
		if err := cfg.DB.Create(&rule).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建取款规则失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "取款规则创建成功"})
	}
}

// GinListWithdrawalRulesHandler Gin版本的列出取款规则处理器
func GinListWithdrawalRulesHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		var rules []models.Withdrawal
		if err := cfg.DB.Where("user_id = ? AND deleted_at IS NULL", user.ID).Find(&rules).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取取款规则失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"rules": rules})
	}
}

// GinUpdateWithdrawalRuleHandler Gin版本的更新取款规则处理器
func GinUpdateWithdrawalRuleHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		id := c.Param("id")
		var rule models.Withdrawal
		if err := cfg.DB.Where("id = ? AND user_id = ?", id, user.ID).First(&rule).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "取款规则未找到"})
			return
		}

		var updatedRule models.Withdrawal
		if err := c.ShouldBindJSON(&updatedRule); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求体"})
			return
		}

		rule.Asset = updatedRule.Asset
		rule.Address = updatedRule.Address
		rule.Threshold = updatedRule.Threshold
		rule.Amount = updatedRule.Amount
		rule.Enabled = updatedRule.Enabled

		if err := cfg.DB.Save(&rule).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新取款规则失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "取款规则更新成功"})
	}
}

// GinDeleteWithdrawalRuleHandler Gin版本的删除取款规则处理器
func GinDeleteWithdrawalRuleHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		id := c.Param("id")
		var rule models.Withdrawal
		if err := cfg.DB.Where("id = ? AND user_id = ?", id, user.ID).First(&rule).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "取款规则未找到"})
			return
		}

		if err := cfg.DB.Delete(&rule).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除取款规则失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "取款规则删除成功"})
	}
}

// 添加其他需要的handlers占位符，你可以根据需要实现
func GinOrdersHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"orders": []interface{}{}})
	}
}

func GinCancelledOrdersHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"orders": []interface{}{}})
	}
}

func GinCreateOrderHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "订单创建功能待实现"})
	}
}

func GinCancelOrderHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "订单取消功能待实现"})
	}
}

func GinCreateStrategyHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "策略创建功能待实现"})
	}
}

func GinListStrategiesHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"strategies": []interface{}{}})
	}
}

func GinToggleStrategyHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "策略切换功能待实现"})
	}
}

func GinDeleteStrategyHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "策略删除功能待实现"})
	}
}

func GinListSymbolsHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		var symbols []models.CustomSymbol
		if err := cfg.DB.Where("user_id = ? AND deleted_at IS NULL", user.ID).Find(&symbols).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取交易对失败"})
			return
		}

		symbolList := make([]string, len(symbols))
		for i, s := range symbols {
			symbolList[i] = s.Symbol
		}

		c.JSON(http.StatusOK, gin.H{"symbols": symbolList})
	}
}
