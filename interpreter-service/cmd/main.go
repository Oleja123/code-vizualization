package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/Oleja123/code-vizualization/interpreter-service/internal/handler"
	configinfra "github.com/Oleja123/code-vizualization/interpreter-service/internal/infrastructure/config"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to YAML config")
	flag.Parse()

	cfg := configinfra.LoadOrDefault(*configPath)

	listenPort := cfg.Port

	http.Handle("/snapshot", handler.NewSnapshotHandler(*configPath))

	address := fmt.Sprintf(":%d", listenPort)
	log.Printf("interpreter-service listening on %s", address)

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
