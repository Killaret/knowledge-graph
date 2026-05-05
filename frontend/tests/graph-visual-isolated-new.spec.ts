/**
 * Isolated visual regression tests for graph rendering
 * Uses query params to test different node/link types
 * @visual @regression @isolated
 */

import { test, expect } from '@playwright/test';

test.describe.configure({ mode: 'serial' });

test.describe('GraphCanvas Visual - Isolated Node Types @visual @isolated', () => {
  const nodeTypes = ['star', 'planet', 'comet', 'galaxy', 'asteroid'];

  for (const type of nodeTypes) {
    test(`should render ${type} node correctly`, async ({ page }) => {
      // Navigate to test page with query param
      await page.goto(`/test/isolated-node?type=${type}`);
      
      // Wait for canvas to be ready
      const canvas = page.locator('canvas');
      await canvas.waitFor({ timeout: 30000 });
      
      // Wait for D3 simulation to stabilize
      await page.waitForTimeout(3000);
      
      // Take screenshot
      await expect(canvas).toHaveScreenshot(`${type}-node-isolated.png`, {
        maxDiffPixels: 500,
        threshold: 0.4,
        animations: 'disabled'
      });
    });
  }
});

test.describe('GraphCanvas Visual - Link Types @visual @links', () => {
  const linkTypes = ['reference', 'dependency', 'related', 'custom'];

  for (const linkType of linkTypes) {
    test(`should render ${linkType} link correctly`, async ({ page }) => {
      // Navigate to link test page
      await page.goto(`/test/link-pair?linkType=${linkType}`);
      
      // Wait for canvas
      const canvas = page.locator('canvas');
      await canvas.waitFor({ timeout: 30000 });
      
      // Wait for simulation
      await page.waitForTimeout(3000);
      
      // Take screenshot
      await expect(canvas).toHaveScreenshot(`${linkType}-link-isolated.png`, {
        maxDiffPixels: 600,
        threshold: 0.4,
        animations: 'disabled'
      });
    });
  }
});
