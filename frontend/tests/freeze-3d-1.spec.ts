import { test, expect } from "@playwright/test";

const FRONTEND_URL = process.env.FRONTEND_URL || "http://127.0.0.1:3002";

/**
 * FREEZE-3D-1 (решение 82): frontend.graph.3d.enabled=false — кнопки «3D» нет
 * в переключателе видов, /graph/3d открывает 2D-граф. Живой снимок верхней
 * панели — критерий приёмки.
 */
test.describe("FREEZE-3D-1", () => {
  test("top bar shows no 3D toggle and /graph/3d serves 2D", async ({ page }) => {
    await page.goto(`${FRONTEND_URL}/`, { timeout: 60000, waitUntil: "networkidle" });
    await expect(page.locator(".splash-screen")).toBeHidden({ timeout: 30000 });

    await expect(page.locator('[data-testid="view-toggle-graph"]')).toBeVisible();
    await expect(page.locator('[data-testid="view-toggle-list"]')).toBeVisible();
    await expect(page.locator('[data-testid="view-toggle-3d"]')).toHaveCount(0);

    // The 3D route must land on the 2D graph.
    await page.goto(`${FRONTEND_URL}/graph/3d`, {
      timeout: 60000,
      waitUntil: "networkidle",
    });
    await expect(page).toHaveURL(/\/graph$/, { timeout: 20000 });
    await expect(page.locator('[data-testid="graph-canvas"]')).toBeVisible({
      timeout: 30000,
    });
    await expect(page.locator('[data-testid="view-toggle-3d"]')).toHaveCount(0);

    const topBar = page.locator('[data-testid="view-toggle-graph"]').locator("..").locator("..");
    await topBar.screenshot({
      path: "../docs/agents/screenshots/freeze-3d-1/top-bar.png",
    });
  });
});
