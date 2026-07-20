package main

import (
	"fmt"
)

// EvaluateLayer0 performs Ingest, roll reconstruction, and normalizes gear level/reforge state.
func EvaluateLayer0(gear Gear) (Gear, L0Trace) {
	trace := L0Trace{
		ParseConfidence:    1.0, // Manual/API entry is 100% confident
		RollReconstruction: make(map[string]int),
		Ambiguities:        []string{},
	}

	// We work on a copy of the gear to prevent side effects
	normalizedGear := gear

	level := normalizedGear.Level
	rarity := normalizedGear.Rarity
	if level != 85 && level != 88 {
		trace.Ambiguities = append(trace.Ambiguities, fmt.Sprintf("Gear level %d is out of scope (85 or 88 only)", level))
		return normalizedGear, trace
	}

	// Step 1: Expected roll budget calculation
	expectedRolls := GetExpectedRolls(rarity, normalizedGear.Enhance)

	// Step 2: Generate candidate roll counts for each substat
	candidatesPerSub := make([][]int, len(normalizedGear.Substats))
	baseValuesPerSub := make([][]float64, len(normalizedGear.Substats)) // Base value before reforge

	for i, sub := range normalizedGear.Substats {
		cands, bases := getCandidateRolls(sub.Type, sub.Value, level, rarity, normalizedGear.Reforged)
		candidatesPerSub[i] = cands
		baseValuesPerSub[i] = bases

		if len(cands) == 0 {
			trace.Ambiguities = append(trace.Ambiguities, fmt.Sprintf("Substat %s value %.1f has no valid roll count candidates", sub.Type, sub.Value))
		} else if len(cands) > 1 {
			trace.Ambiguities = append(trace.Ambiguities, fmt.Sprintf("Substat %s value %.1f is ambiguous: candidates %v", sub.Type, sub.Value, cands))
		}
	}

	// Step 3: Run conservation check solver to find the combination of rolls that sums to expectedRolls
	var validCombination []int
	var validBases []float64
	var foundMultiple bool

	var search func(index int, currentSum int, currentCombo []int, currentBases []float64)
	search = func(index int, currentSum int, currentCombo []int, currentBases []float64) {
		if index == len(normalizedGear.Substats) {
			if currentSum == expectedRolls {
				if len(validCombination) > 0 {
					foundMultiple = true
				} else {
					validCombination = make([]int, len(currentCombo))
					copy(validCombination, currentCombo)
					validBases = make([]float64, len(currentBases))
					copy(validBases, currentBases)
				}
			}
			return
		}

		subCands := candidatesPerSub[index]
		subBases := baseValuesPerSub[index]
		for k, r := range subCands {
			search(index+1, currentSum+r, append(currentCombo, r), append(currentBases, subBases[k]))
		}
	}

	search(0, 0, []int{}, []float64{})

	if len(validCombination) > 0 {
		// Update substat roll counts and subtract reforge increments in the normalized gear if reforged
		for i := range normalizedGear.Substats {
			normalizedGear.Substats[i].Rolls = validCombination[i]
			trace.RollReconstruction[normalizedGear.Substats[i].Type] = validCombination[i]
			// If reforged, we store base values internally or we can keep the reforged value on the Gear
			// but we want the normalizer/L5 to know the base rolls.
		}

		if foundMultiple {
			trace.Ambiguities = append(trace.Ambiguities, "Multiple valid roll combinations sum to the expected roll budget")
		}
	} else {
		// Fallback: if conservation check fails, assign the closest single-stat candidate counts
		trace.Ambiguities = append(trace.Ambiguities, fmt.Sprintf("Conservation check failed: sum of rolls does not match expected total %d", expectedRolls))
		for i, cands := range candidatesPerSub {
			if len(cands) > 0 {
				// pick first candidate
				normalizedGear.Substats[i].Rolls = cands[0]
				trace.RollReconstruction[normalizedGear.Substats[i].Type] = cands[0]
			} else {
				normalizedGear.Substats[i].Rolls = 1
			}
		}
	}

	return normalizedGear, trace
}

// GetExpectedRolls calculates the exact number of rolls a gear should have based on rarity and enhancement
func GetExpectedRolls(rarity string, enhance int) int {
	// Epic: starts with 4 rolls, gets 1 at each +3/+6/+9/+12/+15 (up to 9 rolls total)
	if rarity == "Epic" || rarity == "epic" {
		return 4 + enhance/3
	}

	// Heroic: starts with 3 rolls, gets 1 at +3/+6/+9. Unlocks 4th sub at +12 (gets 1 forced roll), and +15 (gets another forced roll).
	// +0 to +9: 3 + enhance/3
	// +12: 7 rolls (3 base + 3 random + 1 forced 4th sub)
	// +15: 8 rolls (3 base + 3 random + 2 forced 4th sub)
	if enhance < 12 {
		return 3 + enhance/3
	} else if enhance < 15 {
		return 7
	}
	return 8
}

// getCandidateRolls returns the valid roll count candidates and corresponding base values
func getCandidateRolls(statType string, value float64, level int, rarity string, reforged bool) ([]int, []float64) {
	var candidates []int
	var bases []float64

	// Ranges for level & rarity
	rangesLvl, ok := SubstatRollRanges[level]
	if !ok {
		return candidates, bases
	}
	ranges, ok := rangesLvl[rarity]
	if !ok {
		return candidates, bases
	}
	rBounds, ok := ranges[statType]
	if !ok {
		return candidates, bases
	}

	minRoll := rBounds[0]
	maxRoll := rBounds[1]

	// Determine check ranges (1 to 6 rolls)
	for r := 1; r <= 6; r++ {
		baseVal := value
		if reforged {
			increments, hasIncr := ReforgeIncrements[statType]
			if hasIncr && r < len(increments) {
				baseVal = value - increments[r]
			}
		}

		// Check if base value fits within roll boundaries
		// For floats (like Speed, Percent stats), we check within rounding threshold (0.01)
		// Flat stats are integers in game, but we parse them as floats.
		minBound := float64(r) * minRoll
		maxBound := float64(r) * maxRoll

		if baseVal >= minBound-0.05 && baseVal <= maxBound+0.05 {
			candidates = append(candidates, r)
			bases = append(bases, baseVal)
		}
	}

	// Fallback for 0 rolls (e.g. if stat was never rolled, it shouldn't show up in substats,
	// but if value is 0, then 0 rolls is the only option)
	if value == 0 {
		return []int{0}, []float64{0}
	}

	return candidates, bases
}

// getSavKey maps internal stat type to the SAV key string
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

// isRightSideSlot returns true if the slot is Necklace, Ring or Boots
func isRightSideSlot(slot string) bool {
	return slot == "Necklace" || slot == "Ring" || slot == "Boots"
}

// isFlatMainStat returns true if the main stat is a flat Atk/HP/Def
func isFlatMainStat(statType string) bool {
	return statType == "Attack" || statType == "Health" || statType == "Defense"
}
