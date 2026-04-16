# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: camera-position.spec.ts >> 3D Graph - Camera Position and Navigation >> should position camera to show start node in center for local graph
- Location: tests\camera-position.spec.ts:16:3

# Error details

```
Error: expect(received).toBeTruthy()

Received: undefined
```

# Page snapshot

```yaml
- generic [ref=e3]:
  - generic [ref=e5] [cursor=pointer]:
    - checkbox "Показать все заметки (выключено)" [ref=e6]
    - generic [ref=e7]: Показать все заметки (выключено)
  - generic [ref=e8]:
    - generic [ref=e9]:
      - strong [ref=e10]: "1"
      - text: nodes
    - generic [ref=e11]:
      - strong [ref=e12]: "0"
      - text: links
    - generic [ref=e13]: (Local view)
  - generic [ref=e17] [cursor=pointer]: Center Node
```

# Test source

```ts
  1   | import { test, expect } from '@playwright/test';
  2   | 
  3   | /**
  4   |  * Tests for Camera Position and Navigation in 3D Graph
  5   |  * Verifies camera positioning, auto-zoom, and route transitions
  6   |  */
  7   | 
  8   | test.describe('3D Graph - Camera Position and Navigation', () => {
  9   |   
  10  |   test.beforeEach(async ({ page }) => {
  11  |     await page.goto('http://localhost:5173/');
  12  |     await page.waitForLoadState('networkidle');
  13  |     await page.waitForTimeout(1000);
  14  |   });
  15  | 
  16  |   test('should position camera to show start node in center for local graph', async ({ page, request }) => {
  17  |     // Create a note with connections
  18  |     const centerNote = await request.post('http://localhost:8080/notes', {
  19  |       data: { title: 'Center Node', content: 'Main node for camera test', type: 'star' }
  20  |     });
  21  |     const centerId = (await centerNote.json()).id;
  22  |     
  23  |     const linkedNote = await request.post('http://localhost:8080/notes', {
  24  |       data: { title: 'Linked Node', content: 'Secondary node', type: 'planet' }
  25  |     });
  26  |     const linkedId = (await linkedNote.json()).id;
  27  |     
  28  |     await request.post('http://localhost:8080/links', {
  29  |       data: { sourceNoteId: centerId, targetNoteId: linkedId, weight: 0.8 }
  30  |     });
  31  |     
  32  |     // Navigate to 3D graph for center node
  33  |     await page.goto(`http://localhost:5173/graph/3d/${centerId}`);
  34  |     await page.waitForLoadState('networkidle');
  35  |     await page.waitForTimeout(4000); // Wait for simulation and auto-zoom
  36  |     
  37  |     const container = page.locator('.graph-3d-container').first();
  38  |     await expect(container).toBeVisible({ timeout: 10000 });
  39  |     
  40  |     // Check that camera logs show centering on nodes
  41  |     const logs: string[] = await page.evaluate(() => {
  42  |       return (window as any).consoleLogs || [];
  43  |     });
  44  |     
  45  |     // Verify auto-zoom was called
  46  |     const autoZoomLog = logs.find(log => log.includes('autoZoomToFit'));
> 47  |     expect(autoZoomLog).toBeTruthy();
      |                         ^ Error: expect(received).toBeTruthy()
  48  |   });
  49  | 
  50  |   test('should position camera appropriately for full 3D graph', async ({ page, request }) => {
  51  |     // Create multiple notes for full graph
  52  |     const notes = [];
  53  |     for (let i = 0; i < 5; i++) {
  54  |       const note = await request.post('http://localhost:8080/notes', {
  55  |         data: { title: `Full Graph Node ${i}`, content: `Node ${i}`, type: i === 0 ? 'star' : 'planet' }
  56  |       });
  57  |       notes.push((await note.json()).id);
  58  |     }
  59  |     
  60  |     // Create some links
  61  |     for (let i = 0; i < notes.length - 1; i++) {
  62  |       await request.post('http://localhost:8080/links', {
  63  |         data: { sourceNoteId: notes[i], targetNoteId: notes[i + 1], weight: 0.7 }
  64  |       });
  65  |     }
  66  |     
  67  |     // Navigate to full 3D graph
  68  |     await page.goto('http://localhost:5173/graph/3d');
  69  |     await page.waitForLoadState('networkidle');
  70  |     await page.waitForTimeout(5000);
  71  |     
  72  |     const container = page.locator('.graph-3d-container, .lazy-loading, .lazy-error').first();
  73  |     await expect(container).toBeVisible({ timeout: 10000 });
  74  |     
  75  |     // Stats should show multiple nodes
  76  |     const statsBar = page.locator('.stats-bar').first();
  77  |     if (await statsBar.isVisible().catch(() => false)) {
  78  |       const statsText = await statsBar.textContent();
  79  |       const nodeMatch = statsText?.match(/(\d+)\s*nodes/);
  80  |       if (nodeMatch) {
  81  |         const nodeCount = parseInt(nodeMatch[1], 10);
  82  |         expect(nodeCount).toBeGreaterThanOrEqual(5);
  83  |       }
  84  |     }
  85  |   });
  86  | 
  87  |   test('should adjust camera when toggling full graph mode', async ({ page, request }) => {
  88  |     const note = await request.post('http://localhost:8080/notes', {
  89  |       data: { title: 'Toggle Camera Test', content: 'Testing camera on toggle', type: 'star' }
  90  |     });
  91  |     const noteId = (await note.json()).id;
  92  |     
  93  |     await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
  94  |     await page.waitForLoadState('networkidle');
  95  |     await page.waitForTimeout(3000);
  96  |     
  97  |     const container = page.locator('.graph-3d-container').first();
  98  |     await expect(container).toBeVisible();
  99  |     
  100 |     // Find and click the toggle
  101 |     const toggle = page.locator('.toggle input[type="checkbox"]').first();
  102 |     if (await toggle.isVisible().catch(() => false)) {
  103 |       // Get initial state
  104 |       const isChecked = await toggle.isChecked();
  105 |       
  106 |       // Toggle to other mode
  107 |       await toggle.click();
  108 |       await page.waitForTimeout(3000); // Wait for camera adjustment
  109 |       
  110 |       // Graph should still be visible after camera adjustment
  111 |       await expect(container).toBeVisible();
  112 |       
  113 |       // Stats should update
  114 |       const statsBar = page.locator('.stats-bar').first();
  115 |       if (await statsBar.isVisible().catch(() => false)) {
  116 |         const statsText = await statsBar.textContent();
  117 |         // Should show appropriate mode
  118 |         const hasModeIndicator = statsText?.toLowerCase().includes('full') || 
  119 |                                  statsText?.toLowerCase().includes('local');
  120 |         expect(hasModeIndicator).toBe(true);
  121 |       }
  122 |       
  123 |       // Toggle back
  124 |       await toggle.click();
  125 |       await page.waitForTimeout(3000);
  126 |       
  127 |       // Should still be visible
  128 |       await expect(container).toBeVisible();
  129 |     }
  130 |   });
  131 | 
  132 |   test('should maintain camera position on route navigation', async ({ page, request }) => {
  133 |     // Create test note
  134 |     const note = await request.post('http://localhost:8080/notes', {
  135 |       data: { title: 'Navigation Test', content: 'Testing route navigation', type: 'star' }
  136 |     });
  137 |     const noteId = (await note.json()).id;
  138 |     
  139 |     // Go to 3D graph
  140 |     await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
  141 |     await page.waitForLoadState('networkidle');
  142 |     await page.waitForTimeout(4000);
  143 |     
  144 |     const container3D = page.locator('.graph-3d-container').first();
  145 |     await expect(container3D).toBeVisible();
  146 |     
  147 |     // Navigate to home page
```