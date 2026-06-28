import { test, expect } from '@playwright/test';

/**
 * Comparison Test: Dev vs Personal Stack Graph Rendering
 * This test compares the graph visualization between dev and personal stacks
 */

test.describe('Stack Comparison: Dev vs Personal', () => {
  
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => {
      (window as any).__SKIP_AUTH__ = true;
    });
  });

  test('Dev stack graph view', async ({ page }) => {
    await page.goto('http://localhost:5173/graph');
    await page.waitForLoadState('networkidle');
    
    // Wait for graph canvas to load
    await expect(page.locator('canvas')).toBeVisible({ timeout: 15000 });
    await page.waitForTimeout(2000); // Allow graph simulation to stabilize
    
    await page.screenshot({ 
      path: 'screenshots/dev-stack-graph.png',
      fullPage: true 
    });
  });

  test('Personal stack graph view', async ({ page }) => {
    await page.goto('http://localhost:3001/graph');
    await page.waitForLoadState('networkidle');
    
    // Wait for graph canvas to load
    await expect(page.locator('canvas')).toBeVisible({ timeout: 15000 });
    await page.waitForTimeout(2000); // Allow graph simulation to stabilize
    
    await page.screenshot({ 
      path: 'screenshots/personal-stack-graph.png',
      fullPage: true 
    });
  });
});