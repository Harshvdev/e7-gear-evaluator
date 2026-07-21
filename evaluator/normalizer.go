package main

import (
	"sort"
	"strings"
)

// ---------------------------------------------------------------------------
// normalizer.go — Shared scoring, normalization, and contribution calculations.
// ---------------------------------------------------------------------------

// GetPriorityWeight retrieves the signed weight for a 9-position priority slider (-3..+5) per Q-E7-Architecture §2.2
func GetPriorityWeight(prio int) float64 {
	if w, ok := PRIORITY_WEIGHTS[prio]; ok {
		return w
	}
	if prio > 5 {
		return PRIORITY_WEIGHTS[5]
	}
	if prio < -3 {
		return PRIORITY_WEIGHTS[-3]
	}
	return 0.0
}

// GetSingleRollQuality calculates quality q(s,v) in [0, 1] per Q-E7-Architecture §6
func GetSingleRollQuality(statType string, value float64, level int, rarity string) float64 {
	rangesLvl, ok := SubstatRollRanges[level]
	if !ok {
		level = 85
		rangesLvl = SubstatRollRanges[85]
	}
	rarityKey := "Epic"
	if strings.EqualFold(rarity, "Heroic") {
		rarityKey = "Heroic"
	}
	ranges, ok := rangesLvl[rarityKey]
	if !ok {
		return 0.5
	}
	rBounds, ok := ranges[statType]
	if !ok {
		return 0.5
	}
	rmin, rmax := rBounds[0], rBounds[1]
	if rmax <= rmin {
		return 0.5
	}
	q := (value - rmin) / (rmax - rmin)
	if q > 1.0 {
		return 1.0
	} else if q < 0.0 {
		return 0.0
	}
	return q
}

// GetStepValue calculates single-step value = w(s) * q(s,v) per Q-E7-Architecture §6
func GetStepValue(statType string, value float64, prio int, level int, rarity string) float64 {
	w := GetPriorityWeight(prio)
	q := GetSingleRollQuality(statType, value, level, rarity)
	return w * q
}

// ClassifyStep evaluates step class (GOOD, NEUTRAL, WASTED, HARM, MISS) per Q-E7-Architecture §6
func ClassifyStep(mode string, statType string, delta float64, prio int, level int, rarity string) string {
	if strings.EqualFold(mode, "SPEED") {
		if NormalizeStatType(statType) == "Speed" {
			return "GOOD"
		}
		return "MISS"
	}

	w := GetPriorityWeight(prio)
	q := GetSingleRollQuality(statType, delta, level, rarity)
	stepVal := w * q

	if w < 0 {
		return "HARM"
	}
	if (w <= 1.0 && q <= 0.25) || (stepVal <= 0 && w <= 1.0) {
		return "WASTED"
	}
	if stepVal >= 1.5 {
		return "GOOD"
	}
	return "NEUTRAL"
}

// NormalizeStatValue converts any stat value (flat, speed, critical) into the
// "percent-equivalent" space for a specific hero's base stats.
func NormalizeStatValue(statType string, value float64, baseStats HeroBaseStats) float64 {
	switch statType {
	case "AttackPercent", "HealthPercent", "DefensePercent", "EffectivenessPercent", "EffectResistancePercent":
		return value
	case "Speed":
		return value * 2.0
	case "CritHitChancePercent":
		return value * 1.6
	case "CritHitDamagePercent":
		return value * (8.0 / 7.0) // 1.142857
	case "Attack":
		if baseStats.Atk > 0 {
			return (value / baseStats.Atk) * 100.0
		}
		return (value / 1000.0) * 100.0 // fallback default
	case "Defense":
		if baseStats.Def > 0 {
			return (value / baseStats.Def) * 100.0
		}
		return (value / 600.0) * 100.0 // fallback default
	case "Health":
		if baseStats.Hp > 0 {
			return (value / baseStats.Hp) * 100.0
		}
		return (value / 6000.0) * 100.0 // fallback default
	default:
		return 0
	}
}

// GetWSSScore calculates the generic weighted substat score (standard Gear Score) for a gear piece
func GetWSSScore(gear Gear) float64 {
	var total float64
	for _, sub := range gear.Substats {
		coef, ok := WSSCoefficients[sub.Type]
		if ok {
			total += sub.Value * coef
		}
	}
	return total
}

// CalculateAbsencePenalty evaluates absence cost for hero's defining stats (+5 or +4) that the slot could have rolled
func CalculateAbsencePenalty(gear Gear, profile HeroProfile) float64 {
	var penalty float64
	allStats := []string{"atk", "def", "hp", "spd", "cc", "cd", "eff", "res"}

	allowedSubs, hasSlot := SubstatsBySlot[gear.Slot]
	if !hasSlot {
		return 0.0
	}

	mainSavKey := getSavKey(gear.Main.Type)
	existingStats := make(map[string]bool)
	existingStats[mainSavKey] = true
	for _, sub := range gear.Substats {
		existingStats[getSavKey(sub.Type)] = true
	}

	for _, statKey := range allStats {
		prio := profile.Priorities[statKey]
		if prio >= 4 { // ESSENTIAL (+5) or CORE (+4)
			// Check if slot COULD roll this stat as a substat
			slotCanRoll := false
			for legalSub := range allowedSubs {
				if getSavKey(legalSub) == statKey {
					slotCanRoll = true
					break
				}
			}
			if !slotCanRoll {
				continue // slot-impossible (e.g. Def on Weapon, Atk on Armor) -> no penalty
			}

			// Check if piece omitted it
			if !existingStats[statKey] {
				if cost, ok := ABSENCE_COST[prio]; ok {
					penalty += cost
				}
			}
		}
	}

	return penalty
}

func isLeftSideSlot(slot string) bool {
	s := strings.ToLower(slot)
	return s == "weapon" || s == "helmet" || s == "armor"
}

func getMainWeight(slot string, prio int) float64 {
	if isLeftSideSlot(slot) {
		if prio < 0 {
			return GetPriorityWeight(prio)
		}
		return 1.0
	}
	return GetPriorityWeight(prio)
}

// CalculateMaxFit computes the theoretical maximum raw fit score the slot can give to the hero per Q-E7-Architecture §3.4
func CalculateMaxFit(gear Gear, profile HeroProfile, baseStats HeroBaseStats) float64 {
	// 1. Find best legal main for slot
	legalMains, ok := MainStatsBySlot[gear.Slot]
	if !ok {
		return 100.0
	}

	level := gear.Level
	if level != 88 {
		level = 85
	}

	bestMainScore := 0.0
	bestMainType := ""
	for mainType := range legalMains {
		prio := profile.Priorities[getSavKey(mainType)]
		w := getMainWeight(gear.Slot, prio)
		if w <= 0 {
			continue
		}
		valAt15 := MainStatValuesAtPlus15[level][mainType]
		normVal := NormalizeStatValue(mainType, valAt15, baseStats)
		score := w * normVal
		if score > bestMainScore {
			bestMainScore = score
			bestMainType = mainType
		}
	}

	// 2. Draw top 4 distinct legal substats for slot excluding chosen main
	legalSubs := SubstatsBySlot[gear.Slot]
	type subCandidate struct {
		statType  string
		unitScore float64
	}

	rarityKey := "Epic"
	if strings.EqualFold(gear.Rarity, "Heroic") {
		rarityKey = "Heroic"
	}

	var candidates []subCandidate
	for subType := range legalSubs {
		if NormalizeStatType(subType) == NormalizeStatType(bestMainType) {
			continue
		}
		prio := profile.Priorities[getSavKey(subType)]
		w := GetPriorityWeight(prio)
		if w <= 0 {
			continue
		}
		ranges, ok := SubstatRollRanges[level][rarityKey]
		if !ok {
			continue
		}
		maxRollVal := ranges[subType][1]
		unitNormVal := NormalizeStatValue(subType, maxRollVal, baseStats)
		candidates = append(candidates, subCandidate{statType: subType, unitScore: w * unitNormVal})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].unitScore > candidates[j].unitScore
	})

	// Distribute 9 total max rolls (5 enhancements + 4 base rolls) across top legal substats
	topSubScore := 0.0
	if len(candidates) >= 4 {
		topSubScore = 6.0*candidates[0].unitScore + 1.0*candidates[1].unitScore + 1.0*candidates[2].unitScore + 1.0*candidates[3].unitScore
	} else if len(candidates) == 3 {
		topSubScore = 6.0*candidates[0].unitScore + 2.0*candidates[1].unitScore + 1.0*candidates[2].unitScore
	} else if len(candidates) == 2 {
		topSubScore = 6.0*candidates[0].unitScore + 3.0*candidates[1].unitScore
	} else if len(candidates) == 1 {
		topSubScore = 9.0*candidates[0].unitScore
	}

	maxFit := bestMainScore + topSubScore
	if maxFit <= 0 {
		return 100.0 // fallback denominator
	}

	return maxFit
}

// CalculateHeroFit computes rawFit, maxFit, signed fitPct in [-100, +100], and fitClass per Q-E7-Architecture §3
func CalculateHeroFit(gear Gear, profile HeroProfile, baseStats HeroBaseStats) (rawFit float64, maxFit float64, fitPct float64, fitClass string, absencePenalty float64) {
	// Substat signed contributions
	var subContrib float64
	for _, sub := range gear.Substats {
		savKey := getSavKey(sub.Type)
		prio := profile.Priorities[savKey]
		w := GetPriorityWeight(prio)
		normVal := NormalizeStatValue(sub.Type, sub.Value, baseStats)
		subContrib += w * normVal
	}

	// Main stat signed contribution
	mainSavKey := getSavKey(gear.Main.Type)
	mainPrio := profile.Priorities[mainSavKey]
	mainW := getMainWeight(gear.Slot, mainPrio)
	mainNormVal := NormalizeStatValue(gear.Main.Type, gear.Main.Value, baseStats)
	mainContrib := mainW * mainNormVal

	// Absence penalty
	absencePenalty = CalculateAbsencePenalty(gear, profile)

	// Raw Fit
	rawFit = subContrib + mainContrib - absencePenalty

	// Max Fit denominator
	maxFit = CalculateMaxFit(gear, profile, baseStats)

	// fit% = clamp( rawFit / maxFit * 100 , -100, +100 )
	fitPct = (rawFit / maxFit) * 100.0
	if fitPct > 100.0 {
		fitPct = 100.0
	} else if fitPct < -100.0 {
		fitPct = -100.0
	}

	// Fit Class Bands
	if fitPct >= 70.0 {
		fitClass = "CORE"
	} else if fitPct >= 45.0 {
		fitClass = "USABLE"
	} else if fitPct >= 20.0 {
		fitClass = "MARGINAL"
	} else {
		fitClass = "REJECT"
	}

	return rawFit, maxFit, fitPct, fitClass, absencePenalty
}

// GetHeroWeightedScore calculates custom score for legacy UI backward compatibility using signed priority weights
func GetHeroWeightedScore(gear Gear, profile HeroProfile, baseStats HeroBaseStats) float64 {
	rawFit, _, _, _, _ := CalculateHeroFit(gear, profile, baseStats)
	return rawFit
}

// GetHeroWeightedScoreForSubstats calculates custom score using a slice of substats (for projections)
func GetHeroWeightedScoreForSubstats(substats []Substat, profile HeroProfile, baseStats HeroBaseStats) float64 {
	var total float64
	for _, sub := range substats {
		savKey := getSavKey(sub.Type)
		prio := profile.Priorities[savKey]
		w := GetPriorityWeight(prio)
		normVal := NormalizeStatValue(sub.Type, sub.Value, baseStats)
		total += w * normVal
	}
	return total
}
