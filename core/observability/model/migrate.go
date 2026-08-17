package model

import "gorm.io/gorm"

// AutoMigrate 自动迁移所有观测相关的数据表（Span、Trace、Evaluation）。
// 该方法会根据结构体定义自动创建或更新表结构，通常在应用启动时调用。
// 必须传入已初始化的 *gorm.DB 实例。
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Span{},       // 跨度表
		&Trace{},      // 链路追踪表
		&Evaluation{}, // 评估表
	)
}
