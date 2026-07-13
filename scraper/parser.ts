/**
 * Parsing layer.
 *
 * Reads structured values out of the rendered hero-library DOM and turns raw
 * strings ("32.6%", "4,440", "./assets/setdestruction.png") into typed data.
 *
 * Nothing in this module navigates or clicks — that lives in `scraper.ts` —
 * so the parsing logic is easy to unit-test or reuse against a saved HTML
 * snapshot.
 */
import type { AverageStats, HeroOption } from "./types.ts";
import { normalizeSets } from "./sets.ts";

/** Map a stat-preview image filename stem to a normalized stat key. */
const STAT_BY_IMAGE_STEM: Record<string, keyof AverageStats> = {
  statatkdark: "atk",
  statdefdark: "def",
  stathpdark: "hp",
  statspddark: "spd",
  statcrdark: "cc", // critical rate
  statcddark: "cd", // critical damage
  stateffdark: "eff",
  statresdark: "res",
  star: "gs", // gear score uses star.png
  statdacdark: "dac", // dual attack chance (future-proofing)
};

/** Map the page's stat element ids to normalized stat keys. */
const STAT_BY_ID: Record<string, keyof AverageStats> = {
  atkStatBefore: "atk",
  defStatBefore: "def",
  hpStatBefore: "hp",
  spdStatBefore: "spd",
  crStatBefore: "cc",
  cdStatBefore: "cd",
  effStatBefore: "eff",
  resStatBefore: "res",
  gsStatBefore: "gs",
};

export interface RawBuildRow {
  rank: number;
  /** Raw usage text, e.g. "32.6%". */
  usageText: string;
  /** Raw `<img src>` values for the set icons in display order. */
  setImageSrcs: string[];
}

export interface RawArtifactRow {
  usageText: string;
  name: string;
}

/* ----------------------------- pure helpers ----------------------------- */

/** Parse a percentage string like "32.6%" into the number 32.6. */
export function parsePercent(text: string | undefined | null): number {
  if (!text) return 0;
  const m = text.replace(/,/g, "").match(/-?\d+(\.\d+)?/);
  return m ? parseFloat(m[0]) : 0;
}

/** Parse a possibly-formatted integer stat ("4,440", "4440", "") -> number. */
export function parseIntStat(text: string | undefined | null): number | undefined {
  if (!text) return undefined;
  const cleaned = text.trim().replace(/,/g, "");
  if (cleaned === "" || cleaned === "-" || cleaned === "?") return undefined;
  const n = Number(cleaned);
  return Number.isFinite(n) ? n : undefined;
}

/** Extract the filename stem from an asset URL/path. */
function imageStem(src: string): string {
  const file = src.split(/[\\/]/).pop() ?? src;
  return file.replace(/\.(png|jpe?g|webp|gif|svg)$/i, "");
}

/* --------------------------- DOM extraction ----------------------------- */

/**
 * Read every hero option from the underlying `<select>` (populated by the
 * page's JS from herodata.json). The first option is the blank placeholder.
 */
export async function getHeroList(page: import("playwright").Page): Promise<HeroOption[]> {
  const options = await page.evaluate(() => {
    const sel = document.querySelector("#heroSelector0");
    if (!sel) return [];
    return Array.from(sel.querySelectorAll("option"))
      .map((o) => ({ id: o.value, name: (o.textContent ?? "").trim() }))
      .filter((o) => o.id && o.name);
  });
  return options;
}

/** Currently-selected hero name from the select2 control. */
export async function getSelectedHeroName(
  page: import("playwright").Page,
): Promise<string | undefined> {
  return page.evaluate(() => {
    const data = (window as any).jQuery?.("#heroSelector0").select2("data");
    return data?.[0]?.text as string | undefined;
  });
}

/**
 * Read the top-N build archetype rows from `#setCombos`.
 * Returns at most `topN` rows, in usage order.
 */
export async function readBuildRows(
  page: import("playwright").Page,
  topN: number,
): Promise<RawBuildRow[]> {
  return page.evaluate(
    (n) => {
      const rows = Array.from(document.querySelectorAll<HTMLElement>("#setCombos .setComboRow"));
      const out: { rank: number; usageText: string; setImageSrcs: string[] }[] = [];
      rows.slice(0, n).forEach((row, i) => {
        const usageEl = row.querySelector(".setComboRowText");
        const imgs = Array.from(row.querySelectorAll<HTMLImageElement>(".setComboRowImages img"));
        out.push({
          rank: i + 1,
          usageText: (usageEl?.textContent ?? "").trim(),
          setImageSrcs: imgs.map((img) => img.getAttribute("src") ?? img.src ?? ""),
        });
      });
      return out;
    },
    topN,
  );
}

/**
 * Read the top recommended artifact (first row of `#artifactCombos`).
 * This is the hero-level recommended artifact shown in the UI.
 */
export async function readTopArtifact(
  page: import("playwright").Page,
): Promise<RawArtifactRow | undefined> {
  return page.evaluate(() => {
    const row = document.querySelector("#artifactCombos .artifactComboRow");
    if (!row) return undefined;
    const usage = row.querySelector(".setArtifactRowLeft")?.textContent ?? "";
    const name = row.querySelector(".setArtifactRowRight")?.textContent ?? "";
    const u = usage.trim();
    const nm = name.trim();
    if (!u && !nm) return undefined;
    return { usageText: u, name: nm };
  });
}

/**
 * Read the "Average" stat panel.
 *
 * Primary path: read the known stat element ids (#atkStatBefore, ...).
 * Defensive path: also scan every `.statPreviewRow` and map its icon to a key,
 * so any *additional* stat the UI may render in the future is still captured
 * (per the "any other average stat displayed" requirement).
 */
export async function readAverageStats(
  page: import("playwright").Page,
): Promise<AverageStats> {
  return page.evaluate((statById) => {
    const stats: Record<string, number | undefined> = {};

    // 1. Known ids — fast path for the stats we know the page renders.
    for (const [id, key] of Object.entries(statById)) {
      const el = document.getElementById(id);
      const txt = (el?.textContent ?? "").trim();
      if (txt) {
        const n = Number(txt.replace(/,/g, ""));
        stats[key as string] = Number.isFinite(n) ? n : undefined;
      }
    }

    // 2. Generic scan for any *other* stat rows (future-proofing). We map the
    //    row's icon filename to a stat key; anything already captured is
    //    skipped so the known-ids values always win.
    const stemToKey: Record<string, string> = {
      statatkdark: "atk",
      statdefdark: "def",
      stathpdark: "hp",
      statspddark: "spd",
      statcrdark: "cc",
      statcddark: "cd",
      stateffdark: "eff",
      statresdark: "res",
      star: "gs",
      statdacdark: "dac",
    };
    const rows = document.querySelectorAll("#selectedBuild .statPreviewRow");
    rows.forEach((row) => {
      const img = row.querySelector<HTMLImageElement>(".statPreviewImg");
      const valEl = row.querySelector(".statPreviewBefore");
      const txt = (valEl?.textContent ?? "").trim();
      if (!img || !txt) return;
      const src = img.getAttribute("src") ?? img.src ?? "";
      const file = src.split(/[\\/]/).pop() ?? src;
      const stem = file.replace(/\.(png|jpe?g|webp|gif|svg)$/i, "");
      const key = stemToKey[stem];
      if (!key || stats[key] !== undefined) return;
      const n = Number(txt.replace(/,/g, ""));
      if (Number.isFinite(n)) stats[key] = n;
    });

    return stats;
  }, STAT_BY_ID);
}

/* --------------------------- normalization ------------------------------ */

/** Turn a raw build row into a normalized sets[] list. */
export function buildSets(row: RawBuildRow): string[] {
  return normalizeSets(row.setImageSrcs);
}

/** Map a stat image stem to a stat key (exported for tests/debugging). */
export function statKeyFromImage(src: string): keyof AverageStats | undefined {
  return STAT_BY_IMAGE_STEM[imageStem(src)];
}
