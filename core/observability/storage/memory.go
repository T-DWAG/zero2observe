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
