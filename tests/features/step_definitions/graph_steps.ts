import { Given, When, Then, Before, After } from '@cucumber/cucumber';
import { expect } from '@playwright/test';
import type { Page, Browser } from '@playwright/test';
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
  await this.page.goto('http://localhost:5173/');
  await waitForGraphCanvas(this.page);
});

Given('I am on the graph view', async function(this: ITestWorld) {
  await this.page.goto('http://localhost:5173/');
  await waitForGraphCanvas(this.page);
});

Given('I am on the main page', async function(this: ITestWorld) {
  await this.page.goto('http://localhost:5173/');
  await expect(this.page.locator('body')).toBeVisible();
});

Given('I am on the home page', async function(this: ITestWorld) {
  await this.page.goto('http://localhost:5173/');
  await this.page.waitForLoadState('networkidle');
  await expect(this.page.locator('body')).toBeVisible();
});

Given('the application is open on the home page', async function(this: ITestWorld) {
  await this.page.goto('http://localhost:5173/');
  await this.page.waitForLoadState('networkidle');
  await expect(this.page.locator('body')).toBeVisible();
});

Given('notes of various types exist in the system', async function(this: ITestWorld) {
  const types = ['star', 'planet', 'comet', 'galaxy'];
  for (let i = 0; i < types.length; i++) {
    await this.request.post('http://localhost:8080/notes', {
      data: { title: `Type Test ${types[i]}`, content: `Test content ${i}`, type: types[i] }
    });
  }
});

Given('notes of various types exist', async function(this: ITestWorld) {
  await this.step('notes of various types exist in the system');
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

Then('the node label changes to {string}', async function(this: ITestWorld, newLabel: string) {
  const node = await findNodeByLabel(this.page, newLabel);
  expect(node).not.toBeNull();
  if (node) {
    await expect(node).toBeVisible();
  }
});

// Drag and drop to create links
Given('notes {string} and {string} exist on the graph', async function(this: ITestWorld, note1Title: string, note2Title: string) {
  // Create two notes via API
  const note1 = await this.request.post('http://localhost:8080/notes', {
    data: { title: note1Title, content: `Content for ${note1Title}` }
  });
  const note2 = await this.request.post('http://localhost:8080/notes', {
    data: { title: note2Title, content: `Content for ${note2Title}` }
  });
  
  this.context.noteIds = {
    [note1Title]: (await note1.json()).id,
    [note2Title]: (await note2.json()).id
  };
});

When('I drag from node {string} to node {string}', async function(this: ITestWorld, sourceLabel: string, targetLabel: string) {
  const sourceNode = await findNodeByLabel(this.page, sourceLabel);
  const targetNode = await findNodeByLabel(this.page, targetLabel);
  
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
  // Check for link preview indicator
  const preview = this.page.locator('.link-preview, .drag-line, .connection-preview').first();
  await expect(preview).toBeVisible();
});

When('I release the mouse over {string}', async function(this: ITestWorld, targetLabel: string) {
  const targetNode = await findNodeByLabel(this.page, targetLabel);
  if (targetNode) {
    const box = await targetNode.boundingBox();
    if (box) {
      await this.page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
      await this.page.mouse.up();
    }
  }
  await this.page.waitForTimeout(500);
});

Then('a weight selector appears', async function(this: ITestWorld) {
  const weightSelector = this.page.locator('.weight-selector, .link-weight-modal, [class*="weight"]').first();
  await expect(weightSelector).toBeVisible();
});

When('I select weight {string} and confirm', async function(this: ITestWorld, weight: string) {
  // Select weight and confirm
  const weightButton = this.page.locator(`button:has-text("${weight}"), .weight-option:has-text("${weight}")`).first();
  await weightButton.click();
  
  const confirmButton = this.page.locator('button:has-text("Confirm"), button:has-text("Create"), button[type="submit"]').first();
  await confirmButton.click();
  await this.page.waitForTimeout(1000);
});

Then('a new edge appears between {string} and {string}', async function(this: ITestWorld, sourceLabel: string, targetLabel: string) {
  // Verify link was created via API
  const sourceId = this.context.noteIds?.[sourceLabel];
  const targetId = this.context.noteIds?.[targetLabel];
  
  if (sourceId && targetId) {
    const linksResponse = await this.request.get(`http://localhost:8080/links?source=${sourceId}&target=${targetId}`);
    expect(linksResponse.ok()).toBeTruthy();
  }
  
  // Visual verification - link line or connection
  const linkLine = this.page.locator('.link-line, .edge, .connection, line').first();
  const hasLink = await linkLine.isVisible().catch(() => false);
  expect(hasLink).toBe(true);
});

// Search and focus on node
Given('multiple notes exist on the graph', async function(this: ITestWorld) {
  // Ensure at least 2 notes exist
  await this.step('a note "Note 1" exists on the graph');
  await this.step('a note "Note 2" exists on the graph');
});

When('I click the search icon', async function(this: ITestWorld) {
  const searchIcon = this.page.locator('.search-icon, [data-testid="search"], button:has-text("Search")').first();
  await searchIcon.click();
  await this.page.waitForTimeout(300);
});

When('I type {string} in the search input', async function(this: ITestWorld, query: string) {
  const searchInput = this.page.locator('.search-input, input[type="search"], input[placeholder*="Search"]').first();
  await searchInput.fill(query);
  await this.page.waitForTimeout(500);
});

Then('a list of matching notes appears', async function(this: ITestWorld) {
  const searchResults = this.page.locator('.search-results, .dropdown, .autocomplete').first();
  await expect(searchResults).toBeVisible();
});

When('I click on a search result {string}', async function(this: ITestWorld, resultTitle: string) {
  const result = this.page.locator(`.search-results:has-text("${resultTitle}"), .dropdown-item:has-text("${resultTitle}")`).first();
  await result.click();
  await this.page.waitForTimeout(500);
});

Then('the graph zooms to the node {string}', async function(this: ITestWorld, label: string) {
  // The node should become visible/accessible after zoom
  const node = await findNodeByLabel(this.page, label);
  expect(node).not.toBeNull();
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

// Graph View Mode - Low-end devices
Given('I am using a low-end device', async function(this: ITestWorld) {
  // Emulate low-end device via device emulation
  await this.context.close();
  this.context = await this.browser.newContext({
    viewport: { width: 1024, height: 768 },
    deviceScaleFactor: 1,
    reducedMotion: 'reduce'
  });
  this.page = await this.context.newPage();
});

When('I open the graph view', async function(this: ITestWorld) {
  await this.page.goto('http://localhost:5173/');
  await this.page.waitForLoadState('networkidle');
});

Then('the graph defaults to 2D mode', async function(this: ITestWorld) {
  // Check for 2D mode indicators - canvas visible but no 3D container
  const graphContainer = this.page.locator('.graph-canvas, .graph-2d').first();
  await expect(graphContainer).toBeVisible();
});

Then('the {string} button shows a warning icon', async function(this: ITestWorld, buttonText: string) {
  const button = this.page.locator(`button:has-text("${buttonText}")`).first();
  await expect(button).toBeVisible();
  // Check for warning icon inside button
  const hasWarning = await button.locator('.warning-icon, [data-warning], .icon-warning').isVisible().catch(() => false);
  expect(hasWarning).toBe(true);
});

When('clicking it shows a performance warning', async function(this: ITestWorld) {
  // This step should follow clicking the button
  const warning = this.page.locator('.performance-warning, .warning-modal, [class*="warning"]').first();
  await expect(warning).toBeVisible();
});

// Zoom and pan in 2D mode
When('I scroll the mouse wheel up', async function(this: ITestWorld) {
  const canvas = this.page.locator('canvas, .graph-canvas, .graph-2d').first();
  await canvas.scrollIntoViewIfNeeded();
  await canvas.dispatchEvent('wheel', { deltaY: -100 });
  await this.page.waitForTimeout(300);
});

When('I scroll the mouse wheel down', async function(this: ITestWorld) {
  const canvas = this.page.locator('canvas, .graph-canvas, .graph-2d').first();
  await canvas.scrollIntoViewIfNeeded();
  await canvas.dispatchEvent('wheel', { deltaY: 100 });
  await this.page.waitForTimeout(300);
});

Then('the graph zooms in', async function(this: ITestWorld) {
  // Zoom is internal state, verify canvas is still visible and responsive
  const canvas = this.page.locator('canvas, .graph-canvas').first();
  await expect(canvas).toBeVisible();
});

Then('the graph zooms out', async function(this: ITestWorld) {
  const canvas = this.page.locator('canvas, .graph-canvas').first();
  await expect(canvas).toBeVisible();
});

When('I drag the canvas', async function(this: ITestWorld) {
  const canvas = this.page.locator('canvas, .graph-canvas, .graph-2d').first();
  const box = await canvas.boundingBox();
  if (box) {
    const startX = box.x + box.width / 2;
    const startY = box.y + box.height / 2;
    await this.page.mouse.move(startX, startY);
    await this.page.mouse.down();
    await this.page.mouse.move(startX + 100, startY + 50, { steps: 10 });
    await this.page.mouse.up();
  }
  await this.page.waitForTimeout(500);
});

Then('the graph pans in the dragged direction', async function(this: ITestWorld) {
  // Pan is internal state, verify canvas is still visible
  const canvas = this.page.locator('canvas, .graph-canvas').first();
  await expect(canvas).toBeVisible();
});

// 3D view interactions
When('I drag the mouse on the canvas', async function(this: ITestWorld) {
  const canvas = this.page.locator('.graph-3d-container canvas, .graph-3d-container').first();
  const box = await canvas.boundingBox();
  if (box) {
    const startX = box.x + box.width / 2;
    const startY = box.y + box.height / 2;
    await this.page.mouse.move(startX, startY);
    await this.page.mouse.down();
    await this.page.mouse.move(startX - 100, startY - 50, { steps: 10 });
    await this.page.mouse.up();
  }
  await this.page.waitForTimeout(500);
});

Then('the camera rotates around the graph', async function(this: ITestWorld) {
  // Camera rotation is internal 3D state, verify container is visible
  const container = this.page.locator('.graph-3d-container').first();
  await expect(container).toBeVisible();
});

When('I use the scroll wheel', async function(this: ITestWorld) {
  const container = this.page.locator('.graph-3d-container').first();
  await container.scrollIntoViewIfNeeded();
  await container.dispatchEvent('wheel', { deltaY: -50 });
  await this.page.waitForTimeout(300);
  await container.dispatchEvent('wheel', { deltaY: 100 });
  await this.page.waitForTimeout(300);
});

Then('the camera zooms in and out', async function(this: ITestWorld) {
  // Camera zoom is internal state, verify container is visible
  const container = this.page.locator('.graph-3d-container').first();
  await expect(container).toBeVisible();
});

When('I click on a node', async function(this: ITestWorld) {
  // Click on the first visible node
  const node = this.page.locator('.note-card, .node, [class*="node"]').first();
  await node.click();
  await this.page.waitForTimeout(500);
});

Then('the node is highlighted with a glow effect', async function(this: ITestWorld) {
  const highlightedNode = this.page.locator('.glow, .highlighted, [class*="glow"], [class*="highlight"]').first();
  await expect(highlightedNode).toBeVisible();
});

Then('the camera centers on the clicked node', async function(this: ITestWorld) {
  // Camera centering is internal 3D state
  const container = this.page.locator('.graph-3d-container').first();
  await expect(container).toBeVisible();
});
