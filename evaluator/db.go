package main

// ---------------------------------------------------------------------------
// db.go — Database loading and hero metadata.
//
// Owns:
//   - Global in-memory state (hero builds, icons, roles, etc.)
//   - initDatabase() — called once at startup
//   - Path resolution helpers
// ---------------------------------------------------------------------------

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

// ---------------------------------------------------------------------------
// Global state (populated once at startup by initDatabase)
// ---------------------------------------------------------------------------

var (
	heroDatabase   []HeroStats
	heroIcons      = make(map[string]string)
	heroRarities   = make(map[string]int)
	heroRoles      = make(map[string]string)
	heroAttributes = make(map[string]string)
)

// ---------------------------------------------------------------------------
// Startup
// ---------------------------------------------------------------------------

// initDatabase loads average_build_stats.json and stats.json into memory.
// Returns an error if the main database cannot be loaded; icon/metadata
// failures are logged as warnings but do not abort startup.
func initDatabase() error {
	dbPath := findDBPath()
	if dbPath == "" {
		return fmt.Errorf("could not find average_build_stats.json in any expected path")
	}
	fmt.Printf("Loading hero database from: %s\n", dbPath)

	data, err := os.ReadFile(dbPath)
	if err != nil {
		return fmt.Errorf("failed to read database: %w", err)
	}
	if err := json.Unmarshal(data, &heroDatabase); err != nil {
		return fmt.Errorf("failed to parse database: %w", err)
	}
	fmt.Printf("Loaded %d heroes.\n", len(heroDatabase))

	loadHeroMetadata()
	return nil
}

// loadHeroMetadata loads icon, rarity, role, and attribute from stats.json.
// Non-fatal: missing or malformed stats.json is logged as a warning.
func loadHeroMetadata() {
	statsPath := findStatsPath()
	if statsPath == "" {
		log.Println("Warning: stats.json not found — hero icons and metadata will be unavailable.")
		return
	}

	data, err := os.ReadFile(statsPath)
	if err != nil {
		log.Printf("Warning: could not read stats.json: %v\n", err)
		return
	}

	var stats map[string]StatsHero
	if err := json.Unmarshal(data, &stats); err != nil {
		log.Printf("Warning: could not parse stats.json: %v\n", err)
		return
	}

	for _, h := range stats {
		key := normalizeName(h.Name)
		heroIcons[key] = h.Icon
		heroRarities[key] = h.Rarity
		heroRoles[key] = h.Role
		heroAttributes[key] = h.Attribute
	}
	fmt.Printf("Loaded metadata for %d heroes.\n", len(heroIcons))
}

// ---------------------------------------------------------------------------
// Path resolution
// ---------------------------------------------------------------------------

func findDBPath() string {
	candidates := []string{
		"/home/thesky/Documents/Projects/Tools/e7evaluator-z/v2/v3/data/heroes/average_build_stats.json",
		"data/heroes/average_build_stats.json",
		"../data/heroes/average_build_stats.json",
		"../../data/heroes/average_build_stats.json",
	}
	return firstExisting(candidates)
}

func findStatsPath() string {
	candidates := []string{
		"/home/thesky/Documents/Projects/Tools/e7evaluator-z/v2/v3/data/heroes/stats.json",
		"data/heroes/stats.json",
		"../data/heroes/stats.json",
		"../../data/heroes/stats.json",
	}
	return firstExisting(candidates)
}

func firstExisting(paths []string) string {
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// normalizeName produces a lowercase, punctuation-stripped key used to
// cross-reference hero names between the two JSON files.
func normalizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "")
	name = strings.ReplaceAll(name, "-", "")
	name = strings.ReplaceAll(name, "'", "")
	name = strings.ReplaceAll(name, ".", "")
	return name
}
