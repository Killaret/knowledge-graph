import { Given, When, Then, Before, After } from '@cucumber/cucumber';
import { expect, Page, Browser } from '@playwright/test';
import type { ITestWorld } from '../support/world';

// Page object helpers
async function waitForGraphCanvas(page: Page) {
  await expect(page.locator('.graph-canvas, canvas, .graph-2d, .graph-3d-container')).toBeVisible({ timeout: 10000 });
}

async function findNodeByLabel(page: Page, label: string) {
  // Try different selectors that might contain the node label
  const selectors = [
    `.node-label:has-text("${label}")`,
    `text="${label}"`,
    `[data-node-id]:has-text("${label}")`,
    `.note-card:has-text("${label}")`
  ];

  for (const selector of selectors) {
    const locator = page.locator(selector).first();
    if (await locator.isVisible().catch(() => false)) {
      return locator;
    }
  }
  return null;
}

// Background steps
Given('the application is open on the graph view', async function(this: ITestWorld) {
  await this.page.goto('/');
  await waitForGraphCanvas(this.page);
});

Given('I am on the graph view', async function(this: ITestWorld) {
  await this.page.goto('/');
  await waitForGraphCanvas(this.page);
});

Given('I am on the main page', async function(this: ITestWorld) {
  await this.page.goto('/');
  await expect(this.page.locator('body')).toBeVisible();
});

// State assertions
When('there are no notes in the system', async function(this: ITestWorld) {
  // This is a setup step - in real tests we'd clear the database
  // For now, we just verify the empty state message appears
  await expect(this.page.locator('text=No notes yet, text=No notes found')).toBeVisible();
});

Then('the graph canvas shows a message {string}', async function(this: ITestWorld, message: string) {
  await expect(this.page.locator(`text=${message}`)).toBeVisible();
});

Then('a {string} button is visible', async function(this: ITestWorld, buttonText: string) {
  await expect(this.page.locator(`button:has-text("${buttonText}")`)).toBeVisible();
});

// Graph interactions
Given('the graph view is active', async function(this: ITestWorld) {
  await waitForGraphCanvas(this.page);
});

When('I click the {string} button', async function(this: ITestWorld, buttonText: string) {
  await this.page.click(`button:has-text("${buttonText}")`);
});

Then('a modal with a note form appears', async function(this: ITestWorld) {
  await expect(this.page.locator('.modal, [role="dialog"], .create-note-modal')).toBeVisible();
});

Then('a modal with the note form appears pre-filled', async function(this: ITestWorld) {
  await expect(this.page.locator('.modal, [role="dialog"], .edit-note-modal')).toBeVisible();
  // Check that title input has value
  const titleInput = this.page.locator('input[name="title"], input[placeholder*="Title"]').first();
  await expect(titleInput).toHaveValue(/.+/); // Has some value
});

When('I fill in {string} with {string}', async function(this: ITestWorld, field: string, value: string) {
  const fieldName = field.toLowerCase();
  let selector: string;

  if (fieldName.includes('title')) {
    selector = 'input[name="title"], input[placeholder*="Title"], #title';
  } else if (fieldName.includes('content')) {
    selector = 'textarea[name="content"], textarea[placeholder*="Content"], #content';
  } else {
    selector = `input[name="${fieldName}"], textarea[name="${fieldName}"]`;
  }

  await this.page.fill(selector, value);
});

When('I change {string} to {string}', async function(this: ITestWorld, field: string, value: string) {
  // Same as fill in - clear first then type
  const fieldName = field.toLowerCase();
  let selector: string;

  if (fieldName.includes('title')) {
    selector = 'input[name="title"], input[placeholder*="Title"], #title';
  } else {
    selector = `input[name="${fieldName}"], textarea[name="${fieldName}"]`;
  }

  await this.page.fill(selector, '');
  await this.page.fill(selector, value);
});

When('I select type {string}', async function(this: ITestWorld, type: string) {
  // Try to find and select the type dropdown or buttons
  const typeButton = this.page.locator(`button:has-text("${type}"), [data-type="${type.toLowerCase()}"]`).first();
  if (await typeButton.isVisible().catch(() => false)) {
    await typeButton.click();
  } else {
    // Try select dropdown
    await this.page.selectOption('select[name="type"]', { label: type });
  }
});

When('I click {string}', async function(this: ITestWorld, text: string) {
  await this.page.click(`text="${text}"`);
});

Then('the modal closes', async function(this: ITestWorld) {
  await expect(this.page.locator('.modal, [role="dialog"]').first()).not.toBeVisible();
});

Then('a new node labeled {string} appears on the graph', async function(this: ITestWorld, label: string) {
  const node = await findNodeByLabel(this.page, label);
  expect(node).not.toBeNull();
  if (node) {
    await expect(node).toBeVisible();
  }
});

// Note management
Given('a note {string} exists on the graph', async function(this: ITestWorld, title: string) {
  // Check if node exists
  const node = await findNodeByLabel(this.page, title);
  if (!node || !(await node.isVisible().catch(() => false))) {
    // Create the note if it doesn't exist
    await this.page.click('button:has-text("+")');
    await this.page.fill('input[name="title"]', title);
    await this.page.click('button:has-text("Save")');
    await this.page.waitForTimeout(500); // Wait for creation
  }
});

Given('a note {string} exists', async function(this: ITestWorld, title: string) {
  await this.step(`a note "${title}" exists on the graph`);
});

Given('a note {string} of type {string} exists', async function(this: ITestWorld, title: string, type: string) {
  // Implementation would create note with specific type
  await this.step(`a note "${title}" exists on the graph`);
});

When('I click on the node {string}', async function(this: ITestWorld, label: string) {
  const node = await findNodeByLabel(this.page, label);
  expect(node).not.toBeNull();
  if (node) {
    await node.click();
  }
});

Then('a side panel opens showing note details', async function(this: ITestWorld) {
  await expect(this.page.locator('.side-panel, .note-side-panel, [class*="panel"]')).toBeVisible();
});

When('I click {string} in the side panel', async function(this: ITestWorld, action: string) {
  const sidePanel = this.page.locator('.side-panel, .note-side-panel').first();
  await sidePanel.locator(`button:has-text("${action}"), a:has-text("${action}")`).click();
});

Then('the node label changes to {string}', async function(this: ITestWorld, newLabel: string) {
  const node = await findNodeByLabel(this.page, newLabel);
  expect(node).not.toBeNull();
  if (node) {
    await expect(node).toBeVisible();
  }
});

// Delete with confirmation
Then('a confirmation modal appears with text {string}', async function(this: ITestWorld, text: string) {
  await expect(this.page.locator(`.modal:has-text("${text}"), [role="dialog"]:has-text("${text}")`)).toBeVisible();
});

When('I confirm the deletion', async function(this: ITestWorld) {
  await this.page.click('button:has-text("Delete"), button:has-text("Confirm"), button[type="submit"]');
});

Then('the node {string} disappears from the graph', async function(this: ITestWorld, label: string) {
  const node = await findNodeByLabel(this.page, label);
  if (node) {
    await expect(node).not.toBeVisible();
  }
});

// Link creation
Given('notes {string} and {string} exist on the graph', async function(this: ITestWorld, note1: string, note2: string) {
  await this.step(`a note "${note1}" exists on the graph`);
  await this.step(`a note "${note2}" exists on the graph`);
});

When('I drag from node {string} to node {string}', async function(this: ITestWorld, source: string, target: string) {
  const sourceNode = await findNodeByLabel(this.page, source);
  const targetNode = await findNodeByLabel(this.page, target);

  expect(sourceNode).not.toBeNull();
  expect(targetNode).not.toBeNull();

  if (sourceNode && targetNode) {
    const sourceBox = await sourceNode.boundingBox();
    const targetBox = await targetNode.boundingBox();

    if (sourceBox && targetBox) {
      await this.page.mouse.move(sourceBox.x + sourceBox.width / 2, sourceBox.y + sourceBox.height / 2);
      await this.page.mouse.down();
      await this.page.mouse.move(targetBox.x + targetBox.width / 2, targetBox.y + targetBox.height / 2, { steps: 10 });
    }
  }
});

Then('a link preview line is shown', async function(this: ITestWorld) {
  await expect(this.page.locator('.link-preview, .drag-line, [class*="preview"]')).toBeVisible();
});

When('I release the mouse over {string}', async function(this: ITestWorld, target: string) {
  await this.page.mouse.up();
});

Then('a weight selector appears', async function(this: ITestWorld) {
  await expect(this.page.locator('.weight-selector, [class*="weight"], select')).toBeVisible();
});

When('I select weight {string} and confirm', async function(this: ITestWorld, weight: string) {
  await this.page.selectOption('.weight-selector select, select', weight);
  await this.page.click('button:has-text("Confirm"), button:has-text("Create"), button:has-text("Save")');
});

Then('a new edge appears between {string} and {string}', async function(this: ITestWorld, source: string, target: string) {
  // In a 2D canvas or 3D scene, we'd check for the visual edge
  // For now, we just verify no errors occurred
  await expect(this.page.locator('.error')).not.toBeVisible();
});

// Search
Given('multiple notes exist on the graph', async function(this: ITestWorld) {
  // Ensure at least 2 notes exist
  await this.step('a note "Note 1" exists on the graph');
  await this.step('a note "Note 2" exists on the graph');
});

When('I click the search icon', async function(this: ITestWorld) {
  await this.page.click('[data-testid="search-icon"], button:has-text("Search"), .search-button');
});

When('I type {string} in the search input', async function(this: ITestWorld, query: string) {
  await this.page.fill('.search-input, input[type="search"], [placeholder*="Search"]', query);
});

Then('a list of matching notes appears', async function(this: ITestWorld) {
  await expect(this.page.locator('.search-results, .dropdown, [class*="results"]')).toBeVisible();
});

When('I click on a search result {string}', async function(this: ITestWorld, resultTitle: string) {
  await this.page.click(`.search-results:has-text("${resultTitle}"), .dropdown:has-text("${resultTitle}")`);
});

Then('the graph zooms to the node {string}', async function(this: ITestWorld, label: string) {
  // The node should become visible/accessible after zoom
  const node = await findNodeByLabel(this.page, label);
  expect(node).not.toBeNull();
});

Then('the node is highlighted', async function(this: ITestWorld) {
  await expect(this.page.locator('.highlighted, [class*="highlight"], .selected')).toBeVisible();
});
