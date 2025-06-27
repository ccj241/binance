package handlers

import (
	"bytes"
	"context"
	"fmt"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

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

// GinBalanceHandler Gin版本的余额处理器 - 修复版本
// GinBalanceHandler Gin版本的余额处理器 - 修复版本（使用解密的API密钥）
func GinBalanceHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			log.Printf("获取用户失败: %v", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		log.Printf("获取用户 %d (%s) 的余额", user.ID, user.Username)

		// 解密API密钥
		apiKey, err := user.GetDecryptedAPIKey()
		if err != nil {
			log.Printf("解密用户 %d API Key失败: %v", user.ID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "API密钥解密失败"})
			return
		}

		secretKey, err := user.GetDecryptedSecretKey()
		if err != nil {
			log.Printf("解密用户 %d Secret Key失败: %v", user.ID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Secret密钥解密失败"})
			return
		}

		if apiKey == "" || secretKey == "" {
			log.Printf("用户 %d 未设置 API 密钥", user.ID)
			c.JSON(http.StatusBadRequest, gin.H{"error": "API 密钥未设置"})
			return
		}

		// 创建币安客户端
		client := binance.NewClient(apiKey, secretKey)

		// 设置超时上下文
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// 获取账户信息
		account, err := client.NewGetAccountService().Do(ctx)
		if err != nil {
			log.Printf("获取用户 %d 的账户信息失败: %v", user.ID, err)

			// 检查具体的错误类型
			errStr := err.Error()
			if strings.Contains(errStr, "Invalid API-key") || strings.Contains(errStr, "API-key format invalid") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "API 密钥无效，请检查您的密钥"})
				return
			} else if strings.Contains(errStr, "Signature for this request is not valid") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Secret 密钥无效，请检查您的密钥"})
				return
			} else if strings.Contains(errStr, "Timestamp for this request") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "时间同步错误，请检查系统时间"})
				return
			} else if strings.Contains(errStr, "IP address is not allowed") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "IP 地址未在白名单中，请检查 API 设置"})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取余额失败"})
			return
		}

		// 处理余额数据
		balances := make([]map[string]interface{}, 0)
		for _, b := range account.Balances {
			free, err := strconv.ParseFloat(b.Free, 64)
			if err != nil {
				log.Printf("解析 free 余额失败: asset=%s, value=%s, error=%v", b.Asset, b.Free, err)
				continue
			}

			locked, err := strconv.ParseFloat(b.Locked, 64)
			if err != nil {
				log.Printf("解析 locked 余额失败: asset=%s, value=%s, error=%v", b.Asset, b.Locked, err)
				continue
			}

			// 只返回有余额的资产
			if free > 0 || locked > 0 {
				balances = append(balances, map[string]interface{}{
					"asset":  b.Asset,
					"free":   free,
					"locked": locked,
				})
			}
		}

		log.Printf("成功获取用户 %d 的余额，共 %d 个资产有余额", user.ID, len(balances))
		c.JSON(http.StatusOK, gin.H{"balances": balances})
	}
}

// GinTradesHandler Gin版本的交易记录处理器 - 修复版本
func GinTradesHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		// 首先从数据库获取历史交易记录
		var dbTrades []models.Trade
		if err := cfg.DB.Where("user_id = ?", user.ID).Order("created_at desc").Find(&dbTrades).Error; err != nil {
			log.Printf("获取用户 %d 的交易记录失败: %v", user.ID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取交易记录失败"})
			return
		}

		// 如果用户设置了API密钥，尝试从币安获取最新交易
		if user.APIKey != "" && user.SecretKey != "" {
			// 解密API密钥
			apiKey, err := user.GetDecryptedAPIKey()
			if err != nil {
				log.Printf("解密用户 %d API Key失败: %v", user.ID, err)
				// 继续返回数据库中的交易记录
			} else {
				secretKey, err := user.GetDecryptedSecretKey()
				if err != nil {
					log.Printf("解密用户 %d Secret Key失败: %v", user.ID, err)
					// 继续返回数据库中的交易记录
				} else if apiKey != "" && secretKey != "" {
					// 使用解密后的密钥创建客户端
					client := binance.NewClient(apiKey, secretKey)

					// 获取用户的所有交易对
					var symbols []string
					cfg.DB.Model(&models.CustomSymbol{}).
						Where("user_id = ? AND deleted_at IS NULL", user.ID).
						Pluck("symbol", &symbols)

					// 为每个交易对获取最近的交易
					for _, symbol := range symbols {
						// 获取最近24小时的交易
						endTime := time.Now().UnixMilli()
						startTime := time.Now().Add(-24 * time.Hour).UnixMilli()

						trades, err := client.NewListTradesService().
							Symbol(symbol).
							StartTime(startTime).
							EndTime(endTime).
							Limit(100). // 每个交易对最多获取100条记录
							Do(context.Background())

						if err != nil {
							log.Printf("获取 %s 交易记录失败: %v", symbol, err)
							continue
						}

						// 将新交易保存到数据库
						for _, trade := range trades {
							price, _ := strconv.ParseFloat(trade.Price, 64)
							qty, _ := strconv.ParseFloat(trade.Quantity, 64) // 使用 Quantity 而不是 Qty

							// 检查交易是否已存在
							var exists bool
							cfg.DB.Model(&models.Trade{}).
								Where("user_id = ? AND symbol = ? AND time = ?", user.ID, symbol, trade.Time).
								Select("count(*) > 0").
								Find(&exists)

							if !exists {
								newTrade := models.Trade{
									UserID: user.ID,
									Symbol: symbol,
									Price:  price,
									Qty:    qty,
									Time:   trade.Time,
								}
								if err := cfg.DB.Create(&newTrade).Error; err != nil {
									log.Printf("保存交易记录失败: %v", err)
								}
							}
						}
					}

					// 重新查询数据库以获取所有交易（包括新添加的）
					cfg.DB.Where("user_id = ?", user.ID).Order("time desc").Find(&dbTrades)
				}
			}
		}

		// 格式化交易记录
		trades := make([]map[string]interface{}, 0, len(dbTrades))
		for _, trade := range dbTrades {
			trades = append(trades, map[string]interface{}{
				"id":        trade.ID,
				"symbol":    trade.Symbol,
				"price":     trade.Price,
				"qty":       trade.Qty,
				"time":      trade.Time,
				"createdAt": trade.CreatedAt,
			})
		}

		c.JSON(http.StatusOK, gin.H{"trades": trades})
	}
}

// GinOrdersHandler Gin版本的订单处理器 - 修复版本
func GinOrdersHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		// 先从数据库获取所有订单（包括历史订单）
		var dbOrders []models.Order
		if err := cfg.DB.Where("user_id = ?", user.ID).Order("created_at desc").Find(&dbOrders).Error; err != nil {
			log.Printf("获取用户 %d 的订单失败: %v", user.ID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取订单失败"})
			return
		}

		// 如果用户设置了API密钥，同步开放订单
		if user.APIKey != "" && user.SecretKey != "" {
			// 解密API密钥
			apiKey, err := user.GetDecryptedAPIKey()
			if err != nil {
				log.Printf("解密用户 %d API Key失败: %v", user.ID, err)
				// 继续返回数据库中的订单
			} else if apiKey == "" {
				log.Printf("用户 %d 解密后的API Key为空", user.ID)
			} else if len(apiKey) != 64 {
				log.Printf("用户 %d API Key格式错误，长度=%d，期望=64", user.ID, len(apiKey))
			} else {
				// API Key格式正确，继续解密Secret Key
				secretKey, err := user.GetDecryptedSecretKey()
				if err != nil {
					log.Printf("解密用户 %d Secret Key失败: %v", user.ID, err)
				} else if secretKey == "" {
					log.Printf("用户 %d 解密后的Secret Key为空", user.ID)
				} else if len(secretKey) != 64 {
					log.Printf("用户 %d Secret Key格式错误，长度=%d，期望=64", user.ID, len(secretKey))
				} else {
					// 两个密钥都正确，尝试获取开放订单
					client := binance.NewClient(apiKey, secretKey)

					// 获取所有开放订单
					openOrders, err := client.NewListOpenOrdersService().Do(context.Background())
					if err != nil {
						log.Printf("获取开放订单失败: %v", err)
						// 即使API调用失败，也继续返回数据库中的订单
					} else {
						// 创建开放订单映射
						openOrderMap := make(map[int64]bool)
						for _, order := range openOrders {
							openOrderMap[order.OrderID] = true

							// 检查订单是否已在数据库中
							var dbOrder models.Order
							result := cfg.DB.Where("order_id = ? AND user_id = ?", order.OrderID, user.ID).First(&dbOrder)

							price, _ := strconv.ParseFloat(order.Price, 64)
							quantity, _ := strconv.ParseFloat(order.OrigQuantity, 64)

							if result.Error != nil {
								// 订单不存在，创建新订单
								newOrder := models.Order{
									UserID:      user.ID,
									Symbol:      order.Symbol,
									Side:        string(order.Side),
									Price:       price,
									Quantity:    quantity,
									OrderID:     order.OrderID,
									Status:      "pending",
									CancelAfter: time.Now().Add(2 * time.Hour),
								}
								cfg.DB.Create(&newOrder)
							} else {
								// 更新现有订单状态
								if dbOrder.Status != "pending" {
									cfg.DB.Model(&dbOrder).Update("status", "pending")
								}
							}
						}

						// 更新数据库中的订单状态
						for i := range dbOrders {
							if dbOrders[i].Status == "pending" && !openOrderMap[dbOrders[i].OrderID] {
								// 订单不在开放订单列表中，可能已完成或取消
								// 这里暂时不更新状态，让后台任务处理
							}
						}
					}

					// 重新查询数据库
					cfg.DB.Where("user_id = ?", user.ID).Order("created_at desc").Find(&dbOrders)
				}
			}
		}

		// 格式化订单数据
		orders := make([]map[string]interface{}, 0, len(dbOrders))
		for _, order := range dbOrders {
			orders = append(orders, map[string]interface{}{
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

		c.JSON(http.StatusOK, gin.H{"orders": orders})
	}
}

// GinCancelledOrdersHandler Gin版本的已取消订单处理器
func GinCancelledOrdersHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		var orders []models.Order
		if err := cfg.DB.Where("user_id = ? AND status IN (?, ?, ?)",
			user.ID, "cancelled", "expired", "rejected").
			Order("updated_at desc").
			Find(&orders).Error; err != nil {
			log.Printf("获取已取消订单失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取已取消订单失败"})
			return
		}

		// 格式化订单数据
		formattedOrders := make([]map[string]interface{}, 0, len(orders))
		for _, order := range orders {
			formattedOrders = append(formattedOrders, map[string]interface{}{
				"id":        order.ID,
				"orderId":   order.OrderID,
				"symbol":    order.Symbol,
				"side":      order.Side,
				"price":     order.Price,
				"quantity":  order.Quantity,
				"status":    order.Status,
				"createdAt": order.CreatedAt,
				"updatedAt": order.UpdatedAt,
			})
		}

		c.JSON(http.StatusOK, gin.H{"orders": formattedOrders})
	}
}

// GinCreateOrderHandler Gin版本的创建订单处理器
func GinCreateOrderHandler(cfg *config.Config) gin.HandlerFunc {
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

		var orderReq struct {
			Symbol   string  `json:"symbol" binding:"required"`
			Side     string  `json:"side" binding:"required"`
			Quantity float64 `json:"quantity" binding:"required,gt=0"`
			Price    float64 `json:"price" binding:"required,gt=0"`
		}

		if err := c.ShouldBindJSON(&orderReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求", "details": err.Error()})
			return
		}

		// 验证side
		if orderReq.Side != "BUY" && orderReq.Side != "SELL" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的交易方向"})
			return
		}

		client := binance.NewClient(user.APIKey, user.SecretKey)

		// 创建订单
		order, err := client.NewCreateOrderService().
			Symbol(orderReq.Symbol).
			Side(binance.SideType(orderReq.Side)).
			Type(binance.OrderTypeLimit).
			TimeInForce(binance.TimeInForceTypeGTC).
			Quantity(fmt.Sprintf("%.8f", orderReq.Quantity)).
			Price(fmt.Sprintf("%.8f", orderReq.Price)).
			Do(context.Background())

		if err != nil {
			log.Printf("创建订单失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("创建订单失败: %v", err)})
			return
		}

		// 保存到数据库
		dbOrder := models.Order{
			UserID:      user.ID,
			Symbol:      orderReq.Symbol,
			Side:        orderReq.Side,
			Price:       orderReq.Price,
			Quantity:    orderReq.Quantity,
			OrderID:     order.OrderID,
			Status:      "pending",
			CancelAfter: time.Now().Add(2 * time.Hour),
		}

		if err := cfg.DB.Create(&dbOrder).Error; err != nil {
			log.Printf("保存订单失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存订单失败"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "订单创建成功",
			"orderId": order.OrderID,
			"order": map[string]interface{}{
				"id":       dbOrder.ID,
				"orderId":  dbOrder.OrderID,
				"symbol":   dbOrder.Symbol,
				"side":     dbOrder.Side,
				"price":    dbOrder.Price,
				"quantity": dbOrder.Quantity,
				"status":   dbOrder.Status,
			},
		})
	}
}

// GinCancelOrderHandler Gin版本的取消订单处理器
func GinCancelOrderHandler(cfg *config.Config) gin.HandlerFunc {
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

		orderIDStr := c.Param("orderId")
		orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的订单ID"})
			return
		}

		// 查找订单
		var order models.Order
		if err := cfg.DB.Where("order_id = ? AND user_id = ?", orderID, user.ID).First(&order).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "订单未找到"})
			return
		}

		if order.Status != "pending" {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("订单状态为 %s，无法取消", order.Status)})
			return
		}

		client := binance.NewClient(user.APIKey, user.SecretKey)

		// 取消订单
		_, err = client.NewCancelOrderService().
			Symbol(order.Symbol).
			OrderID(order.OrderID).
			Do(context.Background())

		if err != nil {
			// 检查是否因为订单已经不存在
			if strings.Contains(err.Error(), "Order does not exist") {
				// 更新本地状态
				cfg.DB.Model(&order).Update("status", "cancelled")
				c.JSON(http.StatusOK, gin.H{"message": "订单已取消"})
				return
			}

			log.Printf("取消订单失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("取消订单失败: %v", err)})
			return
		}

		// 更新订单状态
		if err := cfg.DB.Model(&order).Update("status", "cancelled").Error; err != nil {
			log.Printf("更新订单状态失败: %v", err)
		}

		c.JSON(http.StatusOK, gin.H{"message": "订单已取消"})
	}
}

// GinBatchCancelOrdersHandler 批量取消订单处理器
func GinBatchCancelOrdersHandler(cfg *config.Config) gin.HandlerFunc {
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

		var request struct {
			OrderIDs []int64 `json:"orderIds" binding:"required"`
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
			return
		}

		if len(request.OrderIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请提供要取消的订单ID"})
			return
		}

		// 查询所有订单
		var orders []models.Order
		if err := cfg.DB.Where("order_id IN ? AND user_id = ? AND status = ?",
			request.OrderIDs, user.ID, "pending").Find(&orders).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询订单失败"})
			return
		}

		if len(orders) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "未找到可取消的订单"})
			return
		}

		client := binance.NewClient(user.APIKey, user.SecretKey)
		results := struct {
			Success []int64 `json:"success"`
			Failed  []struct {
				OrderID int64  `json:"orderId"`
				Error   string `json:"error"`
			} `json:"failed"`
		}{
			Success: []int64{},
			Failed: []struct {
				OrderID int64  `json:"orderId"`
				Error   string `json:"error"`
			}{},
		}

		// 批量取消订单
		for _, order := range orders {
			_, err := client.NewCancelOrderService().
				Symbol(order.Symbol).
				OrderID(order.OrderID).
				Do(context.Background())

			if err != nil {
				// 检查是否因为订单已经不存在
				if strings.Contains(err.Error(), "Order does not exist") {
					// 更新本地状态
					cfg.DB.Model(&order).Update("status", "cancelled")
					results.Success = append(results.Success, order.OrderID)
				} else {
					results.Failed = append(results.Failed, struct {
						OrderID int64  `json:"orderId"`
						Error   string `json:"error"`
					}{
						OrderID: order.OrderID,
						Error:   err.Error(),
					})
				}
			} else {
				// 更新订单状态
				cfg.DB.Model(&order).Update("status", "cancelled")
				results.Success = append(results.Success, order.OrderID)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("成功取消 %d 个订单，失败 %d 个",
				len(results.Success), len(results.Failed)),
			"results": results,
		})
	}
}

// GinWithdrawalHistoryHandler Gin版本的提币历史处理器 - 修复版本
func GinWithdrawalHistoryHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		// 从数据库获取提币历史
		var history []models.WithdrawalHistory
		if err := cfg.DB.Where("user_id = ?", user.ID).
			Order("created_at desc").
			Find(&history).Error; err != nil {
			log.Printf("获取用户 %d 的提币历史失败: %v", user.ID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取提币历史失败"})
			return
		}

		// 如果用户设置了API密钥，尝试从币安获取最新提币历史
		if user.APIKey != "" && user.SecretKey != "" && len(history) == 0 {
			client := binance.NewClient(user.APIKey, user.SecretKey)

			// 获取最近90天的提币历史
			endTime := time.Now().UnixMilli()
			startTime := time.Now().AddDate(0, 0, -90).UnixMilli()

			withdrawals, err := client.NewListWithdrawsService().
				StartTime(startTime).
				EndTime(endTime).
				Do(context.Background())

			if err != nil {
				log.Printf("获取币安提币历史失败: %v", err)
			} else {
				// 保存提币历史到数据库
				for _, w := range withdrawals {
					amount, _ := strconv.ParseFloat(w.Amount, 64)

					withdrawalHistory := models.WithdrawalHistory{
						UserID:       user.ID,
						Asset:        w.Coin,
						Amount:       amount,
						Address:      w.Address,
						WithdrawalID: w.ID,
						TxID:         w.TxID,
						Status:       fmt.Sprintf("%d", w.Status), // 将状态码转为字符串
					}

					// 检查是否已存在
					var exists bool
					cfg.DB.Model(&models.WithdrawalHistory{}).
						Where("user_id = ? AND withdrawal_id = ?", user.ID, w.ID).
						Select("count(*) > 0").
						Find(&exists)

					if !exists {
						if err := cfg.DB.Create(&withdrawalHistory).Error; err != nil {
							log.Printf("保存提币历史失败: %v", err)
						}
					}
				}

				// 重新查询数据库
				cfg.DB.Where("user_id = ?", user.ID).Order("created_at desc").Find(&history)
			}
		}

		// 格式化历史记录
		formattedHistory := make([]map[string]interface{}, 0, len(history))
		for _, h := range history {
			formattedHistory = append(formattedHistory, map[string]interface{}{
				"id":           h.ID,
				"asset":        h.Asset,
				"amount":       h.Amount,
				"address":      h.Address,
				"withdrawalId": h.WithdrawalID,
				"txId":         h.TxID,
				"status":       h.Status,
				"createdAt":    h.CreatedAt,
			})
		}

		c.JSON(http.StatusOK, gin.H{"history": formattedHistory})
	}
}

// GinAddSymbolHandler 添加交易对处理器
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

		// 启动价格监控
		tasks.MonitorNewSymbol(request.Symbol, user.ID, cfg)

		c.JSON(http.StatusOK, gin.H{"message": "Symbol 添加成功"})
	}
}

// GinDeleteSymbolHandler 删除交易对处理器 - 修复版本
func GinDeleteSymbolHandler(cfg *config.Config) gin.HandlerFunc {
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

		if request.Symbol == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "交易对不能为空"})
			return
		}

		// 标准化交易对名称
		symbolName := strings.ToUpper(request.Symbol)

		// 查找要删除的交易对
		var symbol models.CustomSymbol
		if err := cfg.DB.Where("user_id = ? AND symbol = ? AND deleted_at IS NULL", user.ID, symbolName).First(&symbol).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "交易对未找到"})
				return
			}
			log.Printf("查询交易对失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询交易对失败"})
			return
		}

		// 检查是否有相关的活跃策略
		var activeStrategyCount int64
		if err := cfg.DB.Model(&models.Strategy{}).Where(
			"user_id = ? AND symbol = ? AND enabled = ? AND deleted_at IS NULL",
			user.ID, symbolName, true,
		).Count(&activeStrategyCount).Error; err != nil {
			log.Printf("检查策略失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "检查策略失败"})
			return
		}

		if activeStrategyCount > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("无法删除交易对 %s，存在 %d 个活跃策略，请先禁用或删除相关策略", symbolName, activeStrategyCount),
			})
			return
		}

		// 检查是否有待处理的订单
		var pendingOrderCount int64
		if err := cfg.DB.Model(&models.Order{}).Where(
			"user_id = ? AND symbol = ? AND status = ? AND deleted_at IS NULL",
			user.ID, symbolName, "pending",
		).Count(&pendingOrderCount).Error; err != nil {
			log.Printf("检查订单失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "检查订单失败"})
			return
		}

		if pendingOrderCount > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("无法删除交易对 %s，存在 %d 个待处理订单，请先取消相关订单", symbolName, pendingOrderCount),
			})
			return
		}

		// 执行软删除
		if err := cfg.DB.Delete(&symbol).Error; err != nil {
			log.Printf("删除用户 %d 的交易对 %s 失败: %v", user.ID, symbolName, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除交易对失败"})
			return
		}

		// 停止价格监控
		tasks.StopSymbolMonitoring(symbolName, user.ID)

		log.Printf("用户 %d 成功删除交易对 %s", user.ID, symbolName)
		c.JSON(http.StatusOK, gin.H{"message": "交易对删除成功"})
	}
}

// GinCreateWithdrawalRuleHandler 创建提币规则处理器 - 修复版本
func GinCreateWithdrawalRuleHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		// 读取原始请求体
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("读取请求体失败: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求体失败"})
			return
		}

		// 重新设置请求体以供绑定使用
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// 使用自定义结构体接收请求数据
		var req struct {
			Asset     string  `json:"asset" binding:"required"`
			Threshold float64 `json:"threshold" binding:"required,gt=0"`
			Amount    float64 `json:"amount" binding:"min=0"` // 允许为0，表示提取最大值
			Address   string  `json:"address" binding:"required"`
			Enabled   bool    `json:"enabled"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("绑定提币规则请求失败: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "无效的请求体",
				"details": err.Error(),
			})
			return
		}

		// 额外验证
		if len(strings.TrimSpace(req.Asset)) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "资产名称不能为空"})
			return
		}

		if len(strings.TrimSpace(req.Address)) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "提币地址不能为空"})
			return
		}

		// 检查是否已存在相同资产和地址的规则
		var existingRule models.Withdrawal
		if err := cfg.DB.Where("user_id = ? AND asset = ? AND address = ? AND deleted_at IS NULL",
			user.ID, strings.ToUpper(strings.TrimSpace(req.Asset)), strings.TrimSpace(req.Address)).
			First(&existingRule).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "相同资产和地址的提币规则已存在"})
			return
		}

		// 创建提币规则
		rule := models.Withdrawal{
			UserID:    user.ID,
			Asset:     strings.ToUpper(strings.TrimSpace(req.Asset)),
			Threshold: req.Threshold,
			Amount:    req.Amount, // 如果为0，表示提取最大可用金额
			Address:   strings.TrimSpace(req.Address),
			Enabled:   req.Enabled,
			Status:    "active",
		}

		if err := cfg.DB.Create(&rule).Error; err != nil {
			log.Printf("创建提币规则失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建提币规则失败"})
			return
		}

		log.Printf("用户 %d 创建提币规则成功: %s, 阈值=%.8f, 金额=%.8f",
			user.ID, rule.Asset, rule.Threshold, rule.Amount)

		c.JSON(http.StatusOK, gin.H{
			"message": "提币规则创建成功",
			"rule": map[string]interface{}{
				"id":        rule.ID,
				"asset":     rule.Asset,
				"threshold": rule.Threshold,
				"amount":    rule.Amount,
				"address":   rule.Address,
				"enabled":   rule.Enabled,
				"status":    rule.Status,
				"createdAt": rule.CreatedAt,
			},
		})
	}
}

// GinListWithdrawalRulesHandler 获取提币规则列表处理器
func GinListWithdrawalRulesHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		var rules []models.Withdrawal
		if err := cfg.DB.Where("user_id = ? AND deleted_at IS NULL", user.ID).
			Order("created_at desc").
			Find(&rules).Error; err != nil {
			log.Printf("获取用户 %d 的提币规则失败: %v", user.ID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取提币规则失败"})
			return
		}

		// 格式化返回数据
		formattedRules := make([]map[string]interface{}, 0, len(rules))
		for _, rule := range rules {
			formattedRules = append(formattedRules, map[string]interface{}{
				"id":        rule.ID,
				"asset":     rule.Asset,
				"threshold": rule.Threshold,
				"amount":    rule.Amount,
				"address":   rule.Address,
				"enabled":   rule.Enabled,
				"status":    rule.Status,
				"createdAt": rule.CreatedAt,
				"updatedAt": rule.UpdatedAt,
			})
		}

		c.JSON(http.StatusOK, gin.H{"rules": formattedRules})
	}
}

// GinUpdateWithdrawalRuleHandler 更新提币规则处理器
func GinUpdateWithdrawalRuleHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		id := c.Param("id")
		var rule models.Withdrawal
		if err := cfg.DB.Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, user.ID).
			First(&rule).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "提币规则未找到"})
			return
		}

		// 读取原始请求体
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Printf("读取请求体失败: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "读取请求体失败"})
			return
		}

		// 重新设置请求体
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// 使用自定义结构体接收更新数据
		var updateReq struct {
			Asset     string  `json:"asset"`
			Threshold float64 `json:"threshold"`
			Amount    float64 `json:"amount"`
			Address   string  `json:"address"`
			Enabled   bool    `json:"enabled"`
		}

		if err := c.ShouldBindJSON(&updateReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "无效的请求体",
				"details": err.Error(),
			})
			return
		}

		// 验证更新数据
		if len(strings.TrimSpace(updateReq.Asset)) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "资产名称不能为空"})
			return
		}

		if updateReq.Threshold <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "阈值必须大于0"})
			return
		}

		if updateReq.Amount < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "提币金额不能为负数"})
			return
		}

		if len(strings.TrimSpace(updateReq.Address)) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "提币地址不能为空"})
			return
		}

		// 更新规则
		updates := map[string]interface{}{
			"asset":     strings.ToUpper(strings.TrimSpace(updateReq.Asset)),
			"threshold": updateReq.Threshold,
			"amount":    updateReq.Amount,
			"address":   strings.TrimSpace(updateReq.Address),
			"enabled":   updateReq.Enabled,
		}

		if err := cfg.DB.Model(&rule).Updates(updates).Error; err != nil {
			log.Printf("更新提币规则失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新提币规则失败"})
			return
		}

		log.Printf("用户 %d 更新提币规则 %d 成功", user.ID, rule.ID)

		// 重新获取更新后的规则
		if err := cfg.DB.First(&rule, rule.ID).Error; err != nil {
			log.Printf("重新获取规则失败: %v", err)
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "提币规则更新成功",
			"rule": map[string]interface{}{
				"id":        rule.ID,
				"asset":     rule.Asset,
				"threshold": rule.Threshold,
				"amount":    rule.Amount,
				"address":   rule.Address,
				"enabled":   rule.Enabled,
				"status":    rule.Status,
				"updatedAt": rule.UpdatedAt,
			},
		})
	}
}

// GinDeleteWithdrawalRuleHandler 删除提币规则处理器
func GinDeleteWithdrawalRuleHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		id := c.Param("id")
		var rule models.Withdrawal
		if err := cfg.DB.Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, user.ID).
			First(&rule).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "提币规则未找到"})
			return
		}

		if err := cfg.DB.Delete(&rule).Error; err != nil {
			log.Printf("删除提币规则失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除提币规则失败"})
			return
		}

		log.Printf("用户 %d 删除提币规则 %d 成功", user.ID, rule.ID)

		c.JSON(http.StatusOK, gin.H{"message": "提币规则删除成功"})
	}
}

// GinCreateStrategyHandler 创建策略处理器 - 修复版本
func GinCreateStrategyHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		var strategyReq struct {
			Symbol             string    `json:"symbol" binding:"required"`
			StrategyType       string    `json:"strategyType" binding:"required"`
			Side               string    `json:"side" binding:"required"`
			Price              float64   `json:"price" binding:"required,gt=0"`
			TotalQuantity      float64   `json:"totalQuantity" binding:"required,gt=0"`
			BuyQuantities      []float64 `json:"buyQuantities"`
			SellQuantities     []float64 `json:"sellQuantities"`
			BuyDepthLevels     []int     `json:"buyDepthLevels"`
			SellDepthLevels    []int     `json:"sellDepthLevels"`
			BuyBasisPoints     []float64 `json:"buyBasisPoints"`  // 新增：买入万分比
			SellBasisPoints    []float64 `json:"sellBasisPoints"` // 新增：卖出万分比
			CancelAfterMinutes int       `json:"cancelAfterMinutes"`
		}

		if err := c.ShouldBindJSON(&strategyReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据", "details": err.Error()})
			return
		}

		// 验证策略类型
		if strategyReq.StrategyType != "simple" && strategyReq.StrategyType != "iceberg" && strategyReq.StrategyType != "custom" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的策略类型"})
			return
		}

		// 验证交易方向
		if strategyReq.Side != "BUY" && strategyReq.Side != "SELL" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的交易方向"})
			return
		}

		// 设置默认的取消时间
		if strategyReq.CancelAfterMinutes <= 0 {
			strategyReq.CancelAfterMinutes = 120
		}

		// 处理默认值
		if strategyReq.StrategyType == "simple" {
			if strategyReq.Side == "BUY" {
				strategyReq.BuyQuantities = []float64{1.0}
				strategyReq.BuyDepthLevels = []int{1}
				strategyReq.SellQuantities = []float64{}
				strategyReq.SellDepthLevels = []int{}
			} else {
				strategyReq.SellQuantities = []float64{1.0}
				strategyReq.SellDepthLevels = []int{1}
				strategyReq.BuyQuantities = []float64{}
				strategyReq.BuyDepthLevels = []int{}
			}
		} else if strategyReq.StrategyType == "iceberg" {
			if strategyReq.Side == "BUY" {
				if len(strategyReq.BuyQuantities) == 0 {
					strategyReq.BuyQuantities = []float64{0.35, 0.25, 0.2, 0.1, 0.1}
					strategyReq.BuyDepthLevels = []int{1, 3, 5, 7, 9}
				}
				strategyReq.SellQuantities = []float64{}
				strategyReq.SellDepthLevels = []int{}
			} else {
				if len(strategyReq.SellQuantities) == 0 {
					strategyReq.SellQuantities = []float64{0.35, 0.25, 0.2, 0.1, 0.1}
					strategyReq.SellDepthLevels = []int{1, 3, 5, 7, 9}
				}
				strategyReq.BuyQuantities = []float64{}
				strategyReq.BuyDepthLevels = []int{}
			}
		}

		// 验证自定义策略配置 - 修复版本
		if strategyReq.StrategyType == "custom" {
			if strategyReq.Side == "BUY" {
				// 买入策略需要买入数量
				if len(strategyReq.BuyQuantities) == 0 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "自定义买入策略需要设置数量"})
					return
				}

				// 如果提供了万分比，使用万分比；否则使用深度级别
				if len(strategyReq.BuyBasisPoints) > 0 {
					// 使用万分比，不需要深度级别
					if len(strategyReq.BuyQuantities) != len(strategyReq.BuyBasisPoints) {
						c.JSON(http.StatusBadRequest, gin.H{"error": "买入数量和万分比数量不匹配"})
						return
					}
					// 为了兼容，设置默认深度级别
					strategyReq.BuyDepthLevels = make([]int, len(strategyReq.BuyQuantities))
					for i := range strategyReq.BuyDepthLevels {
						strategyReq.BuyDepthLevels[i] = i + 1
					}
				} else if len(strategyReq.BuyDepthLevels) > 0 {
					// 使用深度级别
					if len(strategyReq.BuyQuantities) != len(strategyReq.BuyDepthLevels) {
						c.JSON(http.StatusBadRequest, gin.H{"error": "买入数量和深度级别数量不匹配"})
						return
					}
				} else {
					// 两者都没有，使用默认深度级别
					strategyReq.BuyDepthLevels = make([]int, len(strategyReq.BuyQuantities))
					for i := range strategyReq.BuyDepthLevels {
						strategyReq.BuyDepthLevels[i] = i + 1
					}
				}
			} else {
				// 卖出策略需要卖出数量
				if len(strategyReq.SellQuantities) == 0 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "自定义卖出策略需要设置数量"})
					return
				}

				// 如果提供了万分比，使用万分比；否则使用深度级别
				if len(strategyReq.SellBasisPoints) > 0 {
					// 使用万分比，不需要深度级别
					if len(strategyReq.SellQuantities) != len(strategyReq.SellBasisPoints) {
						c.JSON(http.StatusBadRequest, gin.H{"error": "卖出数量和万分比数量不匹配"})
						return
					}
					// 为了兼容，设置默认深度级别
					strategyReq.SellDepthLevels = make([]int, len(strategyReq.SellQuantities))
					for i := range strategyReq.SellDepthLevels {
						strategyReq.SellDepthLevels[i] = i + 1
					}
				} else if len(strategyReq.SellDepthLevels) > 0 {
					// 使用深度级别
					if len(strategyReq.SellQuantities) != len(strategyReq.SellDepthLevels) {
						c.JSON(http.StatusBadRequest, gin.H{"error": "卖出数量和深度级别数量不匹配"})
						return
					}
				} else {
					// 两者都没有，使用默认深度级别
					strategyReq.SellDepthLevels = make([]int, len(strategyReq.SellQuantities))
					for i := range strategyReq.SellDepthLevels {
						strategyReq.SellDepthLevels[i] = i + 1
					}
				}
			}
		}

		// 转换为字符串存储
		buyQuantitiesStr := ""
		if len(strategyReq.BuyQuantities) > 0 {
			strs := make([]string, len(strategyReq.BuyQuantities))
			for i, q := range strategyReq.BuyQuantities {
				strs[i] = fmt.Sprintf("%.8f", q)
			}
			buyQuantitiesStr = strings.Join(strs, ",")
		}

		sellQuantitiesStr := ""
		if len(strategyReq.SellQuantities) > 0 {
			strs := make([]string, len(strategyReq.SellQuantities))
			for i, q := range strategyReq.SellQuantities {
				strs[i] = fmt.Sprintf("%.8f", q)
			}
			sellQuantitiesStr = strings.Join(strs, ",")
		}

		buyDepthLevelsStr := ""
		if len(strategyReq.BuyDepthLevels) > 0 {
			strs := make([]string, len(strategyReq.BuyDepthLevels))
			for i, d := range strategyReq.BuyDepthLevels {
				strs[i] = fmt.Sprintf("%d", d)
			}
			buyDepthLevelsStr = strings.Join(strs, ",")
		}

		sellDepthLevelsStr := ""
		if len(strategyReq.SellDepthLevels) > 0 {
			strs := make([]string, len(strategyReq.SellDepthLevels))
			for i, d := range strategyReq.SellDepthLevels {
				strs[i] = fmt.Sprintf("%d", d)
			}
			sellDepthLevelsStr = strings.Join(strs, ",")
		}

		// 新增：转换万分比为字符串
		buyBasisPointsStr := ""
		if len(strategyReq.BuyBasisPoints) > 0 {
			strs := make([]string, len(strategyReq.BuyBasisPoints))
			for i, bp := range strategyReq.BuyBasisPoints {
				strs[i] = fmt.Sprintf("%.2f", bp)
			}
			buyBasisPointsStr = strings.Join(strs, ",")
		}

		sellBasisPointsStr := ""
		if len(strategyReq.SellBasisPoints) > 0 {
			strs := make([]string, len(strategyReq.SellBasisPoints))
			for i, bp := range strategyReq.SellBasisPoints {
				strs[i] = fmt.Sprintf("%.2f", bp)
			}
			sellBasisPointsStr = strings.Join(strs, ",")
		}

		// 创建策略
		strategy := models.Strategy{
			UserID:             user.ID,
			Symbol:             strings.ToUpper(strategyReq.Symbol),
			StrategyType:       strategyReq.StrategyType,
			Side:               strategyReq.Side,
			Price:              strategyReq.Price,
			TotalQuantity:      strategyReq.TotalQuantity,
			Status:             "active",
			Enabled:            true,
			BuyQuantities:      buyQuantitiesStr,
			SellQuantities:     sellQuantitiesStr,
			BuyDepthLevels:     buyDepthLevelsStr,
			SellDepthLevels:    sellDepthLevelsStr,
			BuyBasisPoints:     buyBasisPointsStr,  // 新增
			SellBasisPoints:    sellBasisPointsStr, // 新增
			CancelAfterMinutes: strategyReq.CancelAfterMinutes,
		}

		if err := cfg.DB.Create(&strategy).Error; err != nil {
			log.Printf("创建策略失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建策略失败"})
			return
		}

		// 启动价格监控
		tasks.MonitorNewSymbol(strategy.Symbol, user.ID, cfg)

		log.Printf("策略创建成功: ID=%d, Symbol=%s, Type=%s, Side=%s, UserID=%d",
			strategy.ID, strategy.Symbol, strategy.StrategyType, strategy.Side, user.ID)

		c.JSON(http.StatusOK, gin.H{
			"message":    "策略创建成功",
			"strategyId": strategy.ID,
		})
	}
}

// GinListStrategiesHandler 获取策略列表处理器
func GinListStrategiesHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		var strategies []models.Strategy
		if err := cfg.DB.Where("user_id = ? AND deleted_at IS NULL", user.ID).
			Order("created_at desc").
			Find(&strategies).Error; err != nil {
			log.Printf("获取策略失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取策略失败"})
			return
		}

		// 格式化策略数据
		formattedStrategies := make([]map[string]interface{}, 0, len(strategies))
		for _, s := range strategies {
			// 解析数量和深度配置
			buyQuantities := []float64{}
			if s.BuyQuantities != "" {
				for _, q := range strings.Split(s.BuyQuantities, ",") {
					if qty, err := strconv.ParseFloat(strings.TrimSpace(q), 64); err == nil {
						buyQuantities = append(buyQuantities, qty)
					}
				}
			}

			sellQuantities := []float64{}
			if s.SellQuantities != "" {
				for _, q := range strings.Split(s.SellQuantities, ",") {
					if qty, err := strconv.ParseFloat(strings.TrimSpace(q), 64); err == nil {
						sellQuantities = append(sellQuantities, qty)
					}
				}
			}

			buyDepthLevels := []int{}
			if s.BuyDepthLevels != "" {
				for _, d := range strings.Split(s.BuyDepthLevels, ",") {
					if lvl, err := strconv.Atoi(strings.TrimSpace(d)); err == nil {
						buyDepthLevels = append(buyDepthLevels, lvl)
					}
				}
			}

			sellDepthLevels := []int{}
			if s.SellDepthLevels != "" {
				for _, d := range strings.Split(s.SellDepthLevels, ",") {
					if lvl, err := strconv.Atoi(strings.TrimSpace(d)); err == nil {
						sellDepthLevels = append(sellDepthLevels, lvl)
					}
				}
			}

			// 新增：解析万分比
			buyBasisPoints := []float64{}
			if s.BuyBasisPoints != "" {
				for _, bp := range strings.Split(s.BuyBasisPoints, ",") {
					if basisPoint, err := strconv.ParseFloat(strings.TrimSpace(bp), 64); err == nil {
						buyBasisPoints = append(buyBasisPoints, basisPoint)
					}
				}
			}

			sellBasisPoints := []float64{}
			if s.SellBasisPoints != "" {
				for _, bp := range strings.Split(s.SellBasisPoints, ",") {
					if basisPoint, err := strconv.ParseFloat(strings.TrimSpace(bp), 64); err == nil {
						sellBasisPoints = append(sellBasisPoints, basisPoint)
					}
				}
			}

			formattedStrategies = append(formattedStrategies, map[string]interface{}{
				"id":                 s.ID,
				"symbol":             s.Symbol,
				"strategyType":       s.StrategyType,
				"side":               s.Side,
				"price":              s.Price,
				"totalQuantity":      s.TotalQuantity,
				"status":             s.Status,
				"enabled":            s.Enabled,
				"buyQuantities":      buyQuantities,
				"sellQuantities":     sellQuantities,
				"buyDepthLevels":     buyDepthLevels,
				"sellDepthLevels":    sellDepthLevels,
				"buyBasisPoints":     buyBasisPoints,  // 新增
				"sellBasisPoints":    sellBasisPoints, // 新增
				"pendingBatch":       s.PendingBatch,
				"cancelAfterMinutes": s.CancelAfterMinutes,
				"createdAt":          s.CreatedAt,
				"updatedAt":          s.UpdatedAt,
			})
		}

		c.JSON(http.StatusOK, gin.H{"strategies": formattedStrategies})
	}
}

// GinToggleStrategyHandler 切换策略状态处理器
func GinToggleStrategyHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		var req struct {
			ID uint `json:"id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
			return
		}

		var strategy models.Strategy
		if err := cfg.DB.Where("id = ? AND user_id = ? AND deleted_at IS NULL",
			req.ID, user.ID).First(&strategy).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "策略未找到"})
			return
		}

		// 切换启用状态
		strategy.Enabled = !strategy.Enabled
		if err := cfg.DB.Save(&strategy).Error; err != nil {
			log.Printf("切换策略状态失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "切换策略状态失败"})
			return
		}

		// 如果策略被禁用，取消所有待处理的订单
		if !strategy.Enabled {
			var orders []models.Order
			if err := cfg.DB.Where("strategy_id = ? AND status = ?", strategy.ID, "pending").
				Find(&orders).Error; err == nil && len(orders) > 0 {

				// 获取用户API密钥
				if user.APIKey != "" && user.SecretKey != "" {
					client := binance.NewClient(user.APIKey, user.SecretKey)
					for _, order := range orders {
						client.NewCancelOrderService().
							Symbol(order.Symbol).
							OrderID(order.OrderID).
							Do(context.Background())

						cfg.DB.Model(&order).Update("status", "cancelled")
					}
				}
			}

			// 重置pending_batch标志
			cfg.DB.Model(&strategy).Update("pending_batch", false)
		}

		log.Printf("策略 %d 状态切换为: %v", strategy.ID, strategy.Enabled)

		c.JSON(http.StatusOK, gin.H{
			"message": "策略状态切换成功",
			"enabled": strategy.Enabled,
		})
	}
}

// GinDeleteStrategyHandler 删除策略处理器
func GinDeleteStrategyHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := getUserFromGinContext(c, cfg)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
			return
		}

		var req struct {
			ID uint `json:"id" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据"})
			return
		}

		var strategy models.Strategy
		if err := cfg.DB.Where("id = ? AND user_id = ? AND deleted_at IS NULL",
			req.ID, user.ID).First(&strategy).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "策略未找到"})
			return
		}

		// 取消所有相关的待处理订单
		var orders []models.Order
		if err := cfg.DB.Where("strategy_id = ? AND status = ?", strategy.ID, "pending").
			Find(&orders).Error; err == nil && len(orders) > 0 {

			// 获取用户API密钥
			if user.APIKey != "" && user.SecretKey != "" {
				client := binance.NewClient(user.APIKey, user.SecretKey)
				for _, order := range orders {
					client.NewCancelOrderService().
						Symbol(order.Symbol).
						OrderID(order.OrderID).
						Do(context.Background())

					cfg.DB.Model(&order).Update("status", "cancelled")
				}
			}
		}

		// 软删除策略
		if err := cfg.DB.Delete(&strategy).Error; err != nil {
			log.Printf("删除策略失败: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "删除策略失败"})
			return
		}

		log.Printf("策略 %d 已删除", strategy.ID)

		c.JSON(http.StatusOK, gin.H{"message": "策略删除成功"})
	}
}

// GinListSymbolsHandler 获取交易对列表处理器
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
