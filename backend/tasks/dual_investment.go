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
	OrderId         int64  `json:"orderId"` // 添加orderId字段
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

	// 添加更多可能的字段名
	AnnualizedYield string `json:"annualizedYield"` // 可能的年化收益率字段
	Yield           string `json:"yield"`           // 可能的收益率字段
	SettleDate      int64  `json:"settleDate"`      // 可能的结算日期字段
	ExpiryDate      int64  `json:"expiryDate"`      // 可能的到期日期字段
	Status          string `json:"status"`          // 可能的状态字段

	// 币安实际返回的字段名
	APR                  string   `json:"apr"`                  // 年化收益率（币安使用apr而不是apy）
	CanPurchase          bool     `json:"canPurchase"`          // 是否可购买
	CreateTimestamp      int64    `json:"createTimestamp"`      // 创建时间戳
	IsAutoCompoundEnable bool     `json:"isAutoCompoundEnable"` // 是否支持自动复投
	OptionType           string   `json:"optionType"`           // 期权类型 PUT/CALL
	PurchaseDecimal      int      `json:"purchaseDecimal"`      // 购买精度
	AutoCompoundPlanList []string `json:"autoCompoundPlanList"` // 自动复投计划列表
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

	// 调试日志：打印完整的请求信息
	log.Printf("双币投资API请求详情:")
	log.Printf("  - Endpoint: %s %s", method, endpoint)
	log.Printf("  - Parameters: %+v", params)
	log.Printf("  - Query String: %s", query)

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

	// 调试日志：打印请求URL（POST请求）
	if method == "POST" {
		log.Printf("  - Full URL: %s", fullURL)
		log.Printf("  - Request Body: %s", query)
	}

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

	// 调试日志：打印响应
	log.Printf("  - Response Status: %d", resp.StatusCode)
	//log.Printf("  - Response Body: %s", string(body))

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

	// 重要修改：从数据库获取所有用户已添加的交易对
	// 从 Symbol 表和 CustomSymbol 表中获取
	symbolMap := make(map[string]bool)

	// 获取 Symbol 表中的交易对
	var symbols []models.Symbol
	if err := cfg.DB.Where("deleted_at IS NULL").Find(&symbols).Error; err != nil {
		log.Printf("获取交易对列表失败: %v", err)
		return
	}

	for _, sym := range symbols {
		symbolMap[sym.Symbol] = true
	}

	// 获取 CustomSymbol 表中的交易对
	var customSymbols []models.CustomSymbol
	if err := cfg.DB.Where("deleted_at IS NULL").Find(&customSymbols).Error; err != nil {
		log.Printf("获取自定义交易对列表失败: %v", err)
	} else {
		for _, sym := range customSymbols {
			symbolMap[sym.Symbol] = true
		}
	}

	// 转换为数组
	var uniqueSymbols []string
	for symbol := range symbolMap {
		uniqueSymbols = append(uniqueSymbols, symbol)
	}

	if len(uniqueSymbols) == 0 {
		log.Printf("没有找到已添加的交易对")
		return
	}

	log.Printf("准备同步 %d 个交易对的双币投资产品: %v", len(uniqueSymbols), uniqueSymbols)

	totalSynced := 0
	for _, symbol := range uniqueSymbols {
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
		savedCount := 0
		for _, p := range products {
			strikePrice, _ := strconv.ParseFloat(p.StrikePrice, 64)

			// 尝试多个可能的APY字段，优先使用APR
			apy := 0.0
			if p.APR != "" {
				apy, _ = strconv.ParseFloat(p.APR, 64)
			} else if p.Apy != "" {
				apy, _ = strconv.ParseFloat(p.Apy, 64)
			} else if p.AnnualizedYield != "" {
				apy, _ = strconv.ParseFloat(p.AnnualizedYield, 64)
			} else if p.Yield != "" {
				apy, _ = strconv.ParseFloat(p.Yield, 64)
			}

			minAmount, _ := strconv.ParseFloat(p.MinAmount, 64)
			maxAmount, _ := strconv.ParseFloat(p.MaxAmount, 64)

			// 尝试多个可能的日期字段
			var settlementTime time.Time
			if p.SettleDate > 0 {
				settlementTime = time.Unix(p.SettleDate/1000, 0)
			} else if p.DeliveryDate > 0 {
				settlementTime = time.Unix(p.DeliveryDate/1000, 0)
			} else if p.ExpiryDate > 0 {
				settlementTime = time.Unix(p.ExpiryDate/1000, 0)
			} else if p.PurchaseEndTime > 0 {
				settlementTime = time.Unix(p.PurchaseEndTime/1000, 0).AddDate(0, 0, p.Duration)
			} else {
				settlementTime = time.Now().AddDate(0, 0, p.Duration)
			}

			// 处理状态
			status := "active"
			if p.CanPurchase {
				status = "active"
			} else {
				status = "sold_out"
			}
			if p.ProductStatus != "" {
				status = mapProductStatus(p.ProductStatus)
			}
			if status == "unknown" && p.Status != "" {
				status = mapProductStatus(p.Status)
			}

			// 计算深度级别
			priceOffset := abs((strikePrice - currentPrice) / currentPrice * 100)
			depthLevel := int(priceOffset/0.5) + 1

			// 构造保存用的ProductID，可能需要包含orderId信息
			productIdToSave := p.Id
			if p.OrderId > 0 {
				// 如果有orderId，可以考虑将其包含在ProductID中，用分隔符
				productIdToSave = fmt.Sprintf("%s|%d", p.Id, p.OrderId)
			}

			// 调试日志
			if symbol == "SOLUSDT" && savedCount < 3 {
				log.Printf("SOLUSDT产品示例 - ID: %s, OrderID: %d, APY: %.2f%%, StrikePrice: %.2f",
					p.Id, p.OrderId, apy*100, strikePrice)
			}

			product := models.DualInvestmentProduct{
				Symbol:         p.Symbol,
				Direction:      p.Direction,
				StrikePrice:    strikePrice,
				APY:            apy * 100, // API返回的是小数，转换为百分比
				Duration:       p.Duration,
				MinAmount:      minAmount,
				MaxAmount:      maxAmount,
				SettlementTime: settlementTime,
				ProductID:      productIdToSave, // 保存包含orderId的信息
				Status:         status,
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
			} else {
				savedCount++
			}
		}

		if savedCount > 0 {
			log.Printf("同步 %s 产品 %d 个", symbol, savedCount)
		}
		totalSynced += savedCount
	}

	// 清理过期产品
	if err := cfg.DB.Model(&models.DualInvestmentProduct{}).
		Where("settlement_time < ? AND status = ?", time.Now(), "active").
		Update("status", "expired").Error; err != nil {
		log.Printf("更新过期产品失败: %v", err)
	}

	// 清理不在交易对列表中的产品（可选）
	if len(uniqueSymbols) > 0 {
		if err := cfg.DB.Where("symbol NOT IN ? AND deleted_at IS NULL", uniqueSymbols).
			Delete(&models.DualInvestmentProduct{}).Error; err != nil {
			log.Printf("清理无效产品失败: %v", err)
		}
	}

	log.Printf("双币投资产品同步完成，共同步 %d 个产品", totalSynced)
}

// 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
		// 查询所有启用的策略
		var strategies []models.DualInvestmentStrategy
		now := time.Now()
		if err := cfg.DB.Where("enabled = ? AND (status = ? OR (status = ? AND strategy_type != ?)) AND (next_check_time IS NULL OR next_check_time <= ?)",
			true, "active", "completed", "price_trigger", now).
			Find(&strategies).Error; err != nil {
			log.Printf("查询双币投资策略失败: %v", err)
			continue
		}

		if len(strategies) > 0 {
			log.Printf("检查 %d 个双币投资策略", len(strategies))
		}

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
		return
	}

	// 检查投资限额
	if strategy.CurrentInvested >= strategy.TotalInvestmentLimit {
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

	log.Printf("=== 双币投资下单调试信息 ===")
	log.Printf("产品信息:")
	log.Printf("  - ProductID: %s", product.ProductID)
	log.Printf("  - Symbol: %s", product.Symbol)
	log.Printf("  - Direction: %s", product.Direction)
	log.Printf("  - StrikePrice: %.2f", product.StrikePrice)
	log.Printf("  - InvestAmount: %.8f", investAmount)

	// 首先需要获取产品的orderId
	// 重新查询产品列表以获取orderId
	products, err := getDCIProductList(apiKey, secretKey, product.Symbol)
	if err != nil {
		log.Printf("获取产品列表失败: %v", err)
		return false
	}

	// 查找匹配的产品以获取orderId
	var targetProduct *DCIProductListItem

	// 处理可能包含 orderId 的 ProductID
	productIDParts := strings.Split(product.ProductID, "|")
	searchProductID := productIDParts[0] // 获取纯产品ID部分

	for _, p := range products {
		if p.Id == searchProductID {
			targetProduct = &p
			break
		}
	}

	if targetProduct == nil {
		log.Printf("未找到匹配的产品，ProductID: %s, 搜索ID: %s", product.ProductID, searchProductID)
		return false
	}

	log.Printf("找到产品详情: ID=%s, OrderID=%d", targetProduct.Id, targetProduct.OrderId)

	// 根据官方文档构造参数
	params := map[string]interface{}{
		"id":               targetProduct.Id,                         // 产品ID
		"orderId":          fmt.Sprintf("%d", targetProduct.OrderId), // 订单ID需要转为字符串
		"depositAmount":    fmt.Sprintf("%.8f", investAmount),        // 申购金额
		"autoCompoundPlan": "NONE",                                   // 关闭自动复投
		"recvWindow":       5000,
	}

	log.Printf("下单参数: %+v", params)

	res, err := dciRequest(apiKey, secretKey, "POST", "/sapi/v1/dci/product/subscribe", params)
	if err != nil {
		log.Printf("调用双币投资下单API失败: %v", err)

		// 如果orderId为0，尝试只使用id
		if targetProduct.OrderId == 0 {
			log.Printf("OrderID为0，尝试使用产品ID作为OrderID")
			params["orderId"] = targetProduct.Id
			res, err = dciRequest(apiKey, secretKey, "POST", "/sapi/v1/dci/product/subscribe", params)
			if err != nil {
				log.Printf("重试失败: %v", err)
				return false
			}
		} else {
			return false
		}
	}

	log.Printf("下单成功，解析响应...")

	// 解析响应
	var subscribeResp map[string]interface{}
	if err := json.Unmarshal(res, &subscribeResp); err != nil {
		log.Printf("解析下单响应失败: %v, 原始响应: %s", err, string(res))
		return false
	}

	// 从响应中获取positionId
	positionIdFloat, ok := subscribeResp["positionId"].(float64)
	if !ok {
		log.Printf("响应中未找到positionId")
		return false
	}
	positionId := fmt.Sprintf("%.0f", positionIdFloat)

	// 确定投资资产
	investAsset := ""
	if investCoin, ok := subscribeResp["investCoin"].(string); ok {
		investAsset = investCoin
	} else {
		// 如果响应中没有，则根据方向判断
		if product.Direction == "UP" {
			investAsset = product.BaseAsset
		} else {
			investAsset = product.QuoteAsset
		}
	}

	// 创建订单记录
	order := models.DualInvestmentOrder{
		UserID:         user.ID,
		StrategyID:     &strategy.ID,
		ProductID:      product.ID,
		OrderID:        positionId,
		Symbol:         product.Symbol,
		InvestAsset:    investAsset,
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

	log.Printf("双币投资订单创建成功: 用户=%d, 策略=%d, %s %s, 金额=%.2f %s, 年化=%.2f%%, PositionID=%s",
		user.ID, strategy.ID, product.Symbol, product.Direction, investAmount, investAsset, product.APY, positionId)
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
			if time.Now().After(order.SettlementTime) {
				order.Status = "settled"
				cfg.DB.Save(&order)
			}
		}
	}
}

// getDCIProductList 获取双币投资产品列表
func getDCIProductList(apiKey, secretKey, symbol string) ([]DCIProductListItem, error) {
	// 从交易对中解析基础资产和计价资产
	var baseAsset, quoteAsset string
	if strings.HasSuffix(symbol, "USDT") {
		baseAsset = strings.TrimSuffix(symbol, "USDT")
		quoteAsset = "USDT"
	} else if strings.HasSuffix(symbol, "BUSD") {
		baseAsset = strings.TrimSuffix(symbol, "BUSD")
		quoteAsset = "BUSD"
	} else {
		if len(symbol) >= 6 {
			baseAsset = symbol[:3]
			quoteAsset = symbol[3:]
		} else {
			return nil, fmt.Errorf("无法解析交易对: %s", symbol)
		}
	}

	var allProducts []DCIProductListItem

	// 重要修改：修正期权类型和方向的映射关系
	// PUT期权 = 低买策略（用USDT买入基础资产）
	params := map[string]interface{}{
		"optionType":    "PUT",
		"investCoin":    quoteAsset, // USDT - 投资币种
		"exercisedCoin": baseAsset,  // SOL - 行权后获得的币种
		"pageSize":      100,
		"pageIndex":     1,
	}

	res, err := dciRequest(apiKey, secretKey, "GET", "/sapi/v1/dci/product/list", params)
	if err != nil {
		log.Printf("获取 %s PUT产品失败: %v", symbol, err)
	} else {
		var response DCIProductListResponse
		if err := json.Unmarshal(res, &response); err == nil {
			for i := range response.List {
				// PUT期权对应UP方向（低买）
				response.List[i].Direction = "UP"
				if response.List[i].Symbol == "" {
					response.List[i].Symbol = baseAsset + quoteAsset
				}
				response.List[i].BaseAsset = baseAsset
				response.List[i].QuoteAsset = quoteAsset
				response.List[i].InvestAsset = response.List[i].InvestCoin
			}
			allProducts = append(allProducts, response.List...)
		}
	}

	// CALL期权 = 高卖策略（用基础资产卖出）
	params = map[string]interface{}{
		"optionType":    "CALL",
		"investCoin":    baseAsset,  // SOL - 投资币种
		"exercisedCoin": quoteAsset, // USDT - 行权后获得的币种
		"pageSize":      100,
		"pageIndex":     1,
	}

	res, err = dciRequest(apiKey, secretKey, "GET", "/sapi/v1/dci/product/list", params)
	if err != nil {
		log.Printf("获取 %s CALL产品失败: %v", symbol, err)
	} else {
		var response DCIProductListResponse
		if err := json.Unmarshal(res, &response); err == nil {
			for i := range response.List {
				// CALL期权对应DOWN方向（高卖）
				response.List[i].Direction = "DOWN"
				if response.List[i].Symbol == "" {
					response.List[i].Symbol = baseAsset + quoteAsset
				}
				response.List[i].BaseAsset = baseAsset
				response.List[i].QuoteAsset = quoteAsset
				response.List[i].InvestAsset = response.List[i].InvestCoin
			}
			allProducts = append(allProducts, response.List...)
		}
	}

	// 过滤只保留指定交易对的产品
	var filteredProducts []DCIProductListItem
	for _, product := range allProducts {
		if product.Symbol == symbol {
			filteredProducts = append(filteredProducts, product)
		}
	}

	return filteredProducts, nil
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
		return nil, fmt.Errorf("解析持仓响应失败: %v", err)
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
			order.PnL = profitAmount
			order.PnLPercent = (profitAmount / order.InvestAmount) * 100
		} else {
			order.PnL = profitAmount
			order.PnLPercent = 0
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

		log.Printf("双币投资订单结算: 订单=%s, 结算金额=%.2f %s, 盈亏=%.2f",
			order.OrderID, settlementAmount, order.SettlementAsset, profitAmount)
	}

	// 保存更新
	if err := cfg.DB.Save(order).Error; err != nil {
		log.Printf("更新订单 %s 状态失败: %v", order.OrderID, err)
	}
}

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

		if createDualInvestmentOrder(cfg, user, strategy, product, investAmount) {
			log.Printf("自动复投成功: 策略=%d, 原订单=%s, 复投金额=%.2f", strategy.ID, order.OrderID, investAmount)
		}
	}
}

// executeLadderStrategy 执行梯度投资策略
func executeLadderStrategy(cfg *config.Config, strategy models.DualInvestmentStrategy, user models.User, symbol string) {
	// 解密API密钥
	apiKey, err := user.GetDecryptedAPIKey()
	if err != nil {
		return
	}
	secretKey, err := user.GetDecryptedSecretKey()
	if err != nil {
		return
	}

	// 获取当前价格
	client := binance.NewClient(apiKey, secretKey)
	prices, err := client.NewListPricesService().Symbol(symbol).Do(context.Background())
	if err != nil || len(prices) == 0 {
		log.Printf("获取 %s 价格失败: %v", symbol, err)
		return
	}

	_, _ = strconv.ParseFloat(prices[0].Price, 64)

	// 解析梯度配置
	var ladderConfig []models.LadderConfigItem
	if err := json.Unmarshal([]byte(strategy.LadderConfig), &ladderConfig); err != nil {
		log.Printf("解析梯度配置失败: %v", err)
		return
	}

	if len(ladderConfig) == 0 {
		return
	}

	// 根据方向偏好查询产品
	query := cfg.DB.Where("symbol = ? AND status = ?", symbol, "active")

	// 基准价格过滤
	if strategy.DirectionPreference == "UP" {
		query = query.Where("direction = ? AND strike_price <= ?", "UP", strategy.BasePrice)
	} else if strategy.DirectionPreference == "DOWN" {
		query = query.Where("direction = ? AND strike_price >= ?", "DOWN", strategy.BasePrice)
	} else {
		query = query.Where(
			"(direction = ? AND strike_price <= ?) OR (direction = ? AND strike_price >= ?)",
			"UP", strategy.BasePrice, "DOWN", strategy.BasePrice)
	}

	// APY和期限筛选
	if strategy.TargetAPYMin > 0 {
		query = query.Where("apy >= ?", strategy.TargetAPYMin)
	}
	if strategy.TargetAPYMax > 0 {
		query = query.Where("apy <= ?", strategy.TargetAPYMax)
	}
	if strategy.MinDuration > 0 {
		query = query.Where("duration >= ?", strategy.MinDuration)
	}
	if strategy.MaxDuration > 0 {
		query = query.Where("duration <= ?", strategy.MaxDuration)
	}

	var products []models.DualInvestmentProduct
	if err := query.Find(&products).Error; err != nil {
		return
	}

	if len(products) == 0 {
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
		}
	}

	// 更新下次检查时间
	nextCheckTime := time.Now().Add(10 * time.Minute)
	cfg.DB.Model(&strategy).Update("next_check_time", nextCheckTime)

	if successCount > 0 {
		log.Printf("梯度策略 %d 执行: 成功投资 %d 笔，总金额 %.2f", strategy.ID, successCount, totalInvested)
	}
}

// executePriceTriggerStrategy 执行价格触发策略
func executePriceTriggerStrategy(cfg *config.Config, strategy models.DualInvestmentStrategy, user models.User, symbol string) {
	// 解密API密钥
	apiKey, err := user.GetDecryptedAPIKey()
	if err != nil {
		return
	}
	secretKey, err := user.GetDecryptedSecretKey()
	if err != nil {
		return
	}

	// 获取当前价格
	client := binance.NewClient(apiKey, secretKey)
	prices, err := client.NewListPricesService().Symbol(symbol).Do(context.Background())
	if err != nil || len(prices) == 0 {
		return
	}

	currentPrice, _ := strconv.ParseFloat(prices[0].Price, 64)

	// 检查是否触发
	triggered := false
	if strategy.TriggerType == "above" && currentPrice >= strategy.TriggerPrice {
		triggered = true
	} else if strategy.TriggerType == "below" && currentPrice <= strategy.TriggerPrice {
		triggered = true
	}

	if !triggered {
		// 更新下次检查时间
		nextCheckTime := time.Now().Add(1 * time.Minute)
		cfg.DB.Model(&strategy).Update("next_check_time", nextCheckTime)
		return
	}

	log.Printf("价格触发策略 %d 触发: %s 当前价格 %.2f, 触发条件 %s %.2f",
		strategy.ID, symbol, currentPrice, strategy.TriggerType, strategy.TriggerPrice)

	// 重要修改：添加额外的日志说明策略意图
	log.Printf("策略意图: 方向=%s, 基准价格=%.2f, 目标年化=%v%%-%v%%",
		strategy.DirectionPreference, strategy.BasePrice, strategy.TargetAPYMin, strategy.TargetAPYMax)

	// 查找最佳产品
	product := findBestProduct(cfg, strategy, symbol)
	if product == nil {
		log.Printf("未找到符合条件的产品：方向=%s, 基准价格=%.2f",
			strategy.DirectionPreference, strategy.BasePrice)
		// 继续监控
		nextCheckTime := time.Now().Add(5 * time.Minute)
		cfg.DB.Model(&strategy).Update("next_check_time", nextCheckTime)
		return
	}

	// 添加产品信息日志
	log.Printf("找到产品: 方向=%s, 执行价格=%.2f, 年化=%.2f%%, 投资资产=%s",
		product.Direction, product.StrikePrice, product.APY, product.BaseAsset)

	investAmount := calculateInvestAmount(strategy, product)
	if investAmount <= 0 {
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

// findBestProduct 查找最佳产品
func findBestProduct(cfg *config.Config, strategy models.DualInvestmentStrategy, symbol string) *models.DualInvestmentProduct {
	query := cfg.DB.Model(&models.DualInvestmentProduct{}).
		Where("symbol = ? AND status = ?", symbol, "active")

	// 方向筛选
	if strategy.DirectionPreference != "BOTH" {
		query = query.Where("direction = ?", strategy.DirectionPreference)
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

	// 重要修改：基准价格筛选逻辑
	if strategy.BasePrice > 0 {
		// UP方向（低买）：选择执行价格低于基准价格的产品
		// DOWN方向（高卖）：选择执行价格高于基准价格的产品
		if strategy.DirectionPreference == "UP" {
			query = query.Where("strike_price < ?", strategy.BasePrice)
		} else if strategy.DirectionPreference == "DOWN" {
			query = query.Where("strike_price > ?", strategy.BasePrice)
		}
	}

	// 执行价格偏离度筛选
	if strategy.MaxStrikePriceOffset > 0 {
		// 先获取产品，然后在应用层过滤
		var products []models.DualInvestmentProduct
		if err := query.Find(&products).Error; err != nil {
			log.Printf("查询产品失败: %v", err)
			return nil
		}

		// 在应用层进行价格偏离度过滤
		var filteredProducts []models.DualInvestmentProduct
		for _, product := range products {
			if product.CurrentPrice > 0 {
				offset := abs((product.StrikePrice - product.CurrentPrice) / product.CurrentPrice * 100)
				if offset <= strategy.MaxStrikePriceOffset {
					filteredProducts = append(filteredProducts, product)
				}
			}
		}

		// 按APY排序选择最佳
		if len(filteredProducts) > 0 {
			sort.Slice(filteredProducts, func(i, j int) bool {
				return filteredProducts[i].APY > filteredProducts[j].APY
			})

			// 添加日志
			log.Printf("找到 %d 个符合条件的产品，选择APY最高的: %.2f%%",
				len(filteredProducts), filteredProducts[0].APY)

			return &filteredProducts[0]
		}

		log.Printf("没有找到符合价格偏离度要求的产品")
		return nil
	}

	// 没有价格偏离度限制的情况
	var product models.DualInvestmentProduct
	if err := query.Order("apy desc").First(&product).Error; err != nil {
		log.Printf("查询最佳产品失败: %v", err)
		return nil
	}

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
		return 0
	}

	return amount
}

// cleanupOrphanedProducts 清理孤立的双币投资产品（可选功能）
// cleanupOrphanedProducts 清理孤立的双币投资产品（可选功能）
func cleanupOrphanedProducts(cfg *config.Config) {
	// 获取所有已添加的交易对
	symbolMap := make(map[string]bool)

	// 从 Symbol 表获取
	var symbols []models.Symbol
	if err := cfg.DB.Where("deleted_at IS NULL").Find(&symbols).Error; err == nil {
		for _, sym := range symbols {
			symbolMap[sym.Symbol] = true
		}
	}

	// 从 CustomSymbol 表获取
	var customSymbols []models.CustomSymbol
	if err := cfg.DB.Where("deleted_at IS NULL").Find(&customSymbols).Error; err == nil {
		for _, sym := range customSymbols {
			symbolMap[sym.Symbol] = true
		}
	}

	// 转换为数组
	var validSymbols []string
	for symbol := range symbolMap {
		validSymbols = append(validSymbols, symbol)
	}

	if len(validSymbols) == 0 {
		return
	}

	// 软删除不在交易对列表中的产品
	result := cfg.DB.Where("symbol NOT IN ? AND deleted_at IS NULL", validSymbols).
		Delete(&models.DualInvestmentProduct{})

	if result.Error != nil {
		log.Printf("清理孤立产品失败: %v", result.Error)
	} else if result.RowsAffected > 0 {
		log.Printf("清理了 %d 个孤立的双币投资产品", result.RowsAffected)
	}
}
