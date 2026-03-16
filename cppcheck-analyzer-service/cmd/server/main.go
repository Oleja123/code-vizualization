package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Oleja123/code-vizualization/cppcheck-analyzer-service/internal/application/analyzer"
	"github.com/Oleja123/code-vizualization/cppcheck-analyzer-service/internal/handler"
	configinfra "github.com/Oleja123/code-vizualization/cppcheck-analyzer-service/internal/infrastructure/config"
)

func main() {
	port := flag.Int("port", 0, "Port to listen on (overrides config)")
	configPath := flag.String("config", "config.yaml", "Path to YAML config")
	flag.Parse()

	cfg := configinfra.LoadOrDefault(*configPath)
	if *port > 0 {
		cfg.Server.Port = *port
	}

	engine := analyzer.New(
		cfg.Cppcheck.Path,
		cfg.Cppcheck.Std,
		cfg.Cppcheck.Enable,
		cfg.Cppcheck.Inconclusive,
		cfg.Cppcheck.MaxIssues,
	)

	timeout := time.Duration(cfg.Cppcheck.TimeoutSeconds) * time.Second

	http.HandleFunc("/analyze", handler.NewAnalyzeHandler(engine, timeout))
	http.HandleFunc("/health", handler.NewHealthHandler())
	http.HandleFunc("/info", handler.NewInfoHandler())

	address := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("cppcheck-analyzer-service listening on %s", address)

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
