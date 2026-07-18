import type { Page, APIRequestContext } from "@playwright/test";
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
  data: { title: string; content?: string; type?: string; email?: string },
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
  },
): Promise<void> {
  await createLinkAdvanced(
    request,
    data.source_note_id,
    data.target_note_id,
    data.weight ?? 0.5,
    data.link_type ?? "related",
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
 */
export async function clickBySelector(
  page: Page,
  selector: string,
): Promise<void> {
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
export async function clickFloatingControl(
  page: Page,
  dataTestId: string,
): Promise<void> {
  await clickBySelector(page, `[data-testid="${dataTestId}"]`);
}

/**
 * Click create note button in floating controls
 */
export async function clickCreateNoteButton(page: Page): Promise<void> {
  await clickFloatingControl(page, "create-note-button");
}

/**
 * Click view toggle button in floating controls
 */
export async function clickViewToggle(
  page: Page,
  view: "list" | "graph" | "3d",
): Promise<void> {
  const testId = `view-toggle-${view}`;
  await clickFloatingControl(page, testId);
}

/**
 * Click filter chip in floating controls
 */
export async function clickFilterChip(
  page: Page,
  filter: string,
): Promise<void> {
  const filterId = filter.toLowerCase().replace(/s$/, ""); // stars -> star
  await clickFloatingControl(page, `filter-chip-${filterId}`);
}

/**
 * Fill search input in floating controls
 */
export async function fillSearchInput(
  page: Page,
  query: string,
): Promise<void> {
  await page.evaluate((q) => {
    const input = document.querySelector(
      '[data-testid="search-input"]',
    ) as HTMLInputElement;
    if (input) {
      input.value = q;
      input.dispatchEvent(new Event("input", { bubbles: true }));
    } else {
      throw new Error("Search input not found");
    }
  }, query);
}

/**
 * Click search button in floating controls
 */
export async function clickSearchButton(page: Page): Promise<void> {
  // Search button doesn't have data-testid, use class selector via JS
  await page.evaluate(() => {
    const button = document.querySelector(".search-btn");
    if (button) {
      (button as HTMLElement).click();
    } else {
      throw new Error("Search button not found");
    }
  });
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
