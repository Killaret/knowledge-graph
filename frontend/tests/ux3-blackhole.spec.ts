import { test, expect, type Page, type Locator } from "@playwright/test";

const FRONTEND_URL = process.env.FRONTEND_URL || "http://127.0.0.1:3002";

/** Wait until the 2D canvas exposed its debug API and the simulation
 *  assigned real coordinates to at least one node. */
async function waitForGraphCanvas(page: Page): Promise<Locator> {
  const canvas = page.locator('[data-testid="graph-canvas"]');
  await expect(canvas).toBeVisible({ timeout: 20000 });
  await page.waitForFunction(
    () => {
      const api = (window as any).__graphCanvas;
      if (!api) return false;
      const nodes = api.getSimulationNodes();
      return nodes.length > 0 && nodes.some((n: { x?: number }) => n.x != null);
    },
    { timeout: 20000 }
  );
  return canvas;
}

/**
 * UX-3 review follow-up: with the link-type legend expanded the black hole
 * must stay fully visible (the accretion ring draws at 1.44× the core radius)
 * and still accept a dropped note. Mirrors the live check requested in
 * docs/tasks/UX-3-review-findings.md.
 */
test.describe("UX-3 black hole beside the expanded legend", () => {
  test("hole glyph clears the expanded legend and accepts a note drop", async ({ page }) => {
    await page.goto(`${FRONTEND_URL}/`, { timeout: 60000, waitUntil: "networkidle" });
    // The splash screen overlays the canvas during boot — wait for it to
    // leave so screenshots show the real graph.
    await expect(page.locator(".splash-screen")).toBeHidden({ timeout: 30000 });
    const canvas = await waitForGraphCanvas(page);
    const canvasBox = await canvas.boundingBox();
    expect(canvasBox).not.toBeNull();

    const legend = page.locator(".link-type-legend");
    await expect(legend).toBeVisible({ timeout: 10000 });
    if (await legend.evaluate((el) => el.classList.contains("collapsed"))) {
      await legend.locator("button").first().click();
      await expect(legend).not.toHaveClass(/collapsed/);
    }
    // Let the ResizeObserver publish the measured size and a frame run.
    await page.waitForTimeout(400);

    const hole = await page.evaluate(() => window.__graphCanvas?.getBlackHole?.());
    expect(hole).toBeTruthy();
    const legendBox = await legend.boundingBox();
    expect(legendBox).not.toBeNull();
    const extent = hole!.radius * 1.44;

    // The whole glyph — core plus accretion ring — is inside the canvas and
    // completely to the left of the legend.
    expect(hole!.x - extent).toBeGreaterThanOrEqual(0);
    expect(hole!.y - extent).toBeGreaterThanOrEqual(0);
    expect(hole!.x + extent).toBeLessThanOrEqual(canvasBox!.width);
    expect(hole!.y + extent).toBeLessThanOrEqual(canvasBox!.height);
    expect(canvasBox!.x + hole!.x + extent).toBeLessThanOrEqual(legendBox!.x + 1);

    await page.screenshot({ path: "test-results/ux3-blackhole-legend.png" });

    // Drag a real note onto the hole: it must accept the drop.
    const api = await page.evaluate(() => {
      const gc = window.__graphCanvas!;
      const node = gc
        .getSimulationNodes()
        .find((n: any) => n.x != null && n.y != null && n.type !== "technical") as any;
      return {
        transform: { ...gc.transform },
        node: { id: node.id, x: node.x, y: node.y },
        count: gc.getSimulationNodes().length,
      };
    });
    const nodeScreenX = canvasBox!.x + api.node.x * api.transform.k + api.transform.x;
    const nodeScreenY = canvasBox!.y + api.node.y * api.transform.k + api.transform.y;
    const holeScreenX = canvasBox!.x + hole!.x;
    const holeScreenY = canvasBox!.y + hole!.y;

    await page.mouse.move(nodeScreenX, nodeScreenY);
    await page.mouse.down();
    await page.mouse.move(holeScreenX, holeScreenY, { steps: 15 });
    await page.mouse.up();

    // The drop opens the delete confirmation — the hole accepted the note.
    const confirm = page.locator('[data-testid="confirm-modal-confirm"]');
    await expect(confirm).toBeVisible({ timeout: 5000 });
    await page.screenshot({ path: "test-results/ux3-blackhole-drop.png" });
    await confirm.click();

    await page.waitForFunction(
      (n) => window.__graphCanvas!.getSimulationNodes().length < n,
      api.count,
      { timeout: 15000 }
    );
  });
});
