package main

type RollRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type ProbItem struct {
	Val  float64 `json:"val"`
	Prob float64 `json:"prob"`
}

var SETS = []string{
	"HealthSet",
	"DefenseSet",
	"AttackSet",
	"SpeedSet",
	"CriticalSet",
	"HitSet",
	"DestructionSet",
	"LifestealSet",
	"CounterSet",
	"ResistSet",
	"UnitySet",
	"RageSet",
	"ImmunitySet",
	"PenetrationSet",
	"RevengeSet",
	"InjurySet",
	"ProtectionSet",
	"TorrentSet",
	"ReversalSet",
	"RiposteSet",
	"WarfareSet",
	"PursuitSet",
	"WeakeningSet",
	"FervorSet",
}

var SET_LABELS = map[string]string{
	"HealthSet":               "Health",
	"DefenseSet":              "Defense",
	"AttackSet":               "Attack",
	"SpeedSet":                "Speed",
	"CriticalSet":             "Critical",
	"HitSet":                  "Hit (Effect)",
	"DestructionSet":          "Destruction",
	"LifestealSet":            "Lifesteal",
	"CounterSet":              "Counter",
	"ResistSet":               "Resist",
	"UnitySet":                "Unity",
	"RageSet":                 "Rage",
	"ImmunitySet":             "Immunity",
	"PenetrationSet":          "Penetration",
	"RevengeSet":              "Revenge",
	"InjurySet":               "Injury",
	"ProtectionSet":           "Protection",
	"TorrentSet":              "Torrent",
	"ReversalSet":             "Reversal",
	"RiposteSet":              "Riposte",
	"WarfareSet":              "Warfare",
	"PursuitSet":              "Pursuit",
	"WeakeningSet":            "Weakening",
	"FervorSet":               "Fervor",
}

var SLOTS = []string{"Weapon", "Helmet", "Armor", "Necklace", "Ring", "Boots"}

var STAT_TYPES = []string{
	"Attack",
	"AttackPercent",
	"Health",
	"HealthPercent",
	"Defense",
	"DefensePercent",
	"Speed",
	"CritHitChancePercent",
	"CritHitDamagePercent",
	"EffectivenessPercent",
	"EffectResistancePercent",
}

var STAT_LABELS = map[string]string{
	"Attack":                  "Attack",
	"AttackPercent":           "Attack %",
	"Health":                  "Health",
	"HealthPercent":           "Health %",
	"Defense":                 "Defense",
	"DefensePercent":          "Defense %",
	"Speed":                   "Speed",
	"CritHitChancePercent":    "Crit Chance %",
	"CritHitDamagePercent":    "Crit Damage %",
	"EffectivenessPercent":    "Effectiveness %",
	"EffectResistancePercent": "Effect Resist %",
}

var FIXED_MAIN_BY_SLOT = map[string]string{
	"Weapon": "Attack",
	"Helmet": "Health",
	"Armor":  "Defense",
}

var FLEX_MAIN_BY_SLOT = map[string][]string{
	"Necklace": {"AttackPercent", "HealthPercent", "DefensePercent", "CritHitChancePercent", "CritHitDamagePercent", "Health", "Defense", "Attack"},
	"Ring":     {"AttackPercent", "HealthPercent", "DefensePercent", "EffectivenessPercent", "EffectResistancePercent", "Health", "Defense", "Attack"},
	"Boots":    {"AttackPercent", "HealthPercent", "DefensePercent", "Speed", "Health", "Defense", "Attack"},
}

var ROLL_RANGES_85 = map[string]map[string]RollRange{
	"Epic": {
		"AttackPercent":           {Min: 4, Max: 8},
		"HealthPercent":           {Min: 4, Max: 8},
		"DefensePercent":          {Min: 4, Max: 8},
		"EffectivenessPercent":    {Min: 4, Max: 8},
		"EffectResistancePercent": {Min: 4, Max: 8},
		"CritHitDamagePercent":    {Min: 4, Max: 7},
		"CritHitChancePercent":    {Min: 3, Max: 5},
		"Speed":                   {Min: 2, Max: 5},
		"Attack":                  {Min: 33, Max: 46},
		"Health":                  {Min: 157, Max: 202},
		"Defense":                 {Min: 28, Max: 35},
	},
	"Heroic": {
		"AttackPercent":           {Min: 4, Max: 8},
		"HealthPercent":           {Min: 4, Max: 8},
		"DefensePercent":          {Min: 4, Max: 8},
		"EffectivenessPercent":    {Min: 4, Max: 8},
		"EffectResistancePercent": {Min: 4, Max: 8},
		"CritHitDamagePercent":    {Min: 4, Max: 7},
		"CritHitChancePercent":    {Min: 3, Max: 5},
		"Speed":                   {Min: 1, Max: 4},
		"Attack":                  {Min: 31, Max: 44},
		"Health":                  {Min: 149, Max: 192},
		"Defense":                 {Min: 26, Max: 33},
	},
}

var ROLL_RANGES_88 = map[string]map[string]RollRange{
	"Epic": {
		"AttackPercent":           {Min: 5, Max: 9},
		"HealthPercent":           {Min: 5, Max: 9},
		"DefensePercent":          {Min: 5, Max: 9},
		"EffectivenessPercent":    {Min: 5, Max: 9},
		"EffectResistancePercent": {Min: 5, Max: 9},
		"CritHitDamagePercent":    {Min: 4, Max: 8},
		"CritHitChancePercent":    {Min: 3, Max: 6},
		"Speed":                   {Min: 3, Max: 5},
		"Attack":                  {Min: 37, Max: 53},
		"Health":                  {Min: 178, Max: 229},
		"Defense":                 {Min: 32, Max: 40},
	},
	"Heroic": {
		"AttackPercent":           {Min: 5, Max: 9},
		"HealthPercent":           {Min: 5, Max: 9},
		"DefensePercent":          {Min: 5, Max: 9},
		"EffectivenessPercent":    {Min: 5, Max: 9},
		"EffectResistancePercent": {Min: 5, Max: 9},
		"CritHitDamagePercent":    {Min: 4, Max: 8},
		"CritHitChancePercent":    {Min: 3, Max: 6},
		"Speed":                   {Min: 2, Max: 4},
		"Attack":                  {Min: 36, Max: 50},
		"Health":                  {Min: 169, Max: 218},
		"Defense":                 {Min: 30, Max: 38},
	},
}

var SPEED_PROBABILITIES = map[int]map[string][]ProbItem{
	85: {
		"Epic": {
			{Val: 2, Prob: 0.33223},
			{Val: 3, Prob: 0.33223},
			{Val: 4, Prob: 0.33223},
			{Val: 5, Prob: 0.00331},
		},
		"Heroic": {
			{Val: 1, Prob: 0.03833},
			{Val: 2, Prob: 0.34843},
			{Val: 3, Prob: 0.34843},
			{Val: 4, Prob: 0.26481},
		},
	},
	88: {
		"Epic": {
			{Val: 3, Prob: 0.49751},
			{Val: 4, Prob: 0.49751},
			{Val: 5, Prob: 0.00498},
		},
		"Heroic": {
			{Val: 2, Prob: 0.08333},
			{Val: 3, Prob: 0.52083},
			{Val: 4, Prob: 0.39583},
		},
	},
}

var MAIN_STAT_PROGRESSION = map[int]map[string]map[int]float64{
	85: {
		"Attack":                  {0: 100, 3: 150, 6: 200, 9: 250, 12: 325, 15: 525},
		"Health":                  {0: 540, 3: 810, 6: 1080, 9: 1350, 12: 1755, 15: 2835},
		"Defense":                 {0: 60, 3: 90, 6: 120, 9: 150, 12: 195, 15: 310},
		"AttackPercent":           {0: 12, 3: 18, 6: 24, 9: 30, 12: 39, 15: 65},
		"HealthPercent":           {0: 12, 3: 18, 6: 24, 9: 30, 12: 39, 15: 65},
		"DefensePercent":          {0: 12, 3: 18, 6: 24, 9: 30, 12: 39, 15: 65},
		"CritHitChancePercent":    {0: 11, 3: 15, 6: 20, 9: 25, 12: 33, 15: 60},
		"CritHitDamagePercent":    {0: 13, 3: 19, 6: 25, 9: 31, 12: 40, 15: 70},
		"EffectivenessPercent":    {0: 12, 3: 18, 6: 24, 9: 30, 12: 39, 15: 65},
		"EffectResistancePercent": {0: 12, 3: 18, 6: 24, 9: 30, 12: 39, 15: 65},
		"Speed":                   {0: 8, 3: 12, 6: 16, 9: 20, 12: 26, 15: 45},
	},
	88: {
		"Attack":                  {0: 103, 3: 154, 6: 206, 9: 258, 12: 336, 15: 550},
		"Health":                  {0: 553, 3: 829, 6: 1105, 9: 1381, 12: 1795, 15: 2970},
		"Defense":                 {0: 62, 3: 93, 6: 124, 9: 155, 12: 202, 15: 320},
		"AttackPercent":           {0: 13, 3: 19, 6: 25, 9: 31, 12: 40, 15: 70},
		"HealthPercent":           {0: 13, 3: 19, 6: 25, 9: 31, 12: 40, 15: 70},
		"DefensePercent":          {0: 13, 3: 19, 6: 25, 9: 31, 12: 40, 15: 70},
		"CritHitChancePercent":    {0: 12, 3: 18, 6: 24, 9: 30, 12: 36, 15: 65},
		"CritHitDamagePercent":    {0: 14, 3: 21, 6: 28, 9: 35, 12: 45, 15: 75},
		"EffectivenessPercent":    {0: 13, 3: 19, 6: 25, 9: 31, 12: 40, 15: 70},
		"EffectResistancePercent": {0: 13, 3: 19, 6: 25, 9: 31, 12: 40, 15: 70},
		"Speed":                   {0: 9, 3: 13, 6: 17, 9: 21, 12: 28, 15: 50},
	},
}

var FLEXIBLE_MAIN_STATS = []string{
	"AttackPercent", "HealthPercent", "DefensePercent", "Speed",
	"CritHitChancePercent", "CritHitDamagePercent", "EffectivenessPercent", "EffectResistancePercent",
	"Attack", "Health", "Defense",
}
