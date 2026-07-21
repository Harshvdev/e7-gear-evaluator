package main

import (
	"fmt"
	"strings"
)

// EvaluateLayer6 decides the final verdict and salvage plan for the gear per Q-E7-Architecture §6.
func EvaluateLayer6(
	gear Gear,
	l3 L3Trace,
	l4Results []L4HeroGateResult,
	l5Scenarios map[string]L5Trace, // hero_id -> L5Trace
	profiles []HeroProfile,
	baseStats map[string]HeroBaseStats, // hero_id -> base stats
	modBudget string, // "none" | "surplus"
	reforgeBudget string, // "none" | "surplus"
) L6Trace {
	if modBudget == "" {
		modBudget = "none"
	}
	if reforgeBudget == "" {
		reforgeBudget = "none"
	}

	trace := L6Trace{
		Verdict:       "SELL_EXTRACT",
		WinnerHero:    "",
		WinnerBuild:   0,
		RunnerUps:     []string{},
		SalvageDetail: nil,
		ModBudget:     modBudget,
		ReforgeBudget: reforgeBudget,
	}

	bestFitPct := -999.0
	var winningProfile *HeroProfile
	var winningClass string

	// 1. Check SALVAGE_MOD (§6 Decisive Rule: as-is class MUST be >= USABLE, modding raises to CORE, mod_budget == surplus)
	if modBudget == "surplus" {
		var salvageWinnerProfile *HeroProfile
		var salvagePlan *SalvagePlan
		bestSalvageFitPct := -999.0

		for i, res := range l4Results {
			profile := profiles[i]
			heroBase := baseStats[profile.HeroID]

			if !res.GateResults[0] || !res.GateResults[1] {
				continue
			}

			l5Trace, hasL5 := l5Scenarios[profile.HeroID]
			if !hasL5 || len(l5Trace.Scenarios) == 0 {
				continue
			}

			expFitPct := l5Trace.Scenarios[0].HeroFitPcts[profile.HeroID]
			// Axiom 5: As-is projected fit-class MUST be >= USABLE (expFitPct >= 45.0)
			if expFitPct < 45.0 {
				continue // Sub-USABLE base: NEVER salvage! (Modding cannot rescue a sub-USABLE base)
			}

			expectedScen := &l5Trace.Scenarios[0]
			if len(expectedScen.Substats) != 4 {
				continue
			}

			projSubs := expectedScen.Substats
			if gear.Level == 85 && len(expectedScen.Reforged) == 4 {
				projSubs = expectedScen.Reforged
			}

			// Find dead/harmful substat (prio <= 0)
			deadSubIdx := -1
			var deadSub Substat

			for idx, sub := range projSubs {
				savKey := getSavKey(sub.Type)
				if profile.Priorities[savKey] <= 0 {
					deadSub = sub
					deadSubIdx = idx
					break
				}
			}

			if deadSubIdx >= 0 && !deadSub.Modified {
				var other3 []string
				for idx, sub := range projSubs {
					if idx != deadSubIdx {
						other3 = append(other3, sub.Type)
					}
				}
				legalReplacements := GetLegalSubstats(gear.Slot, gear.Main.Type, other3)

				rarityKey := "Epic"
				if strings.EqualFold(gear.Rarity, "Heroic") {
					rarityKey = "Heroic"
				}

				for _, repType := range legalReplacements {
					repSavKey := getSavKey(repType)
					if profile.Priorities[repSavKey] >= 2 {
						ranges, hasRanges := SubstatRollRanges[gear.Level][rarityKey]
						if !hasRanges {
							continue
						}
						minRoll := ranges[repType][0]
						deadSubRolls := deadSub.Rolls
						if deadSubRolls < 1 {
							deadSubRolls = 1
						}
						moddedVal := float64(deadSubRolls) * minRoll

						moddedSubs := make([]Substat, len(projSubs))
						copy(moddedSubs, projSubs)
						moddedSubs[deadSubIdx] = Substat{
							Type:  repType,
							Value: moddedVal,
							Rolls: deadSub.Rolls,
						}

						moddedGear := Gear{
							Slot:     gear.Slot,
							Level:    gear.Level,
							Rarity:   gear.Rarity,
							Enhance:  15,
							Main:     gear.Main,
							Substats: moddedSubs,
						}

						rescoredRawFit, _, rescoredFitPct, rescoredClass, _ := CalculateHeroFit(moddedGear, profile, heroBase)

						// Modding must raise the piece to CORE class (fitPct >= 70.0)
						if rescoredClass == "CORE" || rescoredFitPct >= 70.0 {
							if rescoredFitPct > bestSalvageFitPct {
								bestSalvageFitPct = rescoredFitPct
								salvageWinnerProfile = &profiles[i]
								salvagePlan = &SalvagePlan{
									DeadSubStat:    deadSub.Type,
									TargetStat:     repType,
									ExpectedValue:  moddedVal,
									RescoredFit:    rescoredRawFit,
									RescoredFitPct: rescoredFitPct,
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
	}

	// 2. Check KEEP_ENHANCE (as-is projected +15 fit-class in {CORE, USABLE})
	for i, res := range l4Results {
		if res.GateResults[0] && res.GateResults[1] && res.GateResults[2] {
			profile := profiles[i]
			l5Trace, hasL5 := l5Scenarios[profile.HeroID]
			if !hasL5 || len(l5Trace.Scenarios) == 0 {
				continue
			}

			expFitPct := l5Trace.Scenarios[0].HeroFitPcts[profile.HeroID]
			expFitClass := "REJECT"
			if expFitPct >= 70.0 {
				expFitClass = "CORE"
			} else if expFitPct >= 45.0 {
				expFitClass = "USABLE"
			} else if expFitPct >= 20.0 {
				expFitClass = "MARGINAL"
			}

			// Controller stop card action check
			if l5Trace.StopCard != nil && l5Trace.StopCard.Recommended == "STOP" && gear.Enhance < 15 {
				continue
			}

			if expFitClass == "CORE" || expFitClass == "USABLE" {
				if expFitPct > bestFitPct {
					bestFitPct = expFitPct
					winningProfile = &profiles[i]
					winningClass = expFitClass
				}
			}
		}
	}

	if winningProfile != nil {
		trace.Verdict = "KEEP_ENHANCE"
		trace.WinnerHero = winningProfile.HeroName
		trace.WinnerBuild = winningProfile.BuildRank
		for i, res := range l4Results {
			if res.Pass && profiles[i].HeroID != winningProfile.HeroID {
				trace.RunnerUps = append(trace.RunnerUps, fmt.Sprintf("%s (Build %d)", profiles[i].HeroName, profiles[i].BuildRank))
			}
		}

		// Optional Reforge Tag upgrade if reforge_budget == surplus
		if gear.Level == 85 && reforgeBudget == "surplus" {
			l5Trace := l5Scenarios[winningProfile.HeroID]
			if len(l5Trace.Scenarios) > 0 {
				refFitPct := l5Trace.Scenarios[0].HeroFitPcts[winningProfile.HeroID]
				if refFitPct > bestFitPct && winningClass != "CORE" {
					trace.Verdict = "REFORGE_TAG"
				}
			}
		}
		return trace
	}

	// 3. Check KEEP_MARGINAL (as-is projected fit-class MARGINAL on primary hero)
	for i, res := range l4Results {
		if res.GateResults[0] && res.GateResults[1] && res.GateResults[2] {
			profile := profiles[i]
			if res.RosterTier != "primary" {
				continue
			}
			l5Trace, hasL5 := l5Scenarios[profile.HeroID]
			if !hasL5 || len(l5Trace.Scenarios) == 0 {
				continue
			}
			expFitPct := l5Trace.Scenarios[0].HeroFitPcts[profile.HeroID]
			if expFitPct >= 20.0 && expFitPct < 45.0 {
				if expFitPct > bestFitPct {
					bestFitPct = expFitPct
					winningProfile = &profiles[i]
				}
			}
		}
	}

	if winningProfile != nil {
		trace.Verdict = "KEEP_MARGINAL"
		trace.WinnerHero = winningProfile.HeroName
		trace.WinnerBuild = winningProfile.BuildRank
		return trace
	}

	// 4. Check SPEED_VAULT next
	if l3.Tagged {
		trace.Verdict = "SPEED_VAULT"
		return trace
	}

	// 5. Otherwise SELL_EXTRACT
	trace.Verdict = "SELL_EXTRACT"
	return trace
}
