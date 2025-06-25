package routes

import (
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/controllers"
	"github.com/ccj241/binance/handlers"
	"github.com/ccj241/binance/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"time"
)

// SetupRoutes 配置路由
func SetupRoutes(router *gin.Engine, cfg *config.Config) {
	// 配置CORS中间件
	corsConfig := cors.Config{
		AllowOrigins:     []string{"http://localhost:8080", "http://127.0.0.1:8080", "*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(corsConfig))

	// 创建用户控制器实例
	userController := &controllers.UserController{Config: cfg}

	// 公共路由，无需认证
	router.POST("/register", gin.WrapH(handlers.RegisterHandler(cfg))) // 注册新用户
	router.POST("/login", gin.WrapH(handlers.LoginHandler(cfg)))       // 用户登录

	// 受保护路由，需要认证
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		// API密钥管理 - 修正路由路径以匹配前端
		protected.POST("/api-key", userController.SetAPIKey)             // 设置API密钥
		protected.GET("/api-key", userController.GetAPIKey)              // 获取API密钥
		protected.DELETE("/api-key/delete", userController.DeleteAPIKey) // 删除API密钥

		// 订单管理 - 转换为Gin handlers
		protected.GET("/orders", handlers.GinOrdersHandler(cfg))
		protected.GET("/cancelled_orders", handlers.GinCancelledOrdersHandler(cfg))
		protected.POST("/order", handlers.GinCreateOrderHandler(cfg))
		protected.POST("/cancel_order/:orderId", handlers.GinCancelOrderHandler(cfg))

		// 策略管理
		protected.POST("/strategy", handlers.GinCreateStrategyHandler(cfg))
		protected.GET("/strategies", handlers.GinListStrategiesHandler(cfg))
		protected.POST("/toggle_strategy", handlers.GinToggleStrategyHandler(cfg))
		protected.POST("/delete_strategy", handlers.GinDeleteStrategyHandler(cfg))
		protected.DELETE("/delete_strategy", handlers.GinDeleteStrategyHandler(cfg))

		// 交易对和价格
		protected.GET("/symbols", handlers.GinListSymbolsHandler(cfg))
		protected.POST("/symbols", handlers.GinAddSymbolHandler(cfg))
		protected.GET("/prices", handlers.GinPricesHandler(cfg))

		// 账户信息
		protected.GET("/balance", handlers.GinBalanceHandler(cfg))
		protected.GET("/trades", handlers.GinTradesHandler(cfg))

		// 提币管理
		protected.GET("/withdrawalhistory", handlers.GinWithdrawalHistoryHandler(cfg))
		protected.POST("/withdrawals", handlers.GinCreateWithdrawalRuleHandler(cfg))
		protected.GET("/withdrawals", handlers.GinListWithdrawalRulesHandler(cfg))
		protected.PUT("/withdrawals/:id", handlers.GinUpdateWithdrawalRuleHandler(cfg))
		protected.DELETE("/withdrawals/:id", handlers.GinDeleteWithdrawalRuleHandler(cfg))
	}
}
