package collector

import (
	"context"

	obsmodel "github.com/T-DWAG/zero2observe/model"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	ucb "github.com/cloudwego/eino/utils/callbacks"
)

type spanIDKey struct{}

func withSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, spanIDKey{}, spanID)
}

func spanIDFromCtx(ctx context.Context) string {
	id, _ := ctx.Value(spanIDKey{}).(string)
	return id
}

// NewHandler 构建 Eino Callback Handler。
func NewHandler(state *State) callbacks.Handler {
	return ucb.NewHandlerHelper().
		Agent(&ucb.AgentCallbackHandler{
			OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *adk.AgentCallbackInput) context.Context {
				sp := state.startSpan(obsmodel.SpanTypeAgent, info.Name)
				if input != nil && input.Input != nil && len(input.Input.Messages) > 0 {
					sp.Reasoning = input.Input.Messages[len(input.Input.Messages)-1].Content
				}
				return withSpanID(ctx, sp.SpanID)
			},
			OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *adk.AgentCallbackOutput) context.Context {
				state.finishSpan(spanIDFromCtx(ctx), func(sp *obsmodel.Span) {
					sp.Status = obsmodel.SpanStatusSuccess
				})
				if output != nil && output.Events != nil {
					go drainAgentEvents(output.Events)
				}
				return ctx
			},
		}).
		ChatModel(&ucb.ModelCallbackHandler{
			OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *model.CallbackInput) context.Context {
				sp := state.startSpan(obsmodel.SpanTypeLLM, info.Name)
				sp.ModelName = info.Type
				if input != nil && input.Config != nil {
					sp.ModelName = input.Config.Model
				}
				return withSpanID(ctx, sp.SpanID)
			},
			OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *model.CallbackOutput) context.Context {
				state.finishSpan(spanIDFromCtx(ctx), func(sp *obsmodel.Span) {
					sp.Status = obsmodel.SpanStatusSuccess
					if output == nil {
						return
					}
					if output.Config != nil && output.Config.Model != "" {
						sp.ModelName = output.Config.Model
					}
					if output.TokenUsage != nil {
						sp.PromptTokens = int64(output.TokenUsage.PromptTokens)
						sp.CompletionTokens = int64(output.TokenUsage.CompletionTokens)
						sp.TotalTokens = int64(output.TokenUsage.TotalTokens)
						sp.Cost = EstimateCost(sp.ModelName, sp.PromptTokens, sp.CompletionTokens)
						state.addLLMTokens(sp.PromptTokens, sp.CompletionTokens, sp.TotalTokens, sp.Cost)
					}
				})
				return ctx
			},
			OnError: func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
				state.failSpan(spanIDFromCtx(ctx), err)
				return ctx
			},
		}).
		Tool(&ucb.ToolCallbackHandler{
			OnStart: func(ctx context.Context, info *callbacks.RunInfo, input *tool.CallbackInput) context.Context {
				sp := state.startSpan(obsmodel.SpanTypeTool, info.Name)
				sp.ToolName = info.Name
				if input != nil {
					sp.ToolInput = input.ArgumentsInJSON
				}
				return withSpanID(ctx, sp.SpanID)
			},
			OnEnd: func(ctx context.Context, info *callbacks.RunInfo, output *tool.CallbackOutput) context.Context {
				state.finishSpan(spanIDFromCtx(ctx), func(sp *obsmodel.Span) {
					sp.Status = obsmodel.SpanStatusSuccess
					if output != nil {
						if output.Response != "" {
							sp.ToolOutput = output.Response
						} else if output.ToolOutput != nil {
							sp.ToolOutput = toJSON(output.ToolOutput)
						}
					}
				})
				return ctx
			},
			OnError: func(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
				state.failSpan(spanIDFromCtx(ctx), err)
				return ctx
			},
		}).
		Handler()
}

func drainAgentEvents(events *adk.AsyncIterator[*adk.AgentEvent]) {
	for {
		if _, ok := events.Next(); !ok {
			break
		}
	}
}
