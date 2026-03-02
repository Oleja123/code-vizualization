package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/application/eventdispatcher"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/application/interpreter"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/snapshot"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/step"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/infrastructure/cache"
	configinfra "github.com/Oleja123/code-vizualization/interpreter-service/internal/infrastructure/config"
	"github.com/Oleja123/code-vizualization/semantic-analyzer-service/pkg/onecompiler"
	"github.com/Oleja123/code-vizualization/semantic-analyzer-service/pkg/validator"
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
	StepsCount  int                `json:"steps_count,omitempty"`
	Result      *int               `json:"result,omitempty"`
	Snapshot    *snapshot.Snapshot `json:"snapshot,omitempty"`
}

func NewSnapshotHandler(oneCompilerConfigPath string, cacher cache.Cacher) http.HandlerFunc {
	conv := converter.New()
	cfg := configinfra.LoadOrDefault(oneCompilerConfigPath)
	val := buildValidator(cfg)

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

		cacheKey := fmt.Sprintf("code:%s:max_elements:%d:max_steps:%d", req.Code, cfg.MaxAllocatedElements, cfg.MaxSteps)

		var steps []step.Step
		var stepBegin int
		var result *int
		var execErr error

		if cacher != nil {
			cachedInfo, err := cacher.Get(r.Context(), cacheKey)
			if err == nil && cachedInfo.Value != nil {
				steps = cachedInfo.Value
				stepBegin = cachedInfo.StepBegin
				result = cachedInfo.Result
				execErr = cachedInfo.Err
			}
		}

		if steps == nil && execErr == nil {
			program, parseErr := conv.ParseToAST(req.Code)
			if parseErr != nil {
				writeJSON(w, http.StatusBadRequest, SnapshotResponse{Success: false, Error: "parse error: " + parseErr.Error()})
				return
			}

			if err := val.ValidateProgram(program, req.Code); err != nil {
				var unavailableErr validator.CompileUnavailableError
				if errors.As(err, &unavailableErr) {
					writeJSON(w, http.StatusServiceUnavailable, SnapshotResponse{Success: false, Error: err.Error()})
					return
				}

				writeJSON(w, http.StatusBadRequest, SnapshotResponse{Success: false, Error: "semantic error: " + err.Error()})
				return
			}

			runner := interpreter.NewInterpreterWithLimits(cfg.MaxAllocatedElements, cfg.MaxSteps)
			result, steps, stepBegin, execErr = runner.ExecuteProgram(program)
			if execErr != nil && steps == nil {
				writeJSON(w, http.StatusBadRequest, SnapshotResponse{Success: false, Error: "error: " + execErr.Error()})
				return
			}

			if cacher != nil {
				cachedInfo := cache.CachedInfo{
					Value:     steps,
					StepBegin: stepBegin,
					Result:    result,
					Err:       execErr,
				}
				_ = cacher.Set(r.Context(), cacheKey, cachedInfo)
			}
		}

		ed := eventdispatcher.NewEventDispatcher(stepBegin)
		ed.Steps = steps
		if err := ed.ApplyStep(req.Step); err != nil {
			writeJSON(w, http.StatusBadRequest, SnapshotResponse{Success: false, Error: err.Error()})
			return
		}

		stepsCount := len(steps) - stepBegin
		if stepsCount < 0 {
			stepsCount = 0
		}

		currentStep := ed.GetCurrentStep() - stepBegin
		if currentStep < 0 {
			currentStep = 0
		}

		writeJSON(w, http.StatusOK, SnapshotResponse{
			Success:     true,
			Step:        req.Step,
			CurrentStep: currentStep,
			StepsCount:  stepsCount,
			Result:      result,
			Snapshot:    ed.GetSnapshot(),
		})
	}
}

func buildValidator(cfg *configinfra.Config) *validator.SemanticValidator {
	if cfg == nil || !cfg.Enabled {
		return validator.New()
	}

	timeout := cfg.TimeoutSeconds
	if timeout == 0 {
		timeout = 10
	}

	client := onecompiler.NewClient(
		cfg.APIURL,
		cfg.APIKey,
		timeout,
	)

	return validator.NewWithOneCompilerClient(client)
}

func writeJSON(w http.ResponseWriter, status int, body SnapshotResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
