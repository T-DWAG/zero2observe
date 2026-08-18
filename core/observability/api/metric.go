package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/T-DWAG/zero2observe/metrics"
)

func (s *Server) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	//1、检查聚合器
	if s.agg == nil {
		writeError(w, http.StatusServiceUnavailable, "metrics not configured")
		return
	}
	//2、获取scope
	scope := r.URL.Query().Get("scope")
	if scope == "" {
		scope = metrics.ScopeLast24h
	}
	scope = strings.TrimSpace(scope)
	//3、聚合
	snap, err := s.agg.Aggregate(r.Context(), scope, time.Now().UTC())
	if err != nil {
		if strings.Contains(err.Error(), "invalid scope") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, snap)
}
