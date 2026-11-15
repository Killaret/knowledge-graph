import { test, expect } from "@playwright/test";

const FRONTEND_URL = process.env.FRONTEND_URL || "http://127.0.0.1:3002";

/**
 * NOTE-HEALTH-1 этап 0 (решение: строка «Качество» → «Обработка», только при
 * проблеме; декоративный HEALTH в нижней панели скрыт до настоящей метрики).
 * Живой снимок нижней панели — критерий приёмки.
 */
test.describe("NOTE-HEALTH-1 stage 0", () => {
  test("bottom HUD has no decorative HEALTH block, real metrics intact", async ({ page }) => {
    await page.goto(`${FRONTEND_URL}/`, { timeout: 60000, waitUntil: "networkidle" });
    await expect(page.locator(".splash-screen")).toBeHidden({ timeout: 30000 });

    const panel = page.locator('[data-testid="cockpit-bottom-panel"]');
    await expect(panel).toBeVisible({ timeout: 20000 });

    // Decorative HEALTH is gone; real HUD metrics still render.
    await expect(panel.locator('[data-testid="hud-health"]')).toHaveCount(0);
    await expect(panel.locator('[data-testid="hud-node-count"]')).toBeVisible();
    await expect(panel.locator('[data-testid="hud-link-count"]')).toBeVisible();

    await panel.screenshot({
      path: "../docs/agents/screenshots/note-health-1/bottom-panel.png",
    });
    await page.screenshot({
      path: "../docs/agents/screenshots/note-health-1/page.png",
      fullPage: false,
    });
  });
});
