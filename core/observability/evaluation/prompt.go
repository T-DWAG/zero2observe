package evaluation

import (
	"fmt"
	"strings"

	"github.com/T-DWAG/zero2observe/model"
)

const systemPrompt = `你是 Agent 执行质量评估员。根据用户输入、Agent 最终输出、执行步骤，从三个维度打分。
分数范围 0.0~1.0。只输出 JSON，不要 markdown：
{
  "accuracy":   {"score": 0.0, "reason": "..."},
  "tool_usage": {"score": 0.0, "reason": "..."},
  "efficiency": {"score": 0.0, "reason": "..."}
}
维度含义：
- accuracy：输出是否满足用户需求
- tool_usage：工具是否用得合理
- efficiency：步骤/token 是否冗长`

func buildUserPrompt(tr *model.Trace, spans []*model.Span) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "## 用户输入\n%s\n\n", tr.UserInput)
	fmt.Fprintf(&sb, "## Agent 输出\n%s\n\n", tr.AgentOutput)
	fmt.Fprintf(&sb, "## 执行步骤（%d）\n", len(spans))
	for i, sp := range spans {
		fmt.Fprintf(&sb, "%d. [%s] %s status=%s tokens=%d tool=%s\n",
			i+1, sp.SpanType, sp.SpanName, sp.Status, sp.TotalTokens, sp.ToolName)
	}
	fmt.Fprintf(&sb, "\n## Trace 汇总\ntokens=%d cost=%.4f duration_ms=%d status=%s\n",
		tr.TotalTokens, tr.TotalCost, tr.DurationMs, tr.Status)
	return sb.String()
}
