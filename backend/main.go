package main

import (
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/migrations"
	"github.com/ccj241/binance/models"
	"github.com/ccj241/binance/routes"
	"github.com/ccj241/binance/tasks"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	cfg := config.NewConfig()

	// 数据库迁移
	if err := models.MigrateDB(cfg.DB); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 迁移双币投资相关表
	if err := models.MigrateDualInvestmentTables(cfg.DB); err != nil {
		log.Fatalf("双币投资表迁移失败: %v", err)
	}
	// 添加性能优化索引
	if err := migrations.AddPerformanceIndexes(cfg.DB); err != nil {
		log.Printf("添加性能索引失败: %v", err)
		// 不要因为索引失败而退出，可能索引已存在
	}

	// 设置路由
	router := gin.Default()
	routes.SetupRoutes(router, cfg)

	// 启动后台任务
	go tasks.StartPriceMonitoring(cfg)
	go tasks.CheckOrders(cfg)
	go tasks.CheckWithdrawals(cfg)
	go tasks.StartDualInvestmentTasks(cfg) // 双币投资任务

	// 启动服务器
	log.Printf("服务器启动在端口 8081")
	log.Fatal(router.Run(":23337")) //23337

}
