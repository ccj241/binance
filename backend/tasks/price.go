package tasks

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
)

var PriceMonitor sync.Map
var MonitoredSymbols sync.Map
var strategyLocks sync.Map

// MonitorNewSymbol 启动对新交易对的监控
func MonitorNewSymbol(symbol string, userID uint, cfg *config.Config) {
	if _, loaded := MonitoredSymbols.LoadOrStore(symbol+"|"+fmt.Sprint(userID), true); !loaded {
		log.Printf("Starting price monitoring for symbol: %s, userID: %d", symbol, userID)
		go monitorPrice(symbol, userID, cfg)
	}
}

// StartPriceMonitoring 开始监控价格
func StartPriceMonitoring(cfg *config.Config) {
	if cfg.DB == nil {
		log.Println("Database not initialized, skipping price monitoring")
		return
	}
	var symbols []models.CustomSymbol
	if err := cfg.DB.Select("DISTINCT symbol, user_id").Where("deleted_at IS NULL").Find(&symbols).Error; err != nil {
		log.Printf("Failed to fetch custom symbols: %v", err)
		return
	}
	for _, symbol := range symbols {
		go monitorPrice(symbol.Symbol, symbol.UserID, cfg)
	}
}

// monitorPrice 监控单个交易对的价格
func monitorPrice(symbol string, userID uint, cfg *config.Config) {
	var user models.User
	if err := cfg.DB.First(&user, userID).Error; err != nil {
		log.Printf("User not found: ID=%d", userID)
		return
	}
	if user.APIKey == "" || user.SecretKey == "" {
		log.Printf("API Key not set for user %d, skipping monitorPrice for %s", user.ID, symbol)
		return
	}
	client := binance.NewClient(user.APIKey, user.SecretKey)
	lastLogTime := time.Now()
	lastDepthLogTime := time.Now()
	logInterval := 60 * time.Second
	wsDepthHandler := func(event *binance.WsDepthEvent) {
		if len(event.Asks) < 3 || len(event.Bids) < 3 {
			if time.Since(lastDepthLogTime) >= logInterval {
				log.Printf("Insufficient depth data for %s: Asks=%d, Bids=%d", symbol, len(event.Asks), len(event.Bids))
				lastDepthLogTime = time.Now()
			}
			return
		}
		asks := event.Asks[:3]
		bids := event.Bids[:3]
		currentPrice, err := strconv.ParseFloat(event.Asks[0].Price, 64)
		if err != nil {
			log.Printf("Error parsing price for %s: %v", symbol, err)
			return
		}
		PriceMonitor.Store(symbol+"|"+fmt.Sprint(userID), currentPrice)
		price := models.Price{
			Symbol:    symbol,
			Price:     fmt.Sprintf("%.8f", currentPrice),
			UpdatedAt: time.Now(),
		}
		if err := cfg.DB.Where("symbol = ?", symbol).Assign(price).FirstOrCreate(&price).Error; err != nil {
			log.Printf("Failed to save price for %s: %v", symbol, err)
		}
		if time.Since(lastLogTime) >= logInterval {
			log.Printf("Price updated for %s (user %d): %.2f", symbol, userID, currentPrice)
			lastLogTime = time.Now()
		}

		var strategies []models.Strategy
		if err := cfg.DB.Where("user_id = ? AND symbol = ? AND status = ? AND enabled = ? AND deleted_at IS NULL", userID, symbol, "active", true).Find(&strategies).Error; err != nil {
			log.Printf("Error fetching strategies for %s (user %d): %v", symbol, userID, err)
			return
		}
		for _, strategy := range strategies {
			lock, _ := strategyLocks.LoadOrStore(fmt.Sprintf("%d", strategy.ID), &sync.Mutex{})
			mutex := lock.(*sync.Mutex)
			if !mutex.TryLock() {
				continue
			}
			defer mutex.Unlock()

			var orders []models.Order
			if err := cfg.DB.Where("strategy_id = ? AND status = ?", strategy.ID, "pending").Find(&orders).Error; err != nil {
				log.Printf("Error fetching orders for strategy %d: %v", strategy.ID, err)
				continue
			}
			for _, order := range orders {
				binanceOrder, err := client.NewGetOrderService().Symbol(order.Symbol).OrderID(order.OrderID).Do(context.Background())
				if err != nil {
					log.Printf("Failed to fetch Binance order %d for strategy %d: %v", order.OrderID, strategy.ID, err)
					continue
				}
				if binanceOrder.Status == binance.OrderStatusTypeFilled {
					order.Status = "filled"
				} else if binanceOrder.Status == binance.OrderStatusTypePendingCancel || binanceOrder.Status == binance.OrderStatusTypeExpired {
					order.Status = "cancelled"
				}
				if order.Status != "pending" {
					if err := cfg.DB.Save(&order).Error; err != nil {
						log.Printf("Failed to update order %d status to %s: %v", order.OrderID, order.Status, err)
					}
				}
			}

			if err := cfg.DB.Where("strategy_id = ? AND status NOT IN (?, ?)", strategy.ID, "filled", "cancelled").Find(&orders).Error; err != nil {
				log.Printf("Error fetching orders for strategy %d: %v", strategy.ID, err)
				continue
			}
			if len(orders) == 0 && strategy.PendingBatch {
				if err := cfg.DB.Model(&strategy).Update("pending_batch", false).Error; err != nil {
					log.Printf("Failed to reset PendingBatch for strategy %d: %v", strategy.ID, err)
				} else {
					log.Printf("Reset PendingBatch for strategy %d", strategy.ID)
				}
			}
		}

		if err := cfg.DB.Where("user_id = ? AND symbol = ? AND status = ? AND enabled = ? AND deleted_at IS NULL AND pending_batch = ?", userID, symbol, "active", true, false).Find(&strategies).Error; err != nil {
			log.Printf("Error fetching strategies for %s (user %d): %v", symbol, userID, err)
			return
		}
		for _, strategy := range strategies {
			lock, _ := strategyLocks.LoadOrStore(fmt.Sprintf("%d", strategy.ID), &sync.Mutex{})
			mutex := lock.(*sync.Mutex)
			if !mutex.TryLock() {
				continue
			}
			defer mutex.Unlock()

			if strategy.PendingBatch {
				log.Printf("Strategy %d has pending batch, skipping", strategy.ID)
				continue
			}

			if err := cfg.DB.Model(&strategy).Update("pending_batch", true).Error; err != nil {
				log.Printf("Failed to set PendingBatch for strategy %d: %v", strategy.ID, err)
				continue
			}

			if strategy.Side == "SELL" && currentPrice >= strategy.Price {
				if err := placeOrders(client, strategy, userID, asks, "SELL", cfg); err != nil {
					log.Printf("Failed to place SELL orders for strategy %d: %v", strategy.ID, err)
					if err := cfg.DB.Model(&strategy).Update("pending_batch", false).Error; err != nil {
						log.Printf("Failed to reset PendingBatch for strategy %d: %v", strategy.ID, err)
					}
				}
			} else if strategy.Side == "BUY" && currentPrice <= strategy.Price {
				if err := placeOrders(client, strategy, userID, bids, "BUY", cfg); err != nil {
					log.Printf("Failed to place BUY orders for strategy %d: %v", strategy.ID, err)
					if err := cfg.DB.Model(&strategy).Update("pending_batch", false).Error; err != nil {
						log.Printf("Failed to reset PendingBatch for strategy %d: %v", strategy.ID, err)
					}
				}
			}
		}
	}
	wsErrHandler := func(err error) {
		log.Printf("WebSocket error for %s (user %d): %v", symbol, userID, err)
	}
	doneC, _, err := binance.WsDepthServe(symbol, wsDepthHandler, wsErrHandler)
	if err != nil {
		log.Printf("Failed to start WebSocket for %s (user %d): %v", symbol, userID, err)
		time.Sleep(5 * time.Second)
		monitorPrice(symbol, userID, cfg)
		return
	}
	<-doneC
}

// placeOrders 放置订单
func placeOrders(client *binance.Client, strategy models.Strategy, userID uint, levels interface{}, side string, cfg *config.Config) error {
	var quantities []float64
	var depthLevels []float64
	var placedOrders []models.Order

	log.Printf("Processing %s strategy: ID=%d, Type=%s, SellQuantities='%s', SellDepthLevels='%s', BuyQuantities='%s', BuyDepthLevels='%s'",
		side, strategy.ID, strategy.StrategyType, strategy.SellQuantities, strategy.SellDepthLevels, strategy.BuyQuantities, strategy.BuyDepthLevels)

	// 获取交易所信息
	exchangeInfo, err := client.NewExchangeInfoService().Symbol(strategy.Symbol).Do(context.Background())
	if err != nil {
		log.Printf("Failed to fetch exchange info for %s: %v", strategy.Symbol, err)
		return fmt.Errorf("failed to fetch exchange info: %v", err)
	}
	var pricePrecision, quantityPrecision int
	var minNotional float64
	for _, symbolInfo := range exchangeInfo.Symbols {
		if symbolInfo.Symbol == strategy.Symbol {
			for _, filter := range symbolInfo.Filters {
				switch filter["filterType"] {
				case "PRICE_FILTER":
					if tickSize, ok := filter["tickSize"].(string); ok {
						if ts, err := strconv.ParseFloat(tickSize, 64); err == nil {
							pricePrecision = int(-math.Log10(ts))
						}
					}
				case "LOT_SIZE":
					if stepSize, ok := filter["stepSize"].(string); ok {
						if ss, err := strconv.ParseFloat(stepSize, 64); err == nil {
							quantityPrecision = int(-math.Log10(ss))
						}
					}
				case "NOTIONAL":
					if minNot, ok := filter["minNotional"].(string); ok {
						minNotional, _ = strconv.ParseFloat(minNot, 64)
					}
				}
			}
			break
		}
	}
	log.Printf("Market %s: pricePrecision=%d, quantityPrecision=%d, minNotional=%.2f", strategy.Symbol, pricePrecision, quantityPrecision, minNotional)

	// 验证 TotalQuantity
	if strategy.TotalQuantity <= 0 {
		log.Printf("Invalid TotalQuantity %.2f for strategy %d", strategy.TotalQuantity, strategy.ID)
		return fmt.Errorf("invalid total quantity")
	}

	// 处理 Quantities 和 DepthLevels
	if side == "SELL" {
		if strategy.SellQuantities == "" || strategy.SellDepthLevels == "" {
			log.Printf("Empty SellQuantities or SellDepthLevels for strategy %d, using defaults", strategy.ID)
			quantities = []float64{1.0}
			depthLevels = []float64{0.0}
		} else {
			quantities, depthLevels, err = parseQuantitiesAndDepthLevels(strategy.SellQuantities, strategy.SellDepthLevels, strategy.ID)
			if err != nil {
				log.Printf("Failed to parse SELL quantities/depth for strategy %d: %v", strategy.ID, err)
				return err
			}
		}
	} else {
		if strategy.BuyQuantities == "" || strategy.BuyDepthLevels == "" {
			log.Printf("Empty BuyQuantities or BuyDepthLevels for strategy %d, using defaults", strategy.ID)
			quantities = []float64{1.0}
			depthLevels = []float64{0.0}
		} else {
			quantities, depthLevels, err = parseQuantitiesAndDepthLevels(strategy.BuyQuantities, strategy.BuyDepthLevels, strategy.ID)
			if err != nil {
				log.Printf("Failed to parse BUY quantities/depth for strategy %d: %v", strategy.ID, err)
				return err
			}
		}
	}

	// 处理价格水平
	var priceLevels []struct {
		Price    string
		Quantity string
	}
	if strategy.StrategyType == "iceberg" {
		var basePrice float64
		if side == "SELL" {
			asks, ok := levels.([]binance.Ask)
			if !ok || len(asks) == 0 {
				log.Printf("Invalid or empty asks for SELL: %T, len=%d", levels, len(asks))
				return fmt.Errorf("invalid asks")
			}
			basePrice, err = strconv.ParseFloat(asks[0].Price, 64)
			if err != nil {
				log.Printf("Failed to parse ask price for strategy %d: %v", strategy.ID, err)
				return fmt.Errorf("invalid ask price")
			}
			factors := []float64{1.0, 0.9999, 0.9997, 0.9995, 0.9993}
			for i := 0; i < len(quantities) && i < len(factors); i++ {
				price := basePrice * factors[i]
				priceLevels = append(priceLevels, struct {
					Price    string
					Quantity string
				}{Price: fmt.Sprintf("%.*f", pricePrecision, price)})
			}
		} else {
			bids, ok := levels.([]binance.Bid)
			if !ok || len(bids) == 0 {
				log.Printf("Invalid or empty bids for BUY: %T, len=%d", levels, len(bids))
				return fmt.Errorf("invalid bids")
			}
			basePrice, err = strconv.ParseFloat(bids[0].Price, 64)
			if err != nil {
				log.Printf("Failed to parse bid price for strategy %d: %v", strategy.ID, err)
				return fmt.Errorf("invalid bid price")
			}
			factors := []float64{1.0, 1.0001, 1.0003, 1.0005, 1.0007}
			for i := 0; i < len(quantities) && i < len(factors); i++ {
				price := basePrice * factors[i]
				priceLevels = append(priceLevels, struct {
					Price    string
					Quantity string
				}{Price: fmt.Sprintf("%.*f", pricePrecision, price)})
			}
		}
	} else if strategy.StrategyType == "custom" {
		var basePrice float64
		if side == "SELL" {
			asks, ok := levels.([]binance.Ask)
			if !ok || len(asks) == 0 {
				log.Printf("Invalid or empty asks for SELL: %T, len=%d", levels, len(asks))
				return fmt.Errorf("invalid asks")
			}
			basePrice, err = strconv.ParseFloat(asks[0].Price, 64)
			if err != nil {
				log.Printf("Failed to parse ask price for strategy %d: %v", strategy.ID, err)
				return fmt.Errorf("invalid ask price")
			}
			for i := 0; i < len(quantities) && i < len(depthLevels); i++ {
				price := basePrice * (1 - depthLevels[i])
				priceLevels = append(priceLevels, struct {
					Price    string
					Quantity string
				}{Price: fmt.Sprintf("%.*f", pricePrecision, price)})
			}
		} else {
			bids, ok := levels.([]binance.Bid)
			if !ok || len(bids) == 0 {
				log.Printf("Invalid or empty bids for BUY: %T, len=%d", levels, len(bids))
				return fmt.Errorf("invalid bids")
			}
			basePrice, err = strconv.ParseFloat(bids[0].Price, 64)
			if err != nil {
				log.Printf("Failed to parse bid price for strategy %d: %v", strategy.ID, err)
				return fmt.Errorf("invalid bid price")
			}
			for i := 0; i < len(quantities) && i < len(depthLevels); i++ {
				price := basePrice * (1 + depthLevels[i])
				priceLevels = append(priceLevels, struct {
					Price    string
					Quantity string
				}{Price: fmt.Sprintf("%.*f", pricePrecision, price)})
			}
		}
	} else {
		if side == "SELL" {
			asks, ok := levels.([]binance.Ask)
			if !ok || len(asks) == 0 {
				log.Printf("Invalid or empty asks for SELL: %T, len=%d", levels, len(asks))
				return fmt.Errorf("invalid asks")
			}
			priceLevels = append(priceLevels, struct {
				Price    string
				Quantity string
			}{Price: asks[0].Price, Quantity: asks[0].Quantity})
			quantities[0] = math.Max(quantities[0], 0.00000001)
		} else {
			bids, ok := levels.([]binance.Bid)
			if !ok || len(bids) == 0 {
				log.Printf("Invalid or empty bids for BUY: %T, len=%d", levels, len(bids))
				return fmt.Errorf("invalid bids")
			}
			priceLevels = append(priceLevels, struct {
				Price    string
				Quantity string
			}{Price: bids[0].Price, Quantity: bids[0].Quantity})
			quantities[0] = math.Max(quantities[0], 0.00000001)
		}
	}

	if len(priceLevels) == 0 {
		log.Printf("No valid price levels for %s strategy %d", side, strategy.ID)
		return fmt.Errorf("no valid price levels")
	}

	// 放置订单
	for i, level := range priceLevels {
		if i >= len(quantities) {
			log.Printf("Quantity index %d exceeds quantities length for strategy %d", i, strategy.ID)
			break
		}
		price, err := strconv.ParseFloat(level.Price, 64)
		if err != nil {
			log.Printf("Failed to parse price for strategy %d: %v", strategy.ID, err)
			continue
		}
		quantity := strategy.TotalQuantity * quantities[i]
		priceStr := fmt.Sprintf("%.*f", pricePrecision, price)
		quantityStr := fmt.Sprintf("%.*f", quantityPrecision, quantity)
		notional := price * quantity
		if notional < minNotional {
			log.Printf("Order notional %.8f below minimum %.8f for strategy %d, adjusting qty", notional, minNotional, strategy.ID)
			quantity = math.Ceil(minNotional/price*math.Pow(10, float64(quantityPrecision))) / math.Pow(10, float64(quantityPrecision))
			quantityStr = fmt.Sprintf("%.*f", quantityPrecision, quantity)
			notional = price * quantity
		}
		log.Printf("Placing %s order: StrategyID=%d, Price=%s, Quantity=%s, Notional=%.8f", side, strategy.ID, priceStr, quantityStr, notional)
		order, err := client.NewCreateOrderService().
			Symbol(strategy.Symbol).
			Side(binance.SideType(side)).
			Type(binance.OrderTypeLimit).
			TimeInForce(binance.TimeInForceTypeGTC).
			Quantity(quantityStr).
			Price(priceStr).
			Do(context.Background())
		if err != nil {
			log.Printf("Failed to place %s order for strategy %d: %v", side, strategy.ID, err)
			for _, po := range placedOrders {
				_, cancelErr := client.NewCancelOrderService().Symbol(strategy.Symbol).OrderID(po.OrderID).Do(context.Background())
				if cancelErr != nil {
					log.Printf("Failed to cancel order %d for strategy %d: %v", po.OrderID, strategy.ID, cancelErr)
				}
			}
			return fmt.Errorf("failed to place order: %v", err)
		}
		dbOrder := models.Order{
			StrategyID:  strategy.ID,
			UserID:      userID,
			Symbol:      strategy.Symbol,
			Side:        side,
			Price:       price,
			Quantity:    quantity,
			OrderID:     order.OrderID,
			Status:      "pending",
			CancelAfter: time.Now().Add(2 * time.Hour),
		}
		if err := cfg.DB.Create(&dbOrder).Error; err != nil {
			log.Printf("Failed to save order %d for strategy %d: %v", dbOrder.OrderID, strategy.ID, err)
			for _, po := range placedOrders {
				_, cancelErr := client.NewCancelOrderService().Symbol(strategy.Symbol).OrderID(po.OrderID).Do(context.Background())
				if cancelErr != nil {
					log.Printf("Failed to cancel order %d for strategy %d: %v", po.OrderID, strategy.ID, cancelErr)
				}
			}
			_, cancelErr := client.NewCancelOrderService().Symbol(strategy.Symbol).OrderID(order.OrderID).Do(context.Background())
			if cancelErr != nil {
				log.Printf("Failed to cancel order %d for strategy %d: %v", order.OrderID, strategy.ID, cancelErr)
			}
			return fmt.Errorf("failed to save order: %v", err)
		}
		placedOrders = append(placedOrders, dbOrder)
	}

	if len(placedOrders) == 0 {
		log.Printf("No orders placed for %s strategy %d", side, strategy.ID)
		return fmt.Errorf("no orders placed")
	}

	for _, dbOrder := range placedOrders {
		log.Printf("Order placed: StrategyID=%d, OrderID=%d, Side=%s, Price=%.2f, Quantity=%.8f", strategy.ID, dbOrder.OrderID, side, dbOrder.Price, dbOrder.Quantity)
	}
	return nil
}

// parseQuantitiesAndDepthLevels 解析 Quantities 和 DepthLevels
func parseQuantitiesAndDepthLevels(quantitiesStr, depthLevelsStr string, strategyID uint) ([]float64, []float64, error) {
	var quantities, depthLevels []float64
	quantitiesStr = strings.Trim(quantitiesStr, "[]")
	depthLevelsStr = strings.Trim(depthLevelsStr, "[]")
	if quantitiesStr == "" || depthLevelsStr == "" {
		return nil, nil, fmt.Errorf("empty quantities or depth levels for strategy %d", strategyID)
	}

	qStrs := strings.Split(quantitiesStr, ",")
	dStrs := strings.Split(depthLevelsStr, ",")
	for i, q := range qStrs {
		q = strings.TrimSpace(q)
		if q == "" {
			log.Printf("Empty quantity at index %d for strategy %d", i, strategyID)
			continue
		}
		if qty, err := strconv.ParseFloat(q, 64); err == nil && qty > 0 {
			quantities = append(quantities, qty)
		} else {
			log.Printf("Invalid quantity '%s' at index %d for strategy %d: %v", q, i, strategyID, err)
		}
	}
	for i, d := range dStrs {
		d = strings.TrimSpace(d)
		if d == "" {
			log.Printf("Empty depth level at index %d for strategy %d", i, strategyID)
			continue
		}
		if lvl, err := strconv.ParseFloat(d, 64); err == nil && lvl >= 0 {
			depthLevels = append(depthLevels, lvl)
		} else {
			log.Printf("Invalid depth level '%s' at index %d for strategy %d: %v", d, i, strategyID, err)
		}
	}

	if len(quantities) == 0 || len(depthLevels) == 0 || len(quantities) != len(depthLevels) {
		return nil, nil, fmt.Errorf("invalid quantities or depth levels for strategy %d: quantities=%v, depthLevels=%v", strategyID, quantities, depthLevels)
	}
	return quantities, depthLevels, nil
}
