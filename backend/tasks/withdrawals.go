package tasks

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2"
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
)

// CheckWithdrawals 定期检查并执行自动提币规则
func CheckWithdrawals(cfg *config.Config) {
	ticker := time.NewTicker(5 * time.Minute) // 每5分钟检查一次
	defer ticker.Stop()

	// 立即执行一次
	processWithdrawalRules(cfg)

	for range ticker.C {
		processWithdrawalRules(cfg)
	}
}

// processWithdrawalRules 处理所有启用的提币规则
func processWithdrawalRules(cfg *config.Config) {
	var rules []models.Withdrawal
	if err := cfg.DB.Where("enabled = ? AND deleted_at IS NULL", true).Find(&rules).Error; err != nil {
		log.Printf("获取自动提币规则失败: %v", err)
		return
	}

	if len(rules) == 0 {
		return
	}

	log.Printf("检查 %d 个自动提币规则", len(rules))

	// 按用户分组规则
	userRules := make(map[uint][]models.Withdrawal)
	for _, rule := range rules {
		userRules[rule.UserID] = append(userRules[rule.UserID], rule)
	}

	// 处理每个用户的规则
	for userID, userRuleList := range userRules {
		processUserWithdrawalRules(cfg, userID, userRuleList)
	}
}

// processUserWithdrawalRules 处理单个用户的提币规则
func processUserWithdrawalRules(cfg *config.Config, userID uint, rules []models.Withdrawal) {
	// 获取用户信息
	var user models.User
	if err := cfg.DB.First(&user, userID).Error; err != nil {
		log.Printf("提币规则用户未找到: ID=%d", userID)
		return
	}

	// 解密API密钥
	apiKey, err := user.GetDecryptedAPIKey()
	if err != nil {
		log.Printf("解密用户 %d API Key失败: %v", userID, err)
		return
	}
	secretKey, err := user.GetDecryptedSecretKey()
	if err != nil {
		log.Printf("解密用户 %d Secret Key失败: %v", userID, err)
		return
	}

	if apiKey == "" || secretKey == "" {
		log.Printf("用户 %d 未设置 API 密钥，跳过提币规则检查", user.ID)
		return
	}

	client := binance.NewClient(apiKey, secretKey)

	// 获取账户余额
	account, err := client.NewGetAccountService().Do(context.Background())
	if err != nil {
		log.Printf("获取用户 %d 账户余额失败: %v", userID, err)
		return
	}

	// 创建余额映射
	balanceMap := make(map[string]float64)
	for _, balance := range account.Balances {
		free, _ := strconv.ParseFloat(balance.Free, 64)
		if free > 0 {
			balanceMap[balance.Asset] = free
		}
	}

	// 检查每个规则
	for _, rule := range rules {
		processWithdrawalRule(cfg, client, user, rule, balanceMap)
	}
}

// processWithdrawalRule 处理单个提币规则
func processWithdrawalRule(cfg *config.Config, client *binance.Client, user models.User, rule models.Withdrawal, balanceMap map[string]float64) {
	balance, exists := balanceMap[rule.Asset]
	if !exists || balance == 0 {
		log.Printf("用户 %d 的 %s 余额为0，跳过规则 %d", user.ID, rule.Asset, rule.ID)
		return
	}

	// 检查是否达到阈值
	if balance < rule.Threshold {
		log.Printf("用户 %d 的 %s 余额 %.4f 未达到阈值 %.4f，跳过规则 %d",
			user.ID, rule.Asset, balance, rule.Threshold, rule.ID)
		return
	}

	// 注意：暂时跳过网络验证，因为数据库模型中还没有Network字段
	// TODO: 等数据库模型更新后再启用网络验证
	// if rule.Network == "" {
	//     log.Printf("规则 %d 未配置网络，跳过提币", rule.ID)
	//     return
	// }
	// if !isValidAssetNetwork(rule.Asset, rule.Network) {
	//     log.Printf("币种 %s 与网络 %s 不兼容，跳过规则 %d", rule.Asset, rule.Network, rule.ID)
	//     return
	// }

	// 确定提币金额
	var withdrawAmount float64
	if rule.Amount == 0 {
		// 如果规则金额为0，提取最大可用金额
		withdrawAmount = balance
		log.Printf("规则 %d 设置为提取最大金额，将提取 %.4f %s", rule.ID, withdrawAmount, rule.Asset)
	} else {
		// 否则提取指定金额，但不超过可用余额
		withdrawAmount = rule.Amount
		if withdrawAmount > balance {
			withdrawAmount = balance
			log.Printf("规则 %d 指定金额 %.4f 超过可用余额，调整为 %.4f %s",
				rule.ID, rule.Amount, withdrawAmount, rule.Asset)
		}
	}

	// 获取提币手续费和最小提币金额
	// 注意：暂时使用默认网络信息，等数据库模型更新后再使用rule.Network
	withdrawInfo, err := getWithdrawInfo(client, rule.Asset, "")
	if err != nil {
		log.Printf("获取 %s 提币信息失败: %v", rule.Asset, err)
		return
	}

	// 检查是否满足最小提币金额
	if withdrawAmount < withdrawInfo.MinWithdrawAmount {
		log.Printf("提币金额 %.4f %s 小于最小提币金额 %.4f，跳过",
			withdrawAmount, rule.Asset, withdrawInfo.MinWithdrawAmount)
		return
	}

	// 计算实际到账金额（扣除手续费）
	actualAmount := withdrawAmount - withdrawInfo.WithdrawFee
	if actualAmount <= 0 {
		log.Printf("扣除手续费后金额为负，跳过提币")
		return
	}

	// 执行提币
	log.Printf("准备提币: 用户=%d, 资产=%s, 金额=%.4f, 地址=%s",
		user.ID, rule.Asset, withdrawAmount, rule.Address)

	// 创建提币请求
	withdrawReq := client.NewCreateWithdrawService().
		Coin(rule.Asset).
		Address(rule.Address).
		Amount(fmt.Sprintf("%.4f", withdrawAmount))
	// 注意：暂时不添加网络参数，等数据库模型更新后再启用
	// .Network(rule.Network)

	withdrawResp, err := withdrawReq.Do(context.Background())
	if err != nil {
		log.Printf("提币失败: %v", err)
		// 记录失败历史
		recordWithdrawalHistory(cfg, user.ID, rule, withdrawAmount, "", "failed", err.Error())
		return
	}

	// 记录成功的提币历史
	recordWithdrawalHistory(cfg, user.ID, rule, withdrawAmount, withdrawResp.ID, "processing", "")

	log.Printf("提币成功: ID=%s, 用户=%d, %s %.8f -> %s",
		withdrawResp.ID, user.ID, rule.Asset, withdrawAmount, rule.Address)
}

// WithdrawInfo 提币信息
type WithdrawInfo struct {
	MinWithdrawAmount float64
	WithdrawFee       float64
}

// getWithdrawInfo 获取特定网络的提币信息
func getWithdrawInfo(client *binance.Client, asset, network string) (*WithdrawInfo, error) {
	// 根据币种和网络返回相应的提币信息
	networkInfo := getAssetNetworkInfo(asset, network)
	if networkInfo != nil {
		return networkInfo, nil
	}

	// 如果没有预设信息，返回默认值
	return &WithdrawInfo{
		MinWithdrawAmount: 0.001,
		WithdrawFee:       0.0005,
	}, nil
}

// getAssetNetworkInfo 获取币种在特定网络的信息
func getAssetNetworkInfo(asset, network string) *WithdrawInfo {
	// 预设的币种网络信息映射
	assetNetworkMap := map[string]map[string]*WithdrawInfo{
		"BTC": {
			"BTC": {
				MinWithdrawAmount: 0.001,
				WithdrawFee:       0.0005,
			},
			"BEP20": {
				MinWithdrawAmount: 0.0001,
				WithdrawFee:       0.0000035,
			},
		},
		"ETH": {
			"ERC20": {
				MinWithdrawAmount: 0.01,
				WithdrawFee:       0.005,
			},
			"BEP20": {
				MinWithdrawAmount: 0.001,
				WithdrawFee:       0.0002,
			},
			"ARBITRUM": {
				MinWithdrawAmount: 0.001,
				WithdrawFee:       0.0001,
			},
			"POLYGON": {
				MinWithdrawAmount: 0.001,
				WithdrawFee:       0.0001,
			},
		},
		"USDT": {
			"ERC20": {
				MinWithdrawAmount: 10,
				WithdrawFee:       25,
			},
			"TRC20": {
				MinWithdrawAmount: 1,
				WithdrawFee:       1,
			},
			"BEP20": {
				MinWithdrawAmount: 1,
				WithdrawFee:       0.8,
			},
			"POLYGON": {
				MinWithdrawAmount: 1,
				WithdrawFee:       0.8,
			},
			"ARBITRUM": {
				MinWithdrawAmount: 1,
				WithdrawFee:       0.8,
			},
			"OPTIMISM": {
				MinWithdrawAmount: 1,
				WithdrawFee:       0.8,
			},
		},
		"USDC": {
			"ERC20": {
				MinWithdrawAmount: 10,
				WithdrawFee:       25,
			},
			"TRC20": {
				MinWithdrawAmount: 1,
				WithdrawFee:       1,
			},
			"BEP20": {
				MinWithdrawAmount: 1,
				WithdrawFee:       0.8,
			},
			"POLYGON": {
				MinWithdrawAmount: 1,
				WithdrawFee:       0.8,
			},
			"ARBITRUM": {
				MinWithdrawAmount: 1,
				WithdrawFee:       0.1,
			},
		},
		"BNB": {
			"BEP20": {
				MinWithdrawAmount: 0.01,
				WithdrawFee:       0.005,
			},
			"BEP2": {
				MinWithdrawAmount: 0.01,
				WithdrawFee:       0.00075,
			},
		},
		"ADA": {
			"ADA": {
				MinWithdrawAmount: 1,
				WithdrawFee:       1,
			},
		},
		"DOT": {
			"DOT": {
				MinWithdrawAmount: 1,
				WithdrawFee:       0.1,
			},
		},
		"SOL": {
			"SOL": {
				MinWithdrawAmount: 0.01,
				WithdrawFee:       0.01,
			},
		},
		"MATIC": {
			"POLYGON": {
				MinWithdrawAmount: 0.1,
				WithdrawFee:       0.01,
			},
			"ERC20": {
				MinWithdrawAmount: 10,
				WithdrawFee:       15,
			},
		},
		"AVAX": {
			"AVAXC": {
				MinWithdrawAmount: 0.01,
				WithdrawFee:       0.005,
			},
		},
		"TRX": {
			"TRC20": {
				MinWithdrawAmount: 1,
				WithdrawFee:       1,
			},
		},
	}

	if assetInfo, exists := assetNetworkMap[asset]; exists {
		if networkInfo, exists := assetInfo[network]; exists {
			return networkInfo
		}
	}

	return nil
}

// isValidAssetNetwork 验证币种和网络的兼容性
func isValidAssetNetwork(asset, network string) bool {
	validNetworks := getSupportedNetworks(asset)
	for _, validNetwork := range validNetworks {
		if validNetwork == network {
			return true
		}
	}
	return false
}

// getSupportedNetworks 获取币种支持的网络列表
func getSupportedNetworks(asset string) []string {
	supportedNetworks := map[string][]string{
		"BTC":   {"BTC", "BEP20"},
		"ETH":   {"ERC20", "BEP20", "ARBITRUM", "POLYGON"},
		"USDT":  {"ERC20", "TRC20", "BEP20", "POLYGON", "ARBITRUM", "OPTIMISM"},
		"USDC":  {"ERC20", "TRC20", "BEP20", "POLYGON", "ARBITRUM"},
		"BNB":   {"BEP20", "BEP2"},
		"ADA":   {"ADA"},
		"DOT":   {"DOT"},
		"SOL":   {"SOL"},
		"MATIC": {"POLYGON", "ERC20"},
		"AVAX":  {"AVAXC"},
		"TRX":   {"TRC20"},
		"LTC":   {"LTC"},
		"BCH":   {"BCH"},
		"XRP":   {"XRP"},
		"DOGE":  {"DOGE"},
		"SHIB":  {"ERC20", "BEP20"},
		"UNI":   {"ERC20", "BEP20"},
		"LINK":  {"ERC20", "BEP20"},
		"ATOM":  {"ATOM"},
		"FTM":   {"FTM", "ERC20", "BEP20"},
		"NEAR":  {"NEAR"},
		"ALGO":  {"ALGO"},
		"VET":   {"VET"},
		"ICP":   {"ICP"},
		"THETA": {"THETA", "ERC20", "BEP20"},
		"FIL":   {"FIL"},
		"XTZ":   {"XTZ"},
		"EOS":   {"EOS"},
		"AAVE":  {"ERC20", "BEP20"},
		"MKR":   {"ERC20"},
		"COMP":  {"ERC20", "BEP20"},
		"YFI":   {"ERC20", "BEP20"},
		"SNX":   {"ERC20", "BEP20"},
		"CRV":   {"ERC20", "BEP20"},
		"SUSHI": {"ERC20", "BEP20"},
		"1INCH": {"ERC20", "BEP20"},
		"BAT":   {"ERC20", "BEP20"},
		"ZRX":   {"ERC20", "BEP20"},
		"ENJ":   {"ERC20", "BEP20"},
		"MANA":  {"ERC20", "BEP20", "POLYGON"},
		"SAND":  {"ERC20", "BEP20", "POLYGON"},
		"AXS":   {"ERC20", "BEP20"},
		"GALA":  {"ERC20", "BEP20"},
		"CHZ":   {"ERC20", "BEP20"},
	}

	if networks, exists := supportedNetworks[asset]; exists {
		return networks
	}

	// 如果没有预设的网络，返回一些通用网络
	return []string{"ERC20", "BEP20"}
}

// recordWithdrawalHistory 记录提币历史
func recordWithdrawalHistory(cfg *config.Config, userID uint, rule models.Withdrawal, amount float64, withdrawalID, status, errorMsg string) {
	history := models.WithdrawalHistory{
		UserID: userID,
		Asset:  rule.Asset,
		// Network:      rule.Network, // TODO: 等数据库模型更新后再启用
		Amount:       amount,
		Address:      rule.Address,
		WithdrawalID: withdrawalID,
		Status:       status,
	}

	if errorMsg != "" {
		// 可以将错误信息存储在某个字段中，或记录到日志
		log.Printf("提币错误: %s", errorMsg)
	}

	if err := cfg.DB.Create(&history).Error; err != nil {
		log.Printf("记录提币历史失败: %v", err)
	}
}

// GetSupportedNetworksForAsset 导出函数供API使用，获取币种支持的网络
func GetSupportedNetworksForAsset(asset string) []string {
	return getSupportedNetworks(asset)
}

// ValidateAssetNetwork 导出函数供API使用，验证币种网络兼容性
func ValidateAssetNetwork(asset, network string) bool {
	return isValidAssetNetwork(asset, network)
}

// GetNetworkInfo 导出函数供API使用，获取网络信息
func GetNetworkInfo(asset, network string) *WithdrawInfo {
	return getAssetNetworkInfo(asset, network)
}
