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

    const input = page.locator("textarea#import-list");
    await input.fill("First test | https://example.com/one\nhttps://example.com/two");

    await page.getByRole("button", { name: /Preview/i }).click();

    await expect(page.locator(".preview-table tbody tr")).toHaveCount(2, { timeout: 10000 });

    await page.getByRole("button", { name: /Import/i }).click();

    await expect(page.locator(".progress")).toBeVisible({ timeout: 10000 });

    // Wait for the async task to complete (poll up to 15s)
    await expect(async () => {
      const text = await page.locator(".progress").textContent();
      expect(text).toMatch(/Created/);
    }).toPass({ timeout: 15000 });

    // Verify the first note exists in the backend
    const listResponse = await request.get(`${getBackendUrl()}/api/v1/notes`);
    expect(listResponse.ok()).toBeTruthy();
    const notesData = await listResponse.json();
    const notes = notesData.notes ?? [];
    const created = notes.find((n: { title: string }) => n.title === "First test");
    expect(created).toBeDefined();
    expect(created.content).toContain("https://example.com/one");
  });
});
