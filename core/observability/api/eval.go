package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/T-DWAG/zero2observe/storage"
)

type createEvaluationRequest struct {
	TraceID string `json:"trace_id"`
}

func (s *Server) handleCreateEvaluation(w http.ResponseWriter, r *http.Request) {
	if s.judge == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "evaluation not configured"})
		return
	}

	var req createEvaluationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TraceID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	evals, err := s.judge.Evaluate(r.Context(), req.TraceID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "trace not found"})
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"trace_id": req.TraceID,
		"items":    evals,
		"total":    len(evals),
	})
}

func (s *Server) handleListEvaluations(w http.ResponseWriter, r *http.Request) {
	//1、 request 解析 trace id
	traceID := r.PathValue("trace_id")
	if traceID == "" {
		writeError(w, http.StatusBadRequest, "trace_id is required")
		return
	}
	//2、库里面找 trace
	if _, err := s.store.GetTrace(r.Context(), traceID); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "trace not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	//3、库里面找 evaluation
	list, err := s.store.ListEvaluations(r.Context(), traceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"trace_id": traceID,
		"items":    list,
		"total":    len(list),
	})
}
