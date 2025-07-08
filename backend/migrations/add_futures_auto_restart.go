package migrations

import (
	"gorm.io/gorm"
	"log"
)

// AddFuturesAutoRestart 添加期货策略自动重启字段
func AddFuturesAutoRestart(db *gorm.DB) error {
	type FuturesStrategy struct {
		AutoRestart bool `gorm:"default:false;comment:完成后自动重启" json:"autoRestart"`
	}

	// 检查并添加字段
	if !db.Migrator().HasColumn(&FuturesStrategy{}, "auto_restart") {
		if err := db.Exec("ALTER TABLE futures_strategies ADD COLUMN auto_restart BOOLEAN DEFAULT FALSE COMMENT '完成后自动重启'").Error; err != nil {
			log.Printf("添加 auto_restart 字段失败: %v", err)
			return err
		}
		log.Println("成功添加 auto_restart 字段")
	}

	return nil
}
