import { test, expect, type Page } from "@playwright/test";
import { argosScreenshot } from "@argos-ci/playwright";
import { clickViewToggle, openCockpitPanel } from "../helpers/testUtils";

/**
 * Visual Regression Tests with Argos Playwright SDK
 *
 * - Uses argosScreenshot for automatic stabilization and upload.
 * - Injects a seeded Math.random and __SKIP_AUTH__ before each test.
 * - Runs against the isolated test stack with seeded data.
 */

const STABLE_RENDER = "?stableRender=true";

const VIEWPORTS = [
  { width: 1920, height: 1080 },
  { width: 768, height: 1024 },
  { width: 375, height: 667 },
];

test.describe("Visual Regression @visual", { tag: "@visual" }, () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => {
      // Enable auth bypass for isolated visual tests
      (window as any).__SKIP_AUTH__ = true;

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

  test("Home page - default view", async ({ page }) => {
    await page.goto("/" + STABLE_RENDER);
    await waitForApp(page);
    await argosScreenshot(page, "home-default", { fullPage: true });
  });

  test("Home page - list view", async ({ page }) => {
    await page.goto("/" + STABLE_RENDER);
    await waitForApp(page);

    await clickViewToggle(page, "list");
    await page.waitForTimeout(500);

    await argosScreenshot(page, "home-list-view", { fullPage: true });
  });

  test("Home page - with star filter", async ({ page }) => {
    await page.goto("/" + STABLE_RENDER);
    await waitForApp(page);

    await openCockpitPanel(page, "left");
    const filterButton = page.locator('[data-testid="filter-chip-star"]');
    await expect(filterButton).toBeVisible({ timeout: 5000 });
    await filterButton.click();
    await page.waitForTimeout(500);

    await argosScreenshot(page, "home-filtered-stars", { fullPage: true });
  });

  test("2D Graph - full view with links", async ({ page }) => {
    await page.goto("/graph" + STABLE_RENDER);
    await waitForGraph(page);

    await argosScreenshot(page, "2d-graph-full", { fullPage: true });
  });

  test("2D Graph - ghost node creation form", async ({ page }) => {
    await page.goto("/graph" + STABLE_RENDER);
    await waitForGraph(page);

    await page.keyboard.press("N");
    await page.waitForTimeout(500);

    await argosScreenshot(page, "2d-ghost-node-form", { fullPage: true });
  });

  test("2D Graph - help modal", async ({ page }) => {
    await page.goto("/graph" + STABLE_RENDER);
    await waitForGraph(page);

    await page.keyboard.press("?");
    await page.waitForTimeout(500);

    await argosScreenshot(page, "2d-help-modal", { fullPage: true });
  });

  test("NoteCard - selected state", async ({ page }) => {
    await page.goto("/" + STABLE_RENDER);
    await waitForApp(page);

    await clickViewToggle(page, "list");
    await page.waitForTimeout(500);

    const selectButton = page.locator('[data-testid="select-mode-toggle"]');
    await expect(selectButton).toBeVisible({ timeout: 5000 });
    await selectButton.click();
    await page.waitForTimeout(500);

    const firstNote = page.locator('[data-testid="note-card"]').first();
    await expect(firstNote).toBeVisible({ timeout: 5000 });
    await firstNote.click();
    await page.waitForTimeout(500);

    await argosScreenshot(page, "notecard-selected", { fullPage: true });
  });

  test("Search page", async ({ page }) => {
    await page.goto("/search" + STABLE_RENDER);
    await waitForApp(page);
    await argosScreenshot(page, "search-page", { fullPage: true });
  });

  test("Search with query", async ({ page }) => {
    await page.goto("/search?q=star" + STABLE_RENDER);
    await waitForApp(page);

    const searchInput = page.locator('[data-testid="search-input"]').first();
    await expect(searchInput).toBeVisible({ timeout: 5000 });
    await page.waitForTimeout(500);

    await argosScreenshot(page, "search-with-query", { fullPage: true });
  });

  test("Empty state", async ({ page }) => {
    await page.goto("/search?q=nonexistentquery123456789" + STABLE_RENDER);
    await waitForApp(page);
    await argosScreenshot(page, "empty-state", { fullPage: true });
  });

  test("3D Graph - renders 3D view", async ({ page }) => {
    await page.goto("/graph/3d" + STABLE_RENDER);
    const viewer = page.locator('[data-testid="graph-3d-viewer"]');
    await expect(viewer).toBeVisible({ timeout: 15000 });
    await argosScreenshot(page, "3d-graph-view", { fullPage: true });
  });

  test("Home responsive viewports", async ({ page }) => {
    await page.goto("/" + STABLE_RENDER);
    await waitForApp(page);
    await argosScreenshot(page, "home-responsive", {
      fullPage: true,
      viewports: VIEWPORTS,
    });
  });
});
