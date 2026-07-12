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
  Necklace: ['AttackPercent', 'HealthPercent', 'DefensePercent', 'CritHitChancePercent', 'CritHitDamagePercent'],
  Ring: ['AttackPercent', 'HealthPercent', 'DefensePercent', 'EffectivenessPercent', 'EffectResistancePercent'],
  Boots: ['AttackPercent', 'HealthPercent', 'DefensePercent', 'Speed'],
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

export const MAIN_STAT_TABLE = {
  85: {
    Weapon: { Attack: { base: 35, max: 525 } },
    Helmet: { Health: { base: 320, max: 4725 } },
    Armor: { Defense: { base: 25, max: 350 } },
    Necklace: {
      AttackPercent: { base: 7, max: 65 },
      HealthPercent: { base: 7, max: 65 },
      DefensePercent: { base: 7, max: 65 },
      CritHitChancePercent: { base: 5, max: 60 },
      CritHitDamagePercent: { base: 5, max: 60 },
    },
    Ring: {
      AttackPercent: { base: 7, max: 65 },
      HealthPercent: { base: 7, max: 65 },
      DefensePercent: { base: 7, max: 65 },
      EffectivenessPercent: { base: 8, max: 65 },
      EffectResistancePercent: { base: 8, max: 65 },
    },
    Boots: {
      AttackPercent: { base: 7, max: 65 },
      HealthPercent: { base: 7, max: 65 },
      DefensePercent: { base: 7, max: 65 },
      Speed: { base: 7, max: 45 },
    },
  },
  88: {
    Weapon: { Attack: { base: 37, max: 560 } },
    Helmet: { Health: { base: 340, max: 5050 } },
    Armor: { Defense: { base: 26, max: 375 } },
    Necklace: {
      AttackPercent: { base: 8, max: 70 },
      HealthPercent: { base: 8, max: 70 },
      DefensePercent: { base: 8, max: 70 },
      CritHitChancePercent: { base: 6, max: 65 },
      CritHitDamagePercent: { base: 6, max: 65 },
    },
    Ring: {
      AttackPercent: { base: 8, max: 70 },
      HealthPercent: { base: 8, max: 70 },
      DefensePercent: { base: 8, max: 70 },
      EffectivenessPercent: { base: 9, max: 70 },
      EffectResistancePercent: { base: 9, max: 70 },
    },
    Boots: {
      AttackPercent: { base: 8, max: 70 },
      HealthPercent: { base: 8, max: 70 },
      DefensePercent: { base: 8, max: 70 },
      Speed: { base: 8, max: 48 },
    },
  },
};
