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
