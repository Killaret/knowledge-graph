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
      - strong [ref=e8]: "2911"
      - text: links
  - generic [ref=e9]:
    - button "Reset View" [ref=e11] [cursor=pointer]:
      - img [ref=e12]
      - generic [ref=e14]: Reset View
    - generic:
      - generic [ref=e16] [cursor=pointer]: Fog Link 2
      - generic [ref=e17] [cursor=pointer]: Fog Link 1
      - generic [ref=e18] [cursor=pointer]: Fog Link 0
      - generic [ref=e19] [cursor=pointer]: Fog Test Note
      - generic [ref=e20] [cursor=pointer]: Target Node
      - generic [ref=e21] [cursor=pointer]: Structure Test
      - generic [ref=e22] [cursor=pointer]: Linked to Second
      - generic [ref=e23] [cursor=pointer]: Linked to First
      - generic [ref=e24] [cursor=pointer]: Second Graph Note
      - generic [ref=e25] [cursor=pointer]: First Graph Note
      - generic [ref=e26] [cursor=pointer]: Loading Test Link 7
      - generic [ref=e27] [cursor=pointer]: Loading Test Link 6
      - generic [ref=e28] [cursor=pointer]: Loading Test Link 5
      - generic [ref=e29] [cursor=pointer]: Loading Test Link 4
      - generic [ref=e30] [cursor=pointer]: Loading Test Link 3
      - generic [ref=e31] [cursor=pointer]: Loading Test Link 2
      - generic [ref=e32] [cursor=pointer]: Loading Test Link 0
      - generic [ref=e33] [cursor=pointer]: Progressive Loadi...
      - generic [ref=e34] [cursor=pointer]: Linked Note 4
      - generic [ref=e35] [cursor=pointer]: Linked Note 3
      - generic [ref=e36] [cursor=pointer]: Secondary 2
      - generic [ref=e37] [cursor=pointer]: Linked Note 2
      - generic [ref=e38] [cursor=pointer]: Node Display Test
      - generic [ref=e39] [cursor=pointer]: 3D Stats Test
      - generic [ref=e40] [cursor=pointer]: Secondary 1
      - generic [ref=e41] [cursor=pointer]: Secondary 0
      - generic [ref=e42] [cursor=pointer]: Linked Note 0
      - generic [ref=e43] [cursor=pointer]: Central Progressi...
      - generic [ref=e44] [cursor=pointer]: Stats Test Node 2
      - generic [ref=e45] [cursor=pointer]: Stats Test Node 1
      - generic [ref=e46] [cursor=pointer]: Toggle Test
      - generic [ref=e47] [cursor=pointer]: Comet Node
      - generic [ref=e48] [cursor=pointer]: Isolated Note No ...
      - generic [ref=e49] [cursor=pointer]: Asteroid Node
      - generic [ref=e50] [cursor=pointer]: Note 4
      - generic [ref=e51] [cursor=pointer]: Linked Note
      - generic [ref=e52] [cursor=pointer]: Main Note
      - generic [ref=e53] [cursor=pointer]: Source
      - generic [ref=e54] [cursor=pointer]: Weak Link
      - generic [ref=e55] [cursor=pointer]: Strong Link
      - generic [ref=e56] [cursor=pointer]: 3D Toggle Test 2
      - generic [ref=e57] [cursor=pointer]: 3D Toggle Test 1
      - generic [ref=e58] [cursor=pointer]: Labeled Node
      - generic [ref=e59] [cursor=pointer]: Scene Test Note
      - generic [ref=e60] [cursor=pointer]: Isolated Node
      - generic [ref=e61] [cursor=pointer]: Direct URL Test
      - generic [ref=e62] [cursor=pointer]: Full Graph Node 4
      - generic [ref=e63] [cursor=pointer]: Toggle Camera Test
      - generic [ref=e64] [cursor=pointer]: Direct URL Test
      - generic [ref=e65] [cursor=pointer]: Canvas Dimension ...
      - generic [ref=e66] [cursor=pointer]: Isolated Node
      - generic [ref=e67] [cursor=pointer]: Camera Reset Test
      - generic [ref=e68] [cursor=pointer]: Target Node
      - generic [ref=e69] [cursor=pointer]: planet Node
      - generic [ref=e70] [cursor=pointer]: Linked Note 1
      - generic [ref=e71] [cursor=pointer]: Navigation Test Note
      - generic [ref=e72] [cursor=pointer]: Loading Test Link 1
```

# Test source

```ts
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
  278 |     // Local mode should be indicated - wait for stats to appear
  279 |     const statsLocal = page.locator('[data-testid="graph-stats"], .stats-bar').first();
  280 |     await expect(statsLocal).toBeVisible({ timeout: 5000 });
  281 |     const statsText = await statsLocal.textContent();
  282 |     const hasLocalMode = statsText?.toLowerCase().includes('local') || statsText?.toLowerCase().includes('view');
  283 |     expect(hasLocalMode).toBe(true);
  284 | 
  285 |     // Test direct access to full 3D graph
  286 |     await page.goto('/graph/3d');
  287 |     await page.waitForLoadState('networkidle');
  288 |     await page.waitForTimeout(4000);
  289 |     
  290 |     const fullContainer = page.locator('.graph-3d-container, .lazy-loading').first();
  291 |     await expect(fullContainer).toBeVisible();
  292 |     
  293 |     // Enable full graph mode by clicking toggle (default is local view)
  294 |     const toggle = page.locator('.toggle input[type="checkbox"]').first();
  295 |     const hasToggle = await toggle.isVisible().catch(() => false);
  296 |     if (hasToggle) {
  297 |       await toggle.click();
  298 |       await page.waitForTimeout(3000);
  299 |     }
  300 |     
  301 |     // Full mode should be indicated - wait for stats to appear
  302 |     const statsFull = page.locator('[data-testid="graph-stats"], .stats-bar').first();
  303 |     await expect(statsFull).toBeVisible({ timeout: 5000 });
  304 |     const statsTextFull = await statsFull.textContent();
  305 |     const hasFullMode = statsTextFull?.toLowerCase().includes('full') || statsTextFull?.toLowerCase().includes('all');
> 306 |     expect(hasFullMode).toBe(true);
      |                         ^ Error: expect(received).toBe(expected) // Object.is equality
  307 |   });
  308 | 
  309 |   test('should show empty state with appropriate camera position when no notes', async ({ page }) => {
  310 |     // Navigate to full 3D graph when no notes exist (or after clearing)
  311 |     await page.goto('/graph/3d');
  312 |     await page.waitForLoadState('networkidle');
  313 |     await page.waitForTimeout(3000);
  314 |     
  315 |     // Should show empty state or loading state, not error
  316 |     const emptyState = page.locator('.no-data-message, .empty-state, .lazy-loading').first();
  317 |     const graphContainer = page.locator('.graph-3d-container').first();
  318 |     
  319 |     const hasEmpty = await emptyState.isVisible().catch(() => false);
  320 |     const hasGraph = await graphContainer.isVisible().catch(() => false);
  321 |     
  322 |     expect(hasEmpty || hasGraph).toBe(true);
  323 |   });
  324 | 
  325 |   test('should position camera correctly for isolated single node', async ({ page, request }) => {
  326 |     // Create note with NO connections
  327 |     const isolatedNote = await createNote(request, {
  328 |       title: 'Isolated Node',
  329 |       content: 'No connections',
  330 |       type: 'galaxy'
  331 |     });
  332 |     const noteId = isolatedNote.id;
  333 | 
  334 |     await page.goto(`/graph/3d/${noteId}`);
  335 |     await page.waitForLoadState('networkidle');
  336 |     await page.waitForTimeout(4000);
  337 |     
  338 |     // Should render without errors
  339 |     const container = page.locator('.graph-3d-container, .no-data-message').first();
  340 |     await expect(container).toBeVisible();
  341 |     
  342 |     // Stats should show 1 node (just the isolated note itself)
  343 |     const statsBar = page.locator('.stats-bar').first();
  344 |     if (await statsBar.isVisible().catch(() => false)) {
  345 |       const statsText = await statsBar.textContent();
  346 |       const nodeMatch = statsText?.match(/(\d+)\s*nodes/);
  347 |       if (nodeMatch) {
  348 |         const nodeCount = parseInt(nodeMatch[1], 10);
  349 |         expect(nodeCount).toBeGreaterThanOrEqual(1);
  350 |       }
  351 |     }
  352 |   });
  353 | });
  354 | // End of Camera Position and Navigation test suite
  355 | 
```