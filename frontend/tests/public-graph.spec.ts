import { test, expect } from "@playwright/test";

// Public graph requires real auth; in SKIP_AUTH stack all notes are private
// and the graph-service public endpoint returns an empty graph.
test.skip(process.env.SKIP_AUTH === "true", "Public graph not available in SKIP_AUTH stack");

/**
 * Public graph: unauthenticated users should see the read-only graph.
 */
test.describe("Public Graph", () => {
  test("loads graph for unauthenticated users", async ({ page }) => {
    // Ensure we are not in SKIP_AUTH mode
    await page.addInitScript(() => {
      localStorage.removeItem("__SKIP_AUTH__");
      if ("__SKIP_AUTH__" in window) {
        delete (window as any).__SKIP_AUTH__;
      }
    });

    await page.goto("/");
    await page.waitForLoadState("networkidle");

    const canvas = page.locator('[data-testid="graph-canvas"]');
    await expect(canvas).toBeVisible({ timeout: 15000 });

    const stats = page.locator('[data-testid="graph-stats"]');
    await expect(stats).toBeVisible({ timeout: 15000 });
    // Require a non-zero node count explicitly — a bare /\d+/ regex would
    // incorrectly pass on "0 nodes", hiding an empty-graph regression (see
    // public-graph-real-auth.spec.ts for the historical "empty public graph" bug).
    const statsText = await stats.textContent({ timeout: 5000 });
    expect(statsText).toMatch(/[1-9]\d*\s+nodes?/i);
    expect(statsText).toMatch(/\d+\s+links?/i);
  });
});
