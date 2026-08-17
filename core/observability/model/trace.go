package model

import "time"

// Trace 表示一次完整的链路追踪（如用户会话、Agent调用等）
type Trace struct {
	ID        int64  `gorm:"primaryKey;autoIncrement"`     // 数据库自增主键
	TraceID   string `gorm:"uniqueIndex;size:64;not null"` // 全局唯一的追踪ID
	SessionID string `gorm:"index;size:64"`                // 所属会话ID，可选

	UserInput   string `gorm:"type:text"` // 用户输入的内容
	AgentOutput string `gorm:"type:text"` // Agent输出的内容

	StartTime  time.Time `gorm:"index;not null"` // 追踪开始时间
	EndTime    time.Time // 追踪结束时间
	DurationMs int64     // 持续时间，单位ms

	SpanCount   int     // 关联的Span数量
	TotalTokens int64   // 累计token数
	TotalCost   float64 // 累计成本
	Status      string  `gorm:"size:16;index;not null;default:running"` // 追踪状态（如success、error、running）
}

// TableName 返回Trace对应的数据库表名
func (Trace) TableName() string {
	return TableTraces
}
