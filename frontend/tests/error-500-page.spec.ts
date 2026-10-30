import { test, expect } from "@playwright/test";

/**
 * Regression: the global 500 error page renders as a full-viewport screen
 * with the "disconnected extension cord" server-error illustration.
 */

test("500 error page covers the full viewport", async ({ page }) => {
  const viewport = page.viewportSize() ?? { width: 1280, height: 720 };
  await page.setViewportSize(viewport);

  await page.goto("/test/500?trigger=500");

  const errorPage = page.locator('[data-testid="error-page"]');
  await expect(errorPage).toBeVisible({ timeout: 15000 });

  const box = await errorPage.boundingBox();
  expect(box).not.toBeNull();
  expect(box!.width).toBe(viewport.width);
  expect(box!.height).toBe(viewport.height);

  await expect(page.getByText("Internal Server Error")).toBeVisible();
  await expect(page.getByText("Refresh page")).toBeVisible();
  await expect(page.getByText("Go home")).toBeVisible();

  const illustration = page.locator('[data-testid="state-illustration-server-error"]');
  await expect(illustration).toBeVisible();
});
