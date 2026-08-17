package storage

import (
	"context"
	"sync"

	"github.com/T-DWAG/zero2observe/model"
)

type MemoryStorage struct {
	mu     sync.RWMutex
	Traces []*model.Trace
	Spans  []*model.Span
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{}
}

func (m *MemoryStorage) SaveSpan(_ context.Context, span *model.Span) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Spans = append(m.Spans, span)
	return nil
}

func (m *MemoryStorage) SaveTrace(_ context.Context, trace *model.Trace) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Traces = append(m.Traces, trace)
	return nil
}

func (m *MemoryStorage) GetTrace(_ context.Context, traceID string) (*model.Trace, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, tr := range m.Traces {
		if tr.TraceID == traceID {
			cp := *tr
			return &cp, nil
		}
	}
	return nil, ErrNotFound
}

func (m *MemoryStorage) GetTraceSpans(_ context.Context, traceID string) ([]*model.Span, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*model.Span, 0)
	for _, sp := range m.Spans {
		if sp.TraceID == traceID {
			cp := *sp
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (m *MemoryStorage) ListTraces(_ context.Context, filter TraceFilter) ([]*model.Trace, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	filter = filter.normalize()

	matched := make([]*model.Trace, 0)
	for _, tr := range m.Traces {
		if filter.SessionID != "" && tr.SessionID != filter.SessionID {
			continue
		}
		if filter.Status != "" && tr.Status != filter.Status {
			continue
		}
		if !filter.StartTime.IsZero() && tr.StartTime.Before(filter.StartTime) {
			continue
		}
		if !filter.EndTime.IsZero() && tr.StartTime.After(filter.EndTime) {
			continue
		}
		cp := *tr
		matched = append(matched, &cp)
	}

	total := int64(len(matched))
	start := filter.offset()
	if start >= len(matched) {
		return []*model.Trace{}, total, nil
	}
	end := start + filter.Size
	if end > len(matched) {
		end = len(matched)
	}
	return matched[start:end], total, nil
}
