package routes

import (
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/controllers"
	"github.com/ccj241/binance/middleware"
	"github.com/gin-gonic/gin"
)

// SetupFuturesRoutes 配置永续期货相关路由
func SetupFuturesRoutes(router *gin.RouterGroup, cfg *config.Config) {
	futuresController := &controllers.FuturesController{Config: cfg}

	// 永续期货路由组
	futuresGroup := router.Group("/futures")
	futuresGroup.Use(middleware.AuthMiddleware(cfg))
	{
		// 策略管理
		futuresGroup.GET("/strategies", futuresController.GetStrategies)         // 获取策略列表
		futuresGroup.POST("/strategies", futuresController.CreateStrategy)       // 创建策略
		futuresGroup.PUT("/strategies/:id", futuresController.UpdateStrategy)    // 更新策略
		futuresGroup.DELETE("/strategies/:id", futuresController.DeleteStrategy) // 删除策略

		// 订单管理
		futuresGroup.GET("/orders", futuresController.GetOrders) // 获取订单列表

		// 持仓管理
		futuresGroup.GET("/positions", futuresController.GetPositions) // 获取持仓列表

		// 统计信息
		futuresGroup.GET("/stats", futuresController.GetStats) // 获取统计信息
	}
}
