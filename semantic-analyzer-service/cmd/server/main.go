package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
	"github.com/Oleja123/code-vizualization/semantic-analyzer-service/internal/infrastructure/config"
	"github.com/Oleja123/code-vizualization/semantic-analyzer-service/internal/infrastructure/onecompiler"
	"github.com/Oleja123/code-vizualization/semantic-analyzer-service/pkg/validator"
)

type ValidateRequest struct {
	Code string `json:"code"`
}

type ValidateResponse struct {
	Success bool               `json:"success"`
	Program *converter.Program `json:"program,omitempty"`
	Error   string             `json:"error,omitempty"`
}

var (
	conv     *converter.CConverter
	val      *validator.SemanticValidator
	ocClient *onecompiler.Client
	cfg      *config.Config
	logger   *slog.Logger
)

func init() {
	conv = converter.NewCConverter()
	val = validator.New()
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// handleValidate обрабатывает POST запрос с кодом на валидацию
func handleValidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ValidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ValidateResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	// Парсим код
	tree, err := conv.Parse([]byte(req.Code))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ValidateResponse{
			Success: false,
			Error:   "Parse error: " + err.Error(),
		})
		return
	}

	// Конвертируем в AST
	program, err := conv.ConvertToProgram(tree, []byte(req.Code))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ValidateResponse{
			Success: false,
			Error:   "Conversion error: " + err.Error(),
		})
		return
	}

	prog := program.(*converter.Program)

	// Валидируем
	if err := val.ValidateProgram(prog); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ValidateResponse{
			Success: false,
			Error:   "Semantic error: " + err.Error(),
		})
		return
	}

	// Check compilation if OneCompiler is enabled
	if cfg.OneCompiler.Enabled && ocClient != nil {
		result, err := ocClient.CompileC(req.Code)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(ValidateResponse{
				Success: false,
				Error:   "Compilation check unavailable: " + err.Error(),
			})
			return
		}

		// Check if compilation failed (error in stderr)
		if result.Stderr != "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ValidateResponse{
				Success: false,
				Error:   "Compilation error: " + result.Stderr,
			})
			return
		}
	}

	// Успех
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ValidateResponse{
		Success: true,
		Program: prog,
	})
}

// handleHealth проверяет статус сервиса
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "semantic-analyzer-service",
	})
}

// handleInfo выводит информацию об API
func handleInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	info := map[string]interface{}{
		"service": "Semantic Analyzer Service",
		"version": "1.0.0",
		"endpoints": map[string]string{
			"POST /validate": "Validate C code",
			"GET /health":    "Health check",
			"GET /info":      "Service information",
		},
		"supported_types": []string{"int", "void"},
		"supported_operators": map[string][]string{
			"assignment": {"=", "+=", "-=", "/=", "%="},
			"unary":      {"-", "!", "++", "--"},
			"binary":     {"+", "-", "*", "/", "%", "==", "!=", "<", "<=", ">", ">=", "&&", "||"},
		},
		"unsupported_operators": []string{"&", "|", "^", "<<", ">>"},
	}
	json.NewEncoder(w).Encode(info)
}

func main() {
	port := flag.Int("port", 0, "Port to listen on (overrides config)")
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	flag.Parse()

	// Load config
	cfg = config.LoadConfigOrDefault(*configPath)

	// Override port if provided via flag
	if *port > 0 {
		cfg.Server.Port = *port
	}

	// Initialize OneCompiler client if enabled
	if cfg.OneCompiler.Enabled {
		ocClient = onecompiler.NewClient(
			cfg.OneCompiler.APIURL,
			cfg.OneCompiler.APIKey,
			cfg.OneCompiler.TimeoutSeconds,
		)
		logger.Info("OneCompiler client initialized",
			slog.Int("timeout_seconds", cfg.OneCompiler.TimeoutSeconds))
	}

	// Регистрируем обработчики
	http.HandleFunc("/validate", handleValidate)
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/info", handleInfo)

	address := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Info("Starting Semantic Analyzer Server",
		slog.String("address", address))

	if err := http.ListenAndServe(address, nil); err != nil {
		logger.Error("Server error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
