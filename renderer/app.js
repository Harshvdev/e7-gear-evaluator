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
let allHeroes = null;
let heroConfig = {};
let tempHeroConfig = {};

// DOM Elements - Main Page
const configBtn = document.getElementById('config-btn');
const heroConfigBtn = document.getElementById('hero-config-btn');
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

// DOM Elements - Hero Config Modal Settings
const heroConfigDialog = document.getElementById('hero-config-dialog');
const heroConfigCloseBtn = document.getElementById('hero-config-close-btn');
const heroConfigSaveBtn = document.getElementById('hero-config-save-btn');

// Initialization
document.addEventListener('DOMContentLoaded', async () => {
  await fetchConfig();
  await fetchHeroes();
  updateControls();
  render();

  // Event Listeners - Main UI
  configBtn.addEventListener('click', () => configDialog.showModal());
  heroConfigBtn.addEventListener('click', openHeroConfigDialog);
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

  // Event Listeners - Hero Config Modal
  heroConfigCloseBtn.addEventListener('click', () => heroConfigDialog.close());
  heroConfigSaveBtn.addEventListener('click', saveHeroConfig);
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
    const reqBody = {
      gear: gear,
      excludedBuilds: getExcludedBuildsPayload()
    };
    const res = await fetch(`${EVALUATOR_API_BASE}/api/evaluate`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(reqBody)
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

/**
 * Builds the payload map of excluded builds for the evaluation API
 */
function getExcludedBuildsPayload() {
  const excluded = {};
  if (!allHeroes) return excluded;

  allHeroes.forEach(h => {
    const config = heroConfig[h.name];
    if (!config) return;

    if (!config.enabled) {
      // Exclude all builds of this hero
      excluded[h.name] = h.builds.map(b => b.rank);
    } else {
      // Exclude only builds that are not in the checked builds array
      const excludedRanks = [];
      h.builds.forEach(b => {
        if (!config.builds.includes(b.rank)) {
          excludedRanks.push(b.rank);
        }
      });
      if (excludedRanks.length > 0) {
        excluded[h.name] = excludedRanks;
      }
    }
  });

  return excluded;
}

/**
 * Fetches the list of all heroes from Go Evaluator backend API
 */
async function fetchHeroes() {
  try {
    const res = await fetch(`${EVALUATOR_API_BASE}/api/heroes`);
    if (!res.ok) throw new Error(`HTTP error! status: ${res.status}`);
    allHeroes = await res.json();
    
    // Sort alphabetically by hero name
    allHeroes.sort((a, b) => a.name.localeCompare(b.name));

    // Load saved config
    const stored = localStorage.getItem('e7_hero_config');
    if (stored) {
      try {
        heroConfig = JSON.parse(stored);
      } catch (e) {
        heroConfig = {};
      }
    } else {
      heroConfig = {};
    }

    setupHeroConfigDialog();
  } catch (err) {
    console.error('Failed to fetch heroes list:', err);
  }
}

/**
 * Renders the hero lists and configures listeners for filtering and selection
 */
function setupHeroConfigDialog() {
  if (!allHeroes) return;

  const listContainer = document.getElementById('hero-config-list-container');
  if (!listContainer) return;

  const searchInput = document.getElementById('hero-config-search');
  const roleSelect = document.getElementById('hero-config-filter-role');
  const elementSelect = document.getElementById('hero-config-filter-element');
  const raritySelect = document.getElementById('hero-config-filter-rarity');

  const SHORT_SET_NAMES = {
    'Speed': 'Spd',
    'Health': 'HP',
    'Defense': 'Def',
    'Attack': 'Atk',
    'Critical': 'Crit',
    'Hit': 'Hit',
    'Resistance': 'Res',
    'Immunity': 'Imm',
    'Penetration': 'Pen',
    'Lifesteal': 'LS',
    'Destruction': 'Dest',
    'Counter': 'Counter',
    'Rage': 'Rage',
    'Revenge': 'Rev',
    'Torrent': 'Tor',
    'Injury': 'Inj',
    'Unity': 'Unity',
    'Protection': 'Prot'
  };

  // Render the DOM elements once
  listContainer.innerHTML = allHeroes.map(h => {
    const iconSrc = h.icon ? `/data/heroes/icons/${h.icon}` : '/data/heroes/icons/default.png';
    const buildsHtml = h.builds.map(b => {
      const sets = b.sets.filter(s => s && s !== 'Undefined').map(s => SHORT_SET_NAMES[s] || s).join('/');
      const setsStr = sets ? sets : 'General';
      const usageStr = b.usage ? ` (${b.usage}%)` : '';
      
      const stats = b.averageStats || {};
      const statsHtml = `
        <div class="build-card-stats-grid">
          <div>ATK: ${stats.atk || 0}</div>
          <div>DEF: ${stats.def || 0}</div>
          <div>HP: ${stats.hp || 0}</div>
          <div>SPD: ${stats.spd || 0}</div>
          <div>CRT: ${stats.cc || 0}%</div>
          <div>CDG: ${stats.cd || 0}%</div>
          <div>EFF: ${stats.eff || 0}%</div>
          <div>RES: ${stats.res || 0}%</div>
        </div>
      `;

      return `
        <label class="build-card-check">
          <input type="checkbox" class="build-checkbox" data-hero-name="${h.name}" data-build-rank="${b.rank}">
          <div class="build-card-content">
            <div class="build-card-header">
              <span class="build-card-rank">Build ${b.rank}</span>
              <span class="build-card-sets" title="${setsStr}">${setsStr}${usageStr}</span>
            </div>
            ${statsHtml}
          </div>
        </label>
      `;
    }).join('');

    const formattedRole = h.role === 'manauser' ? 'Soul Weaver' : h.role.charAt(0).toUpperCase() + h.role.slice(1);
    const formattedElement = h.attribute.charAt(0).toUpperCase() + h.attribute.slice(1);

    return `
      <div class="hero-config-row" data-hero-name="${h.name}">
        <div class="hero-config-left">
          <img class="hero-config-icon" src="${iconSrc}" onerror="this.src='data:image/svg+xml;utf8,<svg xmlns=%22http://www.w3.org/2000/svg%22 width=%2224%22 height=%2224%22 viewBox=%220 0 24 24%22><circle cx=%2212%22 cy=%2212%22 r=%2210%22 fill=%22%232a2a2a%22/><text x=%2212%22 y=%2216%22 font-size=%2212%22 fill=%22%23888%22 text-anchor=%22middle%22 font-weight=%22bold%22>${h.name[0]}</text></svg>';" style="width: 40px; height: 40px;">
          <div class="hero-config-info">
            <span class="hero-config-name">${h.name}</span>
            <span class="hero-config-meta">${h.rarity}★ ${formattedRole} | ${formattedElement}</span>
          </div>
          <label class="hero-toggle">
            <input type="checkbox" class="hero-main-checkbox" data-hero-name="${h.name}">
            <span class="hero-toggle-slider"></span>
          </label>
        </div>
        <div class="hero-config-right">
          ${buildsHtml}
        </div>
      </div>
    `;
  }).join('');

  // Filtering function
  const filterList = () => {
    const search = searchInput.value.trim().toLowerCase();
    const role = roleSelect.value;
    const element = elementSelect.value;
    const rarity = raritySelect.value ? parseInt(raritySelect.value, 10) : '';

    const rows = listContainer.querySelectorAll('.hero-config-row');
    rows.forEach(row => {
      const heroName = row.getAttribute('data-hero-name');
      const h = allHeroes.find(x => x.name === heroName);
      if (!h) return;

      const matchesSearch = h.name.toLowerCase().includes(search);
      const matchesRole = !role || h.role === role;
      const matchesElement = !element || h.attribute === element;
      const matchesRarity = !rarity || h.rarity === rarity;

      if (matchesSearch && matchesRole && matchesElement && matchesRarity) {
        row.style.display = 'flex';
      } else {
        row.style.display = 'none';
      }
    });
  };

  // Attach search and filter event listeners
  searchInput.addEventListener('input', filterList);
  roleSelect.addEventListener('change', filterList);
  elementSelect.addEventListener('change', filterList);
  raritySelect.addEventListener('change', filterList);

  // Main checkbox change listener (Include/Exclude hero)
  listContainer.querySelectorAll('.hero-main-checkbox').forEach(checkbox => {
    checkbox.addEventListener('change', (e) => {
      const heroName = e.target.getAttribute('data-hero-name');
      const enabled = e.target.checked;

      if (!tempHeroConfig[heroName]) {
        const h = allHeroes.find(x => x.name === heroName);
        tempHeroConfig[heroName] = { enabled: true, builds: h.builds.map(b => b.rank) };
      }
      tempHeroConfig[heroName].enabled = enabled;

      if (enabled) {
        if (!tempHeroConfig[heroName].builds || tempHeroConfig[heroName].builds.length === 0) {
          tempHeroConfig[heroName].builds = [1];
        }
      } else {
        tempHeroConfig[heroName].builds = [];
      }

      const row = listContainer.querySelector(`.hero-config-row[data-hero-name="${heroName}"]`);
      if (row) {
        if (enabled) {
          row.classList.remove('disabled');
        } else {
          row.classList.add('disabled');
        }
      }

      // Check/uncheck and enable/disable build boxes inside this row
      const buildBoxes = listContainer.querySelectorAll(`.build-checkbox[data-hero-name="${heroName}"]`);
      buildBoxes.forEach(box => {
        box.disabled = !enabled;
        const rank = parseInt(box.getAttribute('data-build-rank'), 10);
        box.checked = enabled && tempHeroConfig[heroName].builds.includes(rank);
      });
    });
  });

  // Build checkbox change listener
  listContainer.querySelectorAll('.build-checkbox').forEach(checkbox => {
    checkbox.addEventListener('change', (e) => {
      const heroName = e.target.getAttribute('data-hero-name');
      const rank = parseInt(e.target.getAttribute('data-build-rank'), 10);
      const checked = e.target.checked;

      if (!tempHeroConfig[heroName]) {
        const h = allHeroes.find(x => x.name === heroName);
        tempHeroConfig[heroName] = { enabled: true, builds: h.builds.map(b => b.rank) };
      }

      if (checked) {
        if (!tempHeroConfig[heroName].builds.includes(rank)) {
          tempHeroConfig[heroName].builds.push(rank);
        }
      } else {
        tempHeroConfig[heroName].builds = tempHeroConfig[heroName].builds.filter(r => r !== rank);
        // If all builds are unchecked, disable the hero
        if (tempHeroConfig[heroName].builds.length === 0) {
          const mainToggle = listContainer.querySelector(`.hero-main-checkbox[data-hero-name="${heroName}"]`);
          if (mainToggle) {
            mainToggle.checked = false;
            mainToggle.dispatchEvent(new Event('change'));
          }
        }
      }
    });
  });

  // Bulk selectors
  document.getElementById('hero-bulk-enable-all').addEventListener('click', () => {
    const rows = listContainer.querySelectorAll('.hero-config-row');
    rows.forEach(row => {
      if (row.style.display === 'none') return; // Only affect visible/filtered heroes
      const name = row.getAttribute('data-hero-name');
      const h = allHeroes.find(x => x.name === name);
      if (h) {
        tempHeroConfig[name] = { enabled: true, builds: h.builds.map(b => b.rank) };
        row.classList.remove('disabled');
        const mainCheckbox = row.querySelector('.hero-main-checkbox');
        if (mainCheckbox) mainCheckbox.checked = true;
        const buildBoxes = row.querySelectorAll('.build-checkbox');
        buildBoxes.forEach(box => {
          box.disabled = false;
          box.checked = true;
        });
      }
    });
  });

  document.getElementById('hero-bulk-disable-all').addEventListener('click', () => {
    const rows = listContainer.querySelectorAll('.hero-config-row');
    rows.forEach(row => {
      if (row.style.display === 'none') return;
      const name = row.getAttribute('data-hero-name');
      if (tempHeroConfig[name]) {
        tempHeroConfig[name].enabled = false;
        tempHeroConfig[name].builds = [];
      } else {
        tempHeroConfig[name] = { enabled: false, builds: [] };
      }
      row.classList.add('disabled');
      const mainCheckbox = row.querySelector('.hero-main-checkbox');
      if (mainCheckbox) mainCheckbox.checked = false;
      const buildBoxes = row.querySelectorAll('.build-checkbox');
      buildBoxes.forEach(box => {
        box.disabled = true;
        box.checked = false;
      });
    });
  });

  document.getElementById('hero-bulk-rank1-only').addEventListener('click', () => {
    const rows = listContainer.querySelectorAll('.hero-config-row');
    rows.forEach(row => {
      if (row.style.display === 'none') return;
      const name = row.getAttribute('data-hero-name');
      tempHeroConfig[name] = { enabled: true, builds: [1] };
      row.classList.remove('disabled');
      const mainCheckbox = row.querySelector('.hero-main-checkbox');
      if (mainCheckbox) mainCheckbox.checked = true;
      const buildBoxes = row.querySelectorAll('.build-checkbox');
      buildBoxes.forEach(box => {
        box.disabled = false;
        const rank = parseInt(box.getAttribute('data-build-rank'), 10);
        box.checked = (rank === 1);
      });
    });
  });
}

/**
 * Prepares the temp copy and opens the modal dialog, resetting any active filters
 */
function openHeroConfigDialog() {
  if (!allHeroes) return;

  // Clone active config
  tempHeroConfig = JSON.parse(JSON.stringify(heroConfig));

  // Reset inputs
  document.getElementById('hero-config-search').value = '';
  document.getElementById('hero-config-filter-role').value = '';
  document.getElementById('hero-config-filter-element').value = '';
  document.getElementById('hero-config-filter-rarity').value = '';

  const listContainer = document.getElementById('hero-config-list-container');
  const rows = listContainer.querySelectorAll('.hero-config-row');
  rows.forEach(row => {
    row.style.display = 'flex';
    const name = row.getAttribute('data-hero-name');
    const h = allHeroes.find(x => x.name === name);
    if (!h) return;

    // Use stored config if available, default to all enabled otherwise
    const config = tempHeroConfig[name] || { enabled: true, builds: h.builds.map(b => b.rank) };
    
    if (config.enabled) {
      row.classList.remove('disabled');
    } else {
      row.classList.add('disabled');
    }

    const mainCheckbox = row.querySelector('.hero-main-checkbox');
    if (mainCheckbox) mainCheckbox.checked = config.enabled;

    const buildBoxes = row.querySelectorAll('.build-checkbox');
    buildBoxes.forEach(box => {
      const rank = parseInt(box.getAttribute('data-build-rank'), 10);
      box.checked = config.enabled && config.builds.includes(rank);
      box.disabled = !config.enabled;
    });
  });

  heroConfigDialog.showModal();
}

/**
 * Saves tempHeroConfig to heroConfig, writes to localStorage, and updates evaluation
 */
function saveHeroConfig() {
  heroConfig = JSON.parse(JSON.stringify(tempHeroConfig));
  localStorage.setItem('e7_hero_config', JSON.stringify(heroConfig));
  heroConfigDialog.close();

  // Re-evaluate if gear is loaded
  if (currentGear) {
    fetchEvaluatorData(currentGear).then(() => {
      render();
    });
  }
}

