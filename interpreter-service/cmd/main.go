package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/Oleja123/code-vizualization/interpreter-service/internal/handler"
)

func main() {
	port := flag.Int("port", 8080, "Port to listen on")
	flag.Parse()

	http.Handle("/snapshot", handler.NewSnapshotHandler())

	address := fmt.Sprintf(":%d", *port)
	log.Printf("interpreter-service listening on %s", address)

	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
