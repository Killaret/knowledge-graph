/**
 * Visual regression tests using mock data
 * Does not require backend - uses hardcoded graph data
 * @visual @regression @mock
 */

import { test, expect } from '@playwright/test';

test.describe.configure({ mode: 'serial' });

test.describe('GraphCanvas Visual - Mock Node Types @visual @mock', () => {
  const nodeTypes = ['star', 'planet', 'comet', 'galaxy', 'asteroid'];

  for (const type of nodeTypes) {
    test(`should render ${type} node correctly`, async ({ page }) => {
      // Navigate to test page with mock data
      await page.goto(`/test/isolated-node?type=${type}`);
      
      // Wait for canvas to be ready
      await page.waitForSelector('canvas', { timeout: 10000 });
      
      // Wait for simulation to stabilize
      await page.waitForTimeout(1000);
      
      // Take screenshot of the canvas
      const canvas = page.locator('canvas').first();
      await expect(canvas).toHaveScreenshot(`${type}-node-mock.png`, {
        maxDiffPixels: 1000,
        threshold: 0.6,
        animations: 'disabled'
      });
    });
  }
});

test.describe('GraphCanvas Visual - Mock Link Types @visual @mock', () => {
  const linkTypes = ['reference', 'dependency', 'related', 'custom'];

  for (const linkType of linkTypes) {
    test(`should render ${linkType} link correctly`, async ({ page }) => {
      // Navigate to link test page with mock data
      await page.goto(`/test/link-pair?linkType=${linkType}`);
      
      // Wait for canvas to be ready
      await page.waitForSelector('canvas', { timeout: 10000 });
      
      // Wait for simulation to stabilize
      await page.waitForTimeout(1000);
      
      // Take screenshot
      const canvas = page.locator('canvas').first();
      await expect(canvas).toHaveScreenshot(`${linkType}-link-mock.png`, {
        maxDiffPixels: 1200,
        threshold: 0.6,
        animations: 'disabled'
      });
    });
  }
});
