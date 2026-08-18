package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/T-DWAG/zero2observe/metrics"
	"github.com/T-DWAG/zero2observe/model"
	"github.com/T-DWAG/zero2observe/storage"
)

func TestGetMetrics(t *testing.T) {
	//1、准备内存存储并写入一条窗口内的成功 Trace
	store := storage.NewMemoryStorage()
	now := time.Now().UTC()
	_ = store.SaveTrace(context.Background(), &model.Trace{
		TraceID: "tm", StartTime: now.Add(-time.Hour),
		Status: model.SpanStatusSuccess, TotalTokens: 10, DurationMs: 100,
	})

	//2、挂载聚合器并构造 GET /api/v1/metrics 请求
	srv := NewServer(store).WithAggregator(metrics.NewAggregator(store, time.Minute))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics?scope=last_24h", nil)
	rec := httptest.NewRecorder()

	//3、执行请求，期望 200
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}

	//4、解析快照并校验聚合结果
	var body metrics.Snapshot
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.TotalTraces != 1 {
		t.Fatalf("traces=%d", body.TotalTraces)
	}
}

func TestGetMetrics_BadScope(t *testing.T) {
	//1、挂载聚合器（ttl=0 走默认缓存时间）
	srv := NewServer(storage.NewMemoryStorage()).
		WithAggregator(metrics.NewAggregator(storage.NewMemoryStorage(), 0))

	//2、使用非法 scope 请求指标接口
	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics?scope=nope", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	//3、期望 400 Bad Request
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rec.Code)
	}
}
