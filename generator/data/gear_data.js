// Epic Seven Gear Data Constants
// Extracted from original Next.js application & Fribbels database.

export const SETS = [
  'HealthSet',
  'DefenseSet',
  'AttackSet',
  'SpeedSet',
  'CriticalSet',
  'HitSet',
  'DestructionSet',
  'LifestealSet',
  'CounterSet',
  'ResistSet',
  'UnitySet',
  'RageSet',
  'ImmunitySet',
  'PenetrationSet',
  'RevengeSet',
  'InjurySet',
  'ProtectionSet',
  'TorrentSet',
  'ReversalSet',
  'RiposteSet',
  'WarfareSet',
  'PursuitSet',
  'WeakeningSet',
  'FervorSet',
];

export const SET_LABELS = {
  HealthSet: 'Health',
  DefenseSet: 'Defense',
  AttackSet: 'Attack',
  SpeedSet: 'Speed',
  CriticalSet: 'Critical',
  HitSet: 'Hit (Effect)',
  DestructionSet: 'Destruction',
  LifestealSet: 'Lifesteal',
  CounterSet: 'Counter',
  ResistSet: 'Resist',
  UnitySet: 'Unity',
  RageSet: 'Rage',
  ImmunitySet: 'Immunity',
  PenetrationSet: 'Penetration',
  RevengeSet: 'Revenge',
  InjurySet: 'Injury',
  ProtectionSet: 'Protection',
  TorrentSet: 'Torrent',
  ReversalSet: 'Reversal',
  RiposteSet: 'Riposte',
  WarfareSet: 'Warfare',
  PursuitSet: 'Pursuit',
  WeakeningSet: 'Weakening',
  FervorSet: 'Fervor',
};

export const SLOTS = ['Weapon', 'Helmet', 'Armor', 'Necklace', 'Ring', 'Boots'];

export const STAT_TYPES = [
  'Attack',
  'AttackPercent',
  'Health',
  'HealthPercent',
  'Defense',
  'DefensePercent',
  'Speed',
  'CritHitChancePercent',
  'CritHitDamagePercent',
  'EffectivenessPercent',
  'EffectResistancePercent',
];

export const STAT_LABELS = {
  Attack: 'Attack',
  AttackPercent: 'Attack %',
  Health: 'Health',
  HealthPercent: 'Health %',
  Defense: 'Defense',
  DefensePercent: 'Defense %',
  Speed: 'Speed',
  CritHitChancePercent: 'Crit Chance %',
  CritHitDamagePercent: 'Crit Damage %',
  EffectivenessPercent: 'Effectiveness %',
  EffectResistancePercent: 'Effect Resist %',
};

export const STAT_SHORT = {
  Attack: 'ATK',
  AttackPercent: 'ATK%',
  Health: 'HP',
  HealthPercent: 'HP%',
  Defense: 'DEF',
  DefensePercent: 'DEF%',
  Speed: 'SPD',
  CritHitChancePercent: 'C.Chance',
  CritHitDamagePercent: 'C.Dmg',
  EffectivenessPercent: 'EFF',
  EffectResistancePercent: 'EFF.RES',
};

// Main stat rules per slot
export const FIXED_MAIN_BY_SLOT = {
  Weapon: 'Attack',
  Helmet: 'Health',
  Armor: 'Defense',
};

export const FLEX_MAIN_BY_SLOT = {
  Necklace: ['AttackPercent', 'HealthPercent', 'DefensePercent', 'CritHitChancePercent', 'CritHitDamagePercent', 'Health', 'Defense', 'Attack'],
  Ring: ['AttackPercent', 'HealthPercent', 'DefensePercent', 'EffectivenessPercent', 'EffectResistancePercent', 'Health', 'Defense', 'Attack'],
  Boots: ['AttackPercent', 'HealthPercent', 'DefensePercent', 'Speed', 'Health', 'Defense', 'Attack'],
};

// Level 85 roll ranges by rarity
export const ROLL_RANGES_85 = {
  Epic: {
    AttackPercent: { min: 4, max: 8 },
    HealthPercent: { min: 4, max: 8 },
    DefensePercent: { min: 4, max: 8 },
    EffectivenessPercent: { min: 4, max: 8 },
    EffectResistancePercent: { min: 4, max: 8 },
    CritHitDamagePercent: { min: 4, max: 7 },
    CritHitChancePercent: { min: 3, max: 5 },
    Speed: { min: 2, max: 5 },
    Attack: { min: 33, max: 46 },
    Health: { min: 157, max: 202 },
    Defense: { min: 28, max: 35 },
  },
  Heroic: {
    AttackPercent: { min: 4, max: 8 },
    HealthPercent: { min: 4, max: 8 },
    DefensePercent: { min: 4, max: 8 },
    EffectivenessPercent: { min: 4, max: 8 },
    EffectResistancePercent: { min: 4, max: 8 },
    CritHitDamagePercent: { min: 4, max: 7 },
    CritHitChancePercent: { min: 3, max: 5 },
    Speed: { min: 1, max: 4 },
    Attack: { min: 31, max: 44 },
    Health: { min: 149, max: 192 },
    Defense: { min: 26, max: 33 },
  },
};

// Level 88 roll ranges by rarity
export const ROLL_RANGES_88 = {
  Epic: {
    AttackPercent: { min: 5, max: 9 },
    HealthPercent: { min: 5, max: 9 },
    DefensePercent: { min: 5, max: 9 },
    EffectivenessPercent: { min: 5, max: 9 },
    EffectResistancePercent: { min: 5, max: 9 },
    CritHitDamagePercent: { min: 4, max: 8 },
    CritHitChancePercent: { min: 3, max: 6 },
    Speed: { min: 3, max: 5 },
    Attack: { min: 37, max: 53 },
    Health: { min: 178, max: 229 },
    Defense: { min: 32, max: 40 },
  },
  Heroic: {
    // Note: Technically Heroic Lvl 88 is not obtainable, but we include it for completion.
    AttackPercent: { min: 5, max: 9 },
    HealthPercent: { min: 5, max: 9 },
    DefensePercent: { min: 5, max: 9 },
    EffectivenessPercent: { min: 5, max: 9 },
    EffectResistancePercent: { min: 5, max: 9 },
    CritHitDamagePercent: { min: 4, max: 8 },
    CritHitChancePercent: { min: 3, max: 6 },
    Speed: { min: 2, max: 4 },
    Attack: { min: 36, max: 50 },
    Health: { min: 169, max: 218 },
    Defense: { min: 30, max: 38 },
  },
};

// Exact Speed Roll Probability Weights from Fribbels
export const SPEED_PROBABILITIES = {
  85: {
    Epic: [
      { val: 2, prob: 0.33223 },
      { val: 3, prob: 0.33223 },
      { val: 4, prob: 0.33223 },
      { val: 5, prob: 0.00331 },
    ],
    Heroic: [
      { val: 1, prob: 0.03833 },
      { val: 2, prob: 0.34843 },
      { val: 3, prob: 0.34843 },
      { val: 4, prob: 0.26481 },
    ],
  },
  88: {
    Epic: [
      { val: 3, prob: 0.49751 },
      { val: 4, prob: 0.49751 },
      { val: 5, prob: 0.00498 },
    ],
    Heroic: [
      { val: 2, prob: 0.08333 },
      { val: 3, prob: 0.52083 },
      { val: 4, prob: 0.39583 },
    ],
  },
};

// Standard average roll ranges for ES calculation
export const SUBSTAT_ROLL_RANGES = {
  AttackPercent: { min: 4, max: 8 },
  HealthPercent: { min: 4, max: 8 },
  DefensePercent: { min: 4, max: 8 },
  Attack: { min: 33, max: 46 },
  Health: { min: 157, max: 202 },
  Defense: { min: 28, max: 35 },
  Speed: { min: 2, max: 5 },
  CritHitChancePercent: { min: 3, max: 5 },
  CritHitDamagePercent: { min: 4, max: 7 },
  EffectivenessPercent: { min: 4, max: 8 },
  EffectResistancePercent: { min: 4, max: 8 },
};

export const MAIN_STAT_PROGRESSION = {
  85: {
    Attack: { 0: 100, 3: 150, 6: 200, 9: 250, 12: 325, 15: 525 },
    Health: { 0: 540, 3: 810, 6: 1080, 9: 1350, 12: 1755, 15: 2835 },
    Defense: { 0: 60, 3: 90, 6: 120, 9: 150, 12: 195, 15: 310 },
    AttackPercent: { 0: 12, 3: 18, 6: 24, 9: 30, 12: 39, 15: 65 },
    HealthPercent: { 0: 12, 3: 18, 6: 24, 9: 30, 12: 39, 15: 65 },
    DefensePercent: { 0: 12, 3: 18, 6: 24, 9: 30, 12: 39, 15: 65 },
    CritHitChancePercent: { 0: 11, 3: 15, 6: 20, 9: 25, 12: 33, 15: 60 },
    CritHitDamagePercent: { 0: 13, 3: 19, 6: 25, 9: 31, 12: 40, 15: 70 },
    EffectivenessPercent: { 0: 12, 3: 18, 6: 24, 9: 30, 12: 39, 15: 65 },
    EffectResistancePercent: { 0: 12, 3: 18, 6: 24, 9: 30, 12: 39, 15: 65 },
    Speed: { 0: 8, 3: 12, 6: 16, 9: 20, 12: 26, 15: 45 }
  },
  88: {
    Attack: { 0: 103, 3: 154, 6: 206, 9: 258, 12: 336, 15: 550 },
    Health: { 0: 553, 3: 829, 6: 1105, 9: 1381, 12: 1795, 15: 2970 },
    Defense: { 0: 62, 3: 93, 6: 124, 9: 155, 12: 202, 15: 320 },
    AttackPercent: { 0: 13, 3: 19, 6: 25, 9: 31, 12: 40, 15: 70 },
    HealthPercent: { 0: 13, 3: 19, 6: 25, 9: 31, 12: 40, 15: 70 },
    DefensePercent: { 0: 13, 3: 19, 6: 25, 9: 31, 12: 40, 15: 70 },
    CritHitChancePercent: { 0: 12, 3: 18, 6: 24, 9: 30, 12: 36, 15: 65 },
    CritHitDamagePercent: { 0: 14, 3: 21, 6: 28, 9: 35, 12: 45, 15: 75 },
    EffectivenessPercent: { 0: 13, 3: 19, 6: 25, 9: 31, 12: 40, 15: 70 },
    EffectResistancePercent: { 0: 13, 3: 19, 6: 25, 9: 31, 12: 40, 15: 70 },
    Speed: { 0: 9, 3: 13, 6: 17, 9: 21, 12: 28, 15: 50 }
  }
};
