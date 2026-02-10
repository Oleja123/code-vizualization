package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
)

// ParseRequest представляет запрос на парсинг кода
type ParseRequest struct {
	Code string `json:"code"`
}

// ParseResponse представляет успешный ответ
type ParseResponse struct {
	AST *converter.Program `json:"ast"`
}

// ErrorResponse представляет ошибку парсинга
type ErrorResponse struct {
	Error    string              `json:"error"`
	Code     converter.ErrorCode `json:"code"`
	Message  string              `json:"message"`
	Location *converter.Location `json:"location,omitempty"`
	NodeType string              `json:"nodeType,omitempty"`
}

// handleParse парсит C код и возвращает AST или ошибку
func handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Only POST method is allowed",
		})
		return
	}

	// Парсим запрос
	var req ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid JSON request: " + err.Error(),
		})
		return
	}

	if req.Code == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Code field is empty",
		})
		return
	}

	// Парсим код
	conv := converter.New()
	program, err := conv.ParseToAST(req.Code)

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		// Ошибка парсинга - возвращаем 400 Bad Request с информацией об ошибке
		w.WriteHeader(http.StatusBadRequest)

		errResp := ErrorResponse{
			Error:    "Parse error",
			Code:     err.Code,
			Message:  err.Message,
			NodeType: err.NodeType,
		}

		if err.Loc.Line > 0 {
			errResp.Location = &err.Loc
		}

		json.NewEncoder(w).Encode(errResp)
		return
	}

	// Успешный парс - возвращаем 200 OK с AST
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ParseResponse{
		AST: program,
	})
}

// handleHealth проверка живой ли сервис
func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "cst-to-ast-service",
		"version": "1.0.0",
	})
}

// handleInfo информация об API
func handleInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	info := map[string]interface{}{
		"name":        "CST-to-AST Converter",
		"description": "Converts C code to Abstract Syntax Tree",
		"endpoints": map[string]interface{}{
			"POST /parse": "Parse C code and return AST or error",
			"GET /health": "Health check",
			"GET /info":   "API information",
		},
		"supported_constructs": map[string]interface{}{
			"types": []string{"int", "int*", "int**", "int[N]"},
			"statements": []string{
				"variable declaration", "function declaration",
				"if/else if/else", "while", "for",
				"return", "break", "continue", "block",
			},
			"expressions": []string{
				"variables", "integer literals", "binary operations",
				"unary operations", "assignments", "function calls",
				"array access", "array initialization",
			},
			"operators": map[string][]string{
				"binary":     {"+", "-", "*", "/", "%", "==", "!=", "<", ">", "<=", ">=", "&&", "||", "&", "|", "^", "<<", ">>"},
				"unary":      {"-", "!", "*", "&", "++", "--"},
				"assignment": {"=", "+=", "-=", "*=", "/=", "%=", "&=", "|=", "^=", "<<=", ">>="},
			},
		},
	}

	json.NewEncoder(w).Encode(info)
}

func main() {
	// Регистрируем обработчики
	http.HandleFunc("/parse", handleParse)
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/info", handleInfo)

	// Запускаем сервер
	port := ":8080"
	fmt.Printf("🚀 CST-to-AST Service starting on http://localhost%s\n", port)
	fmt.Println("\nEndpoints:")
	fmt.Println("  POST /parse  - Parse C code")
	fmt.Println("  GET  /health - Health check")
	fmt.Println("  GET  /info   - API information")
	fmt.Println("\nExample:")
	fmt.Println(`  curl -X POST http://localhost:8080/parse \
    -H "Content-Type: application/json" \
    -d '{"code": "int x = 42;"}'`)
	fmt.Println()

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
