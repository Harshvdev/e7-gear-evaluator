package main

import (
	"fmt"
	"math"
)

// RunEvaluationPipeline orchestrates layers L0 through L7.
func RunEvaluationPipeline(
	gear Gear,
	speedCheckToggle string,
	modBudget string,
	reforgeBudget string,
	excludedBuilds map[string][]int,
	customProfiles map[string]map[int]HeroProfile,
) EvaluationResponse {
	// L0: Ingest & Normalize
	normalizedGear, l0Trace := EvaluateLayer0(gear)

	// L1: Legality Check
	l1Trace := EvaluateLayer1(normalizedGear, l0Trace)

	// L2: Universal Discard
	l2Trace := EvaluateLayer2(normalizedGear)

	// L3: Speed Check Toggle
	l3Trace := EvaluateLayer3(normalizedGear, speedCheckToggle)

	// Early check: if legality check fails or universal discard fires, terminate early
	hasL1Violations := len(l1Trace.Violations) > 0
	l2Fired := l2Trace.Fired

	if hasL1Violations || l2Fired {
		verdict := "Sell"
		l6Verdict := "SELL_EXTRACT"
		if l2Fired {
			verdict = "Discard"
			l6Verdict = "DISCARD_WORTHLESS"
		}
		
		// Construct early trace
		l6Trace := L6Trace{
			Verdict:       l6Verdict,
			ModBudget:     modBudget,
			ReforgeBudget: reforgeBudget,
		}
		
		l4Trace := L4Trace{PerHero: []L4HeroGateResult{}}
		l5TracesMap := make(map[string]L5Trace)
		
		trace := EvaluateLayer7(normalizedGear.ID, speedCheckToggle, l0Trace, l1Trace, l2Trace, l3Trace, l4Trace, l5TracesMap, l6Trace, []HeroProfile{})
		
		return EvaluationResponse{
			GearID:       gear.ID,
			Verdict:      verdict,
			SuitedBuilds: []L1SuitedBuild{},
			HeroDetails:  []L1HeroDetail{},
			Trace:        &trace,
		}
	}

	// Step 4: Compile profiles and load base stats
	var profiles []HeroProfile
	baseStatsMap := make(map[string]HeroBaseStats)

	for _, hero := range heroDatabase {
		normName := normalizeName(hero.Hero)
		// Check if hero has base stats
		base, ok := heroBaseStatsMap[normName]
		if !ok {
			// Fallback default base stats
			base = HeroBaseStats{
				Atk: 900, Def: 600, Hp: 6000, Spd: 110,
				Cc: 0.15, Cd: 1.5, Eff: 0, Res: 0,
			}
		}

		for _, build := range hero.Builds {
			if isBuildExcluded(hero.Hero, build.Rank, excludedBuilds) {
				continue
			}

			profile := CompileHeroProfile(hero.Hero, build)

			// Merge custom profile settings from request if present
			if customProfiles != nil {
				if heroCustom, exists := customProfiles[hero.Hero]; exists {
					if custom, ok := heroCustom[build.Rank]; ok {
						if len(custom.Priorities) > 0 {
							for k, val := range custom.Priorities {
								profile.Priorities[k] = val
							}
						}
						if custom.StatRanges != nil {
							if profile.StatRanges == nil {
								profile.StatRanges = make(map[string]StatBounds)
							}
							for k, val := range custom.StatRanges {
								profile.StatRanges[k] = val
							}
						}
						if len(custom.Sets) > 0 {
							profile.Sets = custom.Sets
						}
						if custom.MinQuality != nil {
							profile.MinQuality = custom.MinQuality
						}
						if custom.WeightMode != "" {
							profile.WeightMode = custom.WeightMode
						}
						if custom.RosterTier != "" {
							profile.RosterTier = custom.RosterTier
						}
						if custom.RiskTolerance > 0 {
							profile.RiskTolerance = custom.RiskTolerance
						}
						if len(custom.AccessoryMains) > 0 {
							profile.AccessoryMains = custom.AccessoryMains
						}
					}
				}
			}

			profiles = append(profiles, profile)
			baseStatsMap[profile.HeroID] = base
		}
	}

	// Step 5: Run Layer 4 (Hero matching) & Layer 5 (Projection)
	l4Trace := L4Trace{PerHero: []L4HeroGateResult{}}
	l5TracesMap := make(map[string]L5Trace)

	for _, profile := range profiles {
		base := baseStatsMap[profile.HeroID]
		// L4
		l4Res := EvaluateLayer4(normalizedGear, profile, base)
		l4Trace.PerHero = append(l4Trace.PerHero, l4Res)
		// L5
		l5Res := EvaluateLayer5(normalizedGear, profile, base)
		l5TracesMap[profile.HeroID] = l5Res
	}

	// Step 6: Run Layer 6 (Decision & Salvage)
	l6Trace := EvaluateLayer6(normalizedGear, l3Trace, l4Trace.PerHero, l5TracesMap, profiles, baseStatsMap, modBudget, reforgeBudget)

	// Step 7: Run Layer 7 (Trace generation)
	trace := EvaluateLayer7(normalizedGear.ID, speedCheckToggle, l0Trace, l1Trace, l2Trace, l3Trace, l4Trace, l5TracesMap, l6Trace, profiles)

	// Map L6 verdict to legacy Verdict
	legacyGlobalVerdict := "Sell"
	switch l6Trace.Verdict {
	case "KEEP_ENHANCE", "KEEP_MARGINAL", "SALVAGE_MOD", "REFORGE_TAG":
		legacyGlobalVerdict = "Worthy"
	case "SPEED_VAULT":
		legacyGlobalVerdict = "Speed Check"
	default:
		legacyGlobalVerdict = "Sell"
	}

	// Construct legacy UI details
	heroDetailsMap := make(map[string]*L1HeroDetail)
	suitedBuilds := []L1SuitedBuild{}

	for i, profile := range profiles {
		l4Res := l4Trace.PerHero[i]
		normName := normalizeName(profile.HeroName)

		isBuildWorthy := l4Res.Pass
		// If salvage mod winner, mark it as worthy for legacy UI representation
		if l6Trace.Verdict == "SALVAGE_MOD" && l6Trace.WinnerHero == profile.HeroName && l6Trace.WinnerBuild == profile.BuildRank {
			isBuildWorthy = true
		}

		buildVerdict := "Sell"
		if isBuildWorthy {
			buildVerdict = "Worthy"
		}

		// Calculate WAS Pct (mapped to clamped FitPct in [-100, +100] per Q-E7-Architecture T-CLAMP)
		wasPct := math.Round(l4Res.FitPct*10) / 10

		// Missing Core
		missingCore := 0
		if profile.StatRanges != nil {
			for k, bounds := range profile.StatRanges {
				if bounds.Min != nil {
					hasStat := false
					if getSavKey(normalizedGear.Main.Type) == k {
						hasStat = true
					} else {
						for _, sub := range normalizedGear.Substats {
							if getSavKey(sub.Type) == k {
								hasStat = true
								break
							}
						}
					}
					if !hasStat {
						missingCore++
					}
				}
			}
		}

		// Landmines (any stats on gear with priority <= 0 in profile)
		landmines := []L1Landmine{}
		for _, sub := range normalizedGear.Substats {
			savKey := getSavKey(sub.Type)
			prio := profile.Priorities[savKey]
			if prio <= 0 {
				landmines = append(landmines, L1Landmine{
					StatType: sub.Type,
					Label:    STAT_LABELS[sub.Type],
					Weight:   0.0,
				})
			}
		}

		buildDetail := L1BuildDetail{
			Rank:        profile.BuildRank,
			Usage:       buildDetailUsage(profile.HeroName, profile.BuildRank),
			Sets:        profile.Sets[0],
			Verdict:     buildVerdict,
			WAS:         math.Round(l4Res.FitScore*10000) / 10000,
			WASPct:      wasPct,
			Threshold:   math.Round(l4Res.Threshold*10000) / 10000,
			MissingCore: missingCore,
			Landmines:   landmines,
			Sav:         profile.OriginalSav,
		}

		// Group by hero
		heroDetail, exists := heroDetailsMap[profile.HeroName]
		if !exists {
			heroDetail = &L1HeroDetail{
				HeroName:  profile.HeroName,
				Icon:      heroIcons[normName],
				Rarity:    heroRarities[normName],
				Role:      heroRoles[normName],
				Attribute: heroAttributes[normName],
				Builds:    []L1BuildDetail{},
			}
			heroDetailsMap[profile.HeroName] = heroDetail
		}
		heroDetail.Builds = append(heroDetail.Builds, buildDetail)

		if isBuildWorthy {
			landmineLabels := make([]string, len(landmines))
			for idx, lm := range landmines {
				landmineLabels[idx] = lm.Label
			}
			suitedBuilds = append(suitedBuilds, L1SuitedBuild{
				HeroName:  profile.HeroName,
				Rank:      profile.BuildRank,
				Usage:     buildDetail.Usage,
				Sets:      buildDetail.Sets,
				Landmines: landmineLabels,
				Rarity:    heroRarities[normName],
				Role:      heroRoles[normName],
				Attribute: heroAttributes[normName],
			})
		}
	}

	// Flatten heroDetails
	heroDetails := []L1HeroDetail{}
	for _, hd := range heroDetailsMap {
		heroDetails = append(heroDetails, *hd)
	}

	return EvaluationResponse{
		GearID:       gear.ID,
		Verdict:      legacyGlobalVerdict,
		SuitedBuilds: suitedBuilds,
		HeroDetails:  heroDetails,
		Trace:        &trace,
	}
}

// CompileHeroProfile compiles a Fribbels HeroBuild build statistics block into the canonical HeroProfile model.
func CompileHeroProfile(heroName string, build HeroBuild) HeroProfile {
	profile := HeroProfile{
		HeroID:        fmt.Sprintf("%s_Build_%d", heroName, build.Rank),
		HeroName:      heroName,
		BuildRank:     build.Rank,
		Selected:      true,
		RosterTier:    "primary",
		RiskTolerance: 0.5,
		StatRanges:    make(map[string]StatBounds),
		Sets:          [][]string{build.Sets},
		Priorities:    make(map[string]int),
		WeightMode:    "weighted",
		OriginalSav:   build.Sav,
	}

	// Determine max SAV to normalize weights
	maxSav := 0.0
	stats := []string{"atk", "def", "hp", "spd", "cc", "cd", "eff", "res"}
	for _, s := range stats {
		val := build.Sav.Get(s)
		if val > maxSav {
			maxSav = val
		}
	}

	// Convert SAV to signed priorities (-3..+5) per Q-E7-Architecture §2.2
	for _, s := range stats {
		val := build.Sav.Get(s)
		var prio int
		if maxSav > 0 {
			ratio := val / maxSav
			if ratio >= 0.85 {
				prio = 5 // essential
			} else if ratio >= 0.60 {
				prio = 4 // core
			} else if ratio >= 0.35 {
				prio = 3 // wanted
			} else if ratio >= 0.20 {
				prio = 2 // useful
			} else if ratio >= 0.05 {
				prio = 1 // filler
			} else {
				prio = 0 // neutral
			}
		} else {
			prio = 1
		}
		profile.Priorities[s] = prio
	}

	// Set min capability bounds for important core stats (priority >= 2)
	// Min values are scaled from the average build stats
	if build.AverageStats.Atk > 0 && profile.Priorities["atk"] >= 2 {
		profile.StatRanges["atk"] = StatBounds{Min: valPtr(build.AverageStats.Atk)}
	}
	if build.AverageStats.Def > 0 && profile.Priorities["def"] >= 2 {
		profile.StatRanges["def"] = StatBounds{Min: valPtr(build.AverageStats.Def)}
	}
	if build.AverageStats.Hp > 0 && profile.Priorities["hp"] >= 2 {
		profile.StatRanges["hp"] = StatBounds{Min: valPtr(build.AverageStats.Hp)}
	}
	if build.AverageStats.Spd > 0 && profile.Priorities["spd"] >= 2 {
		profile.StatRanges["spd"] = StatBounds{Min: valPtr(build.AverageStats.Spd)}
	}
	if build.AverageStats.Cc > 0 && profile.Priorities["cc"] >= 2 {
		profile.StatRanges["cc"] = StatBounds{Min: valPtr(build.AverageStats.Cc)}
	}
	if build.AverageStats.Cd > 0 && profile.Priorities["cd"] >= 2 {
		profile.StatRanges["cd"] = StatBounds{Min: valPtr(build.AverageStats.Cd)}
	}
	if build.AverageStats.Eff > 0 && profile.Priorities["eff"] >= 2 {
		profile.StatRanges["eff"] = StatBounds{Min: valPtr(build.AverageStats.Eff)}
	}
	if build.AverageStats.Res > 0 && profile.Priorities["res"] >= 2 {
		profile.StatRanges["res"] = StatBounds{Min: valPtr(build.AverageStats.Res)}
	}

	// Set CC max cap to 100 if CC is a priority
	if profile.Priorities["cc"] > 0 {
		ccBounds := profile.StatRanges["cc"]
		ccBounds.Max = valPtr(100.0)
		profile.StatRanges["cc"] = ccBounds
	}

	return profile
}

// isBuildExcluded checks if a build variant is marked as excluded.
func isBuildExcluded(hero string, rank int, excluded map[string][]int) bool {
	if excluded == nil {
		return false
	}
	ranks, found := excluded[hero]
	if !found {
		return false
	}
	for _, r := range ranks {
		if r == rank {
			return true
		}
	}
	return false
}

// Helpers
func valPtr(v float64) *float64 {
	return &v
}

func buildDetailUsage(hero string, rank int) float64 {
	for _, h := range heroDatabase {
		if h.Hero == hero {
			for _, b := range h.Builds {
				if b.Rank == rank {
					return b.Usage
				}
			}
		}
	}
	return 0.0
}
