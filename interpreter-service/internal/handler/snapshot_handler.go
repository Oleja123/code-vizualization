package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/application/eventdispatcher"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/application/interpreter"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/snapshot"
)

type SnapshotRequest struct {
	Code string `json:"code"`
	Step int    `json:"step"`
}

type SnapshotResponse struct {
	Success     bool               `json:"success"`
	Error       string             `json:"error,omitempty"`
	Step        int                `json:"step,omitempty"`
	CurrentStep int                `json:"current_step,omitempty"`
	StepBegin   int                `json:"step_begin,omitempty"`
	StepsCount  int                `json:"steps_count,omitempty"`
	Result      *int               `json:"result,omitempty"`
	Snapshot    *snapshot.Snapshot `json:"snapshot,omitempty"`
}

func NewSnapshotHandler() http.HandlerFunc {
	conv := converter.New()
	runner := interpreter.NewInterpreter()

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, SnapshotResponse{Success: false, Error: "method not allowed"})
			return
		}

		var req SnapshotRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, SnapshotResponse{Success: false, Error: "invalid request body: " + err.Error()})
			return
		}

		if strings.TrimSpace(req.Code) == "" {
			writeJSON(w, http.StatusBadRequest, SnapshotResponse{Success: false, Error: "code is required"})
			return
		}

		if req.Step < 0 {
			writeJSON(w, http.StatusBadRequest, SnapshotResponse{Success: false, Error: "step must be non-negative"})
			return
		}

		program, parseErr := conv.ParseToAST(req.Code)
		if parseErr != nil {
			writeJSON(w, http.StatusBadRequest, SnapshotResponse{Success: false, Error: "parse error: " + parseErr.Error()})
			return
		}

		result, steps, stepBegin, execErr := runner.ExecuteProgram(program)
		if execErr != nil {
			writeJSON(w, http.StatusBadRequest, SnapshotResponse{Success: false, Error: "runtime error: " + execErr.Error()})
			return
		}

		ed := eventdispatcher.NewEventDispatcher(stepBegin)
		ed.Steps = steps
		if err := ed.ApplyStep(req.Step); err != nil {
			writeJSON(w, http.StatusBadRequest, SnapshotResponse{Success: false, Error: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, SnapshotResponse{
			Success:     true,
			Step:        req.Step,
			CurrentStep: ed.GetCurrentStep(),
			StepBegin:   stepBegin,
			StepsCount:  len(steps),
			Result:      result,
			Snapshot:    ed.GetSnapshot(),
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, body SnapshotResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
