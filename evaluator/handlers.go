package main

import (
	"encoding/json"
	"net/http"
)

// ---------------------------------------------------------------------------
// handlers.go — HTTP handlers and middleware for the Evaluator API.
// ---------------------------------------------------------------------------

func handleEvaluate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var gear Gear
	if err := json.NewDecoder(r.Body).Decode(&gear); err != nil {
		http.Error(w, "Invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Layer 1: Which heroes/builds is this piece suited for?
	l1Result := EvaluateLayer1(gear)

	// Layer 2: Is it worth rolling? (To be implemented)
	l2Result := EvaluateLayer2(gear, l1Result)

	// Construct unified response containing all evaluation layers
	resp := EvaluationResponse{
		GearID:       gear.ID,
		Layer1Result: l1Result, // Embedded for backward-compatible top-level JSON fields
		Layer1:       l1Result, // Explicit "layer1" envelope for clean layered access
		Layer2:       l2Result,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

// corsMiddleware adds standard CORS headers allowing local development across ports.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
