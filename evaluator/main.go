package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

// Data structures from average_build_stats.json

type Sav struct {
	Atk float64 `json:"atk"`
	Def float64 `json:"def"`
	Hp  float64 `json:"hp"`
	Spd float64 `json:"spd"`
	Cc  float64 `json:"cc"`
	Cd  float64 `json:"cd"`
	Eff float64 `json:"eff"`
	Res float64 `json:"res"`
}

func (s Sav) Get(key string) float64 {
	switch key {
	case "atk":
		return s.Atk
	case "def":
		return s.Def
	case "hp":
		return s.Hp
	case "spd":
		return s.Spd
	case "cc":
		return s.Cc
	case "cd":
		return s.Cd
	case "eff":
		return s.Eff
	case "res":
		return s.Res
	default:
		return 0.0
	}
}

type HeroBuild struct {
	Rank  int      `json:"rank"`
	Usage float64  `json:"usage"`
	Sets  []string `json:"sets"`
	Sav   Sav      `json:"sav"`
}

type HeroStats struct {
	Hero   string      `json:"hero"`
	Builds []HeroBuild `json:"builds"`
}

// Stats.json structures

type StatsHero struct {
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	Rarity    int    `json:"rarity"`
	Role      string `json:"role"`
	Attribute string `json:"attribute"`
}

// Gear input structures

type Substat struct {
	Type  string  `json:"type"`
	Value float64 `json:"value"`
	Rolls int     `json:"rolls"`
}

type MainStat struct {
	Type  string  `json:"type"`
	Value float64 `json:"value"`
}

type GearScore struct {
	GS float64 `json:"gs"`
	ES int     `json:"es"`
}

type Gear struct {
	ID       string    `json:"id"`
	Set      string    `json:"set"`
	Slot     string    `json:"slot"`
	Rarity   string    `json:"rarity"`
	Level    int       `json:"level"`
	Enhance  int       `json:"enhance"`
	Main     MainStat  `json:"main"`
	Substats []Substat `json:"substats"`
	Score    GearScore `json:"score"`
}

// Response structures

type LandmineStat struct {
	StatType string  `json:"statType"`
	Label    string  `json:"label"`
	Sav      float64 `json:"sav"`
}

type BuildDetail struct {
	Rank      int            `json:"rank"`
	Usage     float64        `json:"usage"`
	Sets      []string       `json:"sets"`
	Verdict   string         `json:"verdict"` // "Worthy", "Sell"
	Landmines []LandmineStat `json:"landmines"`
	Sav       Sav            `json:"sav"` // Include raw SAV for UI display
}

type HeroDetail struct {
	HeroName  string        `json:"heroName"`
	Icon      string        `json:"icon"`
	Rarity    int           `json:"rarity"`
	Role      string        `json:"role"`
	Attribute string        `json:"attribute"`
	Builds    []BuildDetail `json:"builds"`
}

type SuitedBuild struct {
	HeroName  string   `json:"heroName"`
	Rank      int      `json:"rank"`
	Usage     float64  `json:"usage"`
	Sets      []string `json:"sets"`
	Landmines []string `json:"landmines"` // Labels of landmines
	Rarity    int      `json:"rarity"`
	Role      string   `json:"role"`
	Attribute string   `json:"attribute"`
}

type EvaluationResult struct {
	GearID            string        `json:"gearId"`
	Verdict           string        `json:"verdict"`           // "Worthy", "Discard", "Speed Check", "Sell"
	GlobalRuleMatched string        `json:"globalRuleMatched"` // Description if global rule matched
	SuitedBuilds      []SuitedBuild `json:"suitedBuilds"`
	HeroDetails       []HeroDetail  `json:"heroDetails"`
}

var STAT_LABELS = map[string]string{
	"Attack":                  "Attack",
	"AttackPercent":           "Attack %",
	"Health":                  "Health",
	"HealthPercent":           "Health %",
	"Defense":                 "Defense",
	"DefensePercent":          "Defense %",
	"Speed":                   "Speed",
	"CritHitChancePercent":    "Crit Chance %",
	"CritHitDamagePercent":    "Crit Damage %",
	"EffectivenessPercent":    "Effectiveness %",
	"EffectResistancePercent": "Effect Resist %",
}

var database []HeroStats
var heroIcons map[string]string = make(map[string]string)
var heroRarities map[string]int = make(map[string]int)
var heroRoles map[string]string = make(map[string]string)
var heroAttributes map[string]string = make(map[string]string)

func main() {
	dbPath := findDBPath()
	if dbPath == "" {
		log.Fatal("Could not find average_build_stats.json database in any expected path.")
	}
	fmt.Printf("Loading database from: %s\n", dbPath)

	data, err := os.ReadFile(dbPath)
	if err != nil {
		log.Fatalf("Failed to read database: %v", err)
	}

	if err := json.Unmarshal(data, &database); err != nil {
		log.Fatalf("Failed to parse database: %v", err)
	}
	fmt.Printf("Loaded stats for %d heroes.\n", len(database))

	// Load hero icons and filter metadata
	loadHeroIcons()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/evaluate", handleEvaluate)

	handler := corsMiddleware(mux)

	port := ":8081"
	fmt.Printf("Starting E7 Gear Evaluator API Server on http://localhost%s...\n", port)
	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func normalizeName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "")
	name = strings.ReplaceAll(name, "-", "")
	name = strings.ReplaceAll(name, "'", "")
	name = strings.ReplaceAll(name, ".", "")
	return name
}

func loadHeroIcons() {
	statsPath := findStatsPath()
	if statsPath == "" {
		log.Println("Warning: stats.json not found, icons will not be loaded.")
		return
	}

	data, err := os.ReadFile(statsPath)
	if err != nil {
		log.Printf("Warning: Failed to read stats.json: %v\n", err)
		return
	}

	var stats map[string]StatsHero
	if err := json.Unmarshal(data, &stats); err != nil {
		log.Printf("Warning: Failed to parse stats.json: %v\n", err)
		return
	}

	for _, h := range stats {
		normName := normalizeName(h.Name)
		heroIcons[normName] = h.Icon
		heroRarities[normName] = h.Rarity
		heroRoles[normName] = h.Role
		heroAttributes[normName] = h.Attribute
	}
	fmt.Printf("Loaded %d hero icon and metadata mappings.\n", len(heroIcons))
}

func findDBPath() string {
	paths := []string{
		"/home/thesky/Documents/Projects/Tools/e7evaluator-z/v2/v3/data/heroes/average_build_stats.json",
		"data/heroes/average_build_stats.json",
		"../data/heroes/average_build_stats.json",
		"../../data/heroes/average_build_stats.json",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func findStatsPath() string {
	paths := []string{
		"/home/thesky/Documents/Projects/Tools/e7evaluator-z/v2/v3/data/heroes/stats.json",
		"data/heroes/stats.json",
		"../data/heroes/stats.json",
		"../../data/heroes/stats.json",
	}
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getSavKey(statType string) string {
	switch statType {
	case "Attack", "AttackPercent":
		return "atk"
	case "Defense", "DefensePercent":
		return "def"
	case "Health", "HealthPercent":
		return "hp"
	case "Speed":
		return "spd"
	case "CritHitChancePercent":
		return "cc"
	case "CritHitDamagePercent":
		return "cd"
	case "EffectivenessPercent":
		return "eff"
	case "EffectResistancePercent":
		return "res"
	default:
		return ""
	}
}

func isFlatMainStat(statType string) bool {
	return statType == "Attack" || statType == "Health" || statType == "Defense"
}

func isRightSideSlot(slot string) bool {
	return slot == "Necklace" || slot == "Ring" || slot == "Boots"
}

func hasSpeedSubstat(substats []Substat) bool {
	for _, sub := range substats {
		if sub.Type == "Speed" {
			return true
		}
	}
	return false
}

func handleEvaluate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var gear Gear
	if err := json.NewDecoder(r.Body).Decode(&gear); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	result := evaluateGear(gear)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func evaluateGear(gear Gear) EvaluationResult {
	result := EvaluationResult{
		GearID:       gear.ID,
		SuitedBuilds: []SuitedBuild{},
		HeroDetails:  []HeroDetail{},
	}

	// 1. ES Check (ES < 24 -> Discard)
	if gear.Score.ES < 24 {
		result.Verdict = "Discard"
		result.GlobalRuleMatched = fmt.Sprintf("Equipment Score (%d) is lower than 24.", gear.Score.ES)
		return result
	}

	// 2. Right-side Flat no Speed (Ring/Necklace/Boots + Flat Mainstat + No Speed substat -> Discard)
	if isRightSideSlot(gear.Slot) && isFlatMainStat(gear.Main.Type) && !hasSpeedSubstat(gear.Substats) {
		result.Verdict = "Discard"
		result.GlobalRuleMatched = fmt.Sprintf("Right-side %s with flat mainstat %s and no Speed substat.", gear.Slot, STAT_LABELS[gear.Main.Type])
		return result
	}

	// 3. Right-side Flat with Speed (Ring/Necklace/Boots + Flat Mainstat + Speed substat -> Speed Check)
	if isRightSideSlot(gear.Slot) && isFlatMainStat(gear.Main.Type) && hasSpeedSubstat(gear.Substats) {
		result.Verdict = "Speed Check"
		result.GlobalRuleMatched = fmt.Sprintf("Right-side %s with flat mainstat %s and Speed substat present.", gear.Slot, STAT_LABELS[gear.Main.Type])
		return result
	}

	// 4. Evaluate against hero builds
	hasAtLeastOneWorthyBuild := false

	for _, hero := range database {
		normName := normalizeName(hero.Hero)
		iconFile := heroIcons[normName]

		heroDetail := HeroDetail{
			HeroName:  hero.Hero,
			Icon:      iconFile,
			Rarity:    heroRarities[normName],
			Role:      heroRoles[normName],
			Attribute: heroAttributes[normName],
			Builds:    []BuildDetail{},
		}

		for _, build := range hero.Builds {
			landmines := []LandmineStat{}

			// Check substats
			for _, sub := range gear.Substats {
				key := getSavKey(sub.Type)
				if key == "" {
					continue
				}

				savVal := build.Sav.Get(key)
				if savVal < 10.0 {
					landmines = append(landmines, LandmineStat{
						StatType: sub.Type,
						Label:    STAT_LABELS[sub.Type],
						Sav:      savVal,
					})
				}
			}

			// Apply tolerance rules
			var verdict string
			isWorthy := false
			tolerance := 0
			if strings.EqualFold(gear.Rarity, "Heroic") {
				tolerance = 0
			} else { // Epic
				tolerance = 1
			}

			if len(landmines) <= tolerance {
				verdict = "Worthy"
				isWorthy = true
				hasAtLeastOneWorthyBuild = true
			} else {
				verdict = "Sell"
			}

			buildDetail := BuildDetail{
				Rank:      build.Rank,
				Usage:     build.Usage,
				Sets:      build.Sets,
				Verdict:   verdict,
				Landmines: landmines,
				Sav:       build.Sav,
			}
			heroDetail.Builds = append(heroDetail.Builds, buildDetail)

			// If worthy, add to suited builds list
			if isWorthy {
				landmineLabels := make([]string, len(landmines))
				for i, lm := range landmines {
					landmineLabels[i] = lm.Label
				}
				result.SuitedBuilds = append(result.SuitedBuilds, SuitedBuild{
					HeroName:  hero.Hero,
					Rank:      build.Rank,
					Usage:     build.Usage,
					Sets:      build.Sets,
					Landmines: landmineLabels,
					Rarity:    heroRarities[normName],
					Role:      heroRoles[normName],
					Attribute: heroAttributes[normName],
				})
			}
		}

		result.HeroDetails = append(result.HeroDetails, heroDetail)
	}

	// Determine overall verdict
	if hasAtLeastOneWorthyBuild {
		result.Verdict = "Worthy"
	} else {
		result.Verdict = "Sell"
	}

	return result
}
