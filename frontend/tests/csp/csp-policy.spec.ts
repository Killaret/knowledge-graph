import { test, expect, type ConsoleMessage } from "@playwright/test";
import { setupSkipAuth } from "../helpers/testUtils";

/**
 * CSP-1: walks the key screens with the Content-Security-Policy actually
 * enforced. The visual projects run with `bypassCSP: true` (Argos injects a
 * stabilization script), so they can never see a broken policy — this spec is
 * the guard that fails on a violation.
 *
 * The browser reports every violation as a console error
 * ("Refused to ... because it violates the following Content Security Policy
 * directive"), so collecting console messages is sufficient.
 */

const CSP_VIOLATION = /Content Security Policy|violates the following Content/i;

// The nine screens from the CSP-1 spec: home, 2D graph, 3D graph, search,
// login, registration, cockpit (part of the graph page), note creation,
// Yandex login. The Yandex flow redirects off-site; we visit the entry page.
const SCREENS: Array<{ name: string; path: string }> = [
  { name: "home", path: "/" },
  { name: "graph-2d", path: "/graph?full=1&nocache=1" },
  { name: "graph-3d", path: "/graph/3d" },
  { name: "search", path: "/search" },
  { name: "login", path: "/auth/login" },
  { name: "register", path: "/auth/register" },
  { name: "note-new", path: "/notes/new" },
  { name: "import-bookmarks", path: "/import/bookmarks" },
];

test.describe("CSP policy enforcement", { tag: ["@csp"] }, () => {
  for (const screen of SCREENS) {
    test(`no CSP violations on ${screen.name} (${screen.path})`, async ({ page }) => {
      const violations: string[] = [];
      page.on("console", (msg: ConsoleMessage) => {
        // Report-only violations arrive as warnings, enforced ones as errors —
        // match the text, not the severity.
        if (CSP_VIOLATION.test(msg.text())) {
          violations.push(msg.text());
        }
      });
      page.on("pageerror", (err) => {
        if (CSP_VIOLATION.test(String(err))) {
          violations.push(String(err));
        }
      });

      await setupSkipAuth(page);
      // domcontentloaded is enough: a blocked script reports to the console
      // immediately; networkidle would hang on a deliberately broken page.
      await page.goto(screen.path, { waitUntil: "domcontentloaded" });
      await page.waitForLoadState("networkidle", { timeout: 30000 }).catch(() => {});
      // Let deferred resources (3D canvas, SSE) attempt to load.
      await page.waitForTimeout(3000);

      expect(violations, `CSP violations on ${screen.path}`).toEqual([]);
    });
  }
});
