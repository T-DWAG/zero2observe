package evaluation

import "context"

type ChatCompleter interface {
	Complete(ctx context.Context, systemPrompt string, userPrompt string) (string, error)
}

type FakeCompleter struct {
	Response string
	Err      error
}

func (f *FakeCompleter) Complete(ctx context.Context, systemPrompt string, userPrompt string) (string, error) {
	if f.Err != nil {
		return "", f.Err
	}

	if f.Response != "" {
		return f.Response, nil
	}

	// 默认三分都给中等分
	return `{
			"accuracy":   {"score": 0.9, "reason": "回答覆盖用户问题"},
			"tool_usage": {"score": 0.8, "reason": "工具调用合理"},
			"efficiency": {"score": 0.7, "reason": "token 消耗可接受"}
		  }`, nil
}


