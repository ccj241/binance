package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/adshao/go-binance/v2"
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
	"github.com/ccj241/binance/tasks"
)

// getUserFromRequest 从请求中获取用户信息
func getUserFromRequest(r *http.Request, cfg *config.Config) (*models.User, error) {
	// 从Gin中间件设置的header获取用户名
	username := r.Header.Get("username")
	if username == "" {
		return nil, fmt.Errorf("用户名未找到")
	}

	var user models.User
	if err := cfg.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

func PricesHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			log.Printf("PricesHandler 无效方法: %s", r.Method)
			http.Error(w, `{"error": "方法不允许"}`, http.StatusMethodNotAllowed)
			return
		}

		user, err := getUserFromRequest(r, cfg)
		if err != nil {
			log.Printf("获取用户失败: %v", err)
			http.Error(w, `{"error": "用户未找到"}`, http.StatusNotFound)
			return
		}

		priceMap := make(map[string]float64)
		tasks.PriceMonitor.Range(func(key, value any) bool {
			symbolUser, ok := key.(string)
			if !ok {
				log.Printf("PriceMonitor 键类型无效: %T", key)
				return true
			}
			price, ok := value.(float64)
			if !ok {
				log.Printf("PriceMonitor 值类型无效: %T", value)
				return true
			}
			parts := strings.Split(symbolUser, "|")
			if len(parts) != 2 {
				log.Printf("键格式无效: %s", symbolUser)
				return true
			}
			symbol, userIDStr := parts[0], parts[1]
			userID, err := strconv.ParseUint(userIDStr, 10, 32)
			if err != nil {
				log.Printf("键中的 userID 无效: %s, error: %v", userIDStr, err)
				return true
			}
			if uint(userID) == user.ID {
				priceMap[symbol] = price
			}
			return true
		})

		response := map[string]interface{}{
			"prices": priceMap,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("编码价格响应失败: %v", err)
		}
	}
}

func BalanceHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			log.Printf("BalanceHandler 无效方法: %s", r.Method)
			http.Error(w, `{"error": "方法不允许"}`, http.StatusMethodNotAllowed)
			return
		}

		user, err := getUserFromRequest(r, cfg)
		if err != nil {
			log.Printf("获取用户失败: %v", err)
			http.Error(w, `{"error": "用户未找到"}`, http.StatusNotFound)
			return
		}

		if user.APIKey == "" || user.SecretKey == "" {
			log.Printf("用户 %s 未设置 API 密钥", user.Username)
			http.Error(w, `{"error": "API 密钥未设置"}`, http.StatusBadRequest)
			return
		}

		client := binance.NewClient(user.APIKey, user.SecretKey)
		account, err := client.NewGetAccountService().Do(context.Background())
		if err != nil {
			log.Printf("获取余额失败，用户 %d: %v", user.ID, err)
			http.Error(w, `{"error": "获取余额失败"}`, http.StatusInternalServerError)
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

		response := map[string]interface{}{
			"balances": balances,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("编码余额响应失败: %v", err)
		}
	}
}

func TradesHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			log.Printf("TradesHandler 无效方法: %s", r.Method)
			http.Error(w, `{"error": "方法不允许"}`, http.StatusMethodNotAllowed)
			return
		}

		user, err := getUserFromRequest(r, cfg)
		if err != nil {
			log.Printf("获取用户失败: %v", err)
			http.Error(w, `{"error": "用户未找到"}`, http.StatusNotFound)
			return
		}

		var trades []models.Trade
		if err := cfg.DB.Where("user_id = ?", user.ID).Find(&trades).Error; err != nil {
			log.Printf("获取用户 %d 的交易记录失败: %v", user.ID, err)
			http.Error(w, `{"error": "获取交易记录失败"}`, http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"trades": trades,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("编码交易记录响应失败: %v", err)
		}
	}
}

func AddSymbolHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Printf("AddSymbolHandler 无效方法: %s", r.Method)
			http.Error(w, `{"error": "方法不允许"}`, http.StatusMethodNotAllowed)
			return
		}

		user, err := getUserFromRequest(r, cfg)
		if err != nil {
			log.Printf("获取用户失败: %v", err)
			http.Error(w, `{"error": "用户未找到"}`, http.StatusNotFound)
			return
		}

		var request struct {
			Symbol string `json:"symbol"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Printf("解码请求体失败: %v", err)
			http.Error(w, `{"error": "无效的请求体"}`, http.StatusBadRequest)
			return
		}

		if request.Symbol == "" {
			log.Printf("请求中 symbol 为空")
			http.Error(w, `{"error": "需要提供 symbol"}`, http.StatusBadRequest)
			return
		}

		// 检查是否已存在
		var existingSymbol models.CustomSymbol
		if err := cfg.DB.Where("user_id = ? AND symbol = ?", user.ID, request.Symbol).First(&existingSymbol).Error; err == nil {
			// 已存在，直接返回成功
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"message": "Symbol 已存在"}); err != nil {
				log.Printf("编码响应失败: %v", err)
			}
			return
		}

		symbol := models.CustomSymbol{
			UserID: user.ID,
			Symbol: strings.ToUpper(request.Symbol),
		}
		if err := cfg.DB.Create(&symbol).Error; err != nil {
			log.Printf("为用户 %d 添加 symbol %s 失败: %v", user.ID, request.Symbol, err)
			http.Error(w, `{"error": "添加 symbol 失败"}`, http.StatusInternalServerError)
			return
		}

		tasks.MonitorNewSymbol(strings.ToUpper(request.Symbol), user.ID, cfg)
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"message": "Symbol 添加成功"}); err != nil {
			log.Printf("编码响应失败: %v", err)
		}
		log.Printf("为用户 %d 添加 symbol %s", user.ID, request.Symbol)
	}
}

func WithdrawalHistoryHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			log.Printf("WithdrawalHistoryHandler 无效方法: %s", r.Method)
			http.Error(w, `{"error": "方法不允许"}`, http.StatusMethodNotAllowed)
			return
		}

		user, err := getUserFromRequest(r, cfg)
		if err != nil {
			log.Printf("获取用户失败: %v", err)
			http.Error(w, `{"error": "用户未找到"}`, http.StatusNotFound)
			return
		}

		var history []models.WithdrawalHistory
		if err := cfg.DB.Where("user_id = ?", user.ID).Find(&history).Error; err != nil {
			log.Printf("获取用户 %d 的取款历史失败: %v", user.ID, err)
			http.Error(w, `{"error": "获取取款历史失败"}`, http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"history": history,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("编码取款历史响应失败: %v", err)
		}
		log.Printf("为用户 %d 返回 %d 条取款历史记录", user.ID, len(history))
	}
}
