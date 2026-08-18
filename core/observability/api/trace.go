package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/T-DWAG/zero2observe/model"
	"github.com/T-DWAG/zero2observe/storage"
)

type listTracesResponse struct {
	Items []*traceDTO `json:"items"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}

// traceDTO：对外 JSON，避免直接暴露 GORM 自增 ID（可选隐藏）。
// 教学简化：也可以直接返回 *model.Trace；这里演示一层 DTO。
type traceDTO struct {
	TraceID     string  `json:"trace_id"`
	SessionID   string  `json:"session_id"`
	UserInput   string  `json:"user_input"`
	AgentOutput string  `json:"agent_output"`
	Status      string  `json:"status"`
	SpanCount   int     `json:"span_count"`
	TotalTokens int64   `json:"total_tokens"`
	TotalCost   float64 `json:"total_cost"`
	DurationMs  int64   `json:"duration_ms"`
	StartTime   string  `json:"start_time"`
	EndTime     string  `json:"end_time,omitempty"`
}

type spanDTO struct {
	SpanID       string  `json:"span_id"`
	TraceID      string  `json:"trace_id"`
	ParentSpanID string  `json:"parent_span_id,omitempty"`
	SpanType     string  `json:"span_type"`
	SpanName     string  `json:"span_name"`
	Status       string  `json:"status"`
	DurationMs   int64   `json:"duration_ms"`
	ModelName    string  `json:"model_name,omitempty"`
	TotalTokens  int64   `json:"total_tokens,omitempty"`
	Cost         float64 `json:"cost,omitempty"`
	ToolName     string  `json:"tool_name,omitempty"`
	ErrorMsg     string  `json:"error_msg,omitempty"`
	StartTime    string  `json:"start_time"`
	EndTime      string  `json:"end_time,omitempty"`
}

func (s *Server) handleListTraces(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	size, _ := strconv.Atoi(q.Get("size"))

	filter := storage.TraceFilter{
		SessionID: q.Get("session_id"),
		Status:    q.Get("status"),
		Page:      page,
		Size:      size,
	}

	items, total, err := s.store.ListTraces(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	out := make([]*traceDTO, 0, len(items))
	for _, tr := range items {
		out = append(out, toTraceDTO(tr))
	}

	// normalize 后的 page/size 回传（与 storage 默认一致）
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 100 {
		size = 100
	}

	writeJSON(w, http.StatusOK, listTracesResponse{
		Items: out,
		Total: total,
		Page:  page,
		Size:  size,
	})
}

func (s *Server) handleGetTrace(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing trace id")
		return
	}

	tr, err := s.store.GetTrace(r.Context(), id)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "trace not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, toTraceDTO(tr))
}

func (s *Server) handleGetTraceSpans(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing trace id")
		return
	}

	// 先确认 Trace 存在，避免「空 spans」和「不存在」混淆
	if _, err := s.store.GetTrace(r.Context(), id); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "trace not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	spans, err := s.store.GetTraceSpans(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	out := make([]*spanDTO, 0, len(spans))
	for _, sp := range spans {
		out = append(out, toSpanDTO(sp))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"trace_id": id,
		"items":    out,
		"total":    len(out),
	})
}

func toTraceDTO(tr *model.Trace) *traceDTO {
	dto := &traceDTO{
		TraceID:     tr.TraceID,
		SessionID:   tr.SessionID,
		UserInput:   tr.UserInput,
		AgentOutput: tr.AgentOutput,
		Status:      tr.Status,
		SpanCount:   tr.SpanCount,
		TotalTokens: tr.TotalTokens,
		TotalCost:   tr.TotalCost,
		DurationMs:  tr.DurationMs,
		StartTime:   tr.StartTime.UTC().Format(time.RFC3339),
	}
	if !tr.EndTime.IsZero() {
		dto.EndTime = tr.EndTime.UTC().Format(time.RFC3339)
	}
	return dto
}

func toSpanDTO(sp *model.Span) *spanDTO {
	dto := &spanDTO{
		SpanID:       sp.SpanID,
		TraceID:      sp.TraceID,
		ParentSpanID: sp.ParentSpanID,
		SpanType:     sp.SpanType,
		SpanName:     sp.SpanName,
		Status:       sp.Status,
		DurationMs:   sp.DurationMs,
		ModelName:    sp.ModelName,
		TotalTokens:  sp.TotalTokens,
		Cost:         sp.Cost,
		ToolName:     sp.ToolName,
		ErrorMsg:     sp.ErrorMsg,
		StartTime:    sp.StartTime.UTC().Format(time.RFC3339),
	}
	if !sp.EndTime.IsZero() {
		dto.EndTime = sp.EndTime.UTC().Format(time.RFC3339)
	}
	return dto
}
