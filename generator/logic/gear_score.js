// Epic Seven Equipment Score Calculations

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
