/**
 * Output layer.
 *
 * Writes scraped heroes as:
 *   - `heroes.jsonl` — one compact JSON object per line, appended after each
 *     hero completes. This is the durable, append-only store that makes
 *     resuming trivial and crash-safe.
 *   - `heroes.json` — a pretty-printed JSON array, rewritten at the end of
 *     each run (and on exit) for human/programmatic consumption.
 *
 * The JSONL file is the source of truth; the JSON file is a convenience view.
 */
import { promises as fs } from "node:fs";
import path from "node:path";
import type { HeroRecord } from "./types.ts";
import { logger } from "./logger.ts";

export interface OutputPaths {
  outDir: string;
  jsonlPath: string;
  jsonPath: string;
}

export function outputPaths(outDir: string): OutputPaths {
  return {
    outDir,
    jsonlPath: path.join(outDir, "heroes.jsonl"),
    jsonPath: path.join(outDir, "heroes.json"),
  };
}

/** Ensure the output directory exists. */
export async function ensureOutDir(outDir: string): Promise<OutputPaths> {
  await fs.mkdir(outDir, { recursive: true });
  return outputPaths(outDir);
}

/** Append a single hero record to the JSONL store (crash-safe, append-only). */
export async function appendHero(paths: OutputPaths, record: HeroRecord): Promise<void> {
  const line = JSON.stringify(record);
  await fs.appendFile(paths.jsonlPath, line + "\n", "utf8");
}

/**
 * Load every hero record previously written to the JSONL store.
 * Malformed lines are skipped with a warning (never fatal).
 */
export async function loadExistingRecords(paths: OutputPaths): Promise<HeroRecord[]> {
  let raw: string;
  try {
    raw = await fs.readFile(paths.jsonlPath, "utf8");
  } catch {
    return [];
  }
  const records: HeroRecord[] = [];
  const lines = raw.split(/\r?\n/);
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    if (!line.trim()) continue;
    try {
      records.push(JSON.parse(line) as HeroRecord);
    } catch {
      logger.warn(`skipping malformed JSONL line ${i + 1}`);
    }
  }
  return records;
}

/** Rewrite the pretty JSON array from the given records. */
export async function writeFinalJson(paths: OutputPaths, records: HeroRecord[]): Promise<void> {
  const payload = JSON.stringify(records, null, 2);
  await fs.writeFile(paths.jsonPath, payload + "\n", "utf8");
  logger.info("wrote final json", { path: paths.jsonPath, count: records.length });
}

/**
 * Rebuild `heroes.json` from the current JSONL store. Handy after a resumed
 * run or as a standalone "finalize" step.
 */
export async function rebuildFinalJson(paths: OutputPaths): Promise<HeroRecord[]> {
  const records = await loadExistingRecords(paths);
  await writeFinalJson(paths, records);
  return records;
}
