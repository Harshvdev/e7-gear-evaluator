package main

import (
	"fmt"
)

// EvaluateLayer2 evaluates whether the gear should be universally discarded.
func EvaluateLayer2(gear Gear) L2Trace {
	trace := L2Trace{
		Rule:   "L2: Universal Discard Rule",
		Fired:  false,
		Detail: "",
	}

	// Rule: Right-side (Necklace/Ring/Boots) with a flat main stat (Attack/Defense/Health)
	// and NO Speed substat is discarded unconditionally.
	if isRightSideSlot(gear.Slot) && isFlatMainStat(gear.Main.Type) {
		hasSpeed := false
		for _, sub := range gear.Substats {
			if sub.Type == "Speed" {
				hasSpeed = true
				break
			}
		}

		if !hasSpeed {
			trace.Fired = true
			trace.Detail = fmt.Sprintf(
				"Right-side %s with flat main stat %s and no Speed substat. universally worthless.",
				gear.Slot, STAT_LABELS[gear.Main.Type],
			)
		}
	}

	return trace
}
