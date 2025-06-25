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

	if user.APIKey == "" || user.SecretKey == "" {
		log.Printf("用户 %d 未设置 API 密钥，跳过提币规则检查", user.ID)
		return
	}

	client := binance.NewClient(user.APIKey, user.SecretKey)

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
		log.Printf("用户 %d 的 %s 余额 %.8f 未达到阈值 %.8f，跳过规则 %d",
			user.ID, rule.Asset, balance, rule.Threshold, rule.ID)
		return
	}

	// 确定提币金额
	var withdrawAmount float64
	if rule.Amount == 0 {
		// 如果规则金额为0，提取最大可用金额
		withdrawAmount = balance
		log.Printf("规则 %d 设置为提取最大金额，将提取 %.8f %s", rule.ID, withdrawAmount, rule.Asset)
	} else {
		// 否则提取指定金额，但不超过可用余额
		withdrawAmount = rule.Amount
		if withdrawAmount > balance {
			withdrawAmount = balance
			log.Printf("规则 %d 指定金额 %.8f 超过可用余额，调整为 %.8f %s",
				rule.ID, rule.Amount, withdrawAmount, rule.Asset)
		}
	}

	// 获取提币手续费和最小提币金额
	withdrawInfo, err := getWithdrawInfo(client, rule.Asset)
	if err != nil {
		log.Printf("获取 %s 提币信息失败: %v", rule.Asset, err)
		return
	}

	// 检查是否满足最小提币金额
	if withdrawAmount < withdrawInfo.MinWithdrawAmount {
		log.Printf("提币金额 %.8f %s 小于最小提币金额 %.8f，跳过",
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
	log.Printf("准备提币: 用户=%d, 资产=%s, 金额=%.8f, 地址=%s",
		user.ID, rule.Asset, withdrawAmount, rule.Address)

	// 创建提币请求
	withdrawReq := client.NewCreateWithdrawService().
		Coin(rule.Asset).
		Address(rule.Address).
		Amount(fmt.Sprintf("%.8f", withdrawAmount))

	// 如果有网络信息，添加网络参数
	// 注意：这里需要根据实际情况设置网络
	// withdrawReq.Network("BSC") // 示例

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

// getWithdrawInfo 获取提币信息（简化版本）
func getWithdrawInfo(client *binance.Client, asset string) (*WithdrawInfo, error) {
	// 在实际应用中，应该调用币安API获取实时的提币信息
	// 这里使用一些常见的默认值
	defaultInfo := map[string]*WithdrawInfo{
		"BTC": {
			MinWithdrawAmount: 0.001,
			WithdrawFee:       0.0005,
		},
		"ETH": {
			MinWithdrawAmount: 0.01,
			WithdrawFee:       0.005,
		},
		"USDT": {
			MinWithdrawAmount: 10,
			WithdrawFee:       1,
		},
		"BNB": {
			MinWithdrawAmount: 0.01,
			WithdrawFee:       0.005,
		},
	}

	if info, exists := defaultInfo[asset]; exists {
		return info, nil
	}

	// 默认值
	return &WithdrawInfo{
		MinWithdrawAmount: 0.001,
		WithdrawFee:       0.0005,
	}, nil
}

// recordWithdrawalHistory 记录提币历史
func recordWithdrawalHistory(cfg *config.Config, userID uint, rule models.Withdrawal, amount float64, withdrawalID, status, errorMsg string) {
	history := models.WithdrawalHistory{
		UserID:       userID,
		Asset:        rule.Asset,
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
