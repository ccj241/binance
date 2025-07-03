package migrations

import (
	"gorm.io/gorm"
	"log"
)

// AddPerformanceIndexes 添加性能优化索引
func AddPerformanceIndexes(db *gorm.DB) error {
	// 定义需要创建的索引
	indexes := []struct {
		table     string
		indexName string
		columns   string
	}{
		{
			table:     "strategies",
			indexName: "idx_strategies_user_symbol_status",
			columns:   "(user_id, symbol, status, enabled, deleted_at, pending_batch)",
		},
		{
			table:     "users",
			indexName: "idx_users_id_deleted",
			columns:   "(id, deleted_at)",
		},
		{
			table:     "custom_symbols",
			indexName: "idx_custom_symbols_user_deleted",
			columns:   "(user_id, deleted_at)",
		},
		{
			table:     "dual_investment_strategies",
			indexName: "idx_dual_strategies_enabled_status",
			columns:   "(enabled, status, deleted_at, next_check_time)",
		},
		{
			table:     "orders",
			indexName: "idx_orders_status_deleted",
			columns:   "(status, deleted_at)",
		},
		{
			table:     "orders",
			indexName: "idx_orders_strategy_status",
			columns:   "(strategy_id, status)",
		},
		{
			table:     "strategies",
			indexName: "idx_strategies_user_enabled",
			columns:   "(user_id, enabled, deleted_at)",
		},
		{
			table:     "dual_investment_products",
			indexName: "idx_dual_products_symbol_status",
			columns:   "(symbol, status, deleted_at)",
		},
		{
			table:     "withdrawals",
			indexName: "idx_withdrawals_user_enabled",
			columns:   "(user_id, enabled, deleted_at)",
		},
		{
			table:     "futures_strategies",
			indexName: "idx_futures_strategies_user_status",
			columns:   "(user_id, enabled, status, deleted_at)",
		},
		{
			table:     "futures_orders",
			indexName: "idx_futures_orders_user_status",
			columns:   "(user_id, status, strategy_id)",
		},
		{
			table:     "futures_positions",
			indexName: "idx_futures_positions_user_status",
			columns:   "(user_id, status, strategy_id)",
		},
	}

	// 逐个创建索引
	for _, idx := range indexes {
		if err := createIndexIfNotExists(db, idx.table, idx.indexName, idx.columns); err != nil {
			log.Printf("创建索引 %s 失败: %v", idx.indexName, err)
			// 继续创建其他索引，不要因为一个失败而停止
		} else {
			log.Printf("索引 %s 创建成功或已存在", idx.indexName)
		}
	}

	return nil
}

// createIndexIfNotExists 检查索引是否存在，不存在则创建
func createIndexIfNotExists(db *gorm.DB, table, indexName, columns string) error {
	// 检查索引是否已存在
	var count int64
	checkSQL := `
		SELECT COUNT(*) 
		FROM information_schema.statistics 
		WHERE table_schema = DATABASE() 
		AND table_name = ? 
		AND index_name = ?
	`

	if err := db.Raw(checkSQL, table, indexName).Count(&count).Error; err != nil {
		return err
	}

	// 如果索引不存在，则创建
	if count == 0 {
		createSQL := "CREATE INDEX " + indexName + " ON " + table + " " + columns
		if err := db.Exec(createSQL).Error; err != nil {
			return err
		}
	}

	return nil
}

// RemovePerformanceIndexes 移除性能优化索引（用于回滚）
func RemovePerformanceIndexes(db *gorm.DB) error {
	indexes := []struct {
		table     string
		indexName string
	}{
		{"strategies", "idx_strategies_user_symbol_status"},
		{"users", "idx_users_id_deleted"},
		{"custom_symbols", "idx_custom_symbols_user_deleted"},
		{"dual_investment_strategies", "idx_dual_strategies_enabled_status"},
		{"orders", "idx_orders_status_deleted"},
		{"orders", "idx_orders_strategy_status"},
		{"strategies", "idx_strategies_user_enabled"},
		{"dual_investment_products", "idx_dual_products_symbol_status"},
		{"withdrawals", "idx_withdrawals_user_enabled"},
	}

	for _, idx := range indexes {
		dropSQL := "DROP INDEX " + idx.indexName + " ON " + idx.table
		if err := db.Exec(dropSQL).Error; err != nil {
			log.Printf("删除索引 %s 失败: %v", idx.indexName, err)
			// 继续删除其他索引
		}
	}

	return nil
}
