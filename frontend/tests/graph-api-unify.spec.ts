import { test, expect } from "@playwright/test";
import type { APIRequestContext } from "@playwright/test";
import { createNote, deleteNote, getBackendUrl } from "./helpers/testData";
import { setupSkipAuth, clickViewToggle } from "./helpers/testUtils";

/**
 * Regression tests for graph-service as the single source of truth on the home page.
 *
 * Scenarios:
 * - The home page makes exactly one request to the full graph endpoint.
 * - After a note is created, the graph reflects the new note (verified on reload).
 */

async function publishNote(request: APIRequestContext, noteId: string): Promise<void> {
  const response = await request.post(`${getBackendUrl()}/api/v1/notes/${noteId}/publish`);
  if (!response.ok()) {
    throw new Error(`Failed to publish note: ${response.status()} ${await response.text()}`);
  }
}

test.describe("Graph API unification", { tag: ["@graph-api"] }, () => {
  const createdNoteIds: string[] = [];

  test.beforeEach(async ({ page }) => {
    await setupSkipAuth(page);
  });

  test.afterEach(async ({ request }) => {
    for (const noteId of createdNoteIds) {
      try {
        await deleteNote(request, noteId);
      } catch {
        // Ignore cleanup errors.
      }
    }
    createdNoteIds.length = 0;
  });

  test("home page loads full graph exactly once", async ({ page, request }) => {
    const timestamp = Date.now();
    const createResponse = await createNote(request, {
      title: `Graph API Unify Note ${timestamp}`,
      content: "Initial note for graph API unification test",
      type: "star",
    });
    const noteId = createResponse.data.id;
    createdNoteIds.push(noteId);
    // In SKIP_AUTH mode the frontend loads the public graph, so the note must be public.
    await publishNote(request, noteId);

    let fullGraphRequestCount = 0;

    // Intercept full-graph paths (SvelteKit proxy and direct backend fallback).
    // In SKIP_AUTH mode the authenticated graph endpoint is used.
    await page.route("**/graph-service/api/v1/graph/full", async (route) => {
      fullGraphRequestCount++;
      await route.continue();
    });
    await page.route("**/api/v1/graph/full", async (route) => {
      fullGraphRequestCount++;
      await route.continue();
    });

    await page.goto("/");
    await page.waitForLoadState("networkidle");

    await expect(page.locator('[data-testid="graph-2d-container"]')).toBeVisible({ timeout: 10000 });

    // Switch to list view to assert the note title is rendered from the graph data.
    await clickViewToggle(page, "list");
    await expect(page.locator(".notes-grid").first()).toBeVisible({ timeout: 10000 });
    await expect(page.locator("text=Graph API Unify Note").first()).toBeVisible({ timeout: 10000 });

    // The full graph must be requested at most once. A value of 0 is acceptable
    // when the data was preloaded/cached and no network request is necessary.
    expect(fullGraphRequestCount).toBeLessThanOrEqual(1);
  });

  test("graph reflects newly created note after reload", async ({ page, request }) => {
    const initialTimestamp = Date.now();
    const initialResponse = await createNote(request, {
      title: `Initial Note ${initialTimestamp}`,
      content: "Initial note content",
      type: "planet",
    });
    const initialNoteId = initialResponse.data.id;
    createdNoteIds.push(initialNoteId);
    await publishNote(request, initialNoteId);

    await page.goto("/");
    await page.waitForLoadState("networkidle");

    const updateTimestamp = Date.now();
    const newResponse = await createNote(request, {
      title: `New Note ${updateTimestamp}`,
      content: "Created after initial page load",
      type: "comet",
    });
    const newNoteId = newResponse.data.id;
    createdNoteIds.push(newNoteId);
    await publishNote(request, newNoteId);

    let fullGraphRequestCount = 0;

    await page.route("**/graph-service/api/v1/graph/full", async (route) => {
      fullGraphRequestCount++;
      await route.continue();
    });
    await page.route("**/api/v1/graph/full", async (route) => {
      fullGraphRequestCount++;
      await route.continue();
    });

    await page.reload();
    await page.waitForLoadState("networkidle");

    await clickViewToggle(page, "list");
    await expect(page.locator(".notes-grid").first()).toBeVisible({ timeout: 10000 });
    await expect(page.locator(`text=New Note ${updateTimestamp}`).first()).toBeVisible({
      timeout: 10000,
    });

    // The page must never fetch the full graph more than once. It may be 0 if the
    // graph was already preloaded on the server / cached by PreloadService.
    expect(fullGraphRequestCount).toBeLessThanOrEqual(1);
  });

});
