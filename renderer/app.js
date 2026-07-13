// E7 Gear Forge - Application Controller (Renderer Frontend)
import { initRendererConfig, renderGearCard, renderUpgradeHistory } from './ui/gear_renderer.js';
import { renderEvaluation } from './ui/evaluation_renderer.js';

// Configuration
const API_BASE = 'http://localhost:8080';
const EVALUATOR_API_BASE = 'http://localhost:8081'; 

// State Management
let currentGear = null;
let baseGear = null;
let apiConfig = null;
let currentEvaluation = null;

// DOM Elements - Main Page
const configBtn = document.getElementById('config-btn');
const generateBtn = document.getElementById('generate-btn');
const resetUpgradeBtn = document.getElementById('reset-upgrade-btn');
const gearCardContainer = document.getElementById('gear-card-container');
const historyContainer = document.getElementById('history-container');
const evaluationContainer = document.getElementById('evaluation-container');

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

// Initialization
document.addEventListener('DOMContentLoaded', async () => {
  await fetchConfig();
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
  dialogSaveBtn.addEventListener('click', () => {
    configDialog.close();
    render();
  });
});

/**
 * Fetches initial configuration/metadata from Go Backend API
 */
async function fetchConfig() {
  try {
    const res = await fetch(`${API_BASE}/api/config`);
    if (!res.ok) throw new Error(`HTTP error! status: ${res.status}`);
    apiConfig = await res.json();
    
    // Initialize labels inside gear renderer module
    initRendererConfig(apiConfig);
    
    // Setup dialog options dynamically
    setupDialogOptions();
  } catch (err) {
    console.error('Failed to fetch backend configuration:', err);
    alert('Failed to connect to the Go Generator Backend. Please ensure the backend is running.');
  }
}

/**
 * Dynamically builds the multi-select checkbox checklists in the settings dialog
 */
function setupDialogOptions() {
  if (!apiConfig) return;

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
  slotsPillsGrid.innerHTML = apiConfig.slots.map(slot => `
    <label class="check-pill">
      <input type="checkbox" name="slots" value="${slot}" checked>
      <span>${slot}</span>
    </label>
  `).join('');

  // Main Stats Checkboxes
  mainsPillsGrid.innerHTML = apiConfig.flexibleMainStats.map(type => `
    <label class="check-pill">
      <input type="checkbox" name="mainTypes" value="${type}" checked>
      <span>${apiConfig.statLabels[type] || type}</span>
    </label>
  `).join('');

  // Sets Checkboxes
  setsPillsGrid.innerHTML = apiConfig.sets.map(setKey => `
    <label class="check-pill">
      <input type="checkbox" name="sets" value="${setKey}" checked>
      <span>${apiConfig.setLabels[setKey] || setKey.replace('Set', '')}</span>
    </label>
  `).join('');

  // Hero Roles Checkboxes
  const roles = [
    { value: 'warrior', label: 'Warrior' },
    { value: 'knight', label: 'Knight' },
    { value: 'thief', label: 'Thief' },
    { value: 'ranger', label: 'Ranger' },
    { value: 'mage', label: 'Mage' },
    { value: 'manauser', label: 'Soul Weaver' }
  ];
  document.getElementById('hero-roles-pills-grid').innerHTML = roles.map(role => `
    <label class="check-pill">
      <input type="checkbox" name="heroRoles" value="${role.value}" checked>
      <span>${role.label}</span>
    </label>
  `).join('');

  // Hero Stars Checkboxes
  const stars = [
    { value: '3', label: '3★' },
    { value: '4', label: '4★' },
    { value: '5', label: '5★' }
  ];
  document.getElementById('hero-stars-pills-grid').innerHTML = stars.map(star => `
    <label class="check-pill">
      <input type="checkbox" name="heroStars" value="${star.value}" checked>
      <span>${star.label}</span>
    </label>
  `).join('');

  // Hero Elements Checkboxes
  const elements = [
    { value: 'fire', label: 'Fire' },
    { value: 'ice', label: 'Ice' },
    { value: 'wind', label: 'Wind' },
    { value: 'light', label: 'Light' },
    { value: 'dark', label: 'Dark' }
  ];
  document.getElementById('hero-elements-pills-grid').innerHTML = elements.map(elem => `
    <label class="check-pill">
      <input type="checkbox" name="heroElements" value="${elem.value}" checked>
      <span>${elem.label}</span>
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

  if (!epicBox || !heroicBox || !lvl85Box || !lvl88Box) return;

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

  if (!lvl85Box || !lvl88Box || !epicBox || !heroicBox) return;

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
async function handleGenerate() {
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

  try {
    const res = await fetch(`${API_BASE}/api/generate`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(options)
    });
    if (!res.ok) throw new Error(`HTTP error! status: ${res.status}`);
    
    currentGear = await res.json();
    baseGear = JSON.parse(JSON.stringify(currentGear));

    await fetchEvaluatorData(currentGear);

    updateControls();
    render();
  } catch (err) {
    console.error('Failed to generate gear:', err);
    alert('Failed to generate gear. Please make sure the Go generator backend is running.');
  }
}

/**
 * Enhances current gear directly to target enhancement level (+3, +6, +9, +12, +15)
 */
async function handleEnhance(targetLevel) {
  if (!currentGear || currentGear.enhance >= targetLevel) return;

  const reqBody = {
    gear: currentGear,
    targetLevel
  };

  try {
    const res = await fetch(`${API_BASE}/api/enhance`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(reqBody)
    });
    if (!res.ok) throw new Error(`HTTP error! status: ${res.status}`);
    
    currentGear = await res.json();

    await fetchEvaluatorData(currentGear);
    
    updateControls();
    render();
  } catch (err) {
    console.error('Failed to enhance gear:', err);
    alert('Failed to enhance gear. Please make sure the Go generator backend is running.');
  }
}

/**
 * Restores the gear back to the original +0 state, preserving starting stats
 */
function handleResetUpgrade() {
  if (!baseGear) return;
  currentGear = JSON.parse(JSON.stringify(baseGear));
  
  fetchEvaluatorData(currentGear).then(() => {
    updateControls();
    render();
  });
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
  renderEvaluation(currentEvaluation, evaluationContainer, currentGear, getHeroFilters());
}

/**
 * Returns currently configured hero filters from Options dialog
 */
function getHeroFilters() {
  const roles = Array.from(document.querySelectorAll('input[name="heroRoles"]:checked')).map(el => el.value);
  const stars = Array.from(document.querySelectorAll('input[name="heroStars"]:checked')).map(el => parseInt(el.value, 10));
  const elements = Array.from(document.querySelectorAll('input[name="heroElements"]:checked')).map(el => el.value);
  return { roles, stars, elements };
}

/**
 * Fetches evaluation data from the evaluator backend
 */
async function fetchEvaluatorData(gear) {
  if (!gear) {
    currentEvaluation = null;
    return;
  }
  try {
    const res = await fetch(`${EVALUATOR_API_BASE}/api/evaluate`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(gear)
    });
    if (res.ok) {
      currentEvaluation = await res.json();
    } else {
      currentEvaluation = null;
      console.warn('Evaluator backend returned non-OK status:', res.status);
    }
  } catch (err) {
    console.warn('Evaluator service not reachable or failed:', err);
    currentEvaluation = null;
  }
}
