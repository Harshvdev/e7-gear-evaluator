package main

// EvaluateLayer3 implements the speed check toggle rule with optional speed_vault_min threshold.
func EvaluateLayer3(gear Gear, speedCheckToggle string, speedVaultMin ...float64) L3Trace {
	trace := L3Trace{
		Toggle:     speedCheckToggle,
		Tagged:     false,
		SpeedValue: 0.0,
	}

	minThreshold := 0.0
	if len(speedVaultMin) > 0 {
		minThreshold = speedVaultMin[0]
	}

	// Find Speed substat
	var hasSpeed bool
	var speedVal float64
	for _, sub := range gear.Substats {
		if sub.Type == "Speed" {
			hasSpeed = true
			speedVal = sub.Value
			break
		}
	}

	if hasSpeed {
		trace.SpeedValue = speedVal
		// Default to ON if empty/unspecified
		toggleOn := speedCheckToggle == "" || speedCheckToggle == "ON" || speedCheckToggle == "on"
		
		if toggleOn && speedVal >= minThreshold {
			trace.Tagged = true
		}
	}

	return trace
}
