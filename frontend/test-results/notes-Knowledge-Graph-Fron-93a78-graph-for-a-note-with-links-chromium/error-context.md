# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: notes.spec.ts >> Knowledge Graph Frontend >> should open graph for a note with links
- Location: tests\notes.spec.ts:97:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: true
Received: false
```

# Page snapshot

```yaml
- generic [ref=e3]:
  - banner [ref=e4]:
    - button "Go back" [ref=e5] [cursor=pointer]: « Back
    - heading "Knowledge Constellation" [level=1] [ref=e6]
    - generic [ref=e7]: Drag to rotate/pan • Scroll to zoom • Click node to open
  - paragraph [ref=e11]: Loading visualization...
```

# Test source

```ts
  23  |     await page.fill('input[name="title"]', 'Playwright Test ' + Date.now());
  24  |     await page.fill('textarea[name="content"]', 'Automated content');
  25  |     
  26  |     // Click Save button
  27  |     await page.click('button[type="submit"]');
  28  |     
  29  |     // Wait for modal to close
  30  |     await page.waitForTimeout(2000);
  31  |     
  32  |     // Verify via API that note was created
  33  |     const notesResponse = await request.get('http://localhost:8080/notes');
  34  |     const notesData = await notesResponse.json();
  35  |     expect(notesData.total).toBeGreaterThan(0);
  36  |     
  37  |     // Note: Due to API serialization issue, we verify creation via API only
  38  |     // The UI list may not refresh correctly until backend is fixed
  39  |   });
  40  | 
  41  |   test('should edit a note', async ({ page, request }) => {
  42  |     // Create a note via API first
  43  |     const timestamp = Date.now();
  44  |     const note = await request.post('http://localhost:8080/notes', {
  45  |       data: { title: 'Edit Test ' + timestamp, content: 'Original content', type: 'star' }
  46  |     });
  47  |     const noteId = (await note.json()).id;
  48  |     
  49  |     // Navigate to edit page directly
  50  |     await page.goto(`http://localhost:5173/notes/${noteId}/edit`);
  51  |     await page.waitForTimeout(1000);
  52  |     
  53  |     // Update note
  54  |     await page.waitForSelector('input[name="title"]', { timeout: 5000 });
  55  |     await page.fill('input[name="title"]', 'Edited ' + timestamp);
  56  |     await page.fill('textarea[name="content"]', 'Updated content');
  57  |     await page.click('button[type="submit"]');
  58  |     
  59  |     // Wait for redirect to note page
  60  |     await page.waitForURL(`http://localhost:5173/notes/${noteId}`, { timeout: 5000 });
  61  |     
  62  |     // Verify via API
  63  |     const updatedNote = await request.get(`http://localhost:8080/notes/${noteId}`);
  64  |     const noteData = await updatedNote.json();
  65  |     expect(noteData.title).toBe('Edited ' + timestamp);
  66  |   });
  67  | 
  68  |   test('should delete a note', async ({ page, request }) => {
  69  |     // Create a note via API first
  70  |     const timestamp = Date.now();
  71  |     const note = await request.post('http://localhost:8080/notes', {
  72  |       data: { title: 'Delete Test ' + timestamp, content: 'Test content for deletion' }
  73  |     });
  74  |     const noteId = (await note.json()).id;
  75  |     
  76  |     // Navigate directly to note page
  77  |     await page.goto(`http://localhost:5173/notes/${noteId}`);
  78  |     await page.waitForTimeout(1000);
  79  |     
  80  |     // Setup dialog handler before click
  81  |     page.on('dialog', async dialog => {
  82  |       await dialog.accept();
  83  |     });
  84  |     
  85  |     // Click Delete button
  86  |     await page.click('button:has-text("Delete")');
  87  | 
  88  |     // Wait for navigation away from note page (either redirect or URL change)
  89  |     await page.waitForFunction(() => !window.location.pathname.includes('/notes/'), { timeout: 10000 });
  90  |     await page.waitForTimeout(1000);
  91  |     
  92  |     // Verify via API that note is deleted
  93  |     const checkResponse = await request.get(`http://localhost:8080/notes/${noteId}`);
  94  |     expect(checkResponse.status()).toBe(404);
  95  |   });
  96  | 
  97  |   test('should open graph for a note with links', async ({ page, request }) => {
  98  |     // Create two notes and a link via API
  99  |     const note1 = await request.post('http://localhost:8080/notes', {
  100 |       data: { title: 'Node A', content: 'A' }
  101 |     });
  102 |     const note2 = await request.post('http://localhost:8080/notes', {
  103 |       data: { title: 'Node B', content: 'B' }
  104 |     });
  105 |     const id1 = (await note1.json()).id;
  106 |     const id2 = (await note2.json()).id;
  107 |     await request.post('http://localhost:8080/links', {
  108 |       data: { source_note_id: id1, target_note_id: id2, link_type: 'reference', weight: 1.0 }
  109 |     });
  110 | 
  111 |     // Navigate to graph page
  112 |     await page.goto(`http://localhost:5173/graph/${id1}`);
  113 |     await page.waitForLoadState('networkidle');
  114 |     await page.waitForTimeout(3000);
  115 |     
  116 |     // Verify graph visualization is present (canvas or error container)
  117 |     const canvas = page.locator('canvas, .graph-canvas').first();
  118 |     const container = page.locator('.graph-3d-container, .graph-2d, .graph-wrapper, .error-overlay').first();
  119 |     
  120 |     const hasCanvas = await canvas.isVisible().catch(() => false);
  121 |     const hasContainer = await container.isVisible().catch(() => false);
  122 |     
> 123 |     expect(hasCanvas || hasContainer).toBe(true);
      |                                       ^ Error: expect(received).toBe(expected) // Object.is equality
  124 |   });
  125 | 
  126 |   test('should show back button on note detail page', async ({ page, request }) => {
  127 |     // Create a note via API
  128 |     const note = await request.post('http://localhost:8080/notes', {
  129 |       data: { title: 'Back Button Test', content: 'Testing back button functionality' }
  130 |     });
  131 |     const noteId = (await note.json()).id;
  132 |     
  133 |     // Navigate to note detail page
  134 |     await page.goto(`http://localhost:5173/notes/${noteId}`);
  135 |     await page.waitForTimeout(1000);
  136 | 
  137 |     // Check that back button is visible (use first())
  138 |     await expect(page.locator('.back-button').first()).toBeVisible();
  139 |     
  140 |     // Test back button functionality
  141 |     await page.click('.back-button');
  142 |     await expect(page).toHaveURL('http://localhost:5173/');
  143 |   });
  144 | 
  145 |   test('should search for notes', async ({ page, request }) => {
  146 |     // Create a note via API with searchable content
  147 |     const timestamp = Date.now();
  148 |     await request.post('http://localhost:8080/notes', {
  149 |       data: { title: 'Searchable Note ' + timestamp, content: 'Unique search content ' + timestamp, type: 'star' }
  150 |     });
  151 |     
  152 |     // Navigate to home
  153 |     await page.goto('http://localhost:5173');
  154 |     await page.waitForTimeout(1000);
  155 |     
  156 |     // Use search in floating controls
  157 |     await page.fill('.search-input', 'Unique search content');
  158 |     await page.click('.search-btn');
  159 |     
  160 |     // Verify search works via API
  161 |     const searchResponse = await request.get('http://localhost:8080/notes/search?q=Unique+search+content');
  162 |     const searchData = await searchResponse.json();
  163 |     expect(searchData.total).toBeGreaterThan(0);
  164 |   });
  165 | 
  166 |   test('should use browser back when history exists', async ({ page, request }) => {
  167 |     // Create a note via API
  168 |     const note = await request.post('http://localhost:8080/notes', {
  169 |       data: { title: 'History Test', content: 'Testing browser back functionality' }
  170 |     });
  171 |     const noteId = (await note.json()).id;
  172 |     
  173 |     // Navigate to note page
  174 |     await page.goto(`http://localhost:5173/notes/${noteId}`);
  175 |     await page.waitForTimeout(1000);
  176 |     
  177 |     // Navigate to home page
  178 |     await page.goto('http://localhost:5173');
  179 |     await page.waitForTimeout(1000);
  180 |     
  181 |     // Go back to note page
  182 |     await page.goBack();
  183 |     await page.waitForTimeout(1000);
  184 |     
  185 |     // Verify back button is visible
  186 |     await expect(page.locator('.back-button')).toBeVisible();
  187 |     
  188 |     // Click back button - should navigate using browser history
  189 |     await page.click('.back-button');
  190 |     await page.waitForTimeout(2000);
  191 |     
  192 |     // Should be back on home page
  193 |     await expect(page).toHaveURL('http://localhost:5173/');
  194 |   });
  195 | });
  196 | 
```