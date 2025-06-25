package tasks

import (
	"log"
	"time"

	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
)

// CheckWithdrawals 定期检查并执行自动取款规则
func CheckWithdrawals(cfg *config.Config) {
	for {
		var rules []models.Withdrawal
		if err := cfg.DB.Where("enabled = ? AND deleted_at IS NULL", true).Find(&rules).Error; err != nil {
			log.Printf("获取自动取款规则失败: %v", err)
			time.Sleep(60 * time.Second)
			continue
		}
		for _, rule := range rules {
			var user models.User
			if err := cfg.DB.First(&user, rule.UserID).Error; err != nil {
				log.Printf("用户未找到: ID=%d", rule.UserID)
				continue
			}
			if user.APIKey == "" || user.SecretKey == "" {
				log.Printf("用户 %d 未设置 API 密钥，跳过取款规则 %d", user.ID, rule.ID)
				continue
			}
			// 这里应该有逻辑来检查余额并执行取款
			// 例如，使用 Binance API 获取余额，然后如果余额超过阈值，执行取款
			// 但由于具体实现未提供，我将假设有相应的逻辑
		}
		time.Sleep(60 * time.Second)
	}
}
