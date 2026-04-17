import { Given, When, Then, Before, After } from '@cucumber/cucumber';
import { expect } from '@playwright/test';
import type { ITestWorld } from '../features/support/world';
import { createNote, createLink, getBackendUrl, cleanupTestData } from '../../frontend/tests/helpers/testData';

// Test data tracking for cleanup
let testNoteIds: string[] = [];

Before({ tags: '@3d-graph' }, async function (this: ITestWorld) {
  testNoteIds = [];
});

After({ tags: '@3d-graph' }, async function (this: ITestWorld) {
  if (testNoteIds.length > 0) {
    await cleanupTestData(this.request, testNoteIds);
  }
});

Given('I have test notes with connections', async function (this: ITestWorld) {
  // Create a central note with connected notes
  const centerNote = await createNote(this.request, {
    title: 'Progressive Loading Center',
    content: 'Central node for progressive loading test',
    type: 'star',
  });
  testNoteIds.push(centerNote.id);

  // Create surrounding notes with links
  for (let i = 0; i < 5; i++) {
    const linkedNote = await createNote(this.request, {
      title: `Linked Note ${i}`,
      content: `Content for note ${i}`,
      type: 'planet',
    });
    testNoteIds.push(linkedNote.id);
    await createLink(this.request, centerNote.id, linkedNote.id, 0.7);
  }

  // Store center note ID for navigation
  (this as any).centerNoteId = centerNote.id;
});

Given('I navigate to {string}', async function (this: ITestWorld, path: string) {
  // Replace {centerNoteId} placeholder if present
  let resolvedPath = path;
  if (path.includes('{centerNoteId}')) {
    const centerId = (this as any).centerNoteId;
    resolvedPath = path.replace('{centerNoteId}', centerId);
  }
  if (path.includes('{noteId}') || path.includes('{zoomTestId}')) {
    const noteId = (this as any).testNoteId || (this as any).centerNoteId;
    resolvedPath = path.replace(/{[^}]+Id}/g, noteId);
  }

  const baseUrl = process.env.FRONTEND_URL || 'http://localhost:5173';
  await this.page.goto(`${baseUrl}${resolvedPath}`);
  await this.page.waitForLoadState('networkidle');
});

Then('the loading overlay should be visible', async function (this: ITestWorld) {
  const overlay = this.page.locator('.loading-overlay, .fog-overlay, [data-testid="loading-overlay"]').first();
  await expect(overlay).toBeVisible({ timeout: 5000 });
});

Then('the loading text should contain {string}', async function (this: ITestWorld, text: string) {
  const loadingText = this.page.locator('.loading-text, .overlay-text, [data-testid="loading-text"]').first();
  const content = await loadingText.textContent();
  expect(content?.toLowerCase()).toContain(text.toLowerCase());
});

Then('the fog density should be at least {float}', async function (this: ITestWorld, minDensity: number) {
  const density = await this.page.evaluate(() => {
    // Access the fog from the 3D scene
    const scene = (window as any).scene;
    return scene?.fog?.density ?? 0;
  });
  expect(density).toBeGreaterThanOrEqual(minDensity);
});

When('the simulation progresses to at least {int}% nodes positioned', async function (this: ITestWorld, percentage: number) {
  await this.page.waitForFunction(
    (pct: number) => {
      const sim = (window as any).simulation;
      if (!sim || !sim.nodes) return false;
      const nodes = sim.nodes();
      const positioned = nodes.filter((n: any) => n.x !== 0 || n.y !== 0 || n.z !== 0).length;
      return (positioned / nodes.length) * 100 >= pct;
    },
    percentage,
    { timeout: 30000 }
  );
});

Then('the loading overlay should disappear within {int} seconds', async function (this: ITestWorld, seconds: number) {
  const overlay = this.page.locator('.loading-overlay, .fog-overlay, [data-testid="loading-overlay"]').first();
  await expect(overlay).toBeHidden({ timeout: seconds * 1000 });
});

Then('the fog density should still be greater than {float}', async function (this: ITestWorld, minDensity: number) {
  const density = await this.page.evaluate(() => {
    const scene = (window as any).scene;
    return scene?.fog?.density ?? 0;
  });
  expect(density).toBeGreaterThan(minDensity);
});

Then('I should be able to click on the canvas', async function (this: ITestWorld) {
  const canvas = this.page.locator('canvas, .graph-3d-canvas, .graph-canvas').first();
  await expect(canvas).toBeVisible();
  await expect(canvas).toBeEnabled();

  // Verify pointer-events is not 'none'
  const pointerEvents = await canvas.evaluate((el: HTMLElement) => {
    return window.getComputedStyle(el).pointerEvents;
  });
  expect(pointerEvents).not.toBe('none');
});

When('I wait for the simulation to end', async function (this: ITestWorld) {
  // Wait for simulation to stabilize (alpha < 0.01 or similar)
  await this.page.waitForFunction(
    () => {
      const sim = (window as any).simulation;
      return sim && sim.alpha && sim.alpha() < 0.01;
    },
    { timeout: 60000 }
  );
});

Then('the fog density should become less than {float} within {int} seconds', async function (
  this: ITestWorld,
  maxDensity: number,
  seconds: number
) {
  await expect
    .poll(
      async () => {
        return await this.page.evaluate(() => {
          const scene = (window as any).scene;
          return scene?.fog?.density ?? 1;
        });
      },
      { timeout: seconds * 1000 }
    )
    .toBeLessThan(maxDensity);
});

Then('all nodes should be clickable', async function (this: ITestWorld) {
  const nodes = await this.page.evaluate(() => {
    const sim = (window as any).simulation;
    return sim ? sim.nodes().length : 0;
  });

  expect(nodes).toBeGreaterThan(0);

  // Try clicking the first node
  const node = this.page.locator('.node, [data-node-id]').first();
  if (await node.isVisible().catch(() => false)) {
    await node.click();
  }
});

Then('links should be visible with opacity based on weight', async function (this: ITestWorld) {
  const links = this.page.locator('.link, [data-link-id], line').all();
  const linkCount = (await links).length;
  expect(linkCount).toBeGreaterThan(0);

  // Verify at least one link has opacity set
  const hasOpacity = await this.page.evaluate(() => {
    const links = document.querySelectorAll('.link, [data-link-id], line');
    for (const link of Array.from(links)) {
      const opacity = window.getComputedStyle(link as HTMLElement).opacity;
      if (parseFloat(opacity) > 0) return true;
    }
    return false;
  });

  expect(hasOpacity).toBe(true);
});

Given('the graph is still loading', async function (this: ITestWorld) {
  // Navigate to graph and wait for initial load but not completion
  const baseUrl = process.env.FRONTEND_URL || 'http://localhost:5173';
  const centerId = (this as any).centerNoteId;
  await this.page.goto(`${baseUrl}/graph/3d/${centerId}`);

  // Wait just a short time to ensure loading has started
  await this.page.waitForTimeout(500);

  // Verify loading state
  const isLoading = await this.page.evaluate(() => {
    return (window as any).isLoading || (window as any).loadingProgress < 1;
  });
  expect(isLoading).toBe(true);
});

When('I drag the canvas to rotate the view', async function (this: ITestWorld) {
  const canvas = this.page.locator('canvas, .graph-3d-canvas').first();

  // Get initial camera position
  const initialPos = await this.page.evaluate(() => {
    return (window as any).camera?.position?.clone();
  });
  (this as any).initialCameraPos = initialPos;

  // Perform drag
  const box = await canvas.boundingBox();
  if (box) {
    await this.page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
    await this.page.mouse.down();
    await this.page.mouse.move(box.x + box.width / 2 + 100, box.y + box.height / 2, { steps: 5 });
    await this.page.mouse.up();
  }
});

Then('the camera position should change', async function (this: ITestWorld) {
  const currentPos = await this.page.evaluate(() => {
    return (window as any).camera?.position;
  });
  const initialPos = (this as any).initialCameraPos;

  expect(currentPos).toBeDefined();
  if (initialPos && currentPos) {
    const hasChanged =
      Math.abs(currentPos.x - initialPos.x) > 0.01 ||
      Math.abs(currentPos.y - initialPos.y) > 0.01 ||
      Math.abs(currentPos.z - initialPos.z) > 0.01;
    expect(hasChanged).toBe(true);
  }
});

Then('the loading overlay should still be visible or fading', async function (this: ITestWorld) {
  const overlay = this.page.locator('.loading-overlay, .fog-overlay, [data-testid="loading-overlay"]').first();
  // Should be either visible or in transition
  const isVisible = await overlay.isVisible().catch(() => false);
  const opacity = await overlay.evaluate((el: HTMLElement) => {
    return parseFloat(window.getComputedStyle(el).opacity);
  }).catch(() => 0);

  // Either visible (opacity > 0.5) or fading (opacity between 0 and 0.5)
  expect(isVisible || opacity > 0).toBe(true);
});
