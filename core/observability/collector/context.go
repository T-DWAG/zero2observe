package collector

import (
	"context"

	"github.com/T-DWAG/zero2observe/storage"
	"github.com/cloudwego/eino/callbacks"
)

// ctxKey 用于 context.Value 的私有类型 key，避免与其他包 key 冲突。
type ctxKey struct{}

// withState 将采集 State 嵌入到 context.Context。
// 返回一个携带 State 的新 context。
func withState(ctx context.Context, s *State) context.Context {
	return context.WithValue(ctx, ctxKey{}, s)
}

// Config 定义创建 Trace 时所需的业务上下文参数。
// SessionID 表示会话唯一标识，UserInput 为用户输入内容。
type Config struct {
	SessionID string // 本次会话的唯一标识
	UserInput string // 当前用户输入内容
}

// WithObsCallback 准备一次 Agent 执行的采集上下文和回调处理。
// - ctx: 原始业务 context
// - store: 用于持久化 Trace/Span 的存储实现
// - cfg: 业务上下文参数（如 SessionID, UserInput）
// 返回：新 context、回调 handler、任务结束的收尾函数（finish）
func WithObsCallback(
	ctx context.Context,
	store storage.Storage,
	cfg Config,
) (context.Context, callbacks.Handler, func(context.Context, string, error)) {
	// 创建一个采集状态 State（一个 Trace 实例），绑定业务信息
	state := newState(store, cfg.SessionID, cfg.UserInput)
	// 创建独立的 context 和取消函数，驱动后台采集 worker
	workerCtx, cancel := context.WithCancel(context.Background())
	go state.runWorker(workerCtx)

	// 生成业务 Handler 实例
	handler := NewHandler(state)
	// 把 State 写入 context，方便后续取用
	ctx = withState(ctx, state)

	// finish：Trace 任务执行收尾函数（输出保存、资源回收、worker 关闭）
	finish := func(_ context.Context, output string, runErr error) {
		// 完成 Trace 状态收尾
		state.finishTrace(output, runErr)
		// 关闭 worker 的 context
		cancel()
		// 等待所有采集器资源释放
		state.waitDone()
	}
	return ctx, handler, finish
}

// TraceIDFromCtx 从 context.Context 中提取当前 Trace 的 TraceID。
// 若 context 未绑定 State，则返回空串。
func TraceIDFromCtx(ctx context.Context) string {
	if s, ok := ctx.Value(ctxKey{}).(*State); ok && s != nil {
		return s.Trace.TraceID
	}
	return ""
}
