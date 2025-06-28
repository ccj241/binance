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
		log.Printf("绑定API密钥请求失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求数据", "details": err.Error()})
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

	log.Printf("为用户 %d 保存 API 密钥: APIKey长度=%d, SecretKey长度=%d",
		userID, len(input.APIKey), len(input.APISecret))

	// 直接赋值，BeforeSave钩子会自动加密
	user.APIKey = input.APIKey
	user.SecretKey = input.APISecret

	// 保存前打印原始值长度
	log.Printf("保存前 - 用户 %d: APIKey=%s..., SecretKey=%s...",
		userID, maskAPIKey(input.APIKey), maskAPIKey(input.APISecret))

	if err := ctrl.Config.DB.Save(&user).Error; err != nil {
		log.Printf("保存 API 密钥失败，用户 %d: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存 API 密钥失败", "details": err.Error()})
		return
	}

	// 重新查询验证是否保存成功
	var savedUser models.User
	if err := ctrl.Config.DB.First(&savedUser, userID).Error; err != nil {
		log.Printf("重新查询用户失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "验证保存失败"})
		return
	}

	// 验证是否真的保存了
	if savedUser.APIKey == "" || savedUser.SecretKey == "" {
		log.Printf("警告：用户 %d 的API密钥保存后为空", userID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API密钥保存异常"})
		return
	}

	log.Printf("API密钥保存成功 - 用户 %d: 加密后APIKey长度=%d, SecretKey长度=%d",
		userID, len(savedUser.APIKey), len(savedUser.SecretKey))

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

	// 解密API Key
	if user.APIKey != "" {
		decryptedAPIKey, err := user.GetDecryptedAPIKey()
		if err != nil {
			log.Printf("解密API Key失败: %v", err)
			maskedAPIKey = "解密失败"
		} else {
			maskedAPIKey = maskAPIKey(decryptedAPIKey)
		}
	}

	// 解密Secret Key
	if user.SecretKey != "" {
		decryptedSecretKey, err := user.GetDecryptedSecretKey()
		if err != nil {
			log.Printf("解密Secret Key失败: %v", err)
			maskedSecretKey = "解密失败"
		} else {
			maskedSecretKey = maskAPIKey(decryptedSecretKey)
		}
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
