# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: notes.spec.ts >> Knowledge Graph Frontend >> should use browser back when history exists
- Location: tests\notes.spec.ts:135:3

# Error details

```
Test timeout of 30000ms exceeded.
```

```
Error: page.click: Test timeout of 30000ms exceeded.
Call log:
  - waiting for locator('button:has-text("Back")')
    - locator resolved to <button aria-label="Go back" class="back-button s-BYOBscIiO5Lg">Back</button>
  - attempting click action
    2 × waiting for element to be visible, enabled and stable
      - element is visible, enabled and stable
      - scrolling into view if needed
      - done scrolling
      - <vite-error-overlay></vite-error-overlay> intercepts pointer events
    - retrying click action
    - waiting 20ms
    2 × waiting for element to be visible, enabled and stable
      - element is visible, enabled and stable
      - scrolling into view if needed
      - done scrolling
      - <vite-error-overlay></vite-error-overlay> intercepts pointer events
    - retrying click action
      - waiting 100ms
    49 × waiting for element to be visible, enabled and stable
       - element is visible, enabled and stable
       - scrolling into view if needed
       - done scrolling
       - <vite-error-overlay></vite-error-overlay> intercepts pointer events
     - retrying click action
       - waiting 500ms

```

# Page snapshot

```yaml
- generic [active] [ref=e1]:
  - generic [ref=e2]:
    - generic [ref=e3]:
      - button "Go back" [ref=e4] [cursor=pointer]: « Back
      - heading "History Test" [level=1] [ref=e5]
      - generic [ref=e6]: "Created: 4/11/2026, 10:27:30 PM"
      - generic [ref=e7]: Testing browser back functionality
      - generic [ref=e8]:
        - link "Edit" [ref=e9] [cursor=pointer]:
          - /url: /notes/7fbedd72-c477-471c-a017-502f1713dcdc/edit
        - button "Delete" [ref=e10]
        - link "✨ Show constellation" [ref=e11] [cursor=pointer]:
          - /url: /graph/7fbedd72-c477-471c-a017-502f1713dcdc
      - heading "Similar notes" [level=2] [ref=e12]
      - list [ref=e13]:
        - listitem [ref=e14]:
          - link "Navigation Test 1775935641618" [ref=e15] [cursor=pointer]:
            - /url: /notes/ac2f07a2-607c-4154-8306-bf7c52531b5c
          - text: "score: 0.074"
        - listitem [ref=e16]:
          - link "Playwright Test" [ref=e17] [cursor=pointer]:
            - /url: /notes/e22213e6-bea4-4ef0-a6f1-fcd46caba0bc
          - text: "score: 0.071"
        - listitem [ref=e18]:
          - link "Edited" [ref=e19] [cursor=pointer]:
            - /url: /notes/7a16289d-f238-4a17-99ec-885cf95f50d5
          - text: "score: 0.021"
        - listitem [ref=e20]:
          - link "Graph Node B" [ref=e21] [cursor=pointer]:
            - /url: /notes/2b01d96c-b64b-4271-aa91-fa08063fafa8
          - text: "score: -0.002"
        - listitem [ref=e22]:
          - link "Graph Node A" [ref=e23] [cursor=pointer]:
            - /url: /notes/1f41ec64-ef6f-4d07-ab25-10df43a7c04b
          - text: "score: -0.003"
    - generic:
      - generic: Ctrl+N
      - text: — новая заметка
      - generic: Ctrl+F
      - text: — поиск
      - generic: Esc
      - text: — закрыть
  - generic [ref=e27]:
    - generic [ref=e28]: window is not defined
    - generic [ref=e29]: "ReferenceError: window is not defined at file:///D:/knowledge-graph/frontend/node_modules/three-forcegraph/dist/three-forcegraph.mjs:404:15 at ModuleJob.run (node:internal/modules/esm/module_job:271:25) at async onImport.tracePromise.__proto__ (node:internal/modules/esm/loader:547:26) at async nodeImport (file:///D:/knowledge-graph/frontend/node_modules/vite/dist/node/chunks/dep-BK3b2jBa.js:53105:15) at async ssrImport (file:///D:/knowledge-graph/frontend/node_modules/vite/dist/node/chunks/dep-BK3b2jBa.js:52963:16) at async eval (D:/knowledge-graph/frontend/src/lib/components/Graph3D.svelte:8:31) at async instantiateModule (file:///D:/knowledge-graph/frontend/node_modules/vite/dist/node/chunks/dep-BK3b2jBa.js:53021:5"
    - generic [ref=e30]:
      - text: Click outside, press Esc key, or fix the code to dismiss.
      - text: You can also disable this overlay by setting
      - code [ref=e31]: server.hmr.overlay
      - text: to
      - code [ref=e32]: "false"
      - text: in
      - code [ref=e33]: vite.config.ts
      - text: .
```

# Test source

```ts
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
  78  |     await expect(page.locator('canvas')).toBeVisible();
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
> 157 |     await page.click('button:has-text("Back")');
      |                ^ Error: page.click: Test timeout of 30000ms exceeded.
  158 |     await page.waitForTimeout(2000);
  159 |     // Check if we're back on home page
  160 |     await expect(page.locator('h1')).toHaveText('My Notes');
  161 |   });
  162 | });
  163 | 
```