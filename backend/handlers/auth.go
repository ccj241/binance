package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"log"
)

type AuthRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token   string `json:"token,omitempty"`
	Message string `json:"message"`
	UserID  uint   `json:"user_id,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// RegisterHandler 用户注册处理器
func RegisterHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeErrorResponse(w, http.StatusMethodNotAllowed, "方法不允许")
			return
		}

		var req AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("注册请求解码失败: %v", err)
			writeErrorResponse(w, http.StatusBadRequest, "无效的请求格式")
			return
		}

		// 验证输入
		if req.Username == "" || req.Password == "" {
			writeErrorResponse(w, http.StatusBadRequest, "用户名和密码不能为空")
			return
		}

		if len(req.Username) < 3 || len(req.Username) > 50 {
			writeErrorResponse(w, http.StatusBadRequest, "用户名长度必须在3-50个字符之间")
			return
		}

		if len(req.Password) < 6 {
			writeErrorResponse(w, http.StatusBadRequest, "密码长度至少6个字符")
			return
		}

		// 检查用户名是否已存在
		var existingUser models.User
		if err := cfg.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
			writeErrorResponse(w, http.StatusConflict, "用户名已存在")
			return
		}

		// 哈希密码
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("密码哈希失败: %v", err)
			writeErrorResponse(w, http.StatusInternalServerError, "密码处理失败")
			return
		}

		// 创建用户
		user := models.User{
			Username: req.Username,
			Password: string(hashedPassword),
		}

		if err := cfg.DB.Create(&user).Error; err != nil {
			log.Printf("创建用户失败: %v", err)
			writeErrorResponse(w, http.StatusInternalServerError, "用户创建失败")
			return
		}

		log.Printf("用户注册成功: %s (ID: %d)", user.Username, user.ID)
		writeSuccessResponse(w, AuthResponse{
			Message: "注册成功",
			UserID:  user.ID,
		})
	}
}

// LoginHandler 用户登录处理器
func LoginHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeErrorResponse(w, http.StatusMethodNotAllowed, "方法不允许")
			return
		}

		var req AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("登录请求解码失败: %v", err)
			writeErrorResponse(w, http.StatusBadRequest, "无效的请求格式")
			return
		}

		// 验证输入
		if req.Username == "" || req.Password == "" {
			writeErrorResponse(w, http.StatusBadRequest, "用户名和密码不能为空")
			return
		}

		// 查找用户
		var user models.User
		if err := cfg.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
			log.Printf("用户登录失败 - 用户不存在: %s", req.Username)
			writeErrorResponse(w, http.StatusUnauthorized, "用户名或密码错误")
			return
		}

		// 验证密码
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			log.Printf("用户登录失败 - 密码错误: %s", req.Username)
			writeErrorResponse(w, http.StatusUnauthorized, "用户名或密码错误")
			return
		}

		// 生成JWT令牌
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":  float64(user.ID),
			"username": user.Username,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
			"iat":      time.Now().Unix(),
		})

		tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
		if err != nil {
			log.Printf("JWT生成失败: %v", err)
			writeErrorResponse(w, http.StatusInternalServerError, "令牌生成失败")
			return
		}

		log.Printf("用户登录成功: %s (ID: %d)", user.Username, user.ID)
		writeSuccessResponse(w, AuthResponse{
			Token:   tokenString,
			Message: "登录成功",
			UserID:  user.ID,
		})
	}
}

// writeErrorResponse 写入错误响应
func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := ErrorResponse{
		Error:   "API_ERROR",
		Message: message,
	}
	json.NewEncoder(w).Encode(response)
}

// writeSuccessResponse 写入成功响应
func writeSuccessResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}
