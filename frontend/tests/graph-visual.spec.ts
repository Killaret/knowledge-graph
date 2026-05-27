/**
 * Visual regression tests for graph rendering
 * These tests capture screenshots and compare them with baselines
 * @visual @regression
 */

import { test, expect } from '@playwright/test';
import { createNote, createLink } from './helpers/testUtils';

test.describe.configure({ mode: 'serial' });

test.describe('GraphCanvas Isolated Node Visual Tests @visual @isolated', () => {
  const nodeTypes = ['star', 'planet', 'comet', 'galaxy', 'asteroid'];

  for (const type of nodeTypes) {
    test(`should render ${type} node with correct visual appearance`, async ({ page }) => {
      // Navigate to isolated node test page with longer timeout
      await page.goto(`/test/isolated-node?type=${type}&stableRender=true`, { timeout: 15000, waitUntil: 'networkidle' });
      
      // Wait for page to be fully loaded
      await page.waitForLoadState('domcontentloaded');
      
      // Wait for canvas to be ready with longer timeout
      await page.waitForSelector('canvas', { timeout: 15000 });
      
      // Wait for simulation to stabilize
      await page.waitForTimeout(2000);
      
      // Take screenshot of the canvas
      const canvas = page.locator('canvas').first();
      await expect(canvas).toHaveScreenshot(`${type}-node.png`, {
        maxDiffPixels: 350,
        threshold: 0.3,
        animations: 'disabled'
      });
    });
  }
});

test.describe('GraphCanvas Link Visual Tests @visual @links', () => {
  const linkTypes = ['reference', 'dependency', 'related', 'custom'];

  for (const linkType of linkTypes) {
    test(`should render ${linkType} link between nodes`, async ({ page }) => {
      // Navigate to link pair test page
      await page.goto(`/test/link-pair?linkType=${linkType}&sourceType=star&targetType=planet&stableRender=true`);
      
      // Wait for canvas to be ready
      await page.waitForSelector('canvas', { timeout: 10000 });
      
      // Wait for simulation to stabilize and lines to be drawn
      await page.waitForTimeout(1500);
      
      // Take screenshot of the canvas
      const canvas = page.locator('canvas').first();
      await expect(canvas).toHaveScreenshot(`${linkType}-link.png`, {
        maxDiffPixels: 350,
        threshold: 0.3,
        animations: 'disabled'
      });
    });
  }
});

test.describe('GraphCanvas Visual Tests @visual', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the graph page
    await page.goto('/?stableRender=true');
    await page.waitForSelector('canvas', { timeout: 10000 });
    // Wait for any animations to settle
    await page.waitForTimeout(500);
  });

  test('should render star node with correct visual appearance', async ({ page }) => {
    // Create a star node
    await createNote(page.request, {
      title: 'Test Star',
      content: 'Visual test for star rendering',
      type: 'star'
    });

    // Refresh to see the node
    await page.goto('/?stableRender=true');
    await page.waitForSelector('canvas', { timeout: 10000 });
    await page.waitForTimeout(1000);

    // Take screenshot of the canvas
    const canvas = page.locator('canvas').first();
    await expect(canvas).toHaveScreenshot('star-node.png', {
      maxDiffPixels: 250,
      threshold: 0.2
    });
  });

  test('should render planet node with correct visual appearance', async ({ page }) => {
    await createNote(page.request, {
      title: 'Test Planet',
      content: 'Visual test for planet rendering',
      type: 'planet'
    });

    await page.goto('/?stableRender=true');
    await page.waitForSelector('canvas', { timeout: 10000 });
    await page.waitForTimeout(1000);

    const canvas = page.locator('canvas').first();
    await expect(canvas).toHaveScreenshot('planet-node.png', {
      maxDiffPixels: 250,
      threshold: 0.2
    });
  });

  test('should render comet node with correct visual appearance', async ({ page }) => {
    await createNote(page.request, {
      title: 'Test Comet',
      content: 'Visual test for comet rendering',
      type: 'comet'
    });

    await page.goto('/?stableRender=true');
    await page.waitForSelector('canvas', { timeout: 10000 });
    await page.waitForTimeout(1000);

    const canvas = page.locator('canvas').first();
    await expect(canvas).toHaveScreenshot('comet-node.png', {
      maxDiffPixels: 250,
      threshold: 0.2
    });
  });

  test('should render galaxy node with correct visual appearance', async ({ page }) => {
    await createNote(page.request, {
      title: 'Test Galaxy',
      content: 'Visual test for galaxy rendering',
      type: 'galaxy'
    });

    await page.goto('/?stableRender=true');
    await page.waitForSelector('canvas', { timeout: 10000 });
    await page.waitForTimeout(1000);

    const canvas = page.locator('canvas').first();
    await expect(canvas).toHaveScreenshot('galaxy-node.png', {
      maxDiffPixels: 250,
      threshold: 0.2
    });
  });

  test('should render asteroid node with correct visual appearance', async ({ page }) => {
    await createNote(page.request, {
      title: 'Test Asteroid',
      content: 'Visual test for asteroid rendering',
      type: 'asteroid'
    });

    await page.goto('/?stableRender=true');
    await page.waitForSelector('canvas', { timeout: 10000 });
    await page.waitForTimeout(1000);

    const canvas = page.locator('canvas').first();
    await expect(canvas).toHaveScreenshot('asteroid-node.png', {
      maxDiffPixels: 250,
      threshold: 0.2
    });
  });

  test('should render link between two nodes', async ({ page }) => {
    // Create two nodes
    const sourceNote = await createNote(page.request, {
      title: 'Source Node',
      content: 'Source for link test',
      type: 'star'
    });

    const targetNote = await createNote(page.request, {
      title: 'Target Node',
      content: 'Target for link test',
      type: 'planet'
    });

    // Create link between them
    await createLink(page.request, {
      source_note_id: sourceNote.data.id,
      target_note_id: targetNote.data.id,
      link_type: 'related',
      weight: 0.6
    });

    await page.goto('/?stableRender=true');
    await page.waitForSelector('canvas', { timeout: 10000 });
    await page.waitForTimeout(1500);

    const canvas = page.locator('canvas').first();
    await expect(canvas).toHaveScreenshot('link-between-nodes.png', {
      maxDiffPixels: 250,
      threshold: 0.2
    });
  });

  test('should render different link types with distinct styles', async ({ page }) => {
    // Create reference link
    const source1 = await createNote(page.request, {
      title: 'Reference Source',
      type: 'star'
    });
    const target1 = await createNote(page.request, {
      title: 'Reference Target',
      type: 'planet'
    });
    await createLink(page.request, {
      source_note_id: source1.data.id,
      target_note_id: target1.data.id,
      link_type: 'reference',
      weight: 0.8
    });

    // Create dependency link
    const source2 = await createNote(page.request, {
      title: 'Dependency Source',
      type: 'comet'
    });
    const target2 = await createNote(page.request, {
      title: 'Dependency Target',
      type: 'asteroid'
    });
    await createLink(page.request, {
      source_note_id: source2.data.id,
      target_note_id: target2.data.id,
      link_type: 'dependency',
      weight: 0.8
    });

    // Create custom link
    const source3 = await createNote(page.request, {
      title: 'Custom Source',
      type: 'galaxy'
    });
    const target3 = await createNote(page.request, {
      title: 'Custom Target',
      type: 'star'
    });
    await createLink(page.request, {
      source_note_id: source3.data.id,
      target_note_id: target3.data.id,
      link_type: 'custom',
      weight: 0.5
    });

    await page.goto('/?stableRender=true');
    await page.waitForSelector('canvas', { timeout: 10000 });
    await page.waitForTimeout(1500);

    const canvas = page.locator('canvas').first();
    await expect(canvas).toHaveScreenshot('multiple-link-types.png', {
      maxDiffPixels: 300,
      threshold: 0.3
    });
  });
});
