/**
 * Browser lifecycle management.
 *
 * A single Chromium instance (and a single incognito context) is created once
 * and reused for every hero. This is the single most important anti-rate-limit
 * optimization: no repeated browser/process startup, warm V8 isolates, shared
 * HTTP cache, and a single TCP connection pool.
 *
 * Resource blocking: the scraper never needs rendered pixels. Set names are
 * read from `<img src>` *attributes* (present in the DOM regardless of whether
 * the image actually downloads), so we block images, fonts and media to keep
 * each hero load to a handful of XHR calls + the page JS. Analytics/gtag is
 * blocked too.
 */
import { chromium, type Browser, type BrowserContext, type Page } from "playwright";
import { logger } from "./logger.ts";

/** A realistic desktop User-Agent to avoid trivial bot fingerprinting. */
const USER_AGENT =
  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
  "(KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36";

const BLOCKED_HOST_FRAGMENTS = [
  "googletagmanager.com",
  "google-analytics.com",
  "buymeacoffee.com",
];

export interface BrowserHandle {
  browser: Browser;
  context: BrowserContext;
  page: Page;
  close: () => Promise<void>;
}

let handle: BrowserHandle | undefined;

/**
 * Lazily launch a shared browser + context + page and navigate to the hero
 * library once. Subsequent calls return the cached handle.
 */
export async function getBrowser(
  baseUrl: string,
  opts: { headless: boolean; extraHeaders?: Record<string, string> },
): Promise<BrowserHandle> {
  if (handle) return handle;

  logger.info("launching browser", { headless: opts.headless });
  const browser = await chromium.launch({
    headless: opts.headless,
    args: [
      "--disable-blink-features=AutomationControlled",
      "--disable-dev-shm-usage",
      "--no-sandbox",
    ],
  });

  const context = await browser.newContext({
    userAgent: USER_AGENT,
    locale: "en-US",
    timezoneId: "America/Los_Angeles",
    viewport: { width: 1366, height: 900 },
    extraHTTPHeaders: {
      "Accept-Language": "en-US,en;q=0.9",
      ...(opts.extraHeaders ?? {}),
    },
    // Bypass the navigator.webdriver flag.
    bypassCSP: true,
  });

  // Mask `navigator.webdriver` for good measure.
  await context.addInitScript(() => {
    Object.defineProperty(navigator, "webdriver", { get: () => undefined });
  });

  // Aggressively drop resource types we never need.
  await context.route("**/*", (route) => {
    const req = route.request();
    const type = req.resourceType();
    const url = req.url();

    if (type === "image" || type === "font" || type === "media") {
      return route.abort();
    }
    if (BLOCKED_HOST_FRAGMENTS.some((h) => url.includes(h))) {
      return route.abort();
    }
    return route.continue();
  });

  const page = await context.newPage();
  page.setDefaultTimeout(30_000);
  page.setDefaultNavigationTimeout(60_000);

  logger.info("navigating to hero library", { url: baseUrl });
  await page.goto(baseUrl, { waitUntil: "domcontentloaded" });

  // Wait for the hero <select> to be populated by the page's own JS
  // (herodata.json + artifactdata.json must finish loading first).
  await page.waitForFunction(
    () => {
      const sel = document.querySelector("#heroSelector0");
      if (!sel) return false;
      return sel.querySelectorAll("option").length > 1; // first option is the blank ""
    },
    undefined,
    { timeout: 45_000, polling: 500 },
  );
  logger.info("hero library ready");

  handle = {
    browser,
    context,
    page,
    close: async () => {
      try {
        await context.close();
      } catch {
        /* ignore */
      }
      try {
        await browser.close();
      } catch {
        /* ignore */
      }
      handle = undefined;
    },
  };

  return handle;
}

/** Close the shared browser if one is open. Safe to call multiple times. */
export async function closeBrowser(): Promise<void> {
  if (handle) await handle.close();
}
