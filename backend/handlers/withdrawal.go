package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
	"github.com/gorilla/mux"
)

func CreateWithdrawalRuleHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			log.Printf("CreateWithdrawalRuleHandler 无效方法: %s", r.Method)
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
			log.Printf("用户未找到: %s, error: %v", username, err)
			http.Error(w, `{"error": "用户未找到"}`, http.StatusNotFound)
			return
		}
		var rule models.Withdrawal
		if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
			log.Printf("解码请求体失败: %v", err)
			http.Error(w, `{"error": "无效的请求体"}`, http.StatusBadRequest)
			return
		}
		rule.UserID = user.ID
		if err := cfg.DB.Create(&rule).Error; err != nil {
			log.Printf("为用户 %d 创建取款规则失败: %v", user.ID, err)
			http.Error(w, `{"error": "创建取款规则失败"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"message": "取款规则创建成功"}); err != nil {
			log.Printf("编码响应失败: %v", err)
		}
		log.Printf("为用户 %d 创建取款规则", user.ID)
	}
}

func ListWithdrawalRulesHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			log.Printf("ListWithdrawalRulesHandler 无效方法: %s", r.Method)
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
			log.Printf("用户未找到: %s, error: %v", username, err)
			http.Error(w, `{"error": "用户未找到"}`, http.StatusNotFound)
			return
		}
		var rules []models.Withdrawal
		if err := cfg.DB.Where("user_id = ? AND deleted_at IS NULL", user.ID).Find(&rules).Error; err != nil {
			log.Printf("获取用户 %d 的取款规则失败: %v", user.ID, err)
			http.Error(w, `{"error": "获取取款规则失败"}`, http.StatusInternalServerError)
			return
		}
		response := map[string]interface{}{
			"rules": rules,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("编码取款规则响应失败: %v", err)
			http.Error(w, `{"error": "编码响应失败"}`, http.StatusInternalServerError)
			return
		}
		log.Printf("为用户 %d 返回 %d 条取款规则", user.ID, len(rules))
	}
}

func UpdateWithdrawalRuleHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			log.Printf("UpdateWithdrawalRuleHandler 无效方法: %s", r.Method)
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
			log.Printf("用户未找到: %s, error: %v", username, err)
			http.Error(w, `{"error": "用户未找到"}`, http.StatusNotFound)
			return
		}
		vars := mux.Vars(r)
		id := vars["id"]
		var rule models.Withdrawal
		if err := cfg.DB.Where("id = ? AND user_id = ?", id, user.ID).First(&rule).Error; err != nil {
			log.Printf("取款规则 %s 未找到，用户 %d: %v", id, user.ID, err)
			http.Error(w, `{"error": "取款规则未找到"}`, http.StatusNotFound)
			return
		}
		var updatedRule models.Withdrawal
		if err := json.NewDecoder(r.Body).Decode(&updatedRule); err != nil {
			log.Printf("解码请求体失败: %v", err)
			http.Error(w, `{"error": "无效的请求体"}`, http.StatusBadRequest)
			return
		}
		rule.Asset = updatedRule.Asset
		rule.Address = updatedRule.Address
		rule.Threshold = updatedRule.Threshold
		rule.Amount = updatedRule.Amount
		rule.Enabled = updatedRule.Enabled
		if err := cfg.DB.Save(&rule).Error; err != nil {
			log.Printf("更新取款规则 %s 失败，用户 %d: %v", id, user.ID, err)
			http.Error(w, `{"error": "更新取款规则失败"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"message": "取款规则更新成功"}); err != nil {
			log.Printf("编码响应失败: %v", err)
		}
		log.Printf("取款规则 %s 更新成功，用户 %d", id, user.ID)
	}
}

func DeleteWithdrawalRuleHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			log.Printf("DeleteWithdrawalRuleHandler 无效方法: %s", r.Method)
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
			log.Printf("用户未找到: %s, error: %v", username, err)
			http.Error(w, `{"error": "用户未找到"}`, http.StatusNotFound)
			return
		}
		vars := mux.Vars(r)
		id := vars["id"]
		var rule models.Withdrawal
		if err := cfg.DB.Where("id = ? AND user_id = ?", id, user.ID).First(&rule).Error; err != nil {
			log.Printf("取款规则 %s 未找到，用户 %d: %v", id, user.ID, err)
			http.Error(w, `{"error": "取款规则未找到"}`, http.StatusNotFound)
			return
		}
		if err := cfg.DB.Delete(&rule).Error; err != nil {
			log.Printf("删除取款规则 %s 失败，用户 %d: %v", id, user.ID, err)
			http.Error(w, `{"error": "删除取款规则失败"}`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]interface{}{"message": "取款规则删除成功"}); err != nil {
			log.Printf("编码响应失败: %v", err)
		}
		log.Printf("取款规则 %s 删除成功，用户 %d", id, user.ID)
	}
}
