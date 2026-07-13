/**
 * Entry point / orchestrator.
 *
 * Wires together the browser, hero-list discovery, per-hero scraping, output
 * writing and resumable progress. Designed to be run directly with bun:
 *
 *   bun run index.ts                      # full run (fresh)
 *   bun run index.ts --resume             # resume an interrupted run
 *   bun run index.ts --limit 20 --delay 2000
 *   bun run index.ts --hero "Afternoon Soak Flan"   # single-hero test
 */
import { parseArgs } from "./config.ts";
import { getBrowser, closeBrowser } from "./browser.ts";
import { getHeroList } from "./parser.ts";
import { scrapeHero } from "./scraper.ts";
import {
  appendHero,
  ensureOutDir,
  loadExistingRecords,
  rebuildFinalJson,
  writeFinalJson,
  type OutputPaths,
} from "./output.ts";
import {
  emptyProgress,
  loadProgress,
  logProgress,
  markCompleted,
  markFailed,
  saveProgress,
  skipSet,
} from "./state.ts";
import { logger, setLogLevel } from "./logger.ts";
import type { HeroOption, ProgressState } from "./types.ts";

/** Rebuild the pretty JSON file every N heroes so partial results are visible. */
const FLUSH_EVERY = 10;

async function main(): Promise<void> {
  const config = parseArgs(process.argv.slice(2));
  if (process.env.DEBUG) setLogLevel("debug");

  const paths = await ensureOutDir(config.outDir);
  logger.info("e7 hero-library scraper starting", {
    topN: config.topN,
    delay: config.heroDelayMs,
    limit: config.limit || "unlimited",
    resume: config.resume,
    hero: config.hero ?? "-",
    out: config.outDir,
  });

  const handle = await getBrowser(config.baseUrl, {
    headless: config.headless,
    extraHeaders: config.extraHeaders,
  });
  const page = handle.page;

  // Install graceful shutdown so progress is never lost.
  let shuttingDown = false;
  const shutdown = async (sig: string) => {
    if (shuttingDown) return;
    shuttingDown = true;
    logger.warn(`received ${sig}, shutting down gracefully...`);
    try {
      await rebuildFinalJson(paths);
    } catch {
      /* ignore */
    }
    await closeBrowser();
    process.exit(0);
  };
  process.on("SIGINT", () => void shutdown("SIGINT"));
  process.on("SIGTERM", () => void shutdown("SIGTERM"));

  // Discover the full hero list once.
  const allHeroes = await getHeroList(page);
  logger.info(`discovered ${allHeroes.length} heroes`);

  // Single-hero test mode: scrape one, print, write a tiny file, exit.
  if (config.hero) {
    const hero = findHero(allHeroes, config.hero);
    if (!hero) {
      logger.error(`hero not found: "${config.hero}"`);
      await closeBrowser();
      process.exit(1);
    }
    logger.info(`single-hero mode: "${hero.name}" (id=${hero.id})`);
    const { record } = await scrapeHero(page, hero, config);
    console.log(JSON.stringify(record, null, 2));
    await writeFinalJson(paths, [record]);
    await closeBrowser();
    return;
  }

  // Batch mode: figure out where to start.
  let progress: ProgressState;
  let completedNames = new Set<string>();
  if (config.resume) {
    const persisted = await loadProgress(config.outDir);
    progress = persisted ?? emptyProgress(allHeroes.length);
    progress.total = allHeroes.length;
    // Also dedupe against anything already in the JSONL store.
    const existing = await loadExistingRecords(paths);
    completedNames = new Set(existing.map((r) => r.hero));
    logger.info(`resuming: ${progress.completed.length} already done, ` +
      `${completedNames.size} in jsonl`);
  } else {
    progress = emptyProgress(allHeroes.length);
    // Fresh run: clear the JSONL store so we don't duplicate.
    const { promises: fs } = await import("node:fs");
    await fs.writeFile(paths.jsonlPath, "", "utf8");
  }

  const skip = skipSet(progress);

  let processed = 0;
  let consecutiveRateLimited = 0;
  /** Abort the run if this many heroes are rate-limited in a row. */
  const RATE_LIMIT_ABORT_THRESHOLD = 5;

  for (let i = 0; i < allHeroes.length; i++) {
    if (shuttingDown) break;
    const hero = allHeroes[i];
    progress.nextIndex = i;

    // Skip already-completed heroes (by id from progress, or by name from jsonl).
    if (skip.has(hero.id) || completedNames.has(hero.name)) {
      continue;
    }

    if (config.limit > 0 && processed >= config.limit) {
      logger.info(`reached --limit ${config.limit}, stopping`);
      break;
    }

    const t0 = Date.now();
    try {
      const { record, buildCount, status } = await scrapeHero(page, hero, config);
      const elapsed = ((Date.now() - t0) / 1000).toFixed(1);

      if (status === "ratelimited") {
        // Soft ban in effect. Do NOT mark completed/failed so a later
        // --resume retries this hero. Since the ban remained after our 5.5-minute backoff, abort immediately.
        logger.warn(
          `✗ ${hero.name} — rate-limited (${elapsed}s); ban remains after backoff. ` +
            `Aborting run — re-run with --resume later.`,
        );
        break;
      } else if (status === "error") {
        // Non-recoverable error: mark failed so we don't loop on a broken hero.
        consecutiveRateLimited = 0;
        markFailed(progress, hero.id);
        logger.warn(`✗ ${hero.name} — error (${elapsed}s); marked failed`);
      } else {
        // "ok" or "nodata": persist + mark completed.
        consecutiveRateLimited = 0;
        await appendHero(paths, record);
        completedNames.add(hero.name);
        markCompleted(progress, hero.id);
        const tag = status === "nodata" ? " (no data)" : "";
        logger.info(`✓ ${hero.name} — ${buildCount} builds${tag} (${elapsed}s)`);
      }
    } catch (err) {
      consecutiveRateLimited = 0;
      markFailed(progress, hero.id);
      logger.error(`✗ ${hero.name} failed: ${String(err)}`);
    }

    processed++;
    await saveProgress(config.outDir, progress);

    if (processed % FLUSH_EVERY === 0) {
      await rebuildFinalJson(paths);
      logProgress(progress);
    }

    // Conservative pacing between heroes to avoid rate limiting / soft bans.
    if (i < allHeroes.length - 1) {
      await sleep(config.heroDelayMs);
    }
  }

  // Final flush.
  progress.nextIndex = allHeroes.length;
  await saveProgress(config.outDir, progress);
  const finalRecords = await rebuildFinalJson(paths);
  logProgress(progress);
  logger.info(`done. ${finalRecords.length} heroes written to ${paths.jsonPath}`);

  await closeBrowser();
}

/** Find a hero by exact name, falling back to case-insensitive containment. */
function findHero(heroes: HeroOption[], query: string): HeroOption | undefined {
  return (
    heroes.find((h) => h.name === query) ??
    heroes.find((h) => h.name.toLowerCase() === query.toLowerCase()) ??
    heroes.find((h) => h.name.toLowerCase().includes(query.toLowerCase()))
  );
}

function sleep(ms: number): Promise<void> {
  return new Promise((r) => setTimeout(r, ms));
}

main().catch((err) => {
  logger.error("fatal", err);
  closeBrowser().finally(() => process.exit(1));
});
