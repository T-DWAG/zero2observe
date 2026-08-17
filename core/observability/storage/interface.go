package storage

import (
	"context"

	"github.com/T-DWAG/zero2observe/model"
)

// Storage 存储抽象；Step 3 提供 PostgreSQL 实现，Step 2 可用 MemoryStorage 验证采集。
type Storage interface {
	SaveSpan(ctx context.Context, span *model.Span) error
	SaveTrace(ctx context.Context, trace *model.Trace) error
}
