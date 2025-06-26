package main

import (
	"github.com/ccj241/binance/config"
	"github.com/ccj241/binance/models"
	"github.com/ccj241/binance/routes"
	"github.com/ccj241/binance/tasks"
	"github.com/ccj241/binance/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
)

// MigrateEncryptAPIKeys 加密现有的明文API密钥
func MigrateEncryptAPIKeys(db *gorm.DB) error {
	var users []models.User

	// 查询所有用户
	if err := db.Find(&users).Error; err != nil {
		return err
	}

	successCount := 0
	failCount := 0

	for _, user := range users {
		needUpdate := false

		// 检查API Key是否需要加密（长度小于100通常表示是明文）
		if user.APIKey != "" && len(user.APIKey) < 100 {
			encrypted, err := utils.Encrypt(user.APIKey)
			if err != nil {
				log.Printf("加密用户 %d 的API Key失败: %v", user.ID, err)
				failCount++
				continue
			}
			user.APIKey = encrypted
			needUpdate = true
		}

		// 检查Secret Key是否需要加密
		if user.SecretKey != "" && len(user.SecretKey) < 100 {
			encrypted, err := utils.Encrypt(user.SecretKey)
			if err != nil {
				log.Printf("加密用户 %d 的Secret Key失败: %v", user.ID, err)
				failCount++
				continue
			}
			user.SecretKey = encrypted
			needUpdate = true
		}

		// 如果需要更新，保存到数据库
		if needUpdate {
			// 使用原生SQL更新，避免触发BeforeSave钩子
			if err := db.Exec(
				"UPDATE users SET api_key = ?, secret_key = ? WHERE id = ?",
				user.APIKey, user.SecretKey, user.ID,
			).Error; err != nil {
				log.Printf("更新用户 %d 的加密密钥失败: %v", user.ID, err)
				failCount++
				continue
			}
			successCount++
			log.Printf("成功加密用户 %d 的API密钥", user.ID)
		}
	}

	log.Printf("API密钥加密迁移完成: 成功 %d 个，失败 %d 个", successCount, failCount)
	return nil
}

func main() {
	cfg := config.NewConfig()
	//执行API密钥加密迁移
	if err := MigrateEncryptAPIKeys(cfg.DB); err != nil {
		log.Printf("API密钥加密迁移失败: %v", err)
	}

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
