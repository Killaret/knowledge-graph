import { test, expect } from "@playwright/test";
import { loginAsTestUser } from "./helpers/auth";

test.describe("Child note creation @auth-real", () => {
  test("creates a child note from the note details panel", async ({ page, request }) => {
    const parentTitle = `Parent Note ${Date.now()}`;
    const childTitle = `Child Note ${Date.now()}`;
    const childContent = "Created as a child note";

    await loginAsTestUser(page, request);
    await page.goto("/?nocache=1", { waitUntil: "networkidle" });

    // Create a parent note.
    await page.locator('[data-testid="create-note-button"]').click();
    await page.locator('[data-testid="create-note-title"]').fill(parentTitle);
    await page.locator('[data-testid="create-note-content"]').fill("Parent content");
    await page.locator('[data-testid="create-note-submit"]').click();

    await expect(page.locator('[data-testid="create-note-title"]')).toHaveCount(0, {
      timeout: 10000,
    });

    // Switch to list view and select the parent note.
    await page.locator('[data-testid="view-toggle-list"]').click();
    const parentCard = page.locator('[data-testid="note-card"]').filter({ hasText: parentTitle });
    await parentCard.click();

    // Open the child note creation flow from the right panel.
    await expect(page.locator('[data-testid="note-details-create-child"]')).toBeVisible({
      timeout: 10000,
    });
    await page.locator('[data-testid="note-details-create-child"]').click();

    // Verify the modal shows the parent breadcrumb.
    await expect(page.locator('[data-testid="create-note-parent"]')).toContainText(parentTitle, {
      timeout: 5000,
    });

    // Fill in the child note and submit.
    await page.locator('[data-testid="create-note-title"]').fill(childTitle);
    await page.locator('[data-testid="create-note-content"]').fill(childContent);
    await page.locator('[data-testid="create-note-submit"]').click();

    await expect(page.locator('[data-testid="create-note-title"]')).toHaveCount(0, {
      timeout: 10000,
    });

    // Verify the child note appears in the list view.
    await expect(
      page.locator('[data-testid="note-title"]').filter({ hasText: childTitle })
    ).toBeVisible({ timeout: 20000 });
  });
});
