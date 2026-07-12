// Epic Seven Equipment Score and Archetype/Synergy Calculations
import { SUBSTAT_ROLL_RANGES } from '../data/gear_data.js';

/**
 * Calculates standard E7 Equipment Score (ES).
 * ES normalizes each substat by dividing by the stat's average roll value, then sums.
 */
export function calculateEquipmentScore(gear) {
  let es = 0;
  for (const sub of gear.substats) {
    const range = SUBSTAT_ROLL_RANGES[sub.type];
    if (!range) continue;
    const avgRoll = (range.min + range.max) / 2;
    es += sub.value / avgRoll;
  }
  return Math.round(es * 10) / 10;
}

/**
 * Infers the best-fit gear archetype based on substat types.
 */
export function inferGearArchetype(gear) {
  const types = new Set(gear.substats.map((s) => s.type));
  const has = (t) => types.has(t);

  const dpsScore =
    (has('AttackPercent') ? 1 : 0) +
    (has('Attack') ? 0.5 : 0) +
    (has('CritHitChancePercent') ? 1 : 0) +
    (has('CritHitDamagePercent') ? 1 : 0) +
    (has('Speed') ? 1 : 0);

  const bruiserScore =
    (has('CritHitChancePercent') ? 1 : 0) +
    (has('CritHitDamagePercent') ? 1 : 0) +
    (has('HealthPercent') ? 1 : 0) +
    (has('DefensePercent') ? 1 : 0) +
    (has('Speed') ? 0.5 : 0);

  const tankScore =
    (has('Speed') ? 1 : 0) +
    (has('EffectResistancePercent') ? 1 : 0) +
    (has('HealthPercent') ? 1 : 0) +
    (has('DefensePercent') ? 1 : 0) +
    (has('Health') ? 0.5 : 0) +
    (has('Defense') ? 0.5 : 0);

  const supportScore =
    (has('Speed') ? 1 : 0) +
    (has('EffectResistancePercent') ? 1 : 0) +
    (has('HealthPercent') ? 1 : 0) +
    (has('EffectivenessPercent') ? 1 : 0) +
    (has('DefensePercent') ? 0.5 : 0);

  const scores = [
    { type: 'dps', score: dpsScore },
    { type: 'bruiser', score: bruiserScore },
    { type: 'tank', score: tankScore },
    { type: 'support', score: supportScore },
  ];
  scores.sort((a, b) => b.score - a.score);

  if (scores[0].score <= 1) return 'mixed';
  return scores[0].type;
}

/**
 * Computes substat synergy and flags conflicting landmine stats.
 */
export function calculateSynergy(gear, archetype) {
  const types = gear.substats.map((s) => s.type);
  const landmines = [];
  const matched = [];

  const DPS_STATS = ['AttackPercent', 'Attack', 'CritHitChancePercent', 'CritHitDamagePercent', 'Speed'];
  const BRUISER_STATS = ['CritHitChancePercent', 'CritHitDamagePercent', 'HealthPercent', 'DefensePercent', 'Speed'];
  const TANK_STATS = ['Speed', 'EffectResistancePercent', 'HealthPercent', 'DefensePercent', 'Health', 'Defense'];
  const SUPPORT_STATS = ['Speed', 'EffectResistancePercent', 'HealthPercent', 'EffectivenessPercent', 'DefensePercent'];

  let goodStats = [];
  let badStats = [];

  switch (archetype) {
    case 'dps':
      goodStats = DPS_STATS;
      badStats = ['EffectivenessPercent', 'EffectResistancePercent'];
      break;
    case 'bruiser':
      goodStats = BRUISER_STATS;
      badStats = ['EffectivenessPercent'];
      break;
    case 'tank':
      goodStats = TANK_STATS;
      badStats = ['CritHitChancePercent', 'CritHitDamagePercent', 'Attack', 'AttackPercent', 'EffectivenessPercent'];
      break;
    case 'support':
      goodStats = SUPPORT_STATS;
      badStats = ['CritHitChancePercent', 'CritHitDamagePercent', 'Attack', 'AttackPercent'];
      break;
    case 'mixed':
      goodStats = ['Speed'];
      badStats = [];
      break;
  }

  for (const t of types) {
    if (goodStats.includes(t)) matched.push(t);
    if (badStats.includes(t)) landmines.push(t);
  }

  const synergy = matched.length / Math.max(1, types.length);
  return { synergy, landmines, matched };
}

/**
 * Checks for a flat main stat on right-side gear (Necklace/Ring/Boots)
 * and warns if the percent version is missing.
 */
export function checkFlatMain(gear) {
  const rightSide = ['Necklace', 'Ring', 'Boots'];
  if (!rightSide.includes(gear.slot)) {
    return { isFlatMain: false, hasPercentSub: false, shouldSell: false };
  }

  const mainType = gear.main.type;
  const isFlatMain = mainType === 'Attack' || mainType === 'Health' || mainType === 'Defense';
  if (!isFlatMain) {
    return { isFlatMain: false, hasPercentSub: false, shouldSell: false };
  }

  const percentVersion =
    mainType === 'Attack' ? 'AttackPercent' : mainType === 'Health' ? 'HealthPercent' : mainType === 'Defense' ? 'DefensePercent' : null;

  const hasPercentSub = percentVersion
    ? gear.substats.some((s) => s.type === percentVersion)
    : false;

  return { isFlatMain, hasPercentSub, shouldSell: !hasPercentSub };
}
