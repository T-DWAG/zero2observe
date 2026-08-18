package metrics

import "time"

// Snapshot 表示某一时间窗口内的可观测性指标快照。
type Snapshot struct {
	Scope         string     `json:"scope"`           // 统计范围（如服务、Agent 等）
	From          time.Time  `json:"from"`            // 窗口起始时间
	To            time.Time  `json:"to"`              // 窗口结束时间
	TotalTraces   int64      `json:"total_traces"`    // 窗口内总 Trace 数
	TotalTokens   int64      `json:"total_tokens"`    // 窗口内总 Token 消耗
	TotalCost     float64    `json:"total_cost"`      // 窗口内总成本
	AvgDurationMs float64    `json:"avg_duration_ms"` // 平均耗时（毫秒）
	SuccessRate   float64    `json:"success_rate"`    // 成功率，取值范围 0~1
	TopTools      []ToolStat `json:"top_tools"`       // 调用次数靠前的工具统计
}

type ToolStat struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// 支持的 scope 常量
const (
	ScopeLast24h = "last_24h"
	ScopeLast7d  = "last_7d"
	ScopeLast30d = "last_30d"
)
