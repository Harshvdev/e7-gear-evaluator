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

	var req EvaluateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Layer 1: Which heroes/builds is this piece suited for?
	l1Result := EvaluateLayer1(req.Gear, req.ExcludedBuilds)

	// Layer 2: Is it worth rolling? (To be implemented)
	l2Result := EvaluateLayer2(req.Gear, l1Result)

	// Construct unified response containing all evaluation layers
	resp := EvaluationResponse{
		GearID:       req.Gear.ID,
		Layer1Result: l1Result, // Embedded for backward-compatible top-level JSON fields
		Layer1:       l1Result, // Explicit "layer1" envelope for clean layered access
		Layer2:       l2Result,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
}

func handleGetHeroes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	type HeroListItem struct {
		Name      string      `json:"name"`
		Icon      string      `json:"icon"`
		Rarity    int         `json:"rarity"`
		Role      string      `json:"role"`
		Attribute string      `json:"attribute"`
		Builds    []HeroBuild `json:"builds"`
	}

	heroes := make([]HeroListItem, 0, len(heroDatabase))
	for _, h := range heroDatabase {
		normName := normalizeName(h.Hero)
		item := HeroListItem{
			Name:      h.Hero,
			Icon:      heroIcons[normName],
			Rarity:    heroRarities[normName],
			Role:      heroRoles[normName],
			Attribute: heroAttributes[normName],
			Builds:    h.Builds,
		}
		heroes = append(heroes, item)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(heroes); err != nil {
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
