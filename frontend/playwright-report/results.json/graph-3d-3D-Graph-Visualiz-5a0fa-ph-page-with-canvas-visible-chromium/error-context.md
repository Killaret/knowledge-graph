# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: graph-3d.spec.ts >> 3D Graph Visualization >> should render graph page with canvas visible
- Location: tests\graph-3d.spec.ts:26:3

# Error details

```
Error: expect(locator).toHaveText(expected) failed

Locator:  locator('h1')
Expected: "Knowledge Constellation"
Received: "500"
Timeout:  5000ms

Call log:
  - Expect "toHaveText" with timeout 5000ms
  - waiting for locator('h1')
    9 × locator resolved to <h1>500</h1>
      - unexpected value "500"

```

# Page snapshot

```yaml
- generic [ref=e2]:
  - heading "500" [level=1] [ref=e3]
  - paragraph [ref=e4]: Internal Error
  - generic:
    - generic: Ctrl+N
    - text: — новая заметка
    - generic: Ctrl+F
    - text: — поиск
    - generic: Esc
    - text: — закрыть
```

# Test source

```ts
  1   | import { test, expect } from '@playwright/test';
  2   | 
  3   | /**
  4   |  * Tests for 3D Graph Visualization
  5   |  * These tests verify that the graph page renders correctly
  6   |  * with either 2D or 3D canvas
  7   |  */
  8   | 
  9   | // Store created note IDs for cleanup
  10  | const createdNoteIds: string[] = [];
  11  | 
  12  | test.afterEach(async ({ request }) => {
  13  |   // Cleanup all created notes
  14  |   for (const noteId of createdNoteIds) {
  15  |     try {
  16  |       await request.delete(`http://localhost:8080/notes/${noteId}`);
  17  |     } catch (e) {
  18  |       // Ignore cleanup errors
  19  |     }
  20  |   }
  21  |   createdNoteIds.length = 0;
  22  | });
  23  | 
  24  | test.describe('3D Graph Visualization', () => {
  25  |   
  26  |   test('should render graph page with canvas visible', async ({ page, request }) => {
  27  |     // Create a note via API
  28  |     const timestamp = Date.now();
  29  |     const note = await request.post('http://localhost:8080/notes', {
  30  |       data: { title: 'Graph Test ' + timestamp, content: 'Test content' }
  31  |     });
  32  |     const noteData = await note.json();
  33  |     const noteId = noteData.id;
  34  |     createdNoteIds.push(noteId);
  35  |     
  36  |     // Navigate to graph page
  37  |     await page.goto(`http://localhost:5173/graph/${noteId}`);
  38  |     await page.waitForLoadState('networkidle');
  39  |     
  40  |     // Verify page title
> 41  |     await expect(page.locator('h1')).toHaveText('Knowledge Constellation');
      |                                      ^ Error: expect(locator).toHaveText(expected) failed
  42  |     
  43  |     // Verify canvas or graph container is visible
  44  |     const canvas = page.locator('canvas');
  45  |     const graphContainer = page.locator('.graph-container');
  46  |     await expect(canvas.or(graphContainer)).toBeVisible({ timeout: 10000 });
  47  |     
  48  |     // Verify WebGL is available (if canvas exists)
  49  |     const hasWebGL = await page.evaluate(() => {
  50  |       const canvas = document.querySelector('canvas');
  51  |       if (!canvas) return false;
  52  |       const gl = canvas.getContext('webgl') || canvas.getContext('experimental-webgl');
  53  |       return !!gl;
  54  |     });
  55  |     expect(hasWebGL).toBe(true);
  56  |   });
  57  | 
  58  |   test('should show graph container with correct styling', async ({ page, request }) => {
  59  |     // Create a note via API
  60  |     const timestamp = Date.now();
  61  |     const note = await request.post('http://localhost:8080/notes', {
  62  |       data: { title: 'Styling Test ' + timestamp, content: 'Testing styling' }
  63  |     });
  64  |     const noteData = await note.json();
  65  |     const noteId = noteData.id;
  66  |     createdNoteIds.push(noteId);
  67  |     
  68  |     // Navigate to graph page
  69  |     await page.goto(`http://localhost:5173/graph/${noteId}`);
  70  |     await page.waitForLoadState('networkidle');
  71  |     
  72  |     // Verify graph container has correct structure
  73  |     await expect(page.locator('.graph-page')).toBeVisible();
  74  |     await expect(page.locator('.graph-header')).toBeVisible();
  75  |     await expect(page.locator('.graph-header h1')).toHaveText('Knowledge Constellation');
  76  |     await expect(page.locator('.hint')).toBeVisible();
  77  |     await expect(page.locator('.graph-container')).toBeVisible();
  78  |     
  79  |     // Verify canvas has correct data-testid
  80  |     const canvas = page.locator('[data-testid="graph-canvas"]');
  81  |     await expect(canvas).toBeVisible();
  82  |   });
  83  | 
  84  |   test('should handle back button navigation from graph page', async ({ page, request }) => {
  85  |     // Create a note via API
  86  |     const timestamp = Date.now();
  87  |     const note = await request.post('http://localhost:8080/notes', {
  88  |       data: { title: 'Navigation Test ' + timestamp, content: 'Testing navigation' }
  89  |     });
  90  |     const noteData = await note.json();
  91  |     const noteId = noteData.id;
  92  |     createdNoteIds.push(noteId);
  93  |     
  94  |     // Navigate to graph page
  95  |     await page.goto(`http://localhost:5173/graph/${noteId}`);
  96  |     await page.waitForLoadState('networkidle');
  97  |     
  98  |     // Click back button
  99  |     await page.click('.back-button');
  100 |     
  101 |     // Should navigate back to home
  102 |     await page.waitForURL(/\/$/);
  103 |     
  104 |     // Verify we're on home page
  105 |     await expect(page.locator('h1')).toHaveText('My Notes');
  106 |   });
  107 | 
  108 |   test('should display performance mode indicator', async ({ page, request }) => {
  109 |     // Create a note via API
  110 |     const timestamp = Date.now();
  111 |     const note = await request.post('http://localhost:8080/notes', {
  112 |       data: { title: 'Performance Test ' + timestamp, content: 'Testing performance indicator' }
  113 |     });
  114 |     const noteData = await note.json();
  115 |     const noteId = noteData.id;
  116 |     createdNoteIds.push(noteId);
  117 |     
  118 |     // Navigate to graph page
  119 |     await page.goto(`http://localhost:5173/graph/${noteId}`);
  120 |     await page.waitForLoadState('networkidle');
  121 |     
  122 |     // Wait for performance hint to appear (with better selector)
  123 |     const performanceHint = page.locator('.performance-hint');
  124 |     await expect(performanceHint).toBeVisible({ timeout: 5000 });
  125 |     
  126 |     // Check for performance hint text
  127 |     const hintText = await performanceHint.textContent();
  128 |     expect(hintText).toMatch(/(3D Mode|2D Mode)/);
  129 |   });
  130 | 
  131 |   test('should handle graph page with no connections gracefully', async ({ page, request }) => {
  132 |     // Create an isolated note via API
  133 |     const timestamp = Date.now();
  134 |     const note = await request.post('http://localhost:8080/notes', {
  135 |       data: { title: 'Isolated Note ' + timestamp, content: 'No connections' }
  136 |     });
  137 |     const noteData = await note.json();
  138 |     const noteId = noteData.id;
  139 |     createdNoteIds.push(noteId);
  140 |     
  141 |     // Navigate to graph page
```