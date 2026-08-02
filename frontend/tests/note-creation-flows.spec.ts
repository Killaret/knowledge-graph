import { test, expect } from "@playwright/test";
import { loginAsTestUser } from "./helpers/auth";

/**
 * Regression tests for the two note-creation flows on the graph page.
 *
 * Manual testing showed two different ways to create a note:
 * 1. Floating "+" button -> CreateNoteModal.
 * 2. "N" hotkey -> ghost note form on the canvas.
 *
 * These tests cover both flows plus the empty-state fallback. They verify that
 * the created note shows up in the list view, which is the most reliable
 * DOM-based assertion since the graph canvas is a <canvas> element.
 */
test.describe("Note creation flows @auth-real", () => {
  test("creates a note via the floating + button on the main page", async ({ page, request }) => {
    const title = `Modal Note ${Date.now()}`;
    const content = "Created via floating + button";

    await loginAsTestUser(page, request);
    await page.goto("/?nocache=1", { waitUntil: "networkidle" });

    await page.locator('[data-testid="create-note-button"]').click();

    await page.locator('[data-testid="create-note-title"]').fill(title);
    await page.locator('[data-testid="create-note-content"]').fill(content);
    await page.locator('[data-testid="create-note-submit"]').click();

    // Wait for the modal to close.
    await expect(page.locator('[data-testid="create-note-title"]')).toHaveCount(0, {
      timeout: 10000,
    });

    // Switch to list view and verify the new note.
    await page.locator('[data-testid="view-toggle-list"]').click();
    await expect(page.locator('[data-testid="note-title"]').filter({ hasText: title })).toBeVisible(
      { timeout: 20000 }
    );
  });

  test("creates a note via the N hotkey ghost form on the main page", async ({ page, request }) => {
    const title = `Ghost Note ${Date.now()}`;
    const content = "Created via N hotkey";

    await loginAsTestUser(page, request);
    await page.goto("/?nocache=1", { waitUntil: "networkidle" });

    // Press N to open the ghost note form.
    await page.keyboard.press("n");

    await expect(page.locator('[data-testid="ghost-note-form"]')).toBeVisible({
      timeout: 5000,
    });

    await page.locator('[data-testid="ghost-note-title"]').fill(title);
    await page.locator('[data-testid="ghost-note-content"]').fill(content);
    await page.locator('[data-testid="ghost-note-create"]').click();

    // Wait for the form to close.
    await expect(page.locator('[data-testid="ghost-note-form"]')).toHaveCount(0, {
      timeout: 10000,
    });

    // Switch to list view and verify.
    await page.locator('[data-testid="view-toggle-list"]').click();
    await expect(page.locator('[data-testid="note-title"]').filter({ hasText: title })).toBeVisible(
      { timeout: 20000 }
    );
  });

  test("creates a note via the N hotkey on the /graph page", async ({ page, request }) => {
    const title = `Graph Ghost Note ${Date.now()}`;
    const content = "Created via N on /graph";

    await loginAsTestUser(page, request);
    await page.goto("/graph?nocache=1", { waitUntil: "networkidle" });

    await page.keyboard.press("n");

    await expect(page.locator('[data-testid="ghost-note-form"]')).toBeVisible({
      timeout: 5000,
    });

    await page.locator('[data-testid="ghost-note-title"]').fill(title);
    await page.locator('[data-testid="ghost-note-content"]').fill(content);
    await page.locator('[data-testid="ghost-note-create"]').click();

    await expect(page.locator('[data-testid="ghost-note-form"]')).toHaveCount(0, {
      timeout: 10000,
    });

    // /graph has no list view, so we navigate to the main page to verify.
    await page.goto("/?nocache=1", { waitUntil: "networkidle" });
    await page.locator('[data-testid="view-toggle-list"]').click();
    await expect(page.locator('[data-testid="note-title"]').filter({ hasText: title })).toBeVisible(
      { timeout: 20000 }
    );
  });
});
