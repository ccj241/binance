package controllers

import (
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
	"github.com/dgrijalva/jwt-go"
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

// SetAPIKey 保存用户的 API 密钥
func (ctrl *UserController) SetAPIKey(c *gin.Context) {
	var input APIKeyInput
	if err := c.ShouldBindJSON(&input); err != nil {
		log.Printf("无效的请求体: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, exists := c.Get("claims")
	if !exists {
		log.Printf("上下文缺少 claims")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的 token claims"})
		return
	}

	userID, ok := claims.(jwt.MapClaims)["user_id"].(float64)
	if !ok {
		log.Printf("claims 中的 user_id 无效: %v", claims)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的 user_id"})
		return
	}

	var user models.User
	if err := ctrl.Config.DB.First(&user, uint(userID)).Error; err != nil {
		log.Printf("用户未找到: userID=%d, error=%v", uint(userID), err)
		c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
		return
	}

	log.Printf("为用户 %d 保存 API 密钥: APIKey=%s, APISecret=%s", uint(userID), input.APIKey, input.APISecret)
	user.APIKey = input.APIKey
	user.SecretKey = input.APISecret // 修改：从 user.APISecret 改为 user.SecretKey
	if err := ctrl.Config.DB.Save(&user).Error; err != nil {
		log.Printf("保存 API 密钥失败，用户 %d: %v", uint(userID), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存 API 密钥失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API 密钥更新成功"})
}

// GetAPIKey 获取用户的 API 密钥（部分掩码）
func (ctrl *UserController) GetAPIKey(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		log.Printf("上下文缺少 claims")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的 token claims"})
		return
	}

	userID, ok := claims.(jwt.MapClaims)["user_id"].(float64)
	if !ok {
		log.Printf("claims 中的 user_id 无效: %v", claims)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的 user_id"})
		return
	}

	var user models.User
	if err := ctrl.Config.DB.First(&user, uint(userID)).Error; err != nil {
		log.Printf("用户未找到: userID=%d, error=%v", uint(userID), err)
		c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
		return
	}

	var maskedAPIKey, maskedSecretKey string
	if user.APIKey != "" {
		if len(user.APIKey) > 12 {
			maskedAPIKey = user.APIKey[:6] + strings.Repeat("*", len(user.APIKey)-12) + user.APIKey[len(user.APIKey)-6:]
		} else {
			maskedAPIKey = strings.Repeat("*", len(user.APIKey))
		}
	}
	if user.SecretKey != "" { // 修改：从 user.APISecret 改为 user.SecretKey
		if len(user.SecretKey) > 12 {
			maskedSecretKey = user.SecretKey[:6] + strings.Repeat("*", len(user.SecretKey)-12) + user.SecretKey[len(user.SecretKey)-6:]
		} else {
			maskedSecretKey = strings.Repeat("*", len(user.SecretKey))
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"apiKey":    maskedAPIKey,
		"secretKey": maskedSecretKey,
	})
}

// DeleteAPIKey 删除用户的 API 密钥
func (ctrl *UserController) DeleteAPIKey(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		log.Printf("上下文缺少 claims")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的 token claims"})
		return
	}

	userID, ok := claims.(jwt.MapClaims)["user_id"].(float64)
	if !ok {
		log.Printf("claims 中的 user_id 无效: %v", claims)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的 user_id"})
		return
	}

	var user models.User
	if err := ctrl.Config.DB.First(&user, uint(userID)).Error; err != nil {
		log.Printf("用户未找到: userID=%d, error=%v", uint(userID), err)
		c.JSON(http.StatusNotFound, gin.H{"error": "用户未找到"})
		return
	}

	user.APIKey = ""
	user.SecretKey = "" // 修改：从 user.APISecret 改为 user.SecretKey
	if err := ctrl.Config.DB.Save(&user).Error; err != nil {
		log.Printf("删除 API 密钥失败，用户 %d: %v", uint(userID), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除 API 密钥失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API 密钥删除成功"})
}
