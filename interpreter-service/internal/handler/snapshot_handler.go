package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/application/eventdispatcher"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/application/interpreter"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/snapshot"
	"github.com/Oleja123/code-vizualization/semantic-analyzer-service/pkg/onecompiler"
	"github.com/Oleja123/code-vizualization/semantic-analyzer-service/pkg/validator"
	"gopkg.in/yaml.v3"
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

type oneCompilerConfigFile struct {
	OneCompiler struct {
		APIURL         string `yaml:"api_url"`
		APIKey         string `yaml:"api_key"`
		Enabled        bool   `yaml:"enabled"`
		TimeoutSeconds int    `yaml:"timeout_seconds"`
	} `yaml:"onecompiler"`
}

func NewSnapshotHandler(oneCompilerConfigPath string) http.HandlerFunc {
	conv := converter.New()
	val := buildValidator(oneCompilerConfigPath)

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

		if err := val.ValidateProgram(program, req.Code); err != nil {
			var unavailableErr validator.CompileUnavailableError
			if errors.As(err, &unavailableErr) {
				writeJSON(w, http.StatusServiceUnavailable, SnapshotResponse{Success: false, Error: err.Error()})
				return
			}

			writeJSON(w, http.StatusBadRequest, SnapshotResponse{Success: false, Error: "semantic error: " + err.Error()})
			return
		}

		runner := interpreter.NewInterpreter()
		result, steps, stepBegin, execErr := runner.ExecuteProgram(program)
		if execErr != nil {
			writeJSON(w, http.StatusBadRequest, SnapshotResponse{Success: false, Error: "error: " + execErr.Error()})
			return
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

func buildValidator(oneCompilerConfigPath string) *validator.SemanticValidator {
	if strings.TrimSpace(oneCompilerConfigPath) == "" {
		return validator.New()
	}

	cfg, err := loadOneCompilerConfig(oneCompilerConfigPath)
	if err != nil || !cfg.OneCompiler.Enabled {
		return validator.New()
	}

	timeout := cfg.OneCompiler.TimeoutSeconds
	if timeout == 0 {
		timeout = 10
	}

	client := onecompiler.NewClient(
		cfg.OneCompiler.APIURL,
		cfg.OneCompiler.APIKey,
		timeout,
	)

	return validator.NewWithOneCompilerClient(client)
}

func loadOneCompilerConfig(path string) (*oneCompilerConfigFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read onecompiler config: %w", err)
	}

	var cfg oneCompilerConfigFile
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse onecompiler config: %w", err)
	}

	return &cfg, nil
}

func writeJSON(w http.ResponseWriter, status int, body SnapshotResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
