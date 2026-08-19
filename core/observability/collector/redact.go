package collector

import (
	"regexp"

	"github.com/T-DWAG/zero2observe/model"
)

var phoneRE = regexp.MustCompile(`1\d{10}`)

func redactText(s string) string {
	if s == "" {
		return s
	}
	return phoneRE.ReplaceAllString(s, "1XXXXXXXXXX")
}

// applyContentPolicy 在落盘前调用：先按 NoContent 清空，再按 Redact 打码。
func applyContentPolicy(cfg Config, tr *model.Trace, sp *model.Span) {
	if cfg.NoContent {
		if tr != nil {
			tr.UserInput = ""
			tr.AgentOutput = ""
		}
		if sp != nil {
			sp.Reasoning = ""
			sp.ToolInput = ""
			sp.ToolOutput = ""
			sp.ErrorMsg = ""
		}
		return
	}
	if !cfg.Redact {
		return
	}
	if tr != nil {
		tr.UserInput = redactText(tr.UserInput)
		tr.AgentOutput = redactText(tr.AgentOutput)
	}
	if sp != nil {
		sp.Reasoning = redactText(sp.Reasoning)
		sp.ToolInput = redactText(sp.ToolInput)
		sp.ToolOutput = redactText(sp.ToolOutput)
		sp.ErrorMsg = redactText(sp.ErrorMsg)
	}
}
