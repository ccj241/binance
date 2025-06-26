package routes

import (
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/controllers"
	"github.com/ccj241/binance/middleware"
	"github.com/gin-gonic/gin"
)

// SetupDualInvestmentRoutes 配置双币投资相关路由
func SetupDualInvestmentRoutes(router *gin.RouterGroup, cfg *config.Config) {
	dualController := &controllers.DualInvestmentController{Config: cfg}

	// 双币投资路由组
	dualGroup := router.Group("/dual-investment")
	dualGroup.Use(middleware.AuthMiddleware(cfg))
	{
		// 产品相关
		dualGroup.GET("/products", dualController.GetProducts)         // 获取可投资产品列表
		dualGroup.POST("/simulate", dualController.SimulateInvestment) // 模拟投资计算

		// 策略管理
		dualGroup.GET("/strategies", dualController.GetStrategies)         // 获取策略列表
		dualGroup.POST("/strategies", dualController.CreateStrategy)       // 创建策略
		dualGroup.PUT("/strategies/:id", dualController.UpdateStrategy)    // 更新策略
		dualGroup.DELETE("/strategies/:id", dualController.DeleteStrategy) // 删除策略

		// 订单管理
		dualGroup.POST("/orders", dualController.CreateOrder) // 创建订单
		dualGroup.GET("/orders", dualController.GetOrders)    // 获取订单列表

		// 统计信息
		dualGroup.GET("/stats", dualController.GetStats) // 获取统计信息
	}
}
