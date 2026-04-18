# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: notes.spec.ts >> Knowledge Graph Frontend >> should use browser back when history exists
- Location: tests\notes.spec.ts:171:3

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

# Test source

```ts
  92  | 
  93  |     // Wait for navigation away from note page (either redirect or URL change)
  94  |     await page.waitForFunction(() => !window.location.pathname.includes('/notes/'), { timeout: 10000 });
  95  |     await page.waitForTimeout(1000);
  96  |     
  97  |     // Verify via API that note is deleted
  98  |     const checkResponse = await request.get(`${getBackendUrl()}/notes/${noteId}`);
  99  |     expect(checkResponse.status()).toBe(404);
  100 |   });
  101 | 
  102 |   test('should open 3D graph for a note with links', async ({ page, request }) => {
  103 |     // Create two notes and a link via API using helper
  104 |     const note1 = await createNote(request, { title: 'Node A', content: 'A' });
  105 |     const note2 = await createNote(request, { title: 'Node B', content: 'B' });
  106 |     const id1 = note1.id;
  107 |     const id2 = note2.id;
  108 |     await createLink(request, id1, id2, 1.0, 'reference');
  109 | 
  110 |     // Navigate to 3D graph page directly
  111 |     await page.goto(`/graph/3d/${id1}`);
  112 |     await page.waitForLoadState('networkidle');
  113 |     await page.waitForTimeout(2000);
  114 |     
  115 |     // Verify 3D graph container is visible immediately (no lazy loading)
  116 |     const graphContainer = page.locator('.graph-3d-container').first();
  117 |     await expect(graphContainer).toBeVisible({ timeout: 3000 });
  118 |     
  119 |     // Verify stats bar shows node and link counts
  120 |     const statsBar = page.locator('.stats-bar').first();
  121 |     await expect(statsBar).toBeVisible();
  122 |     
  123 |     const statsText = await statsBar.textContent();
  124 |     expect(statsText).toMatch(/\d+\s*nodes?/i);
  125 |     expect(statsText).toMatch(/\d+\s*links?/i);
  126 |   });
  127 | 
  128 |   test('should show back button on note detail page', async ({ page, request }) => {
  129 |     // Create a note via API using helper
  130 |     const note = await createNote(request, {
  131 |       title: 'Back Button Test',
  132 |       content: 'Testing back button functionality'
  133 |     });
  134 |     const noteId = note.id;
  135 | 
  136 |     // Navigate to note detail page
  137 |     await page.goto(`/notes/${noteId}`);
  138 |     await page.waitForTimeout(1000);
  139 | 
  140 |     // Check that back button is visible (use first())
  141 |     await expect(page.locator('.back-button').first()).toBeVisible();
  142 |     
  143 |     // Test back button functionality
  144 |     await page.click('.back-button');
  145 |     await expect(page).toHaveURL('/');
  146 |   });
  147 | 
  148 |   test('should search for notes', async ({ page, request }) => {
  149 |     // Create a note via API with searchable content using helper
  150 |     const timestamp = Date.now();
  151 |     await createNote(request, {
  152 |       title: 'Searchable Note ' + timestamp,
  153 |       content: 'Unique search content ' + timestamp,
  154 |       type: 'star'
  155 |     });
  156 | 
  157 |     // Navigate to home
  158 |     await page.goto('/');
  159 |     await page.waitForTimeout(1000);
  160 | 
  161 |     // Use search in floating controls
  162 |     await page.fill('.search-input', 'Unique search content');
  163 |     await page.click('.search-btn');
  164 | 
  165 |     // Verify search works via API
  166 |     const searchResponse = await request.get(`${getBackendUrl()}/notes/search?q=Unique+search+content`);
  167 |     const searchData = await searchResponse.json();
  168 |     expect(searchData.total).toBeGreaterThan(0);
  169 |   });
  170 | 
  171 |   test('should use browser back when history exists', async ({ page, request }) => {
  172 |     // Create a note via API using helper
  173 |     const note = await createNote(request, {
  174 |       title: 'History Test',
  175 |       content: 'Testing browser back functionality'
  176 |     });
  177 |     const noteId = note.id;
  178 | 
  179 |     // Navigate to note page
  180 |     await page.goto(`/notes/${noteId}`);
  181 |     await page.waitForTimeout(1000);
  182 | 
  183 |     // Navigate to home page
  184 |     await page.goto('/');
  185 |     await page.waitForTimeout(1000);
  186 | 
  187 |     // Go back to note page
  188 |     await page.goBack();
  189 |     await page.waitForTimeout(1000);
  190 | 
  191 |     // Verify back button is visible
> 192 |     await expect(page.locator('.back-button')).toBeVisible();
      |                                                ^ Error: expect(locator).toBeVisible() failed
  193 | 
  194 |     // Click back button - should navigate using browser history
  195 |     await page.click('.back-button');
  196 |     await page.waitForTimeout(2000);
  197 | 
  198 |     // Should be back on home page
  199 |     await expect(page).toHaveURL('/');
  200 |   });
  201 | });
  202 | 
```