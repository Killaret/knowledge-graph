# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: graph-3d.spec.ts >> 3D Graph Visualization >> should render graph page with canvas visible
- Location: tests\graph-3d.spec.ts:26:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('canvas').or(locator('.graph-container'))
Expected: visible
Error: strict mode violation: locator('canvas').or(locator('.graph-container')) resolved to 2 elements:
    1) <div class="graph-container s-PnnjJa1Ln-ae">…</div> aka locator('div').nth(5)
    2) <canvas width="192" height="128" class="map-canvas s-xF5NUJj7tVq3"></canvas> aka locator('canvas')

Call log:
  - Expect "toBeVisible" with timeout 10000ms
  - waiting for locator('canvas').or(locator('.graph-container'))

```

# Page snapshot

```yaml
- generic [ref=e3]:
  - banner [ref=e4]:
    - generic [ref=e5]:
      - generic [ref=e6]: 🌌
      - generic [ref=e7]: Knowledge Graph
    - button "Поиск (Ctrl+F)" [ref=e9] [cursor=pointer]:
      - img [ref=e10]
  - main [ref=e13]:
    - generic [ref=e14]:
      - generic [ref=e15]:
        - button "Go back" [ref=e16] [cursor=pointer]: « Back
        - heading "Knowledge Constellation" [level=1] [ref=e17]
        - generic [ref=e18]: Drag to rotate/pan • Scroll to zoom • Click node to open
      - generic [ref=e20]:
        - generic [ref=e21]: ⭐
        - heading "Это одинокая звезда" [level=3] [ref=e22]
        - paragraph [ref=e23]: Создайте связи, чтобы увидеть созвездие
        - button "Создать связь" [ref=e24] [cursor=pointer]
  - navigation "Main navigation" [ref=e25]:
    - button "Новая звезда" [ref=e26] [cursor=pointer]:
      - img [ref=e27]
    - button "Граф" [ref=e29] [cursor=pointer]:
      - img [ref=e30]
    - button "Все заметки" [ref=e36] [cursor=pointer]:
      - img [ref=e37]
    - button "Импорт" [ref=e38] [cursor=pointer]:
      - img [ref=e39]
    - button "Экспорт" [ref=e42] [cursor=pointer]:
      - img [ref=e43]
    - button "Настройки" [ref=e46] [cursor=pointer]:
      - img [ref=e47]
  - generic [ref=e50]:
    - button "Обзор" [ref=e51] [cursor=pointer]:
      - generic [ref=e52]: Обзор
      - img [ref=e53]
    - generic [ref=e55]:
      - generic [ref=e56]:
        - generic [ref=e57]:
          - generic [ref=e58]: "0"
          - generic [ref=e59]: узлов
        - generic [ref=e61]:
          - generic [ref=e62]: "0"
          - generic [ref=e63]: связей
      - generic [ref=e66]:
        - img [ref=e67]
        - generic [ref=e73]: Мини-карта
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
  41  |     await expect(page.locator('h1')).toHaveText('Knowledge Constellation');
  42  |     
  43  |     // Verify canvas or graph container is visible
  44  |     const canvas = page.locator('canvas');
  45  |     const graphContainer = page.locator('.graph-container');
> 46  |     await expect(canvas.or(graphContainer)).toBeVisible({ timeout: 10000 });
      |                                             ^ Error: expect(locator).toBeVisible() failed
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
  142 |     await page.goto(`http://localhost:5173/graph/${noteId}`);
  143 |     await page.waitForLoadState('networkidle');
  144 |     
  145 |     // Page should render without errors even with no connections
  146 |     await expect(page.locator('h1')).toHaveText('Knowledge Constellation');
```