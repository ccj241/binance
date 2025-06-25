package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
	"github.com/gorilla/mux"
)

func OrdersHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.Header.Get("username")
		var user models.User
		if err := cfg.DB.Where("username = ?", username).First(&user).Error; err != nil || user.APIKey == "" || user.SecretKey == "" {
			log.Printf("用户未找到或未设置 API 密钥: %s", username)
			http.Error(w, `{"error": "API 密钥未设置"}`, http.StatusBadRequest)
			return
		}
		client := binance.NewClient(user.APIKey, user.SecretKey)
		orders, err := client.NewListOpenOrdersService().Do(context.Background())
		if err != nil {
			log.Printf("获取订单失败: %v", err)
			http.Error(w, fmt.Sprintf(`{"error": "获取订单失败: %v"}`, err), http.StatusInternalServerError)
			return
		}
		for _, o := range orders {
			price, _ := strconv.ParseFloat(o.Price, 64)
			quantity, _ := strconv.ParseFloat(o.OrigQuantity, 64)
			dbOrder := models.Order{
				UserID:      user.ID,
				Symbol:      o.Symbol,
				Side:        string(o.Side),
				Price:       price,
				Quantity:    quantity,
				OrderID:     o.OrderID,
				Status:      "pending",
				CancelAfter: time.Now().Add(2 * time.Hour),
			}
			if err := cfg.DB.FirstOrCreate(&dbOrder, models.Order{OrderID: o.OrderID, UserID: user.ID}).Error; err != nil {
				log.Printf("同步订单 %d 失败: %v", o.OrderID, err)
			}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"orders": orders})
	}
}

func CancelledOrdersHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.Header.Get("username")
		var user models.User
		if err := cfg.DB.Where("username = ?", username).First(&user).Error; err != nil {
			log.Printf("用户未找到: %s", username)
			http.Error(w, `{"error": "用户未找到"}`, http.StatusNotFound)
			return
		}
		var orders []models.Order
		if err := cfg.DB.Where("user_id = ? AND status = ?", user.ID, "cancelled").Find(&orders).Error; err != nil {
			log.Printf("获取用户 %d 的已取消订单失败: %v", user.ID, err)
			http.Error(w, `{"error": "获取已取消订单失败"}`, http.StatusInternalServerError)
			return
		}
		log.Printf("为用户 %d 获取 %d 个已取消订单", user.ID, len(orders))
		json.NewEncoder(w).Encode(map[string]interface{}{"orders": orders})
	}
}

func CancelOrderHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.Header.Get("username")
		var user models.User
		if err := cfg.DB.Where("username = ?", username).First(&user).Error; err != nil || user.APIKey == "" || user.SecretKey == "" {
			log.Printf("用户未找到或未设置 API 密钥: %s", username)
			http.Error(w, `{"error": "API 密钥未设置"}`, http.StatusBadRequest)
			return
		}
		vars := mux.Vars(r)
		orderIDStr := vars["orderId"]
		orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
		if err != nil {
			log.Printf("无效的订单 ID: %s, error: %v", orderIDStr, err)
			http.Error(w, `{"error": "无效的订单 ID"}`, http.StatusBadRequest)
			return
		}
		var order models.Order
		if err := cfg.DB.Where("order_id = ? AND user_id = ?", orderID, user.ID).First(&order).Error; err != nil {
			log.Printf("订单未找到: OrderID=%d, UserID=%d, error: %v", orderID, user.ID, err)
			http.Error(w, `{"error": "订单未找到"}`, http.StatusNotFound)
			return
		}
		if order.Symbol == "" {
			log.Printf("订单 symbol 为空: OrderID=%d, UserID=%d", orderID, user.ID)
			http.Error(w, `{"error": "订单 symbol 为空"}`, http.StatusBadRequest)
			return
		}
		client := binance.NewClient(user.APIKey, user.SecretKey)
		_, err = client.NewCancelOrderService().Symbol(order.Symbol).OrderID(order.OrderID).Do(context.Background())
		if err != nil {
			if strings.Contains(err.Error(), "Order does not exist") {
				order.Status = "cancelled"
				if err := cfg.DB.Save(&order).Error; err != nil {
					log.Printf("更新订单状态失败: %v", err)
				}
				json.NewEncoder(w).Encode(map[string]string{"message": "订单已取消或不存在"})
				return
			}
			log.Printf("取消订单失败: OrderID=%d, Symbol=%s, error: %v", order.OrderID, order.Symbol, err)
			http.Error(w, fmt.Sprintf(`{"error": "取消订单失败: %v"}`, err), http.StatusInternalServerError)
			return
		}
		order.Status = "cancelled"
		if err := cfg.DB.Save(&order).Error; err != nil {
			log.Printf("更新订单状态失败: %v", err)
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "订单已取消"})
	}
}

func CreateOrderHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.Header.Get("username")
		var user models.User
		if err := cfg.DB.Where("username = ?", username).First(&user).Error; err != nil || user.APIKey == "" || user.SecretKey == "" {
			log.Printf("用户未找到或未设置 API 密钥: %s", username)
			http.Error(w, `{"error": "API 密钥未设置"}`, http.StatusBadRequest)
			return
		}
		var orderReq struct {
			Symbol   string  `json:"symbol"`
			Side     string  `json:"side"`
			Quantity float64 `json:"quantity"`
			Price    float64 `json:"price"`
		}
		if err := json.NewDecoder(r.Body).Decode(&orderReq); err != nil {
			log.Printf("JSON 解码错误: %v", err)
			http.Error(w, `{"error": "无效的请求"}`, http.StatusBadRequest)
			return
		}
		client := binance.NewClient(user.APIKey, user.SecretKey)
		order, err := client.NewCreateOrderService().
			Symbol(orderReq.Symbol).
			Side(binance.SideType(orderReq.Side)).
			Type(binance.OrderTypeLimit).
			TimeInForce(binance.TimeInForceTypeGTC).
			Quantity(fmt.Sprintf("%.8f", orderReq.Quantity)).
			Price(fmt.Sprintf("%.8f", orderReq.Price)).
			Do(context.Background())
		if err != nil {
			log.Printf("下单失败: %v", err)
			http.Error(w, fmt.Sprintf(`{"error": "下单失败: %v"}`, err), http.StatusInternalServerError)
			return
		}
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
			http.Error(w, `{"error": "保存订单失败"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"message": "订单已下单", "orderId": order.OrderID})
	}
}
