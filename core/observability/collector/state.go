package collector

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/T-DWAG/zero2observe/model"
	"github.com/T-DWAG/zero2observe/storage"
	"github.com/google/uuid"
)

const spanChBuffer = 256

// State 一次 Trace 运行期间的采集状态。
// State 表示一次 Trace 采集的状态信息，包括 Trace 对象、span 堆栈、待提交的 span、存储及相关 channel。
type State struct {
	// 存储接口，用于持久化 Trace 数据
	storage storage.Storage

	// 当前 Trace
	Trace *model.Trace
	// spanStack 维护当前活跃的 span ID（栈结构）
	spanStack []string
	// pending 存储还未结束的 span，key 为 span ID
	pending map[string]*model.Span

	// spanCh 用于异步发送采集到的 Span
	spanCh chan *model.Span
	// traceCh 用于异步发送已完成 Trace
	traceCh chan *model.Trace
	// done 用于通知采集任务完成或退出
	done chan struct{}
}

func newState(store storage.Storage, sessionID, userInput string) *State {
	now := time.Now()
	return &State{
		storage: store,
		Trace: &model.Trace{
			TraceID:   uuid.New().String(),
			SessionID: sessionID,
			UserInput: userInput,
			StartTime: now,
			Status:    model.SpanStatusPending,
		},
		pending: make(map[string]*model.Span),
		spanCh:  make(chan *model.Span, spanChBuffer),
		traceCh: make(chan *model.Trace, 1),
		done:    make(chan struct{}),
	}
}

func (s *State) runWorker(ctx context.Context) {
	defer close(s.done)
	for {
		select {
		case <-ctx.Done():
			return
		case span, ok := <-s.spanCh:
			if !ok {
				return
			}
			if err := s.storage.SaveSpan(ctx, span); err != nil {
				log.Printf("failed to save span: %v", err)
			}
		case trace, ok := <-s.traceCh:
			if !ok {
				return
			}
			if err := s.storage.SaveTrace(ctx, trace); err != nil {
				log.Printf("failed to save trace: %v", err)
			}
		}

	}
}

func (s *State) parentSpanID() string {
	if len(s.spanStack) == 0 {
		return ""
	}
	return s.spanStack[len(s.spanStack)-1]
}

func (s *State) pushSpan(spanID string) { s.spanStack = append(s.spanStack, spanID) }

func (s *State) popSpan() {
	if len(s.spanStack) > 0 {
		s.spanStack = s.spanStack[:len(s.spanStack)-1]
	}
}

func (s *State) startSpan(spanType, name string) *model.Span {
	span := &model.Span{
		SpanID:       uuid.New().String(),
		TraceID:      s.Trace.TraceID,
		ParentSpanID: s.parentSpanID(),
		SpanType:     spanType,
		SpanName:     name,
		StartTime:    time.Now(),
		Status:       model.SpanStatusPending,
	}
	s.pending[span.SpanID] = span
	s.pushSpan(span.SpanID)
	return span
}

// mutate参数是一个可选的回调函数，用于在结束span前对其进行自定义修改。
// 如果调用方需要在span结束前修改span的某些字段（如状态、标签等），可以传入一个处理span的函数；
// 如果不需要额外修改，则传nil即可。
func (s *State) finishSpan(spanID string, mutate func(*model.Span)) {
	span, ok := s.pending[spanID]
	if !ok {
		return
	}
	if mutate != nil {
		mutate(span)
	}

	span.EndTime = time.Now()
	span.DurationMs = span.EndTime.Sub(span.StartTime).Milliseconds()
	delete(s.pending, spanID)
	s.popSpan()
	s.Trace.SpanCount++
	select {
	case s.spanCh <- span:
	default:
		log.Printf("[obs] span channel full, drop span %s", spanID)
	}
}

// failSpan 用于将指定 span 标记为失败（错误），并记录错误信息。
// 它的实现通过调用 finishSpan，并传入一个回调函数：
//   - 回调会把 span 的 Status 字段设置为错误状态（SpanStatusError）
//   - 并将错误信息 err.Error() 记录到 span 的 ErrorMsg 字段。
//
// 这样在结束 span 的同时，也能携带详细的错误原因。调用方只需提供 spanID 和错误对象即可。
func (s *State) failSpan(spanID string, err error) {
	s.finishSpan(spanID, func(sp *model.Span) {
		sp.Status = model.SpanStatusError
		sp.ErrorMsg = err.Error()
	})
}

func (s *State) addLLMTokens(_, _, total int64, cost float64) {
	s.Trace.TotalTokens += total
	s.Trace.TotalCost += cost
}

// finishTrace 用于结束一次 Trace（链路追踪）的采集，封装收尾逻辑：
// 1. 设置 Trace 的结束时间和总持续时间
// 2. 记录 Agent 执行的输出
// 3. 根据 runErr 是否为 nil，设置 Trace 的状态（成功/失败）
// 4. 关闭 spanCh，表示后续不再采集新的 Span
// 5. 尝试将 Trace 信息写入 traceCh（若信道满则丢弃并打印日志）
func (s *State) finishTrace(output string, runErr error) {
	// 记录 Trace 的结束时间
	s.Trace.EndTime = time.Now()
	// 计算 Trace 持续的毫秒数
	s.Trace.DurationMs = s.Trace.EndTime.Sub(s.Trace.StartTime).Milliseconds()
	// 保存 Agent 执行的输出内容
	s.Trace.AgentOutput = output
	// 判断是否有错误，根据情况设置 Trace 状态
	if runErr != nil {
		s.Trace.Status = model.SpanStatusError
	} else {
		s.Trace.Status = model.SpanStatusSuccess
	}
	// 关闭 spanCh，表示之后不会再写 Span 进来
	close(s.spanCh)
	// 非阻塞地尝试把 Trace 写入 traceCh，如果满了就丢弃并日志告警
	select {
	case s.traceCh <- s.Trace:
	default:
		log.Printf("[obs] trace channel full, drop trace %s", s.Trace.TraceID)
	}
}

func (s *State) waitDone() { <-s.done }

func toJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
