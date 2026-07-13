/**
 * Scraping layer.
 *
 * Owns all navigation and DOM interaction against the hero-library page:
 *   - selecting a hero via the select2 control (no full page reload, so the
 *     big herodata.json/artifactdata.json are fetched only once),
 *   - triggering a search and waiting for the build list to render,
 *   - clicking each build archetype row so the page recomputes that build's
 *     average stats (we only ever read values the UI itself displays),
 *   - returning a fully-assembled HeroRecord.
 *
 * Parsing of raw DOM strings -> typed data lives in `parser.ts`; set-name
 * normalization lives in `sets.ts`. This module only orchestrates.
 */
import type { Page } from "playwright";
import type { BuildArchetype, HeroOption, HeroRecord, ScraperConfig } from "./types.ts";
import { logger } from "./logger.ts";
import {
  buildSets,
  parsePercent,
  parseIntStat,
  readAverageStats,
  readBuildRows,
  readTopArtifact,
} from "./parser.ts";

/** Fragment of the getBuilds API URL, used to detect the per-hero response. */
const GETBUILDS_URL_FRAGMENT = "/dev/getBuilds";

/** Selector for the ag-grid "no data" overlay shown when a hero has 0 builds. */
const NO_DATA_SELECTOR = ".ag-overlay-no-rows-center";

/** HTTP statuses the API returns when throttling us (soft ban). */
const RATE_LIMIT_STATUSES = new Set([403, 429, 503]);

/** Outcome of a single hero scrape. */
export type ScrapeStatus = "ok" | "nodata" | "ratelimited" | "error";

export interface ScrapeResult {
  record: HeroRecord;
  /** How many build rows were actually present (<= topN). */
  buildCount: number;
  /** Outcome classification — drives how the orchestrator records it. */
  status: ScrapeStatus;
}

const sleep = (ms: number) => new Promise<void>((r) => setTimeout(r, ms));

/**
 * Scrape a single hero. Assumes the shared page is already on the hero library
 * and the hero <select> is populated.
 */
export async function scrapeHero(
  page: Page,
  hero: HeroOption,
  config: ScraperConfig,
): Promise<ScrapeResult> {
  const empty: HeroRecord = { hero: hero.name, builds: [] };

  // 1. Select the hero in the select2 control (programmatic, no dropdown UI).
  await selectHero(page, hero.id);

  // 2. Trigger search with retry/backoff for rate-limiting. The getBuilds API
  //    (AWS API Gateway) soft-bans IPs that burst too many requests, returning
  //    403. We retry with exponential backoff; only a genuine 200 + empty data
  //    counts as "no data" — rate-limits are never recorded as completed.
  let resp;
  for (let attempt = 0; attempt <= config.maxRetries; attempt++) {
    // Register the response listener *before* clicking Search so we never race.
    const respPromise = page.waitForResponse(
      (r) => r.url().includes(GETBUILDS_URL_FRAGMENT),
      { timeout: config.heroTimeoutMs },
    );
    await page.evaluate(() => {
      const btn = document.getElementById("buildsSearchButton") as HTMLButtonElement | null;
      btn?.click();
    });

    try {
      resp = await respPromise;
    } catch (err) {
      logger.warn(`hero "${hero.name}": getBuilds response timeout (attempt ${attempt + 1})`, {
        error: String(err),
      });
      if (attempt < config.maxRetries) {
        const timeoutWait = 5_000;
        logger.info(`retrying after timeout in ${timeoutWait}ms...`);
        await sleep(timeoutWait);
        continue;
      }
      return { record: empty, buildCount: 0, status: "error" };
    }

    if (resp.ok()) break; // got a usable response

    const status = resp.status();
    if (RATE_LIMIT_STATUSES.has(status) && attempt < config.maxRetries) {
      const wait = backoffFor(config, attempt);
      logger.warn(
        `hero "${hero.name}": getBuilds HTTP ${status} (rate-limited), ` +
          `retry ${attempt + 1}/${config.maxRetries} in ${wait}ms`,
      );
      await sleep(wait);
      continue;
    }
    if (RATE_LIMIT_STATUSES.has(status)) {
      logger.warn(`hero "${hero.name}": getBuilds HTTP ${status} — retries exhausted`);
      return { record: empty, buildCount: 0, status: "ratelimited" };
    }
    logger.warn(`hero "${hero.name}": getBuilds HTTP ${status} (unrecoverable)`);
    return { record: empty, buildCount: 0, status: "error" };
  }

  // 3. Wait for the build list to render OR the no-data overlay to appear.
  const settled = await waitForBuildsSettled(page, config.heroTimeoutMs);
  if (settled === "error" || settled === "loading") {
    logger.warn(`hero "${hero.name}": builds never settled (${settled})`);
    return { record: empty, buildCount: 0, status: "error" };
  }

  if (settled === "nodata") {
    logger.debug(`hero "${hero.name}": no build data`);
    return { record: empty, buildCount: 0, status: "nodata" };
  }

  // 5. Read the top-N build rows (usage % + set images) and the top artifact.
  const rawRows = await readBuildRows(page, config.topN);
  const rawArtifact = await readTopArtifact(page);
  if (rawRows.length === 0) {
    logger.warn(`hero "${hero.name}": builds settled but no rows found`);
    return { record: empty, buildCount: 0, status: "nodata" };
  }

  // 6. Click each build row so the page recomputes that combo's average stats,
  //    then read the Average panel.
  //
  //    Important quirk of the source page: its `objectsAreEqual` only compares
  //    key *counts* (not values), so the click handler treats "select filter X"
  //    as a toggle — if the currently-active filter has the same number of set
  //    keys as the one being clicked, it toggles the filter OFF (showing *all*
  //    builds) instead of switching to the new combo. To always end up with the
  //    target combo selected, we first deactivate whichever row is currently
  //    active (clicking it toggles it off → filter = null), then click the
  //    target row. This is exactly the interaction a careful user would perform
  //    and yields the genuine per-archetype average the UI can display.
  const builds: BuildArchetype[] = [];
  for (let i = 0; i < rawRows.length; i++) {
    const row = rawRows[i];
    const rowSel = `#setComboRow${i}`;

    // 6a. Deactivate the currently-active row (if any) so the filter is null.
    await page.evaluate(() => {
      const active = document.querySelector<HTMLElement>("#setCombos .setComboRow.active");
      active?.dispatchEvent(
        new MouseEvent("click", { bubbles: true, cancelable: true, view: window }),
      );
    });
    await page.waitForTimeout(Math.min(80, config.buildDelayMs));

    // 6b. Click the target row -> sets the filter to this combo.
    await page.evaluate((sel) => {
      document.querySelector(sel)?.dispatchEvent(
        new MouseEvent("click", { bubbles: true, cancelable: true, view: window }),
      );
    }, rowSel);

    // The click handler + ag-grid filter + average recompute are synchronous,
    // but give the page a beat to settle and to pace ourselves.
    await page.waitForTimeout(config.buildDelayMs);

    // 6c. Sanity check: the target row should now be the active one. If the
    //     toggle quirk struck anyway (e.g. combo had a different key count),
    //     we still read whatever the UI displays — never invent values.
    const stats = await readAverageStats(page);

    builds.push({
      rank: row.rank,
      usage: parsePercent(row.usageText),
      sets: buildSets(row),
      averageStats: stats,
      artifact: {
        name: rawArtifact?.name ?? "",
        usage: parsePercent(rawArtifact?.usageText),
      },
    });
  }

  return { record: { hero: hero.name, builds }, buildCount: builds.length, status: "ok" };
}

/** Pick the backoff delay for a given (0-based) attempt, clamped to the table. */
function backoffFor(config: ScraperConfig, attempt: number): number {
  const table = config.rateLimitBackoffMs;
  return table[Math.min(attempt, table.length - 1)] ?? 10_000;
}

/* ----------------------------- internals -------------------------------- */

/** Programmatically select a hero in the select2 control. */
async function selectHero(page: Page, heroId: string): Promise<void> {
  await page.evaluate((id) => {
    const $ = (window as any).jQuery;
    if (!$) return;
    $("#heroSelector0").val(id).trigger("change");
  }, heroId);
}

type SettleState = "loading" | "ready" | "nodata" | "error";

/**
 * Poll until the build list either renders rows ("ready") or the page shows
 * its "No data" overlay ("nodata"). Bounded by `timeoutMs`.
 */
async function waitForBuildsSettled(page: Page, timeoutMs: number): Promise<SettleState> {
  const deadline = Date.now() + Math.min(timeoutMs, 15_000);
  while (Date.now() < deadline) {
    const state = await page.evaluate((noDataSel) => {
      const rows = document.querySelectorAll("#setCombos .setComboRow").length;
      if (rows > 0) return "ready";
      const overlay = document.querySelector(noDataSel);
      const txt = (overlay?.textContent ?? "").trim();
      if (txt.includes("No data")) return "nodata";
      // The initial "Select a hero" overlay also counts as not-yet-loaded.
      return "loading";
    }, NO_DATA_SELECTOR);
    if (state === "ready" || state === "nodata") return state as SettleState;
    await page.waitForTimeout(150);
  }
  return "error";
}
