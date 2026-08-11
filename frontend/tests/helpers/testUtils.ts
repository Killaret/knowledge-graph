import { expect, type Page, type APIRequestContext } from "@playwright/test";
import {
  createNote as createNoteAdvanced,
  createLink as createLinkAdvanced,
  getBackendUrl,
} from "./testData";

/**
 * Get frontend URL from environment or use default
 */
export function getFrontendUrl(): string {
  return process.env.FRONTEND_URL || "http://localhost:5173";
}

/**
 * Re-export from testData.ts for backward compatibility
 */
export { getBackendUrl };

/**
 * Re-export createNote from testData.ts with simplified signature
 */
export async function createNote(
  request: APIRequestContext,
  data: { title: string; content?: string; type?: string; email?: string }
): Promise<{ data: { id: string; title: string } }> {
  const result = await createNoteAdvanced(request, {
    title: data.title,
    content: data.content,
    type: data.type,
    metadata: data.email ? { email: data.email } : undefined,
  });
  return { data: { id: result.data.id, title: result.data.title } };
}

/**
 * Re-export createLink from testData.ts
 */
export async function createLink(
  request: APIRequestContext,
  data: {
    source_note_id: string;
    target_note_id: string;
    link_type?: string;
    weight?: number;
  }
): Promise<void> {
  await createLinkAdvanced(
    request,
    data.source_note_id,
    data.target_note_id,
    data.weight ?? 0.5,
    data.link_type ?? "related"
  );
}

/**
 * Re-export deleteNote from testData.ts
 */
export { deleteNote as deleteNoteSimple } from "./testData";

/**
 * Re-export isBackendAvailable from testData.ts
 */
export { isBackendAvailable } from "./testData";

/**
 * Click an element using JavaScript to bypass viewport checks.
 * Useful for position:fixed elements that Playwright considers outside viewport.
 * Waits for the element to be attached first to avoid hydration race conditions.
 */
export async function clickBySelector(page: Page, selector: string): Promise<void> {
  await page.waitForSelector(selector, { state: "attached", timeout: 5000 });
  await page.evaluate((sel) => {
    const element = document.querySelector(sel);
    if (element) {
      (element as HTMLElement).click();
    } else {
      throw new Error(`Element with selector "${sel}" not found`);
    }
  }, selector);
}

/**
 * Click a floating controls button using JavaScript to bypass viewport checks.
 * Floating controls use position:fixed which Playwright considers outside viewport.
 */
export async function clickFloatingControl(page: Page, dataTestId: string): Promise<void> {
  await clickBySelector(page, `[data-testid="${dataTestId}"]`);
}

/**
 * Open a cockpit panel by clicking its visible handle and waiting for the slide-out animation.
 * Waits for the handle to mount to avoid hydration race conditions; if it never appears,
 * the panel may already be open (or the page doesn't have this cockpit).
 */
export async function openCockpitPanel(
  page: Page,
  position: "top" | "left" | "right" | "bottom" = "top"
): Promise<void> {
  const handle = page.locator(`[data-testid="cockpit-handle-${position}"]`).first();
  try {
    await handle.waitFor({ state: "attached", timeout: 2000 });
  } catch {
    // Handle not present: panel is already open or this page has no cockpit.
    return;
  }
  await clickBySelector(page, `[data-testid="cockpit-handle-${position}"]`);
  await page.waitForTimeout(500);
}

/**
 * Click create note button in the top bar
 */
export async function clickCreateNoteButton(page: Page): Promise<void> {
  await clickBySelector(page, '[data-testid="create-note-button"]');
}

/**
 * Click view toggle button in the top bar
 */
export async function clickViewToggle(page: Page, view: "list" | "graph" | "3d"): Promise<void> {
  const testId = `view-toggle-${view}`;
  await clickBySelector(page, `[data-testid="${testId}"]`);
}

/**
 * Click filter chip in the top bar type dropdown
 */
export async function clickFilterChip(page: Page, filter: string): Promise<void> {
  const filterId = filter.toLowerCase().replace(/s$/, ""); // stars -> star
  // Open the type dropdown
  await clickBySelector(page, '[data-testid="type-dropdown-toggle"]');
  await page.waitForTimeout(300);
  await clickBySelector(page, `[data-testid="filter-chip-${filterId}"]`);
  await page.waitForTimeout(300);
}

/**
 * Fill search input in the top bar (home/graph pages) or search page
 */
export async function fillSearchInput(page: Page, query: string): Promise<void> {
  // Try the top bar search input first, then fall back to the dedicated search page input.
  const searchInput = page.locator('[data-testid="top-bar-search-input"], [data-testid="search-input"]').first();
  await expect(searchInput).toBeVisible({ timeout: 5000 });
  await searchInput.fill(query);
  await searchInput.dispatchEvent("input");
}

/**
 * Submit the active search input (the top bar no longer has a separate search button)
 */
export async function clickSearchButton(page: Page): Promise<void> {
  const searchInput = page.locator('[data-testid="top-bar-search-input"], [data-testid="search-input"]').first();
  await searchInput.press("Enter");
  await page.waitForTimeout(300);
}

/**
 * Setup SKIP_AUTH mode for testing
 * Must be called before navigating to protected routes
 *
 * Works with both dev server (localStorage) and production build (query param)
 * Simplified version that doesn't navigate to auth page
 */
export async function setupSkipAuth(page: Page): Promise<void> {
  // Set SKIP_AUTH flag directly without navigating to auth page
  // Use addInitScript only - evaluate might fail due to localStorage restrictions
  await page.addInitScript(() => {
    try {
      localStorage.setItem("__SKIP_AUTH__", "true");
    } catch {
      // localStorage might not be available in some contexts
      console.log("[TEST] localStorage not available, using window flag");
    }
    (window as any).__SKIP_AUTH__ = true;
    console.log("[TEST] SKIP_AUTH enabled via initScript");
  });
}
