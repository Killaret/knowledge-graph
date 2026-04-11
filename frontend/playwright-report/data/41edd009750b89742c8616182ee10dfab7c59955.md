# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: graph-3d.spec.ts >> 3D Graph Visualization >> should render graph page with canvas visible
- Location: tests\graph-3d.spec.ts:17:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('canvas').first().or(locator('.graph-2d, .graph-3d, .graph-container canvas').first())
Expected: visible
Error: strict mode violation: locator('canvas').first().or(locator('.graph-2d, .graph-3d, .graph-container canvas').first()) resolved to 2 elements:
    1) <div class="graph-wrapper graph-3d s-rnaYcBrWX2vf">…</div> aka locator('div').nth(3)
    2) <canvas width="1264" height="639" data-engine="three.js r160"></canvas> aka locator('canvas')

Call log:
  - Expect "toBeVisible" with timeout 10000ms
  - waiting for locator('canvas').first().or(locator('.graph-2d, .graph-3d, .graph-container canvas').first())

```

# Page snapshot

```yaml
- generic [ref=e3]:
  - banner [ref=e4]:
    - button "Go back" [ref=e5] [cursor=pointer]: « Back
    - heading "Knowledge Constellation" [level=1] [ref=e6]
    - generic [ref=e7]: Drag to rotate/pan • Scroll to zoom • Click node to open
  - generic [ref=e9]:
    - generic: 3D Mode
```

# Test source

```ts
  1   | import { test, expect } from '@playwright/test';
  2   | 
  3   | /**
  4   |  * Tests for 3D Graph Visualization
  5   |  * These tests verify that the 3D graph canvas renders correctly
  6   |  * and that the SmartGraph component properly selects 2D or 3D mode
  7   |  */
  8   | 
  9   | test.describe('3D Graph Visualization', () => {
  10  |   
  11  |   test.beforeEach(async ({ page }) => {
  12  |     // Navigate to home page first to create a note
  13  |     await page.goto('/');
  14  |     await page.waitForLoadState('networkidle');
  15  |   });
  16  | 
  17  |   test('should render graph page with canvas visible', async ({ page }) => {
  18  |     // Create a new note first
  19  |     await page.click('a:has-text("New Note")');
  20  |     await page.waitForURL(/\/notes\/new/);
  21  |     
  22  |     // Fill in note details
  23  |     await page.fill('input[name="title"]', 'Graph Test Note');
  24  |     await page.fill('textarea[name="content"]', 'This is a test note for graph visualization');
  25  |     await page.click('button[type="submit"]');
  26  |     
  27  |     // Wait for note creation and redirect to note page
  28  |     await page.waitForURL(/\/notes\/[a-f0-9-]+/);
  29  |     
  30  |     // Get the note ID from URL
  31  |     const noteUrl = page.url();
  32  |     const noteId = noteUrl.split('/').pop();
  33  |     
  34  |     // Navigate to graph page
  35  |     await page.goto(`/graph/${noteId}`);
  36  |     await page.waitForLoadState('networkidle');
  37  |     
  38  |     // Verify page title
  39  |     await expect(page.locator('h1')).toHaveText('Knowledge Constellation');
  40  |     
  41  |     // Verify canvas or 2D fallback is visible
  42  |     const canvas = page.locator('canvas').first();
  43  |     const graph2D = page.locator('.graph-2d, .graph-3d, .graph-container canvas').first();
  44  |     
> 45  |     await expect(canvas.or(graph2D)).toBeVisible({ timeout: 10000 });
      |                                      ^ Error: expect(locator).toBeVisible() failed
  46  |   });
  47  | 
  48  |   test('should show graph container with correct styling', async ({ page }) => {
  49  |     // Create a new note
  50  |     await page.click('a:has-text("New Note")');
  51  |     await page.waitForURL(/\/notes\/new/);
  52  |     
  53  |     await page.fill('input[name="title"]', 'Styling Test Note');
  54  |     await page.fill('textarea[name="content"]', 'Testing graph styling');
  55  |     await page.click('button[type="submit"]');
  56  |     
  57  |     await page.waitForURL(/\/notes\/[a-f0-9-]+/);
  58  |     const noteUrl = page.url();
  59  |     const noteId = noteUrl.split('/').pop();
  60  |     
  61  |     // Navigate to graph page
  62  |     await page.goto(`/graph/${noteId}`);
  63  |     await page.waitForLoadState('networkidle');
  64  |     
  65  |     // Verify graph container has correct background
  66  |     const graphPage = page.locator('.graph-page');
  67  |     await expect(graphPage).toBeVisible();
  68  |     
  69  |     // Verify header elements
  70  |     await expect(page.locator('.graph-header')).toBeVisible();
  71  |     await expect(page.locator('.graph-header h1')).toHaveText('Knowledge Constellation');
  72  |     await expect(page.locator('.hint')).toBeVisible();
  73  |   });
  74  | 
  75  |   test('should handle back button navigation from graph page', async ({ page }) => {
  76  |     // Create a new note
  77  |     await page.click('a:has-text("New Note")');
  78  |     await page.waitForURL(/\/notes\/new/);
  79  |     
  80  |     await page.fill('input[name="title"]', 'Navigation Test Note');
  81  |     await page.fill('textarea[name="content"]', 'Testing navigation from graph');
  82  |     await page.click('button[type="submit"]');
  83  |     
  84  |     await page.waitForURL(/\/notes\/[a-f0-9-]+/);
  85  |     const noteUrl = page.url();
  86  |     const noteId = noteUrl.split('/').pop();
  87  |     
  88  |     // Navigate to graph page
  89  |     await page.goto(`/graph/${noteId}`);
  90  |     await page.waitForLoadState('networkidle');
  91  |     
  92  |     // Click back button
  93  |     await page.click('.back-button');
  94  |     
  95  |     // Should navigate back
  96  |     await page.waitForURL(/\/$/);
  97  |     await expect(page).toHaveURL('/');
  98  |   });
  99  | 
  100 |   test('should display performance mode indicator', async ({ page }) => {
  101 |     // Create a new note
  102 |     await page.click('a:has-text("New Note")');
  103 |     await page.waitForURL(/\/notes\/new/);
  104 |     
  105 |     await page.fill('input[name="title"]', 'Performance Test Note');
  106 |     await page.fill('textarea[name="content"]', 'Testing performance mode indicator');
  107 |     await page.click('button[type="submit"]');
  108 |     
  109 |     await page.waitForURL(/\/notes\/[a-f0-9-]+/);
  110 |     const noteUrl = page.url();
  111 |     const noteId = noteUrl.split('/').pop();
  112 |     
  113 |     // Navigate to graph page
  114 |     await page.goto(`/graph/${noteId}`);
  115 |     await page.waitForLoadState('networkidle');
  116 |     
  117 |     // Wait for graph to load (either 2D or 3D)
  118 |     await page.waitForTimeout(2000);
  119 |     
  120 |     // Check for performance hint (either "3D Mode" or "2D Mode")
  121 |     const performanceHint = page.locator('.performance-hint');
  122 |     if (await performanceHint.isVisible().catch(() => false)) {
  123 |       const hintText = await performanceHint.textContent();
  124 |       expect(hintText).toMatch(/(3D Mode|2D Mode)/);
  125 |     }
  126 |   });
  127 | 
  128 |   test('should handle graph page with no nodes gracefully', async ({ page }) => {
  129 |     // Create a note with no links
  130 |     await page.click('a:has-text("New Note")');
  131 |     await page.waitForURL(/\/notes\/new/);
  132 |     
  133 |     await page.fill('input[name="title"]', 'Isolated Note');
  134 |     await page.fill('textarea[name="content"]', 'This note has no connections');
  135 |     await page.click('button[type="submit"]');
  136 |     
  137 |     await page.waitForURL(/\/notes\/[a-f0-9-]+/);
  138 |     const noteUrl = page.url();
  139 |     const noteId = noteUrl.split('/').pop();
  140 |     
  141 |     // Navigate to graph page
  142 |     await page.goto(`/graph/${noteId}`);
  143 |     await page.waitForLoadState('networkidle');
  144 |     
  145 |     // Page should still render without errors
```