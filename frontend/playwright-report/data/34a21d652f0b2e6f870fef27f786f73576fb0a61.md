# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: notes.spec.ts >> Knowledge Graph Frontend >> should open graph for a note with links
- Location: tests\notes.spec.ts:64:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('[data-testid="main-graph-canvas"]')
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('[data-testid="main-graph-canvas"]')

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
      - region "Knowledge graph" [ref=e20]
  - navigation "Main navigation":
    - button "Новая звезда" [ref=e22] [cursor=pointer]:
      - img [ref=e23]
    - button "Граф" [ref=e24] [cursor=pointer]:
      - img [ref=e25]
    - button "Все заметки" [ref=e31] [cursor=pointer]:
      - img [ref=e32]
    - button "Импорт" [ref=e33] [cursor=pointer]:
      - img [ref=e34]
    - button "Экспорт" [ref=e37] [cursor=pointer]:
      - img [ref=e38]
    - button "Настройки" [ref=e41] [cursor=pointer]:
      - img [ref=e42]
  - generic [ref=e45]:
    - button "Обзор" [ref=e46] [cursor=pointer]:
      - generic [ref=e47]: Обзор
      - img [ref=e48]
    - generic [ref=e50]:
      - generic [ref=e51]:
        - generic [ref=e52]:
          - generic [ref=e53]: "0"
          - generic [ref=e54]: узлов
        - generic [ref=e56]:
          - generic [ref=e57]: "0"
          - generic [ref=e58]: связей
      - generic [ref=e61]:
        - img [ref=e62]
        - generic [ref=e68]: Мини-карта
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
  3   | test.describe('Knowledge Graph Frontend', () => {
  4   |   test.beforeEach(async ({ page }) => {
  5   |     await page.goto('http://localhost:5173');
  6   |     await page.waitForLoadState('networkidle');
  7   |   });
  8   | 
  9   |   test('should create a new note', async ({ page }) => {
  10  |     // Click new note button in left toolbar
  11  |     await page.click('[data-testid="toolbar-new-note"]');
  12  |     await page.fill('input[placeholder="Title"]', 'Playwright Test');
  13  |     await page.fill('textarea', 'Automated content');
  14  |     await page.click('button:has-text("Create")');
  15  |     // Wait for navigation to complete with explicit timeout
  16  |     await page.waitForURL(/\/notes\/[a-f0-9-]+/, { timeout: 5000 });
  17  |     await expect(page.locator('h1')).toHaveText('Playwright Test');
  18  |   });
  19  | 
  20  |   test('should edit a note', async ({ page }) => {
  21  |     // Сначала создадим заметку через API или UI
  22  |     await page.click('[data-testid="toolbar-new-note"]');
  23  |     await page.fill('input[placeholder="Title"]', 'To Edit');
  24  |     await page.fill('textarea', 'Original');
  25  |     await page.click('button:has-text("Create")');
  26  |     await page.waitForURL(/\/notes\/[a-f0-9-]+/, { timeout: 5000 });
  27  |     // Wait additional time for page to fully load
  28  |     await page.waitForTimeout(1000);
  29  | 
  30  |     await page.click('button:has-text("Редактировать")');
  31  |     await page.fill('input[placeholder="Title"]', 'Edited');
  32  |     await page.fill('textarea', 'New content');
  33  |     await page.click('button:has-text("Update")');
  34  |     await expect(page.locator('h1')).toHaveText('Edited');
  35  |     await expect(page.locator('.content')).toHaveText('New content');
  36  |   });
  37  | 
  38  |   test('should delete a note', async ({ page, request }) => {
  39  |     // Create a note via API first
  40  |     const timestamp = Date.now();
  41  |     const note = await request.post('http://localhost:8080/notes', {
  42  |       data: { title: 'Delete Test ' + timestamp, content: 'Test content for deletion' }
  43  |     });
  44  |     const noteId = (await note.json()).id;
  45  |     const noteTitle = 'Delete Test ' + timestamp;
  46  |     
  47  |     // Go to notes list to see the note
  48  |     await page.goto('http://localhost:5173/notes');
  49  |     await page.waitForSelector('.note-card', { timeout: 5000 });
  50  |     await expect(page.locator('text=' + noteTitle)).toBeVisible();
  51  |     
  52  |     // Delete the note via API
  53  |     await request.delete(`http://localhost:8080/notes/${noteId}`);
  54  |     
  55  |     // Wait and reload to see the changes
  56  |     await page.waitForTimeout(1000);
  57  |     await page.goto('http://localhost:5173/notes');
  58  |     await page.waitForLoadState('networkidle');
  59  |     
  60  |     // Check that the specific note is no longer present
  61  |     await expect(page.locator('text=' + noteTitle)).not.toBeVisible();
  62  |   });
  63  | 
  64  |   test('should open graph for a note with links', async ({ page, request }) => {
  65  |     // Сначала создадим две заметки и связь через API
  66  |     const note1 = await request.post('http://localhost:8080/notes', {
  67  |       data: { title: 'Node A', content: 'A' }
  68  |     });
  69  |     const note2 = await request.post('http://localhost:8080/notes', {
  70  |       data: { title: 'Node B', content: 'B' }
  71  |     });
  72  |     const id1 = (await note1.json()).id;
  73  |     const id2 = (await note2.json()).id;
  74  |     await request.post('http://localhost:8080/links', {
  75  |       data: { source_note_id: id1, target_note_id: id2, link_type: 'reference', weight: 1.0 }
  76  |     });
  77  | 
  78  |     await page.goto(`http://localhost:5173/graph/${id1}`);
> 79  |     await expect(page.locator('[data-testid="main-graph-canvas"]')).toBeVisible();
      |                                                                     ^ Error: expect(locator).toBeVisible() failed
  80  |     // Ждём, пока d3-force немного стабилизируется
  81  |     await page.waitForTimeout(1000);
  82  |     // Проверяем, что canvas не пустой
  83  |     const canvas = page.locator('[data-testid="main-graph-canvas"]');
  84  |     await expect(canvas).toBeVisible();
  85  |   });
  86  | 
  87  |   test('should show back button on note detail page', async ({ page, request }) => {
  88  |     // Create a note via API first
  89  |     const note = await request.post('http://localhost:8080/notes', {
  90  |       data: { title: 'Back Button Test', content: 'Testing back button functionality' }
  91  |     });
  92  |     const noteId = (await note.json()).id;
  93  |     
  94  |     // Navigate to note detail page
  95  |     await page.goto(`http://localhost:5173/notes/${noteId}`);
  96  |     await page.waitForTimeout(1000);
  97  | 
  98  |     // Check that back button is visible
  99  |     await expect(page.locator('[data-testid="back-button"]')).toBeVisible();
  100 |     
  101 |     // Test back button functionality - should go back to home page (graph)
  102 |     await page.click('[data-testid="back-button"]');
  103 |     await expect(page).toHaveURL('http://localhost:5173/');
  104 |   });
  105 | 
  106 |   test('should show back button on graph page', async ({ page, request }) => {
  107 |     // Create a note via API first
  108 |     const note = await request.post('http://localhost:8080/notes', {
  109 |       data: { title: 'Graph Back Test', content: 'Testing back button on graph' }
  110 |     });
  111 |     const noteId = (await note.json()).id;
  112 | 
  113 |     // First navigate to home page to create history
  114 |     await page.goto('http://localhost:5173/');
  115 |     await page.waitForTimeout(1000);
  116 |     
  117 |     // Then navigate to note page
  118 |     await page.goto(`http://localhost:5173/notes/${noteId}`);
  119 |     await page.waitForTimeout(1000);
  120 |     
  121 |     // Then navigate to graph page
  122 |     await page.goto(`http://localhost:5173/graph/${noteId}`);
  123 |     await page.waitForTimeout(1000);
  124 |     
  125 |     // Check that back button is visible
  126 |     await expect(page.locator('[data-testid="back-button"]')).toBeVisible();
  127 |     
  128 |     // Test back button functionality - should go back to note page using browser history
  129 |     await page.click('[data-testid="back-button"]');
  130 |     await page.waitForTimeout(1000);
  131 |     
  132 |     // Should be back on note page
  133 |     await expect(page).toHaveURL(`http://localhost:5173/notes/${noteId}`);
  134 |   });
  135 | 
  136 |   test('should use browser back when history exists', async ({ page, request }) => {
  137 |     // Create a note via API first
  138 |     const note = await request.post('http://localhost:8080/notes', {
  139 |       data: { title: 'History Test', content: 'Testing browser back functionality' }
  140 |     });
  141 |     const noteId = (await note.json()).id;
  142 |     const noteUrl = `http://localhost:5173/notes/${noteId}`;
  143 |     
  144 |     // Navigate to note page to create history
  145 |     await page.goto(noteUrl);
  146 |     await page.waitForTimeout(1000);
  147 |     
  148 |     // Navigate to home page to create history
  149 |     await page.goto('http://localhost:5173');
  150 |     await page.waitForTimeout(1000);
  151 |     
  152 |     // Go back to note page using browser history
  153 |     await page.goBack();
  154 |     await page.waitForTimeout(1000);
  155 |     await expect(page.locator('[data-testid="back-button"]')).toBeVisible();
  156 |     
  157 |     // Click back button - should use browser history to go home (graph page)
  158 |     await page.click('[data-testid="back-button"]');
  159 |     await page.waitForTimeout(2000);
  160 |     // Check if we're back on home page (graph)
  161 |     await expect(page.locator('.home-graph-page')).toBeVisible();
  162 |   });
  163 | });
  164 | 
```