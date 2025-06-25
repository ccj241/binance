package middleware

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// ValidationError 验证错误结构
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationMiddleware 数据验证中间件
func ValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只对 POST、PUT、PATCH 请求进行验证
		if c.Request.Method != "POST" && c.Request.Method != "PUT" && c.Request.Method != "PATCH" {
			c.Next()
			return
		}

		// 根据路径进行特定验证
		path := c.Request.URL.Path

		switch {
		case strings.HasSuffix(path, "/api-key"):
			validateAPIKey(c)
		case strings.HasSuffix(path, "/strategy"):
			validateStrategy(c)
		case strings.HasSuffix(path, "/order"):
			validateOrder(c)
		case strings.HasSuffix(path, "/withdrawals") || strings.Contains(path, "/withdrawals/"):
			validateWithdrawal(c)
		default:
			c.Next()
		}
	}
}

// validateAPIKey 验证API密钥
func validateAPIKey(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "无效的请求数据",
			"details": err.Error(),
		})
		c.Abort()
		return
	}

	errors := []ValidationError{}

	// 验证 API Key
	apiKey, ok := data["apiKey"].(string)
	if !ok || apiKey == "" {
		errors = append(errors, ValidationError{
			Field:   "apiKey",
			Message: "API Key 不能为空",
		})
	} else if len(apiKey) != 64 {
		errors = append(errors, ValidationError{
			Field:   "apiKey",
			Message: "API Key 长度必须为 64 个字符",
		})
	}

	// 验证 API Secret
	apiSecret, ok := data["apiSecret"].(string)
	if !ok || apiSecret == "" {
		errors = append(errors, ValidationError{
			Field:   "apiSecret",
			Message: "API Secret 不能为空",
		})
	} else if len(apiSecret) != 64 {
		errors = append(errors, ValidationError{
			Field:   "apiSecret",
			Message: "API Secret 长度必须为 64 个字符",
		})
	}

	if len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "验证失败",
			"errors": errors,
		})
		c.Abort()
		return
	}

	// 重新绑定数据以供后续使用
	c.Set("validated_data", data)
	c.Next()
}

// validateStrategy 验证策略
func validateStrategy(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "读取请求体失败",
		})
		c.Abort()
		return
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求数据",
		})
		c.Abort()
		return
	}

	errors := []ValidationError{}

	// 验证交易对
	symbol, ok := data["symbol"].(string)
	if !ok || symbol == "" {
		errors = append(errors, ValidationError{
			Field:   "symbol",
			Message: "交易对不能为空",
		})
	} else if !isValidSymbol(symbol) {
		errors = append(errors, ValidationError{
			Field:   "symbol",
			Message: "无效的交易对格式",
		})
	}

	// 验证策略类型
	strategyType, ok := data["strategyType"].(string)
	if !ok || strategyType == "" {
		errors = append(errors, ValidationError{
			Field:   "strategyType",
			Message: "策略类型不能为空",
		})
	} else if strategyType != "simple" && strategyType != "iceberg" && strategyType != "custom" {
		errors = append(errors, ValidationError{
			Field:   "strategyType",
			Message: "策略类型必须是 simple、iceberg 或 custom",
		})
	}

	// 验证方向
	side, ok := data["side"].(string)
	if !ok || side == "" {
		errors = append(errors, ValidationError{
			Field:   "side",
			Message: "交易方向不能为空",
		})
	} else if side != "BUY" && side != "SELL" {
		errors = append(errors, ValidationError{
			Field:   "side",
			Message: "交易方向必须是 BUY 或 SELL",
		})
	}

	// 验证价格
	price, ok := getFloat64(data["price"])
	if !ok || price <= 0 {
		errors = append(errors, ValidationError{
			Field:   "price",
			Message: "价格必须大于 0",
		})
	}

	// 验证数量
	totalQuantity, ok := getFloat64(data["totalQuantity"])
	if !ok || totalQuantity <= 0 {
		errors = append(errors, ValidationError{
			Field:   "totalQuantity",
			Message: "总数量必须大于 0",
		})
	}

	// 验证自定义策略的额外参数
	if strategyType == "custom" {
		if side == "BUY" {
			buyQuantities, _ := data["buyQuantities"].([]interface{})
			buyDepthLevels, _ := data["buyDepthLevels"].([]interface{})

			if len(buyQuantities) == 0 {
				errors = append(errors, ValidationError{
					Field:   "buyQuantities",
					Message: "买入策略需要设置数量分配",
				})
			}

			if len(buyDepthLevels) == 0 {
				errors = append(errors, ValidationError{
					Field:   "buyDepthLevels",
					Message: "买入策略需要设置深度级别",
				})
			}

			if len(buyQuantities) != len(buyDepthLevels) {
				errors = append(errors, ValidationError{
					Field:   "buyQuantities",
					Message: "数量分配和深度级别的数量必须相同",
				})
			}
		} else {
			sellQuantities, _ := data["sellQuantities"].([]interface{})
			sellDepthLevels, _ := data["sellDepthLevels"].([]interface{})

			if len(sellQuantities) == 0 {
				errors = append(errors, ValidationError{
					Field:   "sellQuantities",
					Message: "卖出策略需要设置数量分配",
				})
			}

			if len(sellDepthLevels) == 0 {
				errors = append(errors, ValidationError{
					Field:   "sellDepthLevels",
					Message: "卖出策略需要设置深度级别",
				})
			}

			if len(sellQuantities) != len(sellDepthLevels) {
				errors = append(errors, ValidationError{
					Field:   "sellQuantities",
					Message: "数量分配和深度级别的数量必须相同",
				})
			}
		}
	}

	if len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "验证失败",
			"errors": errors,
		})
		c.Abort()
		return
	}

	// 重新设置请求体
	c.Request.Body = &bodyReader{data: body}
	c.Next()
}

// validateOrder 验证订单
func validateOrder(c *gin.Context) {
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求数据",
		})
		c.Abort()
		return
	}

	errors := []ValidationError{}

	// 验证交易对
	symbol, ok := data["symbol"].(string)
	if !ok || symbol == "" {
		errors = append(errors, ValidationError{
			Field:   "symbol",
			Message: "交易对不能为空",
		})
	}

	// 验证方向
	side, ok := data["side"].(string)
	if !ok || side == "" {
		errors = append(errors, ValidationError{
			Field:   "side",
			Message: "交易方向不能为空",
		})
	} else if side != "BUY" && side != "SELL" {
		errors = append(errors, ValidationError{
			Field:   "side",
			Message: "交易方向必须是 BUY 或 SELL",
		})
	}

	// 验证价格
	price, ok := getFloat64(data["price"])
	if !ok || price <= 0 {
		errors = append(errors, ValidationError{
			Field:   "price",
			Message: "价格必须大于 0",
		})
	}

	// 验证数量
	quantity, ok := getFloat64(data["quantity"])
	if !ok || quantity <= 0 {
		errors = append(errors, ValidationError{
			Field:   "quantity",
			Message: "数量必须大于 0",
		})
	}

	if len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "验证失败",
			"errors": errors,
		})
		c.Abort()
		return
	}

	c.Set("validated_data", data)
	c.Next()
}

// validateWithdrawal 验证提币规则 - 修复版本
func validateWithdrawal(c *gin.Context) {
	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "读取请求体失败",
		})
		c.Abort()
		return
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的请求数据",
		})
		c.Abort()
		return
	}

	errors := []ValidationError{}

	// 验证资产
	asset, ok := data["asset"].(string)
	if !ok || asset == "" {
		errors = append(errors, ValidationError{
			Field:   "asset",
			Message: "资产不能为空",
		})
	} else if len(asset) < 2 || len(asset) > 10 {
		errors = append(errors, ValidationError{
			Field:   "asset",
			Message: "资产名称长度必须在2-10个字符之间",
		})
	}

	// 验证地址
	address, ok := data["address"].(string)
	if !ok || address == "" {
		errors = append(errors, ValidationError{
			Field:   "address",
			Message: "提币地址不能为空",
		})
	} else if len(address) < 10 {
		errors = append(errors, ValidationError{
			Field:   "address",
			Message: "提币地址格式不正确",
		})
	}

	// 验证阈值
	threshold, ok := getFloat64(data["threshold"])
	if !ok || threshold <= 0 {
		errors = append(errors, ValidationError{
			Field:   "threshold",
			Message: "阈值必须大于 0",
		})
	}

	// 验证金额 - 允许为0（表示提取全部）
	amount, ok := getFloat64(data["amount"])
	if !ok || amount < 0 {
		errors = append(errors, ValidationError{
			Field:   "amount",
			Message: "提币金额不能为负数",
		})
	}

	// 验证启用状态（可选字段）
	if enabled, exists := data["enabled"]; exists {
		if _, ok := enabled.(bool); !ok {
			errors = append(errors, ValidationError{
				Field:   "enabled",
				Message: "启用状态必须是布尔值",
			})
		}
	}

	if len(errors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "验证失败",
			"errors": errors,
		})
		c.Abort()
		return
	}

	// 重新设置请求体以供后续使用
	c.Request.Body = &bodyReader{data: body}
	c.Set("validated_data", data)
	c.Next()
}

// Helper functions

// isValidSymbol 验证交易对格式
func isValidSymbol(symbol string) bool {
	// 币安交易对格式：基础货币+计价货币，如 BTCUSDT
	matched, _ := regexp.MatchString(`^[A-Z]{2,10}[A-Z]{2,10}$`, symbol)
	return matched
}

// isValidAddress 验证地址格式（简单验证）
func isValidAddress(address string) bool {
	// 基本的地址格式验证
	// BTC地址：以1、3或bc1开头
	// ETH地址：以0x开头，40个十六进制字符
	// 这里只做基本验证
	if strings.HasPrefix(address, "0x") && len(address) == 42 {
		return true
	}
	if (strings.HasPrefix(address, "1") || strings.HasPrefix(address, "3") || strings.HasPrefix(address, "bc1")) && len(address) >= 26 && len(address) <= 62 {
		return true
	}
	// 其他格式暂时通过
	return len(address) >= 10 && len(address) <= 100
}

// getFloat64 安全获取float64值
func getFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}

// bodyReader 用于重新读取请求体
type bodyReader struct {
	data []byte
	pos  int
}

func (r *bodyReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (r *bodyReader) Close() error {
	return nil
}
