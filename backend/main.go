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
	// 迁移双币投资相关表
	if err := models.MigrateDualInvestmentTables(cfg.DB); err != nil {
		log.Fatalf("双币投资表迁移失败: %v", err)
	}
	// 初始化默认管理员账号
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("密码哈希失败: %v", err)
	}
	adminUser := models.User{
		Username: "admin",
		Password: string(hashedPassword),
		Role:     "admin",
		Status:   "active",
	}
	if err := cfg.DB.FirstOrCreate(&adminUser, models.User{Username: "admin"}).Error; err != nil {
		log.Fatalf("创建管理员用户失败: %v", err)
	}
	log.Printf("默认管理员账号: admin / admin123")

	// 初始化测试用户（默认需要审核）
	testPassword, err := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("密码哈希失败: %v", err)
	}
	testUser := models.User{
		Username: "testuser",
		Password: string(testPassword),
		Role:     "user",
		Status:   "active", // 为了兼容现有代码，测试用户默认激活
	}
	if err := cfg.DB.FirstOrCreate(&testUser, models.User{Username: "testuser"}).Error; err != nil {
		log.Fatalf("创建测试用户失败: %v", err)
	}

	// 设置路由
	router := gin.Default()
	routes.SetupRoutes(router, cfg)

	// 启动后台任务，传递 config
	go tasks.StartPriceMonitoring(cfg)
	go tasks.CheckOrders(cfg)
	go tasks.CheckWithdrawals(cfg)
	go tasks.StartDualInvestmentTasks(cfg) // 新增：启动双币投资任务

	// 启动服务器
	log.Printf("服务器启动在端口 8081")
	log.Fatal(router.Run(":8081"))
}
