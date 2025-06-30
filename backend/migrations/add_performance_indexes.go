package migrations

import (
	"gorm.io/gorm"
)

// AddPerformanceIndexes 添加性能优化索引
func AddPerformanceIndexes(db *gorm.DB) error {
	// 为 strategies 表添加复合索引
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_strategies_user_symbol_status 
		ON strategies(user_id, symbol, status, enabled, deleted_at, pending_batch)
	`).Error; err != nil {
		return err
	}

	// 为 users 表优化主键查询
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_users_id_deleted 
		ON users(id, deleted_at)
	`).Error; err != nil {
		return err
	}

	// 为 custom_symbols 表添加索引
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_custom_symbols_user_deleted 
		ON custom_symbols(user_id, deleted_at)
	`).Error; err != nil {
		return err
	}

	// 为 dual_investment_strategies 表添加索引
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_dual_strategies_enabled_status 
		ON dual_investment_strategies(enabled, status, deleted_at, next_check_time)
	`).Error; err != nil {
		return err
	}

	// 为 orders 表添加索引
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_orders_status_deleted 
		ON orders(status, deleted_at)
	`).Error; err != nil {
		return err
	}

	return nil
}
