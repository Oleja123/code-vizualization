package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Oleja123/code-vizualization/cppcheck-analyzer-service/internal/application/analyzer"
)

type AnalyzeRequest struct {
	Code string `json:"code"`
}

type AnalyzeResponse struct {
	Success  bool             `json:"success"`
	Passed   bool             `json:"passed,omitempty"`
	Issues   []analyzer.Issue `json:"issues,omitempty"`
	Count    int              `json:"count,omitempty"`
	Error    string           `json:"error,omitempty"`
	Analyzer string           `json:"analyzer,omitempty"`
}

func NewAnalyzeHandler(engine *analyzer.Analyzer, timeout time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, AnalyzeResponse{Success: false, Error: "method not allowed"})
			return
		}

		var req AnalyzeRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, AnalyzeResponse{Success: false, Error: "invalid request body: " + err.Error()})
			return
		}

		if strings.TrimSpace(req.Code) == "" {
			writeJSON(w, http.StatusBadRequest, AnalyzeResponse{Success: false, Error: "code is required"})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		issues, _, err := engine.Analyze(ctx, req.Code)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, AnalyzeResponse{Success: false, Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, AnalyzeResponse{
			Success:  true,
			Passed:   len(issues) == 0,
			Issues:   issues,
			Count:    len(issues),
			Analyzer: "cppcheck",
		})
	}
}

func NewHealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "healthy",
			"service": "cppcheck-analyzer-service",
		})
	}
}

func NewInfoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"service":  "Cppcheck Analyzer Service",
			"version":  "1.0.0",
			"analyzer": "cppcheck",
			"language": "c",
			"endpoints": map[string]string{
				"POST /analyze": "Analyze C code with cppcheck",
				"GET /health":   "Health check",
				"GET /info":     "Service information",
			},
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
