package main

import (
	"fmt"
)

// EvaluateLayer6 decides the final verdict and salvage plan for the gear.
func EvaluateLayer6(
	gear Gear,
	l3 L3Trace,
	l4Results []L4HeroGateResult,
	l5Scenarios map[string]L5Trace, // hero_id -> L5Trace
	profiles []HeroProfile,
	baseStats map[string]HeroBaseStats, // hero_id -> base stats
) L6Trace {
	trace := L6Trace{
		Verdict:       "SELL_EXTRACT",
		WinnerHero:    "",
		WinnerBuild:   0,
		RunnerUps:     []string{},
		SalvageDetail: nil,
	}

	bestFitScore := -999.0
	var winningProfile *HeroProfile

	// Try KEEP_ENHANCE first
	for i, res := range l4Results {
		if res.Pass {
			profile := profiles[i]
			l5Trace := l5Scenarios[profile.HeroID]
			expectedFit := 0.0

			// Find "expected" scenario fit score
			for _, scen := range l5Trace.Scenarios {
				if scen.ScenarioName == "expected" {
					expectedFit = scen.HeroScores[profile.HeroID]
					break
				}
			}

			if expectedFit >= res.Threshold {
				if expectedFit > bestFitScore {
					bestFitScore = expectedFit
					winningProfile = &profiles[i]
				}
			}
		}
	}

	if winningProfile != nil {
		trace.Verdict = "KEEP_ENHANCE"
		trace.WinnerHero = winningProfile.HeroName
		trace.WinnerBuild = winningProfile.BuildRank
		// Collect runner-ups
		for i, res := range l4Results {
			if res.Pass && profiles[i].HeroID != winningProfile.HeroID {
				trace.RunnerUps = append(trace.RunnerUps, fmt.Sprintf("%s (Build %d)", profiles[i].HeroName, profiles[i].BuildRank))
			}
		}
		return trace
	}

	// Try SALVAGE_MOD next
	var salvageWinnerProfile *HeroProfile
	var salvagePlan *SalvagePlan
	bestSalvageFit := -999.0

	for i, res := range l4Results {
		profile := profiles[i]
		heroBase := baseStats[profile.HeroID]

		// Must match basic legality and set gates, but might fail the stats gate
		if !res.GateResults[0] || !res.GateResults[1] {
			continue
		}

		// The gear must have 4 substats for salvage mod check to be meaningful
		if len(gear.Substats) != 4 {
			continue
		}

		// Check if exactly 3 of 4 substats fit the hero build (prio >= 1)
		fittingCount := 0
		var deadSub Substat
		deadSubIdx := -1

		for idx, sub := range gear.Substats {
			savKey := getSavKey(sub.Type)
			if profile.Priorities[savKey] >= 1 {
				fittingCount++
			} else {
				deadSub = sub
				deadSubIdx = idx
			}
		}

		if fittingCount == 3 && deadSubIdx >= 0 && !deadSub.Modified && deadSub.Rolls <= 1 {
			// Find allowed replacements for this slot that fit the hero profile
			allowed := SubstatsBySlot[gear.Slot]
			for repType := range allowed {
				// Must not collide with main stat
				if repType == gear.Main.Type {
					continue
				}
				// Must not collide with other 3 substats
				collides := false
				for idx, sub := range gear.Substats {
					if idx != deadSubIdx && sub.Type == repType {
						collides = true
						break
					}
				}
				if collides {
					continue
				}

				// Must be a stat the hero cares about
				repSavKey := getSavKey(repType)
				if profile.Priorities[repSavKey] >= 1 {
					// Modded value: rolls * min roll value for that stat
					ranges, hasRanges := SubstatRollRanges[gear.Level][gear.Rarity]
					if !hasRanges {
						continue
					}
					minRoll := ranges[repType][0]
					moddedVal := float64(deadSub.Rolls) * minRoll

					// Rescore with the replaced stat
					moddedSubs := make([]Substat, len(gear.Substats))
					copy(moddedSubs, gear.Substats)
					moddedSubs[deadSubIdx] = Substat{
						Type:  repType,
						Value: moddedVal,
						Rolls: deadSub.Rolls,
					}

					// Get projected expected value with this replaced stat
					// For simple rescored fit, we just evaluate the fit score of these substats
					rescoredFit := GetHeroWeightedScoreForSubstats(moddedSubs, profile, heroBase)

					// Add main stat contribution
					mainSavKey := getSavKey(gear.Main.Type)
					prio := profile.Priorities[mainSavKey]
					mainWeight := getHeroWeightMultiplier(prio)
					normMain := NormalizeStatValue(gear.Main.Type, gear.Main.Value, heroBase)
					rescoredFit += normMain * mainWeight

					// We check if it clears the threshold
					if rescoredFit >= res.Threshold {
						if rescoredFit > bestSalvageFit {
							bestSalvageFit = rescoredFit
							salvageWinnerProfile = &profiles[i]
							salvagePlan = &SalvagePlan{
								DeadSubStat:   deadSub.Type,
								TargetStat:    repType,
								ExpectedValue: moddedVal,
								RescoredFit:   rescoredFit,
							}
						}
					}
				}
			}
		}
	}

	if salvageWinnerProfile != nil {
		trace.Verdict = "SALVAGE_MOD"
		trace.WinnerHero = salvageWinnerProfile.HeroName
		trace.WinnerBuild = salvageWinnerProfile.BuildRank
		trace.SalvageDetail = salvagePlan
		return trace
	}

	// Try SPEED_VAULT next
	if l3.Tagged {
		trace.Verdict = "SPEED_VAULT"
		return trace
	}

	// Otherwise SELL_EXTRACT
	trace.Verdict = "SELL_EXTRACT"
	return trace
}
