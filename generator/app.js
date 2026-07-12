// E7 Gear Forge - Application Controller
import { SLOTS, SETS, SET_LABELS, STAT_LABELS } from './data/gear_data.js';
import { generateGear, enhanceToLevel } from './logic/gear_engine.js';
import { renderGearCard, renderUpgradeHistory } from './ui/gear_renderer.js';

// State Management
let currentGear = null;
let baseGear = null;

// DOM Elements - Main Page
const configBtn = document.getElementById('config-btn');
const generateBtn = document.getElementById('generate-btn');
const resetUpgradeBtn = document.getElementById('reset-upgrade-btn');
const gearCardContainer = document.getElementById('gear-card-container');
const historyContainer = document.getElementById('history-container');

// DOM Elements - Enhance Buttons
const enhance3Btn = document.getElementById('enhance-3-btn');
const enhance6Btn = document.getElementById('enhance-6-btn');
const enhance9Btn = document.getElementById('enhance-9-btn');
const enhance12Btn = document.getElementById('enhance-12-btn');
const enhance15Btn = document.getElementById('enhance-15-btn');

// DOM Elements - Modal Settings
const configDialog = document.getElementById('config-dialog');
const dialogCloseBtn = document.getElementById('dialog-close-btn');
const dialogSaveBtn = document.getElementById('dialog-save-btn');
const raritiesPillsGrid = document.getElementById('rarities-pills-grid');
const levelsPillsGrid = document.getElementById('levels-pills-grid');
const slotsPillsGrid = document.getElementById('slots-pills-grid');
const mainsPillsGrid = document.getElementById('mains-pills-grid');
const setsPillsGrid = document.getElementById('sets-pills-grid');

// Flexible main stat types for Necklace/Ring/Boots
const FLEXIBLE_MAIN_STATS = [
  'AttackPercent', 'HealthPercent', 'DefensePercent', 'Speed',
  'CritHitChancePercent', 'CritHitDamagePercent', 'EffectivenessPercent', 'EffectResistancePercent'
];

// Initialization
document.addEventListener('DOMContentLoaded', () => {
  setupDialogOptions();
  updateControls();
  render();

  // Event Listeners - Main UI
  configBtn.addEventListener('click', () => configDialog.showModal());
  generateBtn.addEventListener('click', handleGenerate);
  resetUpgradeBtn.addEventListener('click', handleResetUpgrade);

  // Event Listeners - Enhancements
  enhance3Btn.addEventListener('click', () => handleEnhance(3));
  enhance6Btn.addEventListener('click', () => handleEnhance(6));
  enhance9Btn.addEventListener('click', () => handleEnhance(9));
  enhance12Btn.addEventListener('click', () => handleEnhance(12));
  enhance15Btn.addEventListener('click', () => handleEnhance(15));

  // Event Listeners - Modal
  dialogCloseBtn.addEventListener('click', () => configDialog.close());
  dialogSaveBtn.addEventListener('click', () => configDialog.close());
});

/**
 * Dynamically builds the multi-select checkbox checklists in the settings dialog
 */
function setupDialogOptions() {
  // Rarities Checkboxes
  raritiesPillsGrid.innerHTML = ['Epic', 'Heroic'].map(rarity => `
    <label class="check-pill">
      <input type="checkbox" name="rarities" value="${rarity}" checked>
      <span>${rarity}</span>
    </label>
  `).join('');

  // Levels Checkboxes
  levelsPillsGrid.innerHTML = [85, 88].map(lvl => `
    <label class="check-pill">
      <input type="checkbox" name="levels" value="${lvl}" checked>
      <span>Lv. ${lvl}</span>
    </label>
  `).join('');

  // Slots Checkboxes
  slotsPillsGrid.innerHTML = SLOTS.map(slot => `
    <label class="check-pill">
      <input type="checkbox" name="slots" value="${slot}" checked>
      <span>${slot}</span>
    </label>
  `).join('');

  // Main Stats Checkboxes
  mainsPillsGrid.innerHTML = FLEXIBLE_MAIN_STATS.map(type => `
    <label class="check-pill">
      <input type="checkbox" name="mainTypes" value="${type}" checked>
      <span>${STAT_LABELS[type] || type}</span>
    </label>
  `).join('');

  // Sets Checkboxes
  setsPillsGrid.innerHTML = SETS.map(setKey => `
    <label class="check-pill">
      <input type="checkbox" name="sets" value="${setKey}" checked>
      <span>${SET_LABELS[setKey] || setKey.replace('Set', '')}</span>
    </label>
  `).join('');

  // Attach change validation listeners
  document.querySelectorAll('input[name="rarities"]').forEach(el => {
    el.addEventListener('change', handleRarityChange);
  });
  document.querySelectorAll('input[name="levels"]').forEach(el => {
    el.addEventListener('change', handleLevelChange);
  });

  // Setup initial constraint states
  handleRarityChange();
  handleLevelChange();
}

/**
 * Ensures Heroic rarity locks level to 85.
 * If Epic is unchecked, Level 88 must be unchecked and disabled.
 */
function handleRarityChange(e) {
  const epicBox = document.querySelector('input[name="rarities"][value="Epic"]');
  const heroicBox = document.querySelector('input[name="rarities"][value="Heroic"]');
  const lvl85Box = document.querySelector('input[name="levels"][value="85"]');
  const lvl88Box = document.querySelector('input[name="levels"][value="88"]');

  // Prevent both rarities from being unchecked simultaneously
  if (!epicBox.checked && !heroicBox.checked) {
    if (e && e.target) {
      e.target.checked = true;
    } else {
      epicBox.checked = true;
    }
    return;
  }

  if (!epicBox.checked && heroicBox.checked) {
    // Only Heroic is selected: Level 88 must be disabled and unchecked
    lvl88Box.checked = false;
    lvl88Box.disabled = true;
    // Ensure Level 85 is checked so we don't have all levels unchecked
    lvl85Box.checked = true;
  } else {
    // Epic is selected: Level 88 is valid to select
    // Only enable Level 88 if Level 85 checkbox is not the only one checked, or check level constraints
    lvl88Box.disabled = false;
  }
}

/**
 * Ensures Level 88 option locks rarity to Epic.
 * If Level 85 is unchecked, Heroic rarity must be unchecked and disabled.
 */
function handleLevelChange(e) {
  const lvl85Box = document.querySelector('input[name="levels"][value="85"]');
  const lvl88Box = document.querySelector('input[name="levels"][value="88"]');
  const epicBox = document.querySelector('input[name="rarities"][value="Epic"]');
  const heroicBox = document.querySelector('input[name="rarities"][value="Heroic"]');

  // Prevent both levels from being unchecked simultaneously
  if (!lvl85Box.checked && !lvl88Box.checked) {
    if (e && e.target) {
      e.target.checked = true;
    } else {
      lvl85Box.checked = true;
    }
    return;
  }

  if (!lvl85Box.checked && lvl88Box.checked) {
    // Only Level 88 is selected: Heroic is invalid, so disable and uncheck it
    heroicBox.checked = false;
    heroicBox.disabled = true;
    // Ensure Epic is checked so we don't have all rarities unchecked
    epicBox.checked = true;
  } else {
    // Level 85 is selected: Heroic is valid to select (provided Epic is checked)
    if (epicBox.checked) {
      heroicBox.disabled = false;
    }
  }
}

/**
 * Reads multi-select settings from checkboxes and generates a new +0 gear piece
 */
function handleGenerate() {
  // Extract checked arrays
  const rarities = Array.from(document.querySelectorAll('input[name="rarities"]:checked')).map(el => el.value);
  const levels = Array.from(document.querySelectorAll('input[name="levels"]:checked')).map(el => parseInt(el.value, 10));
  const slots = Array.from(document.querySelectorAll('input[name="slots"]:checked')).map(el => el.value);
  const sets = Array.from(document.querySelectorAll('input[name="sets"]:checked')).map(el => el.value);
  const mainTypes = Array.from(document.querySelectorAll('input[name="mainTypes"]:checked')).map(el => el.value);

  // Validations
  if (rarities.length === 0) {
    alert('Please select at least one gear rarity.');
    configDialog.showModal();
    return;
  }
  if (levels.length === 0) {
    alert('Please select at least one gear level.');
    configDialog.showModal();
    return;
  }
  if (slots.length === 0) {
    alert('Please select at least one gear slot.');
    configDialog.showModal();
    return;
  }
  if (sets.length === 0) {
    alert('Please select at least one gear set.');
    configDialog.showModal();
    return;
  }

  const options = {
    rarities,
    levels,
    slots,
    sets,
    mainTypes
  };

  currentGear = generateGear(options);
  baseGear = JSON.parse(JSON.stringify(currentGear));

  updateControls();
  render();
}

/**
 * Enhances current gear directly to target enhancement level (+3, +6, +9, +12, +15)
 */
function handleEnhance(targetLevel) {
  if (!currentGear || currentGear.enhance >= targetLevel) return;
  currentGear = enhanceToLevel(currentGear, targetLevel);
  
  updateControls();
  render();
}

/**
 * Restores the gear back to the original +0 state, preserving starting stats
 */
function handleResetUpgrade() {
  if (!baseGear) return;
  currentGear = JSON.parse(JSON.stringify(baseGear));
  
  updateControls();
  render();
}

/**
 * Manages enabled/disabled states and styling classes of buttons
 */
function updateControls() {
  const hasGear = currentGear !== null;
  const enhanceVal = hasGear ? currentGear.enhance : 0;

  // Enable re-roll if enhanced
  resetUpgradeBtn.disabled = !hasGear || enhanceVal === 0;

  // Set individual target enhance buttons
  enhance3Btn.disabled = !hasGear || enhanceVal >= 3;
  enhance6Btn.disabled = !hasGear || enhanceVal >= 6;
  enhance9Btn.disabled = !hasGear || enhanceVal >= 9;
  enhance12Btn.disabled = !hasGear || enhanceVal >= 12;
  enhance15Btn.disabled = !hasGear || enhanceVal >= 15;

  // Add / remove active state highlights
  const btns = [
    { el: enhance3Btn, val: 3 },
    { el: enhance6Btn, val: 6 },
    { el: enhance9Btn, val: 9 },
    { el: enhance12Btn, val: 12 },
    { el: enhance15Btn, val: 15 }
  ];

  btns.forEach(btn => {
    if (hasGear && enhanceVal === btn.val) {
      btn.el.classList.add('active-enhance');
    } else {
      btn.el.classList.remove('active-enhance');
    }
  });
}

/**
 * Redraws UI
 */
function render() {
  renderGearCard(currentGear, gearCardContainer);
  renderUpgradeHistory(currentGear, historyContainer);
}
