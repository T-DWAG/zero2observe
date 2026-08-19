package metrics

import (
	"context"
	"testing"
	"time"

	"github.com/T-DWAG/zero2observe/model"
	"github.com/T-DWAG/zero2observe/storage"
)

func mustPostgresStore(t *testing.T) storage.Storage {
	t.Helper()

	db, err := storage.OpenPostgres(storage.PostgresDSN())
	if err != nil {
		t.Skipf("postgres unavailable: %v", err)
	}
	if err := model.AutoMigrate(db); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("TRUNCATE obs_spans, obs_traces, obs_evaluations RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatal(err)
	}
	return storage.NewPostgresStorage(db)
}

func TestAggregate_Last24h(t *testing.T) {
	store := mustPostgresStore(t)
	ctx := context.Background()
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)

	// 窗内 2 条（1 success 1 error）
	_ = store.SaveTrace(ctx, &model.Trace{
		TraceID: "t1", StartTime: now.Add(-2 * time.Hour), EndTime: now.Add(-2*time.Hour + time.Second),
		Status: model.SpanStatusSuccess, TotalTokens: 100, TotalCost: 0.01, DurationMs: 1000,
	})
	_ = store.SaveTrace(ctx, &model.Trace{
		TraceID: "t2", StartTime: now.Add(-1 * time.Hour), EndTime: now.Add(-1*time.Hour + time.Second),
		Status: model.SpanStatusError, TotalTokens: 50, TotalCost: 0.02, DurationMs: 2000,
	})
	// 窗外
	_ = store.SaveTrace(ctx, &model.Trace{
		TraceID: "t-old", StartTime: now.Add(-48 * time.Hour),
		Status: model.SpanStatusSuccess, TotalTokens: 999, DurationMs: 1,
	})

	_ = store.SaveSpan(ctx, &model.Span{
		SpanID: "s1", TraceID: "t1", SpanType: model.SpanTypeTool, ToolName: "get_weather",
		StartTime: now.Add(-2 * time.Hour), Status: model.SpanStatusSuccess,
	})
	_ = store.SaveSpan(ctx, &model.Span{
		SpanID: "s2", TraceID: "t1", SpanType: model.SpanTypeTool, ToolName: "get_weather",
		StartTime: now.Add(-2 * time.Hour), Status: model.SpanStatusSuccess,
	})
	_ = store.SaveSpan(ctx, &model.Span{
		SpanID: "s3", TraceID: "t2", SpanType: model.SpanTypeTool, ToolName: "search",
		StartTime: now.Add(-1 * time.Hour), Status: model.SpanStatusSuccess,
	})

	agg := NewAggregator(store, time.Minute)
	snap, err := agg.Aggregate(ctx, ScopeLast24h, now)
	if err != nil {
		t.Fatal(err)
	}
	if snap.TotalTraces != 2 {
		t.Fatalf("traces=%d", snap.TotalTraces)
	}
	if snap.TotalTokens != 150 {
		t.Fatalf("tokens=%d", snap.TotalTokens)
	}
	if snap.SuccessRate != 0.5 {
		t.Fatalf("success_rate=%v", snap.SuccessRate)
	}
	if snap.AvgDurationMs != 1500 {
		t.Fatalf("avg=%v", snap.AvgDurationMs)
	}
	if len(snap.TopTools) < 1 || snap.TopTools[0].Name != "get_weather" || snap.TopTools[0].Count != 2 {
		t.Fatalf("top_tools=%+v", snap.TopTools)
	}
}

func TestAggregate_InvalidScope(t *testing.T) {
	agg := NewAggregator(storage.NewMemoryStorage(), 0)
	_, err := agg.Aggregate(context.Background(), "yesterday", time.Now())
	if err == nil {
		t.Fatal("want error")
	}
}

func TestAggregate_CacheHit(t *testing.T) {
	store := storage.NewMemoryStorage()
	ctx := context.Background()
	now := time.Now().UTC()
	_ = store.SaveTrace(ctx, &model.Trace{
		TraceID: "t1", StartTime: now.Add(-time.Hour),
		Status: model.SpanStatusSuccess, DurationMs: 10,
	})

	agg := NewAggregator(store, time.Minute)
	s1, err := agg.Aggregate(ctx, ScopeLast24h, now)
	if err != nil {
		t.Fatal(err)
	}
	// 再写一条，若走缓存则 TotalTraces 仍为 1
	_ = store.SaveTrace(ctx, &model.Trace{
		TraceID: "t2", StartTime: now.Add(-30 * time.Minute),
		Status: model.SpanStatusSuccess, DurationMs: 10,
	})
	s2, err := agg.Aggregate(ctx, ScopeLast24h, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if s1.TotalTraces != s2.TotalTraces {
		t.Fatalf("cache miss? %d vs %d", s1.TotalTraces, s2.TotalTraces)
	}
}
