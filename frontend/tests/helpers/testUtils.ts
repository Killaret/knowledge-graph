import type { Page } from '@playwright/test';

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
