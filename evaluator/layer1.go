package main

import (
	"math"
	"strings"
)

// EvaluateLayer1 verifies the gear against E7 game rules.
func EvaluateLayer1(gear Gear, l0 L0Trace) L1Trace {
	trace := L1Trace{
		RulesRun:   []string{},
		Violations: []string{},
	}

	runRule := func(name string, condition bool, violationCode string) {
		trace.RulesRun = append(trace.RulesRun, name)
		if !condition {
			trace.Violations = append(trace.Violations, violationCode)
		}
	}

	// R1.8 Level & Rarity check
	isLevelValid := gear.Level == 85 || gear.Level == 88
	isRarityValid := strings.EqualFold(gear.Rarity, "Epic") || strings.EqualFold(gear.Rarity, "Heroic")
	runRule("R1.8: Scope Check (level/rarity)", isLevelValid && isRarityValid, "L1_OUT_OF_SCOPE")

	if !isLevelValid || !isRarityValid {
		return trace // abort further checks if scope is wrong
	}

	// R1.1 Main stat slot legality
	allowedMains, ok := MainStatsBySlot[gear.Slot]
	runRule("R1.1: Slot Main Stat Legality", ok && allowedMains[gear.Main.Type], "L1_ILLEGAL_MAIN_FOR_SLOT")

	// R1.2 Main stat != Substat collision
	collisionFound := false
	for _, sub := range gear.Substats {
		if sub.Type == gear.Main.Type {
			collisionFound = true
			break
		}
	}
	runRule("R1.2: Main Sub Stat Collision", !collisionFound, "L1_MAIN_SUB_COLLISION")

	// R1.3 Allowed Substats by Slot
	substatsLegal := true
	allowedSubs, ok := SubstatsBySlot[gear.Slot]
	if ok {
		for _, sub := range gear.Substats {
			if !allowedSubs[sub.Type] {
				substatsLegal = false
				break
			}
		}
	} else {
		substatsLegal = false
	}
	runRule("R1.3: Slot Sub Stat Legality", substatsLegal, "L1_ILLEGAL_SUB_FOR_SLOT")

	// R1.4 Substats distinct
	distinct := true
	seenSubs := make(map[string]bool)
	for _, sub := range gear.Substats {
		if seenSubs[sub.Type] {
			distinct = false
			break
		}
		seenSubs[sub.Type] = true
	}
	runRule("R1.4: Distinct Substats Check", distinct, "L1_DUPLICATE_SUB")

	// R1.5 Substat count matches rarity + enhance
	expectedSubCount := 4
	if strings.EqualFold(gear.Rarity, "Heroic") {
		if gear.Enhance < 12 {
			expectedSubCount = 3
		} else {
			expectedSubCount = 4
		}
	}
	runRule("R1.5: Substat Count Check", len(gear.Substats) == expectedSubCount, "L1_WRONG_SUB_COUNT")

	// R1.6 Reconstructed rolls feasible and substat values within legal roll range bounds
	rollsFeasible := true
	for _, sub := range gear.Substats {
		if sub.Rolls <= 0 && sub.Value > 0 {
			rollsFeasible = false
			break
		}
		if rangesLvl, ok := SubstatRollRanges[gear.Level]; ok {
			var rarityKey string
			if strings.EqualFold(gear.Rarity, "Epic") {
				rarityKey = "Epic"
			} else if strings.EqualFold(gear.Rarity, "Heroic") {
				rarityKey = "Heroic"
			}
			if rangesRarity, ok := rangesLvl[rarityKey]; ok {
				if rBounds, ok := rangesRarity[sub.Type]; ok {
					minBound := float64(sub.Rolls) * rBounds[0]
					maxBound := float64(sub.Rolls) * rBounds[1]

					checkVal := sub.Value
					if gear.Reforged {
						if increments, ok := ReforgeIncrements[sub.Type]; ok {
							rIdx := sub.Rolls
							if rIdx >= len(increments) {
								rIdx = len(increments) - 1
							}
							if rIdx >= 0 {
								checkVal -= increments[rIdx]
							}
						}
					}

					if checkVal < minBound-0.1 || checkVal > maxBound+0.1 {
						rollsFeasible = false
						break
					}
				}
			}
		}
	}
	runRule("R1.6: Substat Value Feasibility", rollsFeasible, "L1_VALUE_OUT_OF_RANGE")


	// R1.7 Main stat value matches progression
	mainValCorrect := true
	// Check if enhance level is a multiple of 3
	if gear.Enhance%3 == 0 {
		if lvlProg, ok := MainStatProgression[gear.Level]; ok {
			if enhProg, ok := lvlProg[gear.Enhance]; ok {
				if expectedVal, ok := enhProg[gear.Main.Type]; ok {
					actualVal := gear.Main.Value
					if gear.Reforged {
						actualVal -= MainStatReforgeBonuses[gear.Main.Type]
					}
					// Allow +/- 1 tolerance for rounding errors
					diff := math.Abs(actualVal - expectedVal)
					if diff > 1.05 {
						mainValCorrect = false
					}
				}
			}
		}
	}
	runRule("R1.7: Main Value Progression Check", mainValCorrect, "L1_BAD_MAIN_VALUE")

	// R1.10 At most one modified substat
	modifiedCount := 0
	for _, sub := range gear.Substats {
		if sub.Modified {
			modifiedCount++
		}
	}
	runRule("R1.10: Modification Stone Count Check", modifiedCount <= 1, "L1_MOD_RULE_VIOLATION")

	return trace
}
