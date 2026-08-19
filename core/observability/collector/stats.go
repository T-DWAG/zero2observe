package collector

import "sync/atomic"

type Stats struct {
	DroppedSpans   atomic.Int64 // 丢弃的 span 数量
	SaveSpanFails  atomic.Int64 // 保存 span 失败的数量
	SaveTraceFails atomic.Int64 // 保存 trace 失败的数量
	SaveSpanOK     atomic.Int64 // 保存 span 成功的数量
	SaveTraceOK    atomic.Int64 // 保存 trace 成功的数量
}

func (s *Stats) snapshot() (dropped, spanFail, traceFail, spanOK, traceOK int64) {
	return s.DroppedSpans.Load(),
		s.SaveSpanFails.Load(),
		s.SaveTraceFails.Load(),
		s.SaveSpanOK.Load(),
		s.SaveTraceOK.Load()
}
