package api

import (
	"net/http"

	"github.com/T-DWAG/zero2observe/storage"
)

type Server struct {
	store storage.Storage
}

func NewServer(store storage.Storage) *Server {
	return &Server{store: store}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /api/v1/traces", s.handleListTraces)
	mux.HandleFunc("GET /api/v1/traces/{id}", s.handleGetTrace)
	mux.HandleFunc("GET /api/v1/traces/{id}/spans", s.handleGetTraceSpans)
	return mux
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
