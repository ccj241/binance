package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
	"github.com/ccj241/binance/tasks"
)

func CreateStrategyHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Printf("CreateStrategyHandler 无效方法: %s", r.Method)
			http.Error(w, `{"error": "方法不允许"}`, http.StatusMethodNotAllowed)
			return
		}
		username := r.Header.Get("username")
		if username == "" {
			log.Printf("请求头缺少 username")
			http.Error(w, `{"error": "未授权"}`, http.StatusUnauthorized)
			return
		}
		log.Printf("为用户 %s 创建策略", username)
		var user models.User
		if err := cfg.DB.Where("username = ?", username).First(&user).Error; err != nil {
			log.Printf("用户未找到: %s, error: %v", username, err)
			http.Error(w, `{"error": "用户未找到"}`, http.StatusNotFound)
			return
		}
		var strategyReq struct {
			Symbol          string    `json:"symbol"`
			StrategyType    string    `json:"strategyType"`
			Side            string    `json:"side"`
			Price           float64   `json:"price"`
			TotalQuantity   float64   `json:"totalQuantity"`
			BuyQuantities   []float64 `json:"buyQuantities"`
			SellQuantities  []float64 `json:"sellQuantities"`
			BuyDepthLevels  []int     `json:"buyDepthLevels"`
			SellDepthLevels []int     `json:"sellDepthLevels"`
		}
		if err := json.NewDecoder(r.Body).Decode(&strategyReq); err != nil {
			log.Printf("JSON 解码错误: %v", err)
			http.Error(w, `{"error": "无效的请求"}`, http.StatusBadRequest)
			return
		}
		log.Printf("收到策略请求: %+v", strategyReq)
		if strategyReq.StrategyType != "simple" && strategyReq.StrategyType != "iceberg" && strategyReq.StrategyType != "custom" {
			http.Error(w, `{"error": "无效的策略类型"}`, http.StatusBadRequest)
			return
		}
		if strategyReq.Side != "BUY" && strategyReq.Side != "SELL" {
			http.Error(w, `{"error": "无效的方向"}`, http.StatusBadRequest)
			return
		}
		// 设置默认值并验证
		if strategyReq.StrategyType == "simple" {
			if strategyReq.Side == "BUY" {
				if len(strategyReq.BuyQuantities) == 0 {
					strategyReq.BuyQuantities = []float64{1.0}
					strategyReq.BuyDepthLevels = []int{1}
				}
				strategyReq.SellQuantities = []float64{}
				strategyReq.SellDepthLevels = []int{}
			} else {
				if len(strategyReq.SellQuantities) == 0 {
					strategyReq.SellQuantities = []float64{1.0}
					strategyReq.SellDepthLevels = []int{1}
				}
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
		} else if strategyReq.StrategyType == "custom" {
			if strategyReq.Side == "BUY" {
				if len(strategyReq.BuyQuantities) == 0 || len(strategyReq.BuyDepthLevels) == 0 {
					http.Error(w, `{"error": "自定义策略需要提供买入数量和深度级别"}`, http.StatusBadRequest)
					return
				}
				strategyReq.SellQuantities = []float64{}
				strategyReq.SellDepthLevels = []int{}
			} else {
				if len(strategyReq.SellQuantities) == 0 || len(strategyReq.SellDepthLevels) == 0 {
					http.Error(w, `{"error": "自定义策略需要提供卖出数量和深度级别"}`, http.StatusBadRequest)
					return
				}
				strategyReq.BuyQuantities = []float64{}
				strategyReq.BuyDepthLevels = []int{}
			}
		}
		if len(strategyReq.BuyQuantities) != len(strategyReq.BuyDepthLevels) || len(strategyReq.SellQuantities) != len(strategyReq.SellDepthLevels) {
			http.Error(w, `{"error": "数量和深度级别长度不匹配"}`, http.StatusBadRequest)
			return
		}
		buyQuantitiesStr := ""
		if len(strategyReq.BuyQuantities) > 0 {
			strs := make([]string, len(strategyReq.BuyQuantities))
			for i, q := range strategyReq.BuyQuantities {
				if q <= 0 {
					http.Error(w, `{"error": "数量必须为正"}`, http.StatusBadRequest)
					return
				}
				strs[i] = fmt.Sprintf("%.8f", q)
			}
			buyQuantitiesStr = strings.Join(strs, ",")
		}
		sellQuantitiesStr := ""
		if len(strategyReq.SellQuantities) > 0 {
			strs := make([]string, len(strategyReq.SellQuantities))
			for i, q := range strategyReq.SellQuantities {
				if q <= 0 {
					http.Error(w, `{"error": "数量必须为正"}`, http.StatusBadRequest)
					return
				}
				strs[i] = fmt.Sprintf("%.8f", q)
			}
			sellQuantitiesStr = strings.Join(strs, ",")
		}
		buyDepthLevelsStr := ""
		if len(strategyReq.BuyDepthLevels) > 0 {
			strs := make([]string, len(strategyReq.BuyDepthLevels))
			for i, d := range strategyReq.BuyDepthLevels {
				if d <= 0 {
					http.Error(w, `{"error": "深度级别必须为正"}`, http.StatusBadRequest)
					return
				}
				strs[i] = fmt.Sprintf("%d", d)
			}
			buyDepthLevelsStr = strings.Join(strs, ",")
		}
		sellDepthLevelsStr := ""
		if len(strategyReq.SellDepthLevels) > 0 {
			strs := make([]string, len(strategyReq.SellDepthLevels))
			for i, d := range strategyReq.SellDepthLevels {
				if d <= 0 {
					http.Error(w, `{"error": "深度级别必须为正"}`, http.StatusBadRequest)
					return
				}
				strs[i] = fmt.Sprintf("%d", d)
			}
			sellDepthLevelsStr = strings.Join(strs, ",")
		}
		dbStrategy := models.Strategy{
			UserID:          user.ID,
			Symbol:          strings.ToUpper(strategyReq.Symbol),
			StrategyType:    strategyReq.StrategyType,
			Side:            strategyReq.Side,
			Price:           strategyReq.Price,
			TotalQuantity:   strategyReq.TotalQuantity,
			Status:          "active",
			Enabled:         true,
			BuyQuantities:   buyQuantitiesStr,
			SellQuantities:  sellQuantitiesStr,
			BuyDepthLevels:  buyDepthLevelsStr,
			SellDepthLevels: sellDepthLevelsStr,
		}
		if err := cfg.DB.Create(&dbStrategy).Error; err != nil {
			log.Printf("创建策略失败: %v", err)
			http.Error(w, `{"error": "创建策略失败"}`, http.StatusInternalServerError)
			return
		}
		log.Printf("策略创建成功: ID=%d, Symbol=%s, Type=%s, Side=%s, UserID=%d, BuyQuantities='%s', SellQuantities='%s', BuyDepthLevels='%s', SellDepthLevels='%s'",
			dbStrategy.ID, dbStrategy.Symbol, dbStrategy.StrategyType, dbStrategy.Side, user.ID, buyQuantitiesStr, sellQuantitiesStr, buyDepthLevelsStr, sellDepthLevelsStr)
		// 确保传递 cfg 参数以修复错误
		tasks.MonitorNewSymbol(dbStrategy.Symbol, user.ID, cfg)
		json.NewEncoder(w).Encode(map[string]interface{}{"message": "策略创建成功", "strategyId": dbStrategy.ID})
	}
}

func ListStrategiesHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			log.Printf("ListStrategiesHandler 无效方法: %s", r.Method)
			http.Error(w, `{"error": "方法不允许"}`, http.StatusMethodNotAllowed)
			return
		}
		username := r.Header.Get("username")
		if username == "" {
			log.Printf("请求头缺少 username")
			http.Error(w, `{"error": "未授权"}`, http.StatusUnauthorized)
			return
		}
		log.Printf("为用户 %s 列出策略", username)
		var user models.User
		if err := cfg.DB.Where("username = ?", username).First(&user).Error; err != nil {
			log.Printf("用户未找到: %s, error: %v", username, err)
			http.Error(w, `{"error": "用户未找到"}`, http.StatusNotFound)
			return
		}
		var strategies []models.Strategy
		if err := cfg.DB.Where("user_id = ? AND deleted_at IS NULL", user.ID).Find(&strategies).Error; err != nil {
			log.Printf("获取用户 %d 的策略失败: %v", user.ID, err)
			http.Error(w, `{"error": "获取策略失败"}`, http.StatusInternalServerError)
			return
		}
		log.Printf("为用户 %d 找到 %d 个策略: %+v", user.ID, len(strategies), strategies)
		response := map[string]interface{}{
			"strategies": make([]map[string]interface{}, len(strategies)),
		}
		for i, s := range strategies {
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
			response["strategies"].([]map[string]interface{})[i] = map[string]interface{}{
				"id":              s.ID,
				"symbol":          s.Symbol,
				"strategyType":    s.StrategyType,
				"side":            s.Side,
				"price":           s.Price,
				"totalQuantity":   s.TotalQuantity,
				"status":          s.Status,
				"enabled":         s.Enabled,
				"buyQuantities":   buyQuantities,
				"sellQuantities":  sellQuantities,
				"buyDepthLevels":  buyDepthLevels,
				"sellDepthLevels": sellDepthLevels,
				"createdAt":       s.CreatedAt.Format(time.RFC3339),
				"updatedAt":       s.UpdatedAt.Format(time.RFC3339),
			}
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("为用户 %d 编码响应失败: %v", user.ID, err)
			http.Error(w, `{"error": "编码响应失败"}`, http.StatusInternalServerError)
			return
		}
		log.Printf("为用户 %d 成功返回 %d 个策略", user.ID, len(strategies))
	}
}

func ToggleStrategyHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Printf("ToggleStrategyHandler 无效方法: %s", r.Method)
			http.Error(w, `{"error": "方法不允许"}`, http.StatusMethodNotAllowed)
			return
		}
		username := r.Header.Get("username")
		if username == "" {
			log.Printf("请求头缺少 username")
			http.Error(w, `{"error": "未授权"}`, http.StatusUnauthorized)
			return
		}
		var user models.User
		if err := cfg.DB.Where("username = ?", username).First(&user).Error; err != nil {
			log.Printf("用户未找到: %s", username)
			http.Error(w, `{"error": "用户未找到"}`, http.StatusNotFound)
			return
		}
		var req struct {
			ID uint `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("JSON 解码错误: %v", err)
			http.Error(w, `{"error": "无效的请求"}`, http.StatusBadRequest)
			return
		}
		var strategy models.Strategy
		if err := cfg.DB.Where("id = ? AND user_id = ? AND deleted_at IS NULL", req.ID, user.ID).First(&strategy).Error; err != nil {
			log.Printf("策略未找到: ID=%d, UserID=%d", req.ID, user.ID)
			http.Error(w, `{"error": "策略未找到"}`, http.StatusNotFound)
			return
		}
		strategy.Enabled = !strategy.Enabled
		if err := cfg.DB.Save(&strategy).Error; err != nil {
			log.Printf("切换策略 %d 失败: %v", strategy.ID, err)
			http.Error(w, `{"error": "切换策略失败"}`, http.StatusInternalServerError)
			return
		}
		log.Printf("策略切换成功: ID=%d, Enabled=%v", strategy.ID, strategy.Enabled)
		json.NewEncoder(w).Encode(map[string]interface{}{"message": "策略切换成功", "enabled": strategy.Enabled})
	}
}

func DeleteStrategyHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodDelete {
			log.Printf("DeleteStrategyHandler 无效方法: %s", r.Method)
			http.Error(w, `{"error": "方法不允许"}`, http.StatusMethodNotAllowed)
			return
		}
		username := r.Header.Get("username")
		if username == "" {
			log.Printf("请求头缺少 username")
			http.Error(w, `{"error": "未授权"}`, http.StatusUnauthorized)
			return
		}
		var user models.User
		if err := cfg.DB.Where("username = ?", username).First(&user).Error; err != nil {
			log.Printf("用户未找到: %s", username)
			http.Error(w, `{"error": "用户未找到"}`, http.StatusNotFound)
			return
		}
		var req struct {
			ID uint `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("JSON 解码错误: %v", err)
			http.Error(w, `{"error": "无效的请求"}`, http.StatusBadRequest)
			return
		}
		var strategy models.Strategy
		if err := cfg.DB.Where("id = ? AND user_id = ? AND deleted_at IS NULL", req.ID, user.ID).First(&strategy).Error; err != nil {
			log.Printf("策略未找到: ID=%d, UserID=%d", req.ID, user.ID)
			http.Error(w, `{"error": "策略未找到"}`, http.StatusNotFound)
			return
		}
		if err := cfg.DB.Delete(&strategy).Error; err != nil {
			log.Printf("删除策略 %d 失败: %v", strategy.ID, err)
			http.Error(w, `{"error": "删除策略失败"}`, http.StatusInternalServerError)
			return
		}
		log.Printf("策略删除成功: ID=%d", strategy.ID)
		json.NewEncoder(w).Encode(map[string]interface{}{"message": "策略删除成功"})
	}
}

func ListSymbolsHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			log.Printf("ListSymbolsHandler 无效方法: %s", r.Method)
			http.Error(w, `{"error": "方法不允许"}`, http.StatusMethodNotAllowed)
			return
		}
		username := r.Header.Get("username")
		if username == "" {
			log.Printf("请求头缺少 username")
			http.Error(w, `{"error": "未授权"}`, http.StatusUnauthorized)
			return
		}
		log.Printf("为用户 %s 列出交易对", username)
		var user models.User
		if err := cfg.DB.Where("username = ?", username).First(&user).Error; err != nil {
			log.Printf("用户未找到: %s, error: %v", username, err)
			http.Error(w, `{"error": "用户未找到"}`, http.StatusNotFound)
			return
		}
		var symbols []models.CustomSymbol
		if err := cfg.DB.Where("user_id = ? AND deleted_at IS NULL", user.ID).Find(&symbols).Error; err != nil {
			log.Printf("获取用户 %d 的交易对失败: %v", user.ID, err)
			http.Error(w, `{"error": "获取交易对失败"}`, http.StatusInternalServerError)
			return
		}
		response := map[string]interface{}{
			"symbols": make([]string, len(symbols)),
		}
		for i, s := range symbols {
			response["symbols"].([]string)[i] = s.Symbol
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("编码交易对响应失败: %v", err)
			http.Error(w, `{"error": "编码响应失败"}`, http.StatusInternalServerError)
			return
		}
		log.Printf("为用户 %d 成功返回 %d 个交易对", user.ID, len(symbols))
	}
}
