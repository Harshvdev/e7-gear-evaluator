package main

import (
	"math"
	"strings"
)

// EvaluateLayer5 calculates current and projected substat envelopes at +15, including reforge increments.
func EvaluateLayer5(gear Gear, profile HeroProfile, baseStats HeroBaseStats) L5Trace {
	trace := L5Trace{
		CurrentWSS:       GetWSSScore(gear),
		CurrentHeroScore: make(map[string]float64),
		Scenarios:        []ProjectionScenario{},
	}

	trace.CurrentHeroScore[profile.HeroID] = GetHeroWeightedScore(gear, profile, baseStats)

	// Step 1: Calculate remaining rolls
	enhance := gear.Enhance
	rarity := gear.Rarity
	level := gear.Level

	var randRolls int
	var forced4thRolls int
	var forced4thExists bool

	if strings.EqualFold(rarity, "Epic") {
		// Epic: starting 4 rolls, gets 1 at each +3/+6/+9/+12/+15 (up to 9 rolls total)
		// already rolled: 4 + enhance/3
		// remaining: 9 - (4 + enhance/3) = 5 - enhance/3
		randRolls = 5 - enhance/3
		if randRolls < 0 {
			randRolls = 0
		}
	} else {
		// Heroic: starts with 3, gets 1 at each +3/+6/+9.
		// Unlocks 4th sub at +12 (forced roll) and +15 (forced roll).
		if enhance < 12 {
			// remaining random rolls for original 3 subs
			randRolls = 3 - enhance/3
			forced4thRolls = 2
			forced4thExists = false
		} else if enhance < 15 {
			randRolls = 0
			forced4thRolls = 1
			forced4thExists = true
		} else {
			randRolls = 0
			forced4thRolls = 0
			forced4thExists = true
		}
	}

	// Legal stats for 4th substat (if not exists yet)
	var legal4thStats []string
	if !forced4thExists && len(gear.Substats) < 4 {
		allowed := SubstatsBySlot[gear.Slot]
		// Subtract main stat
		// Subtract 3 existing substats
		for s := range allowed {
			if s == gear.Main.Type {
				continue
			}
			exists := false
			for _, sub := range gear.Substats {
				if sub.Type == s {
					exists = true
					break
				}
			}
			if !exists {
				legal4thStats = append(legal4thStats, s)
			}
		}
	}

	// Enumerate scenarios
	scenarios := []string{"expected", "best", "worst"}
	for _, name := range scenarios {
		projSubstats := make([]Substat, len(gear.Substats))
		copy(projSubstats, gear.Substats)

		// 1. Project random rolls into existing stats
		if randRolls > 0 {
			switch name {
			case "expected":
				// Distribute expected rolls uniformly across existing subs
				expectedRandPerSub := float64(randRolls) / float64(len(gear.Substats))
				for i := range projSubstats {
					avgRoll := GetAverageRoll(level, rarity, projSubstats[i].Type)
					projSubstats[i].Value += expectedRandPerSub * avgRoll
					projSubstats[i].Rolls += int(math.Round(expectedRandPerSub))
				}
			case "best":
				// Find the best substat for the hero
				bestIdx := 0
				bestWeight := -1.0
				for i, sub := range projSubstats {
					savKey := getSavKey(sub.Type)
					w := getHeroWeightMultiplier(profile.Priorities[savKey])
					if w > bestWeight {
						bestWeight = w
						bestIdx = i
					}
				}
				// Put all random rolls into the best sub, at max roll value
				ranges, _ := SubstatRollRanges[level][rarity]
				maxRoll := ranges[projSubstats[bestIdx].Type][1]
				projSubstats[bestIdx].Value += float64(randRolls) * maxRoll
				projSubstats[bestIdx].Rolls += randRolls
			case "worst":
				// Find the worst substat for the hero
				worstIdx := 0
				worstWeight := 999.0
				for i, sub := range projSubstats {
					savKey := getSavKey(sub.Type)
					w := getHeroWeightMultiplier(profile.Priorities[savKey])
					if w < worstWeight {
						worstWeight = w
						worstIdx = i
					}
				}
				// Put all random rolls into the worst sub, at min roll value
				ranges, _ := SubstatRollRanges[level][rarity]
				minRoll := ranges[projSubstats[worstIdx].Type][0]
				projSubstats[worstIdx].Value += float64(randRolls) * minRoll
				projSubstats[worstIdx].Rolls += randRolls
			}
		}

		// 2. Project 4th substat forced rolls (if not yet exists)
		if forced4thRolls > 0 && len(legal4thStats) > 0 {
			var chosen4thType string
			var val float64
			var rolls int = forced4thRolls

			switch name {
			case "expected":
				// Probability-weighted average of all legal stats
				// In expected case, we sum the scores and represent it as a virtual average stat contribution.
				// For the list of substats, we can just pick the legal stat with the average weight,
				// or average each legal stat's rolls.
				// To keep it simple and representable, we create a virtual substat representing the expected 4th sub
				// or we pick the legal stat that represents the median priority.
				// Let's pick the legal stat with the highest weight for the expected case as a realistic average,
				// or compute the average of all legal stats.
				// Let's average values:
				var avgVal float64
				var avgType string = "AttackPercent" // placeholder
				for _, s := range legal4thStats {
					avgRoll := GetAverageRoll(level, rarity, s)
					avgVal += float64(forced4thRolls) * avgRoll
				}
				avgVal /= float64(len(legal4thStats))
				chosen4thType = avgType
				val = avgVal
			case "best":
				// Pick the single best legal stat
				bestType := legal4thStats[0]
				bestWeight := -1.0
				for _, s := range legal4thStats {
					w := getHeroWeightMultiplier(profile.Priorities[getSavKey(s)])
					if w > bestWeight {
						bestWeight = w
						bestType = s
					}
				}
				ranges, _ := SubstatRollRanges[level][rarity]
				maxRoll := ranges[bestType][1]
				chosen4thType = bestType
				val = float64(forced4thRolls) * maxRoll
			case "worst":
				// Pick the single worst legal stat
				worstType := legal4thStats[0]
				worstWeight := 999.0
				for _, s := range legal4thStats {
					w := getHeroWeightMultiplier(profile.Priorities[getSavKey(s)])
					if w < worstWeight {
						worstWeight = w
						worstType = s
					}
				}
				ranges, _ := SubstatRollRanges[level][rarity]
				minRoll := ranges[worstType][0]
				chosen4thType = worstType
				val = float64(forced4thRolls) * minRoll
			}

			projSubstats = append(projSubstats, Substat{
				Type:  chosen4thType,
				Value: val,
				Rolls: rolls,
			})
		} else if forced4thRolls > 0 && len(projSubstats) == 4 {
			// 4th sub already exists, add forced rolls into it
			idx := 3
			subType := projSubstats[idx].Type
			switch name {
			case "expected":
				projSubstats[idx].Value += float64(forced4thRolls) * GetAverageRoll(level, rarity, subType)
			case "best":
				ranges, _ := SubstatRollRanges[level][rarity]
				maxRoll := ranges[subType][1]
				projSubstats[idx].Value += float64(forced4thRolls) * maxRoll
			case "worst":
				ranges, _ := SubstatRollRanges[level][rarity]
				minRoll := ranges[subType][0]
				projSubstats[idx].Value += float64(forced4thRolls) * minRoll
			}
			projSubstats[idx].Rolls += forced4thRolls
		}

		// 3. Reforge calculation (L85 -> L90)
		reforgedSubstats := make([]Substat, len(projSubstats))
		copy(reforgedSubstats, projSubstats)

		if level == 85 {
			for i, sub := range reforgedSubstats {
				incrMap, ok := ReforgeIncrements[sub.Type]
				if ok {
					rIndex := sub.Rolls
					if rIndex > 6 {
						rIndex = 6
					}
					if rIndex >= 0 && rIndex < len(incrMap) {
						reforgedSubstats[i].Value += incrMap[rIndex]
					}
				}
			}
		}

		// Calculate WSS and Hero Scores for the scenario
		scenarioScore := GetWSSScore(Gear{Substats: projSubstats})
		if level == 85 {
			scenarioScore = GetWSSScore(Gear{Substats: reforgedSubstats})
		}

		heroScores := make(map[string]float64)
		// fit score
		fit := GetHeroWeightedScoreForSubstats(projSubstats, profile, baseStats)
		// Add main stat contribution
		mainSavKey := getSavKey(gear.Main.Type)
		prio := profile.Priorities[mainSavKey]
		mainWeight := getHeroWeightMultiplier(prio)
		normMain := NormalizeStatValue(gear.Main.Type, gear.Main.Value, baseStats)
		if gear.Reforged || level == 85 { // model reforged main
			normMain = NormalizeStatValue(gear.Main.Type, gear.Main.Value+MainStatReforgeBonuses[gear.Main.Type], baseStats)
		}
		fit += normMain * mainWeight
		heroScores[profile.HeroID] = fit

		trace.Scenarios = append(trace.Scenarios, ProjectionScenario{
			ScenarioName: name,
			Substats:     projSubstats,
			Reforged:     reforgedSubstats,
			Score:        scenarioScore,
			HeroScores:   heroScores,
		})
	}

	return trace
}

func getHeroWeightMultiplier(prio int) float64 {
	switch prio {
	case 1:
		return 1.0
	case 2:
		return 2.5
	case 3:
		return 5.0
	default:
		return 0.0
	}
}
