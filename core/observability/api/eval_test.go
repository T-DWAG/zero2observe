package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/T-DWAG/zero2observe/evaluation"
	"github.com/T-DWAG/zero2observe/model"
	"github.com/T-DWAG/zero2observe/storage"
)

func TestHandleCreateEvaluation(t *testing.T) {
	store := storage.NewMemoryStorage()
	ctx := context.Background()
	now := time.Now().UTC()
	_ = store.SaveTrace(ctx, &model.Trace{
		TraceID: "tr-e", UserInput: "hi", AgentOutput: "hello",
		StartTime: now, Status: model.SpanStatusSuccess,
	})

	srv := NewServer(store).WithJudge(evaluation.NewJudge(store, &evaluation.FakeCompleter{}))

	body, _ := json.Marshal(map[string]string{"trace_id": "tr-e"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluations", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("post status=%d body=%s", rec.Code, rec.Body.String())
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/evaluations/tr-e", nil)
	rec2 := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("get status=%d", rec2.Code)
	}
}
