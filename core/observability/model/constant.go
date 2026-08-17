// 常量表
package model

// 跨度类型
const (
	SpanTypeLLM   = "llm"
	SpanTypeTool  = "tool"
	SpanTypeAgent = "agent"
)

// 跨度状态
const (
	SpanStatusSuccess = "success"
	SpanStatusError   = "error"
	SpanStatusPending = "running" // 进行中；Step 2 采集器会用到此常量
)

// 评估维度
const (
	EvalDimensionAccuracy   = "accuracy"
	EvalDimensionToolUsage  = "tool_usage"
	EvalDimensionEfficiency = "efficiency"
)

// 表名
const (
	TableSpans       = "obs_spans"
	TableTraces      = "obs_traces"
	TableEvaluations = "obs_evaluations"
)
