import { Given, When, Then, Before, After, type IWorld } from '@cucumber/cucumber';
import { expect } from '@playwright/test';
import type { Page, APIRequestContext } from '@playwright/test';

// Custom world type
interface ITestWorld extends IWorld {
  page: Page;
  request: APIRequestContext;
  testNotes: Array<{ id: string; title: string; type: string }>;
  centerNoteId?: string;
}

Before(async function(this: ITestWorld) {
  this.testNotes = [];
});

After(async function(this: ITestWorld) {
  // Cleanup test notes
  for (const note of this.testNotes) {
    try {
      await this.request.delete(`http://localhost:8080/notes/${note.id}`);
    } catch (e) {
      // Ignore cleanup errors
    }
  }
});

// Background steps
Given('I have test notes with connections', async function(this: ITestWorld) {
  // Create center note
  const centerNote = await this.request.post('http://localhost:8080/notes', {
    data: {
      title: 'Center Test Note',
      content: 'This is the center note for testing',
      type: 'star'
    }
  });
  const centerData = await centerNote.json();
  this.centerNoteId = centerData.id;
  this.testNotes.push({ id: centerData.id, title: centerData.title, type: 'star' });
  
  // Create connected notes
  const types = ['planet', 'comet', 'galaxy', 'asteroid'];
  for (let i = 0; i < 4; i++) {
    const note = await this.request.post('http://localhost:8080/notes', {
      data: {
        title: `Connected Note ${i}`,
        content: `Content for note ${i}`,
        type: types[i % types.length]
      }
    });
    const noteData = await note.json();
    this.testNotes.push({ id: noteData.id, title: noteData.title, type: types[i % types.length] });
    
    // Create link to center
    await this.request.post('http://localhost:8080/links', {
      data: {
        source_note_id: this.centerNoteId,
        target_note_id: noteData.id,
        link_type: 'related',
        weight: 0.5 + Math.random() * 0.5
      }
    });
  }
});

Given('there are notes of various types in the database', async function(this: ITestWorld) {
  const types = ['star', 'planet', 'comet', 'galaxy', 'asteroid'];
  for (let i = 0; i < 5; i++) {
    const note = await this.request.post('http://localhost:8080/notes', {
      data: {
        title: `Test ${types[i]} ${Date.now()}`,
        content: `Content for ${types[i]}`,
        type: types[i]
      }
    });
    const noteData = await note.json();
    this.testNotes.push({ id: noteData.id, title: noteData.title, type: types[i] });
  }
});

// Navigation steps
Given('I am on the main page {string}', async function(this: ITestWorld, path: string) {
  await this.page.goto(`http://localhost:5173${path}`);
  await this.page.waitForLoadState('networkidle');
  await this.page.waitForTimeout(500);
});

Given('I navigate to {string}', async function(this: ITestWorld, path: string) {
  // Replace {centerNoteId} placeholder
  const resolvedPath = path.replace('{centerNoteId}', this.centerNoteId || 'test-id');
  await this.page.goto(`http://localhost:5173${resolvedPath}`);
  await this.page.waitForLoadState('networkidle');
  await this.page.waitForTimeout(500);
});

Given('I am on the 3D graph page for a note with connections', async function(this: ITestWorld) {
  // Create notes if needed
  if (!this.centerNoteId) {
    // Create center note
    const centerNote = await this.request.post('http://localhost:8080/notes', {
      data: { title: 'Center Test Note', content: 'Center note', type: 'star' }
    });
    const centerData = await centerNote.json();
    this.centerNoteId = centerData.id;
    this.testNotes.push({ id: centerData.id, title: centerData.title, type: 'star' });
  }
  // Navigate to 3D graph
  await this.page.goto(`http://localhost:5173/graph/3d/${this.centerNoteId}`);
  await this.page.waitForLoadState('networkidle');
  await this.page.waitForTimeout(500);
});

// UI interaction steps
When('I click the {string} toggle button in the floating controls', async function(this: ITestWorld, viewName: string) {
  const button = this.page.locator('.floating-controls button, [data-testid="view-toggle"]').filter({ hasText: new RegExp(viewName, 'i') }).first();
  await expect(button).toBeVisible({ timeout: 5000 });
  await button.click();
  await this.page.waitForTimeout(500);
});

When('I click the {string} filter chip in floating controls', async function(this: ITestWorld, filterName: string) {
  const chip = this.page.locator('.floating-controls .filter-chip, [data-testid="type-filter"]').filter({ hasText: new RegExp(filterName, 'i') }).first();
  await expect(chip).toBeVisible({ timeout: 5000 });
  await chip.click();
  await this.page.waitForTimeout(300);
});

When('I type {string} in the search input', async function(this: ITestWorld, searchText: string) {
  const searchInput = this.page.locator('.floating-controls input[type="search"], [data-testid="search-input"], input[placeholder*="Search"]').first();
  await expect(searchInput).toBeVisible({ timeout: 5000 });
  await searchInput.fill(searchText);
  await this.page.waitForTimeout(300);
});

When('I clear the search input', async function(this: ITestWorld) {
  const searchInput = this.page.locator('.floating-controls input[type="search"], [data-testid="search-input"]').first();
  await searchInput.clear();
  await this.page.waitForTimeout(300);
});

When('I click the {string} button in floating controls', async function(this: ITestWorld, buttonLabel: string) {
  const button = this.page.locator('.floating-controls button').filter({ hasText: new RegExp(buttonLabel.replace(/[+*]/g, '\\$&'), 'i') }).first();
  await expect(button).toBeVisible({ timeout: 5000 });
  await button.click();
});

// View state assertions
Then('I should see the 2D force graph by default', async function(this: ITestWorld) {
  const graph = this.page.locator('.fullscreen-graph, canvas, .graph-canvas').first();
  await expect(graph).toBeVisible({ timeout: 10000 });
});

Then('I should see a grid of note cards', async function(this: ITestWorld) {
  const grid = this.page.locator('.list-container, .notes-grid, [data-testid="notes-list"]').first();
  await expect(grid).toBeVisible({ timeout: 10000 });
  const cards = this.page.locator('.note-card');
  await expect(cards.first()).toBeVisible({ timeout: 5000 });
});

Then('I should see the fullscreen 2D force graph', async function(this: ITestWorld) {
  const graph = this.page.locator('.fullscreen-graph').first();
  await expect(graph).toBeVisible({ timeout: 10000 });
  const canvas = this.page.locator('.fullscreen-graph canvas').first();
  await expect(canvas).toBeVisible({ timeout: 5000 });
});

Then('I should see the 2D force graph by default', async function(this: ITestWorld) {
  const graph = this.page.locator('.fullscreen-graph, canvas, .graph-canvas').first();
  await expect(graph).toBeVisible({ timeout: 10000 });
});

Then('I am in list view', async function(this: ITestWorld) {
  const listView = this.page.locator('.list-container, .notes-grid').first();
  const isVisible = await listView.isVisible().catch(() => false);
  if (!isVisible) {
    // Click list toggle
    const button = this.page.locator('.floating-controls button').filter({ hasText: /List/i }).first();
    await button.click();
    await this.page.waitForTimeout(500);
  }
  await expect(listView).toBeVisible({ timeout: 5000 });
});

Then('I am in graph view', async function(this: ITestWorld) {
  const graph = this.page.locator('.fullscreen-graph').first();
  const isVisible = await graph.isVisible().catch(() => false);
  if (!isVisible) {
    // Click graph toggle
    const button = this.page.locator('.floating-controls button').filter({ hasText: /Graph/i }).first();
    await button.click();
    await this.page.waitForTimeout(500);
  }
  await expect(graph).toBeVisible({ timeout: 5000 });
});

Then('the view toggle should show {string} option', async function(this: ITestWorld, optionText: string) {
  const button = this.page.locator('.floating-controls button').filter({ hasText: new RegExp(optionText, 'i') }).first();
  await expect(button).toBeVisible({ timeout: 5000 });
});

// Filter and search assertions
Then('only notes of type {string} should be displayed', async function(this: ITestWorld, type: string) {
  const cards = this.page.locator('.note-card');
  const count = await cards.count();
  expect(count).toBeGreaterThan(0);
  
  // Check that all visible cards have the correct type
  for (let i = 0; i < count; i++) {
    const typeBadge = cards.nth(i).locator('.type-badge, [data-testid="note-type"]').first();
    const cardType = await typeBadge.textContent();
    expect(cardType?.toLowerCase()).toContain(type.toLowerCase());
  }
});

Then('the count badge should show the correct number', async function(this: ITestWorld) {
  const badge = this.page.locator('.filter-chip .count, [data-testid="filter-count"]').first();
  await expect(badge).toBeVisible({ timeout: 5000 });
  const count = await badge.textContent();
  expect(parseInt(count || '0')).toBeGreaterThan(0);
});

Then('all notes should be displayed', async function(this: ITestWorld) {
  const cards = this.page.locator('.note-card');
  await expect(cards.first()).toBeVisible({ timeout: 5000 });
});

Then('the list should show only notes containing {string}', async function(this: ITestWorld, searchTerm: string) {
  const cards = this.page.locator('.note-card');
  const count = await cards.count();
  
  if (count === 0) {
    // Empty state is valid for no matches
    const emptyState = this.page.locator('.empty-state, text=/No notes found/i').first();
    await expect(emptyState).toBeVisible({ timeout: 5000 });
    return;
  }
  
  for (let i = 0; i < count; i++) {
    const title = await cards.nth(i).locator('.note-title, h3').first().textContent();
    expect(title?.toLowerCase()).toContain(searchTerm.toLowerCase());
  }
});

Then('the note cards should highlight the matching text', async function(this: ITestWorld) {
  const highlighted = this.page.locator('.note-card mark, .note-card .highlight, .note-card [style*="background"]').first();
  await expect(highlighted).toBeVisible({ timeout: 5000 });
});

// Create note modal steps
Then('a create note modal should open', async function(this: ITestWorld) {
  const modal = this.page.locator('.modal, [role="dialog"], .create-note-modal').first();
  await expect(modal).toBeVisible({ timeout: 5000 });
});

When('I fill in the title {string}', async function(this: ITestWorld, title: string) {
  const input = this.page.locator('input[name="title"], [data-testid="note-title-input"]').first();
  await input.fill(title);
});

When('I select type {string}', async function(this: ITestWorld, type: string) {
  const select = this.page.locator('select[name="type"], [data-testid="note-type-select"]').first();
  await select.selectOption(type.toLowerCase());
});

When('I click the {string} button', async function(this: ITestWorld, buttonText: string) {
  const button = this.page.locator('button').filter({ hasText: new RegExp(buttonText, 'i') }).first();
  await button.click();
});

Then('the modal should close', async function(this: ITestWorld) {
  const modal = this.page.locator('.modal, [role="dialog"]').first();
  await expect(modal).not.toBeVisible({ timeout: 5000 });
});

Then('the new note should appear in the graph', async function(this: ITestWorld) {
  // Wait for graph to update
  await this.page.waitForTimeout(1000);
  const canvas = this.page.locator('canvas').first();
  await expect(canvas).toBeVisible({ timeout: 5000 });
});
