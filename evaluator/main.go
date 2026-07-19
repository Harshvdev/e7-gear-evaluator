package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	if err := initDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/evaluate", handleEvaluate)
	mux.HandleFunc("/api/heroes", handleGetHeroes)

	handler := corsMiddleware(mux)

	port := ":8081"
	fmt.Printf("Starting E7 Gear Evaluator API on http://localhost%s...\n", port)
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
