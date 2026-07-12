// Epic Seven Gear Generation and Enhancement Engine
import {
  SLOTS,
  SETS,
  STAT_TYPES,
  FIXED_MAIN_BY_SLOT,
  FLEX_MAIN_BY_SLOT,
  ROLL_RANGES_85,
  ROLL_RANGES_88,
  SPEED_PROBABILITIES,
  MAIN_STAT_PROGRESSION
} from '../data/gear_data.js';

// Helper utilities
export function randomChoice(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}

export function randomInt(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

/**
 * Rolls a substat value using standard uniform ranges OR weighted speed probabilities.
 */
export function rollSubstatValue(type, level, rarity) {
  const cleanRarity = (rarity === 'Epic' || rarity === 'Heroic') ? rarity : 'Epic';
  const cleanLevel = level === 88 ? 88 : 85;

  if (type === 'Speed') {
    const list = SPEED_PROBABILITIES[cleanLevel]?.[cleanRarity] || [];
    if (list.length > 0) {
      const r = Math.random();
      let cumulative = 0;
      for (const item of list) {
        cumulative += item.prob;
        if (r <= cumulative) {
          return item.val;
        }
      }
      return list[list.length - 1].val; // Fallback
    }
  }

  // Uniform distribution for other stats
  const table = cleanLevel === 88 ? ROLL_RANGES_88 : ROLL_RANGES_85;
  const range = table[cleanRarity]?.[type] ?? { min: 1, max: 1 };
  return randomInt(range.min, range.max);
}

/**
 * Returns the main stat value for a slot, stat type, enhancement, and gear level.
 */
export function mainStatValue(slot, type, enhance, level = 85) {
  const cleanLevel = level === 88 ? 88 : 85;
  const stats = MAIN_STAT_PROGRESSION[cleanLevel]?.[type];
  if (!stats) return 0;
  
  const e = Math.max(0, Math.min(15, enhance));
  if (stats[e] !== undefined) {
    return stats[e];
  }
  
  // Interpolate between the two nearest keys
  const keys = [0, 3, 6, 9, 12, 15];
  let lowerKey = 0;
  let upperKey = 15;
  for (let i = 0; i < keys.length; i++) {
    if (keys[i] <= e) lowerKey = keys[i];
    if (keys[i] >= e) {
      upperKey = keys[i];
      break;
    }
  }
  if (lowerKey === upperKey) return stats[lowerKey];
  const fraction = (e - lowerKey) / (upperKey - lowerKey);
  return Math.round(stats[lowerKey] + fraction * (stats[upperKey] - stats[lowerKey]));
}

/**
 * Generates a starting gear piece (+0) based on selected options.
 * Options support array inputs for sets, slots, and mainTypes (multi-select).
 */
export function generateGear(options = {}) {
  // Select Rarity and Level: Pick from selected options
  let rarity = (options.rarities && options.rarities.length > 0)
    ? randomChoice(options.rarities)
    : 'Epic';
  let level = (options.levels && options.levels.length > 0)
    ? randomChoice(options.levels)
    : 85;

  // Enforce game rule: Heroic is always level 85
  if (rarity === 'Heroic') {
    level = 85;
  }

  // Select Slot: Pick from selected slots or random
  const slot = (options.slots && options.slots.length > 0)
    ? randomChoice(options.slots)
    : randomChoice(SLOTS);

  // Select Set: Pick from selected sets or random
  const set = (options.sets && options.sets.length > 0)
    ? randomChoice(options.sets)
    : randomChoice(SETS);

  // Select Main Stat Type based on Slot constraints
  let mainType = null;
  const fixed = FIXED_MAIN_BY_SLOT[slot];
  if (fixed) {
    mainType = fixed; // Left-side gear: weapon, helmet, armor are fixed
  } else {
    // Right-side gear: necklace, ring, boots are flexible
    const allowedMains = FLEX_MAIN_BY_SLOT[slot] || [];
    
    // Intersect user-selected main types with the slot's valid main types
    const selectedAllowed = (options.mainTypes || []).filter(t => allowedMains.includes(t));
    if (selectedAllowed.length > 0) {
      mainType = randomChoice(selectedAllowed);
    } else {
      mainType = randomChoice(allowedMains);
    }
  }

  const mainVal = mainStatValue(slot, mainType, 0, level);

  // Starting substat counts: Epic starts with 4, Heroic starts with 3.
  const startCount = rarity === 'Epic' ? 4 : 3;
  const substats = [];
  const usedTypes = new Set([mainType]);

  // Gather allowed substat types (cannot overlap with main stat or conflict with slot constraints)
  const allowedPool = STAT_TYPES.filter(type => {
    if (usedTypes.has(type)) return false;
    
    // Weapon constraints: no flat Defense, no DefensePercent
    if (slot === 'Weapon' && (type === 'Defense' || type === 'DefensePercent')) {
      return false;
    }
    // Armor constraints: no flat Attack, no AttackPercent
    if (slot === 'Armor' && (type === 'Attack' || type === 'AttackPercent')) {
      return false;
    }
    return true;
  });

  // Shuffle and pick starting substats
  const shuffled = [...allowedPool].sort(() => Math.random() - 0.5);
  const pickCount = Math.min(startCount, shuffled.length);

  for (let i = 0; i < pickCount; i++) {
    const type = shuffled[i];
    const value = rollSubstatValue(type, level, rarity);
    substats.push({ type, value, rolls: 1 });
    usedTypes.add(type);
  }

  return {
    id: `gear-${Date.now()}-${Math.floor(Math.random() * 100000)}`,
    set,
    slot,
    rarity,
    level,
    enhance: 0,
    main: { type: mainType, value: mainVal },
    substats,
    history: []
  };
}

/**
 * Applies one upgrade step (+3) to the gear piece.
 */
export function enhanceStep(gear) {
  if (gear.enhance >= 15) {
    return { gear: { ...gear } };
  }

  const nextEnhance = Math.min(15, gear.enhance + 3);
  const mainVal = mainStatValue(gear.slot, gear.main.type, nextEnhance, gear.level);
  
  // Clone substats
  const substats = gear.substats.map(sub => ({ ...sub }));
  const history = [...(gear.history || [])];
  
  // Heroic gear unlocks 4th sub at +12 (enhance goes from 9 to 12)
  const shouldAddNewSub = gear.rarity === 'Heroic' && substats.length < 4 && gear.enhance === 9;

  if (shouldAddNewSub) {
    const usedTypes = new Set([gear.main.type, ...substats.map(s => s.type)]);
    const pool = STAT_TYPES.filter(type => {
      if (usedTypes.has(type)) return false;
      if (gear.slot === 'Weapon' && (type === 'Defense' || type === 'DefensePercent')) return false;
      if (gear.slot === 'Armor' && (type === 'Attack' || type === 'AttackPercent')) return false;
      return true;
    });

    const newType = pool.length > 0 ? randomChoice(pool) : 'Speed';
    const rolledVal = rollSubstatValue(newType, gear.level, gear.rarity);
    
    substats.push({ type: newType, value: rolledVal, rolls: 1 });
    history.push({
      step: nextEnhance,
      type: 'unlock',
      stat: newType,
      value: rolledVal,
      prevValue: 0,
      newValue: rolledVal
    });
  } else {
    // Upgrade an existing substat
    const idx = Math.floor(Math.random() * substats.length);
    const target = substats[idx];
    const prevVal = target.value;
    const rolledVal = rollSubstatValue(target.type, gear.level, gear.rarity);
    
    target.value += rolledVal;
    target.rolls = (target.rolls || 1) + 1;
    
    history.push({
      step: nextEnhance,
      type: 'upgrade',
      stat: target.type,
      value: rolledVal,
      prevValue: prevVal,
      newValue: target.value
    });
  }

  return {
    gear: {
      ...gear,
      enhance: nextEnhance,
      main: { ...gear.main, value: mainVal },
      substats,
      history
    }
  };
}

/**
 * Enhances gear to a specific target level (+3, +6, +9, +12, +15).
 */
export function enhanceToLevel(gear, targetLevel) {
  let current = { ...gear };
  const target = Math.min(15, Math.max(0, targetLevel));
  
  while (current.enhance < target) {
    const result = enhanceStep(current);
    current = result.gear;
  }
  
  return current;
}
