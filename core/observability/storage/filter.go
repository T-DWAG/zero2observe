package storage

import "time"

// TraceFilter 列表查询条件；空字段表示不筛选。
type TraceFilter struct {
	SessionID string
	Status    string
	Page      int // 从 1 开始；<=0 当 1
	Size      int // 默认 20；上限建议 100
	StartTime time.Time
	EndTime   time.Time
}

func (f TraceFilter) normalize() TraceFilter {
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.Size <= 0 {
		f.Size = 20
	}
	if f.Size > 100 {
		f.Size = 100
	}
	return f
}

func (f TraceFilter) offset() int {
	return (f.Page - 1) * f.Size
}
