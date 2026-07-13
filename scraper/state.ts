/**
 * Resume / progress state.
 *
 * The scraper is fully resumable: after each hero we persist a small
 * `progress.json` describing what's done. On restart with `--resume`, heroes
 * already in `heroes.jsonl` are skipped, so an interrupted long run picks up
 * exactly where it left off without duplicate work or duplicate API requests.
 */
import { promises as fs } from "node:fs";
import path from "node:path";
import type { ProgressState } from "./types.ts";
import { logger } from "./logger.ts";

export function progressPath(outDir: string): string {
  return path.join(outDir, "progress.json");
}

export function emptyProgress(total = 0): ProgressState {
  return {
    updatedAt: new Date().toISOString(),
    total,
    nextIndex: 0,
    completed: [],
    failed: [],
  };
}

/** Load persisted progress, or return undefined if none exists. */
export async function loadProgress(outDir: string): Promise<ProgressState | undefined> {
  try {
    const raw = await fs.readFile(progressPath(outDir), "utf8");
    return JSON.parse(raw) as ProgressState;
  } catch {
    return undefined;
  }
}

/** Persist progress atomically (write temp then rename). */
export async function saveProgress(outDir: string, state: ProgressState): Promise<void> {
  state.updatedAt = new Date().toISOString();
  const file = progressPath(outDir);
  const tmp = file + ".tmp";
  await fs.writeFile(tmp, JSON.stringify(state, null, 2), "utf8");
  await fs.rename(tmp, file);
}

/** Mark a hero id as completed and advance nextIndex if appropriate. */
export function markCompleted(state: ProgressState, id: string): ProgressState {
  if (!state.completed.includes(id)) state.completed.push(id);
  return state;
}

/** Mark a hero id as failed (won't be retried this run). */
export function markFailed(state: ProgressState, id: string): ProgressState {
  if (!state.failed.includes(id)) state.failed.push(id);
  return state;
}

/** Compute the set of ids to skip (completed or failed). */
export function skipSet(state: ProgressState | undefined): Set<string> {
  if (!state) return new Set();
  return new Set([...state.completed, ...state.failed]);
}

/** Log a one-line progress summary. */
export function logProgress(state: ProgressState): void {
  const done = state.completed.length;
  const failed = state.failed.length;
  const pct = state.total ? ((done + failed) / state.total) * 100 : 0;
  logger.info(
    `progress ${done}/${state.total} done${failed ? `, ${failed} failed` : ""} (${pct.toFixed(1)}%)`,
  );
}
