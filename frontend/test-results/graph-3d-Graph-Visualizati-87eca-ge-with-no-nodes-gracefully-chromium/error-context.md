# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: graph-3d.spec.ts >> Graph Visualization >> should handle graph page with no nodes gracefully
- Location: tests\graph-3d.spec.ts:117:3

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
  - generic [ref=e9]:
    - paragraph [ref=e13]: Loading 3D constellation...
    - generic: 3D Mode
```

# Test source

```ts
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
  117 |   test('should handle graph page with no nodes gracefully', async ({ page, request }) => {
  118 |     // Create a note with no links via API
  119 |     const note = await request.post('http://localhost:8080/notes', {
  120 |       data: { title: 'Isolated Note', content: 'This note has no connections' }
  121 |     });
  122 |     const noteId = (await note.json()).id;
  123 |     
  124 |     // Navigate to graph page
  125 |     await page.goto(`http://localhost:5173/graph/${noteId}`);
  126 |     await page.waitForLoadState('networkidle');
  127 |     await page.waitForTimeout(2000);
  128 |     
  129 |     // Page should render without errors - check for any graph container or error message
  130 |     const graphElements = page.locator('.graph-3d-container, .graph-2d, .graph-wrapper, .error-overlay, .no-data-message').first();
  131 |     await expect(graphElements).toBeVisible();
  132 |     
  133 |     // Either canvas or error/empty state message should be shown
  134 |     const canvas = page.locator('canvas, .graph-canvas').first();
  135 |     const message = page.locator('.error-overlay, .no-data-message, .empty-state').first();
  136 |     
  137 |     const hasCanvas = await canvas.isVisible().catch(() => false);
  138 |     const hasMessage = await message.isVisible().catch(() => false);
  139 |     
  140 |     // At least one visualization element should be present
> 141 |     expect(hasCanvas || hasMessage).toBe(true);
      |                                     ^ Error: expect(received).toBe(expected) // Object.is equality
  142 |   });
  143 | });
  144 | 
```