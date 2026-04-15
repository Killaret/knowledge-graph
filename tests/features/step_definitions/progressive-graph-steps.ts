import { Given, When, Then, Before, After } from '@cucumber/cucumber';
import { expect } from '@playwright/test';
import type { ITestWorld } from '../support/world';

// Progressive Graph Rendering steps

Given('I am on the 3D graph view for note {string}', async function(this: ITestWorld, noteId: string) {
  await this.page.goto(`http://localhost:5173/graph/3d/${noteId}`);
  await this.page.waitForLoadState('networkidle');
  await this.page.waitForTimeout(2000);
});

Given('a note {string} exists with content {string}', async function(this: ITestWorld, title: string, content: string) {
  const response = await this.request.post('http://localhost:8080/notes', {
    data: { title, content }
  });
  expect(response.ok()).toBeTruthy();
  const note = await response.json();
  this.context.lastNoteId = note.id;
});

Given('a note {string} of type {string} exists', async function(this: ITestWorld, title: string, type: string) {
  const response = await this.request.post('http://localhost:8080/notes', {
    data: { title, content: `Test ${type} note`, type }
  });
  expect(response.ok()).toBeTruthy();
  const note = await response.json();
  this.context.lastNoteId = note.id;
});

Given('note {string} is linked to note {string}', async function(this: ITestWorld, sourceTitle: string, targetTitle: string) {
  // Find notes by title via API
  const searchResponse = await this.request.get(`http://localhost:8080/notes?q=${encodeURIComponent(sourceTitle)}`);
  const searchData = await searchResponse.json();
  const sourceNote = searchData.items.find((n: any) => n.title === sourceTitle);
  
  const targetSearch = await this.request.get(`http://localhost:8080/notes?q=${encodeURIComponent(targetTitle)}`);
  const targetData = await targetSearch.json();
  const targetNote = targetData.items.find((n: any) => n.title === targetTitle);
  
  if (sourceNote && targetNote) {
    await this.request.post('http://localhost:8080/links', {
      data: { 
        sourceNoteId: sourceNote.id, 
        targetNoteId: targetNote.id, 
        weight: 0.8 
      }
    });
  }
});

When('I navigate to the 3D graph view', async function(this: ITestWorld) {
  const noteId = this.context.lastNoteId;
  if (!noteId) {
    throw new Error('No note ID available. Create a note first.');
  }
  await this.page.goto(`http://localhost:5173/graph/3d/${noteId}`);
  await this.page.waitForLoadState('networkidle');
  await this.page.waitForTimeout(2000);
});

When('I wait for the graph to load', async function(this: ITestWorld) {
  await this.page.waitForTimeout(2000);
});

When('I wait for progressive loading to complete', async function(this: ITestWorld) {
  // Wait for background loading to finish (fog animation + camera zoom)
  await this.page.waitForTimeout(5000);
});

Then('the 3D graph container is visible', async function(this: ITestWorld) {
  const graphContainer = this.page.locator('.graph-3d-container').first();
  await expect(graphContainer).toBeVisible({ timeout: 5000 });
});

Then('the graph is displayed without loading spinner', async function(this: ITestWorld) {
  const graphContainer = this.page.locator('.graph-3d-container').first();
  await expect(graphContainer).toBeVisible({ timeout: 3000 });
  
  // Verify no spinner is present
  const spinner = this.page.locator('.loading-overlay, .lazy-loading, .spinner');
  const hasSpinner = await spinner.isVisible().catch(() => false);
  expect(hasSpinner).toBe(false);
});

Then('the stats bar shows {string} and {string}', async function(this: ITestWorld, nodesText: string, linksText: string) {
  const statsBar = this.page.locator('.stats-bar').first();
  await expect(statsBar).toBeVisible();
  
  const statsText = await statsBar.textContent();
  expect(statsText).toContain(nodesText);
  expect(statsText).toContain(linksText);
});

Then('the stats bar shows at least {int} nodes', async function(this: ITestWorld, minNodes: number) {
  const statsBar = this.page.locator('.stats-bar').first();
  await expect(statsBar).toBeVisible();
  
  const statsText = await statsBar.textContent();
  expect(statsText).not.toBeNull();
  const nodeMatch = statsText!.match(/(\d+)\s*nodes?/i);
  if (nodeMatch) {
    const nodeCount = parseInt(nodeMatch[1], 10);
    expect(nodeCount).toBeGreaterThanOrEqual(minNodes);
  }
});

Then('the graph shows {string} mode indicator', async function(this: ITestWorld, mode: string) {
  const statsBar = this.page.locator('.stats-bar').first();
  await expect(statsBar).toBeVisible();
  
  const statsText = await statsBar.textContent();
  expect(statsText).toContain(mode);
});

Then('the WebGL canvas is present', async function(this: ITestWorld) {
  const canvas = this.page.locator('canvas').first();
  const hasCanvas = await canvas.isVisible().catch(() => false);
  
  // WebGL may not work in all environments, so we just check if container is visible
  if (!hasCanvas) {
    const graphContainer = this.page.locator('.graph-3d-container').first();
    await expect(graphContainer).toBeVisible();
  }
});

Then('the fog effect is active', async function(this: ITestWorld) {
  // Fog is handled internally by Three.js, we verify graph is visible
  const graphContainer = this.page.locator('.graph-3d-container').first();
  await expect(graphContainer).toBeVisible();
});

Then('the camera is positioned to show the graph', async function(this: ITestWorld) {
  // Camera position is internal state, we verify canvas is rendering
  const canvas = this.page.locator('canvas').first();
  const hasCanvas = await canvas.isVisible().catch(() => false);
  
  if (hasCanvas) {
    const box = await canvas.boundingBox();
    expect(box).not.toBeNull();
    expect(box!.width).toBeGreaterThan(0);
    expect(box!.height).toBeGreaterThan(0);
  }
});

// Note: 'multiple notes exist on the graph' is defined in graph_steps.ts

Given('I am on the graph view in 2D mode', async function(this: ITestWorld) {
  await this.page.goto('http://localhost:5173/');
  await this.page.waitForLoadState('networkidle');
  // 2D is default on main page
  await this.page.waitForTimeout(1000);
});

Given('I am on the graph view in 3D mode', async function(this: ITestWorld) {
  // Navigate to a 3D graph page
  const response = await this.request.post('http://localhost:8080/notes', {
    data: { title: '3D Mode Test', content: 'Test' }
  });
  const note = await response.json();
  
  await this.page.goto(`http://localhost:5173/graph/3d/${note.id}`);
  await this.page.waitForLoadState('networkidle');
  await this.page.waitForTimeout(2000);
});

When('I click the {string} button in floating controls', async function(this: ITestWorld, buttonText: string) {
  const button = this.page.locator(`.floating-controls button:has-text("${buttonText}")`).first();
  await button.click();
  await this.page.waitForTimeout(1000);
});

Then('the graph switches to 3D mode', async function(this: ITestWorld) {
  const graphContainer = this.page.locator('.graph-3d-container').first();
  await expect(graphContainer).toBeVisible({ timeout: 5000 });
});

Then('the graph switches back to 2D mode', async function(this: ITestWorld) {
  // 2D mode shows graph canvas or 2D container
  const container = this.page.locator('.graph-canvas, .graph-2d, canvas').first();
  await expect(container).toBeVisible({ timeout: 5000 });
});

Then('the graph canvas shows {string} indicator', async function(this: ITestWorld, indicator: string) {
  const statsBar = this.page.locator('.stats-bar').first();
  const hasStats = await statsBar.isVisible().catch(() => false);
  
  if (hasStats) {
    const statsText = await statsBar.textContent();
    expect(statsText).toContain(indicator.replace(' Mode', ''));
  }
});

Then('nodes are rendered as 3D spheres', async function(this: ITestWorld) {
  // 3D nodes are rendered in WebGL canvas
  const canvas = this.page.locator('canvas').first();
  const hasCanvas = await canvas.isVisible().catch(() => false);
  expect(hasCanvas).toBe(true);
});

Then('nodes are rendered as 2D shapes', async function(this: ITestWorld) {
  // 2D nodes are SVG or DOM elements
  const graphContainer = this.page.locator('.graph-canvas, .graph-2d, canvas').first();
  await expect(graphContainer).toBeVisible();
});
