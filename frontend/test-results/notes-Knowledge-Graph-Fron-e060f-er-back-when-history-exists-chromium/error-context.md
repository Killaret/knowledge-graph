# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: notes.spec.ts >> Knowledge Graph Frontend >> should use browser back when history exists
- Location: tests\notes.spec.ts:169:3

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('.back-button')
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('.back-button')

```

# Page snapshot

```yaml
- paragraph [ref=e3]: Loading...
```

# Test source

```ts
  89  |     await page.waitForFunction(() => !window.location.pathname.includes('/notes/'), { timeout: 10000 });
  90  |     await page.waitForTimeout(1000);
  91  |     
  92  |     // Verify via API that note is deleted
  93  |     const checkResponse = await request.get(`http://localhost:8080/notes/${noteId}`);
  94  |     expect(checkResponse.status()).toBe(404);
  95  |   });
  96  | 
  97  |   test('should open 3D graph for a note with links', async ({ page, request }) => {
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
  108 |       data: { sourceNoteId: id1, targetNoteId: id2, link_type: 'reference', weight: 1.0 }
  109 |     });
  110 | 
  111 |     // Navigate to 3D graph page directly
  112 |     await page.goto(`http://localhost:5173/graph/3d/${id1}`);
  113 |     await page.waitForLoadState('networkidle');
  114 |     await page.waitForTimeout(2000);
  115 |     
  116 |     // Verify 3D graph container is visible immediately (no lazy loading)
  117 |     const graphContainer = page.locator('.graph-3d-container').first();
  118 |     await expect(graphContainer).toBeVisible({ timeout: 3000 });
  119 |     
  120 |     // Verify stats bar shows node and link counts
  121 |     const statsBar = page.locator('.stats-bar').first();
  122 |     await expect(statsBar).toBeVisible();
  123 |     
  124 |     const statsText = await statsBar.textContent();
  125 |     expect(statsText).toMatch(/\d+\s*nodes?/i);
  126 |     expect(statsText).toMatch(/\d+\s*links?/i);
  127 |   });
  128 | 
  129 |   test('should show back button on note detail page', async ({ page, request }) => {
  130 |     // Create a note via API
  131 |     const note = await request.post('http://localhost:8080/notes', {
  132 |       data: { title: 'Back Button Test', content: 'Testing back button functionality' }
  133 |     });
  134 |     const noteId = (await note.json()).id;
  135 |     
  136 |     // Navigate to note detail page
  137 |     await page.goto(`http://localhost:5173/notes/${noteId}`);
  138 |     await page.waitForTimeout(1000);
  139 | 
  140 |     // Check that back button is visible (use first())
  141 |     await expect(page.locator('.back-button').first()).toBeVisible();
  142 |     
  143 |     // Test back button functionality
  144 |     await page.click('.back-button');
  145 |     await expect(page).toHaveURL('http://localhost:5173/');
  146 |   });
  147 | 
  148 |   test('should search for notes', async ({ page, request }) => {
  149 |     // Create a note via API with searchable content
  150 |     const timestamp = Date.now();
  151 |     await request.post('http://localhost:8080/notes', {
  152 |       data: { title: 'Searchable Note ' + timestamp, content: 'Unique search content ' + timestamp, type: 'star' }
  153 |     });
  154 |     
  155 |     // Navigate to home
  156 |     await page.goto('http://localhost:5173');
  157 |     await page.waitForTimeout(1000);
  158 |     
  159 |     // Use search in floating controls
  160 |     await page.fill('.search-input', 'Unique search content');
  161 |     await page.click('.search-btn');
  162 |     
  163 |     // Verify search works via API
  164 |     const searchResponse = await request.get('http://localhost:8080/notes/search?q=Unique+search+content');
  165 |     const searchData = await searchResponse.json();
  166 |     expect(searchData.total).toBeGreaterThan(0);
  167 |   });
  168 | 
  169 |   test('should use browser back when history exists', async ({ page, request }) => {
  170 |     // Create a note via API
  171 |     const note = await request.post('http://localhost:8080/notes', {
  172 |       data: { title: 'History Test', content: 'Testing browser back functionality' }
  173 |     });
  174 |     const noteId = (await note.json()).id;
  175 |     
  176 |     // Navigate to note page
  177 |     await page.goto(`http://localhost:5173/notes/${noteId}`);
  178 |     await page.waitForTimeout(1000);
  179 |     
  180 |     // Navigate to home page
  181 |     await page.goto('http://localhost:5173');
  182 |     await page.waitForTimeout(1000);
  183 |     
  184 |     // Go back to note page
  185 |     await page.goBack();
  186 |     await page.waitForTimeout(1000);
  187 |     
  188 |     // Verify back button is visible
> 189 |     await expect(page.locator('.back-button')).toBeVisible();
      |                                                ^ Error: expect(locator).toBeVisible() failed
  190 |     
  191 |     // Click back button - should navigate using browser history
  192 |     await page.click('.back-button');
  193 |     await page.waitForTimeout(2000);
  194 |     
  195 |     // Should be back on home page
  196 |     await expect(page).toHaveURL('http://localhost:5173/');
  197 |   });
  198 | });
  199 | 
```