# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: ui-components.spec.ts >> UI Components from Flowchart >> should open left sidebar menu
- Location: tests\ui-components.spec.ts:9:3

# Error details

```
Test timeout of 30000ms exceeded.
```

```
Error: page.click: Test timeout of 30000ms exceeded.
Call log:
  - waiting for locator('button[aria-label="Открыть меню"]')
    - locator resolved to <button aria-expanded="false" aria-label="Открыть меню" class="menu-btn s-jXmVQErCnyUA">…</button>
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
    55 × waiting for element to be visible, enabled and stable
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
    - button "Открыть меню" [ref=e3] [cursor=pointer]:
      - img [ref=e4]
    - button "Открыть поиск" [ref=e5] [cursor=pointer]:
      - img [ref=e6]
      - text: Поиск
    - heading "My Notes" [level=1] [ref=e9]
    - generic [ref=e10]:
      - searchbox "Search" [ref=e11]
      - button "Search" [ref=e12] [cursor=pointer]
    - generic [ref=e13]:
      - generic [ref=e14]:
        - link "Back Button Test Testing back button functionality..." [ref=e15] [cursor=pointer]:
          - /url: /notes/0cc2906a-43be-48ce-946e-d8d511c35acf
          - heading "Back Button Test" [level=3] [ref=e16]
          - paragraph [ref=e17]: Testing back button functionality...
        - generic [ref=e18]:
          - link "Edit" [ref=e19] [cursor=pointer]:
            - /url: /notes/0cc2906a-43be-48ce-946e-d8d511c35acf/edit
          - button "Delete" [ref=e20]
      - generic [ref=e21]:
        - link "Graph Back Test Testing back button on graph..." [ref=e22] [cursor=pointer]:
          - /url: /notes/7337d92a-4646-408f-ac75-28299fbef788
          - heading "Graph Back Test" [level=3] [ref=e23]
          - paragraph [ref=e24]: Testing back button on graph...
        - generic [ref=e25]:
          - link "Edit" [ref=e26] [cursor=pointer]:
            - /url: /notes/7337d92a-4646-408f-ac75-28299fbef788/edit
          - button "Delete" [ref=e27]
      - generic [ref=e28]:
        - link "Node B B..." [ref=e29] [cursor=pointer]:
          - /url: /notes/e554b1dc-d76b-4811-a237-78eefcddcef0
          - heading "Node B" [level=3] [ref=e30]
          - paragraph [ref=e31]: B...
        - generic [ref=e32]:
          - link "Edit" [ref=e33] [cursor=pointer]:
            - /url: /notes/e554b1dc-d76b-4811-a237-78eefcddcef0/edit
          - button "Delete" [ref=e34]
      - generic [ref=e35]:
        - link "Node A A..." [ref=e36] [cursor=pointer]:
          - /url: /notes/a5229876-f4b0-48c8-8053-896fb3f275e7
          - heading "Node A" [level=3] [ref=e37]
          - paragraph [ref=e38]: A...
        - generic [ref=e39]:
          - link "Edit" [ref=e40] [cursor=pointer]:
            - /url: /notes/a5229876-f4b0-48c8-8053-896fb3f275e7/edit
          - button "Delete" [ref=e41]
      - generic [ref=e42]:
        - link "History Test Testing browser back functionality..." [ref=e43] [cursor=pointer]:
          - /url: /notes/7fbedd72-c477-471c-a017-502f1713dcdc
          - heading "History Test" [level=3] [ref=e44]
          - paragraph [ref=e45]: Testing browser back functionality...
        - generic [ref=e46]:
          - link "Edit" [ref=e47] [cursor=pointer]:
            - /url: /notes/7fbedd72-c477-471c-a017-502f1713dcdc/edit
          - button "Delete" [ref=e48]
      - generic [ref=e49]:
        - link "Edited New content..." [ref=e50] [cursor=pointer]:
          - /url: /notes/7a16289d-f238-4a17-99ec-885cf95f50d5
          - heading "Edited" [level=3] [ref=e51]
          - paragraph [ref=e52]: New content...
        - generic [ref=e53]:
          - link "Edit" [ref=e54] [cursor=pointer]:
            - /url: /notes/7a16289d-f238-4a17-99ec-885cf95f50d5/edit
          - button "Delete" [ref=e55]
      - generic [ref=e56]:
        - link "Playwright Test Automated content..." [ref=e57] [cursor=pointer]:
          - /url: /notes/e22213e6-bea4-4ef0-a6f1-fcd46caba0bc
          - heading "Playwright Test" [level=3] [ref=e58]
          - paragraph [ref=e59]: Automated content...
        - generic [ref=e60]:
          - link "Edit" [ref=e61] [cursor=pointer]:
            - /url: /notes/e22213e6-bea4-4ef0-a6f1-fcd46caba0bc/edit
          - button "Delete" [ref=e62]
      - generic [ref=e63]:
        - link "Navigation Test 1775935641618 Testing navigation..." [ref=e64] [cursor=pointer]:
          - /url: /notes/ac2f07a2-607c-4154-8306-bf7c52531b5c
          - heading "Navigation Test 1775935641618" [level=3] [ref=e65]
          - paragraph [ref=e66]: Testing navigation...
        - generic [ref=e67]:
          - link "Edit" [ref=e68] [cursor=pointer]:
            - /url: /notes/ac2f07a2-607c-4154-8306-bf7c52531b5c/edit
          - button "Delete" [ref=e69]
      - generic [ref=e70]:
        - link "Isolated Node No connections..." [ref=e71] [cursor=pointer]:
          - /url: /notes/d432fe33-8c61-4a21-a5a0-1c3e3b805e66
          - heading "Isolated Node" [level=3] [ref=e72]
          - paragraph [ref=e73]: No connections...
        - generic [ref=e74]:
          - link "Edit" [ref=e75] [cursor=pointer]:
            - /url: /notes/d432fe33-8c61-4a21-a5a0-1c3e3b805e66/edit
          - button "Delete" [ref=e76]
      - generic [ref=e77]:
        - link "Graph Node B Content B..." [ref=e78] [cursor=pointer]:
          - /url: /notes/2b01d96c-b64b-4271-aa91-fa08063fafa8
          - heading "Graph Node B" [level=3] [ref=e79]
          - paragraph [ref=e80]: Content B...
        - generic [ref=e81]:
          - link "Edit" [ref=e82] [cursor=pointer]:
            - /url: /notes/2b01d96c-b64b-4271-aa91-fa08063fafa8/edit
          - button "Delete" [ref=e83]
      - generic [ref=e84]:
        - link "Graph Node A Content A..." [ref=e85] [cursor=pointer]:
          - /url: /notes/1f41ec64-ef6f-4d07-ab25-10df43a7c04b
          - heading "Graph Node A" [level=3] [ref=e86]
          - paragraph [ref=e87]: Content A...
        - generic [ref=e88]:
          - link "Edit" [ref=e89] [cursor=pointer]:
            - /url: /notes/1f41ec64-ef6f-4d07-ab25-10df43a7c04b/edit
          - button "Delete" [ref=e90]
    - button "Создать заметку" [ref=e91] [cursor=pointer]:
      - img [ref=e92]
    - generic:
      - generic: Ctrl+N
      - text: — новая заметка
      - generic: Ctrl+F
      - text: — поиск
      - generic: Esc
      - text: — закрыть
  - generic [ref=e96]:
    - generic [ref=e97]: window is not defined
    - generic [ref=e98]: "ReferenceError: window is not defined at file:///D:/knowledge-graph/frontend/node_modules/three-forcegraph/dist/three-forcegraph.mjs:404:15 at ModuleJob.run (node:internal/modules/esm/module_job:271:25) at async onImport.tracePromise.__proto__ (node:internal/modules/esm/loader:547:26) at async nodeImport (file:///D:/knowledge-graph/frontend/node_modules/vite/dist/node/chunks/dep-BK3b2jBa.js:53105:15) at async ssrImport (file:///D:/knowledge-graph/frontend/node_modules/vite/dist/node/chunks/dep-BK3b2jBa.js:52963:16) at async eval (D:/knowledge-graph/frontend/src/lib/components/Graph3D.svelte:8:31) at async instantiateModule (file:///D:/knowledge-graph/frontend/node_modules/vite/dist/node/chunks/dep-BK3b2jBa.js:53021:5"
    - generic [ref=e99]:
      - text: Click outside, press Esc key, or fix the code to dismiss.
      - text: You can also disable this overlay by setting
      - code [ref=e100]: server.hmr.overlay
      - text: to
      - code [ref=e101]: "false"
      - text: in
      - code [ref=e102]: vite.config.ts
      - text: .
```

# Test source

```ts
  1   | import { test, expect } from '@playwright/test';
  2   | 
  3   | test.describe('UI Components from Flowchart', () => {
  4   |   test.beforeEach(async ({ page }) => {
  5   |     await page.goto('http://localhost:5173');
  6   |     await page.waitForLoadState('networkidle');
  7   |   });
  8   | 
  9   |   test('should open left sidebar menu', async ({ page }) => {
  10  |     // Click hamburger menu
> 11  |     await page.click('button[aria-label="Открыть меню"]');
      |                ^ Error: page.click: Test timeout of 30000ms exceeded.
  12  |     
  13  |     // Check sidebar is visible
  14  |     await expect(page.locator('nav[aria-label="Главное меню"]')).toBeVisible();
  15  |     await expect(page.locator('text=Импорт')).toBeVisible();
  16  |     await expect(page.locator('text=Экспорт')).toBeVisible();
  17  |     
  18  |     // Close by clicking overlay
  19  |     await page.click('.sidebar-overlay');
  20  |     await expect(page.locator('nav[aria-label="Главное меню"]')).not.toBeVisible();
  21  |   });
  22  | 
  23  |   test('should open search panel', async ({ page }) => {
  24  |     // Click search button
  25  |     await page.click('button[aria-label="Открыть поиск"]');
  26  |     
  27  |     // Check search panel is visible
  28  |     await expect(page.locator('input[placeholder="Поиск по заметкам..."]')).toBeVisible();
  29  |     
  30  |     // Type search query
  31  |     await page.fill('input[placeholder="Поиск по заметкам..."]', 'test');
  32  |     await page.waitForTimeout(600); // Wait for debounce
  33  |     
  34  |     // Close search
  35  |     await page.keyboard.press('Escape');
  36  |     await expect(page.locator('input[placeholder="Поиск по заметкам..."]')).not.toBeVisible();
  37  |   });
  38  | 
  39  |   test('should open document import modal', async ({ page }) => {
  40  |     // Open sidebar
  41  |     await page.click('button[aria-label="Открыть меню"]');
  42  |     
  43  |     // Click import
  44  |     await page.click('text=Импорт');
  45  |     
  46  |     // Check modal is open
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
```