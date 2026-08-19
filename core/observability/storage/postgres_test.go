package storage

import (
	"context"
	"testing"
	"time"

	"github.com/T-DWAG/zero2observe/model"
	"github.com/google/uuid"
)

func TestPostgresStorage_SaveAndGet(t *testing.T) {
	db, err := OpenPostgres(PostgresDSN())
	if err != nil {
		t.Fatal(err)
	}
	if err := model.AutoMigrate(db); err != nil {
		t.Fatal(err)
	}

	store := NewPostgresStorage(db)
	ctx := context.Background()
	traceID := uuid.New().String()
	now := time.Now()

	tr := &model.Trace{
		TraceID:   traceID,
		SessionID: "s-pg-test",
		UserInput: "hello",
		StartTime: now,
		Status:    model.SpanStatusSuccess,
		SpanCount: 1,
	}
	sp := &model.Span{
		SpanID:    uuid.New().String(),
		TraceID:   traceID,
		SpanType:  model.SpanTypeAgent,
		SpanName:  "agent",
		StartTime: now,
		Status:    model.SpanStatusSuccess,
	}

	if err := store.SaveSpan(ctx, sp); err != nil {
		t.Fatal(err)
	}
	if err := store.SaveTrace(ctx, tr); err != nil {
		t.Fatal(err)
	}

	got, err := store.GetTrace(ctx, traceID)
	if err != nil {
		t.Fatal(err)
	}
	if got.UserInput != "hello" {
		t.Fatalf("user_input = %q", got.UserInput)
	}

	spans, err := store.GetTraceSpans(ctx, traceID)
	if err != nil {
		t.Fatal(err)
	}
	if len(spans) != 1 {
		t.Fatalf("spans = %d", len(spans))
	}
}
