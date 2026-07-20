package main

// ---------------------------------------------------------------------------
// normalizer.go — Shared scoring, normalization, and contribution calculations.
// ---------------------------------------------------------------------------

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
		return value * 1.142857
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

// GetHeroWeightedScore calculates the custom score for a gear piece against a specific hero profile.
// It applies the priority weights: 0 -> 0, 1 -> 1, 2 -> 2.5, 3 -> 5.
func GetHeroWeightedScore(gear Gear, profile HeroProfile, baseStats HeroBaseStats) float64 {
	var total float64

	// Substat weights
	for _, sub := range gear.Substats {
		savKey := getSavKey(sub.Type)
		prio := profile.Priorities[savKey]
		var weight float64
		switch prio {
		case 1:
			weight = 1.0
		case 2:
			weight = 2.5
		case 3:
			weight = 5.0
		default:
			weight = 0.0
		}

		if weight > 0 {
			normVal := NormalizeStatValue(sub.Type, sub.Value, baseStats)
			total += normVal * weight
		}
	}

	// Main stat weight (main stat also contributes to the hero fit score)
	mainSavKey := getSavKey(gear.Main.Type)
	prio := profile.Priorities[mainSavKey]
	var mainWeight float64
	switch prio {
	case 1:
		mainWeight = 1.0
	case 2:
		mainWeight = 2.5
	case 3:
		mainWeight = 5.0
	default:
		mainWeight = 0.0
	}
	if mainWeight > 0 {
		normVal := NormalizeStatValue(gear.Main.Type, gear.Main.Value, baseStats)
		total += normVal * mainWeight
	}

	return total
}

// GetHeroWeightedScoreForSubstats calculates the custom score using only a slice of substats (for projections)
func GetHeroWeightedScoreForSubstats(substats []Substat, profile HeroProfile, baseStats HeroBaseStats) float64 {
	var total float64
	for _, sub := range substats {
		savKey := getSavKey(sub.Type)
		prio := profile.Priorities[savKey]
		var weight float64
		switch prio {
		case 1:
			weight = 1.0
		case 2:
			weight = 2.5
		case 3:
			weight = 5.0
		default:
			weight = 0.0
		}

		if weight > 0 {
			normVal := NormalizeStatValue(sub.Type, sub.Value, baseStats)
			total += normVal * weight
		}
	}
	return total
}
