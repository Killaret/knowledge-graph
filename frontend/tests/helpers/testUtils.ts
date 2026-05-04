import type { Page, APIRequestContext } from '@playwright/test';

/**
 * Get backend URL from environment or use default
 */
export function getBackendUrl(): string {
  return process.env.BACKEND_URL || 'http://localhost:8080/api';
}

/**
 * Create a note via API for testing
 */
export async function createNote(
  request: APIRequestContext,
  data: { title: string; content?: string; type?: string }
): Promise<{ data: { id: string; title: string } }> {
  const response = await request.post(`${getBackendUrl()}/v1/notes`, {
    data: {
      title: data.title,
      content: data.content || 'Test content',
      type: data.type || 'star'
    }
  });
  return await response.json() as { data: { id: string; title: string } };
}

/**
 * Create a link between notes via API for testing
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
  await request.post(`${getBackendUrl()}/v1/links`, {
    data: {
      source_note_id: data.source_note_id,
      target_note_id: data.target_note_id,
      link_type: data.link_type || 'related',
      weight: data.weight ?? 0.5
    }
  });
}

/**
 * Click an element using JavaScript to bypass viewport checks.
 * Useful for position:fixed elements that Playwright considers outside viewport.
 */
export async function clickBySelector(page: Page, selector: string): Promise<void> {
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
 * Click create note button in floating controls
 */
export async function clickCreateNoteButton(page: Page): Promise<void> {
  await clickFloatingControl(page, 'create-note-button');
}

/**
 * Click view toggle button in floating controls
 */
export async function clickViewToggle(page: Page, view: 'list' | 'graph' | '3d'): Promise<void> {
  const testId = `view-toggle-${view}`;
  await clickFloatingControl(page, testId);
}

/**
 * Click filter chip in floating controls
 */
export async function clickFilterChip(page: Page, filter: string): Promise<void> {
  const filterId = filter.toLowerCase().replace(/s$/, ''); // stars -> star
  await clickFloatingControl(page, `filter-chip-${filterId}`);
}

/**
 * Fill search input in floating controls
 */
export async function fillSearchInput(page: Page, query: string): Promise<void> {
  await page.evaluate((q) => {
    const input = document.querySelector('[data-testid="search-input"]') as HTMLInputElement;
    if (input) {
      input.value = q;
      input.dispatchEvent(new Event('input', { bubbles: true }));
    } else {
      throw new Error('Search input not found');
    }
  }, query);
}

/**
 * Click search button in floating controls
 */
export async function clickSearchButton(page: Page): Promise<void> {
  // Search button doesn't have data-testid, use class selector via JS
  await page.evaluate(() => {
    const button = document.querySelector('.search-btn');
    if (button) {
      (button as HTMLElement).click();
    } else {
      throw new Error('Search button not found');
    }
  });
}
