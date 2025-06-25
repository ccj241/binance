package controllers

import (
	"fmt"
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
)

type UserController struct {
	Config *config.Config
}

type APIKeyInput struct {
	APIKey    string `json:"apiKey" binding:"required"`
	APISecret string `json:"apiSecret" binding:"required"`
}

// getUserIDFromContext 从context中获取用户ID
func (ctrl *UserController) getUserIDFromContext(c *gin.Context) (uint, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		log.Printf("上下文缺少 user_id")
		return 0, fmt.Errorf("用户ID不存在")
	}

	uid, ok := userID.(uint)
	if !ok {
		log.Printf("user_id 类型错误: %T", userID)
		return 0, fmt.Errorf("用户ID类型错误")
	}

	return uid, nil
}

// SetAPIKey 保存用户的 API 密钥
func (ctrl *UserController) SetAPIKey(c *gin.Context) {
	var input APIKeyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("无效的请求体: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误", "details": err.Error()})
		return
	}

	userID, err := ctrl.getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户认证"})
		return
	}

	var user models.User
	if err := ctrl.Config.DB.First(&user, userID).Error; err != nil {
		log.Printf("用户未找到: userID=%d, error=%v", userID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
		return
	}

	log.Printf("为用户 %d 保存 API 密钥: APIKey=%s", userID, maskAPIKey(input.APIKey))
	user.APIKey = input.APIKey
	user.SecretKey = input.APISecret
	if err := ctrl.Config.DB.Save(&user).Error; err != nil {
		log.Printf("保存 API 密钥失败，用户 %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存 API 密钥失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API 密钥更新成功"})
}

// GetAPIKey 获取用户的 API 密钥（部分掩码）
func (ctrl *UserController) GetAPIKey(c *gin.Context) {
	userID, err := ctrl.getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户认证"})
		return
	}

	var user models.User
	if err := ctrl.Config.DB.First(&user, userID).Error; err != nil {
		log.Printf("用户未找到: userID=%d, error=%v", userID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
		return
	}

	var maskedAPIKey, maskedSecretKey string
	if user.APIKey != "" {
		maskedAPIKey = maskAPIKey(user.APIKey)
	}
	if user.SecretKey != "" {
		maskedSecretKey = maskAPIKey(user.SecretKey)
	}

	c.JSON(http.StatusOK, gin.H{
		"apiKey":    maskedAPIKey,
		"secretKey": maskedSecretKey,
	})
}

// DeleteAPIKey 删除用户的 API 密钥
func (ctrl *UserController) DeleteAPIKey(c *gin.Context) {
	userID, err := ctrl.getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户认证"})
		return
	}

	var user models.User
	if err := ctrl.Config.DB.First(&user, userID).Error; err != nil {
		log.Printf("用户未找到: userID=%d, error=%v", userID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
		return
	}

	user.APIKey = ""
	user.SecretKey = ""
	if err := ctrl.Config.DB.Save(&user).Error; err != nil {
		log.Printf("删除 API 密钥失败，用户 %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除 API 密钥失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API 密钥删除成功"})
}

// maskAPIKey 对API密钥进行掩码处理
func maskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 12 {
		return strings.Repeat("*", len(key))
	}
	return key[:6] + strings.Repeat("*", len(key)-12) + key[len(key)-6:]
}
