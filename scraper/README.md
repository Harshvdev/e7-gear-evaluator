# e7-hero-library-scraper

A production-quality, resumable scraper for the
[Fribbels Epic Seven Hero Library](https://fribbels.github.io/e7/hero-library.html).

For every hero it extracts the **top 5 most-used build archetypes** (the
summarized build list shown *above* the big build table), each with:

- usage percentage
- normalized, human-readable set names (primary first)
- average stats (ATK, DEF, HP, SPD, CC, CD, EFF, RES, GS — plus any other stat
  the UI may render in the Average panel)
- the hero's recommended artifact + its usage percentage

Results are written to `out/heroes.json` (pretty array) and `out/heroes.jsonl`
(append-only, one record per line — the durable store used for resuming).

## Quick start

```bash
cd scraper
bun install                       # installs playwright (chromium is pre-cached)

# Scrape a single hero (great for a quick check):
bun run index.ts --hero "Afternoon Soak Flan"

# Scrape a limited batch with conservative pacing:
bun run index.ts --limit 20 --delay 2000

# Full run, all 384 heroes:
bun run index.ts

# Resume an interrupted run (skips heroes already in heroes.jsonl):
bun run index.ts --resume
```

### Options

| Flag              | Default | Description                                            |
| ----------------- | ------- | ------------------------------------------------------ |
| `--url <url>`     | see cfg | Base hero-library URL                                  |
| `--top <n>`       | `5`     | Top N build archetypes per hero                        |
| `--delay <ms>`    | `2000`  | Delay between heroes (anti rate-limit pacing)          |
| `--build-delay <ms>` | `250` | Delay between build-row clicks (lets averages settle) |
| `--timeout <ms>`  | `45000` | Per-hero timeout                                       |
| `--limit <n>`     | `0`     | Stop after N heroes (0 = unlimited)                    |
| `--hero <name>`   | —       | Scrape a single hero by name (test mode)               |
| `--resume`        | off     | Resume from persisted progress                         |
| `--headed`        | off     | Show the browser window                                |
| `--out <dir>`     | `./out` | Output/state directory                                 |

Set `DEBUG=1` for verbose logging.

## Architecture

The code is intentionally split into small, single-responsibility modules so
each concern can evolve independently:

```
scraper/
├── index.ts      # Orchestrator + CLI parsing + graceful shutdown
├── config.ts     # Default config + argv parsing
├── browser.ts    # Single reusable Chromium instance + resource blocking
├── scraper.ts    # Navigation / clicking / waiting (the "scraping")
├── parser.ts     # DOM -> raw typed values (the "parsing")
├── sets.ts       # Internal id -> official set name (the "normalization")
├── output.ts     # JSON + JSONL writers (the "output")
├── state.ts      # Resumable progress state
├── logger.ts     # Leveled, timestamped logger
└── types.ts      # Shared domain types
```

### How a hero is scraped

1. The hero is selected in the page's `select2` dropdown **without reloading
   the page**, so the large `herodata.json` / `artifactdata.json` are fetched
   only once for the whole run.
2. The Search button is clicked; we await the `getBuilds` API response, then
   poll until the build list renders (or the page shows its "No data" overlay).
3. The top-N build rows are read (usage % + set-icon `src` attributes).
4. For each build row we click it so the page recomputes that set-combo's
   average stats in the Average panel, then read the panel. We only ever read
   values the UI itself displays — we never compute averages ourselves.
5. The hero's top recommended artifact (first row of the artifact panel) is
   attached to each build, matching the requested output shape.

### Performance / anti-rate-limit design

- **Single browser instance** reused for every hero (warm V8, shared HTTP cache).
- **Resource blocking**: images, fonts, media and analytics are aborted. Set
  names are read from `<img src>` *attributes* (present in the DOM regardless
  of whether the pixel downloads), so blocking images loses no data and cuts
  bandwidth dramatically.
- **No per-hero page reloads** — hero switching happens in-page via select2.
- **Conservative pacing**: configurable delay between heroes (`--delay`,
  default 2500 ms) and between build clicks (`--build-delay`).
- **No duplicate requests**: the `getBuilds` endpoint is hit exactly once per
  hero, never more.
- **Resumable**: `out/progress.json` + append-only `out/heroes.jsonl` mean an
  interrupted run picks up exactly where it left off. SIGINT/SIGTERM trigger a
  graceful shutdown that flushes the final JSON.

### Rate limiting (observed)

The `getBuilds` endpoint is an AWS API Gateway that **soft-bans** IPs which
burst too many requests, responding `HTTP 403` (not `429`). Once tripped, the
ban persists for a while and affects every hero. The scraper handles this
explicitly so no data is silently lost and no hero is wrongly marked "no data":

- Each hero is retried with exponential backoff (`5s → 15s → 40s`, configurable
  via `rateLimitBackoffMs`) on `403/429/503`.
- A rate-limited hero is **never** marked completed — it stays pending so the
  next `--resume` retries it.
- If **5 consecutive** heroes are rate-limited, the run aborts gracefully
  (saving progress) since the IP is soft-banned; re-run `--resume` later.
- Genuine `200 + empty data` is still correctly recorded as `"no data"` (an
  empty `builds` array), distinct from a rate-limit.

For a full 384-hero scrape, use conservative pacing (e.g. `--delay 4000`) and
run in multiple `--resume` sessions — the resume support makes this trivial.
A run that hits the soft ban can simply be re-invoked with `--resume` once the
ban lifts; already-completed heroes are skipped.

### Set-name normalization

`sets.ts` maps the page's internal identifiers to official Epic Seven set
names. It handles both encoding forms the UI uses:

- ingame codes: `set_att` → `Attack`, `set_cri_dmg` → `Destruction`,
  `set_opener` → `Warfare`, `set_res` → `Resistance`, `set_acc` → `Hit`, …
- image filename stems: `setdestruction.png` → `Destruction`, `sethit.png` →
  `Hit`, …

Unknown identifiers fall back to a cleaned, title-cased version so a new set is
never silently dropped. Adding a new set is a one-line change in `sets.ts`.

## Known source-page quirks (faithfully handled)

These are behaviors of the *scraped* site, not the scraper. The scraper reads
exactly what the UI displays.

1. **Buggy filter toggle.** The page's `objectsAreEqual` only compares object
   key *counts*, not values, so clicking a build row whose set-combo has the
   same number of sets as the currently-active filter **toggles the filter
   off** (showing *all* builds) instead of switching to the new combo. The
   scraper works around this by deactivating the active row before each click,
   so every build's average reflects its own set-combo — exactly what a careful
   user would see.

2. **Superset averages for partial-set combos.** The page's row filter uses
   `>=` on raw set piece-counts. For a "full" 4+2 combo (e.g. Destruction +
   Penetration) this matches exactly that combo. For a *partial* / single-set
   combo (e.g. a lone 4-piece Speed with two non-completing pieces) the filter
   matches a **superset** of builds, so the UI's displayed average for that row
   is the superset average, not the exact-combo average. The scraper reports
   what the UI displays (per the "use only data displayed in the UI" /
   "do not calculate anything yourself" requirements); it never invents values.

3. **The 9th Average stat is Gear Score, not Dual Attack Chance.** The example
   in the task spec shows `dac: 430`, but the panel's 9th stat (the `430`
   value) is actually **Gear Score** (rendered with a star icon, element id
   `gsStatBefore`). Dual Attack Chance is not present in the Average panel, so
   it is omitted; the gear score is captured as `gs`. (`dac` would be captured
   automatically if the UI ever adds it — see the generic scan in `parser.ts`.)

## Output shape

```json
{
  "hero": "Afternoon Soak Flan",
  "builds": [
    {
      "rank": 1,
      "usage": 32.6,
      "sets": ["Destruction", "Penetration"],
      "averageStats": {
        "atk": 4440, "def": 1063, "hp": 13116, "spd": 183,
        "cc": 39, "cd": 347, "eff": 23, "res": 3, "gs": 430
      },
      "artifact": { "name": "Dreamlike Holiday", "usage": 87.1 }
    }
  ]
}
```

`heroes.json` is an array of these records; `heroes.jsonl` is the same records,
one per line.
