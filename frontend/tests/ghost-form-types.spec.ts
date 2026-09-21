import { test, expect } from "@playwright/test";
import { setupSkipAuth } from "./helpers/testUtils";

/**
 * IMP-3: the ghost note form (N hotkey) lists all 11 UI types in a column;
 * at a 720 px viewport the form overflows and the bottom types (satellite,
 * asteroid, dust, debris) are unreachable — measured in review. The form must
 * scroll so every type is clickable.
 */
test.describe("Ghost note form — all types reachable (IMP-3)", () => {
  test.use({ viewport: { width: 1280, height: 720 } });

  test("debris is clickable at 720px viewport", async ({ page }) => {
    await setupSkipAuth(page);
    await page.goto("/graph?full=1&nocache=1", { waitUntil: "domcontentloaded" });
    await page.waitForLoadState("networkidle", { timeout: 30000 }).catch(() => {});

    await page.keyboard.press("n");
    const form = page.locator('[data-testid="ghost-note-form"]');
    await expect(form).toBeVisible({ timeout: 10000 });

    // All 11 UI types are rendered.
    await expect(form.locator("[data-type]")).toHaveCount(11);

    // The form must not overflow the viewport height — measured, not eyeballed.
    const box = await form.boundingBox();
    const viewport = page.viewportSize();
    expect(box).not.toBeNull();
    expect(box!.y + box!.height).toBeLessThanOrEqual(viewport!.height);

    // The last type must be reachable: scroll the form and click it.
    const debris = form.locator('[data-type="debris"]');
    await debris.scrollIntoViewIfNeeded();
    await expect(debris).toBeVisible();
    await debris.click();
    await expect(debris).toHaveClass(/active/);
  });
});
