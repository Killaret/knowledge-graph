import { test, expect } from "@playwright/test";
import { loginAsTestUser } from "./helpers/auth";

/**
 * Regression tests for the unified GraphTopBar and cockpit/canvas controls.
 *  - public graph top bar has fog toggle, view toggles and zoom works
 *  - authenticated cockpit panels open/pin/close
 *  - zoom changes the canvas transform (covers black-hole / ghost-node scaling path)
 */

test.describe("Cockpit and canvas controls @auth-real", () => {
  const FRONTEND_URL = process.env.FRONTEND_URL || "http://127.0.0.1:3002";

  test("public graph top bar exposes canvas controls and fog toggle", async ({ page }) => {
    await page.goto(`${FRONTEND_URL}/graph?full=1&nocache=1`, {
      timeout: 60000,
      waitUntil: "networkidle",
    });

    const canvas = page.locator('[data-testid="graph-canvas"]');
    await expect(canvas).toBeVisible({ timeout: 20000 });

    const topBar = page.locator('[data-testid="graph-top-bar"]');
    await expect(topBar).toBeVisible();

    // Core public controls
    await expect(page.locator('[data-testid="top-bar-reset"]')).toBeVisible();
    await expect(page.locator('[data-testid="top-bar-open-search"]')).toBeVisible();
    await expect(page.locator('[data-testid="top-bar-focus"]')).toBeVisible();

    // Fog toggle is now in the unified top bar, not the overlay
    const fogBtn = page.locator('[data-testid="top-bar-fog"]');
    await expect(fogBtn).toBeVisible();
    await expect(fogBtn).toHaveAttribute("aria-pressed", "true");
    await fogBtn.click();
    await expect(fogBtn).toHaveAttribute("aria-pressed", "false");
    await fogBtn.click();
    await expect(fogBtn).toHaveAttribute("aria-pressed", "true");

    // Graph stats are visible and show nodes
    const stats = page.locator('[data-testid="graph-stats"]').first();
    await expect(stats).toBeVisible();
    await expect(stats).toContainText(/[1-9]\d*\s*(?:nodes?|уз(?:лов|ел|ла|ьев)?)/i, {
      timeout: 10000,
    });

    // Auth buttons are visible on the public graph
    await expect(page.locator('[data-testid="top-bar-sign-in"]')).toBeVisible();
    await expect(page.locator('[data-testid="top-bar-register"]')).toBeVisible();
  });

  test("canvas zoom changes transform and keeps the canvas visible", async ({ page }) => {
    await page.goto(`${FRONTEND_URL}/graph?full=1&nocache=1`, {
      timeout: 60000,
      waitUntil: "networkidle",
    });

    const canvas = page.locator('[data-testid="graph-canvas"]');
    await expect(canvas).toBeVisible({ timeout: 20000 });

    // Wait for the debug / controller handle exposed by GraphCanvas
    await page.waitForFunction(() => typeof (window as any).__graphCanvas !== "undefined", {
      timeout: 10000,
    });

    const initialK = await page.evaluate(() => (window as any).__graphCanvas.transform.k);
    expect(initialK).toBeGreaterThan(0);

    const box = await canvas.boundingBox();
    expect(box).toBeTruthy();

    // Zoom in at canvas center
    await page.mouse.move(box!.x + box!.width / 2, box!.y + box!.height / 2);
    await page.mouse.wheel(0, -10);
    await page.waitForTimeout(500);

    const zoomedK = await page.evaluate(() => (window as any).__graphCanvas.transform.k);
    expect(zoomedK).toBeGreaterThan(initialK);

    // Zoom back out
    await page.mouse.wheel(0, 40);
    await page.waitForTimeout(500);

    const zoomedOutK = await page.evaluate(() => (window as any).__graphCanvas.transform.k);
    expect(zoomedOutK).toBeLessThan(zoomedK);

    // The graph should still be rendered after zoom interactions
    await expect(canvas).toBeVisible();
  });

  test("authenticated cockpit panels open, pin and close", async ({ page, request }) => {
    await loginAsTestUser(page, request);

    await page.goto(`${FRONTEND_URL}/graph?full=1&nocache=1`, {
      timeout: 60000,
      waitUntil: "networkidle",
    });

    const canvas = page.locator('[data-testid="graph-canvas"]');
    await expect(canvas).toBeVisible({ timeout: 20000 });

    await expect(page.locator('[data-testid="cosmic-cockpit"]')).toBeVisible();
    await expect(page.locator('[data-testid="graph-top-bar"]')).toBeVisible();

    // Open left panel. The handle is removed from the DOM as soon as it is
    // clicked (it is only rendered while the panel is closed), so we use
    // evaluate to click once without waiting on a detached Playwright handle.
    const leftHandle = page.locator('[data-testid="cockpit-handle-left"]');
    await expect(leftHandle).toBeVisible();
    await page.evaluate(() => {
      document.querySelector<HTMLElement>('[data-testid="cockpit-handle-left"]')?.click();
    });
    const leftPanel = page.locator('[data-testid="cockpit-left-panel"]');
    await expect(leftPanel).toBeInViewport({
      timeout: 5000,
    });
    await expect(leftHandle).not.toBeVisible();

    // Pin, unpin, then close
    const pinBtn = page.locator('[data-testid="cockpit-panel-pin-left"]');
    await expect(pinBtn).toBeVisible();
    await pinBtn.click();
    await pinBtn.click(); // unpin
    await page.locator('[data-testid="cockpit-panel-close-left"]').click();

    // Move the cursor away from the panel so mouseleave triggers and the panel
    // can close (it stays open while hovered when not pinned).
    await page.locator('[data-testid="graph-canvas"]').hover();
    await page.waitForTimeout(400);

    // The handle re-appears once the panel is fully closed.
    await expect(leftHandle).toBeVisible();

    // Open right panel (empty by default until a node is selected)
    const rightHandle = page.locator('[data-testid="cockpit-handle-right"]');
    await expect(rightHandle).toBeVisible();
    await page.evaluate(() => {
      document.querySelector<HTMLElement>('[data-testid="cockpit-handle-right"]')?.click();
    });
    const rightPanel = page.locator('[data-testid="cockpit-right-panel"]');
    await expect(rightPanel).toBeInViewport({
      timeout: 5000,
    });
    await expect(rightHandle).not.toBeVisible();

    // Top bar view controls are present (unified with public top bar)
    await expect(page.locator('[data-testid="view-toggle-graph"]')).toBeVisible();
    await expect(page.locator('[data-testid="view-toggle-3d"]')).toBeVisible();
    await expect(page.locator('[data-testid="view-toggle-list"]')).toBeVisible();

    // Authenticated only: create note button, fog toggle, link/type filters
    await expect(page.locator('[data-testid="create-note-button"]')).toBeVisible();
    await expect(page.locator('[data-testid="top-bar-fog"]')).toBeVisible();
  });
});
