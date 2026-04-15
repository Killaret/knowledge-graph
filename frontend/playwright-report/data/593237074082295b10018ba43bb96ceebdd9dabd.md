# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: graph-3d.spec.ts >> Graph Visualization >> should render graph page with visualization
- Location: tests\graph-3d.spec.ts:17:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: true
Received: false
```

# Page snapshot

```yaml
- generic [ref=e3]:
  - banner [ref=e4]:
    - button "Go back" [ref=e5] [cursor=pointer]: « Back
    - heading "Knowledge Constellation" [level=1] [ref=e6]
    - generic [ref=e7]: Drag to rotate/pan • Scroll to zoom • Click node to open
  - paragraph [ref=e11]: Loading visualization...
```

# Test source

```ts
  1   | import { test, expect } from '@playwright/test';
  2   | 
  3   | /**
  4   |  * Tests for Graph Visualization (2D primary, 3D optional)
  5   |  * These tests verify that the graph renders correctly in both modes
  6   |  * Note: WebGL may be limited in headless environments, so we test for either 2D or 3D
  7   |  */
  8   | 
  9   | test.describe('Graph Visualization', () => {
  10  |   
  11  |   test.beforeEach(async ({ page }) => {
  12  |     // Navigate to home page first
  13  |     await page.goto('http://localhost:5173/');
  14  |     await page.waitForLoadState('networkidle');
  15  |   });
  16  | 
  17  |   test('should render graph page with visualization', async ({ page, request }) => {
  18  |     // Create a note via API
  19  |     const note = await request.post('http://localhost:8080/notes', {
  20  |       data: { title: 'Graph Test Note', content: 'Test note for graph' }
  21  |     });
  22  |     const noteId = (await note.json()).id;
  23  |     
  24  |     // Navigate to graph page
  25  |     await page.goto(`http://localhost:5173/graph/${noteId}`);
  26  |     await page.waitForLoadState('networkidle');
  27  |     await page.waitForTimeout(3000);
  28  |     
  29  |     // Verify either canvas (2D/3D) or error message is shown
  30  |     // WebGL may not work in headless mode, so we check for any graph container
  31  |     const canvas = page.locator('canvas, .graph-canvas').first();
  32  |     const graphContainer = page.locator('.graph-3d-container, .graph-2d, .graph-wrapper, .error-overlay').first();
  33  |     
  34  |     const hasCanvas = await canvas.isVisible().catch(() => false);
  35  |     const hasContainer = await graphContainer.isVisible().catch(() => false);
  36  |     
  37  |     // At least one visualization element should be present
> 38  |     expect(hasCanvas || hasContainer).toBe(true);
      |                                       ^ Error: expect(received).toBe(expected) // Object.is equality
  39  |   });
  40  | 
  41  |   test('should show graph container with correct styling', async ({ page, request }) => {
  42  |     // Create a note via API
  43  |     const note = await request.post('http://localhost:8080/notes', {
  44  |       data: { title: 'Styling Test Note', content: 'Testing graph styling' }
  45  |     });
  46  |     const noteId = (await note.json()).id;
  47  |     
  48  |     // Navigate to graph page
  49  |     await page.goto(`http://localhost:5173/graph/${noteId}`);
  50  |     await page.waitForLoadState('networkidle');
  51  |     await page.waitForTimeout(2000);
  52  |     
  53  |     // Verify graph container is visible (use first())
  54  |     await expect(page.locator('.graph-3d-container, .graph-2d, .graph-wrapper, .error-overlay').first()).toBeVisible();
  55  |     
  56  |     // Canvas may not be available in headless mode due to WebGL limitations
  57  |     // Just verify the container loaded correctly
  58  |     const container = page.locator('.graph-3d-container, .graph-2d, .graph-wrapper').first();
  59  |     const errorOverlay = page.locator('.error-overlay').first();
  60  |     
  61  |     const hasContainer = await container.isVisible().catch(() => false);
  62  |     const hasError = await errorOverlay.isVisible().catch(() => false);
  63  |     
  64  |     // Either graph or error message should be shown
  65  |     expect(hasContainer || hasError).toBe(true);
  66  |   });
  67  | 
  68  |   test('should handle back button navigation from graph page', async ({ page, request }) => {
  69  |     // Create a note via API
  70  |     const note = await request.post('http://localhost:8080/notes', {
  71  |       data: { title: 'Navigation Test Note', content: 'Testing navigation from graph' }
  72  |     });
  73  |     const noteId = (await note.json()).id;
  74  |     
  75  |     // Navigate to graph page
  76  |     await page.goto(`http://localhost:5173/graph/${noteId}`);
  77  |     await page.waitForLoadState('networkidle');
  78  |     await page.waitForTimeout(1000);
  79  |     
  80  |     // Navigate back using browser back
  81  |     await page.goBack();
  82  |     await page.waitForTimeout(1000);
  83  |     
  84  |     // Should be back on home or note page
  85  |     const currentUrl = page.url();
  86  |     expect(currentUrl).toMatch(/http:\/\/localhost:5173(\/|\/notes\/.+)/);
  87  |   });
  88  | 
  89  |   test('should display performance mode or fallback indicator', async ({ page, request }) => {
  90  |     // Create a note via API
  91  |     const note = await request.post('http://localhost:8080/notes', {
  92  |       data: { title: 'Performance Test Note', content: 'Testing performance mode indicator' }
  93  |     });
  94  |     const noteId = (await note.json()).id;
  95  |     
  96  |     // Navigate to graph page
  97  |     await page.goto(`http://localhost:5173/graph/${noteId}`);
  98  |     await page.waitForLoadState('networkidle');
  99  |     
  100 |     // Wait for graph to load (either 2D or 3D)
  101 |     await page.waitForTimeout(3000);
  102 |     
  103 |     // Check for performance hint (optional - may not be visible in headless)
  104 |     const performanceHint = page.locator('.performance-hint');
  105 |     const hasHint = await performanceHint.isVisible().catch(() => false);
  106 |     
  107 |     if (hasHint) {
  108 |       const hintText = await performanceHint.textContent();
  109 |       expect(hintText).toMatch(/(3D Mode|2D Mode|Performance)/i);
  110 |     }
  111 |     
  112 |     // Verify graph container loaded (success or error state)
  113 |     const container = page.locator('.graph-3d-container, .graph-2d, .graph-wrapper, .error-overlay').first();
  114 |     await expect(container).toBeVisible();
  115 |   });
  116 | 
  117 | });
  118 | 
```