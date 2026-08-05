import { test, expect } from "@playwright/test";
import { setupSkipAuth, getBackendUrl } from "./helpers/testUtils";

/**
 * E2E tests for the mass bookmark import flow.
 */

test.describe("Mass bookmark import", { tag: ["@e2e", "@import", "@skip-auth"] }, () => {
  test.beforeEach(async ({ page }) => {
    await setupSkipAuth(page);
  });

  test("previews and imports a batch of URLs", async ({ page, request }) => {
    await page.goto("/import/bookmarks");
    await page.waitForLoadState("networkidle");

    const ts = Date.now();
    const title1 = `Mass import note ${ts}`;
    const title2 = `Mass import note ${ts} second`;
    const input = page.locator("textarea#import-list");
    await input.fill(`${title1} | https://example.com/mass/${ts}\n${title2} | https://example.com/mass/${ts}/2`);

    await page.getByRole("button", { name: /Preview/i }).click();

    await expect(page.locator(".preview-table tbody tr")).toHaveCount(2, { timeout: 10000 });

    await page.getByRole("button", { name: /Import/i }).click();

    await expect(page.locator(".progress")).toBeVisible({ timeout: 10000 });

    // Wait for the async task to create at least one note (poll status up to 30s)
    const getProgressText = async () => (await page.locator(".progress").textContent()) ?? "";
    await expect(async () => {
      const text = await getProgressText();
      expect(text).toMatch(/Created:\s*[1-9]/);
    }).toPass({ timeout: 30000 });

    // Give the worker a moment to persist the notes after the status update
    await page.waitForTimeout(1000);

    // Verify the first note exists in the backend
    const listResponse = await request.get(`${getBackendUrl()}/api/v1/notes`);
    expect(listResponse.ok()).toBeTruthy();
    const notesData = await listResponse.json();
    const notes = notesData.notes ?? [];
    const created = notes.find((n: { title: string }) => n.title === title1);
    expect(created).toBeDefined();
    expect(created.content).toContain(`https://example.com/mass/${ts}`);
  });
});
