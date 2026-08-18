package api

import (
	"net/http"

	"github.com/T-DWAG/zero2observe/evaluation"
	"github.com/T-DWAG/zero2observe/storage"
)

type Server struct {
	store storage.Storage
	judge *evaluation.Judge // 可为 nil：未配置则 POST 返回 503
}

func NewServer(store storage.Storage) *Server {
	return &Server{store: store}
}

func (s *Server) WithJudge(judge *evaluation.Judge) *Server {
	s.judge = judge
	return s
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /api/v1/traces", s.handleListTraces)
	mux.HandleFunc("GET /api/v1/traces/{id}", s.handleGetTrace)
	mux.HandleFunc("GET /api/v1/traces/{id}/spans", s.handleGetTraceSpans)

	//evaluation
	mux.HandleFunc("POST /api/v1/evaluations", s.handleCreateEvaluation)
	mux.HandleFunc("GET /api/v1/evaluations/{trace_id}", s.handleListEvaluations)

	return mux
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
