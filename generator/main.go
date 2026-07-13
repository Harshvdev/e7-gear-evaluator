package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type GenerateOptions struct {
	Rarities  []string `json:"rarities"`
	Levels    []int    `json:"levels"`
	Slots     []string `json:"slots"`
	Sets      []string `json:"sets"`
	MainTypes []string `json:"mainTypes"`
}

type EnhanceRequest struct {
	Gear        Gear `json:"gear"`
	TargetLevel int  `json:"targetLevel"`
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/config", handleConfig)
	mux.HandleFunc("/api/generate", handleGenerateGear)
	mux.HandleFunc("/api/enhance", handleEnhanceGear)

	handler := corsMiddleware(mux)

	port := ":8080"
	fmt.Printf("Starting E7 Gear Generator API Server on http://localhost%s...\n", port)
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := struct {
		Slots             []string          `json:"slots"`
		Sets              []string          `json:"sets"`
		SetLabels         map[string]string `json:"setLabels"`
		StatLabels        map[string]string `json:"statLabels"`
		FlexibleMainStats []string          `json:"flexibleMainStats"`
	}{
		Slots:             SLOTS,
		Sets:              SETS,
		SetLabels:         SET_LABELS,
		StatLabels:        STAT_LABELS,
		FlexibleMainStats: FLEXIBLE_MAIN_STATS,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGenerateGear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var opts GenerateOptions
	err := json.NewDecoder(r.Body).Decode(&opts)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	gear := generateGear(opts)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gear)
}

func handleEnhanceGear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req EnhanceRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	gear := enhanceToLevel(req.Gear, req.TargetLevel)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gear)
}
