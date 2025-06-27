package config

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
	"time"
)

type Config struct {
	DB        *gorm.DB
	JWTSecret string
}

func NewConfig() *Config {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "root:123456@tcp(mysql:3306)/binance?charset=utf8mb4&parseTime=True&loc=Local"
		log.Println("未设置 DATABASE_DSN，使用默认值")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败: ", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("获取数据库实例失败: ", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// 从环境变量加载 JWT 密钥
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Println("警告：未设置 JWT_SECRET，使用默认值（生产环境中应设置此环境变量）")
		jwtSecret = "your_jwt_secret_key_change_in_production" // 生产环境中应避免硬编码
	}

	return &Config{
		DB:        db,
		JWTSecret: jwtSecret,
	}
}
