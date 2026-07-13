/**
 * Set-name normalization.
 *
 * The hero-library UI encodes gear sets in two ways:
 *   1. As image filenames in the rendered set icons, e.g. `./assets/setdestruction.png`.
 *   2. As internal ingame identifiers, e.g. `set_att`, `set_cri_dmg`, `set_opener`.
 *
 * This module turns any of those (and a few defensive fallbacks such as the
 * `DestructionSet` class-name style) into a single, human-readable, official
 * Epic Seven set name (e.g. `Destruction`, `Penetration`, `Warfare`).
 *
 * The mapping is data-driven so adding a new set later is a one-line change.
 */

/**
 * Mapping from internal ingame set identifiers to official, human-readable
 * Epic Seven set names. `set_opener -> Warfare` is included per the task spec
 * example; the rest are derived from the page's own `ingameSetsToSetNames`
 * table (with the `Set` suffix stripped to match the requested output style).
 */
export const SET_BY_INGAME_CODE: Record<string, string> = {
  set_acc: "Hit",
  set_att: "Attack",
  set_coop: "Unity",
  set_counter: "Counter",
  set_cri_dmg: "Destruction",
  set_cri: "Critical",
  set_def: "Defense",
  set_immune: "Immunity",
  set_max_hp: "Health",
  set_penetrate: "Penetration",
  set_rage: "Rage",
  set_res: "Resistance",
  set_revenge: "Revenge",
  set_scar: "Injury",
  set_speed: "Speed",
  set_vampire: "Lifesteal",
  set_shield: "Protection",
  set_torrent: "Torrent",
  set_revenant: "Reversal",
  set_riposte: "Riposte",
  // Spec-provided example mapping.
  set_opener: "Warfare",
};

/**
 * Mapping from the image asset filename stem (no path, no extension) to the
 * official set name. Mirrors the page's `assetsBySet` table.
 */
export const SET_BY_IMAGE_STEM: Record<string, string> = {
  sethealth: "Health",
  setdefense: "Defense",
  setattack: "Attack",
  setspeed: "Speed",
  setcritical: "Critical",
  sethit: "Hit",
  setdestruction: "Destruction",
  setlifesteal: "Lifesteal",
  setcounter: "Counter",
  setresist: "Resistance",
  setunity: "Unity",
  setrage: "Rage",
  setimmunity: "Immunity",
  setrevenge: "Revenge",
  setinjury: "Injury",
  setpenetration: "Penetration",
  setprotection: "Protection",
  settorrent: "Torrent",
  setrevenant: "Reversal",
  setriposte: "Riposte",
};

/** Reverse-lookup helper: official name -> ingame code (useful for debugging). */
export const INGAME_CODE_BY_SET: Record<string, string> = Object.fromEntries(
  Object.entries(SET_BY_INGAME_CODE).map(([code, name]) => [name, code]),
);

/**
 * Normalize any set identifier we might encounter into an official name.
 *
 * Accepts (in priority order):
 *   - internal ingame code (`set_att`, `set_opener`, ...)
 *   - image src / filename stem (`./assets/setdestruction.png`, `setdestruction`)
 *   - `DestructionSet`-style class names
 *   - already-official names (`Destruction`)
 *
 * Falls back to a cleaned title-cased version of the input so the scraper
 * never silently drops a set it hasn't seen before.
 */
export function normalizeSet(raw: string): string {
  if (!raw) return "";

  const trimmed = String(raw).trim();

  // 1. Direct ingame-code match (e.g. "set_att").
  if (SET_BY_INGAME_CODE[trimmed]) return SET_BY_INGAME_CODE[trimmed];

  // 2. Extract a filename stem from a URL/path and try both the stem and the
  //    bare filename (without the leading "set" if it's actually "set_...").
  const stem = trimmed.split(/[\\/]/).pop() ?? trimmed;
  const noExt = stem.replace(/\.(png|jpe?g|webp|gif|svg)$/i, "");

  if (SET_BY_INGAME_CODE[noExt]) return SET_BY_INGAME_CODE[noExt];
  if (SET_BY_IMAGE_STEM[noExt]) return SET_BY_IMAGE_STEM[noExt];

  // 3. `DestructionSet` style -> `Destruction`.
  if (/Set$/i.test(noExt)) {
    const stripped = noExt.replace(/Set$/i, "");
    if (SET_BY_IMAGE_STEM[`set${stripped.toLowerCase()}`]) {
      return SET_BY_IMAGE_STEM[`set${stripped.toLowerCase()}`];
    }
    return stripped;
  }

  // 4. Last resort: title-case the stem so unknown sets are still readable.
  return noExt
    .replace(/^set_?/i, "")
    .replace(/[_-]+/g, " ")
    .replace(/\b\w/g, (c) => c.toUpperCase());
}

/**
 * Normalize a list of raw set identifiers into an ordered, de-duplicated list
 * of official names. Order is preserved (primary set first) because the UI
 * renders 4-piece sets before 2-piece sets and we want to keep that ordering.
 */
export function normalizeSets(rawSets: string[]): string[] {
  const out: string[] = [];
  const seen = new Set<string>();
  for (const raw of rawSets) {
    const name = normalizeSet(raw);
    if (!name || seen.has(name)) continue;
    seen.add(name);
    out.push(name);
  }
  return out;
}
