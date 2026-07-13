package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// ---------------------------------------------------------------------------
// layer1.go — Layer 1 Evaluation (Hero Suitability & WAS Scoring)
//
// Layer 1 determines which hero build(s) a gear piece is suitable for by
// computing a Weighted Alignment Score (WAS) against every build in the database.
// ---------------------------------------------------------------------------

// Constants for Layer 1 scoring
var GS_WEIGHTS = map[string]float64{
	"AttackPercent":           1.0,
	"DefensePercent":          1.0,
	"HealthPercent":           1.0,
	"EffectivenessPercent":    1.0,
	"EffectResistancePercent": 1.0,
	"Speed":                   2.0,
	"CritHitChancePercent":    1.6,
	"CritHitDamagePercent":    8.0 / 7.0,
	"Attack":                  0.088718,
	"Defense":                 0.160968,
	"Health":                  0.017759,
}

const (
	LANDMINE_WEIGHT_THRESHOLD = 0.08
	CORE_COVERAGE_TARGET      = 0.80
	WAS_THRESHOLD_FACTOR      = 0.45
	MISSING_CORE_PENALTY      = 0.15
	HEROIC_THRESHOLD_DISCOUNT = 0.85
)

// EvaluateLayer1 runs the full Layer 1 suitability evaluation for a gear piece.
func EvaluateLayer1(gear Gear) *Layer1Result {
	result := &Layer1Result{
		SuitedBuilds: []L1SuitedBuild{},
		HeroDetails:  []L1HeroDetail{},
	}

	// --- Pre-Filter Rule 1: ES Check (ES < 24 → Discard) ---
	if gear.Score.ES < 24 {
		result.Verdict = "Discard"
		result.GlobalRuleMatched = fmt.Sprintf("Equipment Score (%d) is lower than 24.", gear.Score.ES)
		return result
	}

	// --- Pre-Filter Rule 2: Right-side Flat with no Speed → Discard ---
	if isRightSideSlot(gear.Slot) && isFlatMainStat(gear.Main.Type) && !hasSpeedSubstat(gear.Substats) {
		result.Verdict = "Discard"
		result.GlobalRuleMatched = fmt.Sprintf(
			"Right-side %s with flat mainstat %s and no Speed substat.",
			gear.Slot, STAT_LABELS[gear.Main.Type],
		)
		return result
	}

	// --- Pre-Filter Rule 3: Right-side Flat with Speed → Speed Check ---
	if isRightSideSlot(gear.Slot) && isFlatMainStat(gear.Main.Type) && hasSpeedSubstat(gear.Substats) {
		result.Verdict = "Speed Check"
		result.GlobalRuleMatched = fmt.Sprintf(
			"Right-side %s with flat mainstat %s and Speed substat present.",
			gear.Slot, STAT_LABELS[gear.Main.Type],
		)
		return result
	}

	// --- Full WAS Evaluation against hero database ---
	hasAtLeastOneWorthyBuild := false

	for _, hero := range heroDatabase {
		normName := normalizeName(hero.Hero)
		iconFile := heroIcons[normName]

		heroDetail := L1HeroDetail{
			HeroName:  hero.Hero,
			Icon:      iconFile,
			Rarity:    heroRarities[normName],
			Role:      heroRoles[normName],
			Attribute: heroAttributes[normName],
			Builds:    []L1BuildDetail{},
		}

		for _, build := range hero.Builds {
			heroWeights := computeHeroWeights(build.Sav)

			wasFinal, threshold, missingCore, _ := scoreBuildAlignment(gear, build.Sav)

			isWorthy := wasFinal >= threshold
			verdict := "Sell"
			if isWorthy {
				verdict = "Worthy"
				hasAtLeastOneWorthyBuild = true
			}

			wasPct := 0.0
			if threshold > 0 {
				wasPct = math.Round((wasFinal/threshold)*1000) / 10
			}

			landmines := []L1Landmine{}
			for _, sub := range gear.Substats {
				key := getSavKey(sub.Type)
				if key == "" {
					continue
				}
				w := heroWeights[key]
				if w < LANDMINE_WEIGHT_THRESHOLD {
					landmines = append(landmines, L1Landmine{
						StatType: sub.Type,
						Label:    STAT_LABELS[sub.Type],
						Weight:   math.Round(w*10000) / 100,
					})
				}
			}

			buildDetail := L1BuildDetail{
				Rank:        build.Rank,
				Usage:       build.Usage,
				Sets:        build.Sets,
				Verdict:     verdict,
				WAS:         math.Round(wasFinal*10000) / 10000,
				WASPct:      wasPct,
				Threshold:   math.Round(threshold*10000) / 10000,
				MissingCore: missingCore,
				Landmines:   landmines,
				Sav:         build.Sav,
			}
			heroDetail.Builds = append(heroDetail.Builds, buildDetail)

			if isWorthy {
				landmineLabels := make([]string, len(landmines))
				for i, lm := range landmines {
					landmineLabels[i] = lm.Label
				}
				result.SuitedBuilds = append(result.SuitedBuilds, L1SuitedBuild{
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

	if hasAtLeastOneWorthyBuild {
		result.Verdict = "Worthy"
	} else {
		result.Verdict = "Sell"
	}

	return result
}

// ---------------------------------------------------------------------------
// WAS calculation helpers
// ---------------------------------------------------------------------------

func computeHeroWeights(sav Sav) map[string]float64 {
	total := sav.Atk + sav.Def + sav.Hp + sav.Spd + sav.Cc + sav.Cd + sav.Eff + sav.Res
	if total == 0 {
		return map[string]float64{
			"atk": 0.125, "def": 0.125, "hp": 0.125, "spd": 0.125,
			"cc": 0.125, "cd": 0.125, "eff": 0.125, "res": 0.125,
		}
	}
	return map[string]float64{
		"atk": sav.Atk / total,
		"def": sav.Def / total,
		"hp":  sav.Hp / total,
		"spd": sav.Spd / total,
		"cc":  sav.Cc / total,
		"cd":  sav.Cd / total,
		"eff": sav.Eff / total,
		"res": sav.Res / total,
	}
}

func computeMaxWeight(weights map[string]float64) float64 {
	var maxW float64
	for _, w := range weights {
		if w > maxW {
			maxW = w
		}
	}
	return maxW
}

func computeCoreStats(weights map[string]float64) []string {
	type kv struct {
		key string
		val float64
	}
	sorted := make([]kv, 0, len(weights))
	for k, v := range weights {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].val > sorted[j].val
	})

	var core []string
	var cumulative float64
	for _, item := range sorted {
		core = append(core, item.key)
		cumulative += item.val
		if cumulative >= CORE_COVERAGE_TARGET {
			break
		}
	}
	return core
}

func computeGearContributions(gear Gear) map[string]float64 {
	raw := map[string]float64{}
	var total float64

	for _, sub := range gear.Substats {
		gsWeight, ok := GS_WEIGHTS[sub.Type]
		if !ok {
			continue
		}
		gs := sub.Value * gsWeight
		savKey := getSavKey(sub.Type)
		if savKey == "" {
			continue
		}
		raw[savKey] += gs
		total += gs
	}

	if total == 0 {
		return raw
	}

	result := make(map[string]float64, len(raw))
	for k, v := range raw {
		result[k] = v / total
	}
	return result
}

func computeMainBonus(gear Gear, heroWeights map[string]float64) float64 {
	if gear.Slot == "Weapon" || gear.Slot == "Helmet" || gear.Slot == "Armor" {
		return 1.0
	}

	mainSavKey := getSavKey(gear.Main.Type)
	if mainSavKey == "" {
		return 0.75
	}

	wMain := heroWeights[mainSavKey]
	maxW := computeMaxWeight(heroWeights)
	if maxW == 0 {
		return 1.0
	}

	return 0.5 + 0.5*(wMain/maxW)
}

func scoreBuildAlignment(gear Gear, sav Sav) (wasFinal, threshold float64, missingCore int, coreStats []string) {
	heroWeights := computeHeroWeights(sav)
	maxW := computeMaxWeight(heroWeights)
	coreStats = computeCoreStats(heroWeights)
	gearContributions := computeGearContributions(gear)

	var alignment float64
	for savKey, contrib := range gearContributions {
		alignment += heroWeights[savKey] * contrib
	}

	missingCore = 0
	for _, coreKey := range coreStats {
		if gearContributions[coreKey] == 0 {
			missingCore++
		}
	}
	alignment *= 1.0 - MISSING_CORE_PENALTY*float64(missingCore)

	mainBonus := computeMainBonus(gear, heroWeights)
	wasFinal = alignment * mainBonus

	threshold = maxW * WAS_THRESHOLD_FACTOR
	if strings.EqualFold(gear.Rarity, "Heroic") && len(gear.Substats) < 4 {
		threshold *= HEROIC_THRESHOLD_DISCOUNT
	}

	return wasFinal, threshold, missingCore, coreStats
}

// ---------------------------------------------------------------------------
// Utility helpers
// ---------------------------------------------------------------------------

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
