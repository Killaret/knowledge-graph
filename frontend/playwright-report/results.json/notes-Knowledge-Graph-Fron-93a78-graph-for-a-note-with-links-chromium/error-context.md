# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: notes.spec.ts >> Knowledge Graph Frontend >> should open graph for a note with links
- Location: tests\notes.spec.ts:63:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('canvas')
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('canvas')

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
  1   | import { test, expect } from '@playwright/test';
  2   | 
  3   | test.describe('Knowledge Graph Frontend', () => {
  4   |   test.beforeEach(async ({ page }) => {
  5   |     await page.goto('http://localhost:5173');
  6   |     await page.waitForLoadState('networkidle');
  7   |   });
  8   | 
  9   |   test('should create a new note', async ({ page }) => {
  10  |     await page.click('[data-testid="fab-new-note"]');
  11  |     await page.fill('input[placeholder="Title"]', 'Playwright Test');
  12  |     await page.fill('textarea', 'Automated content');
  13  |     await page.click('button:has-text("Create")');
  14  |     // Wait for navigation to complete with explicit timeout
  15  |     await page.waitForURL(/\/notes\/[a-f0-9-]+/, { timeout: 5000 });
  16  |     await expect(page.locator('h1')).toHaveText('Playwright Test');
  17  |   });
  18  | 
  19  |   test('should edit a note', async ({ page }) => {
  20  |     // Сначала создадим заметку через API или UI
  21  |     await page.click('[data-testid="fab-new-note"]');
  22  |     await page.fill('input[placeholder="Title"]', 'To Edit');
  23  |     await page.fill('textarea', 'Original');
  24  |     await page.click('button:has-text("Create")');
  25  |     await page.waitForURL(/\/notes\/[a-f0-9-]+/, { timeout: 5000 });
  26  |     // Wait additional time for page to fully load
  27  |     await page.waitForTimeout(1000);
  28  | 
  29  |     await page.click('a:has-text("Edit")');
  30  |     await page.fill('input[placeholder="Title"]', 'Edited');
  31  |     await page.fill('textarea', 'New content');
  32  |     await page.click('button:has-text("Update")');
  33  |     await expect(page.locator('h1')).toHaveText('Edited');
  34  |     await expect(page.locator('.content')).toHaveText('New content');
  35  |   });
  36  | 
  37  |   test('should delete a note', async ({ page, request }) => {
  38  |     // Create a note via API first
  39  |     const timestamp = Date.now();
  40  |     const note = await request.post('http://localhost:8080/notes', {
  41  |       data: { title: 'Delete Test ' + timestamp, content: 'Test content for deletion' }
  42  |     });
  43  |     const noteId = (await note.json()).id;
  44  |     const noteTitle = 'Delete Test ' + timestamp;
  45  |     
  46  |     // Go to home page to see the note in the list
  47  |     await page.goto('http://localhost:5173/');
  48  |     await page.waitForSelector('.note-card', { timeout: 5000 });
  49  |     await expect(page.locator('text=' + noteTitle)).toBeVisible();
  50  |     
  51  |     // Delete the note via API
  52  |     await request.delete(`http://localhost:8080/notes/${noteId}`);
  53  |     
  54  |     // Wait and reload to see the changes
  55  |     await page.waitForTimeout(1000);
  56  |     await page.goto('http://localhost:5173/');
  57  |     await page.waitForLoadState('networkidle');
  58  |     
  59  |     // Check that the specific note is no longer present
  60  |     await expect(page.locator('text=' + noteTitle)).not.toBeVisible();
  61  |   });
  62  | 
  63  |   test('should open graph for a note with links', async ({ page, request }) => {
  64  |     // Сначала создадим две заметки и связь через API
  65  |     const note1 = await request.post('http://localhost:8080/notes', {
  66  |       data: { title: 'Node A', content: 'A' }
  67  |     });
  68  |     const note2 = await request.post('http://localhost:8080/notes', {
  69  |       data: { title: 'Node B', content: 'B' }
  70  |     });
  71  |     const id1 = (await note1.json()).id;
  72  |     const id2 = (await note2.json()).id;
  73  |     await request.post('http://localhost:8080/links', {
  74  |       data: { source_note_id: id1, target_note_id: id2, link_type: 'reference', weight: 1.0 }
  75  |     });
  76  | 
  77  |     await page.goto(`http://localhost:5173/graph/${id1}`);
> 78  |     await expect(page.locator('canvas')).toBeVisible();
      |                                          ^ Error: expect(locator).toBeVisible() failed
  79  |     // Ждём, пока d3-force немного стабилизируется
  80  |     await page.waitForTimeout(1000);
  81  |     // Проверяем, что canvas не пустой (можно по цвету пикселя, но сложно)
  82  |     const canvas = page.locator('canvas');
  83  |     await expect(canvas).toBeVisible();
  84  |   });
  85  | 
  86  |   test('should show back button on note detail page', async ({ page, request }) => {
  87  |     // Create a note via API first
  88  |     const note = await request.post('http://localhost:8080/notes', {
  89  |       data: { title: 'Back Button Test', content: 'Testing back button functionality' }
  90  |     });
  91  |     const noteId = (await note.json()).id;
  92  |     
  93  |     // Navigate to note detail page
  94  |     await page.goto(`http://localhost:5173/notes/${noteId}`);
  95  |     await page.waitForTimeout(1000);
  96  | 
  97  |     // Check that back button is visible
  98  |     await expect(page.locator('button:has-text("Back")')).toBeVisible();
  99  |     
  100 |     // Test back button functionality - should go back to home page
  101 |     await page.click('button:has-text("Back")');
  102 |     await expect(page).toHaveURL('http://localhost:5173/');
  103 |   });
  104 | 
  105 |   test('should show back button on graph page', async ({ page, request }) => {
  106 |     // Create a note via API first
  107 |     const note = await request.post('http://localhost:8080/notes', {
  108 |       data: { title: 'Graph Back Test', content: 'Testing back button on graph' }
  109 |     });
  110 |     const noteId = (await note.json()).id;
  111 | 
  112 |     // First navigate to home page to create history
  113 |     await page.goto('http://localhost:5173/');
  114 |     await page.waitForTimeout(1000);
  115 |     
  116 |     // Then navigate to note page
  117 |     await page.goto(`http://localhost:5173/notes/${noteId}`);
  118 |     await page.waitForTimeout(1000);
  119 |     
  120 |     // Then navigate to graph page
  121 |     await page.goto(`http://localhost:5173/graph/${noteId}`);
  122 |     await page.waitForTimeout(1000);
  123 |     
  124 |     // Check that back button is visible
  125 |     await expect(page.locator('button:has-text("Back")')).toBeVisible();
  126 |     
  127 |     // Test back button functionality - should go back to note page using browser history
  128 |     await page.click('button:has-text("Back")');
  129 |     await page.waitForTimeout(1000);
  130 |     
  131 |     // Should be back on note page
  132 |     await expect(page).toHaveURL(`http://localhost:5173/notes/${noteId}`);
  133 |   });
  134 | 
  135 |   test('should use browser back when history exists', async ({ page, request }) => {
  136 |     // Create a note via API first
  137 |     const note = await request.post('http://localhost:8080/notes', {
  138 |       data: { title: 'History Test', content: 'Testing browser back functionality' }
  139 |     });
  140 |     const noteId = (await note.json()).id;
  141 |     const noteUrl = `http://localhost:5173/notes/${noteId}`;
  142 |     
  143 |     // Navigate to note page to create history
  144 |     await page.goto(noteUrl);
  145 |     await page.waitForTimeout(1000);
  146 |     
  147 |     // Navigate to home page to create history
  148 |     await page.goto('http://localhost:5173');
  149 |     await page.waitForTimeout(1000);
  150 |     
  151 |     // Go back to note page using browser history
  152 |     await page.goBack();
  153 |     await page.waitForTimeout(1000);
  154 |     await expect(page.locator('button:has-text("Back")')).toBeVisible();
  155 |     
  156 |     // Click back button - should use browser history to go home
  157 |     await page.click('button:has-text("Back")');
  158 |     await page.waitForTimeout(2000);
  159 |     // Check if we're back on home page
  160 |     await expect(page.locator('h1')).toHaveText('My Notes');
  161 |   });
  162 | });
  163 | 
```