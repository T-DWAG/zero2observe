package model

import "time"

// Evaluation 表示一次对链路追踪（Trace）或其结果的评估/评价。
// 通常用于记录模型输出或Agent执行在某个维度（如准确性、效率等）的评分和理由。
type Evaluation struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"` // 数据库自增主键
	TraceID   string    `gorm:"index;size:64;not null"`   // 被评估的Trace的ID
	Dimension string    `gorm:"size:32;not null"`         // 评估的维度（如accuracy、efficiency等）
	Score     float64   // 得分（可为百分制或其他浮点评分）
	Reason    string    `gorm:"type:text"`      // 评分理由及评语
	CreatedAt time.Time `gorm:"index;not null"` // 评估的创建时间
}

// TableName 返回Evaluation对应的数据库表名
func (Evaluation) TableName() string {
	return TableEvaluations
}
