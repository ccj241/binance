package main

import (
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/migrations" // 添加这行
	"github.com/ccj241/binance/models"
	"github.com/ccj241/binance/routes"
	"github.com/ccj241/binance/tasks"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"time"
)

func main() {
	// 设置Gin模式
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	cfg := config.NewConfig()

	// 数据库迁移
	if err := models.MigrateDB(cfg.DB); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 迁移双币投资相关表
	if err := models.MigrateDualInvestmentTables(cfg.DB); err != nil {
		log.Fatalf("双币投资表迁移失败: %v", err)
	}

	// 迁移永续期货相关表
	if err := models.MigrateFuturesTables(cfg.DB); err != nil {
		log.Fatalf("永续期货表迁移失败: %v", err)
	}
	if err := migrations.AddFuturesIcebergFields(cfg.DB); err != nil {
		log.Fatalf("迁移失败: %v", err)
	}
	if err := migrations.AddFuturesAutoRestart(cfg.DB); err != nil {
		log.Fatalf("添加期货自动重启字段失败: %v", err)
	}
	// 添加性能优化索引
	if err := migrations.AddPerformanceIndexes(cfg.DB); err != nil {
		log.Printf("添加性能索引时出现错误: %v", err)
		// 不要因为索引失败而退出，可能部分索引已创建成功
	}

	// 设置路由
	router := gin.New()
	router.Use(gin.Recovery())

	// 可选：添加自定义的精简日志中间件
	if os.Getenv("GIN_ACCESS_LOG") == "true" {
		router.Use(func(c *gin.Context) {
			start := time.Now()
			c.Next()
			// 只记录错误和慢请求
			if c.Writer.Status() >= 400 || time.Since(start) > 1*time.Second {
				log.Printf("%s %s %d %v", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(start))
			}
		})
	}

	routes.SetupRoutes(router, cfg)

	// 启动后台任务
	go tasks.StartPriceMonitoring(cfg)
	go tasks.CheckOrders(cfg)
	go tasks.CheckWithdrawals(cfg)
	go tasks.StartDualInvestmentTasks(cfg)
	go tasks.StartFuturesMonitoring(cfg) // 添加这行

	// 启动服务器
	log.Printf("服务器启动在端口 23337")
	log.Fatal(router.Run(":23337"))
}
