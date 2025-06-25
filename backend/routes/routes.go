package routes

import (
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/controllers"
	"github.com/ccj241/binance/handlers"
	"github.com/ccj241/binance/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRoutes 配置路由
func SetupRoutes(router *gin.Engine, cfg *config.Config) {
	// 添加CORS中间件
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 创建用户控制器实例
	userController := &controllers.UserController{Config: cfg}

	// 公共路由，无需认证
	router.POST("/register", gin.WrapH(handlers.RegisterHandler(cfg))) // 注册新用户
	router.POST("/login", gin.WrapH(handlers.LoginHandler(cfg)))       // 用户登录

	// 受保护路由，需要认证
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware(cfg))
	{
		// API密钥管理
		protected.POST("/set_api_key", userController.SetAPIKey)
		protected.GET("/get_api_key", userController.GetAPIKey)
		protected.DELETE("/delete_api_key", userController.DeleteAPIKey)

		// 订单管理
		protected.GET("/orders", gin.WrapH(handlers.OrdersHandler(cfg)))
		protected.GET("/cancelled_orders", gin.WrapH(handlers.CancelledOrdersHandler(cfg)))
		protected.POST("/create_order", gin.WrapH(handlers.CreateOrderHandler(cfg)))
		protected.POST("/cancel_order/:orderId", gin.WrapH(handlers.CancelOrderHandler(cfg)))

		// 策略管理
		protected.POST("/create_strategy", gin.WrapH(handlers.CreateStrategyHandler(cfg)))
		protected.GET("/strategies", gin.WrapH(handlers.ListStrategiesHandler(cfg)))
		protected.POST("/toggle_strategy", gin.WrapH(handlers.ToggleStrategyHandler(cfg)))
		protected.POST("/delete_strategy", gin.WrapH(handlers.DeleteStrategyHandler(cfg)))
		protected.DELETE("/delete_strategy", gin.WrapH(handlers.DeleteStrategyHandler(cfg)))

		// 交易对和价格
		protected.GET("/symbols", gin.WrapH(handlers.ListSymbolsHandler(cfg)))
		protected.POST("/add_symbol", gin.WrapH(handlers.AddSymbolHandler(cfg)))
		protected.GET("/prices", gin.WrapH(handlers.PricesHandler(cfg)))

		// 账户信息
		protected.GET("/balance", gin.WrapH(handlers.BalanceHandler(cfg)))
		protected.GET("/trades", gin.WrapH(handlers.TradesHandler(cfg)))

		// 提币管理
		protected.GET("/withdrawal_history", gin.WrapH(handlers.WithdrawalHistoryHandler(cfg)))
		protected.POST("/create_withdrawal_rule", gin.WrapH(handlers.CreateWithdrawalRuleHandler(cfg)))
		protected.GET("/withdrawal_rules", gin.WrapH(handlers.ListWithdrawalRulesHandler(cfg)))
		protected.PUT("/withdrawal_rules/:id", gin.WrapH(handlers.UpdateWithdrawalRuleHandler(cfg)))
		protected.DELETE("/withdrawal_rules/:id", gin.WrapH(handlers.DeleteWithdrawalRuleHandler(cfg)))
	}
}
