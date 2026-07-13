/**
 * Default runtime configuration + CLI argument parsing for the scraper.
 */
import type { ScraperConfig } from "./types.ts";

export const DEFAULT_CONFIG: ScraperConfig = {
  baseUrl: "https://fribbels.github.io/e7/hero-library.html",
  topN: 5,
  heroDelayMs: 2500,
  buildDelayMs: 250,
  heroTimeoutMs: 45000,
  headless: true,
  limit: 0,
  resume: false,
  hero: undefined,
  outDir: "./out",
  // Rate-limit resilience:
  maxRetries: 1,
  rateLimitBackoffMs: [330_000],
  rateLimitCooldownMs: 30_000,
};

/** Parse argv into a ScraperConfig, layered on top of the defaults. */
export function parseArgs(argv: string[]): ScraperConfig {
  const cfg: ScraperConfig = { ...DEFAULT_CONFIG };

  for (let i = 0; i < argv.length; i++) {
    const a = argv[i];
    const next = () => argv[++i];
    switch (a) {
      case "--url":
        cfg.baseUrl = next();
        break;
      case "--top":
        cfg.topN = Number(next());
        break;
      case "--delay":
        cfg.heroDelayMs = Number(next());
        break;
      case "--build-delay":
        cfg.buildDelayMs = Number(next());
        break;
      case "--timeout":
        cfg.heroTimeoutMs = Number(next());
        break;
      case "--limit":
        cfg.limit = Number(next());
        break;
      case "--hero":
        cfg.hero = next();
        break;
      case "--resume":
        cfg.resume = true;
        break;
      case "--headed":
        cfg.headless = false;
        break;
      case "--out":
        cfg.outDir = next();
        break;
      case "--help":
      case "-h":
        printHelp();
        process.exit(0);
        break;
      default:
        if (a.startsWith("--")) {
          console.warn(`[config] unknown flag: ${a} (ignored)`);
        }
    }
  }

  return cfg;
}

function printHelp(): void {
  console.log(`
e7-hero-library-scraper

Scrapes the top build archetypes for every hero from the Fribbels E7 Hero Library.

Usage:
  bun run index.ts [options]

Options:
  --url <url>          Base hero-library URL (default: ${DEFAULT_CONFIG.baseUrl})
  --top <n>            Top N build archetypes per hero (default: ${DEFAULT_CONFIG.topN})
  --delay <ms>         Delay between heroes in ms (default: ${DEFAULT_CONFIG.heroDelayMs})
  --build-delay <ms>   Delay between build-row clicks in ms (default: ${DEFAULT_CONFIG.buildDelayMs})
  --timeout <ms>       Per-hero timeout in ms (default: ${DEFAULT_CONFIG.heroTimeoutMs})
  --limit <n>          Stop after <n> heroes (0 = unlimited, default: 0)
  --hero <name>        Scrape a single hero by name (for testing)
  --resume             Resume from persisted progress (skip completed heroes)
  --headed             Show the browser window (default: headless)
  --out <dir>          Output/state directory (default: ${DEFAULT_CONFIG.outDir})
  -h, --help           Show this help

Examples:
  bun run index.ts --hero "Afternoon Soak Flan"
  bun run index.ts --limit 20 --delay 2000
  bun run index.ts --resume
`);
}
