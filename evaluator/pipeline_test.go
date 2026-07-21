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

func TestEvaluateLayer4_SpeedMinRequirement(t *testing.T) {
	// 0-speed necklace (like the Protection set necklace)
	gearNoSpeed := Gear{
		Slot:    "Necklace",
		Level:   88,
		Rarity:  "Epic",
		Enhance: 0,
		Set:     "Protection",
		Main:    MainStat{Type: "AttackPercent", Value: 13.0},
		Substats: []Substat{
			{Type: "EffectivenessPercent", Value: 5.0, Rolls: 1},
			{Type: "Attack", Value: 42.0, Rolls: 1},
			{Type: "CritHitDamagePercent", Value: 7.0, Rolls: 1},
			{Type: "EffectResistancePercent", Value: 5.0, Rolls: 1},
		},
	}

	spdMin := 251.0
	atkMin := 3562.0
	cdMin := 280.0

	profileWithSpdMin := HeroProfile{
		HeroID:    "Zio_Build_1",
		HeroName:  "Zio",
		BuildRank: 1,
		Selected:  true,
		StatRanges: map[string]StatBounds{
			"spd": {Min: &spdMin},
			"atk": {Min: &atkMin},
			"cd":  {Min: &cdMin},
		},
		Priorities: map[string]int{
			"atk": 3, "cd": 3, "spd": 2, "cc": 1, "eff": 1, "def": 0, "hp": 0, "res": 0,
		},
		WeightMode: "strict",
	}

	baseStats := HeroBaseStats{
		Atk: 1255, Def: 683, Hp: 6266, Spd: 106,
		Cc: 0.15, Cd: 1.5, Eff: 0.18, Res: 0,
	}

	l4Res := EvaluateLayer4(gearNoSpeed, profileWithSpdMin, baseStats)

	// In strict mode, missing spd.min (serving 2/3 min stats) fails Gate 3
	if l4Res.GateResults[2] {
		t.Errorf("Expected Gate 3 to fail for 0-speed gear when spd.min target is set in strict mode, but it passed")
	}

	if l4Res.Pass {
		t.Errorf("Expected Layer 4 evaluation to FAIL for 0-speed gear when spd.min target is set in strict mode, but it passed")
	}
}

func TestEvaluateLayer1_OutOfRange(t *testing.T) {
	gearOutOfRange := Gear{
		Slot:    "Necklace",
		Level:   85,
		Rarity:  "Epic",
		Enhance: 0,
		Main:    MainStat{Type: "CritHitDamagePercent", Value: 13.0},
		Substats: []Substat{
			{Type: "Speed", Value: 99.0, Rolls: 1}, // Impossible Speed roll
			{Type: "AttackPercent", Value: 6.0, Rolls: 1},
			{Type: "CritHitChancePercent", Value: 4.0, Rolls: 1},
			{Type: "HealthPercent", Value: 5.0, Rolls: 1},
		},
	}

	l0Trace := L0Trace{
		RollReconstruction: map[string]int{"Speed": 1, "AttackPercent": 1, "CritHitChancePercent": 1, "HealthPercent": 1},
	}

	l1 := EvaluateLayer1(gearOutOfRange, l0Trace)
	hasOutOfRange := false
	for _, v := range l1.Violations {
		if v == "L1_VALUE_OUT_OF_RANGE" {
			hasOutOfRange = true
			break
		}
	}

	if !hasOutOfRange {
		t.Errorf("Expected L1_VALUE_OUT_OF_RANGE violation for Speed 99.0, got violations: %v", l1.Violations)
	}
}

func TestEvaluateLayer5_ArmorNoAttack(t *testing.T) {
	// Armor slot piece (+0 Heroic, 3 substats)
	gearArmor := Gear{
		Slot:    "Armor",
		Level:   85,
		Rarity:  "Heroic",
		Enhance: 0,
		Main:    MainStat{Type: "Defense", Value: 60.0},
		Substats: []Substat{
			{Type: "DefensePercent", Value: 6.0, Rolls: 1},
			{Type: "HealthPercent", Value: 6.0, Rolls: 1},
			{Type: "Speed", Value: 3.0, Rolls: 1},
		},
	}

	profile := HeroProfile{
		HeroID:    "Abigail_Build_1",
		HeroName:  "Abigail",
		BuildRank: 1,
		Selected:  true,
		Priorities: map[string]int{
			"def": 3, "hp": 3, "spd": 3, "cc": 1,
		},
		WeightMode: "weighted",
	}

	baseStats := HeroBaseStats{
		Atk: 984, Def: 632, Hp: 6148, Spd: 101,
	}

	l5 := EvaluateLayer5(gearArmor, profile, baseStats)

	for _, scen := range l5.Scenarios {
		for _, sub := range scen.Substats {
			if sub.Type == "Attack" || sub.Type == "AttackPercent" {
				t.Errorf("Scenario %s projected illegal Attack stat %s on Armor gear!", scen.ScenarioName, sub.Type)
			}
		}
	}
}

func TestEvaluateLayer4_InertProfile(t *testing.T) {
	gear := Gear{
		Slot:    "Weapon",
		Level:   85,
		Rarity:  "Epic",
		Enhance: 0,
		Main:    MainStat{Type: "Attack", Value: 100.0},
		Substats: []Substat{
			{Type: "AttackPercent", Value: 6.0, Rolls: 1},
			{Type: "Speed", Value: 3.0, Rolls: 1},
			{Type: "CritHitChancePercent", Value: 4.0, Rolls: 1},
			{Type: "CritHitDamagePercent", Value: 5.0, Rolls: 1},
		},
	}

	inertProfile := HeroProfile{
		HeroID:     "Inert_Build",
		HeroName:   "Inert",
		BuildRank:  1,
		Priorities: map[string]int{},
	}

	baseStats := HeroBaseStats{Atk: 1000, Def: 600, Hp: 6000, Spd: 100}

	res := EvaluateLayer4(gear, inertProfile, baseStats)
	if res.Pass {
		t.Errorf("Expected inert profile to fail Layer 4 matching, but it passed!")
	}
}

func TestEvaluateLayer6_SalvageModProjected(t *testing.T) {
	// +0 Epic necklace with 3 great stats + 1 dead stat (EffRes)
	gearSalvage := Gear{
		Slot:    "Necklace",
		Level:   85,
		Rarity:  "Epic",
		Enhance: 0,
		Set:     "Speed",
		Main:    MainStat{Type: "CritHitDamagePercent", Value: 13.0},
		Substats: []Substat{
			{Type: "AttackPercent", Value: 8.0, Rolls: 1},
			{Type: "CritHitChancePercent", Value: 5.0, Rolls: 1},
			{Type: "Speed", Value: 4.0, Rolls: 1},
			{Type: "EffectResistancePercent", Value: 5.0, Rolls: 1}, // Dead stat
		},
	}

	profile := HeroProfile{
		HeroID:    "DPS_Build_1",
		HeroName:  "DPS",
		BuildRank: 1,
		Selected:  true,
		Sets:      [][]string{{"Speed"}},
		Priorities: map[string]int{
			"atk": 3, "cd": 3, "cc": 3, "spd": 3, "res": 0,
		},
		WeightMode: "weighted",
	}

	baseStats := HeroBaseStats{Atk: 1200, Def: 600, Hp: 6000, Spd: 110, Cc: 0.15, Cd: 1.5}

	l4Res := EvaluateLayer4(gearSalvage, profile, baseStats)
	l5Res := EvaluateLayer5(gearSalvage, profile, baseStats)

	l6 := EvaluateLayer6(
		gearSalvage,
		L3Trace{},
		[]L4HeroGateResult{l4Res},
		map[string]L5Trace{profile.HeroID: l5Res},
		[]HeroProfile{profile},
		map[string]HeroBaseStats{profile.HeroID: baseStats},
		"surplus",
		"none",
	)

	if l6.Verdict != "KEEP_ENHANCE" && l6.Verdict != "SALVAGE_MOD" {
		t.Errorf("Expected +0 gear with 3 perfect subs + 1 dead sub to be KEEP_ENHANCE or SALVAGE_MOD, got: %s", l6.Verdict)
	}
}

func TestEvaluateLayer5_Heroic4thSubSanity(t *testing.T) {
	gearHeroic := Gear{
		Slot:    "Helmet",
		Level:   85,
		Rarity:  "Heroic",
		Enhance: 0,
		Main:    MainStat{Type: "Health", Value: 540.0},
		Substats: []Substat{
			{Type: "AttackPercent", Value: 6.0, Rolls: 1},
			{Type: "Speed", Value: 3.0, Rolls: 1},
			{Type: "CritHitChancePercent", Value: 4.0, Rolls: 1},
		},
	}

	profile := HeroProfile{
		HeroID:    "DPS_Heroic_1",
		HeroName:  "DPS",
		BuildRank: 1,
		Selected:  true,
		Priorities: map[string]int{
			"atk": 3, "spd": 3, "cc": 3, "cd": 3,
		},
		WeightMode: "weighted",
	}

	baseStats := HeroBaseStats{Atk: 1000, Def: 600, Hp: 6000, Spd: 100}
	l5 := EvaluateLayer5(gearHeroic, profile, baseStats)

	for _, scen := range l5.Scenarios {
		if len(scen.Substats) == 4 {
			fourthSub := scen.Substats[3]
			if fourthSub.Type == "AttackPercent" && fourthSub.Value > 20.0 {
				t.Errorf("Scenario %s calculated impossible AttackPercent 4th sub value: %.2f", scen.ScenarioName, fourthSub.Value)
			}
			if fourthSub.Type == "Speed" && fourthSub.Value > 10.0 {
				t.Errorf("Scenario %s calculated impossible Speed 4th sub value: %.2f", scen.ScenarioName, fourthSub.Value)
			}
		}
	}
}

func TestGetHeroBaseStatValue_EffResScaling(t *testing.T) {
	baseStats := HeroBaseStats{
		Atk: 1000, Def: 600, Hp: 6000, Spd: 110,
		Cc: 0.15, Cd: 1.5, Eff: 0.18, Res: 0.30,
	}

	effVal := getHeroBaseStatValue("eff", baseStats)
	if effVal != 18.0 {
		t.Errorf("Expected base eff value to be 18.0, got: %.2f", effVal)
	}

	resVal := getHeroBaseStatValue("res", baseStats)
	if resVal != 30.0 {
		t.Errorf("Expected base res value to be 30.0, got: %.2f", resVal)
	}
}

func TestEvaluateLayer6_LowercaseRaritySalvage(t *testing.T) {
	gearSalvage := Gear{
		Slot:    "Necklace",
		Level:   85,
		Rarity:  "epic", // Lowercase
		Enhance: 0,
		Set:     "Speed",
		Main:    MainStat{Type: "CritHitDamagePercent", Value: 13.0},
		Substats: []Substat{
			{Type: "AttackPercent", Value: 8.0, Rolls: 1},
			{Type: "CritHitChancePercent", Value: 5.0, Rolls: 1},
			{Type: "Speed", Value: 4.0, Rolls: 1},
			{Type: "EffectResistancePercent", Value: 5.0, Rolls: 1},
		},
	}

	profile := HeroProfile{
		HeroID:    "DPS_Build_1",
		HeroName:  "DPS",
		BuildRank: 1,
		Selected:  true,
		Sets:      [][]string{{"Speed"}},
		Priorities: map[string]int{
			"atk": 3, "cd": 3, "cc": 3, "spd": 3, "res": 0,
		},
		WeightMode: "weighted",
	}

	baseStats := HeroBaseStats{Atk: 1200, Def: 600, Hp: 6000, Spd: 110, Cc: 0.15, Cd: 1.5}
	l4Res := EvaluateLayer4(gearSalvage, profile, baseStats)
	l5Res := EvaluateLayer5(gearSalvage, profile, baseStats)

	l6 := EvaluateLayer6(
		gearSalvage,
		L3Trace{},
		[]L4HeroGateResult{l4Res},
		map[string]L5Trace{profile.HeroID: l5Res},
		[]HeroProfile{profile},
		map[string]HeroBaseStats{profile.HeroID: baseStats},
		"surplus",
		"none",
	)

	if l6.Verdict != "KEEP_ENHANCE" && l6.Verdict != "SALVAGE_MOD" {
		t.Errorf("Expected lowercase rarity gear to evaluate salvage properly, got verdict: %s", l6.Verdict)
	}
}

func TestEvaluateLayer3_SpeedVaultMinThreshold(t *testing.T) {
	gearLowSpeed := Gear{
		Slot:   "Ring",
		Level:  85,
		Rarity: "Epic",
		Substats: []Substat{
			{Type: "Speed", Value: 2.0, Rolls: 1},
		},
	}

	l3Default := EvaluateLayer3(gearLowSpeed, "ON")
	if !l3Default.Tagged {
		t.Errorf("Expected speed check to tag Speed 2 when threshold is 0")
	}

	l3WithMin := EvaluateLayer3(gearLowSpeed, "ON", 3.0)
	if l3WithMin.Tagged {
		t.Errorf("Expected speed check NOT to tag Speed 2 when threshold is 3.0")
	}
}

// ---------------------------------------------------------------------------
// Validation Suite V1 - V5 (Q-E7-Architecture §8)
// ---------------------------------------------------------------------------

func TestValidationCase_V1_TankHelmetMissingDef(t *testing.T) {
	// Heroic L85 Helmet (HP main), +0, subs Atk% 6, HP% 6, Eff% 5, Efr% 4
	gear := Gear{
		Slot:    "Helmet",
		Level:   85,
		Rarity:  "Heroic",
		Enhance: 0,
		Main:    MainStat{Type: "Health", Value: 540.0},
		Substats: []Substat{
			{Type: "AttackPercent", Value: 6.0, Rolls: 1},
			{Type: "HealthPercent", Value: 6.0, Rolls: 1},
			{Type: "EffectivenessPercent", Value: 5.0, Rolls: 1},
			{Type: "EffectResistancePercent", Value: 4.0, Rolls: 1},
		},
	}

	// Tank priorities: HP +5, Def +5, Eff +3, Spd +2, Efr +1, Atk -2, CC -1, CD -1
	profile := HeroProfile{
		HeroID:     "Tank_1",
		HeroName:   "Tank",
		BuildRank:  1,
		Selected:   true,
		RosterTier: "primary",
		Priorities: map[string]int{
			"hp": 5, "def": 5, "eff": 3, "spd": 2, "res": 1, "atk": -2, "cc": -1, "cd": -1,
		},
		WeightMode: "weighted",
	}

	baseStats := HeroBaseStats{Atk: 900, Def: 650, Hp: 6500, Spd: 100}

	l4Res := EvaluateLayer4(gear, profile, baseStats)
	// Def is essential (+5) and slot-possible on Helmet, so missing Def pays -18 penalty
	if l4Res.AbsencePenalty < 18.0 {
		t.Errorf("Expected absence penalty >= 18.0 for missing essential Def%% on Helmet, got: %.2f", l4Res.AbsencePenalty)
	}

	// fit% should land MARGINAL (< 45%)
	if l4Res.FitClass != "MARGINAL" && l4Res.FitClass != "REJECT" {
		t.Errorf("Expected fitClass to be MARGINAL or REJECT due to absence & dead Atk%%, got: %s (fitPct: %.2f)", l4Res.FitClass, l4Res.FitPct)
	}

	l5Res := EvaluateLayer5(gear, profile, baseStats)
	l6 := EvaluateLayer6(
		gear,
		L3Trace{},
		[]L4HeroGateResult{l4Res},
		map[string]L5Trace{profile.HeroID: l5Res},
		[]HeroProfile{profile},
		map[string]HeroBaseStats{profile.HeroID: baseStats},
		"none",
		"none",
	)

	// mod_budget = none -> SELL_EXTRACT
	if l6.Verdict != "SELL_EXTRACT" && l6.Verdict != "KEEP_MARGINAL" {
		t.Errorf("Expected V1 verdict to be SELL_EXTRACT or KEEP_MARGINAL under mod_budget=none, got: %s", l6.Verdict)
	}
}

func TestValidationCase_V4_ModScarcityAsIsFloor(t *testing.T) {
	// Base fit% is sub-USABLE (as-is projected fit% ~38%). Even with mod_budget = surplus, SALVAGE requires as-is >= USABLE.
	gear := Gear{
		Slot:    "Helmet",
		Level:   85,
		Rarity:  "Epic",
		Enhance: 6,
		Main:    MainStat{Type: "Health", Value: 1200.0},
		Substats: []Substat{
			{Type: "Speed", Value: 4.0, Rolls: 1},
			{Type: "AttackPercent", Value: 8.0, Rolls: 1},
			{Type: "CritHitChancePercent", Value: 5.0, Rolls: 1},
			{Type: "Defense", Value: 110.0, Rolls: 3}, // 2 dead rolls into flatDef
		},
	}

	profile := HeroProfile{
		HeroID:     "DPS_V4",
		HeroName:   "DPS",
		BuildRank:  1,
		Selected:   true,
		RosterTier: "primary",
		Priorities: map[string]int{
			"spd": 5, "cc": 5, "cd": 5, "atk": 2, "def": -3,
		},
		WeightMode: "weighted",
	}

	baseStats := HeroBaseStats{Atk: 1200, Def: 600, Hp: 6000, Spd: 110}

	l4Res := EvaluateLayer4(gear, profile, baseStats)
	l5Res := EvaluateLayer5(gear, profile, baseStats)

	// Under surplus mod budget
	l6Surplus := EvaluateLayer6(
		gear,
		L3Trace{},
		[]L4HeroGateResult{l4Res},
		map[string]L5Trace{profile.HeroID: l5Res},
		[]HeroProfile{profile},
		map[string]HeroBaseStats{profile.HeroID: baseStats},
		"surplus",
		"none",
	)

	// Since as-is fit is sub-USABLE (due to missing essential CD penalty), it must NOT be SALVAGE_MOD
	if l4Res.FitPct < 45.0 && l6Surplus.Verdict == "SALVAGE_MOD" {
		t.Errorf("Axiom 5 violation: sub-USABLE base (fitPct %.2f < 45) was wrongly SALVAGE_MOD!", l4Res.FitPct)
	}
}

func TestValidationCase_V2_CatalogTierDilutionGuard(t *testing.T) {
	gear := Gear{
		Slot:    "Necklace",
		Level:   88,
		Rarity:  "Epic",
		Enhance: 0,
		Main:    MainStat{Type: "CritHitDamagePercent", Value: 14.0},
		Substats: []Substat{
			{Type: "CritHitChancePercent", Value: 5.0, Rolls: 1},
			{Type: "AttackPercent", Value: 8.0, Rolls: 1},
			{Type: "DefensePercent", Value: 6.0, Rolls: 1},
			{Type: "HealthPercent", Value: 6.0, Rolls: 1},
		},
	}

	// Tank primary hero
	tankProfile := HeroProfile{
		HeroID:     "Tank_V2",
		HeroName:   "Tank",
		BuildRank:  1,
		RosterTier: "primary",
		Priorities: map[string]int{
			"hp": 5, "def": 5, "eff": 3, "atk": -2, "cc": -2, "cd": -2,
		},
	}
	// Catalog DPS hero
	catalogDpsProfile := HeroProfile{
		HeroID:     "CatalogDPS_V2",
		HeroName:   "CatalogDPS",
		BuildRank:  1,
		RosterTier: "catalog",
		Priorities: map[string]int{
			"atk": 5, "cd": 5, "cc": 4, "spd": 3,
		},
	}

	baseTank := HeroBaseStats{Atk: 900, Def: 650, Hp: 6500, Spd: 100}
	baseDps := HeroBaseStats{Atk: 1200, Def: 600, Hp: 6000, Spd: 110}

	l4Tank := EvaluateLayer4(gear, tankProfile, baseTank)
	l4Dps := EvaluateLayer4(gear, catalogDpsProfile, baseDps)

	// Tank should reject piece
	if l4Tank.Pass {
		t.Errorf("Expected Tank hero to reject DPS necklace with negative priorities")
	}

	// Catalog hero requires CORE and P90 fitPct (80%)
	if l4Dps.RosterTier == "catalog" && l4Dps.FitPct < 80.0 && l4Dps.Pass {
		t.Errorf("Catalog tier dilution guard failed: catalog match below P90 bar wrongly passed Gate 4!")
	}
}

func TestValidationCase_V3_Plus6StopDecision(t *testing.T) {
	// Boots enhanced to +6 with 2 rolls going into flatDef (dead stat)
	gearBoots := Gear{
		Slot:    "Boots",
		Level:   85,
		Rarity:  "Epic",
		Enhance: 6,
		Main:    MainStat{Type: "AttackPercent", Value: 20.0},
		Substats: []Substat{
			{Type: "Speed", Value: 3.0, Rolls: 1},
			{Type: "CritHitDamagePercent", Value: 5.0, Rolls: 1},
			{Type: "CritHitChancePercent", Value: 4.0, Rolls: 1},
			{Type: "Defense", Value: 90.0, Rolls: 3}, // 2 dead rolls into flatDef
		},
	}

	profile := HeroProfile{
		HeroID:        "DPS_V3",
		HeroName:      "DPS",
		BuildRank:     1,
		RiskTolerance: 0.5,
		Priorities: map[string]int{
			"spd": 5, "cd": 5, "cc": 4, "atk": 3, "def": -3,
		},
	}

	baseStats := HeroBaseStats{Atk: 1200, Def: 600, Hp: 6000, Spd: 110}
	l5Res := EvaluateLayer5(gearBoots, profile, baseStats)

	if l5Res.StopCard == nil {
		t.Fatalf("Expected StopCard to be emitted by Enhancement Controller at +6")
	}

	if l5Res.StopCard.Recommended != "STOP" {
		t.Errorf("Expected Enhancement Controller to recommend STOP at +6 with two dead rolls, got: %s", l5Res.StopCard.Recommended)
	}
}

func TestValidationCase_V5_ReforgeScarcity(t *testing.T) {
	gear85 := Gear{
		Slot:    "Helmet",
		Level:   85,
		Rarity:  "Heroic",
		Enhance: 15,
		Main:    MainStat{Type: "Health", Value: 2835.0},
		Substats: []Substat{
			{Type: "Speed", Value: 9.0, Rolls: 3},
			{Type: "HealthPercent", Value: 12.0, Rolls: 2},
			{Type: "DefensePercent", Value: 10.0, Rolls: 2},
			{Type: "EffectivenessPercent", Value: 6.0, Rolls: 1},
		},
	}

	profile := HeroProfile{
		HeroID:     "Bruiser_V5",
		HeroName:   "Bruiser",
		BuildRank:  1,
		RosterTier: "primary",
		Priorities: map[string]int{
			"hp": 4, "def": 4, "spd": 3, "eff": 1,
		},
	}

	baseStats := HeroBaseStats{Atk: 950, Def: 650, Hp: 6500, Spd: 105}
	l4Res := EvaluateLayer4(gear85, profile, baseStats)
	l5Res := EvaluateLayer5(gear85, profile, baseStats)

	// reforge_budget = none -> verdict KEEP_ENHANCE or KEEP_MARGINAL
	l6None := EvaluateLayer6(
		gear85,
		L3Trace{},
		[]L4HeroGateResult{l4Res},
		map[string]L5Trace{profile.HeroID: l5Res},
		[]HeroProfile{profile},
		map[string]HeroBaseStats{profile.HeroID: baseStats},
		"none",
		"none",
	)

	// reforge_budget = surplus -> REFORGE_TAG if reforge elevates class
	l6Surplus := EvaluateLayer6(
		gear85,
		L3Trace{},
		[]L4HeroGateResult{l4Res},
		map[string]L5Trace{profile.HeroID: l5Res},
		[]HeroProfile{profile},
		map[string]HeroBaseStats{profile.HeroID: baseStats},
		"none",
		"surplus",
	)

	if l6None.Verdict == "REFORGE_TAG" {
		t.Errorf("Expected reforge_budget=none NOT to emit REFORGE_TAG")
	}
	if l6Surplus.Verdict != "REFORGE_TAG" && l6Surplus.Verdict != "KEEP_ENHANCE" {
		t.Errorf("Expected reforge_budget=surplus to evaluate REFORGE_TAG or KEEP_ENHANCE, got: %s", l6Surplus.Verdict)
	}
}

