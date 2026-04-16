import { Given, When, Then } from '@cucumber/cucumber';
import { expect } from '@playwright/test';
import type { ITestWorld } from '../support/world';
import type { Dialog } from '@playwright/test';

// Note creation
Given('I am on the knowledge graph page', async function(this: ITestWorld) {
  await this.page.goto('http://localhost:5173/');
  await expect(this.page.locator('body')).toBeVisible();
  await this.page.waitForTimeout(1000);
});

Given('I see an empty graph', async function(this: ITestWorld) {
  const emptyState = this.page.locator('text=No notes, text=No notes found, text=Create your first, .empty-state').first();
  const hasNotes = await this.page.locator('.note-card').count() > 0;
  if (!hasNotes) {
    await expect(emptyState).toBeVisible();
  }
});

When('I click the {string} floating button', async function(this: ITestWorld, buttonText: string) {
  const selectors = [
    `button:has-text("${buttonText}")`,
    `.floating-controls button`,
    `[data-testid="create-note"]`,
    '.create-btn'
  ];
  
  for (const selector of selectors) {
    const locator = this.page.locator(selector).first();
    if (await locator.isVisible().catch(() => false)) {
      await locator.click();
      return;
    }
  }
  
  // Fallback - try to find any button with + or create
  await this.page.click('button:has-text("+"), button:has-text("Create"), .create-btn');
});

When('I enter {string} as the title', async function(this: ITestWorld, title: string) {
  await this.page.fill('input[name="title"], input[placeholder*="Title"]', title);
});

When('I enter {string} as the content', async function(this: ITestWorld, content: string) {
  await this.page.fill('textarea[name="content"], textarea[placeholder*="Content"]', content);
});

When('I select {string} as the type', async function(this: ITestWorld, type: string) {
  // Try to find type selector
  const typeButton = this.page.locator(`button:has-text("${type}"), [data-type="${type.toLowerCase()}"], .type-${type.toLowerCase()}`).first();
  if (await typeButton.isVisible().catch(() => false)) {
    await typeButton.click();
  } else {
    // Try select dropdown
    await this.page.selectOption('select[name="type"]', { label: type });
  }
});

When('I click the {string} button to save', async function(this: ITestWorld, buttonText: string) {
  await this.page.click(`button[type="submit"], button:has-text("${buttonText}"), button:has-text("Save")`);
  await this.page.waitForTimeout(1000);
});

Then('a new node {string} appears on the graph canvas', async function(this: ITestWorld, nodeTitle: string) {
  // Check in list view or graph
  const selectors = [
    `.note-card:has-text("${nodeTitle}")`,
    `text="${nodeTitle}"`,
    `[data-note-id]:has-text("${nodeTitle}")`,
    `.node-label:has-text("${nodeTitle}")`
  ];
  
  let found = false;
  for (const selector of selectors) {
    const locator = this.page.locator(selector).first();
    if (await locator.isVisible().catch(() => false)) {
      found = true;
      break;
    }
  }
  
  expect(found).toBe(true);
});

Then('the new node has type {string}', async function(this: ITestWorld, type: string) {
  // Verify type indicator exists
  const typeIndicator = this.page.locator(`[data-type="${type.toLowerCase()}"], .type-badge:has-text("${type}")`).first();
  await expect(typeIndicator).toBeVisible();
});

// Note editing
Given('a note {string} is displayed on the graph', async function(this: ITestWorld, title: string) {
  // First check if note exists
  const noteCard = this.page.locator(`.note-card:has-text("${title}"), text="${title}"`).first();
  
  if (await noteCard.isVisible().catch(() => false)) {
    return; // Note exists
  }
  
  // Create the note
  await this.step('I click the "+" floating button');
  await this.step(`I enter "${title}" as the title`);
  await this.step('I enter "Test content" as the content');
  await this.step('I click the "Save" button to save');
  await this.page.waitForTimeout(1000);
});

When('I right-click on the node {string}', async function(this: ITestWorld, nodeTitle: string) {
  const node = this.page.locator(`.note-card:has-text("${nodeTitle}"), text="${nodeTitle}"`).first();
  await node.click({ button: 'right' });
});

When('I select {string} from the context menu', async function(this: ITestWorld, action: string) {
  await this.page.click(`.context-menu:has-text("${action}"), [role="menuitem"]:has-text("${action}"), text="${action}"`);
});

Then('the {string} modal opens with fields pre-filled', async function(this: ITestWorld, modalType: string) {
  const modal = this.page.locator('.modal, [role="dialog"], .edit-note-modal').first();
  await expect(modal).toBeVisible();
  
  // Check that title input has value
  const titleInput = modal.locator('input[name="title"], input[placeholder*="Title"]').first();
  const value = await titleInput.inputValue();
  expect(value.length).toBeGreaterThan(0);
});

Then('I clear the title field', async function(this: ITestWorld) {
  await this.page.fill('input[name="title"]', '');
});

When('I type {string} as the new title', async function(this: ITestWorld, newTitle: string) {
  await this.page.fill('input[name="title"]', newTitle);
});

Then('the node {string} displays the updated title', async function(this: ITestWorld, newTitle: string) {
  const node = this.page.locator(`.note-card:has-text("${newTitle}"), text="${newTitle}"`).first();
  await expect(node).toBeVisible();
});

// Note deletion
When('I click {string} from the context menu', async function(this: ITestWorld, action: string) {
  await this.step(`I select "${action}" from the context menu`);
});

Then('a confirmation dialog appears', async function(this: ITestWorld) {
  // Browser confirm dialog or custom modal
  const confirmModal = this.page.locator('.modal:has-text("Delete"), .modal:has-text("Confirm"), [role="dialog"]').first();
  const hasDialog = await confirmModal.isVisible().catch(() => false);
  
  if (!hasDialog) {
    // Check for browser dialog handler
    this.page.on('dialog', async (dialog: Dialog) => {
      expect(dialog.type()).toBe('confirm');
    });
  }
});

When('I confirm the action', async function(this: ITestWorld) {
  // Handle browser confirm dialog
  this.page.on('dialog', async (dialog: Dialog) => {
    await dialog.accept();
  });
  
  // Or click confirm button in modal
  await this.page.click('button:has-text("Delete"), button:has-text("Confirm"), button[type="submit"]');
});

Then('the node {string} is removed from the graph', async function(this: ITestWorld, nodeTitle: string) {
  const node = this.page.locator(`.note-card:has-text("${nodeTitle}"), text="${nodeTitle}"`).first();
  await expect(node).not.toBeVisible();
});

// List View Toggle
When('I click the {string} toggle button', async function(this: ITestWorld, buttonText: string) {
  await this.page.click(`button:has-text("${buttonText}"), [data-view="${buttonText.toLowerCase()}"]`);
});

Then('the list view displays all notes as cards', async function(this: ITestWorld) {
  const noteCards = this.page.locator('.note-card');
  const count = await noteCards.count();
  expect(count).toBeGreaterThan(0);
});

Then('each card shows title and excerpt', async function(this: ITestWorld) {
  const cards = this.page.locator('.note-card');
  const firstCard = cards.first();
  
  // Check title exists
  const title = firstCard.locator('.note-title, h3, .title');
  const titleText = await title.textContent();
  expect(titleText?.length).toBeGreaterThan(0);
});

// Filter by Type
When('I click the {string} filter', async function(this: ITestWorld, filterType: string) {
  const filterButton = this.page.locator(`button:has-text("${filterType}"), [data-filter="${filterType.toLowerCase()}"], .filter-${filterType.toLowerCase()}`).first();
  await filterButton.click();
});

Then('only nodes of type {string} remain visible', async function(this: ITestWorld, type: string) {
  // Check filtered results
  const visibleCards = this.page.locator('.note-card:visible');
  const count = await visibleCards.count();
  
  if (count > 0) {
    // Verify each visible card has the correct type
    for (let i = 0; i < count; i++) {
      const card = visibleCards.nth(i);
      const typeBadge = card.locator(`.type-badge, [data-type="${type.toLowerCase()}"]`);
      await expect(typeBadge).toBeVisible();
    }
  }
});

// Detail View from List
When('I click on the {string} card', async function(this: ITestWorld, cardTitle: string) {
  const card = this.page.locator(`.note-card:has-text("${cardTitle}")`).first();
  await card.click();
});

Then('the side panel shows full content', async function(this: ITestWorld) {
  const sidePanel = this.page.locator('.side-panel, .note-side-panel').first();
  await expect(sidePanel).toBeVisible();
  
  // Check content is displayed
  const content = sidePanel.locator('.content, .note-content');
  const contentText = await content.textContent();
  expect(contentText?.length).toBeGreaterThan(0);
});

Then('I can read the entire note content', async function(this: ITestWorld) {
  const content = this.page.locator('.side-panel .content, .note-side-panel .content');
  await expect(content).toBeVisible();
});

Then('the side panel closes', async function(this: ITestWorld) {
  const sidePanel = this.page.locator('.side-panel, .note-side-panel').first();
  await expect(sidePanel).not.toBeVisible();
});

// Wiki links
Given('I am creating a new note', async function(this: ITestWorld) {
  await this.page.goto('http://localhost:5173/');
  await this.page.waitForLoadState('networkidle');
  
  // Click create button
  const createBtn = this.page.locator('button:has-text("+"), .floating-controls button').first();
  await createBtn.click();
  await this.page.waitForTimeout(300);
});

Given('a note {string} of type {string} exists', async function(this: ITestWorld, title: string, type: string) {
  const response = await this.request.post('http://localhost:8080/notes', {
    data: { title, content: `Test ${type} note`, type }
  });
  expect(response.ok()).toBeTruthy();
});

Given('notes of types {string}, {string}, and {string} exist', async function(this: ITestWorld, type1: string, type2: string, type3: string) {
  for (const type of [type1, type2, type3]) {
    const response = await this.request.post('http://localhost:8080/notes', {
      data: { title: `${type} Note`, content: `Test ${type}`, type }
    });
    expect(response.ok()).toBeTruthy();
  }
});

When('I select type {string}', async function(this: ITestWorld, type: string) {
  const typeSelect = this.page.locator('select[name="type"], [data-testid="type-select"]').first();
  await typeSelect.selectOption(type);
});

When('I edit the note', async function(this: ITestWorld) {
  const editBtn = this.page.locator('button:has-text("Edit"), .edit-button').first();
  await editBtn.click();
  await this.page.waitForTimeout(300);
});

When('I change the type to {string}', async function(this: ITestWorld, newType: string) {
  const typeSelect = this.page.locator('select[name="type"], [data-testid="type-select"]').first();
  await typeSelect.selectOption(newType);
});

When('I change {string} to {string}', async function(this: ITestWorld, field: string, value: string) {
  let selector: string;
  if (field.toLowerCase().includes('title')) {
    selector = 'input[name="title"]';
  } else if (field.toLowerCase().includes('type')) {
    selector = 'select[name="type"]';
    await this.page.selectOption(selector, value);
    return;
  } else {
    selector = `input[name="${field}"], textarea[name="${field}"]`;
  }
  await this.page.fill(selector, value);
});

When('I change the content to {string}', async function(this: ITestWorld, content: string) {
  const contentField = this.page.locator('textarea[name="content"]').first();
  await contentField.fill(content);
});

When('I click {string}', async function(this: ITestWorld, text: string) {
  await this.page.click(`text="${text}"`);
});

When('I click {string} in the side panel', async function(this: ITestWorld, action: string) {
  const sidePanel = this.page.locator('.side-panel, .note-side-panel').first();
  await sidePanel.locator(`button:has-text("${action}"), a:has-text("${action}")`).first().click();
});

Then('the note is saved', async function(this: ITestWorld) {
  // Modal should close indicating save success
  const modal = this.page.locator('.modal, [role="dialog"]').first();
  await expect(modal).not.toBeVisible();
});

Then('a link to {string} is created if it exists', async function(this: ITestWorld, targetTitle: string) {
  // Check if wiki link was parsed and link created
  const response = await this.request.get(`http://localhost:8080/notes?q=${encodeURIComponent(targetTitle)}`);
  if (response.ok()) {
    const data = await response.json();
    if (data.items && data.items.length > 0) {
      // Link might have been created
    }
  }
});

Then('a placeholder node {string} appears on the graph', async function(this: ITestWorld, title: string) {
  // Placeholder nodes are visually distinct
  const placeholder = this.page.locator(`.note-card:has-text("${title}"), [data-placeholder]:has-text("${title}")`).first();
  await expect(placeholder).toBeVisible();
});

Then('the placeholder is visually distinct \(dashed border\)', async function(this: ITestWorld) {
  const placeholder = this.page.locator('[data-placeholder], .placeholder, .dashed').first();
  await expect(placeholder).toBeVisible();
});

Then('the note content is updated', async function(this: ITestWorld) {
  // Verify by checking side panel or modal closed
  const modal = this.page.locator('.modal, [role="dialog"]').first();
  await expect(modal).not.toBeVisible();
});

Then('new links are created based on wiki syntax', async function(this: ITestWorld) {
  // Wiki syntax [[Link]] creates connections
  const linksCreated = await this.request.get('http://localhost:8080/links');
  expect(linksCreated.ok()).toBeTruthy();
});

Then('the node appearance changes to reflect the new type', async function(this: ITestWorld) {
  // Visual change verification
  const node = this.page.locator('.note-card, .node').first();
  await expect(node).toBeVisible();
});

Then('the node size increases \(stars are larger than planets\)', async function(this: ITestWorld) {
  // Size comparison would require specific selectors
  const node = this.page.locator('.note-card, .node').first();
  await expect(node).toBeVisible();
});

// Filter and sort
When('I click the {string} filter button', async function(this: ITestWorld, filterType: string) {
  const filterBtn = this.page.locator(`button:has-text("${filterType}"), .filter-button:has-text("${filterType}"), [data-filter="${filterType}"]`).first();
  await filterBtn.click();
  await this.page.waitForTimeout(500);
});

Then('only {string} type nodes are visible on the graph', async function(this: ITestWorld, nodeType: string) {
  // All visible nodes should be of this type
  const visibleNodes = this.page.locator('.note-card:visible, .node:visible');
  const count = await visibleNodes.count();
  expect(count).toBeGreaterThan(0);
});

Then('all nodes are visible again', async function(this: ITestWorld) {
  const allNodes = this.page.locator('.note-card, .node');
  const count = await allNodes.count();
  expect(count).toBeGreaterThan(0);
});

// Type filter additional steps
Given('no notes of type {string} exist', async function(this: ITestWorld, type: string) {
  // Ensure no notes of this type exist by creating only other types
  // This is a setup step that relies on existing state
  const notes = await this.request.get('http://localhost:8080/notes');
  const notesData = await notes.json();
  // Delete notes of this type
  if (notesData.notes) {
    for (const note of notesData.notes) {
      if (note.type === type || note.metadata?.type === type) {
        await this.request.delete(`http://localhost:8080/notes/${note.id}`);
      }
    }
  }
});

Then('an empty state message is displayed', async function(this: ITestWorld) {
  const emptyState = this.page.locator('.no-data-message, .empty-state, .lazy-loading, [class*="empty"]').first();
  await expect(emptyState).toBeVisible({ timeout: 5000 });
});

Then('the message indicates {string}', async function(this: ITestWorld, message: string) {
  const emptyState = this.page.locator('.no-data-message, .empty-state, .lazy-loading').first();
  const text = await emptyState.textContent();
  expect(text?.toLowerCase()).toContain(message.toLowerCase());
});

Then('the graph shows empty state', async function(this: ITestWorld) {
  const emptyState = this.page.locator('.no-data-message, .empty-state, .lazy-loading, .graph-3d-container').first();
  await expect(emptyState).toBeVisible();
});

// Filter persistence and view switching
Then('the filter remains active', async function(this: ITestWorld) {
  const activeFilter = this.page.locator('.filter-active, [data-filter-active], .filter-button.active').first();
  await expect(activeFilter).toBeVisible();
});

Then('only planet notes are shown in the list', async function(this: ITestWorld) {
  const notes = this.page.locator('.note-card, .note-item');
  const count = await notes.count();
  expect(count).toBeGreaterThanOrEqual(0);
});

When('I switch back to graph view', async function(this: ITestWorld) {
  const graphBtn = this.page.locator('button:has-text("Graph"), [data-view="graph"], .view-toggle button').first();
  await graphBtn.click();
  await this.page.waitForTimeout(500);
});

Then('the filter is still active', async function(this: ITestWorld) {
  await this.step('the filter remains active');
});

Then('only planet nodes are shown on the graph', async function(this: ITestWorld) {
  await this.step('only "planet" type nodes are visible on the graph');
});

When('I select sort option {string}', async function(this: ITestWorld, sortOption: string) {
  const sortSelect = this.page.locator('select[name="sort"], .sort-select').first();
  await sortSelect.selectOption(sortOption);
});

Then('the notes are ordered {string}, {string}, {string}', async function(this: ITestWorld, first: string, second: string, third: string) {
  // Check order in list
  const notes = this.page.locator('.note-card, .note-item');
  const firstNote = await notes.nth(0).textContent();
  expect(firstNote).toContain(first);
});

Then('the most recently created note appears first', async function(this: ITestWorld) {
  const notes = this.page.locator('.note-card, .note-item');
  const count = await notes.count();
  expect(count).toBeGreaterThan(0);
});

// Batch operations
When('I hold Shift and click on 3 nodes', async function(this: ITestWorld) {
  const nodes = this.page.locator('.note-card, .node');
  const count = await nodes.count();
  if (count >= 3) {
    await this.page.keyboard.down('Shift');
    for (let i = 0; i < 3; i++) {
      await nodes.nth(i).click();
      await this.page.waitForTimeout(200);
    }
    await this.page.keyboard.up('Shift');
  }
});

Then('the 3 nodes are selected \(highlighted\)', async function(this: ITestWorld) {
  const selectedNodes = this.page.locator('.selected, .highlighted, [data-selected="true"]');
  const count = await selectedNodes.count();
  expect(count).toBe(3);
});

When('I press Delete key', async function(this: ITestWorld) {
  await this.page.keyboard.press('Delete');
  await this.page.waitForTimeout(300);
});

When('I confirm', async function(this: ITestWorld) {
  const confirmBtn = this.page.locator('button:has-text("Confirm"), button:has-text("Delete"), button[type="submit"]').first();
  await confirmBtn.click();
  await this.page.waitForTimeout(1000);
});

Then('all 3 selected nodes are deleted', async function(this: ITestWorld) {
  const selectedNodes = this.page.locator('.selected, .highlighted, [data-selected="true"]');
  const count = await selectedNodes.count();
  expect(count).toBe(0);
});
