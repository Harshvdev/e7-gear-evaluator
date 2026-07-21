package main

import (
	"strings"
)


// EvaluateLayer5 calculates current and projected substat envelopes at +15, including reforge increments.
func EvaluateLayer5(gear Gear, profile HeroProfile, baseStats HeroBaseStats) L5Trace {
	trace := L5Trace{
		CurrentWSS:        GetWSSScore(gear),
		CurrentHeroScore:  make(map[string]float64),
		CurrentHeroFitPct: make(map[string]float64),
		Scenarios:         []ProjectionScenario{},
	}

	rawFit, _, currentFitPct, _, _ := CalculateHeroFit(gear, profile, baseStats)
	trace.CurrentHeroScore[profile.HeroID] = rawFit
	trace.CurrentHeroFitPct[profile.HeroID] = currentFitPct

	// Step 1: Calculate remaining rolls
	enhance := gear.Enhance
	rarity := gear.Rarity
	level := gear.Level

	rarityKey := "Epic"
	if strings.EqualFold(rarity, "Heroic") {
		rarityKey = "Heroic"
	}

	var randRolls int
	var forced4thRolls int
	var forced4thExists bool

	if strings.EqualFold(rarity, "Epic") {
		randRolls = 5 - enhance/3
		if randRolls < 0 {
			randRolls = 0
		}
	} else {
		if enhance < 12 {
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
		var existingSubTypes []string
		for _, sub := range gear.Substats {
			existingSubTypes = append(existingSubTypes, sub.Type)
		}
		legal4thStats = GetLegalSubstats(gear.Slot, gear.Main.Type, existingSubTypes)
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
				expectedRandPerSub := float64(randRolls) / float64(len(gear.Substats))
				for i := range projSubstats {
					avgRoll := GetAverageRoll(level, rarityKey, projSubstats[i].Type)
					projSubstats[i].Value += expectedRandPerSub * avgRoll
				}
				for r := 0; r < randRolls; r++ {
					idx := r % len(projSubstats)
					projSubstats[idx].Rolls++
				}
			case "best":
				bestIdx := 0
				bestWeight := -999.0
				for i, sub := range projSubstats {
					savKey := getSavKey(sub.Type)
					w := GetPriorityWeight(profile.Priorities[savKey])
					if w > bestWeight {
						bestWeight = w
						bestIdx = i
					}
				}
				ranges, _ := SubstatRollRanges[level][rarityKey]
				maxRoll := ranges[projSubstats[bestIdx].Type][1]
				projSubstats[bestIdx].Value += float64(randRolls) * maxRoll
				projSubstats[bestIdx].Rolls += randRolls
			case "worst":
				worstIdx := 0
				worstWeight := 999.0
				for i, sub := range projSubstats {
					savKey := getSavKey(sub.Type)
					w := GetPriorityWeight(profile.Priorities[savKey])
					if w < worstWeight {
						worstWeight = w
						worstIdx = i
					}
				}
				ranges, _ := SubstatRollRanges[level][rarityKey]
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
				bestType := legal4thStats[0]
				bestWeight := -999.0
				for _, s := range legal4thStats {
					w := GetPriorityWeight(profile.Priorities[getSavKey(s)])
					if w > bestWeight {
						bestWeight = w
						bestType = s
					}
				}
				chosen4thType = bestType
				val = float64(forced4thRolls) * GetAverageRoll(level, rarityKey, chosen4thType)
			case "best":
				bestType := legal4thStats[0]
				bestWeight := -999.0
				for _, s := range legal4thStats {
					w := GetPriorityWeight(profile.Priorities[getSavKey(s)])
					if w > bestWeight {
						bestWeight = w
						bestType = s
					}
				}
				ranges, _ := SubstatRollRanges[level][rarityKey]
				maxRoll := ranges[bestType][1]
				chosen4thType = bestType
				val = float64(forced4thRolls) * maxRoll
			case "worst":
				worstType := legal4thStats[0]
				worstWeight := 999.0
				for _, s := range legal4thStats {
					w := GetPriorityWeight(profile.Priorities[getSavKey(s)])
					if w < worstWeight {
						worstWeight = w
						worstType = s
					}
				}
				ranges, _ := SubstatRollRanges[level][rarityKey]
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
			idx := 3
			subType := projSubstats[idx].Type
			switch name {
			case "expected":
				projSubstats[idx].Value += float64(forced4thRolls) * GetAverageRoll(level, rarityKey, subType)
			case "best":
				ranges, _ := SubstatRollRanges[level][rarityKey]
				maxRoll := ranges[subType][1]
				projSubstats[idx].Value += float64(forced4thRolls) * maxRoll
			case "worst":
				ranges, _ := SubstatRollRanges[level][rarityKey]
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

		// Create projected gear object at +15 main stat value
		mainValAt15 := MainStatValuesAtPlus15[level][gear.Main.Type]
		if mainValAt15 <= 0 {
			mainValAt15 = gear.Main.Value
		}
		projGear := Gear{
			Slot:     gear.Slot,
			Level:    gear.Level,
			Rarity:   gear.Rarity,
			Enhance:  15,
			Main:     MainStat{Type: gear.Main.Type, Value: mainValAt15},
			Substats: projSubstats,
		}
		if level == 85 {
			projGear.Main.Value += MainStatReforgeBonuses[gear.Main.Type]
			projGear.Substats = reforgedSubstats
		}

		projRawFit, _, projFitPct, _, _ := CalculateHeroFit(projGear, profile, baseStats)

		heroScores := make(map[string]float64)
		heroFitPcts := make(map[string]float64)
		heroScores[profile.HeroID] = projRawFit
		heroFitPcts[profile.HeroID] = projFitPct

		trace.Scenarios = append(trace.Scenarios, ProjectionScenario{
			ScenarioName: name,
			Substats:     projSubstats,
			Reforged:     reforgedSubstats,
			Score:        scenarioScore,
			HeroScores:   heroScores,
			HeroFitPcts:  heroFitPcts,
		})
	}

	// Class Outcome Probabilities over remaining rolls (§4 & §5)
	// Expected scenario gets 50% probability, best 25%, worst 25% (or continuous envelope mapping)
	expFit := trace.Scenarios[0].HeroFitPcts[profile.HeroID]
	bestFit := trace.Scenarios[1].HeroFitPcts[profile.HeroID]
	worstFit := trace.Scenarios[2].HeroFitPcts[profile.HeroID]

	// Classify outcomes across scenarios
	var pCore, pUsable, pMarginal, pReject float64
	scenWeights := []float64{0.50, 0.25, 0.25}
	scenFits := []float64{expFit, bestFit, worstFit}

	for idx, fitVal := range scenFits {
		w := scenWeights[idx]
		if fitVal >= 70.0 {
			pCore += w
		} else if fitVal >= 45.0 {
			pUsable += w
		} else if fitVal >= 20.0 {
			pMarginal += w
		} else {
			pReject += w
		}
	}

	pGood := pCore + pUsable
	evFinal := pCore*85.0 + pUsable*57.0 + pMarginal*32.0 + pReject*0.0

	trace.PCore = pCore
	trace.PUsable = pUsable
	trace.PMarginal = pMarginal
	trace.PReject = pReject
	trace.PGood = pGood
	trace.EVFinal = evFinal

	// Step 4: Enhancement Controller & Hard Caps (§10)
	config := DefaultGlobalConfig()
	missBudget := config.MissBudget
	effFloor := config.EffFloor

	// Calculate observed step classes
	goodCount := 0
	wastedCount := 0
	steerableTaken := 0
	steerableRemaining := 0

	if strings.EqualFold(rarity, "Epic") {
		steerableTaken = enhance / 3
		steerableRemaining = (15 - enhance) / 3
	} else {
		if enhance <= 9 {
			steerableTaken = enhance / 3
			steerableRemaining = (9 - enhance) / 3
		} else {
			steerableTaken = 3
			steerableRemaining = 0
		}
	}

	for _, sub := range gear.Substats {
		savKey := getSavKey(sub.Type)
		prio := profile.Priorities[savKey]
		w := GetPriorityWeight(prio)

		// For existing sub rolls
		if sub.Rolls > 0 {
			avgValPerRoll := sub.Value / float64(sub.Rolls)
			stepClass := ClassifyStep("HERO", sub.Type, avgValPerRoll, prio, level, rarity)
			switch stepClass {
			case "HARM", "WASTED":
				wastedCount += sub.Rolls
			case "GOOD":
				goodCount += sub.Rolls
			}
		} else if w < 0 {
			// Substat present with negative weight counts as HARM
			wastedCount++
		}
	}

	// Evaluate Hard Caps (C1 - C4)
	c1Pass := wastedCount <= missBudget

	effBest := 1.0
	if steerableTaken+steerableRemaining > 0 {
		effBest = float64(goodCount+steerableRemaining) / float64(steerableTaken+steerableRemaining)
	}
	c2Pass := effBest >= effFloor

	// C3: Path exists (best-case finish >= tier required)
	tierRequiredClass := "USABLE"
	if strings.EqualFold(profile.RosterTier, "catalog") {
		tierRequiredClass = "CORE"
	}
	c3Pass := false
	if tierRequiredClass == "CORE" {
		c3Pass = bestFit >= 70.0
	} else {
		c3Pass = bestFit >= 45.0
	}

	// C4: No Conjunctive Hope
	// On Heroic gear at +9, if usable finish requires forced 4th sub to unlock a specific stat and roll high, check if unsatisfied
	c4Pass := true
	if strings.EqualFold(rarity, "Heroic") && enhance >= 9 && !forced4thExists {
		// Check if current fit% is sub-usable and 4th sub must be a specific core stat
		if currentFitPct < 45.0 && len(legal4thStats) > 0 {
			best4thPrio := -999
			for _, s := range legal4thStats {
				p := profile.Priorities[getSavKey(s)]
				if p > best4thPrio {
					best4thPrio = p
				}
			}
			if best4thPrio < 4 {
				c4Pass = false
			}
		}
	}

	hardCaps := HardCapsState{
		C1MissBudget:        c1Pass,
		C2MarginalEff:       c2Pass,
		C3PathExists:        c3Pass,
		C4NoConjunctiveHope: c4Pass,
	}

	action := "CONTINUE"
	reason := ""

	if enhance >= 15 {
		action = "STOP"
		reason = "ALREADY_DONE"
	} else if !c1Pass {
		action = "STOP"
		reason = "MISS_BUDGET_EXCEEDED"
	} else if !c2Pass {
		action = "STOP"
		reason = "MARGINAL_EFFICIENCY_LOW"
	} else if !c3Pass {
		action = "STOP"
		reason = "PATH_EXISTS"
	} else if !c4Pass {
		action = "STOP"
		reason = "CONJUNCTIVE_HOPE"
	} else if pGood < profile.RiskTolerance {
		action = "STOP"
		reason = "LOW_ODDS"
	} else if evFinal < 45.0 {
		action = "STOP"
		reason = "BELOW_FLOOR"
	}

	rollSeq := "Hit core stats on remaining rolls"
	if bestFit >= 70.0 {
		rollSeq = "All remaining rolls hit top core stats"
	}

	trace.StopCard = &StopCard{
		EnhanceAtPoint:      enhance,
		ObservedFitPct:      currentFitPct,
		PGood:               pGood,
		EVFinal:             evFinal,
		Recommended:         action,
		Reason:              reason,
		HardCaps:            hardCaps,
		RollSequenceForCore: rollSeq,
	}

	return trace
}

func getHeroWeightMultiplier(prio int) float64 {
	return GetPriorityWeight(prio)
}
