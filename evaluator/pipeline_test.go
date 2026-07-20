package main

import (
	"testing"
)

func TestGetExpectedRolls(t *testing.T) {
	tests := []struct {
		rarity   string
		enhance  int
		expected int
	}{
		{"Epic", 0, 4},
		{"Epic", 3, 5},
		{"Epic", 15, 9},
		{"Heroic", 0, 3},
		{"Heroic", 9, 6},
		{"Heroic", 12, 7},
		{"Heroic", 15, 8},
	}

	for _, tt := range tests {
		actual := GetExpectedRolls(tt.rarity, tt.enhance)
		if actual != tt.expected {
			t.Errorf("GetExpectedRolls(%s, %d) = %d; expected %d", tt.rarity, tt.enhance, actual, tt.expected)
		}
	}
}

func TestEvaluateLayer1_Legality(t *testing.T) {
	// 1. Collision test
	gearCollision := Gear{
		Slot:    "Necklace",
		Level:   85,
		Rarity:  "Epic",
		Enhance: 0,
		Main:    MainStat{Type: "AttackPercent", Value: 12.0},
		Substats: []Substat{
			{Type: "AttackPercent", Value: 6.0, Rolls: 1},
			{Type: "Speed", Value: 2.0, Rolls: 1},
			{Type: "CritHitChancePercent", Value: 4.0, Rolls: 1},
			{Type: "HealthPercent", Value: 5.0, Rolls: 1},
		},
	}

	l0Trace := L0Trace{
		RollReconstruction: map[string]int{"AttackPercent": 1, "Speed": 1, "CritHitChancePercent": 1, "HealthPercent": 1},
	}
	l1 := EvaluateLayer1(gearCollision, l0Trace)
	hasCollision := false
	for _, v := range l1.Violations {
		if v == "L1_MAIN_SUB_COLLISION" {
			hasCollision = true
		}
	}
	if !hasCollision {
		t.Errorf("Expected Main-Sub Collision violation, got violations: %v", l1.Violations)
	}

	// 2. Valid gear test
	gearValid := Gear{
		Slot:    "Necklace",
		Level:   85,
		Rarity:  "Epic",
		Enhance: 0,
		Main:    MainStat{Type: "CritHitDamagePercent", Value: 13.0},
		Substats: []Substat{
			{Type: "AttackPercent", Value: 6.0, Rolls: 1},
			{Type: "Speed", Value: 2.0, Rolls: 1},
			{Type: "CritHitChancePercent", Value: 4.0, Rolls: 1},
			{Type: "HealthPercent", Value: 5.0, Rolls: 1},
		},
	}

	l1Valid := EvaluateLayer1(gearValid, l0Trace)
	if len(l1Valid.Violations) > 0 {
		t.Errorf("Expected valid gear to have 0 violations, got: %v", l1Valid.Violations)
	}
}

func TestEvaluateLayer2_UniversalDiscard(t *testing.T) {
	// Right-side flat main with no speed
	gearDiscard := Gear{
		Slot:    "Necklace",
		Level:   85,
		Rarity:  "Epic",
		Main:    MainStat{Type: "Attack", Value: 100},
		Substats: []Substat{
			{Type: "AttackPercent", Value: 6.0, Rolls: 1},
			{Type: "HealthPercent", Value: 5.0, Rolls: 1},
			{Type: "DefensePercent", Value: 5.0, Rolls: 1},
			{Type: "CritHitChancePercent", Value: 4.0, Rolls: 1},
		},
	}

	l2 := EvaluateLayer2(gearDiscard)
	if !l2.Fired {
		t.Errorf("Expected universal discard to fire for flat main right side with no Speed sub")
	}

	// Right-side flat main WITH speed should survive L2
	gearSurvive := Gear{
		Slot:    "Necklace",
		Level:   85,
		Rarity:  "Epic",
		Main:    MainStat{Type: "Attack", Value: 100},
		Substats: []Substat{
			{Type: "Speed", Value: 4.0, Rolls: 1},
			{Type: "HealthPercent", Value: 5.0, Rolls: 1},
			{Type: "DefensePercent", Value: 5.0, Rolls: 1},
			{Type: "CritHitChancePercent", Value: 4.0, Rolls: 1},
		},
	}

	l2Survive := EvaluateLayer2(gearSurvive)
	if l2Survive.Fired {
		t.Errorf("Expected gear to survive universal discard since it has Speed substat")
	}
}

func TestEvaluateLayer3_SpeedCheck(t *testing.T) {
	gearWithSpeed := Gear{
		Slot:   "Necklace",
		Rarity: "Epic",
		Substats: []Substat{
			{Type: "Speed", Value: 3.0, Rolls: 1},
		},
	}

	l3On := EvaluateLayer3(gearWithSpeed, "ON")
	if !l3On.Tagged {
		t.Errorf("Expected speed check tag to be true when toggle is ON and Speed sub present")
	}

	l3Off := EvaluateLayer3(gearWithSpeed, "OFF")
	if l3Off.Tagged {
		t.Errorf("Expected speed check tag to be false when toggle is OFF")
	}
}
