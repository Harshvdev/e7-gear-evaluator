package main

// EvaluateLayer7 constructs the final GearTrace audit trace
func EvaluateLayer7(
	gearID string,
	speedCheckState string,
	l0 L0Trace,
	l1 L1Trace,
	l2 L2Trace,
	l3 L3Trace,
	l4 L4Trace,
	l5 map[string]L5Trace, // hero_id -> L5Trace
	l6 L6Trace,
	profiles []HeroProfile,
) GearTrace {
	// Construct reasons codes
	reasonCodes := []string{}

	// If L1 has violations, they are reasons
	for _, v := range l1.Violations {
		reasonCodes = append(reasonCodes, v)
	}

	// If L2 universally discarded, add reason
	if l2.Fired {
		reasonCodes = append(reasonCodes, "L2_FLAT_MAIN_NO_SPEED")
	}

	// L6 Verdict reason codes
	if l6.Verdict == "KEEP_ENHANCE" {
		reasonCodes = append(reasonCodes, "L6_KEEP_SUITABLE")
	} else if l6.Verdict == "SALVAGE_MOD" {
		reasonCodes = append(reasonCodes, "L6_SALVAGEABLE_MOD")
	} else if l6.Verdict == "SPEED_VAULT" {
		reasonCodes = append(reasonCodes, "L6_SPEED_VAULTED")
	} else if l6.Verdict == "SELL_EXTRACT" && len(reasonCodes) == 0 {
		reasonCodes = append(reasonCodes, "L6_NO_HERO_MATCH")
	}

	// For the L5 trace inside the GearTrace, we serialize the winner's trace or general trace
	var winnerL5 L5Trace
	if l6.WinnerHero != "" {
		for _, profile := range profiles {
			if profile.HeroName == l6.WinnerHero && profile.BuildRank == l6.WinnerBuild {
				winnerL5 = l5[profile.HeroID]
				break
			}
		}
	}

	// Fallback to first L5 trace if winnerL5 is empty
	if winnerL5.CurrentWSS == 0 && len(l5) > 0 {
		for _, t := range l5 {
			winnerL5 = t
			break
		}
	}

	trace := GearTrace{
		GearID:          gearID,
		TableVersion:    "1.0.0", // geardata version
		SpeedCheckState: speedCheckState,
		L0:              l0,
		L1:              l1,
		L2:              l2,
		L3:              l3,
		L4:              l4,
		L5:              winnerL5,
		L6:              l6,
		Verdict:         l6.Verdict,
		ReasonCodes:     reasonCodes,
	}

	return trace
}
