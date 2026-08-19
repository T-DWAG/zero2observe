package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

func newChatModel(ctx context.Context) (model.ToolCallingChatModel, error) {
	//1、provider
	provider := strings.ToLower(strings.TrimSpace(envOr("LLM_PROVIDER", "openai")))

	//2、apikey
	apiKey := strings.TrimSpace(envOr("LLM_API_KEY", ""))
	if apiKey == "" {
		return nil, errors.New("LLM_API_KEY is not set")
	}
	//3、modelname
	modelName := strings.TrimSpace(envOr("LLM_MODEL", "gpt-4o-mini"))

	//3、依据provider 构建
	switch provider {
	case "openai":
		cfg := &openai.ChatModelConfig{
			APIKey: apiKey,
			Model:  modelName,
		}
		if base := strings.TrimSpace(os.Getenv("LLM_BASE_URL")); base != "" {
			cfg.BaseURL = base
		}
		return openai.NewChatModel(ctx, cfg)

	case "claude", "anthropic":
		return claude.NewChatModel(ctx, &claude.Config{
			APIKey:    apiKey,
			Model:     modelName,
			MaxTokens: 2048,
		})

	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

// env读取函数
func envOr(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func hasLLMKey() bool {
	return strings.TrimSpace(envOr("LLM_API_KEY", "")) != ""

}
