package tasks

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
	"gorm.io/gorm"
)

// 双币投资API响应结构体
type DCIProductListResponse struct {
	Total int                  `json:"total"`
	List  []DCIProductListItem `json:"list"` // 币安实际返回的是list
}

type DCIProductListItem struct {
	Id              string `json:"id"`
	Symbol          string `json:"symbol"`
	Direction       string `json:"direction"` // UP/DOWN
	StrikePrice     string `json:"strikePrice"`
	Duration        int    `json:"duration"`
	Apy             string `json:"apy"`
	MinAmount       string `json:"minAmount"`
	MaxAmount       string `json:"maxAmount"`
	DeliveryDate    int64  `json:"deliveryDate"`
	ProductStatus   string `json:"productStatus"`
	PurchaseEndTime int64  `json:"purchaseEndTime"`
	BaseAsset       string `json:"baseAsset"`
	QuoteAsset      string `json:"quoteAsset"`
	InvestAsset     string `json:"investAsset"`
	InvestCoin      string `json:"investCoin"`    // 币安返回的字段名
	ExercisedCoin   string `json:"exercisedCoin"` // 币安返回的字段名
}

type DCISubscribeResponse struct {
	PositionId   string `json:"positionId"`
	PurchaseTime int64  `json:"purchaseTime"`
}

type DCIPositionResponse struct {
	Total int               `json:"total"`
	Rows  []DCIPositionItem `json:"rows"`
}

type DCIPositionItem struct {
	Id           string `json:"id"`
	PositionId   string `json:"positionId"`
	ProductId    string `json:"productId"`
	Symbol       string `json:"symbol"`
	Direction    string `json:"direction"`
	StrikePrice  string `json:"strikePrice"`
	Duration     int    `json:"duration"`
	Apy          string `json:"apy"`
	InvestAmount string `json:"investAmount"`
	InvestAsset  string `json:"investAsset"`
	PurchaseTime int64  `json:"purchaseTime"`
	DeliveryDate int64  `json:"deliveryDate"`
	Status       string `json:"status"` // PENDING/ACTIVE/SETTLED
	SettleAsset  string `json:"settleAsset"`
	SettleAmount string `json:"settleAmount"`
	ProfitAmount string `json:"profitAmount"`
	ProfitAsset  string `json:"profitAsset"`
}

// BinanceAPIError 币安API错误响应
type BinanceAPIError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

// dciRequest 用于处理双币投资API请求的辅助函数
func dciRequest(apiKey, secretKey, method, endpoint string, params map[string]interface{}) ([]byte, error) {
	baseURL := "https://api.binance.com"

	// 添加时间戳
	params["timestamp"] = fmt.Sprintf("%d", time.Now().UnixMilli())

	// 构建查询字符串
	query := buildQueryString(params)

	// 生成签名
	signature := sign(query, secretKey)
	query += "&signature=" + signature

	// 构建完整URL
	fullURL := baseURL + endpoint
	if method == "GET" {
		fullURL += "?" + query
	}

	// 创建请求
	var req *http.Request
	var err error

	if method == "GET" {
		req, err = http.NewRequest(method, fullURL, nil)
	} else {
		req, err = http.NewRequest(method, fullURL, strings.NewReader(query))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	req.Header.Set("X-MBX-APIKEY", apiKey)

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		var apiErr BinanceAPIError
		if err := json.Unmarshal(body, &apiErr); err == nil {
			return nil, fmt.Errorf("API错误 [%d]: %s", apiErr.Code, apiErr.Msg)
		}
		return nil, fmt.Errorf("HTTP错误 %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// buildQueryString 构建查询字符串
func buildQueryString(params map[string]interface{}) string {
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		v := params[k]
		parts = append(parts, fmt.Sprintf("%s=%v", k, url.QueryEscape(fmt.Sprintf("%v", v))))
	}

	return strings.Join(parts, "&")
}

// sign 生成签名
func sign(message, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// StartDualInvestmentTasks 启动双币投资相关任务
func StartDualInvestmentTasks(cfg *config.Config) {
	// 产品同步任务 - 每5分钟执行一次
	go syncDualInvestmentProducts(cfg)

	// 策略执行任务 - 每分钟检查一次
	go executeDualInvestmentStrategies(cfg)

	// 订单结算监控 - 每10分钟检查一次
	go monitorDualInvestmentSettlement(cfg)
}

// syncDualInvestmentProducts 同步双币投资产品
func syncDualInvestmentProducts(cfg *config.Config) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	// 立即执行一次
	doSyncProducts(cfg)

	for range ticker.C {
		doSyncProducts(cfg)
	}
}

// doSyncProducts 执行产品同步 - 真实API版本
func doSyncProducts(cfg *config.Config) {
	log.Println("开始同步双币投资产品...")

	// 获取一个有效的用户API密钥用于同步
	var user models.User
	if err := cfg.DB.Where("api_key != ? AND secret_key != ?", "", "").First(&user).Error; err != nil {
		log.Printf("没有找到有效的API密钥用于同步产品")
		return
	}

	// 解密API密钥
	apiKey, err := user.GetDecryptedAPIKey()
	if err != nil {
		log.Printf("解密API Key失败: %v", err)
		return
	}
	secretKey, err := user.GetDecryptedSecretKey()
	if err != nil {
		log.Printf("解密Secret Key失败: %v", err)
		return
	}

	client := binance.NewClient(apiKey, secretKey)

	// 定义要同步的交易对列表
	symbols := []string{
		"BTCUSDT", "ETHUSDT", "BNBUSDT", "SOLUSDT", "ADAUSDT",
		"XRPUSDT", "DOTUSDT", "DOGEUSDT", "AVAXUSDT", "MATICUSDT",
	}

	for _, symbol := range symbols {
		// 获取当前价格
		prices, err := client.NewListPricesService().Symbol(symbol).Do(context.Background())
		if err != nil || len(prices) == 0 {
			log.Printf("获取 %s 价格失败: %v", symbol, err)
			continue
		}
		currentPrice, _ := strconv.ParseFloat(prices[0].Price, 64)

		// 调用双币投资产品列表API
		products, err := getDCIProductList(apiKey, secretKey, symbol)
		if err != nil {
			log.Printf("获取 %s 双币投资产品失败: %v", symbol, err)
			continue
		}

		// 保存产品到数据库
		for _, p := range products {
			strikePrice, _ := strconv.ParseFloat(p.StrikePrice, 64)
			apy, _ := strconv.ParseFloat(p.Apy, 64)
			minAmount, _ := strconv.ParseFloat(p.MinAmount, 64)
			maxAmount, _ := strconv.ParseFloat(p.MaxAmount, 64)

			// 计算深度级别（基于执行价格与当前价格的偏离度）
			priceOffset := abs((strikePrice - currentPrice) / currentPrice * 100)
			depthLevel := int(priceOffset/0.5) + 1 // 每0.5%为一个深度级别

			product := models.DualInvestmentProduct{
				Symbol:         p.Symbol,
				Direction:      p.Direction,
				StrikePrice:    strikePrice,
				APY:            apy * 100, // API返回的是小数，转换为百分比
				Duration:       p.Duration,
				MinAmount:      minAmount,
				MaxAmount:      maxAmount,
				SettlementTime: time.Unix(p.DeliveryDate/1000, 0),
				ProductID:      p.Id,
				Status:         mapProductStatus(p.ProductStatus),
				BaseAsset:      p.BaseAsset,
				QuoteAsset:     p.QuoteAsset,
				CurrentPrice:   currentPrice,
				DepthLevel:     depthLevel,
			}

			// 使用 ProductID 作为唯一标识更新或创建
			if err := cfg.DB.Where("product_id = ?", product.ProductID).
				Assign(product).
				FirstOrCreate(&product).Error; err != nil {
				log.Printf("保存产品失败: %v", err)
			}
		}

		log.Printf("同步 %s 产品完成，共 %d 个产品", symbol, len(products))
	}

	// 清理过期产品
	if err := cfg.DB.Model(&models.DualInvestmentProduct{}).
		Where("settlement_time < ? AND status = ?", time.Now(), "active").
		Update("status", "expired").Error; err != nil {
		log.Printf("更新过期产品失败: %v", err)
	}

	log.Println("双币投资产品同步完成")
}

// getDCIProductList 获取双币投资产品列表
func getDCIProductList(apiKey, secretKey, symbol string) ([]DCIProductListItem, error) {
	// 从交易对中解析基础资产和计价资产
	// 例如 BTCUSDT -> BTC 和 USDT
	var baseAsset, quoteAsset string
	if strings.HasSuffix(symbol, "USDT") {
		baseAsset = strings.TrimSuffix(symbol, "USDT")
		quoteAsset = "USDT"
	} else if strings.HasSuffix(symbol, "BUSD") {
		baseAsset = strings.TrimSuffix(symbol, "BUSD")
		quoteAsset = "BUSD"
	} else {
		// 其他情况，简单处理
		if len(symbol) >= 6 {
			baseAsset = symbol[:3]
			quoteAsset = symbol[3:]
		} else {
			return nil, fmt.Errorf("无法解析交易对: %s", symbol)
		}
	}

	var allProducts []DCIProductListItem

	// 1. 查询看涨期权产品（CALL）
	params := map[string]interface{}{
		"optionType":    "CALL",
		"investCoin":    baseAsset,
		"exercisedCoin": quoteAsset,
		"pageSize":      100,
		"pageIndex":     1,
	}

	res, err := dciRequest(apiKey, secretKey, "GET", "/sapi/v1/dci/product/list", params)
	if err != nil {
		log.Printf("获取 %s CALL产品失败 (投资币种=%s): %v", symbol, baseAsset, err)
	} else {
		var response DCIProductListResponse
		if err := json.Unmarshal(res, &response); err == nil {
			log.Printf("获取到 %d 个CALL产品 (投资币种=%s)", len(response.List), baseAsset)

			// 打印前3个产品的详细信息进行调试
			for i, product := range response.List {
				if i < 3 {
					log.Printf("CALL产品[%d]: ID=%s, Symbol='%s', InvestCoin=%s, ExercisedCoin=%s, StrikePrice=%s",
						i, product.Id, product.Symbol, product.InvestCoin, product.ExercisedCoin, product.StrikePrice)
				}

				// 设置产品属性
				response.List[i].Direction = "UP"

				// 如果Symbol为空，根据投资币种和行权币种构建
				if response.List[i].Symbol == "" {
					response.List[i].Symbol = baseAsset + quoteAsset
					log.Printf("为产品 %s 设置Symbol: %s", product.Id, response.List[i].Symbol)
				}

				// 设置BaseAsset和QuoteAsset
				response.List[i].BaseAsset = baseAsset
				response.List[i].QuoteAsset = quoteAsset
				response.List[i].InvestAsset = product.InvestCoin
			}
			allProducts = append(allProducts, response.List...)
		} else {
			log.Printf("解析CALL响应失败: %v", err)
		}
	}

	// 2. 查询看跌期权产品（PUT）
	params = map[string]interface{}{
		"optionType":    "PUT",
		"investCoin":    quoteAsset,
		"exercisedCoin": baseAsset,
		"pageSize":      100,
		"pageIndex":     1,
	}

	res, err = dciRequest(apiKey, secretKey, "GET", "/sapi/v1/dci/product/list", params)
	if err != nil {
		log.Printf("获取 %s PUT产品失败 (投资币种=%s): %v", symbol, quoteAsset, err)
	} else {
		var response DCIProductListResponse
		if err := json.Unmarshal(res, &response); err == nil {
			log.Printf("获取到 %d 个PUT产品 (投资币种=%s)", len(response.List), quoteAsset)

			// 打印前3个产品的详细信息进行调试
			for i, product := range response.List {
				if i < 3 {
					log.Printf("PUT产品[%d]: ID=%s, Symbol='%s', InvestCoin=%s, ExercisedCoin=%s, StrikePrice=%s",
						i, product.Id, product.Symbol, product.InvestCoin, product.ExercisedCoin, product.StrikePrice)
				}

				// 设置产品属性
				response.List[i].Direction = "DOWN"

				// 如果Symbol为空，根据投资币种和行权币种构建
				if response.List[i].Symbol == "" {
					response.List[i].Symbol = baseAsset + quoteAsset
					log.Printf("为产品 %s 设置Symbol: %s", product.Id, response.List[i].Symbol)
				}

				// 设置BaseAsset和QuoteAsset
				response.List[i].BaseAsset = baseAsset
				response.List[i].QuoteAsset = quoteAsset
				response.List[i].InvestAsset = product.InvestCoin
			}
			allProducts = append(allProducts, response.List...)
		} else {
			log.Printf("解析PUT响应失败: %v", err)
		}
	}

	// 统计所有产品的Symbol
	symbolCount := make(map[string]int)
	for _, product := range allProducts {
		symbolCount[product.Symbol]++
	}

	log.Printf("所有产品的Symbol分布:")
	for sym, count := range symbolCount {
		if count > 0 {
			log.Printf("  '%s': %d 个产品", sym, count)
		}
	}

	// 过滤只保留指定交易对的产品
	var filteredProducts []DCIProductListItem
	for _, product := range allProducts {
		if product.Symbol == symbol {
			filteredProducts = append(filteredProducts, product)
		}
	}

	log.Printf("交易对 %s 共获取到 %d 个有效产品 (期望Symbol='%s')", symbol, len(filteredProducts), symbol)
	return filteredProducts, nil
}

// mapProductStatus 映射产品状态
func mapProductStatus(status string) string {
	switch status {
	case "PURCHASABLE", "PENDING":
		return "active"
	case "SOLD_OUT":
		return "sold_out"
	case "EXPIRED":
		return "expired"
	default:
		return "unknown"
	}
}

// executeDualInvestmentStrategies 执行双币投资策略
func executeDualInvestmentStrategies(cfg *config.Config) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// 查询所有启用的策略，需要检查的策略
		var strategies []models.DualInvestmentStrategy
		now := time.Now()
		// 价格触发策略即使状态是completed也要继续监控，除非被禁用
		if err := cfg.DB.Where("enabled = ? AND (status = ? OR (status = ? AND strategy_type != ?)) AND (next_check_time IS NULL OR next_check_time <= ?)",
			true, "active", "completed", "price_trigger", now).
			Find(&strategies).Error; err != nil {
			log.Printf("查询双币投资策略失败: %v", err)
			continue
		}

		log.Printf("找到 %d 个需要检查的双币投资策略", len(strategies))

		for _, strategy := range strategies {
			go executeStrategy(cfg, strategy)
		}
	}
}

// executeStrategy 执行单个策略
func executeStrategy(cfg *config.Config, strategy models.DualInvestmentStrategy) {
	// 获取用户信息
	var user models.User
	if err := cfg.DB.First(&user, strategy.UserID).Error; err != nil {
		log.Printf("策略用户未找到: %v", err)
		return
	}

	// 解密API密钥
	apiKey, err := user.GetDecryptedAPIKey()
	if err != nil {
		log.Printf("解密用户 %d API Key失败: %v", user.ID, err)
		return
	}

	secretKey, err := user.GetDecryptedSecretKey()
	if err != nil {
		log.Printf("解密用户 %d Secret Key失败: %v", user.ID, err)
		return
	}

	if apiKey == "" || secretKey == "" {
		log.Printf("用户 %d 未设置 API 密钥，跳过策略 %d", user.ID, strategy.ID)
		return
	}

	// 检查投资限额
	if strategy.CurrentInvested >= strategy.TotalInvestmentLimit {
		log.Printf("策略 %d 已达到投资限额", strategy.ID)
		return
	}

	symbol := strategy.BaseAsset + strategy.QuoteAsset

	switch strategy.StrategyType {
	case "single":
		executeSingleStrategy(cfg, strategy, user, symbol)
	case "auto_reinvest":
		executeAutoReinvestStrategy(cfg, strategy, user, symbol)
	case "ladder":
		executeLadderStrategy(cfg, strategy, user, symbol)
	case "price_trigger":
		executePriceTriggerStrategy(cfg, strategy, user, symbol)
	}
}

// createDualInvestmentOrder 创建双币投资订单 - 真实API版本
func createDualInvestmentOrder(cfg *config.Config, user models.User, strategy models.DualInvestmentStrategy,
	product *models.DualInvestmentProduct, investAmount float64) bool {

	// 解密API密钥
	apiKey, err := user.GetDecryptedAPIKey()
	if err != nil {
		log.Printf("创建订单时解密API Key失败: %v", err)
		return false
	}
	secretKey, err := user.GetDecryptedSecretKey()
	if err != nil {
		log.Printf("创建订单时解密Secret Key失败: %v", err)
		return false
	}

	// 调用双币投资下单API
	params := map[string]interface{}{
		"productId":    product.ProductID,
		"investAmount": fmt.Sprintf("%.8f", investAmount),
		"autoCompound": "false", // 是否自动复投
	}

	res, err := dciRequest(apiKey, secretKey, "POST", "/sapi/v1/dci/product/subscribe", params)
	if err != nil {
		log.Printf("调用双币投资下单API失败: %v", err)
		return false
	}

	var subscribeResp DCISubscribeResponse
	if err := json.Unmarshal(res, &subscribeResp); err != nil {
		log.Printf("解析下单响应失败: %v", err)
		return false
	}

	// 创建订单记录
	order := models.DualInvestmentOrder{
		UserID:         user.ID,
		StrategyID:     &strategy.ID,
		ProductID:      product.ID,
		OrderID:        subscribeResp.PositionId, // 使用币安返回的positionId
		Symbol:         product.Symbol,
		InvestAsset:    product.BaseAsset, // 根据产品确定投资资产
		InvestAmount:   investAmount,
		StrikePrice:    product.StrikePrice,
		APY:            product.APY,
		Direction:      product.Direction,
		Duration:       product.Duration,
		SettlementTime: product.SettlementTime,
		Status:         "active",
	}

	// 使用事务
	err = cfg.DB.Transaction(func(tx *gorm.DB) error {
		// 创建订单
		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		// 更新策略已投资金额
		if err := tx.Model(&strategy).Updates(map[string]interface{}{
			"current_invested": gorm.Expr("current_invested + ?", investAmount),
			"last_executed_at": time.Now(),
		}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		log.Printf("保存双币投资订单失败: %v", err)
		return false
	}

	log.Printf("创建双币投资订单成功: 策略=%d, 产品=%s, 金额=%.2f, 币安订单ID=%s",
		strategy.ID, product.Symbol, investAmount, subscribeResp.PositionId)
	return true
}

// monitorDualInvestmentSettlement 监控双币投资结算
func monitorDualInvestmentSettlement(cfg *config.Config) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// 查询所有活跃的订单
		var orders []models.DualInvestmentOrder
		if err := cfg.DB.Where("status = ?", "active").Find(&orders).Error; err != nil {
			log.Printf("查询活跃订单失败: %v", err)
			continue
		}

		// 按用户分组
		userOrders := make(map[uint][]models.DualInvestmentOrder)
		for _, order := range orders {
			userOrders[order.UserID] = append(userOrders[order.UserID], order)
		}

		// 处理每个用户的订单
		for userID, orderList := range userOrders {
			go checkUserOrders(cfg, userID, orderList)
		}
	}
}

// checkUserOrders 检查用户的订单状态
func checkUserOrders(cfg *config.Config, userID uint, orders []models.DualInvestmentOrder) {
	// 获取用户信息
	var user models.User
	if err := cfg.DB.First(&user, userID).Error; err != nil {
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

	// 获取用户的所有持仓
	positions, err := getDCIPositions(apiKey, secretKey)
	if err != nil {
		log.Printf("获取用户 %d 的双币投资持仓失败: %v", userID, err)
		return
	}

	// 创建持仓映射
	positionMap := make(map[string]DCIPositionItem)
	for _, pos := range positions {
		positionMap[pos.PositionId] = pos
	}

	// 更新订单状态
	for _, order := range orders {
		if pos, exists := positionMap[order.OrderID]; exists {
			updateOrderFromPosition(cfg, &order, pos)
		} else {
			// 如果在持仓中找不到，可能已经结算或取消
			// 检查是否超过结算时间
			if time.Now().After(order.SettlementTime) {
				order.Status = "settled"
				cfg.DB.Save(&order)
			}
		}
	}
}

// getDCIPositions 获取双币投资持仓
func getDCIPositions(apiKey, secretKey string) ([]DCIPositionItem, error) {
	params := map[string]interface{}{
		"pageSize":  100,
		"pageIndex": 1,
	}

	res, err := dciRequest(apiKey, secretKey, "GET", "/sapi/v1/dci/product/positions", params)
	if err != nil {
		return nil, err
	}

	var response DCIPositionResponse
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, err
	}

	return response.Rows, nil
}

// updateOrderFromPosition 根据持仓信息更新订单
func updateOrderFromPosition(cfg *config.Config, order *models.DualInvestmentOrder, position DCIPositionItem) {
	// 更新订单状态
	switch position.Status {
	case "PENDING":
		order.Status = "pending"
	case "ACTIVE":
		order.Status = "active"
	case "SETTLED":
		order.Status = "settled"

		// 更新结算信息
		settlementAmount, _ := strconv.ParseFloat(position.SettleAmount, 64)
		profitAmount, _ := strconv.ParseFloat(position.ProfitAmount, 64)

		order.SettlementAsset = position.SettleAsset
		order.SettlementAmount = settlementAmount
		order.SettledAt = &time.Time{}
		*order.SettledAt = time.Now()

		// 计算盈亏
		if position.SettleAsset == position.InvestAsset {
			// 同币种结算，直接计算差额
			order.PnL = profitAmount
			order.PnLPercent = (profitAmount / order.InvestAmount) * 100
		} else {
			// 不同币种结算，需要根据当前价格计算
			// 这里简化处理，实际应该获取结算时的价格
			order.PnL = profitAmount
			order.PnLPercent = 0 // 需要更复杂的计算
		}

		// 计算实际年化收益率
		days := float64(order.Duration)
		order.ActualAPY = (order.PnLPercent / days) * 365

		// 如果有关联策略，更新策略的已投资金额
		if order.StrategyID != nil {
			cfg.DB.Model(&models.DualInvestmentStrategy{}).
				Where("id = ?", *order.StrategyID).
				Update("current_invested", gorm.Expr("current_invested - ?", order.InvestAmount))
		}
	}

	// 保存更新
	if err := cfg.DB.Save(order).Error; err != nil {
		log.Printf("更新订单 %s 状态失败: %v", order.OrderID, err)
	}
}

// 保留其他函数不变...
// executeSingleStrategy, executeAutoReinvestStrategy, executeLadderStrategy, executePriceTriggerStrategy
// findBestProduct, calculateInvestAmount, abs 等函数保持原样

// executeSingleStrategy 执行单次投资策略
func executeSingleStrategy(cfg *config.Config, strategy models.DualInvestmentStrategy, user models.User, symbol string) {
	// 查找符合条件的产品
	product := findBestProduct(cfg, strategy, symbol)
	if product == nil {
		return
	}

	// 计算投资金额
	investAmount := calculateInvestAmount(strategy, product)
	if investAmount <= 0 {
		return
	}

	// 创建订单
	createDualInvestmentOrder(cfg, user, strategy, product, investAmount)
}

// executeAutoReinvestStrategy 执行自动复投策略
func executeAutoReinvestStrategy(cfg *config.Config, strategy models.DualInvestmentStrategy, user models.User, symbol string) {
	// 检查是否有已结算的订单需要复投
	var settledOrders []models.DualInvestmentOrder
	if err := cfg.DB.Where("strategy_id = ? AND status = ? AND created_at > ?",
		strategy.ID, "settled", time.Now().Add(-24*time.Hour)).
		Find(&settledOrders).Error; err != nil {
		return
	}

	// 对每个已结算订单进行复投
	for _, order := range settledOrders {
		// 查找类似的产品
		product := findBestProduct(cfg, strategy, symbol)
		if product == nil {
			continue
		}

		// 使用结算金额进行复投
		investAmount := order.SettlementAmount
		if investAmount > strategy.MaxSingleAmount {
			investAmount = strategy.MaxSingleAmount
		}

		createDualInvestmentOrder(cfg, user, strategy, product, investAmount)
	}
}

// executeLadderStrategy 执行梯度投资策略 - 完全重写版本
func executeLadderStrategy(cfg *config.Config, strategy models.DualInvestmentStrategy, user models.User, symbol string) {
	// 解密API密钥
	apiKey, err := user.GetDecryptedAPIKey()
	if err != nil {
		log.Printf("解密API Key失败: %v", err)
		return
	}
	secretKey, err := user.GetDecryptedSecretKey()
	if err != nil {
		log.Printf("解密Secret Key失败: %v", err)
		return
	}

	// 获取当前价格
	client := binance.NewClient(apiKey, secretKey)
	prices, err := client.NewListPricesService().Symbol(symbol).Do(context.Background())
	if err != nil || len(prices) == 0 {
		log.Printf("获取 %s 价格失败: %v", symbol, err)
		return
	}

	currentPrice, _ := strconv.ParseFloat(prices[0].Price, 64)
	log.Printf("梯度策略 %d: %s 当前价格 %.2f, 基准价格 %.2f",
		strategy.ID, symbol, currentPrice, strategy.BasePrice)

	// 解析梯度配置
	var ladderConfig []models.LadderConfigItem
	if err := json.Unmarshal([]byte(strategy.LadderConfig), &ladderConfig); err != nil {
		log.Printf("解析梯度配置失败: %v", err)
		return
	}

	if len(ladderConfig) == 0 {
		log.Printf("梯度策略 %d 没有配置梯度参数", strategy.ID)
		return
	}

	// 根据方向偏好查询产品
	query := cfg.DB.Where("symbol = ? AND status = ?", symbol, "active")

	// 基准价格过滤
	if strategy.DirectionPreference == "UP" {
		// 看涨：只选择执行价格 <= 基准价格的产品
		query = query.Where("direction = ? AND strike_price <= ?", "UP", strategy.BasePrice)
	} else if strategy.DirectionPreference == "DOWN" {
		// 看跌：只选择执行价格 >= 基准价格的产品
		query = query.Where("direction = ? AND strike_price >= ?", "DOWN", strategy.BasePrice)
	} else {
		// 双向：根据方向选择
		query = query.Where(
			"(direction = ? AND strike_price <= ?) OR (direction = ? AND strike_price >= ?)",
			"UP", strategy.BasePrice, "DOWN", strategy.BasePrice)
	}

	// APY筛选
	if strategy.TargetAPYMin > 0 {
		query = query.Where("apy >= ?", strategy.TargetAPYMin)
	}
	if strategy.TargetAPYMax > 0 {
		query = query.Where("apy <= ?", strategy.TargetAPYMax)
	}

	// 期限筛选
	if strategy.MinDuration > 0 {
		query = query.Where("duration >= ?", strategy.MinDuration)
	}
	if strategy.MaxDuration > 0 {
		query = query.Where("duration <= ?", strategy.MaxDuration)
	}

	var products []models.DualInvestmentProduct
	if err := query.Find(&products).Error; err != nil {
		log.Printf("查询梯度产品失败: %v", err)
		return
	}

	if len(products) == 0 {
		log.Printf("没有找到符合条件的产品")
		return
	}

	// 按深度级别排序产品
	sort.Slice(products, func(i, j int) bool {
		return products[i].DepthLevel < products[j].DepthLevel
	})

	// 根据梯度配置进行投资
	totalInvested := 0.0
	successCount := 0

	for _, config := range ladderConfig {
		// 查找符合深度要求的产品
		var targetProduct *models.DualInvestmentProduct
		for i := range products {
			if products[i].DepthLevel >= config.MinDepth {
				targetProduct = &products[i]
				break
			}
		}

		if targetProduct == nil {
			log.Printf("没有找到深度 >= %d 的产品", config.MinDepth)
			continue
		}

		// 计算投资金额
		investAmount := strategy.MaxSingleAmount * (config.Percentage / 100.0)

		// 确保不超过剩余限额
		remainingLimit := strategy.TotalInvestmentLimit - strategy.CurrentInvested - totalInvested
		if investAmount > remainingLimit {
			investAmount = remainingLimit
		}

		// 确保满足产品最小投资额
		if investAmount < targetProduct.MinAmount {
			log.Printf("投资金额 %.2f 小于产品最小额 %.2f，跳过", investAmount, targetProduct.MinAmount)
			continue
		}

		// 确保不超过产品最大投资额
		if investAmount > targetProduct.MaxAmount {
			investAmount = targetProduct.MaxAmount
		}

		// 创建订单
		if createDualInvestmentOrder(cfg, user, strategy, targetProduct, investAmount) {
			totalInvested += investAmount
			successCount++
			log.Printf("梯度投资成功: 深度=%d, 执行价=%.2f, 年化=%.2f%%, 金额=%.2f",
				targetProduct.DepthLevel, targetProduct.StrikePrice, targetProduct.APY, investAmount)
		}
	}

	// 更新下次检查时间（10分钟后）
	nextCheckTime := time.Now().Add(10 * time.Minute)
	cfg.DB.Model(&strategy).Update("next_check_time", nextCheckTime)

	log.Printf("梯度策略 %d 执行完成: 成功投资 %d 笔，总金额 %.2f，下次检查时间 %s",
		strategy.ID, successCount, totalInvested, nextCheckTime.Format("15:04:05"))
}

// executePriceTriggerStrategy 执行价格触发策略
func executePriceTriggerStrategy(cfg *config.Config, strategy models.DualInvestmentStrategy, user models.User, symbol string) {
	// 解密API密钥
	apiKey, err := user.GetDecryptedAPIKey()
	if err != nil {
		log.Printf("解密API Key失败: %v", err)
		return
	}
	secretKey, err := user.GetDecryptedSecretKey()
	if err != nil {
		log.Printf("解密Secret Key失败: %v", err)
		return
	}

	// 获取当前价格
	client := binance.NewClient(apiKey, secretKey)
	prices, err := client.NewListPricesService().Symbol(symbol).Do(context.Background())
	if err != nil || len(prices) == 0 {
		log.Printf("获取 %s 价格失败: %v", symbol, err)
		return
	}

	currentPrice, _ := strconv.ParseFloat(prices[0].Price, 64)
	log.Printf("价格触发策略 %d: %s 当前价格 %.2f, 触发价格 %.2f, 触发类型 %s",
		strategy.ID, symbol, currentPrice, strategy.TriggerPrice, strategy.TriggerType)

	// 检查是否触发
	triggered := false
	if strategy.TriggerType == "above" && currentPrice >= strategy.TriggerPrice {
		triggered = true
		log.Printf("价格触发策略 %d 触发: 当前价格 %.2f >= 触发价格 %.2f",
			strategy.ID, currentPrice, strategy.TriggerPrice)
	} else if strategy.TriggerType == "below" && currentPrice <= strategy.TriggerPrice {
		triggered = true
		log.Printf("价格触发策略 %d 触发: 当前价格 %.2f <= 触发价格 %.2f",
			strategy.ID, currentPrice, strategy.TriggerPrice)
	}

	if !triggered {
		// 更新下次检查时间（1分钟后）
		nextCheckTime := time.Now().Add(1 * time.Minute)
		cfg.DB.Model(&strategy).Update("next_check_time", nextCheckTime)
		return
	}

	// 触发后，先查询所有该交易对的产品进行调试
	log.Printf("=== 开始调试：查询 %s 的所有产品 ===", symbol)
	var allProducts []models.DualInvestmentProduct
	if err := cfg.DB.Where("symbol = ?", symbol).Find(&allProducts).Error; err != nil {
		log.Printf("查询所有产品失败: %v", err)
	} else {
		log.Printf("找到 %s 的产品总数: %d", symbol, len(allProducts))
		for i, p := range allProducts {
			log.Printf("产品[%d]: ID=%d, 方向=%s, 执行价=%.2f, 年化=%.2f%%, 期限=%d天, 状态=%s, 最小额=%.2f, 最大额=%.2f",
				i+1, p.ID, p.Direction, p.StrikePrice, p.APY, p.Duration, p.Status, p.MinAmount, p.MaxAmount)
		}
	}

	// 打印策略的筛选条件
	log.Printf("=== 策略 %d 的筛选条件 ===", strategy.ID)
	log.Printf("方向偏好: %s", strategy.DirectionPreference)
	log.Printf("目标年化: %.2f%% - %.2f%%", strategy.TargetAPYMin, strategy.TargetAPYMax)
	log.Printf("期限范围: %d - %d 天", strategy.MinDuration, strategy.MaxDuration)
	log.Printf("最大执行价格偏离度: %.2f%%", strategy.MaxStrikePriceOffset)
	log.Printf("单笔最大金额: %.2f", strategy.MaxSingleAmount)
	log.Printf("=== 筛选条件结束 ===")

	// 查找最佳产品
	product := findBestProduct(cfg, strategy, symbol)
	if product == nil {
		log.Printf("价格触发策略 %d 触发但没有找到合适的产品", strategy.ID)
		// 继续监控，5分钟后再次检查
		nextCheckTime := time.Now().Add(5 * time.Minute)
		cfg.DB.Model(&strategy).Update("next_check_time", nextCheckTime)
		return
	}

	log.Printf("找到合适的产品: ID=%d, 方向=%s, 执行价=%.2f, 年化=%.2f%%",
		product.ID, product.Direction, product.StrikePrice, product.APY)

	investAmount := calculateInvestAmount(strategy, product)
	if investAmount <= 0 {
		log.Printf("价格触发策略 %d 触发但投资金额计算为0", strategy.ID)
		return
	}

	// 创建订单
	if createDualInvestmentOrder(cfg, user, strategy, product, investAmount) {
		// 更新策略状态为已完成
		cfg.DB.Model(&strategy).Updates(map[string]interface{}{
			"status":           "completed",
			"last_executed_at": time.Now(),
		})
		log.Printf("价格触发策略 %d 执行成功，策略已完成", strategy.ID)
	} else {
		// 失败后5分钟再试
		nextCheckTime := time.Now().Add(5 * time.Minute)
		cfg.DB.Model(&strategy).Update("next_check_time", nextCheckTime)
	}
}

// findBestProduct 查找最佳产品 - 修复SQL注入
func findBestProduct(cfg *config.Config, strategy models.DualInvestmentStrategy, symbol string) *models.DualInvestmentProduct {
	log.Printf("=== findBestProduct 开始查找产品 ===")

	query := cfg.DB.Model(&models.DualInvestmentProduct{}).
		Where("symbol = ? AND status = ?", symbol, "active")

	// 先统计初始产品数量
	var initialCount int64
	query.Count(&initialCount)
	log.Printf("步骤1: 交易对=%s, 状态=active 的产品数量: %d", symbol, initialCount)

	// 方向筛选
	if strategy.DirectionPreference != "BOTH" {
		query = query.Where("direction = ?", strategy.DirectionPreference)
		var directionCount int64
		query.Count(&directionCount)
		log.Printf("步骤2: 方向筛选(%s)后的产品数量: %d", strategy.DirectionPreference, directionCount)
	}

	// APY筛选
	if strategy.TargetAPYMin > 0 {
		query = query.Where("apy >= ?", strategy.TargetAPYMin)
		var apyMinCount int64
		query.Count(&apyMinCount)
		log.Printf("步骤3: 最小年化(>= %.2f%%)筛选后的产品数量: %d", strategy.TargetAPYMin, apyMinCount)
	}
	if strategy.TargetAPYMax > 0 {
		query = query.Where("apy <= ?", strategy.TargetAPYMax)
		var apyMaxCount int64
		query.Count(&apyMaxCount)
		log.Printf("步骤4: 最大年化(<= %.2f%%)筛选后的产品数量: %d", strategy.TargetAPYMax, apyMaxCount)
	}

	// 期限筛选
	if strategy.MinDuration > 0 {
		query = query.Where("duration >= ?", strategy.MinDuration)
		var minDurCount int64
		query.Count(&minDurCount)
		log.Printf("步骤5: 最小期限(>= %d天)筛选后的产品数量: %d", strategy.MinDuration, minDurCount)
	}
	if strategy.MaxDuration > 0 {
		query = query.Where("duration <= ?", strategy.MaxDuration)
		var maxDurCount int64
		query.Count(&maxDurCount)
		log.Printf("步骤6: 最大期限(<= %d天)筛选后的产品数量: %d", strategy.MaxDuration, maxDurCount)
	}

	// 执行价格偏离度筛选 - 修复SQL注入
	if strategy.MaxStrikePriceOffset > 0 {
		// 先获取产品，然后在应用层过滤
		var products []models.DualInvestmentProduct
		if err := query.Find(&products).Error; err != nil {
			log.Printf("查询产品失败: %v", err)
			return nil
		}

		log.Printf("步骤7: 获取到 %d 个产品，开始进行价格偏离度筛选", len(products))

		// 在应用层进行价格偏离度过滤
		var filteredProducts []models.DualInvestmentProduct
		for _, product := range products {
			if product.CurrentPrice > 0 {
				offset := abs((product.StrikePrice - product.CurrentPrice) / product.CurrentPrice * 100)
				log.Printf("  产品ID=%d: 当前价=%.2f, 执行价=%.2f, 偏离度=%.2f%% (最大允许=%.2f%%)",
					product.ID, product.CurrentPrice, product.StrikePrice, offset, strategy.MaxStrikePriceOffset)
				if offset <= strategy.MaxStrikePriceOffset {
					filteredProducts = append(filteredProducts, product)
				}
			} else {
				log.Printf("  产品ID=%d: 当前价格为0，跳过", product.ID)
			}
		}

		log.Printf("步骤8: 价格偏离度筛选后的产品数量: %d", len(filteredProducts))

		// 按APY排序选择最佳
		if len(filteredProducts) > 0 {
			sort.Slice(filteredProducts, func(i, j int) bool {
				return filteredProducts[i].APY > filteredProducts[j].APY
			})
			selectedProduct := &filteredProducts[0]
			log.Printf("=== 选中产品: ID=%d, 年化=%.2f%%, 执行价=%.2f ===",
				selectedProduct.ID, selectedProduct.APY, selectedProduct.StrikePrice)
			return selectedProduct
		}
		log.Printf("=== 没有找到符合条件的产品 ===")
		return nil
	}

	// 没有价格偏离度限制的情况
	var product models.DualInvestmentProduct
	if err := query.Order("apy desc").First(&product).Error; err != nil {
		log.Printf("查询最佳产品失败: %v", err)
		return nil
	}

	log.Printf("=== 选中产品: ID=%d, 年化=%.2f%%, 执行价=%.2f ===",
		product.ID, product.APY, product.StrikePrice)
	return &product
}

// abs 计算绝对值
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// calculateInvestAmount 计算投资金额
func calculateInvestAmount(strategy models.DualInvestmentStrategy, product *models.DualInvestmentProduct) float64 {
	// 可用额度
	available := strategy.TotalInvestmentLimit - strategy.CurrentInvested
	if available <= 0 {
		return 0
	}

	// 单笔限额
	amount := strategy.MaxSingleAmount
	if amount > available {
		amount = available
	}

	// 产品限额
	if amount > product.MaxAmount {
		amount = product.MaxAmount
	}
	if amount < product.MinAmount {
		return 0 // 低于最小限额
	}

	return amount
}
