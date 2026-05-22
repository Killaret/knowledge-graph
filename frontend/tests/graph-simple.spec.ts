import { test, expect } from '@playwright/test';

test('should render star node', async ({ page }) => {
  await page.goto('http://localhost:5173/test/isolated-node?type=star&stableRender=true');
  await page.waitForSelector('canvas', { timeout: 30000 });
  await page.waitForTimeout(2000);
  
  const canvas = page.locator('canvas').first();
  await expect(canvas).toHaveScreenshot('star-node.png', {
    maxDiffPixels: 500,
    threshold: 0.4
  });
});
