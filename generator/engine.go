package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
)

type Substat struct {
	Type  string  `json:"type"`
	Value float64 `json:"value"`
	Rolls int     `json:"rolls"`
}

type MainStat struct {
	Type  string  `json:"type"`
	Value float64 `json:"value"`
}

type HistoryStep struct {
	Step      int     `json:"step"`
	Type      string  `json:"type"` // "unlock" or "upgrade"
	Stat      string  `json:"stat"`
	Value     float64 `json:"value"`
	PrevValue float64 `json:"prevValue"`
	NewValue  float64 `json:"newValue"`
}

type GearScore struct {
	GS float64 `json:"gs"`
	ES int     `json:"es"`
}

type Gear struct {
	ID       string        `json:"id"`
	Set      string        `json:"set"`
	Slot     string        `json:"slot"`
	Rarity   string        `json:"rarity"`
	Level    int           `json:"level"`
	Enhance  int           `json:"enhance"`
	Main     MainStat      `json:"main"`
	Substats []Substat     `json:"substats"`
	History  []HistoryStep `json:"history"`
	Score    GearScore     `json:"score"`
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func rollSubstatValue(statType string, level int, rarity string) float64 {
	cleanRarity := rarity
	if cleanRarity != "Epic" && cleanRarity != "Heroic" {
		cleanRarity = "Epic"
	}
	cleanLevel := level
	if cleanLevel != 88 {
		cleanLevel = 85
	}

	if statType == "Speed" {
		list := SPEED_PROBABILITIES[cleanLevel][cleanRarity]
		if len(list) > 0 {
			r := rand.Float64()
			var cumulative float64 = 0
			for _, item := range list {
				cumulative += item.Prob
				if r <= cumulative {
					return item.Val
				}
			}
			return list[len(list)-1].Val
		}
	}

	ranges := ROLL_RANGES_85
	if cleanLevel == 88 {
		ranges = ROLL_RANGES_88
	}

	statRange, exists := ranges[cleanRarity][statType]
	if !exists {
		return 1.0
	}
	return float64(randomInt(statRange.Min, statRange.Max))
}

func mainStatValue(statType string, enhance int, level int) float64 {
	cleanLevel := level
	if cleanLevel != 88 {
		cleanLevel = 85
	}
	stats, exists := MAIN_STAT_PROGRESSION[cleanLevel][statType]
	if !exists {
		return 0
	}

	e := enhance
	if e < 0 {
		e = 0
	}
	if e > 15 {
		e = 15
	}

	if val, found := stats[e]; found {
		return val
	}

	keys := []int{0, 3, 6, 9, 12, 15}
	lowerKey := 0
	upperKey := 15
	for i := 0; i < len(keys); i++ {
		if keys[i] <= e {
			lowerKey = keys[i]
		}
		if keys[i] >= e {
			upperKey = keys[i]
			break
		}
	}
	if lowerKey == upperKey {
		return stats[lowerKey]
	}

	fraction := float64(e-lowerKey) / float64(upperKey-lowerKey)
	val := stats[lowerKey] + fraction*(stats[upperKey]-stats[lowerKey])
	return math.Round(val)
}

func generateGear(options GenerateOptions) Gear {
	rarity := "Epic"
	if len(options.Rarities) > 0 {
		rarity = randomChoiceString(options.Rarities)
	}

	level := 85
	if len(options.Levels) > 0 {
		level = randomChoiceInt(options.Levels)
	}

	if rarity == "Heroic" {
		level = 85
	}

	slot := randomChoiceString(SLOTS)
	if len(options.Slots) > 0 {
		slot = randomChoiceString(options.Slots)
	}

	set := randomChoiceString(SETS)
	if len(options.Sets) > 0 {
		set = randomChoiceString(options.Sets)
	}

	var mainType string
	fixed := FIXED_MAIN_BY_SLOT[slot]
	if fixed != "" {
		mainType = fixed
	} else {
		allowedMains := FLEX_MAIN_BY_SLOT[slot]
		var selectedAllowed []string
		for _, t := range options.MainTypes {
			if containsString(allowedMains, t) {
				selectedAllowed = append(selectedAllowed, t)
			}
		}

		if len(selectedAllowed) > 0 {
			mainType = randomChoiceString(selectedAllowed)
		} else {
			mainType = randomChoiceString(allowedMains)
		}
	}

	mainVal := mainStatValue(mainType, 0, level)

	startCount := 4
	if rarity == "Heroic" {
		startCount = 3
	}

	usedTypes := map[string]bool{mainType: true}
	var allowedPool []string
	for _, t := range STAT_TYPES {
		if usedTypes[t] {
			continue
		}
		if slot == "Weapon" && (t == "Defense" || t == "DefensePercent") {
			continue
		}
		if slot == "Armor" && (t == "Attack" || t == "AttackPercent") {
			continue
		}
		allowedPool = append(allowedPool, t)
	}

	shuffled := make([]string, len(allowedPool))
	copy(shuffled, allowedPool)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	pickCount := startCount
	if len(shuffled) < pickCount {
		pickCount = len(shuffled)
	}

	var substats []Substat
	for i := 0; i < pickCount; i++ {
		t := shuffled[i]
		val := rollSubstatValue(t, level, rarity)
		substats = append(substats, Substat{
			Type:  t,
			Value: val,
			Rolls: 1,
		})
		usedTypes[t] = true
	}

	id := fmt.Sprintf("gear-%d-%d", time.Now().UnixNano()/1000, rand.Intn(100000))
	gear := Gear{
		ID:       id,
		Set:      set,
		Slot:     slot,
		Rarity:   rarity,
		Level:    level,
		Enhance:  0,
		Main:     MainStat{Type: mainType, Value: mainVal},
		Substats: substats,
		History:  []HistoryStep{},
	}
	gear.Score = calculateEquipmentScore(gear)
	return gear
}

func enhanceStep(gear Gear) Gear {
	if gear.Enhance >= 15 {
		return gear
	}

	nextEnhance := gear.Enhance + 3
	if nextEnhance > 15 {
		nextEnhance = 15
	}

	mainVal := mainStatValue(gear.Main.Type, nextEnhance, gear.Level)

	substats := make([]Substat, len(gear.Substats))
	copy(substats, gear.Substats)

	history := make([]HistoryStep, len(gear.History))
	copy(history, gear.History)

	shouldAddNewSub := gear.Rarity == "Heroic" && len(substats) < 4 && gear.Enhance == 9

	if shouldAddNewSub {
		usedTypes := map[string]bool{gear.Main.Type: true}
		for _, s := range substats {
			usedTypes[s.Type] = true
		}

		var pool []string
		for _, t := range STAT_TYPES {
			if usedTypes[t] {
				continue
			}
			if gear.Slot == "Weapon" && (t == "Defense" || t == "DefensePercent") {
				continue
			}
			if gear.Slot == "Armor" && (t == "Attack" || t == "AttackPercent") {
				continue
			}
			pool = append(pool, t)
		}

		newType := "Speed"
		if len(pool) > 0 {
			newType = randomChoiceString(pool)
		}

		rolledVal := rollSubstatValue(newType, gear.Level, gear.Rarity)

		substats = append(substats, Substat{
			Type:  newType,
			Value: rolledVal,
			Rolls: 1,
		})

		history = append(history, HistoryStep{
			Step:      nextEnhance,
			Type:      "unlock",
			Stat:      newType,
			Value:     rolledVal,
			PrevValue: 0,
			NewValue:  rolledVal,
		})
	} else {
		if len(substats) > 0 {
			idx := rand.Intn(len(substats))
			target := &substats[idx]
			prevVal := target.Value
			rolledVal := rollSubstatValue(target.Type, gear.Level, gear.Rarity)

			target.Value += rolledVal
			target.Rolls += 1

			history = append(history, HistoryStep{
				Step:      nextEnhance,
				Type:      "upgrade",
				Stat:      target.Type,
				Value:     rolledVal,
				PrevValue: prevVal,
				NewValue:  target.Value,
			})
		}
	}

	gear.Enhance = nextEnhance
	gear.Main.Value = mainVal
	gear.Substats = substats
	gear.History = history
	gear.Score = calculateEquipmentScore(gear)

	return gear
}

func enhanceToLevel(gear Gear, targetLevel int) Gear {
	current := gear
	target := targetLevel
	if target > 15 {
		target = 15
	}
	if target < 0 {
		target = 0
	}

	for current.Enhance < target {
		current = enhanceStep(current)
	}

	return current
}

func calculateEquipmentScore(gear Gear) GearScore {
	gsWeights := map[string]float64{
		"AttackPercent":           1.0,
		"DefensePercent":          1.0,
		"HealthPercent":           1.0,
		"EffectivenessPercent":    1.0,
		"EffectResistancePercent": 1.0,
		"Speed":                   2.0,
		"CritHitChancePercent":    1.6,
		"CritHitDamagePercent":    8.0 / 7.0,
		"Attack":                  0.088718,
		"Defense":                 0.160968,
		"Health":                  0.017759,
	}

	esWeights := map[string]float64{
		"AttackPercent":           1.0,
		"DefensePercent":          1.0,
		"HealthPercent":           1.0,
		"EffectivenessPercent":    1.0,
		"EffectResistancePercent": 1.0,
		"Speed":                   2.0,
		"CritHitChancePercent":    1.6,
		"CritHitDamagePercent":    8.0 / 7.0,
		"Attack":                  3.7373 / 39.4717,
		"Defense":                 4.9853 / 31.0051,
		"Health":                  3.1505 / 179.4031,
	}

	var gsSubstats float64 = 0
	var esSubstats float64 = 0

	for _, sub := range gear.Substats {
		gsSubstats += sub.Value * gsWeights[sub.Type]
		esSubstats += sub.Value * esWeights[sub.Type]
	}

	var esMain float64 = 0
	if gear.Main.Type != "" {
		mainType := gear.Main.Type
		mainVal := gear.Main.Value

		isPercentMain := strings.HasSuffix(mainType, "Percent") ||
			mainType == "CritHitChancePercent" ||
			mainType == "CritHitDamagePercent" ||
			mainType == "EffectivenessPercent" ||
			mainType == "EffectResistancePercent"

		if isPercentMain {
			esMain = mainVal / 2.5
		} else if mainType == "Speed" {
			esMain = mainVal * 2.0
		} else {
			var maxVal85 float64
			switch mainType {
			case "Attack":
				maxVal85 = 525.0
			case "Health":
				maxVal85 = 2835.0
			case "Defense":
				maxVal85 = 310.0
			default:
				maxVal85 = 1.0
			}
			esMain = (mainVal / maxVal85) * 24.2
		}
	}

	finalGS := math.Round(gsSubstats*10.0) / 10.0
	finalES := int(math.Floor(esMain + esSubstats))

	return GearScore{
		GS: finalGS,
		ES: finalES,
	}
}

func randomChoiceString(arr []string) string {
	if len(arr) == 0 {
		return ""
	}
	return arr[rand.Intn(len(arr))]
}

func randomChoiceInt(arr []int) int {
	if len(arr) == 0 {
		return 0
	}
	return arr[rand.Intn(len(arr))]
}

func randomInt(min, max int) int {
	if min >= max {
		return min
	}
	return rand.Intn(max-min+1) + min
}

func containsString(arr []string, val string) bool {
	for _, x := range arr {
		if x == val {
			return true
		}
	}
	return false
}
