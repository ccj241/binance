package migrations

import (
	"gorm.io/gorm"
	"log"
)

// AddFuturesIcebergFields 添加期货冰山策略相关字段
func AddFuturesIcebergFields(db *gorm.DB) error {
	type FuturesStrategy struct {
		StrategyType      string `gorm:"type:varchar(20);default:'simple'" json:"strategyType"`
		IcebergLevels     int    `gorm:"default:5;comment:冰山层数" json:"icebergLevels"`
		IcebergQuantities string `gorm:"type:text;comment:冰山策略各层数量比例" json:"icebergQuantities"`
		IcebergPriceGaps  string `gorm:"type:text;comment:冰山策略各层价格间隔(‰)" json:"icebergPriceGaps"`
	}

	// 检查并添加字段
	if !db.Migrator().HasColumn(&FuturesStrategy{}, "strategy_type") {
		if err := db.Exec("ALTER TABLE futures_strategies ADD COLUMN strategy_type VARCHAR(20) DEFAULT 'simple' COMMENT '策略类型'").Error; err != nil {
			log.Printf("添加 strategy_type 字段失败: %v", err)
			return err
		}
		log.Println("成功添加 strategy_type 字段")
	}

	if !db.Migrator().HasColumn(&FuturesStrategy{}, "iceberg_levels") {
		if err := db.Exec("ALTER TABLE futures_strategies ADD COLUMN iceberg_levels INT DEFAULT 5 COMMENT '冰山层数'").Error; err != nil {
			log.Printf("添加 iceberg_levels 字段失败: %v", err)
			return err
		}
		log.Println("成功添加 iceberg_levels 字段")
	}

	if !db.Migrator().HasColumn(&FuturesStrategy{}, "iceberg_quantities") {
		if err := db.Exec("ALTER TABLE futures_strategies ADD COLUMN iceberg_quantities TEXT COMMENT '冰山策略各层数量比例'").Error; err != nil {
			log.Printf("添加 iceberg_quantities 字段失败: %v", err)
			return err
		}
		log.Println("成功添加 iceberg_quantities 字段")
	}

	if !db.Migrator().HasColumn(&FuturesStrategy{}, "iceberg_price_gaps") {
		if err := db.Exec("ALTER TABLE futures_strategies ADD COLUMN iceberg_price_gaps TEXT COMMENT '冰山策略各层价格间隔(‰)'").Error; err != nil {
			log.Printf("添加 iceberg_price_gaps 字段失败: %v", err)
			return err
		}
		log.Println("成功添加 iceberg_price_gaps 字段")
	}

	// 更新现有记录的默认值
	if err := db.Exec("UPDATE futures_strategies SET strategy_type = 'simple' WHERE strategy_type IS NULL OR strategy_type = ''").Error; err != nil {
		log.Printf("更新现有记录的 strategy_type 默认值失败: %v", err)
	}

	log.Println("期货冰山策略字段迁移完成")
	return nil
}

// RemoveFuturesIcebergFields 回滚：移除期货冰山策略相关字段
func RemoveFuturesIcebergFields(db *gorm.DB) error {
	// 删除字段
	if db.Migrator().HasColumn("futures_strategies", "strategy_type") {
		if err := db.Exec("ALTER TABLE futures_strategies DROP COLUMN strategy_type").Error; err != nil {
			log.Printf("删除 strategy_type 字段失败: %v", err)
			return err
		}
	}

	if db.Migrator().HasColumn("futures_strategies", "iceberg_levels") {
		if err := db.Exec("ALTER TABLE futures_strategies DROP COLUMN iceberg_levels").Error; err != nil {
			log.Printf("删除 iceberg_levels 字段失败: %v", err)
			return err
		}
	}

	if db.Migrator().HasColumn("futures_strategies", "iceberg_quantities") {
		if err := db.Exec("ALTER TABLE futures_strategies DROP COLUMN iceberg_quantities").Error; err != nil {
			log.Printf("删除 iceberg_quantities 字段失败: %v", err)
			return err
		}
	}

	if db.Migrator().HasColumn("futures_strategies", "iceberg_price_gaps") {
		if err := db.Exec("ALTER TABLE futures_strategies DROP COLUMN iceberg_price_gaps").Error; err != nil {
			log.Printf("删除 iceberg_price_gaps 字段失败: %v", err)
			return err
		}
	}

	log.Println("期货冰山策略字段回滚完成")
	return nil
}
