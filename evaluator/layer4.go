package main

import (
	"math"
	"strings"
)

func normalizeSetName(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 3 && strings.EqualFold(s[len(s)-3:], "set") {
		s = strings.TrimSpace(s[:len(s)-3])
	}
	lower := strings.ToLower(s)
	switch lower {
	case "resist", "resistance", "effect resistance", "effectresist":
		return "resistance"
	case "hit", "effectiveness", "hit (effect)":
		return "hit"
	case "crit", "critical", "critical rate", "crit rate":
		return "critical"
	case "cdmg", "crit damage", "crit dmg", "destruction":
		return "destruction"
	default:
		return lower
	}
}

func matchSetNames(set1, set2 string) bool {
	return normalizeSetName(set1) == normalizeSetName(set2)
}

// EvaluateLayer4 matches the gear against a hero profile and runs Gates 1-4.
func EvaluateLayer4(gear Gear, profile HeroProfile, baseStats HeroBaseStats) L4HeroGateResult {
	result := L4HeroGateResult{
		HeroName:    profile.HeroName,
		BuildRank:   profile.BuildRank,
		GateResults: [4]bool{false, false, false, false},
		FitScore:    0.0,
		Threshold:   0.0,
		Pass:        false,
	}

	// --- Gate 1: Set compatibility ---
	gate1Pass := false
	if len(profile.Sets) == 0 {
		gate1Pass = true
	} else {
		for _, combo := range profile.Sets {
			comboMatched := false
			for _, s := range combo {
				if matchSetNames(s, gear.Set) {
					comboMatched = true
					break
				}
			}
			if comboMatched {
				gate1Pass = true
				break
			}
		}
	}
	result.GateResults[0] = gate1Pass

	// --- Gate 2: Slot/main-stat service ---
	gate2Pass := false
	if !isRightSideSlot(gear.Slot) {
		gate2Pass = true
	} else {
		mainSavKey := getSavKey(gear.Main.Type)
		prio := profile.Priorities[mainSavKey]
		hasMinBound := false
		if profile.StatRanges != nil {
			if bounds, ok := profile.StatRanges[mainSavKey]; ok && bounds.Min != nil {
				hasMinBound = true
			}
		}

		if prio >= 1 || hasMinBound {
			gate2Pass = true
		}

		// Accessory main stats checklist validation
		if len(profile.AccessoryMains) > 0 {
			mainStatAllowed := false
			for _, accMain := range profile.AccessoryMains {
				if strings.EqualFold(accMain, gear.Main.Type) {
					mainStatAllowed = true
					break
				}
			}
			if !mainStatAllowed {
				gate2Pass = false
			}
		}

		// Strict-mode exclusion check
		if profile.WeightMode == "strict" && profile.StatRanges != nil {
			if bounds, ok := profile.StatRanges[mainSavKey]; ok && bounds.Max != nil && *bounds.Max <= 0 {
				gate2Pass = false
			}
		}
	}
	result.GateResults[1] = gate2Pass

	// --- Gate 3: Stat-range feasibility (min/max) ---
	gate3Pass := true
	if profile.StatRanges != nil {
		// 1. Min capability checks
		minStats := []string{}
		for k, bounds := range profile.StatRanges {
			if bounds.Min != nil {
				minStats = append(minStats, k)
			}
		}

		servedCount := 0
		for _, statKey := range minStats {
			// Check if gear has this stat as main or sub
			hasStat := false
			if getSavKey(gear.Main.Type) == statKey {
				hasStat = true
			} else {
				for _, sub := range gear.Substats {
					if getSavKey(sub.Type) == statKey {
						hasStat = true
						break
					}
				}
			}

			if hasStat {
				servedCount++
			}
		}

		k := len(minStats)
		if k > 0 {
			if profile.WeightMode == "strict" {
				if servedCount < k {
					gate3Pass = false
				}
			} else {
				// Weighted mode: must serve at least ceil(k/2)
				target := int(math.Ceil(float64(k) / 2.0))
				if servedCount < target {
					gate3Pass = false
				}
			}
		}

		// 2. Max waste guard checks (strict mode fails hard-cap CC overflow)
		if gate3Pass {
			for statKey, bounds := range profile.StatRanges {
				if bounds.Max != nil {
					maxVal := *bounds.Max
					// Calculate gear's contribution to this stat
					contrib := getGearContributionToStat(gear, statKey, baseStats)
					baseVal := getHeroBaseStatValue(statKey, baseStats)

					if baseVal+contrib > maxVal {
						// cc is a hard-cap stat
						if statKey == "cc" && profile.WeightMode == "strict" {
							gate3Pass = false
							break
						}
					}
				}
			}
		}
	}
	result.GateResults[2] = gate3Pass

	// --- Gate 4: Priority fit score ---
	gate4Pass := false
	fitScore := GetHeroWeightedScore(gear, profile, baseStats)
	result.FitScore = fitScore

	// Apply penalty for soft-cap max bounds overflow if present
	if profile.StatRanges != nil {
		for statKey, bounds := range profile.StatRanges {
			if bounds.Max != nil {
				maxVal := *bounds.Max
				contrib := getGearContributionToStat(gear, statKey, baseStats)
				baseVal := getHeroBaseStatValue(statKey, baseStats)
				if baseVal+contrib > maxVal {
					// Soft-cap penalty: subtract the overflow portion scaled down
					overflow := (baseVal + contrib) - maxVal
					result.FitScore -= overflow * 2.0 // penalty multiplier
				}
			}
		}
	}

	// Dynamic threshold based on priorities
	maxW := 0.0
	for _, w := range profile.Priorities {
		var weight float64
		switch w {
		case 1:
			weight = 1.0
		case 2:
			weight = 2.5
		case 3:
			weight = 5.0
		}
		if weight > maxW {
			maxW = weight
		}
	}

	baseMultiplier := 15.0
	if gear.Level == 88 {
		if strings.EqualFold(gear.Rarity, "Epic") {
			baseMultiplier = 18.0
		} else {
			baseMultiplier = 15.0
		}
	} else {
		if strings.EqualFold(gear.Rarity, "Epic") {
			baseMultiplier = 16.0
		} else {
			baseMultiplier = 13.0
		}
	}

	var threshold float64
	if profile.MinQuality != nil && profile.MinQuality.Score != nil {
		topPercent := *profile.MinQuality.Score
		multiplier := baseMultiplier * (100.0 - topPercent) / 85.0
		if multiplier < 0 {
			multiplier = 0
		}
		threshold = maxW * multiplier
	} else {
		threshold = maxW * baseMultiplier
	}
	result.Threshold = threshold

	// Perform checks
	if result.FitScore >= threshold {
		gate4Pass = true
	}

	// Strict mode: every priority-3 stat must be present
	if profile.WeightMode == "strict" {
		for statKey, prio := range profile.Priorities {
			if prio == 3 {
				hasStat := false
				if getSavKey(gear.Main.Type) == statKey {
					hasStat = true
				} else {
					for _, sub := range gear.Substats {
						if getSavKey(sub.Type) == statKey {
							hasStat = true
							break
						}
					}
				}
				if !hasStat {
					gate4Pass = false
					break
				}
			}
		}
	}

	result.GateResults[3] = gate4Pass

	// Aggregation: all gates must pass
	if gate1Pass && gate2Pass && gate3Pass && gate4Pass {
		result.Pass = true
	}

	return result
}

// Helpers

func getGearContributionToStat(gear Gear, statKey string, baseStats HeroBaseStats) float64 {
	var contrib float64
	// Check main stat
	if getSavKey(gear.Main.Type) == statKey {
		contrib += gear.Main.Value
	}
	// Check substats
	for _, sub := range gear.Substats {
		if getSavKey(sub.Type) == statKey {
			contrib += sub.Value
		}
	}

	// Convert flats to percent equivalent if necessary
	if statKey == "atk" || statKey == "def" || statKey == "hp" {
		// If the contributions are mixed, we must sum flat contributions and convert them
		var flats float64
		var percents float64

		if getSavKey(gear.Main.Type) == statKey {
			if isFlatMainStat(gear.Main.Type) {
				flats += gear.Main.Value
			} else {
				percents += gear.Main.Value
			}
		}
		for _, sub := range gear.Substats {
			if getSavKey(sub.Type) == statKey {
				if strings.HasSuffix(sub.Type, "Percent") || sub.Type == "AttackPercent" || sub.Type == "HealthPercent" || sub.Type == "DefensePercent" {
					percents += sub.Value
				} else {
					flats += sub.Value
				}
			}
		}

		baseVal := getHeroBaseStatValue(statKey, baseStats)
		if baseVal > 0 {
			return percents + (flats/baseVal)*100.0
		}
		return percents
	}

	return contrib
}

func getHeroBaseStatValue(statKey string, baseStats HeroBaseStats) float64 {
	switch statKey {
	case "atk":
		return baseStats.Atk
	case "def":
		return baseStats.Def
	case "hp":
		return baseStats.Hp
	case "spd":
		return baseStats.Spd
	case "cc":
		return baseStats.Cc * 100.0 // base stats in stats.json are decimals (e.g. 0.15), convert to %
	case "cd":
		return baseStats.Cd * 100.0
	case "eff":
		return baseStats.Eff
	case "res":
		return baseStats.Res
	default:
		return 0
	}
}
