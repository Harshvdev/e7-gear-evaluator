package main

// EvaluateLayer3 implements the speed check toggle rule.
func EvaluateLayer3(gear Gear, speedCheckToggle string) L3Trace {
	trace := L3Trace{
		Toggle:     speedCheckToggle,
		Tagged:     false,
		SpeedValue: 0.0,
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
		
		// If toggle is ON, tag it as Speed Vault.
		// Note: speed_vault_min threshold can be applied here.
		// Since user request doesn't pass a custom min, we default to 0 (any Speed vaults).
		if toggleOn && speedVal >= 0 {
			trace.Tagged = true
		}
	}

	return trace
}
