import { test, expect } from "@playwright/test";
import { setupSkipAuth } from "./helpers/testUtils";

/**
 * Test to verify SKIP_AUTH is working correctly
 */

test.describe("SKIP_AUTH Verification", () => {
  test("should access main page without authentication", async ({ page }) => {
    // Setup SKIP_AUTH first
    await setupSkipAuth(page);

    // Now navigate to main page
    await page.goto("/");
    await page.waitForTimeout(500);

    // Verify SKIP_AUTH flag is set after navigation
    const skipAuthFlag = await page.evaluate(() => (window as any).__SKIP_AUTH__);
    expect(skipAuthFlag).toBe(true);

    // Should NOT redirect to login
    await expect(page).toHaveURL("/");

    // Should show graph or notes (or at least not redirect)
    const currentUrl = page.url();
    expect(currentUrl).not.toContain("/auth/login");
  });

  test("should access graph page without authentication", async ({ page }) => {
    // Setup SKIP_AUTH first
    await setupSkipAuth(page);

    // Navigate to graph page
    await page.goto("/graph");
    await page.waitForLoadState("domcontentloaded");
    await page.waitForTimeout(2000);

    // Verify SKIP_AUTH flag is set after navigation
    const skipAuthFlag = await page.evaluate(() => (window as any).__SKIP_AUTH__);
    expect(skipAuthFlag).toBe(true);

    // Should NOT redirect to login
    await expect(page).toHaveURL("/graph");

    // Verify no 404 error
    const error404 = page.locator("text=404, text=Not Found").first();
    const has404 = await error404.isVisible().catch(() => false);
    expect(has404).toBe(false);
  });
});
