# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: notes.spec.ts >> Knowledge Graph Frontend >> should delete a note
- Location: tests\notes.spec.ts:38:3

# Error details

```
TimeoutError: page.waitForSelector: Timeout 5000ms exceeded.
Call log:
  - waiting for locator('.note-card') to be visible

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
        - heading "Все звёзды" [level=1] [ref=e17]
        - generic [ref=e18]:
          - button "Выбрать заметки" [ref=e19] [cursor=pointer]:
            - img [ref=e20]
            - text: Выбрать
          - button "Новая звезда" [ref=e22] [cursor=pointer]:
            - img [ref=e23]
            - text: Новая звезда
      - generic [ref=e24]:
        - textbox "Поиск по названию или содержимому..." [ref=e26]
        - generic [ref=e27]:
          - combobox [ref=e28]:
            - option "Все типы" [selected]
            - option "⭐ star"
            - option "🪐 planet"
            - option "☄️ comet"
            - option "🌌 galaxy"
          - combobox [ref=e29]:
            - option "Все теги" [selected]
          - textbox "От" [ref=e30]
          - textbox "До" [ref=e31]
          - combobox [ref=e32]:
            - option "По дате" [selected]
            - option "По алфавиту"
            - option "По связям"
          - button "↓" [ref=e33] [cursor=pointer]
          - combobox [ref=e34]:
            - option "Без группировки" [selected]
            - option "По типу"
            - option "По дате"
      - generic [ref=e35]: Показано 0 из 0 заметок
      - generic [ref=e36]:
        - generic [ref=e37]:
          - status [ref=e38]: Loading...
          - status [ref=e39]: Loading...
        - generic [ref=e40]:
          - status [ref=e41]: Loading...
          - status [ref=e42]: Loading...
        - generic [ref=e43]:
          - status [ref=e44]: Loading...
          - status [ref=e45]: Loading...
        - generic [ref=e46]:
          - status [ref=e47]: Loading...
          - status [ref=e48]: Loading...
        - generic [ref=e49]:
          - status [ref=e50]: Loading...
          - status [ref=e51]: Loading...
        - generic [ref=e52]:
          - status [ref=e53]: Loading...
          - status [ref=e54]: Loading...
  - navigation "Main navigation":
    - button "Новая звезда" [ref=e55] [cursor=pointer]:
      - img [ref=e56]
    - button "Граф" [ref=e57] [cursor=pointer]:
      - img [ref=e58]
    - button "Все заметки" [ref=e64] [cursor=pointer]:
      - img [ref=e65]
    - button "Импорт" [ref=e66] [cursor=pointer]:
      - img [ref=e67]
    - button "Экспорт" [ref=e70] [cursor=pointer]:
      - img [ref=e71]
    - button "Настройки" [ref=e74] [cursor=pointer]:
      - img [ref=e75]
  - generic [ref=e78]:
    - button "Обзор" [ref=e79] [cursor=pointer]:
      - generic [ref=e80]: Обзор
      - img [ref=e81]
    - generic [ref=e83]:
      - generic [ref=e84]:
        - generic [ref=e85]:
          - generic [ref=e86]: "0"
          - generic [ref=e87]: узлов
        - generic [ref=e89]:
          - generic [ref=e90]: "0"
          - generic [ref=e91]: связей
      - generic [ref=e94]:
        - img [ref=e95]
        - generic [ref=e101]: Мини-карта
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
> 49  |     await page.waitForSelector('.note-card', { timeout: 5000 });
      |                ^ TimeoutError: page.waitForSelector: Timeout 5000ms exceeded.
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
  79  |     await expect(page.locator('[data-testid="main-graph-canvas"]')).toBeVisible();
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
```