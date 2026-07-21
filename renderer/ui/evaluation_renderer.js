let currentEvaluation = null;
let containerElement = null;
let currentGear = null;

let activeTab = 'suited'; // 'suited' | 'all' | 'trace'
let searchQuery = '';
let activeFilters = null;
let selectedHero = null;
let lastScrollTop = 0;

// Exported main render function
export function renderEvaluation(evaluation, containerEl, gear, filters) {
	containerElement = containerEl;
	currentGear = gear;
	activeFilters = filters;
	currentEvaluation = evaluation;

	if (!evaluation) {
		containerEl.innerHTML = `
			<div class="eval-empty-state">
				<p>Forge a gear piece to evaluate</p>
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

	// Map verdict class for header badge
	let badgeClass = 'verdict-sell';
	let verdictLabel = currentEvaluation.verdict;

	if (currentEvaluation.trace) {
		switch (currentEvaluation.trace.verdict) {
		case "KEEP_ENHANCE":
			badgeClass = "verdict-keep-enhance";
			verdictLabel = "Keep & Enhance";
			break;
		case "SALVAGE_MOD":
			badgeClass = "verdict-salvage-mod";
			verdictLabel = "Salvage Candidate";
			break;
		case "SPEED_VAULT":
			badgeClass = "verdict-speed-vault";
			verdictLabel = "Speed Vaulted";
			break;
		case "SELL_EXTRACT":
			if (currentEvaluation.verdict === "Discard") {
				badgeClass = "verdict-discard";
				verdictLabel = "Discard";
			} else {
				badgeClass = "verdict-sell-extract";
				verdictLabel = "Sell / Extract";
			}
			break;
		}
	} else {
		// Fallback
		if (currentEvaluation.verdict === 'Worthy') {
			badgeClass = 'verdict-worthy';
		} else if (currentEvaluation.verdict === 'Discard') {
			badgeClass = 'verdict-discard';
		} else if (currentEvaluation.verdict === 'Speed Check') {
			badgeClass = 'verdict-speed-check';
		}
	}
	
	// Calculate suited heroes count (unique heroes that have at least one worthy build)
	const suitedHeroes = currentEvaluation.heroDetails.filter(h => 
		h.builds.some(b => b.verdict === 'Worthy') && matchesFilters(h, activeFilters)
	);
	const suitedCount = suitedHeroes.length;

	containerElement.innerHTML = `
		<div class="eval-header">
			<div class="eval-verdict-row">
				<span class="eval-title">Decision Pipeline</span>
				<span class="verdict-badge ${badgeClass}">${verdictLabel}</span>
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
			<button class="eval-tab-btn ${activeTab === 'trace' ? 'active' : ''}" id="tab-trace-btn">
				Pipeline Trace
			</button>
		</div>

		<!-- Conditional Search Bar / Content -->
		${activeTab !== 'trace' ? `
			<div class="eval-search-container">
				<input type="text" class="eval-search-input" id="eval-search" placeholder="Search hero..." value="${searchQuery}">
			</div>
		` : ''}

		<!-- List Wrapper -->
		<div class="eval-list-wrapper" id="eval-list">
			<!-- Dynamic list rows -->
		</div>
	`;

	// Attach event listeners for tabs
	document.getElementById('tab-suited-btn').addEventListener('click', () => {
		activeTab = 'suited';
		draw();
	});
	document.getElementById('tab-all-btn').addEventListener('click', () => {
		activeTab = 'all';
		draw();
	});
	document.getElementById('tab-trace-btn').addEventListener('click', () => {
		activeTab = 'trace';
		draw();
	});
	
	if (activeTab !== 'trace') {
		const searchInput = document.getElementById('eval-search');
		searchInput.addEventListener('input', (e) => {
			searchQuery = e.target.value.trim().toLowerCase();
			renderList();
		});
		
		// Focus at the end of input if user was typing
		searchInput.focus();
		const valLen = searchInput.value.length;
		searchInput.setSelectionRange(valLen, valLen);
	}

	renderList();
}

/**
 * Filters, sorts, and renders rows in the list
 */
function renderList() {
	const listEl = document.getElementById('eval-list');
	if (!listEl || !currentEvaluation) return;

	// Render Pipeline Trace Tab
	if (activeTab === 'trace') {
		renderPipelineTrace(listEl);
		return;
	}

	let html = '';

	// Helpers to calculate sorting parameters
	const getWorthyCount = (hero) => hero.builds.filter(b => b.verdict === 'Worthy').length;

	// Best WAS% across all builds of a hero (higher = more aligned)
	const getBestWasPct = (hero) => {
		let best = -999;
		hero.builds.forEach(b => {
			const val = typeof b.wasPct === 'number' ? b.wasPct : -999;
			if (val > best) best = val;
		});
		return best === -999 ? 0 : best;
	};

	if (activeTab === 'suited') {
		// Show only heroes that have at least one worthy build
		const suitedHeroes = currentEvaluation.heroDetails.filter(h => 
			h.builds.some(b => b.verdict === 'Worthy') && matchesFilters(h, activeFilters)
		);

		const filteredSuited = suitedHeroes.filter(h => 
			h.heroName.toLowerCase().includes(searchQuery)
		);

		// Sort: 1) Most worthy builds first, 2) Best WAS% second, 3) Alphabetical third
		filteredSuited.sort((a, b) => {
			const diff = getWorthyCount(b) - getWorthyCount(a);
			if (diff !== 0) return diff;

			const wasDiff = getBestWasPct(b) - getBestWasPct(a);
			if (wasDiff !== 0) return wasDiff;

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

		// Sort: 1) Most worthy builds first, 2) Best WAS% second, 3) Alphabetical third
		filteredHeroes.sort((a, b) => {
			const diff = getWorthyCount(b) - getWorthyCount(a);
			if (diff !== 0) return diff;

			const wasDiff = getBestWasPct(b) - getBestWasPct(a);
			if (wasDiff !== 0) return wasDiff;

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

		// WAS / heroFit% progress bar
		const wasPct = b.wasPct || 0;
		const barPct = Math.min(Math.max(wasPct, 0), 100);
		
		let fitClassLabel = "REJECT";
		let barColor = "#ef4444";
		if (wasPct >= 70.0) {
			fitClassLabel = "CORE";
			barColor = "#10b981";
		} else if (wasPct >= 45.0) {
			fitClassLabel = "USABLE";
			barColor = "#3b82f6";
		} else if (wasPct >= 20.0) {
			fitClassLabel = "MARGINAL";
			barColor = "#f59e0b";
		}

		const wasLabel = `heroFit%: ${wasPct.toFixed(1)}% (${fitClassLabel})`;

		// Create rows for each substat of the gear
		let statsGridHtml = '';
		currentGear.substats.forEach(sub => {
			const landmine = b.landmines.find(lm => lm.statType === sub.type);
			const isLandmine = !!landmine;

			const savKey = STAT_TO_SAV_KEY[sub.type];
			const savValue = b.sav ? b.sav[savKey] : 0.0;

			const rowClass = isLandmine ? 'landmine' : 'valid';
			const textClass = isLandmine ? 'landmine-text' : 'valid-text';
			const label = formatStatLabel(sub.type);
			const formattedSav = typeof savValue === 'number' ? savValue.toFixed(1) : '0.0';
			const weightNote = isLandmine ? ` (Low Priority)` : '';

			statsGridHtml += `
				<div class="eval-stat-row ${rowClass}">
					<span class="eval-stat-name">${label}</span>
					<span class="eval-stat-sav ${textClass}">Prio Weight: ${formattedSav}${weightNote}</span>
				</div>
			`;
		});

		buildsHtml += `
			<div class="eval-build-card">
				<div class="eval-build-title">
					<span>Build Rank #${b.rank}</span>
					<div style="display: flex; gap: 0.375rem; align-items: center;">
						<span class="build-rank-pill ${verdictClass}" style="font-size: 0.6875rem; padding: 2px 6px; border-radius: 4px; background: ${barColor}; color: #000; font-weight: 700;">${fitClassLabel}</span>
						<span class="eval-status-badge ${verdictClass}">${b.verdict}</span>
					</div>
				</div>
				<div class="eval-build-sets">${b.sets.join(' / ')} Set</div>
				<div class="eval-was-bar-container" title="heroFit%: ${wasPct.toFixed(1)}%">
					<div class="eval-was-bar-track">
						<div class="eval-was-bar-fill" style="width: ${barPct}%; background: ${barColor};"></div>
					</div>
					<span class="eval-was-label" style="color: ${barColor};">${wasLabel}</span>
				</div>
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
 * Render the interactive 7-Layer Pipeline Trace
 */
function renderPipelineTrace(listEl) {
	const trace = currentEvaluation.trace;
	if (!trace) {
		listEl.innerHTML = `<div class="eval-empty-state"><p>No trace data available</p></div>`;
		return;
	}

	// L0 Ingest info
	const l0Rolls = Object.entries(trace.l0.rollReconstruction || {})
		.map(([stat, rolls]) => `<div class="trace-detail-item"><span class="trace-detail-label">${formatStatLabel(stat)}</span><span class="trace-detail-val">${rolls} ${rolls === 1 ? 'roll' : 'rolls'}</span></div>`)
		.join('');

	// L1 Legality check violations
	const l1Status = trace.l1.violations.length === 0 ? 'status-pass' : 'status-fail';
	const l1StatusLabel = trace.l1.violations.length === 0 ? 'Passed' : 'Failed';
	const l1ViolationsList = trace.l1.violations.map(v => `<li style="color: #ef4444; margin: 0.25rem 0;">Violation: <strong>${v}</strong></li>`).join('');
	const l1Body = trace.l1.violations.length === 0 
		? `<p>Passed all 10 game rule validation invariants (R1.1 - R1.10).</p>`
		: `<ul>${l1ViolationsList}</ul>`;

	// L2 Universal Discard
	const l2Status = trace.l2.fired ? 'status-fail' : 'status-pass';
	const l2StatusLabel = trace.l2.fired ? 'Fired (Discard)' : 'Passed';
	const l2Body = trace.l2.fired
		? `<p style="color: #ef4444; font-weight: 700;">${trace.l2.detail}</p>`
		: `<p>Passed. Gear does not meet the flat-main worthless pre-filters.</p>`;

	// L3 Speed Check Toggle
	const l3Status = trace.l3.tagged ? 'status-tag' : 'status-neutral';
	const l3StatusLabel = trace.l3.tagged ? 'Speed Vaulted' : 'Skipped';
	const l3Body = `<p>Speed Check Option: <strong>${trace.l3.toggle}</strong>. Speed Substat value: <strong>${trace.l3.speedValue}</strong>.</p>
					${trace.l3.tagged ? `<p style="color: #8b5cf6; font-weight: 700; margin-top: 0.25rem;">Tagged for Speed Vault: this piece survives regardless of hero matching.</p>` : ''}`;

	// L4 Hero Matching
	const l4MatchesCount = trace.l4.perHero ? trace.l4.perHero.filter(h => h.pass).length : 0;
	const l4Status = l4MatchesCount > 0 ? 'status-pass' : (trace.l3.tagged ? 'status-tag' : 'status-fail');
	const l4StatusLabel = l4MatchesCount > 0 ? `Matched ${l4MatchesCount}` : (trace.l3.tagged ? 'Vault Bypassed' : '0 Matches');
	
	let l4MatchesHtml = '';
	if (trace.l4.perHero && trace.l4.perHero.length > 0) {
		const topMatches = trace.l4.perHero.filter(h => h.pass || h.fitScore > 0);
		topMatches.sort((a, b) => b.fitScore - a.fitScore);
		
		// limit to top 5
		const limitMatches = topMatches.slice(0, 5);
		const matchesRows = limitMatches.map(m => {
			const setVal = m.gateResults[0] ? '✓' : '✗';
			const mainVal = m.gateResults[1] ? '✓' : '✗';
			const rangeVal = m.gateResults[2] ? '✓' : '✗';
			const scoreVal = m.gateResults[3] ? '✓' : '✗';
			const rowColor = m.pass ? '#10b981' : '#64748b';
			return `<div style="display:flex; justify-content:space-between; margin:0.25rem 0; color:${rowColor};">
				<span>${m.heroName} (Build #${m.buildRank})</span>
				<span>Fit: ${m.fitScore.toFixed(1)} / ${m.threshold.toFixed(1)} [S:${setVal} M:${mainVal} R:${rangeVal} F:${scoreVal}]</span>
			</div>`;
		}).join('');
		l4MatchesHtml = `<div style="margin-top: 0.5rem; border-top: 1px solid rgba(255,255,255,0.05); padding-top: 0.5rem;">
			<div style="font-weight:700; margin-bottom: 0.25rem; color:#fff;">Top Build Matches:</div>
			${matchesRows}
		</div>`;
	}

	// L5 Quality & Projection
	const expectedScen = trace.l5.scenarios ? trace.l5.scenarios.find(s => s.scenarioName === 'expected') : null;
	const bestScen = trace.l5.scenarios ? trace.l5.scenarios.find(s => s.scenarioName === 'best') : null;
	const worstScen = trace.l5.scenarios ? trace.l5.scenarios.find(s => s.scenarioName === 'worst') : null;

	const wssCurrent = trace.l5.currentWss || 0;
	const wssExpected = expectedScen ? expectedScen.score : 0;
	const wssBest = bestScen ? bestScen.score : 0;
	const wssWorst = worstScen ? worstScen.score : 0;

	let stopCardHtml = '';
	if (trace.l5.stopCard) {
		const sc = trace.l5.stopCard;
		const isStop = sc.recommended === 'STOP';
		const badgeCls = isStop ? 'status-fail' : 'status-pass';
		const pGoodPct = (sc.pGood * 100).toFixed(0);

		stopCardHtml = `
			<div class="salvage-plan-card" style="margin-top: 0.75rem; background: rgba(0,0,0,0.25); border-left-color: ${isStop ? '#ef4444' : '#10b981'};">
				<div class="salvage-plan-title" style="display: flex; justify-content: space-between; align-items: center;">
					<span>Enhancement Controller (Stop Point +${sc.enhanceAtPoint})</span>
					<span class="trace-step-status ${badgeCls}">${sc.recommended}</span>
				</div>
				<div class="salvage-plan-body" style="margin-top: 0.375rem;">
					${isStop ? `<p style="color: #ef4444; font-weight: 700; margin-bottom: 0.25rem;">Stop Reason: ${sc.reason}</p>` : ''}
					<p style="margin-bottom: 0.25rem;">Continuation Odds (P_good): <strong>${pGoodPct}%</strong> | Expected Value (EV_final): <strong>${sc.evFinal.toFixed(1)}</strong></p>
					<p style="font-size: 0.6875rem; color: var(--text-muted); font-style: italic;">${sc.rollSequenceForCore}</p>
				</div>
			</div>
		`;
	}

	const l5Body = `
		<p>Current Gear Substat Score: <strong>${wssCurrent.toFixed(1)} WSS</strong></p>
		
		<div class="projection-envelope-card">
			<div class="projection-title">+15 Enhancement Projections</div>
			<div class="projection-row">
				<span class="projection-label">Expected Scenario:</span>
				<span class="projection-val projection-expected">${wssExpected.toFixed(1)} WSS</span>
			</div>
			<div class="projection-row">
				<span class="projection-label">Best-Case Envelope:</span>
				<span class="projection-val projection-best">${wssBest.toFixed(1)} WSS</span>
			</div>
			<div class="projection-row">
				<span class="projection-label">Worst-Case Envelope:</span>
				<span class="projection-val projection-worst">${wssWorst.toFixed(1)} WSS</span>
			</div>
			${currentGear.level === 85 ? `<p style="font-size: 0.6875rem; color: var(--text-muted); margin-top: 0.5rem; font-style: italic;">*Includes reforge calculation to Level 90 increments based on roll counts.</p>` : ''}
		</div>

		${stopCardHtml}
	`;

	// L6 Decision & Salvage
	let l6VerdictClass = 'status-fail';
	let l6VerdictLabel = trace.l6.verdict;
	switch (trace.l6.verdict) {
	case "KEEP_ENHANCE":
		l6VerdictClass = "status-pass";
		l6VerdictLabel = "Keep & Enhance";
		break;
	case "KEEP_MARGINAL":
		l6VerdictClass = "status-tag";
		l6VerdictLabel = "Keep (Marginal)";
		break;
	case "SALVAGE_MOD":
		l6VerdictClass = "status-pass";
		l6VerdictLabel = "Salvage Candidate (Mod)";
		break;
	case "REFORGE_TAG":
		l6VerdictClass = "status-pass";
		l6VerdictLabel = "Reforge Candidate";
		break;
	case "SPEED_VAULT":
		l6VerdictClass = "status-tag";
		l6VerdictLabel = "Speed Vaulted";
		break;
	case "SELL_EXTRACT":
		l6VerdictClass = "status-neutral";
		l6VerdictLabel = "Sell / Extract";
		break;
	}

	let salvageHtml = '';
	if (trace.l6.verdict === 'SALVAGE_MOD' && trace.l6.salvageDetail) {
		salvageHtml = `
			<div class="salvage-plan-card">
				<div class="salvage-plan-title">
					<svg style="width: 14px; height: 14px;" fill="none" stroke="currentColor" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
					</svg>
					Recommended Modification Plan
				</div>
				<div class="salvage-plan-body">
					Replace <span class="salvage-stat-highlight">${formatStatLabel(trace.l6.salvageDetail.deadSubStat)}</span> 
					with <span class="salvage-stat-highlight">${formatStatLabel(trace.l6.salvageDetail.targetStat)}</span> 
					after +15 enhancement.
					<br>
					Expected value at minimum rolls: <strong>${trace.l6.salvageDetail.expectedValue.toFixed(1)}</strong>.
					Rescored Fit: <strong>${trace.l6.salvageDetail.rescoredFit.toFixed(1)}</strong>.
				</div>
			</div>
		`;
	}

	const l6Body = `
		<p>Pipeline Verdict: <strong style="color: #fff;">${l6VerdictLabel}</strong></p>
		${trace.l6.winnerHero ? `<p style="margin-top: 0.125rem;">Best fit hero build: <strong>${trace.l6.winnerHero} (Build #${trace.l6.winnerBuild})</strong></p>` : ''}
		${trace.l6.runnerUps && trace.l6.runnerUps.length > 0 ? `<p style="margin-top: 0.125rem; font-size: 0.6875rem;">Secondary targets: ${trace.l6.runnerUps.join(', ')}</p>` : ''}
		${salvageHtml}
	`;

	listEl.innerHTML = `
		<div class="trace-timeline animate-fade-in" style="padding: 1rem;">
			<!-- Layer 0 -->
			<div class="trace-step">
				<div class="trace-step-header" data-toggle="l0">
					<div class="trace-step-title-group">
						<span class="trace-step-number">00</span>
						<span class="trace-step-name">Ingest & Normalize</span>
					</div>
					<span class="trace-step-status status-pass">Normalized</span>
				</div>
				<div class="trace-step-body" id="body-l0">
					<p>Reconstructed substat roll counts using constraint solver:</p>
					<div class="trace-detail-grid" style="margin-top:0.5rem; margin-bottom:0.5rem;">
						${l0Rolls}
					</div>
					${trace.l0.ambiguities.length > 0 ? `<p style="color: #f59e0b;">Ambiguity note: ${trace.l0.ambiguities.join(', ')}</p>` : ''}
				</div>
			</div>

			<!-- Layer 1 -->
			<div class="trace-step">
				<div class="trace-step-header" data-toggle="l1">
					<div class="trace-step-title-group">
						<span class="trace-step-number">01</span>
						<span class="trace-step-name">Game Rules (Legality Check)</span>
					</div>
					<span class="trace-step-status ${l1Status}">${l1StatusLabel}</span>
				</div>
				<div class="trace-step-body" id="body-l1">
					${l1Body}
				</div>
			</div>

			<!-- Layer 2 -->
			<div class="trace-step">
				<div class="trace-step-header" data-toggle="l2">
					<div class="trace-step-title-group">
						<span class="trace-step-number">02</span>
						<span class="trace-step-name">Universal Discard Filters</span>
					</div>
					<span class="trace-step-status ${l2Status}">${l2StatusLabel}</span>
				</div>
				<div class="trace-step-body" id="body-l2">
					${l2Body}
				</div>
			</div>

			<!-- Layer 3 -->
			<div class="trace-step">
				<div class="trace-step-header" data-toggle="l3">
					<div class="trace-step-title-group">
						<span class="trace-step-number">03</span>
						<span class="trace-step-name">Speed Check Bypass Toggle</span>
					</div>
					<span class="trace-step-status ${l3Status}">${l3StatusLabel}</span>
				</div>
				<div class="trace-step-body" id="body-l3">
					${l3Body}
				</div>
			</div>

			<!-- Layer 4 -->
			<div class="trace-step">
				<div class="trace-step-header" data-toggle="l4">
					<div class="trace-step-title-group">
						<span class="trace-step-number">04</span>
						<span class="trace-step-name">Hero Requirement Matching</span>
					</div>
					<span class="trace-step-status ${l4Status}">${l4StatusLabel}</span>
				</div>
				<div class="trace-step-body" id="body-l4">
					<p>Evaluates set combinations, main stat alignment, capability minimums, and fit score gates.</p>
					${l4MatchesHtml}
				</div>
			</div>

			<!-- Layer 5 -->
			<div class="trace-step">
				<div class="trace-step-header" data-toggle="l5">
					<div class="trace-step-title-group">
						<span class="trace-step-number">05</span>
						<span class="trace-step-name">Quality & +15 Projections</span>
					</div>
					<span class="trace-step-status status-neutral">Projected</span>
				</div>
				<div class="trace-step-body" id="body-l5">
					${l5Body}
				</div>
			</div>

			<!-- Layer 6 -->
			<div class="trace-step">
				<div class="trace-step-header" data-toggle="l6">
					<div class="trace-step-title-group">
						<span class="trace-step-number">06</span>
						<span class="trace-step-name">Decision & Salvage Recommendation</span>
					</div>
					<span class="trace-step-status ${l6VerdictClass}">${l6VerdictLabel}</span>
				</div>
				<div class="trace-step-body" id="body-l6">
					${l6Body}
				</div>
			</div>
		</div>
	`;

	// Setup Accordion toggles
	const headers = listEl.querySelectorAll('.trace-step-header');
	headers.forEach(h => {
		h.addEventListener('click', () => {
			const id = h.getAttribute('data-toggle');
			const body = listEl.querySelector(`#body-${id}`);
			if (body) {
				if (body.style.display === 'none') {
					body.style.display = 'block';
				} else {
					body.style.display = 'none';
				}
			}
		});
	});
}

/**
 * Maps database keys to human readable labels
 */
const STAT_TO_SAV_KEY = {
	"Attack":                  "atk",
	"AttackPercent":           "atk",
	"Health":                  "hp",
	"HealthPercent":           "hp",
	"Defense":                 "def",
	"DefensePercent":          "def",
	"Speed":                   "spd",
	"CritHitChancePercent":    "cc",
	"CritHitDamagePercent":    "cd",
	"EffectivenessPercent":    "eff",
	"EffectResistancePercent": "res"
};

function formatStatLabel(statType) {
	const labels = {
		"Attack":                  "Atk",
		"AttackPercent":           "Atk %",
		"Health":                  "HP",
		"HealthPercent":           "HP %",
		"Defense":                 "Def",
		"DefensePercent":          "Def %",
		"Speed":                   "Speed",
		"CritHitChancePercent":    "Crit Chance",
		"CritHitDamagePercent":    "Crit Damage",
		"EffectivenessPercent":    "Effectiveness",
		"EffectResistancePercent": "Resistance"
	};
	return labels[statType] || statType;
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
