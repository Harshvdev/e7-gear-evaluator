package main

// ---------------------------------------------------------------------------
// types.go — All data structures used across the evaluator.
//
// Sections:
//   1. Database input types   (parsed from average_build_stats.json / stats.json)
//   2. Gear input types       (received from the generator API)
//   3. API response types     (returned from /api/evaluate)
//   4. New 7-Layer Architecture types (HeroProfile, GearTrace, etc.)
//   5. Shared labels / maps
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

type AverageStats struct {
	Atk float64 `json:"atk"`
	Def float64 `json:"def"`
	Hp  float64 `json:"hp"`
	Spd float64 `json:"spd"`
	Cc  float64 `json:"cc"`
	Cd  float64 `json:"cd"`
	Eff float64 `json:"eff"`
	Res float64 `json:"res"`
	Gs  float64 `json:"gs"`
}

// HeroBuild is one ranked build variant for a hero (top N by usage).
type HeroBuild struct {
	Rank         int          `json:"rank"`
	Usage        float64      `json:"usage"`
	Sets         []string     `json:"sets"`
	AverageStats AverageStats `json:"averageStats"`
	Sav          Sav          `json:"sav"`
}

// HeroStats is a single hero entry from average_build_stats.json.
type HeroStats struct {
	Hero   string      `json:"hero"`
	Builds []HeroBuild `json:"builds"`
}

// StatsHero is a single hero entry from stats.json (icon / class metadata).
type StatsHero struct {
	ID           string           `json:"id"`
	Name         string           `json:"name"`
	Icon         string           `json:"icon"`
	Rarity       int              `json:"rarity"`
	Role         string           `json:"role"`
	Attribute    string           `json:"attribute"`
	SelfDevotion SelfDevotionInfo `json:"self_devotion"`
	BaseStats    HeroBaseStatsSet `json:"base_stats"`
}

type SelfDevotionInfo struct {
	Type   string             `json:"type"`
	Grades map[string]float64 `json:"grades"`
}

type HeroBaseStatsSet struct {
	Lv50 HeroBaseStats `json:"lv50"`
	Lv60 HeroBaseStats `json:"lv60"`
}

type HeroBaseStats struct {
	Atk float64 `json:"atk"`
	Def float64 `json:"def"`
	Hp  float64 `json:"hp"`
	Spd float64 `json:"spd"`
	Cc  float64 `json:"chc"` // Note: stats.json uses chc/chd
	Cd  float64 `json:"chd"`
	Eff float64 `json:"eff"`
	Res float64 `json:"efr"` // Note: stats.json uses efr
}

// ---------------------------------------------------------------------------
// 2. Gear input types
// ---------------------------------------------------------------------------

// Substat is one of up to four secondary stats on a gear piece.
type Substat struct {
	Type     string  `json:"type"`
	Value    float64 `json:"value"`
	Rolls    int     `json:"rolls"`
	Modified bool    `json:"modified,omitempty"`
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
	Reforged bool      `json:"reforged,omitempty"`
}

type EvaluateRequest struct {
	Gear           Gear                           `json:"gear"`
	ExcludedBuilds map[string][]int               `json:"excludedBuilds"` // HeroName -> list of excluded build ranks (1-indexed)
	SpeedCheck     string                         `json:"speedCheck,omitempty"`     // "ON" | "OFF" (defaults to ON)
	ModBudget      string                         `json:"modBudget,omitempty"`      // "none" | "surplus" (default "none")
	ReforgeBudget  string                         `json:"reforgeBudget,omitempty"`  // "none" | "surplus" (default "none")
	CustomProfiles map[string]map[int]HeroProfile `json:"customProfiles,omitempty"` // HeroName -> build rank -> custom profile configurations
}

// ---------------------------------------------------------------------------
// 3. API response types
// ---------------------------------------------------------------------------

// EvaluationResponse is the top-level response from /api/evaluate.
type EvaluationResponse struct {
	GearID       string          `json:"gearId"`
	Verdict      string          `json:"verdict"` // Worthy, Sell, Discard, Speed Check
	SuitedBuilds []L1SuitedBuild `json:"suitedBuilds"`
	HeroDetails  []L1HeroDetail  `json:"heroDetails"`

	// New 7-Layer Architecture trace & layers
	Trace  *GearTrace    `json:"trace,omitempty"`
	Layer1 *Layer1Result `json:"layer1,omitempty"` // kept for structure similarity
}

// --- Legacy UI compatibility types ---

// Layer1Result is the output of the Layer 1 hero-suitability evaluation.
type Layer1Result struct {
	Verdict           string          `json:"verdict"`
	GlobalRuleMatched string          `json:"globalRuleMatched,omitempty"`
	SuitedBuilds      []L1SuitedBuild `json:"suitedBuilds"`
	HeroDetails       []L1HeroDetail  `json:"heroDetails"`
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

// L1BuildDetail is the per-build scoring result.
type L1BuildDetail struct {
	Rank        int          `json:"rank"`
	Usage       float64      `json:"usage"`
	Sets        []string     `json:"sets"`
	Verdict     string       `json:"verdict"` // "Worthy" or "Sell"
	WAS         float64      `json:"was"`     // raw Weighted Alignment Score (0–max_weight)
	WASPct      float64      `json:"wasPct"`  // WAS as % of threshold; >100 = passes
	Threshold   float64      `json:"threshold"`
	MissingCore int          `json:"missingCore"`
	Landmines   []L1Landmine `json:"landmines"`
	Sav         Sav          `json:"sav"`
}

// L1Landmine is a gear substat that is low-priority for a specific hero build.
type L1Landmine struct {
	StatType string  `json:"statType"`
	Label    string  `json:"label"`
	Weight   float64 `json:"weight"`
}

// ---------------------------------------------------------------------------
// 4. New 7-Layer Architecture types
// ---------------------------------------------------------------------------

// PRIORITY_WEIGHTS maps the 9-position priority scale (-3..+5) to signed weights per Q-E7-Architecture §2.2
var PRIORITY_WEIGHTS = map[int]float64{
	5:  6.0,  // essential: defining stat
	4:  3.5,  // core: must be present & rolled well
	3:  2.0,  // wanted: strong positive
	2:  1.0,  // useful: counts normally
	1:  0.4,  // filler: marginal positive
	0:  0.0,  // neutral: ignored
	-1: -1.0, // dead: wasted slot
	-2: -2.5, // harmful: actively bad
	-3: -4.0, // poison: carrying it makes piece worse
}

// ABSENCE_COST penalizes a piece for omitting a slot-possible core/essential stat per Q-E7-Architecture §3.2
var ABSENCE_COST = map[int]float64{
	5: 18.0, // missing essential (+5) stat
	4: 10.0, // missing core (+4) stat
}

type StatBounds struct {
	Min *float64 `json:"min"`
	Max *float64 `json:"max"`
}

type MinQuality struct {
	Score      *float64 `json:"score"`
	Efficiency *float64 `json:"efficiency"`
}

// HeroProfile defines the rules/priorities used for evaluating a gear piece for a hero build.
type HeroProfile struct {
	HeroID         string                `json:"heroId"`
	HeroName       string                `json:"heroName"`
	BuildRank      int                   `json:"buildRank"`
	Selected       bool                  `json:"selected"`
	RosterTier     string                `json:"rosterTier,omitempty"`    // "primary" | "bench" | "catalog" (default "primary")
	RiskTolerance  float64               `json:"riskTolerance,omitempty"` // 0.0 .. 1.0 (default 0.5)
	StatRanges     map[string]StatBounds `json:"statRanges"`
	Sets           [][]string            `json:"sets"` // list of acceptable set combinations
	Priorities     map[string]int        `json:"priorities"`
	MinQuality     *MinQuality           `json:"minQuality"`
	WeightMode     string                `json:"weightMode"` // "strict" | "weighted"
	AccessoryMains []string              `json:"accessoryMains"`
	OriginalSav    Sav                   `json:"-"`
}

// Traces per layer
type L0Trace struct {
	ParseConfidence    float64        `json:"parseConfidence"`
	RollReconstruction map[string]int `json:"rollReconstruction"` // statType -> reconstructed rolls
	Ambiguities        []string       `json:"ambiguities"`
}

type L1Trace struct {
	RulesRun   []string `json:"rulesRun"`
	Violations []string `json:"violations"`
}

type L2Trace struct {
	Rule   string `json:"rule"`
	Fired  bool   `json:"fired"`
	Detail string `json:"detail"`
}

type L3Trace struct {
	Toggle     string  `json:"toggle"` // "ON" | "OFF"
	Tagged     bool    `json:"tagged"` // tagged for SPEED_VAULT
	SpeedValue float64 `json:"speedValue"`
}

type L4HeroGateResult struct {
	HeroName       string  `json:"heroName"`
	BuildRank      int     `json:"buildRank"`
	RosterTier     string  `json:"rosterTier"`
	GateResults    [4]bool `json:"gateResults"` // index 0: Set, 1: Main, 2: Range, 3: Fit Score
	FitScore       float64 `json:"fitScore"`
	RawFit         float64 `json:"rawFit"`
	MaxFit         float64 `json:"maxFit"`
	AbsencePenalty float64 `json:"absencePenalty"`
	FitPct         float64 `json:"fitPct"`   // normalized signed fit% in [-100, +100]
	FitClass       string  `json:"fitClass"` // "CORE", "USABLE", "MARGINAL", "REJECT"
	Threshold      float64 `json:"threshold"`
	Pass           bool    `json:"pass"`
}

type L4Trace struct {
	PerHero []L4HeroGateResult `json:"perHero"`
}

type ProjectionScenario struct {
	ScenarioName string             `json:"scenarioName"` // "expected", "best", "worst"
	Substats     []Substat          `json:"substats"`     // projected substats at +15
	Reforged     []Substat          `json:"reforged"`     // projected reforge values
	Score        float64            `json:"score"`        // WSS score
	HeroScores   map[string]float64 `json:"heroScores"`   // hero_rank -> fit score
	HeroFitPcts  map[string]float64 `json:"heroFitPcts"`  // hero_rank -> signed fit%
}

// StopCard is the enhancement depth decision emitted by the Enhancement Controller (§5)
type StopCard struct {
	EnhanceAtPoint      int     `json:"enhanceAtPoint"`
	ObservedFitPct      float64 `json:"observedFitPct"`
	PGood               float64 `json:"pGood"`
	EVFinal             float64 `json:"evFinal"`
	Recommended         string  `json:"recommended"` // "CONTINUE" | "STOP"
	Reason              string  `json:"reason,omitempty"`
	RollSequenceForCore string  `json:"rollSequenceForCore,omitempty"`
}

type L5Trace struct {
	CurrentWSS        float64              `json:"currentWss"`
	CurrentHeroScore  map[string]float64   `json:"currentHeroScore"`  // hero_rank -> fit score
	CurrentHeroFitPct map[string]float64   `json:"currentHeroFitPct"` // hero_rank -> fit%
	PCore             float64              `json:"pCore"`
	PUsable           float64              `json:"pUsable"`
	PMarginal         float64              `json:"pMarginal"`
	PReject           float64              `json:"pReject"`
	PGood             float64              `json:"pGood"`
	EVFinal           float64              `json:"evFinal"`
	StopCard          *StopCard            `json:"stopCard,omitempty"`
	Scenarios         []ProjectionScenario `json:"scenarios"`
}

type SalvagePlan struct {
	DeadSubStat    string  `json:"deadSubStat"`
	TargetStat     string  `json:"targetStat"`
	ExpectedValue  float64 `json:"expectedValue"`
	RescoredFit    float64 `json:"rescoredFit"`
	RescoredFitPct float64 `json:"rescoredFitPct"`
}

type L6Trace struct {
	Verdict       string       `json:"verdict"` // KEEP_ENHANCE, KEEP_MARGINAL, SALVAGE_MOD, REFORGE_TAG, SPEED_VAULT, SELL_EXTRACT
	WinnerHero    string       `json:"winnerHero"`
	WinnerBuild   int          `json:"winnerBuild"`
	RunnerUps     []string     `json:"runnerUps"` // list of "Hero_Name (Build X)"
	SalvageDetail *SalvagePlan `json:"salvageDetail,omitempty"`
	ModBudget     string       `json:"modBudget"`
	ReforgeBudget string       `json:"reforgeBudget"`
}

type GearTrace struct {
	GearID          string    `json:"gearId"`
	TableVersion    string    `json:"tableVersion"`
	SpeedCheckState string    `json:"speedCheckState"`
	L0              L0Trace   `json:"l0"`
	L1              L1Trace   `json:"l1"`
	L2              L2Trace   `json:"l2"`
	L3              L3Trace   `json:"l3"`
	L4              L4Trace   `json:"l4"`
	L5              L5Trace   `json:"l5"`
	L6              L6Trace   `json:"l6"`
	Verdict         string    `json:"verdict"`
	ReasonCodes     []string  `json:"reasonCodes"`
}

// ---------------------------------------------------------------------------
// 5. Shared labels
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
