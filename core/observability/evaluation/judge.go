package evaluation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/T-DWAG/zero2observe/model"
	"github.com/T-DWAG/zero2observe/storage"
)

type Judge struct {
	store storage.Storage
	llm   ChatCompleter
}

func NewJudge(store storage.Storage, llm ChatCompleter) *Judge {
	return &Judge{
		store: store,
		llm:   llm,
	}
}

type dimScore struct {
	Score  float64 `json:"score"`
	Reason string  `json:"reason"`
}

type judgePayload struct {
	Accuracy   dimScore `json:"accuracy"`
	ToolUsage  dimScore `json:"tool_usage"`
	Efficiency dimScore `json:"efficiency"`
}

func (j *Judge) Evaluate(ctx context.Context, traceID string) ([]*model.Evaluation, error) {

	tr, err := j.store.GetTrace(ctx, traceID)
	if err != nil {
		return nil, err
	}

	spans, err := j.store.GetTraceSpans(ctx, traceID)
	if err != nil {
		return nil, fmt.Errorf("get trace spans: %w", err)
	}

	raw, err := j.llm.Complete(ctx, systemPrompt, buildUserPrompt(tr, spans))
	if err != nil {
		return nil, fmt.Errorf("llm complete: %w", err)
	}

	payload, err := parseJudgeJSON(raw)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	evals := []*model.Evaluation{
		{TraceID: traceID, Dimension: model.EvalDimensionAccuracy, Score: clamp01(payload.Accuracy.Score), Reason: payload.Accuracy.Reason, CreatedAt: now},
		{TraceID: traceID, Dimension: model.EvalDimensionToolUsage, Score: clamp01(payload.ToolUsage.Score), Reason: payload.ToolUsage.Reason, CreatedAt: now},
		{TraceID: traceID, Dimension: model.EvalDimensionEfficiency, Score: clamp01(payload.Efficiency.Score), Reason: payload.Efficiency.Reason, CreatedAt: now},
	}
	for _, e := range evals {
		if err := j.store.SaveEvaluation(ctx, e); err != nil {
			return nil, fmt.Errorf("save evaluation: %w", err)
		}
	}
	return evals, nil
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func parseJudgeJSON(raw string) (*judgePayload, error) {
	//去掉空格
	raw = strings.TrimSpace(raw)

	if i := strings.Index(raw, "{"); i >= 0 {
		if j := strings.LastIndex(raw, "}"); j >= 0 {
			raw = raw[i : j+1]
		}
	}

	var p judgePayload
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		return nil, fmt.Errorf("parse judge json: %w", err)
	}
	return &p, nil
}
