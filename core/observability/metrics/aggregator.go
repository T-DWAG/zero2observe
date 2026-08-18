package metrics

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/T-DWAG/zero2observe/model"
	"github.com/T-DWAG/zero2observe/storage"
)

type cacheEntry struct {
	snap      *Snapshot // 指标快照
	expiresAt time.Time // 过期时间
}

type Aggregator struct {
	store    storage.Storage // 存储后端
	cacheTTL time.Duration   // 缓存过期时间

	mu    sync.Mutex            // 互斥锁
	cache map[string]cacheEntry // 缓存
}

func NewAggregator(store storage.Storage, cacheTTL time.Duration) *Aggregator {
	//1、ttl
	if cacheTTL <= 0 {
		cacheTTL = 30 * time.Second
	}

	return &Aggregator{
		store:    store,
		cacheTTL: cacheTTL,
		cache:    make(map[string]cacheEntry),
	}
}

func (a *Aggregator) Aggregate(ctx context.Context, scope string, now time.Time) (*Snapshot, error) {

	//1、获取窗口时间
	from, to, err := windowFormScope(scope, now)
	if err != nil {
		return nil, err
	}

	//2、锁
	a.mu.Lock()

	//3、先缓存
	if entry, ok := a.cache[scope]; ok && entry.expiresAt.After(now) {
		snap := *entry.snap
		a.mu.Unlock()
		return &snap, nil
	}

	a.mu.Unlock()

	//4、从存储获取
	traces, err := listAllTraces(ctx, a.store, storage.TraceFilter{
		StartTime: from,
		EndTime:   to,
	})
	if err != nil {
		return nil, err
	}

	snap := &Snapshot{
		Scope: scope,
		From:  from,
		To:    to,
	}

	var durSum int64
	var success int64

	toolCount := map[string]int64{}
	for _, tr := range traces {
		snap.TotalTraces++
		snap.TotalTokens += tr.TotalTokens
		snap.TotalCost += tr.TotalCost

		durSum += tr.DurationMs

		if tr.Status == model.SpanStatusSuccess {
			success++
		}

		spans, err := a.store.GetTraceSpans(ctx, tr.TraceID)
		if err != nil {
			return nil, fmt.Errorf("spans %s: %w", tr.TraceID, err)
		}

		for _, sp := range spans {
			if sp.SpanType == model.SpanTypeTool && sp.ToolName != "" {
				toolCount[sp.ToolName]++
			}
		}
	}

	if snap.TotalTraces > 0 {
		snap.AvgDurationMs = float64(durSum) / float64(snap.TotalTraces)
		snap.SuccessRate = float64(success) / float64(snap.TotalTraces)
	}

	snap.TopTools = topNTools(toolCount, 5)

	//5、缓存
	a.mu.Lock()
	a.cache[scope] = cacheEntry{
		snap:      snap,
		expiresAt: now.Add(a.cacheTTL),
	}
	a.mu.Unlock()

	out := *snap
	return &out, nil
}

func topNTools(m map[string]int64, n int) []ToolStat {
	out := make([]ToolStat, 0, len(m))
	for name, c := range m {
		out = append(out, ToolStat{Name: name, Count: c})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count == out[j].Count {
			return out[i].Name < out[j].Name
		}
		return out[i].Count > out[j].Count
	})
	if len(out) > n {
		out = out[:n]
	}
	return out
}

func windowFormScope(scope string, now time.Time) (from, to time.Time, err error) {
	to = now.UTC()
	switch scope {
	case ScopeLast24h:
		from = to.Add(-24 * time.Hour)
	case ScopeLast7d:
		from = to.Add(-7 * 24 * time.Hour)
	case ScopeLast30d:
		from = to.Add(-30 * 24 * time.Hour)
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("invalid scope %q", scope)
	}
	return from, to, nil
}

func listAllTraces(ctx context.Context, store storage.Storage, filter storage.TraceFilter) ([]*model.Trace, error) {
	filter.Page = 1
	filter.Size = 100
	var all []*model.Trace
	for {
		items, total, err := store.ListTraces(ctx, filter)
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
		if int64(len(all)) >= total || len(items) == 0 {
			break
		}
		filter.Page++
	}
	return all, nil
}
