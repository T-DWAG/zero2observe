package collector

import "strings"

// EstimateCost 按模型名估算美元成本（每百万 token）。
func EstimateCost(modelName string, promptTokens, completionTokens int64) float64 {
	promptRate, completionRate := lookupRates(modelName)
	return float64(promptTokens)/1_000_000*promptRate +
		float64(completionTokens)/1_000_000*completionRate
}

func lookupRates(modelName string) (promptPerM, completionPerM float64) {
	name := strings.ToLower(modelName)
	switch {
	case strings.Contains(name, "gpt-4o-mini"):
		return 0.15, 0.60
	case strings.Contains(name, "gpt-4o"):
		return 2.5, 10.0
	case strings.Contains(name, "deepseek"):
		return 0.27, 1.1
	case strings.Contains(name, "claude-3.5") || strings.Contains(name, "claude-3-5"):
		return 3.0, 15.0
	default:
		return 0.15, 0.60
	}
}
