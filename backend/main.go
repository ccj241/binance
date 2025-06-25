package main

import (
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
	"github.com/ccj241/binance/routes"
	"github.com/ccj241/binance/tasks"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"log"
)

func main() {
	cfg := config.NewConfig()

	// 数据库迁移
	if err := models.MigrateDB(cfg.DB); err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 初始化测试用户
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("密码哈希失败: %v", err)
	}
	user := models.User{Username: "testuser", Password: string(hashedPassword)}
	if err := cfg.DB.FirstOrCreate(&user, models.User{Username: "testuser"}).Error; err != nil {
		log.Fatalf("创建测试用户失败: %v", err)
	}

	// 设置路由
	router := gin.Default()
	routes.SetupRoutes(router, cfg)

	// 启动后台任务，传递 config
	go tasks.StartPriceMonitoring(cfg)
	go tasks.CheckOrders(cfg)
	go tasks.CheckWithdrawals(cfg)

	// 启动服务器
	log.Printf("服务器启动在端口 8081")
	log.Fatal(router.Run(":8081"))
}
