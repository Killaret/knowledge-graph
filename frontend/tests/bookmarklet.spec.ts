import { test, expect } from "@playwright/test";
import { setupSkipAuth, getBackendUrl } from "./helpers/testUtils";

/**
 * E2E tests for the bookmarklet import flow.
 */

test.describe("Bookmarklet import", { tag: ["@e2e", "@import", "@skip-auth"] }, () => {
  test.beforeEach(async ({ page }) => {
    await setupSkipAuth(page);
  });

  test("creates a note from query parameters", async ({ page, request }) => {
    const title = "Bookmarklet Test Page";
    const url = "https://example.com/test";
    const text = "Important captured text";

    await page.goto(`/import?skip_auth=true&title=${encodeURIComponent(title)}&url=${encodeURIComponent(url)}&text=${encodeURIComponent(text)}`);
    await page.waitForLoadState("networkidle");

    // The page should show the success message
    await expect(page.locator(".status.success")).toBeVisible({ timeout: 10000 });
    await expect(page.locator(".status.success")).toContainText(title);
    await expect(page.locator(".status.success")).toContainText("created successfully");

    // The note should appear in the backend
    const listResponse = await request.get(`${getBackendUrl()}/api/v1/notes`);
    expect(listResponse.ok()).toBeTruthy();
    const notesData = await listResponse.json();
    const notes = notesData.data ?? notesData.notes ?? [];
    const created = notes.find((n: { title: string }) => n.title === title);
    expect(created).toBeDefined();
    expect(created.content).toContain(url);
    expect(created.content).toContain(text);
    expect(created.type).toBe("asteroid");
  });

});
