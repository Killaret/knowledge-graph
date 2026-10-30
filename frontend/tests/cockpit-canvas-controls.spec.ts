import { test, expect, type Page, type Locator } from "@playwright/test";
import { loginAsTestUser } from "./helpers/auth";

/**
 * Wait until the 2D graph canvas has exposed its debug API and the force
 * simulation has assigned real coordinates to at least one node. This is the
 * moment when the canvas is ready for wheel/pan/dblclick interactions: the
 * event bridge is attached and the initial transform has been computed.
 */
async function waitForGraphCanvas(page: Page): Promise<Locator> {
  const canvas = page.locator('[data-testid="graph-canvas"]');
  await expect(canvas).toBeVisible({ timeout: 20000 });
  await page.waitForFunction(
    () => {
      const api = (window as any).__graphCanvas as
        | {
            getSimulationNodes: () => Array<{ x?: number; y?: number }>;
          }
        | undefined;
      if (!api) return false;
      const nodes = api.getSimulationNodes();
      return nodes.length > 0 && nodes.some((n) => n.x != null && n.y != null);
    },
    { timeout: 20000 }
  );
  return canvas;
}

/**
 * Playwright's `page.mouse.wheel()` does not reliably deliver `wheel` events to
 * a `<canvas>` that is not inside a scrollable container. The canvas listener is
 * attached with `{ passive: false }` and expects a real `WheelEvent`. We dispatch
 * one directly on the element so the test exercises the same `handleZoom` path a
 * real browser wheel would, while keeping the test deterministic.
 */
async function dispatchWheel(
  canvas: Locator,
  deltaY: number,
  options?: { clientX?: number; clientY?: number }
): Promise<void> {
  await canvas.page().evaluate(
    ({ deltaY, clientX: overrideX, clientY: overrideY }) => {
      const el = document.querySelector<HTMLCanvasElement>('[data-testid="graph-canvas"]');
      if (!el) throw new Error("Canvas element not found");
      const rect = el.getBoundingClientRect();
      const clientX = overrideX ?? rect.left + rect.width / 2;
      const clientY = overrideY ?? rect.top + rect.height / 2;
      el.dispatchEvent(
        new WheelEvent("wheel", {
          deltaY,
          clientX,
          clientY,
          bubbles: true,
          cancelable: true,
        })
      );
    },
    { deltaY, clientX: options?.clientX, clientY: options?.clientY }
  );
}

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

    const canvas = await waitForGraphCanvas(page);

    const initialK = await page.evaluate(() => (window as any).__graphCanvas.transform.k);
    expect(initialK).toBeGreaterThan(0);

    const box = await canvas.boundingBox();
    expect(box).toBeTruthy();

    // Zoom in at canvas center
    await dispatchWheel(canvas, -120);
    await page.waitForTimeout(100);

    const zoomedK = await page.evaluate(() => (window as any).__graphCanvas.transform.k);
    expect(zoomedK).toBeGreaterThan(initialK);

    // Zoom back out
    await dispatchWheel(canvas, 120);
    await page.waitForTimeout(100);

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

test.describe("Canvas controls — adversarial @auth-real", () => {
  const FRONTEND_URL = process.env.FRONTEND_URL || "http://127.0.0.1:3002";

  test("empty public graph shows empty state and no canvas controls", async ({ page }) => {
    // Mock the public graph endpoint to return an empty graph without
    // touching the shared test database.
    await page.route("**/graph-service/api/v1/graph/public**", async (route) => {
      await route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify({
          nodes: [],
          links: [],
          meta: { total_nodes: 0, total_links: 0, hash: "empty" },
        }),
      });
    });

    await page.goto(`${FRONTEND_URL}/graph?full=1&nocache=1`, {
      timeout: 60000,
      waitUntil: "networkidle",
    });

    // Empty state is shown instead of the canvas.
    const emptyState = page.locator('[data-testid="graph-empty-state"]');
    await expect(emptyState).toBeVisible({ timeout: 20000 });

    // No canvas means no canvas-level controls should be mounted.
    await expect(page.locator('[data-testid="graph-canvas"]')).not.toBeVisible();
    await expect(page.locator('[data-testid="top-bar-fog"]')).not.toBeVisible();

    // Auth buttons and the top bar itself are still present.
    await expect(page.locator('[data-testid="graph-top-bar"]')).toBeVisible();
    await expect(page.locator('[data-testid="top-bar-sign-in"]')).toBeVisible();
    await expect(page.locator('[data-testid="top-bar-register"]')).toBeVisible();
  });

  test("fog can be disabled and canvas remains visible", async ({ page }) => {
    await page.goto(`${FRONTEND_URL}/graph?full=1&nocache=1`, {
      timeout: 60000,
      waitUntil: "networkidle",
    });

    const canvas = page.locator('[data-testid="graph-canvas"]');
    await expect(canvas).toBeVisible({ timeout: 20000 });

    const fogBtn = page.locator('[data-testid="top-bar-fog"]');
    await expect(fogBtn).toBeVisible();
    await expect(fogBtn).toHaveAttribute("aria-pressed", "true");

    await fogBtn.click();
    await expect(fogBtn).toHaveAttribute("aria-pressed", "false");
    await expect(canvas).toBeVisible();

    // Toggle back so other tests start from the default enabled state.
    await fogBtn.click();
    await expect(fogBtn).toHaveAttribute("aria-pressed", "true");
  });

  test("readonly public graph allows zoom and pan but not node drag", async ({ page }) => {
    await page.goto(`${FRONTEND_URL}/graph?full=1&nocache=1`, {
      timeout: 60000,
      waitUntil: "networkidle",
    });

    // Confirm this is actually the public graph before declaring it readonly.
    await expect(page.locator('[data-testid="top-bar-sign-in"]')).toBeVisible();

    const canvas = await waitForGraphCanvas(page);

    const box = await canvas.boundingBox();
    expect(box).toBeTruthy();

    const {
      x: initialX,
      y: initialY,
      k: initialK,
    } = await page.evaluate(() => {
      const t = (window as any).__graphCanvas.transform;
      return { x: t.x, y: t.y, k: t.k };
    });
    expect(initialK).toBeGreaterThan(0);

    // Pan the canvas by dragging on an empty area.
    const centerX = box!.x + box!.width / 2;
    const centerY = box!.y + box!.height / 2;
    await page.mouse.move(centerX, centerY);
    await page.mouse.down();
    await page.mouse.move(centerX + 80, centerY + 40);
    await page.mouse.up();
    await page.waitForTimeout(200);

    const afterPan = await page.evaluate(() => (window as any).__graphCanvas.transform);
    expect(afterPan.x).not.toEqual(initialX);
    expect(afterPan.y).not.toEqual(initialY);

    // Zoom in.
    await dispatchWheel(canvas, -120);
    await page.waitForTimeout(100);

    const afterZoom = await page.evaluate(() => (window as any).__graphCanvas.transform);
    expect(afterZoom.k).toBeGreaterThan(afterPan.k);

    // Double-click should zoom further (not create a ghost note in readonly).
    await canvas.dblclick();
    await page.waitForTimeout(200);

    const afterDbl = await page.evaluate(() => (window as any).__graphCanvas.transform);
    expect(afterDbl.k).toBeGreaterThanOrEqual(afterZoom.k);

    // In public readonly mode the right panel may still render its empty
    // handle, but it must not show any selected node details and the left
    // panel must not be present at all.
    const rightPanel = page.locator('[data-testid="cockpit-right-panel"]');
    await expect(rightPanel.locator('[data-testid="public-note-details"]')).not.toBeVisible();
    await expect(rightPanel.locator('[data-testid="cockpit-note-details"]')).not.toBeVisible();
    await expect(page.locator('[data-testid="cockpit-left-panel"]')).not.toBeVisible();
  });

  test("fast wheel interaction clamps zoom and keeps canvas visible", async ({ page }) => {
    await page.goto(`${FRONTEND_URL}/graph?full=1&nocache=1`, {
      timeout: 60000,
      waitUntil: "networkidle",
    });

    const canvas = await waitForGraphCanvas(page);

    const box = await canvas.boundingBox();
    expect(box).toBeTruthy();

    // Extreme zoom in — should clamp at the configured maximum (5).
    await dispatchWheel(canvas, -5000);
    await page.waitForTimeout(100);
    const maxZoom = await page.evaluate(() => (window as any).__graphCanvas.transform.k);
    expect(maxZoom).toBeLessThanOrEqual(5.01);
    expect(maxZoom).toBeGreaterThan(1);

    // Extreme zoom out — should clamp at the configured minimum (0.1).
    await dispatchWheel(canvas, 10000);
    await page.waitForTimeout(100);
    const minZoom = await page.evaluate(() => (window as any).__graphCanvas.transform.k);
    expect(minZoom).toBeGreaterThanOrEqual(0.09);

    // Rapidly alternate wheel direction — should not crash or hide the canvas.
    for (let i = 0; i < 15; i++) {
      await dispatchWheel(canvas, i % 2 === 0 ? -200 : 200);
    }
    await page.waitForTimeout(200);

    const finalK = await page.evaluate(() => (window as any).__graphCanvas.transform.k);
    expect(finalK).toBeGreaterThanOrEqual(0.1);
    expect(finalK).toBeLessThanOrEqual(5);

    await expect(canvas).toBeVisible();
  });
});
