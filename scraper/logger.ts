/**
 * Tiny leveled logger with timestamps + color-free, grep-friendly prefixes.
 * Kept dependency-free so the scraper stays self-contained.
 */

type Level = "debug" | "info" | "warn" | "error";

const LEVEL_PREFIX: Record<Level, string> = {
  debug: "DBG",
  info: "INF",
  warn: "WRN",
  error: "ERR",
};

let currentLevel: Level = "info";
const ORDER: Record<Level, number> = { debug: 0, info: 1, warn: 2, error: 3 };

export function setLogLevel(level: Level): void {
  currentLevel = level;
}

function ts(): string {
  return new Date().toISOString().replace("T", " ").replace("Z", "");
}

function log(level: Level, msg: string, meta?: unknown): void {
  if (ORDER[level] < ORDER[currentLevel]) return;
  const line = `[${ts()}] ${LEVEL_PREFIX[level]} ${msg}`;
  if (meta === undefined) {
    console[level === "error" ? "error" : "log"](line);
  } else {
    console[level === "error" ? "error" : "log"](line, meta);
  }
}

export const logger = {
  debug: (m: string, meta?: unknown) => log("debug", m, meta),
  info: (m: string, meta?: unknown) => log("info", m, meta),
  warn: (m: string, meta?: unknown) => log("warn", m, meta),
  error: (m: string, meta?: unknown) => log("error", m, meta),
};
