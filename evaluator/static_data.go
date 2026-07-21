package main

import "strings"

// ---------------------------------------------------------------------------

// static_data.go — All immutable game data, roll ranges, and probabilities.
//
// Source: Fribbels E7 Optimizer dataset and geardata.md rules.
// ---------------------------------------------------------------------------

var (
	// MainStatsBySlot maps slot -> legal main stats
	MainStatsBySlot = map[string]map[string]bool{
		"Weapon":   {"Attack": true},
		"Helmet":   {"Health": true},
		"Armor":    {"Defense": true},
		"Necklace": {"AttackPercent": true, "DefensePercent": true, "HealthPercent": true, "CritHitChancePercent": true, "CritHitDamagePercent": true, "Health": true, "Defense": true, "Attack": true},
		"Ring":     {"AttackPercent": true, "DefensePercent": true, "HealthPercent": true, "EffectivenessPercent": true, "EffectResistancePercent": true, "Health": true, "Defense": true, "Attack": true},
		"Boots":    {"AttackPercent": true, "DefensePercent": true, "HealthPercent": true, "Speed": true, "Health": true, "Defense": true, "Attack": true},
	}

	// SubstatsBySlot maps slot -> allowed substats
	SubstatsBySlot = map[string]map[string]bool{
		"Weapon": {
			"AttackPercent": true, "Health": true, "HealthPercent": true, "Speed": true,
			"CritHitChancePercent": true, "CritHitDamagePercent": true, "EffectivenessPercent": true, "EffectResistancePercent": true,
		},
		"Helmet": {
			"Attack": true, "AttackPercent": true, "Defense": true, "DefensePercent": true, "HealthPercent": true, "Speed": true,
			"CritHitChancePercent": true, "CritHitDamagePercent": true, "EffectivenessPercent": true, "EffectResistancePercent": true,
		},
		"Armor": {
			"DefensePercent": true, "Health": true, "HealthPercent": true, "Speed": true,
			"CritHitChancePercent": true, "CritHitDamagePercent": true, "EffectivenessPercent": true, "EffectResistancePercent": true,
		},
		"Necklace": {
			"Attack": true, "AttackPercent": true, "Defense": true, "DefensePercent": true, "Health": true, "HealthPercent": true, "Speed": true,
			"CritHitChancePercent": true, "CritHitDamagePercent": true, "EffectivenessPercent": true, "EffectResistancePercent": true,
		},
		"Ring": {
			"Attack": true, "AttackPercent": true, "Defense": true, "DefensePercent": true, "Health": true, "HealthPercent": true, "Speed": true,
			"CritHitChancePercent": true, "CritHitDamagePercent": true, "EffectivenessPercent": true, "EffectResistancePercent": true,
		},
		"Boots": {
			"Attack": true, "AttackPercent": true, "Defense": true, "DefensePercent": true, "Health": true, "HealthPercent": true, "Speed": true,
			"CritHitChancePercent": true, "CritHitDamagePercent": true, "EffectivenessPercent": true, "EffectResistancePercent": true,
		},
	}

	// WSSCoefficients are the weights for each stat to calculate Gear Score (GS)
	WSSCoefficients = map[string]float64{
		"AttackPercent":           1.0,
		"DefensePercent":          1.0,
		"HealthPercent":           1.0,
		"EffectivenessPercent":    1.0,
		"EffectResistancePercent": 1.0,
		"Speed":                   2.0,
		"CritHitChancePercent":    1.6,
		"CritHitDamagePercent":    8.0 / 7.0, // CDmg is 1.142857 (8/7)
		"Attack":                  0.088718,  // Flat Attack
		"Defense":                 0.160968,  // Flat Defense
		"Health":                  0.017759,  // Flat Health
	}

	// MainStatValuesAtPlus15 maps level -> stat type -> value
	MainStatValuesAtPlus15 = map[int]map[string]float64{
		85: {
			"Attack":                  525,
			"Health":                  2835,
			"Defense":                 310,
			"AttackPercent":           65,
			"DefensePercent":          65,
			"HealthPercent":           65,
			"CritHitChancePercent":    60,
			"CritHitDamagePercent":    70,
			"EffectivenessPercent":    65,
			"EffectResistancePercent": 65,
			"Speed":                   45,
		},
		88: {
			"Attack":                  550,
			"Health":                  2970,
			"Defense":                 320,
			"AttackPercent":           70,
			"DefensePercent":          70,
			"HealthPercent":           70,
			"CritHitChancePercent":    65,
			"CritHitDamagePercent":    75,
			"EffectivenessPercent":    70,
			"EffectResistancePercent": 70,
			"Speed":                   50,
		},
	}

	// MainStatProgressionMaps maps level -> enhancement level (0..15) -> stat type -> value
	// We will fill this in an init block to make the code simpler.
	MainStatProgression = make(map[int]map[int]map[string]float64)

	// SubstatRollRanges maps level -> rarity -> statType -> [min, max]
	SubstatRollRanges = map[int]map[string]map[string][2]float64{
		85: {
			"Heroic": {
				"AttackPercent":           {4, 8},
				"DefensePercent":          {4, 8},
				"HealthPercent":           {4, 8},
				"EffectivenessPercent":    {4, 8},
				"EffectResistancePercent": {4, 8},
				"CritHitChancePercent":    {3, 5},
				"CritHitDamagePercent":    {4, 7},
				"Speed":                   {1, 4},
				"Attack":                  {31, 44},
				"Defense":                 {26, 33},
				"Health":                  {149, 192},
			},
			"Epic": {
				"AttackPercent":           {4, 8},
				"DefensePercent":          {4, 8},
				"HealthPercent":           {4, 8},
				"EffectivenessPercent":    {4, 8},
				"EffectResistancePercent": {4, 8},
				"CritHitChancePercent":    {3, 5},
				"CritHitDamagePercent":    {4, 7},
				"Speed":                   {2, 4},
				"Attack":                  {33, 46},
				"Defense":                 {28, 35},
				"Health":                  {157, 202},
			},
		},
		88: {
			"Heroic": {
				"AttackPercent":           {5, 9},
				"DefensePercent":          {5, 9},
				"HealthPercent":           {5, 9},
				"EffectivenessPercent":    {5, 9},
				"EffectResistancePercent": {5, 9},
				"CritHitChancePercent":    {3, 6},
				"CritHitDamagePercent":    {4, 8},
				"Speed":                   {2, 4},
				"Attack":                  {36, 50},
				"Defense":                 {30, 38},
				"Health":                  {169, 218},
			},
			"Epic": {
				"AttackPercent":           {5, 9},
				"DefensePercent":          {5, 9},
				"HealthPercent":           {5, 9},
				"EffectivenessPercent":    {5, 9},
				"EffectResistancePercent": {5, 9},
				"CritHitChancePercent":    {3, 6},
				"CritHitDamagePercent":    {4, 8},
				"Speed":                   {3, 5},
				"Attack":                  {37, 53},
				"Defense":                 {32, 40},
				"Health":                  {178, 229},
			},
		},
	}

	// Reforge increments per roll count (1..6 rolls) for Level 85 to 90
	ReforgeIncrements = map[string][]float64{
		"AttackPercent":           {0, 1, 3, 4, 5, 7, 8},
		"DefensePercent":          {0, 1, 3, 4, 5, 7, 8},
		"HealthPercent":           {0, 1, 3, 4, 5, 7, 8},
		"EffectivenessPercent":    {0, 1, 3, 4, 5, 7, 8},
		"EffectResistancePercent": {0, 1, 3, 4, 5, 7, 8},
		"CritHitChancePercent":    {0, 1, 2, 3, 4, 5, 6},
		"CritHitDamagePercent":    {0, 1, 2, 3, 4, 6, 7},
		"Speed":                   {0, 0, 1, 2, 3, 4, 4},
		"Attack":                  {0, 11, 22, 33, 44, 55, 66},
		"Defense":                 {0, 9, 18, 27, 36, 45, 54},
		"Health":                  {0, 56, 112, 168, 224, 280, 336},
	}

	// Main stat reforge bonuses (85 -> 90)
	MainStatReforgeBonuses = map[string]float64{
		"Attack":                  25,
		"Health":                  135,
		"Defense":                 10,
		"AttackPercent":           5,
		"DefensePercent":          5,
		"HealthPercent":           5,
		"CritHitChancePercent":    5,
		"CritHitDamagePercent":    5,
		"EffectivenessPercent":    5,
		"EffectResistancePercent": 5,
		"Speed":                   5,
	}

	// Roll value probabilities for Level 85 and 88.
	// Used for expected roll calculation.
	// For uniform distributions, the average is simply (min+max)/2.
	// For speed, we have specific weight distributions.
	SpeedRollAverages = map[int]map[string]float64{
		85: {
			"Epic":   (2.0*0.3322 + 3.0*0.3322 + 4.0*0.3322 + 5.0*0.0033) / (0.3322 + 0.3322 + 0.3322 + 0.0033), // ~3.0
			"Heroic": (1.0*0.0383 + 2.0*0.3484 + 3.0*0.3484 + 4.0*0.2648),                                       // ~2.84
		},
		88: {
			"Epic":   (3.0*0.4975 + 4.0*0.4975 + 5.0*0.0050),            // ~3.5
			"Heroic": (2.0*0.0833 + 3.0*0.5208 + 4.0*0.3958),            // ~3.31
		},
	}
)

func init() {
	// Initialize main stat progression levels
	MainStatProgression[85] = map[int]map[string]float64{
		0:  {"Attack": 100, "Health": 540, "Defense": 60, "AttackPercent": 12, "HealthPercent": 12, "DefensePercent": 12, "CritHitChancePercent": 11, "CritHitDamagePercent": 13, "EffectivenessPercent": 12, "EffectResistancePercent": 12, "Speed": 8},
		3:  {"Attack": 150, "Health": 810, "Defense": 90, "AttackPercent": 18, "HealthPercent": 18, "DefensePercent": 18, "CritHitChancePercent": 15, "CritHitDamagePercent": 19, "EffectivenessPercent": 18, "EffectResistancePercent": 18, "Speed": 12},
		6:  {"Attack": 200, "Health": 1080, "Defense": 120, "AttackPercent": 24, "HealthPercent": 24, "DefensePercent": 24, "CritHitChancePercent": 20, "CritHitDamagePercent": 25, "EffectivenessPercent": 24, "EffectResistancePercent": 24, "Speed": 16},
		9:  {"Attack": 250, "Health": 1350, "Defense": 150, "AttackPercent": 30, "HealthPercent": 30, "DefensePercent": 30, "CritHitChancePercent": 25, "CritHitDamagePercent": 31, "EffectivenessPercent": 30, "EffectResistancePercent": 30, "Speed": 20},
		12: {"Attack": 325, "Health": 1755, "Defense": 195, "AttackPercent": 39, "HealthPercent": 39, "DefensePercent": 39, "CritHitChancePercent": 33, "CritHitDamagePercent": 40, "EffectivenessPercent": 39, "EffectResistancePercent": 39, "Speed": 26},
		15: {"Attack": 525, "Health": 2835, "Defense": 310, "AttackPercent": 65, "HealthPercent": 65, "DefensePercent": 65, "CritHitChancePercent": 60, "CritHitDamagePercent": 70, "EffectivenessPercent": 65, "EffectResistancePercent": 65, "Speed": 45},
	}
	MainStatProgression[88] = map[int]map[string]float64{
		0:  {"Attack": 103, "Health": 553, "Defense": 62, "AttackPercent": 13, "HealthPercent": 13, "DefensePercent": 13, "CritHitChancePercent": 12, "CritHitDamagePercent": 14, "EffectivenessPercent": 13, "EffectResistancePercent": 13, "Speed": 9},
		3:  {"Attack": 154, "Health": 829, "Defense": 93, "AttackPercent": 19, "HealthPercent": 19, "DefensePercent": 19, "CritHitChancePercent": 18, "CritHitDamagePercent": 21, "EffectivenessPercent": 19, "EffectResistancePercent": 19, "Speed": 13},
		6:  {"Attack": 206, "Health": 1105, "Defense": 124, "AttackPercent": 25, "HealthPercent": 25, "DefensePercent": 25, "CritHitChancePercent": 24, "CritHitDamagePercent": 28, "EffectivenessPercent": 25, "EffectResistancePercent": 25, "Speed": 17},
		9:  {"Attack": 258, "Health": 1381, "Defense": 155, "AttackPercent": 31, "HealthPercent": 31, "DefensePercent": 31, "CritHitChancePercent": 30, "CritHitDamagePercent": 35, "EffectivenessPercent": 31, "EffectResistancePercent": 31, "Speed": 21},
		12: {"Attack": 336, "Health": 1795, "Defense": 202, "AttackPercent": 40, "HealthPercent": 40, "DefensePercent": 40, "CritHitChancePercent": 36, "CritHitDamagePercent": 45, "EffectivenessPercent": 40, "EffectResistancePercent": 40, "Speed": 28},
		15: {"Attack": 550, "Health": 2970, "Defense": 320, "AttackPercent": 70, "HealthPercent": 70, "DefensePercent": 70, "CritHitChancePercent": 65, "CritHitDamagePercent": 75, "EffectivenessPercent": 70, "EffectResistancePercent": 70, "Speed": 50},
	}
}

// GetAverageRoll returns the mathematical expected roll value for a stat type.
func GetAverageRoll(level int, rarity string, statType string) float64 {
	// Special Speed values
	if statType == "Speed" {
		if lvlMap, ok := SpeedRollAverages[level]; ok {
			if avg, ok := lvlMap[rarity]; ok {
				return avg
			}
		}
		return 3.0
	}

	// Range-based stats (uniform distribution average)
	ranges, ok := SubstatRollRanges[level]
	if !ok {
		return 0
	}
	rMap, ok := ranges[rarity]
	if !ok {
		return 0
	}
	r, ok := rMap[statType]
	if !ok {
		return 0
	}

	return (r[0] + r[1]) / 2.0
}

// GetLegalSubstats returns the list of allowed substats for a slot, excluding main stat and existing substats.
func GetLegalSubstats(slot string, mainType string, existingSubTypes []string) []string {
	allowed, ok := SubstatsBySlot[slot]
	if !ok {
		return []string{}
	}

	normMain := NormalizeStatType(mainType)
	existingSet := make(map[string]bool)
	for _, sub := range existingSubTypes {
		existingSet[NormalizeStatType(sub)] = true
	}

	var legal []string
	for stat := range allowed {
		normStat := NormalizeStatType(stat)
		if normStat == normMain {
			continue
		}
		if existingSet[normStat] {
			continue
		}
		legal = append(legal, stat)
	}
	return legal
}

// NormalizeStatType standardizes stat names to internal canonical keys (e.g. "AttackPercent", "Speed").
func NormalizeStatType(s string) string {
	sLower := strings.ToLower(strings.TrimSpace(s))
	sClean := strings.ReplaceAll(strings.ReplaceAll(sLower, " ", ""), "_", "")

	switch sClean {
	case "attack", "atk", "flatattack", "flatatk":
		return "Attack"
	case "attackpercent", "atkpercent", "attack%", "atk%", "percentattack":
		return "AttackPercent"
	case "defense", "def", "flatdefense", "flatdef":
		return "Defense"
	case "defensepercent", "defpercent", "defense%", "def%", "percentdefense":
		return "DefensePercent"
	case "health", "hp", "flathealth", "flathp":
		return "Health"
	case "healthpercent", "hppercent", "health%", "hp%", "percenthealth":
		return "HealthPercent"
	case "speed", "spd":
		return "Speed"
	case "crithitchancepercent", "critchancepercent", "critchance%", "crit%", "cc", "chc":
		return "CritHitChancePercent"
	case "crithitdamagepercent", "critdamagepercent", "critdamage%", "cdmg%", "cd", "cdmg", "chd":
		return "CritHitDamagePercent"
	case "effectivenesspercent", "effectiveness%", "eff%", "eff":
		return "EffectivenessPercent"
	case "effectresistancepercent", "effectresistpercent", "effectresistance%", "effectresist%", "res%", "res", "efr":
		return "EffectResistancePercent"
	default:
		return s
	}
}

