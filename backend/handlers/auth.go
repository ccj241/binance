package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
	"github.com/golang-jwt/jwt/v5"
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
	Role    string `json:"role,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// JWT Claims 结构体
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// RegisterHandler 用户注册处理器
func RegisterHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 设置CORS头
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

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

		// 创建用户，默认状态为pending，需要管理员审核
		user := models.User{
			Username: req.Username,
			Password: string(hashedPassword),
			Role:     "user",
			Status:   "pending", // 默认待审核状态
		}

		if err := cfg.DB.Create(&user).Error; err != nil {
			log.Printf("创建用户失败: %v", err)
			writeErrorResponse(w, http.StatusInternalServerError, "用户创建失败")
			return
		}

		log.Printf("用户注册成功: %s (ID: %d), 等待管理员审核", user.Username, user.ID)
		writeSuccessResponse(w, AuthResponse{
			Message: "注册成功，请等待管理员审核后方可登录",
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

		// 检查用户状态
		if user.Status == "pending" {
			log.Printf("用户登录失败 - 账号待审核: %s", req.Username)
			writeErrorResponse(w, http.StatusForbidden, "账号待审核，请联系管理员")
			return
		}

		if user.Status == "disabled" {
			log.Printf("用户登录失败 - 账号已禁用: %s", req.Username)
			writeErrorResponse(w, http.StatusForbidden, "账号已被禁用，请联系管理员")
			return
		}

		// 验证密码
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
			log.Printf("用户登录失败 - 密码错误: %s", req.Username)
			writeErrorResponse(w, http.StatusUnauthorized, "用户名或密码错误")
			return
		}

		// 生成JWT令牌
		claims := Claims{
			UserID:   user.ID,
			Username: user.Username,
			Role:     user.Role,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
		if err != nil {
			log.Printf("JWT生成失败: %v", err)
			writeErrorResponse(w, http.StatusInternalServerError, "令牌生成失败")
			return
		}

		log.Printf("用户登录成功: %s (ID: %d, Role: %s)", user.Username, user.ID, user.Role)
		writeSuccessResponse(w, AuthResponse{
			Token:   tokenString,
			Message: "登录成功",
			UserID:  user.ID,
			Role:    user.Role,
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
	// 直接编码数据，不要包装
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("编码响应失败: %v", err)
	}
}
