import { test, expect, type Page } from "@playwright/test";
import { argosScreenshot } from "@argos-ci/playwright";

/**
 * Anonymous visual baseline — what a visitor without a session sees.
 *
 * These scenarios run without storageState and without the __SKIP_AUTH__
 * injection: the backend must be started with SKIP_AUTH=false, and every
 * screenshot here must render for a logged-out user. If a scenario needs a
 * session, it belongs in visual-authenticated.spec.ts.
 */

const STABLE_RENDER = "?stableRender=true";

const VIEWPORTS = [
  { width: 1920, height: 1080 },
  { width: 768, height: 1024 },
  { width: 375, height: 667 },
];

test.describe("Visual Regression - anonymous @visual", { tag: "@visual" }, () => {
  test.beforeEach(async ({ page }) => {
    await page.emulateMedia({ reducedMotion: "reduce" });
    await page.addInitScript(() => {
      // Seeded Math.random for deterministic canvas / d3-force / particle output
      let seed = 12345;
      const m = 2 ** 31;
      Math.random = function seededRandom() {
        seed = (1103515245 * seed + 12345) % m;
        return seed / m;
      };
    });
  });

  async function waitForApp(page: Page) {
    await expect(page.locator("main")).toBeVisible({ timeout: 15000 });
  }

  async function waitForGraph(page: Page) {
    const canvas = page.locator('[data-testid="graph-canvas"][data-test-stable="true"]');
    await canvas.waitFor({ timeout: 15000 });
  }

  test("Login page", async ({ page }) => {
    await page.goto("/auth/login" + STABLE_RENDER);
    await expect(page.locator('input[name="login"]')).toBeVisible({ timeout: 15000 });
    await argosScreenshot(page, "anon-login-page", { fullPage: true });
  });

  test("Register page", async ({ page }) => {
    await page.goto("/auth/register" + STABLE_RENDER);
    await expect(page.locator("form.register-form")).toBeVisible({ timeout: 15000 });
    await argosScreenshot(page, "anon-register-page", { fullPage: true });
  });

  test("Home page - default view", async ({ page }) => {
    await page.goto("/" + STABLE_RENDER);
    await waitForApp(page);
    await argosScreenshot(page, "anon-home-default", { fullPage: true });
  });

  test("Public 2D graph", async ({ page }) => {
    await page.goto("/graph" + STABLE_RENDER);
    await waitForGraph(page);
    await argosScreenshot(page, "anon-public-graph", { fullPage: true });
  });

  test("Search page", async ({ page }) => {
    await page.goto("/search" + STABLE_RENDER);
    await waitForApp(page);
    await argosScreenshot(page, "anon-search-page", { fullPage: true });
  });

  test("Empty state", async ({ page }) => {
    await page.goto("/search?q=nonexistentquery123456789" + STABLE_RENDER);
    await waitForApp(page);
    await argosScreenshot(page, "anon-empty-state", { fullPage: true });
  });

  test("Home responsive viewports", async ({ page }) => {
    await page.goto("/" + STABLE_RENDER);
    await waitForApp(page);
    await argosScreenshot(page, "anon-home-responsive", {
      fullPage: true,
      viewports: VIEWPORTS,
    });
  });
});
