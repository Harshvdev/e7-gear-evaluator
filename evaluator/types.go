package main

// ---------------------------------------------------------------------------
// types.go — All data structures used across the evaluator.
//
// Sections:
//   1. Database input types   (parsed from average_build_stats.json / stats.json)
//   2. Gear input types       (received from the generator API)
//   3. API response types     (returned from /api/evaluate)
//   4. Shared labels / maps
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// 1. Database input types
// ---------------------------------------------------------------------------

// Sav holds the per-stat "Substat Average Value" for a hero build.
// Each value represents how many rolls worth of that stat an average player
// accumulates across all six gear pieces in that build.
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

// HeroBuild is one ranked build variant for a hero (top N by usage).
type HeroBuild struct {
	Rank  int      `json:"rank"`
	Usage float64  `json:"usage"`
	Sets  []string `json:"sets"`
	Sav   Sav      `json:"sav"`
}

// HeroStats is a single hero entry from average_build_stats.json.
type HeroStats struct {
	Hero   string      `json:"hero"`
	Builds []HeroBuild `json:"builds"`
}

// StatsHero is a single hero entry from stats.json (icon / class metadata).
type StatsHero struct {
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	Rarity    int    `json:"rarity"`
	Role      string `json:"role"`
	Attribute string `json:"attribute"`
}

// ---------------------------------------------------------------------------
// 2. Gear input types
// ---------------------------------------------------------------------------

// Substat is one of up to four secondary stats on a gear piece.
type Substat struct {
	Type  string  `json:"type"`
	Value float64 `json:"value"`
	Rolls int     `json:"rolls"`
}

// MainStat is the primary stat of the gear piece.
type MainStat struct {
	Type  string  `json:"type"`
	Value float64 `json:"value"`
}

// GearScore holds the pre-computed GS and ES from the generator.
type GearScore struct {
	GS float64 `json:"gs"`
	ES int     `json:"es"`
}

// Gear is the full gear piece as received from /api/generate or /api/enhance.
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

// ---------------------------------------------------------------------------
// 3. API response types
// ---------------------------------------------------------------------------

// EvaluationResponse is the top-level response from /api/evaluate.
// It is versioned by layer so future layers can be added without breaking
// existing consumers — Layer 2 fields can simply be appended here.
type EvaluationResponse struct {
	GearID string `json:"gearId"`
	*Layer1Result
	Layer1 *Layer1Result `json:"layer1"`
	Layer2 *Layer2Result `json:"layer2,omitempty"`
}

// --- Layer 1 types ---

// Layer1Result is the output of the Layer 1 hero-suitability evaluation.
// It answers: "Which heroes is this piece worth for?"
type Layer1Result struct {
	// Verdict is the overall decision for this gear piece.
	//   "Worthy"      — gear suits at least one hero build
	//   "Sell"        — gear suits no hero builds
	//   "Discard"     — gear failed a global pre-filter rule
	//   "Speed Check" — right-side flat-main piece with speed; decide manually
	Verdict string `json:"verdict"`

	// GlobalRuleMatched is set when a pre-filter rule short-circuited evaluation.
	// Empty string when the full hero-by-hero scoring was performed.
	GlobalRuleMatched string `json:"globalRuleMatched,omitempty"`

	// SuitedBuilds is a flat list of every (hero, build) pair that passed.
	// Used by the renderer for the compact "suited heroes" tab.
	SuitedBuilds []L1SuitedBuild `json:"suitedBuilds"`

	// HeroDetails contains per-hero, per-build scoring detail for the detail view.
	HeroDetails []L1HeroDetail `json:"heroDetails"`
}

// L1SuitedBuild is a compact record of one passing (hero, build) pair.
type L1SuitedBuild struct {
	HeroName  string   `json:"heroName"`
	Rank      int      `json:"rank"`
	Usage     float64  `json:"usage"`
	Sets      []string `json:"sets"`
	Landmines []string `json:"landmines"` // human-readable labels of low-weight substats
	Rarity    int      `json:"rarity"`
	Role      string   `json:"role"`
	Attribute string   `json:"attribute"`
}

// L1HeroDetail is the full scoring breakdown for one hero across all their builds.
type L1HeroDetail struct {
	HeroName  string          `json:"heroName"`
	Icon      string          `json:"icon"`
	Rarity    int             `json:"rarity"`
	Role      string          `json:"role"`
	Attribute string          `json:"attribute"`
	Builds    []L1BuildDetail `json:"builds"`
}

// L1BuildDetail is the per-build scoring result, surfacing the WAS metrics
// so the UI can display a progress bar and explain the verdict.
type L1BuildDetail struct {
	Rank      int            `json:"rank"`
	Usage     float64        `json:"usage"`
	Sets      []string       `json:"sets"`
	Verdict   string         `json:"verdict"`   // "Worthy" or "Sell"
	WAS       float64        `json:"was"`        // raw Weighted Alignment Score (0–max_weight)
	WASPct    float64        `json:"wasPct"`     // WAS as % of threshold; >100 = passes
	Threshold float64        `json:"threshold"`  // dynamic pass threshold for this build
	MissingCore int          `json:"missingCore"` // count of core stats absent from gear
	Landmines []L1Landmine   `json:"landmines"`
	Sav       Sav            `json:"sav"` // raw SAV kept for the stat-grid UI
}

// L1Landmine is a gear substat that is low-priority for a specific hero build.
// Displayed in the UI to explain why a stat "doesn't count" toward suitability.
type L1Landmine struct {
	StatType string  `json:"statType"` // gear stat type key
	Label    string  `json:"label"`    // human-readable name
	Weight   float64 `json:"weight"`   // hero's normalized priority weight (%) for this stat
}

// ---------------------------------------------------------------------------
// 4. Shared labels
// ---------------------------------------------------------------------------

// STAT_LABELS maps internal stat type keys to human-readable display names.
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
