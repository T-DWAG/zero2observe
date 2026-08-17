package model

import "time"

// Span 表示一次观测到的操作跨度（如LLM调用、工具调用等）。
type Span struct {
	ID           int64  `gorm:"primaryKey;autoIncrement"`     // 数据库自增主键
	SpanID       string `gorm:"uniqueIndex;size:64;not null"` // 全局唯一的跨度ID
	TraceID      string `gorm:"index;size:64;not null"`       // 所属追踪ID
	ParentSpanID string `gorm:"size:64"`                      // 父跨度ID

	SpanType string `gorm:"size:32;index;not null"` // 跨度类型（如llm、tool、agent）
	SpanName string `gorm:"size:128"`               // 跨度名称

	ModelName        string  `gorm:"size:64"` // LLM/模型名称
	PromptTokens     int64   // 输入tokens数
	CompletionTokens int64   // 输出tokens数
	TotalTokens      int64   // 总tokens数
	Cost             float64 // 本次调用成本（如API计费）

	ToolName   string `gorm:"size:64"`   // 工具名称
	ToolInput  string `gorm:"type:text"` // 工具输入内容
	ToolOutput string `gorm:"type:text"` // 工具输出内容

	StartTime  time.Time `gorm:"index;not null"` // 跨度开始时间
	EndTime    time.Time // 跨度结束时间
	DurationMs int64     // 持续时间，单位ms

	Status    string `gorm:"size:16;index;not null;default:running"` // 跨度状态（如success、error、running）
	ErrorMsg  string `gorm:"type:text"`                              // 错误信息
	Reasoning string `gorm:"type:text"`                              // Reasoning/推理链
}

// TableName 返回Span对应的数据库表名
func (Span) TableName() string {
	return TableSpans
}
