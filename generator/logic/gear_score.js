// Epic Seven Equipment Score and Archetype/Synergy Calculations

export function calculateEquipmentScore(gear) {
  if (!gear) return { gs: 0, es: 0 };

  // GS (Gear Score / WSS) Weights
  const gsWeights = {
    AttackPercent: 1.0,
    DefensePercent: 1.0,
    HealthPercent: 1.0,
    EffectivenessPercent: 1.0,
    EffectResistancePercent: 1.0,
    Speed: 2.0,
    CritHitChancePercent: 1.6,
    CritHitDamagePercent: 8 / 7,
    Attack: 0.088718,
    Defense: 0.160968,
    Health: 0.017759
  };

  // ES (Official Equipment Score) Weights
  const esWeights = {
    AttackPercent: 1.0,
    DefensePercent: 1.0,
    HealthPercent: 1.0,
    EffectivenessPercent: 1.0,
    EffectResistancePercent: 1.0,
    Speed: 2.0,
    CritHitChancePercent: 1.6,
    CritHitDamagePercent: 8 / 7,
    Attack: 3.7373 / 39.4717,
    Defense: 4.9853 / 31.0051,
    Health: 3.1505 / 179.4031
  };

  let gsSubstats = 0;
  let esSubstats = 0;

  for (const sub of gear.substats) {
    gsSubstats += sub.value * (gsWeights[sub.type] || 0);
    esSubstats += sub.value * (esWeights[sub.type] || 0);
  }

  // Main Stat contribution for ES
  let esMain = 0;
  if (gear.main) {
    const mainType = gear.main.type;
    const mainVal = gear.main.value;

    const isPercentMain = mainType.endsWith('Percent') ||
                          mainType === 'CritHitChancePercent' ||
                          mainType === 'CritHitDamagePercent' ||
                          mainType === 'EffectivenessPercent' ||
                          mainType === 'EffectResistancePercent';

    if (isPercentMain) {
      esMain = mainVal / 2.5;
    } else if (mainType === 'Speed') {
      esMain = mainVal * 2.0;
    } else {
      // Flat main stat (Attack, Health, Defense)
      const maxVal85 = mainType === 'Attack' ? 525 : (mainType === 'Health' ? 2835 : 310);
      esMain = (mainVal / maxVal85) * 24.2;
    }
  }

  const finalGS = Math.round(gsSubstats * 10) / 10;
  const finalES = Math.floor(esMain + esSubstats);

  return { gs: finalGS, es: finalES };
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
