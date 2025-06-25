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
// backend/routes/routes.go - 添加删除路由

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

	// 添加请求日志中间件
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/health"}, // 跳过健康检查日志
	}))

	// 添加错误恢复中间件
	router.Use(gin.Recovery())

	// 创建用户控制器实例
	userController := &controllers.UserController{Config: cfg}

	// 健康检查端点
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// 公共路由，无需认证
	router.POST("/register", gin.WrapH(handlers.RegisterHandler(cfg)))
	router.POST("/login", gin.WrapH(handlers.LoginHandler(cfg)))

	// 受保护路由，需要认证
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg))
	protected.Use(middleware.ValidationMiddleware()) // 添加验证中间件
	{
		// API密钥管理
		protected.POST("/api-key", userController.SetAPIKey)
		protected.GET("/api-key", userController.GetAPIKey)
		protected.DELETE("/api-key/delete", userController.DeleteAPIKey)

		// 订单管理
		protected.GET("/orders", handlers.GinOrdersHandler(cfg))
		protected.GET("/cancelled_orders", handlers.GinCancelledOrdersHandler(cfg))
		protected.POST("/order", handlers.GinCreateOrderHandler(cfg))
		protected.POST("/cancel_order/:orderId", handlers.GinCancelOrderHandler(cfg))
		protected.POST("/batch_cancel_orders", handlers.GinBatchCancelOrdersHandler(cfg)) // 批量取消订单

		// 策略管理
		protected.POST("/strategy", handlers.GinCreateStrategyHandler(cfg))
		protected.GET("/strategies", handlers.GinListStrategiesHandler(cfg))
		protected.POST("/toggle_strategy", handlers.GinToggleStrategyHandler(cfg))
		protected.POST("/delete_strategy", handlers.GinDeleteStrategyHandler(cfg))
		protected.DELETE("/delete_strategy", handlers.GinDeleteStrategyHandler(cfg))
		protected.GET("/strategy/:id/stats", handlers.GinStrategyStatsHandler(cfg))   // 策略统计
		protected.GET("/strategy/:id/orders", handlers.GinStrategyOrdersHandler(cfg)) // 策略订单

		// 交易对和价格
		protected.GET("/symbols", handlers.GinListSymbolsHandler(cfg))
		protected.POST("/symbols", handlers.GinAddSymbolHandler(cfg))
		protected.DELETE("/symbols", handlers.GinDeleteSymbolHandler(cfg)) // 删除交易对
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

	// 404 处理
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"error":   "NOT_FOUND",
			"message": "请求的资源不存在",
			"path":    c.Request.URL.Path,
		})
	})
}
