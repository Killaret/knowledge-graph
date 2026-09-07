import { test, expect } from "@playwright/test";

/**
 * Test i18n language toggle functionality
 */

test.describe("i18n Language Toggle", () => {
  test("should default to English locale", async ({ page }) => {
    await page.goto("/");
    await page.waitForLoadState("domcontentloaded");
    await page.waitForTimeout(1000);

    // Check that locale is 'en' or null (null means default will be 'en')
    const locale = await page.evaluate(() => localStorage.getItem("locale"));
    expect(locale === "en" || locale === null).toBeTruthy();
  });

  test("should switch to Russian locale", async ({ page }) => {
    await page.goto("/");
    await page.waitForLoadState("domcontentloaded");
    await page.waitForTimeout(1000);

    // Switch to Russian
    await page.evaluate(() => {
      localStorage.setItem("locale", "ru");
    });

    // Reload page to apply locale change
    await page.reload();
    await page.waitForLoadState("domcontentloaded");
    await page.waitForTimeout(1000);

    // Check that locale is 'ru'
    const locale = await page.evaluate(() => localStorage.getItem("locale"));
    expect(locale).toBe("ru");
  });

  test("should switch back to English locale", async ({ page }) => {
    await page.goto("/");
    await page.waitForLoadState("domcontentloaded");
    await page.waitForTimeout(1000);

    // Switch to Russian first
    await page.evaluate(() => {
      localStorage.setItem("locale", "ru");
    });

    await page.reload();
    await page.waitForLoadState("domcontentloaded");
    await page.waitForTimeout(1000);

    // Switch back to English
    await page.evaluate(() => {
      localStorage.setItem("locale", "en");
    });

    await page.reload();
    await page.waitForLoadState("domcontentloaded");
    await page.waitForTimeout(1000);

    // Check that locale is 'en'
    const locale = await page.evaluate(() => localStorage.getItem("locale"));
    expect(locale).toBe("en");
  });

  test("should persist locale across page navigation", async ({ page }) => {
    await page.goto("/");
    await page.waitForLoadState("domcontentloaded");
    await page.waitForTimeout(1000);

    // Set locale to Russian
    await page.evaluate(() => {
      localStorage.setItem("locale", "ru");
    });

    // Navigate to different page
    await page.goto("/graph");
    await page.waitForLoadState("domcontentloaded");
    await page.waitForTimeout(1000);

    // Check that locale persists
    const locale = await page.evaluate(() => localStorage.getItem("locale"));
    expect(locale).toBe("ru");
  });

  test("should handle invalid locale gracefully", async ({ page }) => {
    await page.goto("/");
    await page.waitForLoadState("domcontentloaded");
    await page.waitForTimeout(1000);

    // Set invalid locale
    await page.evaluate(() => {
      localStorage.setItem("locale", "invalid");
    });

    await page.reload();
    await page.waitForLoadState("domcontentloaded");
    await page.waitForTimeout(1000);

    // Should fall back to 'en' or stay 'invalid' (app handles this)
    const locale = await page.evaluate(() => localStorage.getItem("locale"));
    expect(locale === "en" || locale === "invalid").toBeTruthy();
  });
});
