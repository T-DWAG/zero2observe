package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/T-DWAG/zero2observe/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PostgresStorage 用 GORM 把 Span/Trace 写入 PostgreSQL。
type PostgresStorage struct {
	db *gorm.DB
}

// NewPostgresStorage 包装已打开的 *gorm.DB。
// 调用方负责：Open → model.AutoMigrate → NewPostgresStorage。
func NewPostgresStorage(db *gorm.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

// OpenPostgres 打开连接；dsn 示例见手敲清单第三节。
func OpenPostgres(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	return db, nil
}

// SaveSpan：按 span_id 幂等写入（重复采集不炸）。
func (p *PostgresStorage) SaveSpan(ctx context.Context, span *model.Span) error {
	err := p.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "span_id"}},
			UpdateAll: true,
		}).
		Create(span).Error
	if err != nil {
		return fmt.Errorf("save span: %w", err)
	}
	return nil
}

// SaveTrace：按 trace_id 幂等写入。
func (p *PostgresStorage) SaveTrace(ctx context.Context, trace *model.Trace) error {
	err := p.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "trace_id"}},
			UpdateAll: true,
		}).
		Create(trace).Error
	if err != nil {
		return fmt.Errorf("save trace: %w", err)
	}
	return nil
}

func (p *PostgresStorage) GetTrace(ctx context.Context, traceID string) (*model.Trace, error) {
	var tr model.Trace
	err := p.db.WithContext(ctx).Where("trace_id = ?", traceID).First(&tr).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get trace: %w", err)
	}
	return &tr, nil
}

func (p *PostgresStorage) GetTraceSpans(ctx context.Context, traceID string) ([]*model.Span, error) {
	var spans []*model.Span
	err := p.db.WithContext(ctx).
		Where("trace_id = ?", traceID).
		Order("start_time asc").
		Find(&spans).Error
	if err != nil {
		return nil, fmt.Errorf("get trace spans: %w", err)
	}
	return spans, nil
}

func (p *PostgresStorage) ListTraces(ctx context.Context, filter TraceFilter) ([]*model.Trace, int64, error) {
	filter = filter.normalize()
	q := p.db.WithContext(ctx).Model(&model.Trace{})

	if filter.SessionID != "" {
		q = q.Where("session_id = ?", filter.SessionID)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if !filter.StartTime.IsZero() {
		q = q.Where("start_time >= ?", filter.StartTime)
	}
	if !filter.EndTime.IsZero() {
		q = q.Where("start_time <= ?", filter.EndTime)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count traces: %w", err)
	}

	var traces []*model.Trace
	err := q.Order("start_time desc").
		Offset(filter.offset()).
		Limit(filter.Size).
		Find(&traces).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list traces: %w", err)
	}
	return traces, total, nil
}
