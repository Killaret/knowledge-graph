# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: ui-components.spec.ts >> Graph Page Features >> should show note popup on node click
- Location: tests\ui-components.spec.ts:130:3

# Error details

```
TimeoutError: page.waitForSelector: Timeout 10000ms exceeded.
Call log:
  - waiting for locator('canvas') to be visible

```

# Page snapshot

```yaml
- generic [active] [ref=e1]:
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
  - generic [ref=e8]:
    - generic [ref=e9]: window is not defined
    - generic [ref=e10]: "ReferenceError: window is not defined at file:///D:/knowledge-graph/frontend/node_modules/three-forcegraph/dist/three-forcegraph.mjs:404:15 at ModuleJob.run (node:internal/modules/esm/module_job:271:25) at async onImport.tracePromise.__proto__ (node:internal/modules/esm/loader:547:26) at async nodeImport (file:///D:/knowledge-graph/frontend/node_modules/vite/dist/node/chunks/dep-BK3b2jBa.js:53105:15) at async ssrImport (file:///D:/knowledge-graph/frontend/node_modules/vite/dist/node/chunks/dep-BK3b2jBa.js:52963:16) at async eval (D:/knowledge-graph/frontend/src/lib/components/Graph3D.svelte:8:31) at async instantiateModule (file:///D:/knowledge-graph/frontend/node_modules/vite/dist/node/chunks/dep-BK3b2jBa.js:53021:5"
    - generic [ref=e11]:
      - text: Click outside, press Esc key, or fix the code to dismiss.
      - text: You can also disable this overlay by setting
      - code [ref=e12]: server.hmr.overlay
      - text: to
      - code [ref=e13]: "false"
      - text: in
      - code [ref=e14]: vite.config.ts
      - text: .
```

# Test source

```ts
  47  |     await expect(page.locator('text=Импорт документа')).toBeVisible();
  48  |     await expect(page.locator('text=Перетащите PDF или файл сюда')).toBeVisible();
  49  |     
  50  |     // Close modal
  51  |     await page.click('button[aria-label="Закрыть"]');
  52  |     await expect(page.locator('text=Импорт документа')).not.toBeVisible();
  53  |   });
  54  | 
  55  |   test('should open export modal', async ({ page }) => {
  56  |     // Create a note first via API
  57  |     await page.request.post('http://localhost:8080/notes', {
  58  |       data: { title: 'Export Test Note', content: 'Test content' }
  59  |     });
  60  |     
  61  |     // Refresh page to see the note
  62  |     await page.goto('http://localhost:5173');
  63  |     await page.waitForLoadState('networkidle');
  64  |     
  65  |     // Open sidebar
  66  |     await page.click('button[aria-label="Открыть меню"]');
  67  |     
  68  |     // Click export
  69  |     await page.click('text=Экспорт');
  70  |     
  71  |     // Check modal is open
  72  |     await expect(page.locator('text=Экспорт заметок')).toBeVisible();
  73  |     await expect(page.locator('text=JSON')).toBeVisible();
  74  |     await expect(page.locator('text=CSV')).toBeVisible();
  75  |     await expect(page.locator('text=Markdown')).toBeVisible();
  76  |     
  77  |     // Close modal
  78  |     await page.click('button[aria-label="Закрыть"]');
  79  |   });
  80  | 
  81  |   test('should show empty state when no notes', async ({ page, request }) => {
  82  |     // Clear all notes via API
  83  |     const notes = await request.get('http://localhost:8080/notes');
  84  |     const notesData = await notes.json();
  85  |     for (const note of notesData) {
  86  |       await request.delete(`http://localhost:8080/notes/${note.id}`);
  87  |     }
  88  |     
  89  |     // Refresh page
  90  |     await page.goto('http://localhost:5173');
  91  |     await page.waitForLoadState('networkidle');
  92  |     
  93  |     // Check empty state
  94  |     await expect(page.locator('text=Нет данных')).toBeVisible();
  95  |     await expect(page.locator('text=Создайте первую заметку')).toBeVisible();
  96  |   });
  97  | 
  98  |   test('should use keyboard shortcuts', async ({ page }) => {
  99  |     // Test Ctrl+N - should navigate to new note
  100 |     await page.keyboard.press('Control+n');
  101 |     await page.waitForURL('**/notes/new', { timeout: 5000 });
  102 |     await expect(page.locator('text=New Note')).toBeVisible();
  103 |     
  104 |     // Go back
  105 |     await page.goto('http://localhost:5173');
  106 |     
  107 |     // Test Ctrl+F - should open search
  108 |     await page.keyboard.press('Control+f');
  109 |     await expect(page.locator('input[placeholder="Поиск по заметкам..."]')).toBeVisible();
  110 |     
  111 |     // Test Escape - should close search
  112 |     await page.keyboard.press('Escape');
  113 |     await expect(page.locator('input[placeholder="Поиск по заметкам..."]')).not.toBeVisible();
  114 |   });
  115 | 
  116 |   test('should create note via FAB', async ({ page }) => {
  117 |     await page.click('[data-testid="fab-new-note"]');
  118 |     await page.waitForURL('**/notes/new', { timeout: 5000 });
  119 |     
  120 |     await page.fill('input[placeholder="Title"]', 'FAB Created Note');
  121 |     await page.fill('textarea', 'Content from FAB test');
  122 |     await page.click('button:has-text("Create")');
  123 |     
  124 |     await page.waitForURL(/\/notes\/[a-f0-9-]+/, { timeout: 5000 });
  125 |     await expect(page.locator('h1')).toHaveText('FAB Created Note');
  126 |   });
  127 | });
  128 | 
  129 | test.describe('Graph Page Features', () => {
  130 |   test('should show note popup on node click', async ({ page, request }) => {
  131 |     // Create two notes and a link
  132 |     const note1 = await request.post('http://localhost:8080/notes', {
  133 |       data: { title: 'Graph Node A', content: 'Content A' }
  134 |     });
  135 |     const note2 = await request.post('http://localhost:8080/notes', {
  136 |       data: { title: 'Graph Node B', content: 'Content B' }
  137 |     });
  138 |     const id1 = (await note1.json()).id;
  139 |     const id2 = (await note2.json()).id;
  140 |     
  141 |     await request.post('http://localhost:8080/links', {
  142 |       data: { source_note_id: id1, target_note_id: id2, link_type: 'reference', weight: 1.0 }
  143 |     });
  144 |     
  145 |     // Navigate to graph
  146 |     await page.goto(`http://localhost:5173/graph/${id1}`);
> 147 |     await page.waitForSelector('canvas', { timeout: 10000 });
      |                ^ TimeoutError: page.waitForSelector: Timeout 10000ms exceeded.
  148 |     await page.waitForTimeout(2000); // Let graph render
  149 |     
  150 |     // Canvas should be visible
  151 |     await expect(page.locator('canvas')).toBeVisible();
  152 |     await expect(page.locator('text=Knowledge Constellation')).toBeVisible();
  153 |   });
  154 | 
  155 |   test('should show empty state for isolated node', async ({ page, request }) => {
  156 |     // Create single note without links
  157 |     const note = await request.post('http://localhost:8080/notes', {
  158 |       data: { title: 'Isolated Node', content: 'No connections' }
  159 |     });
  160 |     const id = (await note.json()).id;
  161 |     
  162 |     // Navigate to graph
  163 |     await page.goto(`http://localhost:5173/graph/${id}`);
  164 |     await page.waitForLoadState('networkidle');
  165 |     
  166 |     // Should show empty state
  167 |     await expect(page.locator('text=Это одинокая звезда')).toBeVisible();
  168 |     await expect(page.locator('text=Создайте связи, чтобы увидеть созвездие')).toBeVisible();
  169 |   });
  170 | });
  171 | 
```