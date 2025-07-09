package migrations

import (
	"gorm.io/gorm"
	"log"
)

// AddSlowIcebergTimeout 添加慢冰山超时时间字段
func AddSlowIcebergTimeout(db *gorm.DB) error {
	type FuturesStrategy struct {
		SlowIcebergTimeout int `gorm:"default:5;comment:慢冰山各层超时时间(分钟)" json:"slowIcebergTimeout"`
	}

	// 检查并添加字段
	if !db.Migrator().HasColumn(&FuturesStrategy{}, "slow_iceberg_timeout") {
		if err := db.Exec("ALTER TABLE futures_strategies ADD COLUMN slow_iceberg_timeout INT DEFAULT 5 COMMENT '慢冰山各层超时时间(分钟)'").Error; err != nil {
			log.Printf("添加 slow_iceberg_timeout 字段失败: %v", err)
			return err
		}
		log.Println("成功添加 slow_iceberg_timeout 字段")
	}

	return nil
}
