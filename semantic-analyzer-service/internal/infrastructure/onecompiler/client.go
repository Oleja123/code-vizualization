package onecompiler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client клиент для взаимодействия с OneCompiler API
type Client struct {
	apiURL  string
	apiKey  string
	timeout time.Duration
	client  *http.Client
}

// FileContent представляет файл с кодом
type FileContent struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// CompileRequest представляет запрос на компиляцию
type CompileRequest struct {
	Language string        `json:"language"`
	Files    []FileContent `json:"files"`
	Stdin    string        `json:"stdin,omitempty"`
}

// CompileResponse представляет ответ от OneCompiler
type CompileResponse struct {
	Status           string `json:"status"`
	Exception        string `json:"exception"`
	Stdout           string `json:"stdout"`
	Stderr           string `json:"stderr"`
	CompilationTime  int    `json:"compilationTime"`
	ExecutionTime    int    `json:"executionTime"`
	MemoryUsed       int    `json:"memoryUsed"`
	CreditsRemaining int    `json:"creditsRemaining"`
}

// NewClient создает новый клиент OneCompiler
func NewClient(apiURL string, apiKey string, timeoutSeconds int) *Client {
	timeout := time.Duration(timeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	return &Client{
		apiURL:  apiURL,
		apiKey:  apiKey,
		timeout: timeout,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// CompileC компилирует C код
func (c *Client) CompileC(code string) (*CompileResponse, error) {
	return c.compile("c", code)
}

// compile выполняет запрос на компиляцию
func (c *Client) compile(language string, code string) (*CompileResponse, error) {
	req := CompileRequest{
		Language: language,
		Files: []FileContent{
			{
				Name:    "main.c",
				Content: code,
			},
		},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Отправляем запрос
	httpReq, err := http.NewRequest(http.MethodPost, c.apiURL+"/run", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check if we got an error status code
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error: HTTP %d - %s", resp.StatusCode, string(body))
	}

	var compileResp CompileResponse
	if err := json.Unmarshal(body, &compileResp); err != nil {
		return nil, fmt.Errorf("failed to parse response (expected JSON, got %d bytes): %w", len(body), err)
	}

	// Если статус не "success" - была ошибка
	if compileResp.Status != "success" {
		// Выводим stderr если доступен
		if compileResp.Stderr != "" {
			return &compileResp, fmt.Errorf("compilation failed: %s", compileResp.Stderr)
		}
		if compileResp.Exception != "" {
			return &compileResp, fmt.Errorf("compilation error: %s", compileResp.Exception)
		}
		return &compileResp, fmt.Errorf("compilation failed with status: %s", compileResp.Status)
	}

	return &compileResp, nil
}
