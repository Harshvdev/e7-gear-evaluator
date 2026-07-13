/**
 * Core domain types for the Epic Seven Hero Library scraper.
 *
 * These shapes mirror the example output requested in the task spec so the
 * emitted JSON is directly usable without further transformation.
 */

/** A single averaged stat value (always a number when present). */
export interface AverageStats {
  atk?: number;
  def?: number;
  hp?: number;
  spd?: number;
  /** Critical Hit Chance */
  cc?: number;
  /** Critical Hit Damage */
  cd?: number;
  /** Effectiveness */
  eff?: number;
  /** Effect Resistance */
  res?: number;
  /** Dual Attack Chance (only when displayed in the Average panel). */
  dac?: number;
  /** Gear Score (displayed in the Average panel). */
  gs?: number;
  /**
   * Catch-all for any other stat the Average panel may render in the future.
   * Keys are lower-cased human labels.
   */
  [extra: string]: number | undefined;
}

/** Recommended artifact for a hero/build with its overall usage share. */
export interface ArtifactInfo {
  name: string;
  usage: number;
}

/** A single build archetype (one row of the "set combos" list). */
export interface BuildArchetype {
  /** 1-based rank within the hero's build list (1 = most used). */
  rank: number;
  /** Usage percentage of this set combination (e.g. 32.6). */
  usage: number;
  /** Ordered list of normalized, human-readable set names (primary first). */
  sets: string[];
  /** Average stats for the builds that match this set combination. */
  averageStats: AverageStats;
  /** Hero-level recommended artifact (top of the artifact panel). */
  artifact: ArtifactInfo;
}

/** A fully scraped hero record. */
export interface HeroRecord {
  hero: string;
  builds: BuildArchetype[];
}

/** A hero selectable from the dropdown. */
export interface HeroOption {
  id: string;
  name: string;
}

/** Serializable progress state used for resuming an interrupted run. */
export interface ProgressState {
  /** ISO timestamp of the last update. */
  updatedAt: string;
  /** Total number of heroes discovered. */
  total: number;
  /** Index (into the discovered hero list) of the next hero to process. */
  nextIndex: number;
  /** Hero ids that have been completed successfully. */
  completed: string[];
  /** Hero ids that failed and should not be retried this run. */
  failed: string[];
}

/** Runtime configuration for a scrape run. */
export interface ScraperConfig {
  /** Base URL of the hero library. */
  baseUrl: string;
  /** Max number of build archetypes to extract per hero. */
  topN: number;
  /** Delay (ms) between finishing one hero and starting the next. */
  heroDelayMs: number;
  /** Delay (ms) between clicking consecutive build rows (lets averages settle). */
  buildDelayMs: number;
  /** Per-hero navigation/parse timeout (ms). */
  heroTimeoutMs: number;
  /** Whether to run the browser headless. */
  headless: boolean;
  /** Max heroes to process this run (0 = unlimited). */
  limit: number;
  /** Resume from persisted progress instead of starting fresh. */
  resume: boolean;
  /** Optional single hero name to scrape (for testing). */
  hero: string | undefined;
  /** Directory for state + output artifacts. */
  outDir: string;
  /** Max retries when the getBuilds API returns a rate-limit (403/429/503). */
  maxRetries: number;
  /** Backoff schedule (ms) between rate-limit retries; indexed by attempt. */
  rateLimitBackoffMs: number[];
  /** Cooldown (ms) applied after a hero exhausts rate-limit retries. */
  rateLimitCooldownMs: number;
  /** Extra HTTP headers to send on the initial page load. */
  extraHeaders?: Record<string, string>;
}
