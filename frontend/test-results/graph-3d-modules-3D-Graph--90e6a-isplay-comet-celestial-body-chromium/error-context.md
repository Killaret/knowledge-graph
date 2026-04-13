# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: graph-3d-modules.spec.ts >> 3D Graph - Modular Architecture >> should display comet celestial body
- Location: tests\graph-3d-modules.spec.ts:91:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('.graph-3d-container, .lazy-error').first()
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('.graph-3d-container, .lazy-error').first()

```

# Page snapshot

```yaml
- generic [ref=e3]:
  - generic [ref=e5] [cursor=pointer]:
    - checkbox "Показать все заметки (включено)" [checked] [ref=e6]
    - generic [ref=e7]: Показать все заметки (включено)
  - generic [ref=e8]: Failed to load graph data
```

# Test source

```ts
  2   | 
  3   | /**
  4   |  * Tests for 3D Graph Modules (Three.js Refactored)
  5   |  * Verifies the modular architecture, celestial bodies rendering, and link visualization
  6   |  */
  7   | 
  8   | test.describe('3D Graph - Modular Architecture', () => {
  9   |   
  10  |   test.beforeEach(async ({ page }) => {
  11  |     await page.goto('http://localhost:5173/');
  12  |     await page.waitForLoadState('networkidle');
  13  |     await page.waitForTimeout(1000);
  14  |   });
  15  | 
  16  |   test('should render 3D graph with scene setup module', async ({ page, request }) => {
  17  |     // Create a note via API
  18  |     const note = await request.post('http://localhost:8080/notes', {
  19  |       data: { 
  20  |         title: 'Scene Test Note', 
  21  |         content: 'Testing scene setup module',
  22  |         type: 'star'
  23  |       }
  24  |     });
  25  |     const noteId = (await note.json()).id;
  26  |     
  27  |     // Navigate to 3D graph page
  28  |     await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
  29  |     await page.waitForLoadState('networkidle');
  30  |     
  31  |     // Wait for lazy loading to complete (max 10 seconds)
  32  |     await page.waitForTimeout(1000);
  33  |     
  34  |     // Check for any of: lazy-loading, graph container, or error
  35  |     const lazyLoading = page.locator('.lazy-loading').first();
  36  |     const graphContainer = page.locator('.graph-3d-container').first();
  37  |     const lazyError = page.locator('.lazy-error').first();
  38  |     
  39  |     // Wait for lazy loading to finish
  40  |     let attempts = 0;
  41  |     while (attempts < 20) {
  42  |       const isLazyLoading = await lazyLoading.isVisible().catch(() => false);
  43  |       if (!isLazyLoading) break;
  44  |       await page.waitForTimeout(500);
  45  |       attempts++;
  46  |     }
  47  |     
  48  |     // After lazy loading, should have either graph or error
  49  |     const hasGraph = await graphContainer.isVisible().catch(() => false);
  50  |     const hasError = await lazyError.isVisible().catch(() => false);
  51  |     
  52  |     expect(hasGraph || hasError).toBe(true);
  53  |   });
  54  | 
  55  |   test('should display star celestial body', async ({ page, request }) => {
  56  |     const note = await request.post('http://localhost:8080/notes', {
  57  |       data: { title: 'Star Node', content: 'Star type test', type: 'star' }
  58  |     });
  59  |     const noteId = (await note.json()).id;
  60  |     
  61  |     const note2 = await request.post('http://localhost:8080/notes', {
  62  |       data: { title: 'Linked', content: 'Link' }
  63  |     });
  64  |     await request.post('http://localhost:8080/links', {
  65  |       data: { sourceNoteId: noteId, targetNoteId: (await note2.json()).id, weight: 0.8 }
  66  |     });
  67  |     
  68  |     await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
  69  |     await page.waitForLoadState('networkidle');
  70  |     await page.waitForTimeout(4000);
  71  |     
  72  |     // After lazy loading, verify graph or error state
  73  |     const container = page.locator('.graph-3d-container, .lazy-error').first();
  74  |     await expect(container).toBeVisible();
  75  |   });
  76  | 
  77  |   test('should display planet celestial body', async ({ page, request }) => {
  78  |     const note = await request.post('http://localhost:8080/notes', {
  79  |       data: { title: 'Planet Node', content: 'Planet type', type: 'planet' }
  80  |     });
  81  |     const noteId = (await note.json()).id;
  82  |     
  83  |     await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
  84  |     await page.waitForLoadState('networkidle');
  85  |     await page.waitForTimeout(4000);
  86  |     
  87  |     const container = page.locator('.graph-3d-container, .lazy-error').first();
  88  |     await expect(container).toBeVisible();
  89  |   });
  90  | 
  91  |   test('should display comet celestial body', async ({ page, request }) => {
  92  |     const note = await request.post('http://localhost:8080/notes', {
  93  |       data: { title: 'Comet Node', content: 'Comet type', type: 'comet' }
  94  |     });
  95  |     const noteId = (await note.json()).id;
  96  |     
  97  |     await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
  98  |     await page.waitForLoadState('networkidle');
  99  |     await page.waitForTimeout(4000);
  100 |     
  101 |     const container = page.locator('.graph-3d-container, .lazy-error').first();
> 102 |     await expect(container).toBeVisible();
      |                             ^ Error: expect(locator).toBeVisible() failed
  103 |   });
  104 | 
  105 |   test('should display galaxy celestial body', async ({ page, request }) => {
  106 |     const note = await request.post('http://localhost:8080/notes', {
  107 |       data: { title: 'Galaxy Node', content: 'Galaxy type', type: 'galaxy' }
  108 |     });
  109 |     const noteId = (await note.json()).id;
  110 |     
  111 |     await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
  112 |     await page.waitForLoadState('networkidle');
  113 |     await page.waitForTimeout(4000);
  114 |     
  115 |     const container = page.locator('.graph-3d-container, .lazy-error').first();
  116 |     await expect(container).toBeVisible();
  117 |   });
  118 | 
  119 |   test('should render links with weight-based styling', async ({ page, request }) => {
  120 |     // Create notes with different link weights
  121 |     const sourceNote = await request.post('http://localhost:8080/notes', {
  122 |       data: { title: 'Source', content: 'Source note' }
  123 |     });
  124 |     const sourceId = (await sourceNote.json()).id;
  125 |     
  126 |     const strongTarget = await request.post('http://localhost:8080/notes', {
  127 |       data: { title: 'Strong Link', content: 'Strong connection' }
  128 |     });
  129 |     const strongId = (await strongTarget.json()).id;
  130 |     
  131 |     const weakTarget = await request.post('http://localhost:8080/notes', {
  132 |       data: { title: 'Weak Link', content: 'Weak connection' }
  133 |     });
  134 |     const weakId = (await weakTarget.json()).id;
  135 |     
  136 |     // Create links with different weights
  137 |     await request.post('http://localhost:8080/links', {
  138 |       data: { sourceNoteId: sourceId, targetNoteId: strongId, weight: 0.9 }
  139 |     });
  140 |     await request.post('http://localhost:8080/links', {
  141 |       data: { sourceNoteId: sourceId, targetNoteId: weakId, weight: 0.2 }
  142 |     });
  143 |     
  144 |     await page.goto(`http://localhost:5173/graph/3d/${sourceId}`);
  145 |     await page.waitForLoadState('networkidle');
  146 |     await page.waitForTimeout(3000);
  147 |     
  148 |     const container = page.locator('.graph-3d-container');
  149 |     await expect(container).toBeVisible();
  150 |   });
  151 | 
  152 |   test('should auto-zoom camera to fit graph', async ({ page, request }) => {
  153 |     // Create multiple notes to form a graph
  154 |     const notes = [];
  155 |     for (let i = 0; i < 5; i++) {
  156 |       const note = await request.post('http://localhost:8080/notes', {
  157 |         data: { title: `Note ${i}`, content: `Content ${i}` }
  158 |       });
  159 |       notes.push((await note.json()).id);
  160 |     }
  161 |     
  162 |     // Create links between notes
  163 |     for (let i = 0; i < notes.length - 1; i++) {
  164 |       await request.post('http://localhost:8080/links', {
  165 |         data: { sourceNoteId: notes[i], targetNoteId: notes[i + 1], weight: 0.5 }
  166 |       });
  167 |     }
  168 |     
  169 |     await page.goto(`http://localhost:5173/graph/3d/${notes[0]}`);
  170 |     await page.waitForLoadState('networkidle');
  171 |     await page.waitForTimeout(4000); // Wait for simulation to settle and auto-zoom
  172 |     
  173 |     const container = page.locator('.graph-3d-container');
  174 |     await expect(container).toBeVisible();
  175 |     
  176 |     // After auto-zoom, loading should be complete
  177 |     const loading = page.locator('.loading-overlay');
  178 |     const isLoading = await loading.isVisible().catch(() => false);
  179 |     expect(isLoading).toBe(false);
  180 |   });
  181 | 
  182 |   test('should handle full graph toggle', async ({ page, request }) => {
  183 |     const note = await request.post('http://localhost:8080/notes', {
  184 |       data: { title: 'Toggle Test', content: 'Testing full graph toggle' }
  185 |     });
  186 |     const noteId = (await note.json()).id;
  187 |     
  188 |     await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
  189 |     await page.waitForLoadState('networkidle');
  190 |     await page.waitForTimeout(2000);
  191 |     
  192 |     // Find and click the toggle
  193 |     const toggle = page.locator('.toggle input[type="checkbox"]').first();
  194 |     if (await toggle.isVisible().catch(() => false)) {
  195 |       await toggle.click();
  196 |       await page.waitForTimeout(2000);
  197 |       
  198 |       // Verify graph still renders after toggle
  199 |       const container = page.locator('.graph-3d-container');
  200 |       await expect(container).toBeVisible();
  201 |     }
  202 |   });
```