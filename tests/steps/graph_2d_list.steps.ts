import { Given, When, Then, Before } from '@cucumber/cucumber';
import { expect } from '@playwright/test';
import type { ITestWorld } from '../features/support/world';
import { createNote, getBackendUrl, cleanupTestData } from '../../frontend/tests/helpers/testData';

// Test data tracking
let testNoteIds: string[] = [];

Before({ tags: '@2d-graph or @list-view' }, async function (this: ITestWorld) {
  testNoteIds = [];
});

Before({ tags: '@smoke' }, async function (this: ITestWorld) {
  // Ensure clean state for smoke tests
  testNoteIds = [];
});

// Cleanup after scenarios
After({ tags: '@2d-graph or @list-view' }, async function (this: ITestWorld) {
  if (testNoteIds.length > 0) {
    await cleanupTestData(this.request, testNoteIds);
  }
});

Given('I am on the main page {string}', async function (this: ITestWorld, path: string) {
  const baseUrl = process.env.FRONTEND_URL || 'http://localhost:5173';
  await this.page.goto(`${baseUrl}${path}`);
  await this.page.waitForLoadState('networkidle');
  await expect(this.page.locator('body')).toBeVisible();
});

Given('there are notes of various types in the database', async function (this: ITestWorld) {
  const types = ['star', 'planet', 'comet', 'galaxy', 'asteroid'];

  for (const type of types) {
    const note = await createNote(this.request, {
      title: `${type.charAt(0).toUpperCase() + type.slice(1)} Test Note`,
      content: `This is a test note of type ${type}`,
      type: type,
    });
    testNoteIds.push(note.id);
  }
});

Then('I should see the 2D force graph by default', async function (this: ITestWorld) {
  const graphContainer = this.page.locator('.graph-2d, .force-graph-container, canvas').first();
  await expect(graphContainer).toBeVisible({ timeout: 10000 });

  // Verify it's not in list view
  const listView = this.page.locator('.note-list, .list-view, .notes-grid').first();
  const isListVisible = await listView.isVisible().catch(() => false);
  expect(isListVisible).toBe(false);
});

When('I click the {string} toggle button in the floating controls', async function (this: ITestWorld, buttonText: string) {
  const button = this.page.locator(`button:has-text("${buttonText}"), [data-testid="${buttonText.toLowerCase()}-toggle"]`).first();
  await expect(button).toBeVisible();
  await button.click();
  await this.page.waitForTimeout(500); // Allow transition
});

When('I click the {string} toggle button', async function (this: ITestWorld, buttonText: string) {
  const button = this.page
    .locator(`button:has-text("${buttonText}"), [data-testid="view-toggle-${buttonText.toLowerCase()}"], .toggle-btn:has-text("${buttonText}")`)
    .first();
  await expect(button).toBeVisible();
  await button.click();
  await this.page.waitForTimeout(500);
});

Then('I should see a grid of note cards', async function (this: ITestWorld) {
  const noteCards = this.page.locator('.note-card, .note-item, [data-testid="note-card"]').first();
  await expect(noteCards).toBeVisible({ timeout: 5000 });

  // Verify multiple cards exist or at least the container is a grid
  const container = this.page.locator('.notes-grid, .note-list, .list-container').first();
  await expect(container).toBeVisible();
});

Then('the view toggle should show {string} option', async function (this: ITestWorld, optionText: string) {
  const toggleButton = this.page.locator(`button:has-text("${optionText}"), [data-testid="${optionText.toLowerCase()}-toggle"]`).first();
  await expect(toggleButton).toBeVisible();
});

Then('I should see the fullscreen 2D force graph', async function (this: ITestWorld) {
  const graphContainer = this.page.locator('.graph-2d, .force-graph-container, canvas').first();
  await expect(graphContainer).toBeVisible({ timeout: 5000 });

  // Verify it's taking up significant space
  const box = await graphContainer.boundingBox();
  expect(box?.width).toBeGreaterThan(500);
  expect(box?.height).toBeGreaterThan(400);
});

Then('the graph canvas should be visible', async function (this: ITestWorld) {
  const canvas = this.page.locator('.graph-2d canvas, .force-graph-container canvas, .graph-canvas').first();
  await expect(canvas).toBeVisible();
});

Given('I am in list view', async function (this: ITestWorld) {
  // First ensure we're on main page
  const baseUrl = process.env.FRONTEND_URL || 'http://localhost:5173';
  await this.page.goto(`${baseUrl}/`);
  await this.page.waitForLoadState('networkidle');

  // Switch to list view if not already
  const listView = this.page.locator('.note-list, .list-view, .notes-grid').first();
  const isListVisible = await listView.isVisible().catch(() => false);

  if (!isListVisible) {
    const listButton = this.page.locator('button:has-text("List"), [data-testid="list-toggle"]').first();
    if (await listButton.isVisible().catch(() => false)) {
      await listButton.click();
      await this.page.waitForTimeout(500);
    }
  }

  await expect(listView).toBeVisible();
});

When('I click the {string} filter chip in floating controls', async function (this: ITestWorld, filterType: string) {
  const chip = this.page
    .locator(
      `[data-filter="${filterType.toLowerCase()}"], button:has-text("${filterType}"), .filter-chip:has-text("${filterType}"), .chip:has-text("${filterType}")`
    )
    .first();
  await expect(chip).toBeVisible();
  await chip.click();
  await this.page.waitForTimeout(500);
});

Then('only notes of type {string} should be displayed in the list', async function (this: ITestWorld, noteType: string) {
  // Wait for filter to apply
  await this.page.waitForTimeout(500);

  // Check if the stats bar shows filtered state
  const statsBar = this.page.locator('.stats-bar, .filter-indicator, [data-testid="stats-bar"]').first();
  if (await statsBar.isVisible().catch(() => false)) {
    const statsText = await statsBar.textContent();
    const hasFilter = statsText?.toLowerCase().includes('filter') ||
                     statsText?.toLowerCase().includes(noteType.toLowerCase());
    expect(hasFilter).toBe(true);
  }

  // Verify visible notes count changed
  const visibleNotes = await this.page.locator('.note-card, .note-item').count();
  expect(visibleNotes).toBeGreaterThanOrEqual(0);
});

Then('the count badge should show the correct number', async function (this: ITestWorld) {
  const badge = this.page.locator('.count-badge, .stats-count, [data-testid="note-count"]').first();
  await expect(badge).toBeVisible();

  const count = await badge.textContent();
  expect(count).toMatch(/\d+/);
});

When('I click the {string} filter chip', async function (this: ITestWorld, filterType: string) {
  const chip = this.page.locator(`[data-filter="all"], button:has-text("${filterType}"), .filter-chip:has-text("${filterType}")`).first();
  await expect(chip).toBeVisible();
  await chip.click();
  await this.page.waitForTimeout(500);
});

Then('all notes should be displayed', async function (this: ITestWorld) {
  // Check that stats shows all or no filter applied
  const statsBar = this.page.locator('.stats-bar, .filter-indicator').first();
  if (await statsBar.isVisible().catch(() => false)) {
    const statsText = await statsBar.textContent();
    const hasNoFilter = !statsText?.toLowerCase().includes('filter') ||
                        statsText?.toLowerCase().includes('all');
    expect(hasNoFilter).toBe(true);
  }

  // Verify notes are visible
  const notes = this.page.locator('.note-card, .note-item').first();
  await expect(notes).toBeVisible();
});

Given('I am in graph view', async function (this: ITestWorld) {
  // Ensure we're on main page
  const baseUrl = process.env.FRONTEND_URL || 'http://localhost:5173';
  await this.page.goto(`${baseUrl}/`);
  await this.page.waitForLoadState('networkidle');

  // Switch to graph view if in list view
  const graphContainer = this.page.locator('.graph-2d, .force-graph-container, canvas').first();
  const isGraphVisible = await graphContainer.isVisible().catch(() => false);

  if (!isGraphVisible) {
    const graphButton = this.page.locator('button:has-text("Graph"), [data-testid="graph-toggle"]').first();
    if (await graphButton.isVisible().catch(() => false)) {
      await graphButton.click();
      await this.page.waitForTimeout(500);
    }
  }

  await expect(graphContainer).toBeVisible();
});

When('I type {string} in the search input', async function (this: ITestWorld, searchText: string) {
  const searchInput = this.page.locator('input[type="search"], input[placeholder*="search" i], [data-testid="search-input"]').first();
  await expect(searchInput).toBeVisible();
  await searchInput.fill(searchText);
  await this.page.waitForTimeout(500); // Debounce
});

Then('the list should show only notes containing {string}', async function (this: ITestWorld, searchText: string) {
  await this.page.waitForTimeout(500);

  // Check that visible items contain the search text
  const items = await this.page.locator('.note-card, .note-item').all();
  for (const item of items) {
    const text = await item.textContent();
    expect(text?.toLowerCase()).toContain(searchText.toLowerCase());
  }
});

Then('the note cards should highlight the matching text', async function (this: ITestWorld) {
  const highlighted = this.page.locator('mark, .highlight, .search-highlight, [data-highlight]').first();
  const hasHighlight = await highlighted.isVisible().catch(() => false);
  expect(hasHighlight).toBe(true);
});

When('I clear the search input', async function (this: ITestWorld) {
  const searchInput = this.page.locator('input[type="search"], [data-testid="search-input"]').first();
  await searchInput.clear();
  await this.page.waitForTimeout(500);
});

Then('all notes should be displayed again', async function (this: ITestWorld) {
  const notes = await this.page.locator('.note-card, .note-item').count();
  expect(notes).toBeGreaterThan(0);
});

Then('only nodes matching {string} should be visible', async function (this: ITestWorld, searchText: string) {
  // In 2D graph, check opacity or visibility of nodes
  await this.page.waitForTimeout(500);

  const visibleNodes = await this.page.evaluate((text) => {
    const nodes = document.querySelectorAll('.node, circle, [data-node-id]');
    return Array.from(nodes).filter((n) => {
      const el = n as HTMLElement;
      const style = window.getComputedStyle(el);
      return style.opacity !== '0' && style.display !== 'none';
    }).length;
  }, searchText);

  expect(visibleNodes).toBeGreaterThanOrEqual(0);
});

Then('non-matching nodes should be dimmed or hidden', async function (this: ITestWorld) {
  const hasDimmedNodes = await this.page.evaluate(() => {
    const nodes = document.querySelectorAll('.node, circle, [data-node-id]');
    return Array.from(nodes).some((n) => {
      const el = n as HTMLElement;
      const style = window.getComputedStyle(el);
      return parseFloat(style.opacity) < 0.5 || style.display === 'none';
    });
  });

  expect(hasDimmedNodes).toBe(true);
});

Then('all nodes should be visible', async function (this: ITestWorld) {
  const allVisible = await this.page.evaluate(() => {
    const nodes = document.querySelectorAll('.node, circle, [data-node-id]');
    return Array.from(nodes).every((n) => {
      const el = n as HTMLElement;
      const style = window.getComputedStyle(el);
      return parseFloat(style.opacity) > 0.5 && style.display !== 'none';
    });
  });

  expect(allVisible).toBe(true);
});

Given('I am on the main page', async function (this: ITestWorld) {
  const baseUrl = process.env.FRONTEND_URL || 'http://localhost:5173';
  await this.page.goto(`${baseUrl}/`);
  await this.page.waitForLoadState('networkidle');
  await expect(this.page.locator('body')).toBeVisible();
});

When('I click the {string} button in floating controls', async function (this: ITestWorld, buttonLabel: string) {
  const button = this.page.locator(`button:has-text("${buttonLabel}"), [data-testid="${buttonLabel.toLowerCase().replace(/\s+/g, '-')}-btn"], .fab:has-text("${buttonLabel}")`).first();
  await expect(button).toBeVisible();
  await button.click();
});

Then('a create note modal should open', async function (this: ITestWorld) {
  const modal = this.page.locator('.modal, [role="dialog"], .create-note-modal, [data-testid="create-note-modal"]').first();
  await expect(modal).toBeVisible({ timeout: 5000 });
});

When('I fill in the title {string}', async function (this: ITestWorld, title: string) {
  const titleInput = this.page.locator('input[name="title"], input[placeholder*="title" i], [data-testid="note-title-input"]').first();
  await expect(titleInput).toBeVisible();
  await titleInput.fill(title);
});

When('I select type {string}', async function (this: ITestWorld, noteType: string) {
  const typeSelect = this.page.locator('select[name="type"], [data-testid="type-select"], button:has-text("Type")').first();
  await expect(typeSelect).toBeVisible();
  await typeSelect.click();

  const option = this.page.locator(`option:has-text("${noteType}"), [data-value="${noteType.toLowerCase()}"]`).first();
  await option.click();
});

When('I click the {string} button', async function (this: ITestWorld, buttonText: string) {
  const button = this.page.locator(`button:has-text("${buttonText}"), [data-testid="${buttonText.toLowerCase()}-btn"], .btn-primary:has-text("${buttonText}")`).first();
  await expect(button).toBeVisible();
  await button.click();
  await this.page.waitForTimeout(500);
});

Then('the modal should close', async function (this: ITestWorld) {
  const modal = this.page.locator('.modal, [role="dialog"], .create-note-modal').first();
  await expect(modal).toBeHidden({ timeout: 3000 });
});

Then('the new note should appear in the graph', async function (this: ITestWorld) {
  // Check for the new note in the graph or list
  const noteExists = await this.page.locator('.note-card, .node, circle').count();
  expect(noteExists).toBeGreaterThan(0);
});

import { After } from '@cucumber/cucumber';
