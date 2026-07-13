// E7 Gear Forge - Evaluation Panel Renderer

let currentEvaluation = null;
let containerElement = null;
let currentGear = null;

let activeTab = 'suited'; // 'suited' or 'all'
let searchQuery = '';
let selectedHero = null;
let lastScrollTop = 0;
let activeFilters = null;

const STAT_LABELS = {
  'Attack': 'Attack',
  'AttackPercent': 'Attack %',
  'Defense': 'Defense',
  'DefensePercent': 'Defense %',
  'Health': 'Health',
  'HealthPercent': 'Health %',
  'Speed': 'Speed',
  'CritHitChancePercent': 'Crit Chance %',
  'CritHitDamagePercent': 'Crit Damage %',
  'EffectivenessPercent': 'Effectiveness %',
  'EffectResistancePercent': 'Effect Resist %'
};

const STAT_TO_SAV_KEY = {
  'Attack': 'atk',
  'AttackPercent': 'atk',
  'Defense': 'def',
  'DefensePercent': 'def',
  'Health': 'hp',
  'HealthPercent': 'hp',
  'Speed': 'spd',
  'CritHitChancePercent': 'cc',
  'CritHitDamagePercent': 'cd',
  'EffectivenessPercent': 'eff',
  'EffectResistancePercent': 'res'
};

function formatStatLabel(type) {
  return STAT_LABELS[type] || type;
}

/**
 * Main entry point to render the evaluation panel
 */
export function renderEvaluation(evaluation, containerEl, gear, filters) {
  currentEvaluation = evaluation;
  containerElement = containerEl;
  currentGear = gear;
  activeFilters = filters;

  if (!evaluation) {
    containerEl.innerHTML = `
      <div class="eval-empty-state">
        <p>Forge a gear piece to evaluate</p>
      </div>
    `;
    return;
  }

  // Handle global rule matches (Discard / Speed Check)
  if (evaluation.verdict === 'Discard' || evaluation.verdict === 'Speed Check') {
    const verdictClass = evaluation.verdict === 'Discard' ? 'verdict-discard' : 'verdict-speed-check';
    containerEl.innerHTML = `
      <div class="eval-header">
        <div class="eval-verdict-row">
          <span class="eval-title">Layer 1 Evaluation</span>
          <span class="verdict-badge ${verdictClass}">${evaluation.verdict}</span>
        </div>
      </div>
      <div class="eval-list-wrapper animate-fade-in" style="justify-content: center; align-items: center; text-align: center; color: var(--text-muted); font-size: 0.8125rem; padding: 1.5rem; gap: 0.75rem;">
        <svg style="width: 32px; height: 32px; color: ${evaluation.verdict === 'Discard' ? '#ef4444' : '#f59e0b'};" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
        </svg>
        <span style="font-weight: 700; color: #fff;">Global Rule Triggered</span>
        <p class="eval-rule-matched">${evaluation.globalRuleMatched}</p>
      </div>
    `;
    return;
  }

  // Reset detail view if the gear has changed
  if (selectedHero) {
    const heroExists = currentEvaluation.heroDetails.some(h => h.heroName === selectedHero.heroName);
    if (!heroExists) {
      selectedHero = null;
    }
  }

  draw();
}

/**
 * Internal draw function that renders based on current state variables
 */
function draw() {
  if (!containerElement || !currentEvaluation) return;

  if (selectedHero) {
    renderHeroDetail();
    return;
  }

  const verdictClass = currentEvaluation.verdict === 'Worthy' ? 'verdict-worthy' : 'verdict-sell';
  
  // Calculate suited heroes count (unique heroes that have at least one worthy build)
  const suitedHeroes = currentEvaluation.heroDetails.filter(h => 
    h.builds.some(b => b.verdict === 'Worthy') && matchesFilters(h, activeFilters)
  );
  const suitedCount = suitedHeroes.length;

  containerElement.innerHTML = `
    <div class="eval-header">
      <div class="eval-verdict-row">
        <span class="eval-title">Layer 1 Evaluation</span>
        <span class="verdict-badge ${verdictClass}">${currentEvaluation.verdict}</span>
      </div>
    </div>

    <!-- Tabs -->
    <div class="eval-tabs">
      <button class="eval-tab-btn ${activeTab === 'suited' ? 'active' : ''}" id="tab-suited-btn">
        Suited (${suitedCount})
      </button>
      <button class="eval-tab-btn ${activeTab === 'all' ? 'active' : ''}" id="tab-all-btn">
        All Heroes
      </button>
    </div>

    <!-- Search Bar -->
    <div class="eval-search-container">
      <input type="text" class="eval-search-input" id="eval-search" placeholder="Search hero..." value="${searchQuery}">
    </div>

    <!-- List Wrapper -->
    <div class="eval-list-wrapper" id="eval-list">
      <!-- Dynamic list rows -->
    </div>
  `;

  // Attach event listeners for tabs and search
  document.getElementById('tab-suited-btn').addEventListener('click', () => {
    activeTab = 'suited';
    draw();
  });
  document.getElementById('tab-all-btn').addEventListener('click', () => {
    activeTab = 'all';
    draw();
  });
  
  const searchInput = document.getElementById('eval-search');
  searchInput.addEventListener('input', (e) => {
    searchQuery = e.target.value.trim().toLowerCase();
    renderList();
  });
  
  // Focus at the end of input if user was typing
  searchInput.focus();
  const valLen = searchInput.value.length;
  searchInput.setSelectionRange(valLen, valLen);

  renderList();
}

/**
 * Filters, sorts, and renders rows in the list
 */
function renderList() {
  const listEl = document.getElementById('eval-list');
  if (!listEl || !currentEvaluation) return;

  let html = '';

  // Helpers to calculate sorting parameters
  const getWorthyCount = (hero) => hero.builds.filter(b => b.verdict === 'Worthy').length;
  
  const getMinLandmines = (hero) => {
    let min = Infinity;
    hero.builds.forEach(build => {
      const count = build.landmines ? build.landmines.length : 0;
      if (count < min) min = count;
    });
    return min === Infinity ? 0 : min;
  };

  if (activeTab === 'suited') {
    // Show only heroes that have at least one worthy build
    const suitedHeroes = currentEvaluation.heroDetails.filter(h => 
      h.builds.some(b => b.verdict === 'Worthy') && matchesFilters(h, activeFilters)
    );

    const filteredSuited = suitedHeroes.filter(h => 
      h.heroName.toLowerCase().includes(searchQuery)
    );

    // Sort: 1) Most worthy builds first, 2) Least landmines second, 3) Alphabetical third
    filteredSuited.sort((a, b) => {
      const diff = getWorthyCount(b) - getWorthyCount(a);
      if (diff !== 0) return diff;

      const lmDiff = getMinLandmines(a) - getMinLandmines(b);
      if (lmDiff !== 0) return lmDiff;

      return a.heroName.localeCompare(b.heroName);
    });

    if (filteredSuited.length === 0) {
      html = `<div class="eval-empty-state"><p>No suited heroes found</p></div>`;
    } else {
      filteredSuited.forEach(h => {
        const iconSrc = h.icon ? `/data/heroes/icons/${h.icon}` : '/data/heroes/icons/default.png';
        const totalCount = h.builds.length;
        const worthyCount = getWorthyCount(h);

        // Generate rank pills (worthy = green, not worthy = red)
        let buildsPills = '';
        h.builds.forEach(b => {
          const cls = b.verdict === 'Worthy' ? 'worthy' : 'sell';
          buildsPills += `<span class="build-rank-pill ${cls}" title="Build #${b.rank} - ${b.verdict}">${b.rank}</span>`;
        });

        html += `
          <div class="eval-hero-row animate-fade-in" data-hero="${h.heroName}">
            <div class="eval-hero-info">
              <img class="eval-hero-icon" src="${iconSrc}" onerror="this.src='data:image/svg+xml;utf8,<svg xmlns=%22http://www.w3.org/2000/svg%22 width=%2224%22 height=%2224%22 viewBox=%220 0 24 24%22><circle cx=%2212%22 cy=%2212%22 r=%2210%22 fill=%22%232a2a2a%22/><text x=%2212%22 y=%2216%22 font-size=%2212%22 fill=%22%23888%22 text-anchor=%22middle%22 font-weight=%22bold%22>${h.heroName[0]}</text></svg>';">
              <div style="display: flex; flex-direction: column;">
                <span class="eval-hero-name">${h.heroName}</span>
                <div style="display: flex; align-items: center; margin-top: 0.125rem;">
                  <span class="eval-hero-meta">Builds:</span>
                  ${buildsPills}
                </div>
              </div>
            </div>
            <span class="eval-status-badge eval-status-worthy">${worthyCount}/${totalCount}</span>
          </div>
        `;
      });
    }
  } else {
    // Show all heroes
    const filteredHeroes = currentEvaluation.heroDetails.filter(h => 
      h.heroName.toLowerCase().includes(searchQuery) && matchesFilters(h, activeFilters)
    );

    // Sort: 1) Most worthy builds first, 2) Least landmines second, 3) Alphabetical third
    filteredHeroes.sort((a, b) => {
      const diff = getWorthyCount(b) - getWorthyCount(a);
      if (diff !== 0) return diff;

      const lmDiff = getMinLandmines(a) - getMinLandmines(b);
      if (lmDiff !== 0) return lmDiff;

      return a.heroName.localeCompare(b.heroName);
    });

    if (filteredHeroes.length === 0) {
      html = `<div class="eval-empty-state"><p>No heroes found</p></div>`;
    } else {
      filteredHeroes.forEach(h => {
        const iconSrc = h.icon ? `/data/heroes/icons/${h.icon}` : '/data/heroes/icons/default.png';
        const totalCount = h.builds.length;
        const worthyCount = getWorthyCount(h);
        const hasWorthyBuild = worthyCount > 0;
        const badgeClass = hasWorthyBuild ? 'eval-status-worthy' : 'eval-status-sell';

        // Generate rank pills (worthy = green, not worthy = red)
        let buildsPills = '';
        h.builds.forEach(b => {
          const cls = b.verdict === 'Worthy' ? 'worthy' : 'sell';
          buildsPills += `<span class="build-rank-pill ${cls}" title="Build #${b.rank} - ${b.verdict}">${b.rank}</span>`;
        });

        html += `
          <div class="eval-hero-row animate-fade-in" data-hero="${h.heroName}">
            <div class="eval-hero-info">
              <img class="eval-hero-icon" src="${iconSrc}" onerror="this.src='data:image/svg+xml;utf8,<svg xmlns=%22http://www.w3.org/2000/svg%22 width=%2224%22 height=%2224%22 viewBox=%220 0 24 24%22><circle cx=%2212%22 cy=%2212%22 r=%2210%22 fill=%22%232a2a2a%22/><text x=%2212%22 y=%2216%22 font-size=%2212%22 fill=%22%23888%22 text-anchor=%22middle%22 font-weight=%22bold%22>${h.heroName[0]}</text></svg>';">
              <div style="display: flex; flex-direction: column;">
                <span class="eval-hero-name">${h.heroName}</span>
                <div style="display: flex; align-items: center; margin-top: 0.125rem;">
                  <span class="eval-hero-meta">Builds:</span>
                  ${buildsPills}
                </div>
              </div>
            </div>
            <span class="eval-status-badge ${badgeClass}">${worthyCount}/${totalCount}</span>
          </div>
        `;
      });
    }
  }

  listEl.innerHTML = html;

  // Click handler to open detail view
  listEl.querySelectorAll('.eval-hero-row').forEach(row => {
    row.addEventListener('click', () => {
      const heroName = row.getAttribute('data-hero');
      lastScrollTop = listEl.scrollTop;
      selectedHero = currentEvaluation.heroDetails.find(h => h.heroName === heroName);
      draw();
    });
  });
}

/**
 * Render detail view for a specific selected hero
 */
function renderHeroDetail() {
  if (!containerElement || !selectedHero || !currentEvaluation || !currentGear) return;

  const iconSrc = selectedHero.icon ? `/data/heroes/icons/${selectedHero.icon}` : '/data/heroes/icons/default.png';

  let buildsHtml = '';
  selectedHero.builds.forEach(b => {
    const verdictClass = b.verdict === 'Worthy' ? 'eval-status-worthy' : 'eval-status-sell';
    
    // Create rows for each substat of the gear
    let statsGridHtml = '';
    currentGear.substats.forEach(sub => {
      // Find if this substat is a landmine in this build
      const landmine = b.landmines.find(lm => lm.statType === sub.type);
      const isLandmine = !!landmine;
      
      const savKey = STAT_TO_SAV_KEY[sub.type];
      const savValue = b.sav ? b.sav[savKey] : 0.0;
      
      const rowClass = isLandmine ? 'landmine' : 'valid';
      const textClass = isLandmine ? 'landmine-text' : 'valid-text';
      const label = formatStatLabel(sub.type);
      const formattedSav = typeof savValue === 'number' ? savValue.toFixed(1) : '0.0';

      statsGridHtml += `
        <div class="eval-stat-row ${rowClass}">
          <span class="eval-stat-name">${label}</span>
          <span class="eval-stat-sav ${textClass}">SAV: ${formattedSav}</span>
        </div>
      `;
    });

    buildsHtml += `
      <div class="eval-build-card">
        <div class="eval-build-title">
          <span>Build Rank #${b.rank}</span>
          <span class="eval-status-badge ${verdictClass}">${b.verdict}</span>
        </div>
        <div class="eval-build-sets">${b.sets.join(' / ')} Set</div>
        <div class="eval-stat-grid">
          ${statsGridHtml}
        </div>
      </div>
    `;
  });

  containerElement.innerHTML = `
    <div class="eval-detail-container animate-fade-in">
      <div class="eval-detail-header">
        <button class="eval-back-btn" id="eval-back-btn">
          <svg style="width: 14px; height: 14px; margin-right: 2px;" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/>
          </svg>
          Back
        </button>
        <img class="eval-hero-icon" src="${iconSrc}" onerror="this.src='data:image/svg+xml;utf8,<svg xmlns=%22http://www.w3.org/2000/svg%22 width=%2224%22 height=%2224%22 viewBox=%220 0 24 24%22><circle cx=%2212%22 cy=%2212%22 r=%2210%22 fill=%22%232a2a2a%22/><text x=%2212%22 y=%2216%22 font-size=%2212%22 fill=%22%23888%22 text-anchor=%22middle%22 font-weight=%22bold%22>${selectedHero.heroName[0]}</text></svg>';">
        <span class="eval-hero-name" style="font-size: 0.875rem;">${selectedHero.heroName}</span>
      </div>
      
      <div class="eval-list-wrapper">
        ${buildsHtml}
      </div>
    </div>
  `;

  document.getElementById('eval-back-btn').addEventListener('click', () => {
    selectedHero = null;
    draw();
    const listEl = document.getElementById('eval-list');
    if (listEl) {
      listEl.scrollTop = lastScrollTop;
    }
  });
}

/**
 * Returns true if the hero matches the selected filter options
 */
function matchesFilters(h, filters) {
  if (!filters) return true;
  const roleMatches = filters.roles.includes(h.role);
  const starMatches = filters.stars.includes(h.rarity);
  const elementMatches = filters.elements.includes(h.attribute);
  return roleMatches && starMatches && elementMatches;
}
