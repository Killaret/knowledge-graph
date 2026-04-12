# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: ui-components.spec.ts >> UI Components from Flowchart >> should show empty state when no notes
- Location: tests\ui-components.spec.ts:61:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('text=Нет заметок')
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('text=Нет заметок')

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
    - generic [ref=e16]:
      - generic [ref=e17]: 🌌
      - heading "Knowledge Constellation" [level=1] [ref=e18]
      - paragraph [ref=e19]: Ваше созвездие знаний пусто. Создайте первую звезду, чтобы начать.
      - button "Создать первую звезду" [ref=e21] [cursor=pointer]:
        - img [ref=e22]
        - text: Создать первую звезду
  - navigation "Main navigation" [ref=e23]:
    - button "Новая звезда" [ref=e24] [cursor=pointer]:
      - img [ref=e25]
    - button "Граф" [ref=e27] [cursor=pointer]:
      - img [ref=e28]
    - button "Все заметки" [ref=e34] [cursor=pointer]:
      - img [ref=e35]
    - button "Импорт" [ref=e36] [cursor=pointer]:
      - img [ref=e37]
    - button "Экспорт" [ref=e40] [cursor=pointer]:
      - img [ref=e41]
    - button "Настройки" [ref=e44] [cursor=pointer]:
      - img [ref=e45]
  - generic [ref=e48]:
    - button "Обзор" [ref=e49] [cursor=pointer]:
      - generic [ref=e50]: Обзор
      - img [ref=e51]
    - generic [ref=e53]:
      - generic [ref=e54]:
        - generic [ref=e55]:
          - generic [ref=e56]: "0"
          - generic [ref=e57]: узлов
        - generic [ref=e59]:
          - generic [ref=e60]: "0"
          - generic [ref=e61]: связей
      - generic [ref=e64]:
        - img [ref=e65]
        - generic [ref=e71]: Мини-карта
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
  3   | test.describe('UI Components from Flowchart', () => {
  4   |   test.beforeEach(async ({ page }) => {
  5   |     await page.goto('http://localhost:5173');
  6   |     await page.waitForLoadState('networkidle');
  7   |   });
  8   | 
  9   |   test('should show left toolbar with navigation', async ({ page }) => {
  10  |     // Check toolbar is visible
  11  |     await expect(page.locator('nav[aria-label="Main navigation"]')).toBeVisible();
  12  |     
  13  |     // Check toolbar buttons
  14  |     await expect(page.locator('button[aria-label="Новая звезда"]')).toBeVisible();
  15  |     await expect(page.locator('button[aria-label="Граф"]')).toBeVisible();
  16  |     await expect(page.locator('button[aria-label="Все заметки"]')).toBeVisible();
  17  |     await expect(page.locator('button[aria-label="Импорт"]')).toBeVisible();
  18  |     await expect(page.locator('button[aria-label="Экспорт"]')).toBeVisible();
  19  |     await expect(page.locator('button[aria-label="Настройки"]')).toBeVisible();
  20  |   });
  21  | 
  22  |   test('should open search from header', async ({ page }) => {
  23  |     // Click search trigger button in header
  24  |     await page.click('button[aria-label="Поиск (Ctrl+F)"]');
  25  |     
  26  |     // Check search expanded is visible
  27  |     await expect(page.locator('input[placeholder="Поиск по заметкам..."]')).toBeVisible();
  28  |     
  29  |     // Type search query
  30  |     await page.fill('input[placeholder="Поиск по заметкам..."]', 'test');
  31  |     await page.waitForTimeout(600); // Wait for debounce
  32  |     
  33  |     // Close search
  34  |     await page.keyboard.press('Escape');
  35  |   });
  36  | 
  37  |   test('should open document import modal from toolbar', async ({ page }) => {
  38  |     // Click import button in left toolbar
  39  |     await page.click('button[aria-label="Импорт"]');
  40  |     await page.waitForTimeout(300); // Wait for tooltip/handler
  41  |     
  42  |     // Note: Import modal is now triggered via toolbar onImport callback
  43  |     // This test verifies the toolbar button is present and clickable
  44  |     await expect(page.locator('button[aria-label="Импорт"]')).toBeVisible();
  45  |   });
  46  | 
  47  |   test('should open export modal from toolbar', async ({ page }) => {
  48  |     // Create a note first via API
  49  |     await page.request.post('http://localhost:8080/notes', {
  50  |       data: { title: 'Export Test Note', content: 'Test content' }
  51  |     });
  52  |     
  53  |     // Click export button in left toolbar
  54  |     await page.click('button[aria-label="Экспорт"]');
  55  |     await page.waitForTimeout(300);
  56  |     
  57  |     // Verify export button is present
  58  |     await expect(page.locator('button[aria-label="Экспорт"]')).toBeVisible();
  59  |   });
  60  | 
  61  |   test('should show empty state when no notes', async ({ page, request }) => {
  62  |     // Clear all notes via API
  63  |     const notes = await request.get('http://localhost:8080/notes');
  64  |     const notesData = await notes.json();
  65  |     for (const note of notesData) {
  66  |       await request.delete(`http://localhost:8080/notes/${note.id}`);
  67  |     }
  68  |     
  69  |     // Refresh page
  70  |     await page.goto('http://localhost:5173');
  71  |     await page.waitForLoadState('networkidle');
  72  |     
  73  |     // Check empty state
> 74  |     await expect(page.locator('text=Нет заметок')).toBeVisible();
      |                                                    ^ Error: expect(locator).toBeVisible() failed
  75  |     await expect(page.locator('text=Создайте первую заметку, чтобы начать')).toBeVisible();
  76  |   });
  77  | 
  78  |   test('should use keyboard shortcuts', async ({ page }) => {
  79  |     // Test Ctrl+F - should open search
  80  |     await page.keyboard.press('Control+f');
  81  |     await expect(page.locator('input[placeholder="Поиск по заметкам..."]')).toBeVisible();
  82  |     
  83  |     // Test Escape - should close search
  84  |     await page.keyboard.press('Escape');
  85  |     await expect(page.locator('input[placeholder="Поиск по заметкам..."]')).not.toBeVisible();
  86  |   });
  87  | 
  88  |   test('should create note via toolbar', async ({ page }) => {
  89  |     await page.click('[data-testid="toolbar-new-note"]');
  90  |     await page.waitForURL('**/notes/new', { timeout: 5000 });
  91  |     
  92  |     await page.fill('input[placeholder="Title"]', 'Toolbar Created Note');
  93  |     await page.fill('textarea', 'Content from toolbar test');
  94  |     await page.click('button:has-text("Create")');
  95  |     
  96  |     await page.waitForURL(/\/notes\/[a-f0-9-]+/, { timeout: 5000 });
  97  |     await expect(page.locator('h1')).toHaveText('Toolbar Created Note');
  98  |   });
  99  | });
  100 | 
  101 | test.describe('Graph Page Features', () => {
  102 |   test('should show note popup on node click', async ({ page, request }) => {
  103 |     // Create two notes and a link
  104 |     const note1 = await request.post('http://localhost:8080/notes', {
  105 |       data: { title: 'Graph Node A', content: 'Content A' }
  106 |     });
  107 |     const note2 = await request.post('http://localhost:8080/notes', {
  108 |       data: { title: 'Graph Node B', content: 'Content B' }
  109 |     });
  110 |     const id1 = (await note1.json()).id;
  111 |     const id2 = (await note2.json()).id;
  112 |     
  113 |     await request.post('http://localhost:8080/links', {
  114 |       data: { source_note_id: id1, target_note_id: id2, link_type: 'reference', weight: 1.0 }
  115 |     });
  116 |     
  117 |     // Navigate to graph
  118 |     await page.goto(`http://localhost:5173/graph/${id1}`);
  119 |     await page.waitForSelector('[data-testid="main-graph-canvas"]', { timeout: 10000 });
  120 |     await page.waitForTimeout(2000); // Let graph render
  121 |     
  122 |     // Canvas should be visible
  123 |     await expect(page.locator('[data-testid="main-graph-canvas"]')).toBeVisible();
  124 |   });
  125 | 
  126 |   test('should show empty state for isolated node', async ({ page, request }) => {
  127 |     // Create single note without links
  128 |     const note = await request.post('http://localhost:8080/notes', {
  129 |       data: { title: 'Isolated Node', content: 'No connections' }
  130 |     });
  131 |     const id = (await note.json()).id;
  132 |     
  133 |     // Navigate to graph
  134 |     await page.goto(`http://localhost:5173/graph/${id}`);
  135 |     await page.waitForLoadState('networkidle');
  136 |     
  137 |     // Should show empty state for isolated node
  138 |     await expect(page.locator('text=Одинокая звезда')).toBeVisible();
  139 |     await expect(page.locator('text=Создайте связи, чтобы увидеть созвездие')).toBeVisible();
  140 |   });
  141 | });
  142 | 
```