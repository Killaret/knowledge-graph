import { test, expect } from "@playwright/test";
import { setupSkipAuth } from "./helpers/testUtils";

/**
 * E2E for the URL-HEADING-1 stage-A preview contract: the preview row shows
 * a title-candidates dropdown and a collapsible outline; picking a candidate
 * becomes the imported title.
 */

const FIXTURE_ITEM = {
  title: "JSON Formatter",
  url: "https://example.com/article",
  text: "extracted text",
  type: "asteroid",
  is_new: true,
  title_candidates: ["JSON Formatter", "JSON Formatter - example.com", "article"],
  title_source: "rule",
  outline: [
    { level: 2, text: "JSON Formatter" },
    { level: 3, text: "Features" },
    { level: 3, text: "Usage" },
  ],
  noise_dropped: 7,
};

test.describe(
  "Import preview extraction fields",
  { tag: ["@e2e", "@import", "@skip-auth"] },
  () => {
    test.beforeEach(async ({ page }) => {
      await setupSkipAuth(page);
      await page.route("**/api/v1/import/bookmarks/preview", async (route) => {
        await route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify({ data: { items: [FIXTURE_ITEM] } }),
        });
      });
      await page.goto("/import/bookmarks");
      await page.waitForLoadState("networkidle");

      const input = page.locator("textarea#import-list");
      await input.fill(`x | ${FIXTURE_ITEM.url}`);
      await page.getByRole("button", { name: /Preview/i }).click();
      await expect(page.locator(".preview-table tbody tr")).toHaveCount(1, { timeout: 10000 });
    });

    test("shows title candidates dropdown and applies the selection", async ({ page }) => {
      const select = page.getByTestId("title-candidates");
      await expect(select).toBeVisible();
      await expect(select).toHaveValue("JSON Formatter");

      await select.selectOption("JSON Formatter - example.com");
      await expect(page.locator(".table-input")).toHaveValue("JSON Formatter - example.com");
    });

    test("shows a collapsible page outline with noise counter", async ({ page }) => {
      const outline = page.getByTestId("item-outline");
      await expect(outline).toBeVisible();
      await expect(outline.locator("summary")).toContainText("3");
      await expect(outline.locator("summary")).toContainText("7");

      await outline.locator("summary").click();
      const items = outline.locator(".outline-item");
      await expect(items).toHaveCount(3);
      await expect(items.nth(0)).toHaveText("JSON Formatter");
      await expect(items.nth(2)).toHaveText("Usage");
    });
  }
);
