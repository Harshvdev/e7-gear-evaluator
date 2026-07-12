import { STAT_LABELS, SET_LABELS } from '../data/gear_data.js';
import { calculateEquipmentScore } from '../logic/gear_score.js';

// Rarity theme configuration
const RARITY_THEMES = {
  Epic: {
    badgeClass: 'bg-epic',
    cardBorder: 'border-epic',
    titleColor: 'text-epic',
  },
  Heroic: {
    badgeClass: 'bg-heroic',
    cardBorder: 'border-heroic',
    titleColor: 'text-heroic',
  }
};

/**
 * Format stat values (add % for percentage-based stats)
 */
function formatStatValue(type, value) {
  const isPercent = type.endsWith('Percent');
  return `${value}${isPercent ? '%' : ''}`;
}

/**
 * Clean up stat labels for display
 */
function getStatLabel(type) {
  return STAT_LABELS[type] || type;
}

/**
 * Remove percentage symbols from the label name since it is appended to the values
 */
function getCleanStatLabel(type) {
  const label = getStatLabel(type);
  return label.replace(' %', '').replace('%', '').trim();
}

/**
 * Renders the simplified gear card inside the container
 */
export function renderGearCard(gear, containerEl) {
  if (!gear) {
    containerEl.innerHTML = `
      <div class="card-empty-state">
        <p>Forge a gear piece to start</p>
      </div>
    `;
    return;
  }

  const theme = RARITY_THEMES[gear.rarity] || RARITY_THEMES.Epic;
  const setLabel = SET_LABELS[gear.set] || gear.set.replace('Set', '');

  // Generate substats HTML
  let substatsHtml = '';
  gear.substats.forEach(sub => {
    const rollCount = sub.rolls || 1;
    let dots = '';
    for (let i = 0; i < rollCount; i++) {
      dots += `<span class="roll-dot" title="Roll ${i+1}"></span>`;
    }

    substatsHtml += `
      <div class="substat-row">
        <div class="substat-info">
          <span class="substat-label">${getStatLabel(sub.type)}</span>
          <div class="roll-dots-container">
            ${dots}
          </div>
        </div>
        <span class="substat-value">${formatStatValue(sub.type, sub.value)}</span>
      </div>
    `;
  });

  const { gs, es } = calculateEquipmentScore(gear);

  containerEl.innerHTML = `
    <div class="gear-card ${gear.rarity.toLowerCase()} ${theme.cardBorder}">
      <div class="card-bg-overlay"></div>

      <!-- Card Top Info -->
      <div class="card-header">
        <div class="badge-group">
          <span class="rarity-badge ${theme.badgeClass}">${gear.rarity}</span>
          <span class="level-label">Lv. ${gear.level}</span>
        </div>
        <span class="enhance-badge">+${gear.enhance}</span>
      </div>

      <!-- Card Title (Slot & Set) -->
      <div class="card-title-group">
        <h3 class="card-title ${theme.titleColor}">${gear.slot}</h3>
        <p class="card-subtitle">${setLabel} Set</p>
      </div>

      <!-- Main Stat Box -->
      <div class="main-stat-box">
        <div class="main-stat-meta">
          <span class="main-stat-header">Main Stat</span>
          <span class="main-stat-name">${getStatLabel(gear.main.type)}</span>
        </div>
        <span class="main-stat-value">${formatStatValue(gear.main.type, gear.main.value)}</span>
      </div>

      <!-- Substats List -->
      <div class="substats-section">
        <span class="substats-header">Substats</span>
        <div class="substats-list">
          ${substatsHtml}
        </div>
      </div>

      <!-- Scores Section -->
      <div class="card-scores-row" style="margin-top: auto; display: flex; justify-content: space-between; border-top: 1px solid rgba(255, 255, 255, 0.08); padding-top: 0.875rem;">
        <div class="score-item" style="display: flex; flex-direction: column; gap: 0.125rem;">
          <span style="font-size: 0.625rem; font-weight: 700; text-transform: uppercase; color: var(--text-muted); letter-spacing: 0.05em;">Gear Score (GS)</span>
          <span style="font-size: 1.15rem; font-weight: 900; color: #fff; font-family: var(--font-display);">${gs}</span>
        </div>
        <div class="score-item" style="display: flex; flex-direction: column; gap: 0.125rem; align-items: flex-end;">
          <span style="font-size: 0.625rem; font-weight: 700; text-transform: uppercase; color: var(--text-muted); letter-spacing: 0.05em;">Equipment Score (ES)</span>
          <span style="font-size: 1.15rem; font-weight: 900; color: var(--primary); font-family: var(--font-display);">${es}</span>
        </div>
      </div>
    </div>
  `;
}

/**
 * Renders the compact, short-text enhancement history
 */
export function renderUpgradeHistory(gear, containerEl) {
  if (!gear || !gear.history || gear.history.length === 0) {
    containerEl.innerHTML = `
      <div class="log-empty-text">No rolls recorded.</div>
    `;
    return;
  }

  let listHtml = '';
  gear.history.forEach((h) => {
    const isUnlock = h.type === 'unlock';
    const cleanLabel = getCleanStatLabel(h.stat);
    
    // Format precisely and short: "Speed: 3 → 6" or "Effectiveness: Unlocked (8%)"
    let actionText = '';
    if (isUnlock) {
      actionText = `${cleanLabel}: Unlocked (${formatStatValue(h.stat, h.newValue)})`;
    } else {
      actionText = `${cleanLabel}: ${formatStatValue(h.stat, h.prevValue)} → ${formatStatValue(h.stat, h.newValue)}`;
    }

    listHtml += `
      <div class="log-row-item">
        <span class="log-step-tag">+${h.step}</span>
        <span class="log-action-text ${isUnlock ? 'unlock' : ''}" title="${actionText}">
          ${actionText}
        </span>
      </div>
    `;
  });

  containerEl.innerHTML = `
    <div class="log-list">
      ${listHtml}
    </div>
  `;
}
