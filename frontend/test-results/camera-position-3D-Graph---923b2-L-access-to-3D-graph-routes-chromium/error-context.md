# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: camera-position.spec.ts >> 3D Graph - Camera Position and Navigation >> should handle direct URL access to 3D graph routes
- Location: tests\camera-position.spec.ts:262:3

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: true
Received: false
```

# Page snapshot

```yaml
- generic [ref=e3]:
  - generic [ref=e4]:
    - generic [ref=e5]:
      - strong [ref=e6]: "100"
      - text: nodes
    - generic [ref=e7]:
      - strong [ref=e8]: "2444"
      - text: links
  - generic [ref=e9]:
    - generic:
      - generic [ref=e11] [cursor=pointer]: Scene Test Note
      - generic [ref=e12] [cursor=pointer]: Galaxy Node
      - generic [ref=e13] [cursor=pointer]: Isolated Node
      - generic [ref=e14] [cursor=pointer]: Direct URL Test
      - generic [ref=e15] [cursor=pointer]: Toggle Camera Test
      - generic [ref=e16] [cursor=pointer]: Navigation Test
      - generic [ref=e17] [cursor=pointer]: Linked Node
      - generic [ref=e18] [cursor=pointer]: Center Node
      - generic [ref=e19] [cursor=pointer]: 2D to 3D Transition
      - generic [ref=e20] [cursor=pointer]: Full Graph Node 2
      - generic [ref=e21] [cursor=pointer]: Full Graph Node 1
      - generic [ref=e22] [cursor=pointer]: No Type 177642253...
      - generic [ref=e23] [cursor=pointer]: Root Planet 17764...
      - generic [ref=e24] [cursor=pointer]: Canvas Dimension ...
      - generic [ref=e25] [cursor=pointer]: Graph Comet 17764...
      - generic [ref=e26] [cursor=pointer]: Graph Planet 1776...
      - generic [ref=e27] [cursor=pointer]: Graph Star 177642...
      - generic [ref=e28] [cursor=pointer]: Clear Filter Come...
      - generic [ref=e29] [cursor=pointer]: Clear Filter Plan...
      - generic [ref=e30] [cursor=pointer]: Clear Filter Star...
      - generic [ref=e31] [cursor=pointer]: All Filter Comet ...
      - generic [ref=e32] [cursor=pointer]: All Filter Planet...
      - generic [ref=e33] [cursor=pointer]: All Filter Star 1...
      - generic [ref=e34] [cursor=pointer]: Galaxy Filter Tes...
      - generic [ref=e35] [cursor=pointer]: Galaxy Filter Tes...
      - generic [ref=e36] [cursor=pointer]: asteroid Node
      - generic [ref=e37] [cursor=pointer]: comet Node
      - generic [ref=e38] [cursor=pointer]: planet Node
      - generic [ref=e39] [cursor=pointer]: Central Star
      - generic [ref=e40] [cursor=pointer]: Isolated Node
      - generic [ref=e41] [cursor=pointer]: Comet Type 177642...
      - generic [ref=e42] [cursor=pointer]: Planet Type 17764...
      - generic [ref=e43] [cursor=pointer]: Planet Filter Tes...
      - generic [ref=e44] [cursor=pointer]: Star Filter Test ...
      - generic [ref=e45] [cursor=pointer]: Structure Link
      - generic [ref=e46] [cursor=pointer]: Structure Test
      - generic [ref=e47] [cursor=pointer]: Camera Reset Test
      - generic [ref=e48] [cursor=pointer]: Fog Link 0
      - generic [ref=e49] [cursor=pointer]: Fog Test Note
      - generic [ref=e50] [cursor=pointer]: Target Node
      - generic [ref=e51] [cursor=pointer]: Source Node
      - generic [ref=e52] [cursor=pointer]: Linked to First
      - generic [ref=e53] [cursor=pointer]: Second Graph Note
      - generic [ref=e54] [cursor=pointer]: Searchable Note 1...
      - generic [ref=e55] [cursor=pointer]: History Test
      - generic [ref=e56] [cursor=pointer]: Node Display Test
      - generic [ref=e57] [cursor=pointer]: Central Progressi...
      - generic [ref=e58] [cursor=pointer]: Playwright Test 1...
      - generic [ref=e59] [cursor=pointer]: Loading Test Link 0
      - generic [ref=e60] [cursor=pointer]: Progressive Loadi...
      - generic [ref=e61] [cursor=pointer]: Back Button Test
      - generic [ref=e62] [cursor=pointer]: Node B
      - generic [ref=e63] [cursor=pointer]: Edited 1776422498235
      - generic [ref=e64] [cursor=pointer]: List View Test Note
      - generic [ref=e65] [cursor=pointer]: Graph View Test 1...
      - generic [ref=e66] [cursor=pointer]: Side Panel Test 1...
      - generic [ref=e67] [cursor=pointer]: Planet Note 17764...
      - generic [ref=e68] [cursor=pointer]: Star Note 1776422...
      - generic [ref=e69] [cursor=pointer]: Stats Test 2 1776...
      - generic [ref=e70] [cursor=pointer]: Stats Test 1 1776...
      - generic [ref=e71] [cursor=pointer]: Debris Node
      - generic [ref=e72] [cursor=pointer]: 3D Stats Test
      - generic [ref=e73] [cursor=pointer]: Stats Test Node 1
      - generic [ref=e74] [cursor=pointer]: Asteroid Node
      - generic [ref=e75] [cursor=pointer]: Isolated Note No ...
      - generic [ref=e76] [cursor=pointer]: Comet Node
      - generic [ref=e77] [cursor=pointer]: Navigation Test Note
      - generic [ref=e78] [cursor=pointer]: Styling Test Note
      - generic [ref=e79] [cursor=pointer]: Note 4
      - generic [ref=e80] [cursor=pointer]: Note 3
      - generic [ref=e81] [cursor=pointer]: Note 2
      - generic [ref=e82] [cursor=pointer]: Note 1
      - generic [ref=e83] [cursor=pointer]: Note 0
      - generic [ref=e84] [cursor=pointer]: Planet Node
      - generic [ref=e85] [cursor=pointer]: Full Graph Note 2
      - generic [ref=e86] [cursor=pointer]: Full Graph Note 0
      - generic [ref=e87] [cursor=pointer]: 3D Graph Test Note
      - generic [ref=e88] [cursor=pointer]: Linked Note
      - generic [ref=e89] [cursor=pointer]: Main Note
      - generic [ref=e90] [cursor=pointer]: Weak Link
      - generic [ref=e91] [cursor=pointer]: Strong Link
      - generic [ref=e92] [cursor=pointer]: Related Target
      - generic [ref=e93] [cursor=pointer]: Source
      - generic [ref=e94] [cursor=pointer]: 2D Toggle Test 2
      - generic [ref=e95] [cursor=pointer]: Dependency Target
      - generic [ref=e96] [cursor=pointer]: 2D Toggle Test 1
      - generic [ref=e97] [cursor=pointer]: Full Graph Node 4
      - generic [ref=e98] [cursor=pointer]: Full Graph Node 3
      - generic [ref=e99] [cursor=pointer]: Full Graph Node 0
      - generic [ref=e100] [cursor=pointer]: Metadata Star 177...
      - generic [ref=e101] [cursor=pointer]: Comet Filter Test...
      - generic [ref=e102] [cursor=pointer]: Linked to Second
      - generic [ref=e103] [cursor=pointer]: First Graph Note
      - generic [ref=e104] [cursor=pointer]: Linked Note 0
      - generic [ref=e105] [cursor=pointer]: Node A
      - generic [ref=e106] [cursor=pointer]: Test Searchable17...
      - generic [ref=e107] [cursor=pointer]: Home Page Test No...
      - generic [ref=e108] [cursor=pointer]: Stats Test Node 2
      - generic [ref=e109] [cursor=pointer]: Toggle Test
      - generic [ref=e110] [cursor=pointer]: Full Graph Note 1
```

# Test source

```ts
  199 |     
  200 |     // 3D graph should reload correctly
  201 |     const container3DAgain = page.locator('.graph-3d-container').first();
  202 |     await expect(container3DAgain).toBeVisible();
  203 |   });
  204 | 
  205 |   test('should transition from 2D to 3D graph maintaining context', async ({ page, request }) => {
  206 |     const note = await createNote(request, {
  207 |       title: '2D to 3D Transition',
  208 |       content: 'Testing 2D to 3D',
  209 |       type: 'planet'
  210 |     });
  211 |     const noteId = note.id;
  212 | 
  213 |     // Start at 2D graph page
  214 |     await page.goto('/graph');
  215 |     await page.waitForLoadState('networkidle');
  216 |     await page.waitForTimeout(3000);
  217 | 
  218 |     // Verify 2D graph loads
  219 |     const graph2D = page.locator('.fullscreen-graph, canvas').first();
  220 |     await expect(graph2D).toBeVisible();
  221 | 
  222 |     // Navigate to 3D version
  223 |     await page.goto(`/graph/3d/${noteId}`);
  224 |     await page.waitForLoadState('networkidle');
  225 |     await page.waitForTimeout(4000);
  226 |     
  227 |     // Verify 3D graph loads
  228 |     const graph3D = page.locator('.graph-3d-container').first();
  229 |     await expect(graph3D).toBeVisible();
  230 |     
  231 |     // Stats should show the specific note's local graph
  232 |     const statsBar = page.locator('.stats-bar').first();
  233 |     if (await statsBar.isVisible().catch(() => false)) {
  234 |       const statsText = await statsBar.textContent();
  235 |       expect(statsText).toMatch(/\d+\s*nodes/);
  236 |       expect(statsText).toMatch(/\d+\s*links/);
  237 |     }
  238 |   });
  239 | 
  240 |   test('should navigate from home page 3D button to full 3D graph', async ({ page }) => {
  241 |     // Start at home page
  242 |     await page.goto('/');
  243 |     await page.waitForLoadState('networkidle');
  244 |     await page.waitForTimeout(2000);
  245 |     
  246 |     // Find and click the 3D button
  247 |     const button3D = page.locator('.view-button-3d, button:has-text("3D")').first();
  248 |     if (await button3D.isVisible().catch(() => false)) {
  249 |       await button3D.click();
  250 |       await page.waitForTimeout(3000);
  251 |       
  252 |       // Should navigate to /graph/3d
  253 |       const currentUrl = page.url();
  254 |       expect(currentUrl).toContain('/graph/3d');
  255 |       
  256 |       // 3D graph should load
  257 |       const container = page.locator('.graph-3d-container, .lazy-loading').first();
  258 |       await expect(container).toBeVisible({ timeout: 10000 });
  259 |     }
  260 |   });
  261 | 
  262 |   test('should handle direct URL access to 3D graph routes', async ({ page, request }) => {
  263 |     const note = await createNote(request, {
  264 |       title: 'Direct URL Test',
  265 |       content: 'Testing direct URL access',
  266 |       type: 'comet'
  267 |     });
  268 |     const noteId = note.id;
  269 | 
  270 |     // Test direct access to local 3D graph
  271 |     await page.goto(`/graph/3d/${noteId}`);
  272 |     await page.waitForLoadState('networkidle');
  273 |     await page.waitForTimeout(4000);
  274 | 
  275 |     const localContainer = page.locator('.graph-3d-container').first();
  276 |     await expect(localContainer).toBeVisible();
  277 | 
  278 |     // Local mode should be indicated
  279 |     const statsLocal = page.locator('.stats-bar').first();
  280 |     if (await statsLocal.isVisible().catch(() => false)) {
  281 |       const statsText = await statsLocal.textContent();
  282 |       const hasLocalMode = statsText?.toLowerCase().includes('local');
  283 |       expect(hasLocalMode).toBe(true);
  284 |     }
  285 | 
  286 |     // Test direct access to full 3D graph
  287 |     await page.goto('/graph/3d');
  288 |     await page.waitForLoadState('networkidle');
  289 |     await page.waitForTimeout(4000);
  290 |     
  291 |     const fullContainer = page.locator('.graph-3d-container, .lazy-loading').first();
  292 |     await expect(fullContainer).toBeVisible();
  293 |     
  294 |     // Full mode should be indicated
  295 |     const statsFull = page.locator('.stats-bar').first();
  296 |     if (await statsFull.isVisible().catch(() => false)) {
  297 |       const statsText = await statsFull.textContent();
  298 |       const hasFullMode = statsText?.toLowerCase().includes('full');
> 299 |       expect(hasFullMode).toBe(true);
      |                           ^ Error: expect(received).toBe(expected) // Object.is equality
  300 |     }
  301 |   });
  302 | 
  303 |   test('should show empty state with appropriate camera position when no notes', async ({ page }) => {
  304 |     // Navigate to full 3D graph when no notes exist (or after clearing)
  305 |     await page.goto('/graph/3d');
  306 |     await page.waitForLoadState('networkidle');
  307 |     await page.waitForTimeout(3000);
  308 |     
  309 |     // Should show empty state or loading state, not error
  310 |     const emptyState = page.locator('.no-data-message, .empty-state, .lazy-loading').first();
  311 |     const graphContainer = page.locator('.graph-3d-container').first();
  312 |     
  313 |     const hasEmpty = await emptyState.isVisible().catch(() => false);
  314 |     const hasGraph = await graphContainer.isVisible().catch(() => false);
  315 |     
  316 |     expect(hasEmpty || hasGraph).toBe(true);
  317 |   });
  318 | 
  319 |   test('should position camera correctly for isolated single node', async ({ page, request }) => {
  320 |     // Create note with NO connections
  321 |     const isolatedNote = await createNote(request, {
  322 |       title: 'Isolated Node',
  323 |       content: 'No connections',
  324 |       type: 'galaxy'
  325 |     });
  326 |     const noteId = isolatedNote.id;
  327 | 
  328 |     await page.goto(`/graph/3d/${noteId}`);
  329 |     await page.waitForLoadState('networkidle');
  330 |     await page.waitForTimeout(4000);
  331 |     
  332 |     // Should render without errors
  333 |     const container = page.locator('.graph-3d-container, .no-data-message').first();
  334 |     await expect(container).toBeVisible();
  335 |     
  336 |     // Stats should show 1 node (just the isolated note itself)
  337 |     const statsBar = page.locator('.stats-bar').first();
  338 |     if (await statsBar.isVisible().catch(() => false)) {
  339 |       const statsText = await statsBar.textContent();
  340 |       const nodeMatch = statsText?.match(/(\d+)\s*nodes/);
  341 |       if (nodeMatch) {
  342 |         const nodeCount = parseInt(nodeMatch[1], 10);
  343 |         expect(nodeCount).toBeGreaterThanOrEqual(1);
  344 |       }
  345 |     }
  346 |   });
  347 | });
  348 | // End of Camera Position and Navigation test suite
  349 | 
```