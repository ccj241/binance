package tasks

import (
	"context"
	"log"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
)

// CheckOrders 定期检查订单状态并更新
func CheckOrders(cfg *config.Config) {
	for {
		var orders []models.Order
		// 使用 cfg.DB 替代 db.DB，查询待处理订单
		if err := cfg.DB.Where("status = ? AND deleted_at IS NULL", "pending").Find(&orders).Error; err != nil {
			log.Printf("获取待处理订单失败: %v", err)
			time.Sleep(30 * time.Second)
			continue
		}
		for _, order := range orders {
			var user models.User
			// 使用 cfg.DB 查询用户
			if err := cfg.DB.First(&user, order.UserID).Error; err != nil {
				log.Printf("订单 %d 未找到用户: %v", order.OrderID, err)
				continue
			}
			if user.APIKey == "" || user.SecretKey == "" {
				log.Printf("用户 %d 未设置 API 密钥，跳过订单 %d", user.ID, order.OrderID)
				continue
			}
			client := binance.NewClient(user.APIKey, user.SecretKey)
			binanceOrder, err := client.NewGetOrderService().
				Symbol(order.Symbol).
				OrderID(order.OrderID).
				Do(context.Background())
			if err != nil {
				log.Printf("获取订单 %d 失败: %v", order.OrderID, err)
				continue
			}
			var status string
			switch binanceOrder.Status {
			case binance.OrderStatusTypeFilled:
				status = "filled"
			case binance.OrderStatusTypeCanceled, binance.OrderStatusTypeExpired:
				status = "canceled"
			case binance.OrderStatusTypeNew, binance.OrderStatusTypePartiallyFilled:
				if time.Now().After(order.CancelAfter) {
					_, err := client.NewCancelOrderService().
						Symbol(order.Symbol).
						OrderID(order.OrderID).
						Do(context.Background())
					if err != nil {
						log.Printf("取消订单 %d 失败: %v", order.OrderID, err)
						continue
					}
					log.Printf("订单 %d 在2小时后取消", order.OrderID)
					status = "canceled"
				} else {
					continue
				}
			default:
				log.Printf("订单 %d 状态未知: %s", order.OrderID, binanceOrder.Status)
				continue
			}
			// 使用 cfg.DB 更新订单状态
			if err := cfg.DB.Model(&order).Update("status", status).Error; err != nil {
				log.Printf("更新订单 %d 状态失败: %v", order.OrderID, err)
				continue
			}
			log.Printf("订单 %d 更新状态为: %s", order.OrderID, status)

			// 检查策略的订单是否全部完成
			var pendingOrders []models.Order
			// 使用 cfg.DB 查询挂起的订单
			if err := cfg.DB.Where("strategy_id = ? AND status = ? AND deleted_at IS NULL", order.StrategyID, "pending").Find(&pendingOrders).Error; err != nil {
				log.Printf("检查策略 %d 的挂起订单失败: %v", order.StrategyID, err)
				continue
			}
			if len(pendingOrders) == 0 {
				var strategy models.Strategy
				// 使用 cfg.DB 查询策略
				if err := cfg.DB.First(&strategy, order.StrategyID).Error; err != nil {
					log.Printf("策略未找到: ID=%d", order.StrategyID)
					continue
				}
				// 使用 cfg.DB 更新策略
				if err := cfg.DB.Model(&strategy).Update("pending_batch", false).Error; err != nil {
					log.Printf("重置策略 %d 的 pending_batch 失败: %v", strategy.ID, err)
				} else {
					log.Printf("策略 %d 的 pending_batch 已重置", strategy.ID)
				}
			}
		}
		time.Sleep(30 * time.Second)
	}
}
