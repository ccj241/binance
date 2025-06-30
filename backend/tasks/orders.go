package tasks

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
)

// CheckOrders 定期检查订单状态并更新
func CheckOrders(cfg *config.Config) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		checkPendingOrders(cfg)
	}
}

// checkPendingOrders 检查待处理订单
func checkPendingOrders(cfg *config.Config) {
	var orders []models.Order

	// 查询所有待处理订单
	if err := cfg.DB.Where("status = ? AND deleted_at IS NULL", "pending").Find(&orders).Error; err != nil {
		log.Printf("获取待处理订单失败: %v", err)
		return
	}

	if len(orders) == 0 {
		return
	}

	// 移除订单数量日志

	// 按用户分组订单
	userOrders := make(map[uint][]models.Order)
	for _, order := range orders {
		userOrders[order.UserID] = append(userOrders[order.UserID], order)
	}

	// 处理每个用户的订单
	for userID, userOrderList := range userOrders {
		processUserOrders(cfg, userID, userOrderList)
	}
}

// processUserOrders 处理单个用户的订单
func processUserOrders(cfg *config.Config, userID uint, orders []models.Order) {
	// 获取用户信息
	var user models.User
	if err := cfg.DB.First(&user, userID).Error; err != nil {
		log.Printf("订单用户未找到: userID=%d, error=%v", userID, err)
		return
	}

	// 检查加密的API密钥是否存在
	if user.APIKey == "" || user.SecretKey == "" {
		// 移除日志，静默返回
		return
	}

	// 解密API密钥
	apiKey, err := user.GetDecryptedAPIKey()
	if err != nil {
		log.Printf("解密用户 %d API Key失败: %v", userID, err)
		return
	}

	// 验证解密后的API Key格式
	if apiKey == "" || len(apiKey) != 64 {
		// 移除验证日志，只在真正出错时记录
		return
	}

	secretKey, err := user.GetDecryptedSecretKey()
	if err != nil {
		log.Printf("解密用户 %d Secret Key失败: %v", userID, err)
		return
	}

	// 验证解密后的Secret Key格式
	if secretKey == "" || len(secretKey) != 64 {
		// 移除验证日志
		return
	}

	client := binance.NewClient(apiKey, secretKey)

	// 按交易对分组订单，减少API调用
	symbolOrders := make(map[string][]models.Order)
	for _, order := range orders {
		symbolOrders[order.Symbol] = append(symbolOrders[order.Symbol], order)
	}

	// 处理每个交易对的订单
	for symbol, symbolOrderList := range symbolOrders {
		processSymbolOrders(cfg, client, symbol, symbolOrderList)
	}
} // processUserOrders 处理单个用户的订单

// processSymbolOrders 处理特定交易对的订单
func processSymbolOrders(cfg *config.Config, client *binance.Client, symbol string, orders []models.Order) {
	// 批量获取该交易对的所有开放订单
	openOrders, err := client.NewListOpenOrdersService().Symbol(symbol).Do(context.Background())
	if err != nil {
		log.Printf("获取 %s 开放订单失败: %v", symbol, err)
		return
	}

	// 创建订单ID映射
	openOrderMap := make(map[int64]bool)
	for _, order := range openOrders {
		openOrderMap[order.OrderID] = true
	}

	// 检查每个订单的状态
	for _, order := range orders {
		updateOrderStatus(cfg, client, order, openOrderMap)
	}
}

// updateOrderStatus 更新单个订单状态
func updateOrderStatus(cfg *config.Config, client *binance.Client, order models.Order, openOrderMap map[int64]bool) {
	// 如果订单不在开放订单列表中，需要查询具体状态
	if !openOrderMap[order.OrderID] {
		// 查询订单详情
		binanceOrder, err := client.NewGetOrderService().
			Symbol(order.Symbol).
			OrderID(order.OrderID).
			Do(context.Background())

		if err != nil {
			log.Printf("查询订单 %d 失败: %v", order.OrderID, err)

			// 如果订单不存在，可能已被手动取消
			if isOrderNotFoundError(err) {
				updateOrderStatusInDB(cfg, &order, "cancelled")
			}
			return
		}

		// 根据币安订单状态更新本地状态
		switch binanceOrder.Status {
		case binance.OrderStatusTypeFilled:
			updateOrderStatusInDB(cfg, &order, "filled")
		case binance.OrderStatusTypeCanceled:
			updateOrderStatusInDB(cfg, &order, "cancelled")
		case binance.OrderStatusTypeExpired:
			updateOrderStatusInDB(cfg, &order, "expired")
		case binance.OrderStatusTypeRejected:
			updateOrderStatusInDB(cfg, &order, "rejected")
		case binance.OrderStatusTypePartiallyFilled:
			// 部分成交仍然是待处理状态，但需要检查是否超时
			checkOrderTimeout(cfg, client, &order)
		default:
			// 其他状态保持 pending
			checkOrderTimeout(cfg, client, &order)
		}
	} else {
		// 订单仍在开放列表中，检查是否需要超时取消
		checkOrderTimeout(cfg, client, &order)
	}
}

// checkOrderTimeout 检查订单是否超时
func checkOrderTimeout(cfg *config.Config, client *binance.Client, order *models.Order) {
	if time.Now().After(order.CancelAfter) {
		log.Printf("订单 %d 已超时，准备取消", order.OrderID)

		_, err := client.NewCancelOrderService().
			Symbol(order.Symbol).
			OrderID(order.OrderID).
			Do(context.Background())

		if err != nil {
			if !isOrderNotFoundError(err) {
				log.Printf("取消超时订单 %d 失败: %v", order.OrderID, err)
				return
			}
		}

		updateOrderStatusInDB(cfg, order, "cancelled")
		log.Printf("订单 %d 因超时被取消", order.OrderID)
	}
}

// updateOrderStatusInDB 更新数据库中的订单状态
func updateOrderStatusInDB(cfg *config.Config, order *models.Order, status string) {
	if err := cfg.DB.Model(order).Update("status", status).Error; err != nil {
		log.Printf("更新订单 %d 状态为 %s 失败: %v", order.OrderID, status, err)
		return
	}

	log.Printf("订单 %d 状态更新为: %s", order.OrderID, status)

	// 如果订单完成或取消，检查策略状态
	if status == "filled" || status == "cancelled" || status == "expired" || status == "rejected" {
		checkStrategyCompletion(cfg, order.StrategyID)
	}
}

// checkStrategyCompletion 检查策略是否完成
func checkStrategyCompletion(cfg *config.Config, strategyID uint) {
	if strategyID == 0 {
		return
	}

	// 查询该策略的所有待处理订单
	var pendingCount int64
	if err := cfg.DB.Model(&models.Order{}).
		Where("strategy_id = ? AND status = ? AND deleted_at IS NULL", strategyID, "pending").
		Count(&pendingCount).Error; err != nil {
		log.Printf("查询策略 %d 的待处理订单失败: %v", strategyID, err)
		return
	}

	// 如果没有待处理订单，重置策略的 pending_batch 标志
	if pendingCount == 0 {
		var strategy models.Strategy
		if err := cfg.DB.First(&strategy, strategyID).Error; err != nil {
			log.Printf("策略未找到: ID=%d, error=%v", strategyID, err)
			return
		}

		if strategy.PendingBatch {
			if err := cfg.DB.Model(&strategy).Update("pending_batch", false).Error; err != nil {
				log.Printf("重置策略 %d 的 pending_batch 失败: %v", strategy.ID, err)
			} else {
				log.Printf("策略 %d 的所有订单已完成，pending_batch 已重置", strategy.ID)
			}
		}
	}
}

// isOrderNotFoundError 判断是否为订单不存在错误
func isOrderNotFoundError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return contains(errStr, "Order does not exist") ||
		contains(errStr, "-2013") ||
		contains(errStr, "UNKNOWN_ORDER")
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && strings.Contains(s, substr)
}
