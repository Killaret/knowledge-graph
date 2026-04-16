# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: camera-position.spec.ts >> 3D Graph - Camera Position and Navigation >> should handle direct URL access to 3D graph routes
- Location: tests\camera-position.spec.ts:221:3

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
      - strong [ref=e8]: "1837"
      - text: links
  - generic [ref=e9]:
    - paragraph [ref=e12]: Loading 3D constellation...
    - generic:
      - generic [ref=e14] [cursor=pointer]: Scene Test Note
      - generic [ref=e15] [cursor=pointer]: Isolated Node
      - generic [ref=e16] [cursor=pointer]: Direct URL Test
      - generic [ref=e17] [cursor=pointer]: Full Graph Node 4
      - generic [ref=e18] [cursor=pointer]: Full Graph Node 2
      - generic [ref=e19] [cursor=pointer]: Linked Node
      - generic [ref=e20] [cursor=pointer]: 2D to 3D Transition
      - generic [ref=e21] [cursor=pointer]: Toggle Camera Test
      - generic [ref=e22] [cursor=pointer]: Center Node
      - generic [ref=e23] [cursor=pointer]: Автономный и заме...
      - generic [ref=e24] [cursor=pointer]: Настраиваемое и м...
      - generic [ref=e25] [cursor=pointer]: Функциональная и ...
      - generic [ref=e26] [cursor=pointer]: Настраиваемая и п...
      - generic [ref=e27] [cursor=pointer]: Безопасная и глоб...
      - generic [ref=e28] [cursor=pointer]: Выражение no Посв...
      - generic [ref=e29] [cursor=pointer]: Призыв no Направо
      - generic [ref=e30] [cursor=pointer]: Металл no Налоговый
      - generic [ref=e31] [cursor=pointer]: Направо no Освобо...
      - generic [ref=e32] [cursor=pointer]: Точно no Палата
      - generic [ref=e33] [cursor=pointer]: Магазин Потянутьс...
      - generic [ref=e34] [cursor=pointer]: Угроза Мусор Kai
      - generic [ref=e35] [cursor=pointer]: Упорно Мелькнуть ...
      - generic [ref=e36] [cursor=pointer]: Интеллектуальный ...
      - generic [ref=e37] [cursor=pointer]: Экзамен Сравнение!!
      - generic [ref=e38] [cursor=pointer]: Механический Бесп...
      - generic [ref=e39] [cursor=pointer]: Интеграция Лучших...
      - generic [ref=e40] [cursor=pointer]: Максимизация Ульт...
      - generic [ref=e41] [cursor=pointer]: Оптимизация Прибы...
      - generic [ref=e42] [cursor=pointer]: Революция Круглог...
      - generic [ref=e43] [cursor=pointer]: Увеличение B2B Сх...
      - generic [ref=e44] [cursor=pointer]: Оптимизация Безот...
      - generic [ref=e45] [cursor=pointer]: Сравнение Масштаб...
      - generic [ref=e46] [cursor=pointer]: Культивация Богат...
      - generic [ref=e47] [cursor=pointer]: Разнообразная и ц...
      - generic [ref=e48] [cursor=pointer]: Оптимизированное ...
      - generic [ref=e49] [cursor=pointer]: Canvas Dimension ...
      - generic [ref=e50] [cursor=pointer]: Isolated Node
      - generic [ref=e51] [cursor=pointer]: Structure Test
      - generic [ref=e52] [cursor=pointer]: Camera Reset Test
      - generic [ref=e53] [cursor=pointer]: Fog Link 2
      - generic [ref=e54] [cursor=pointer]: Fog Link 1
      - generic [ref=e55] [cursor=pointer]: Fog Link 0
      - generic [ref=e56] [cursor=pointer]: asteroid Node
      - generic [ref=e57] [cursor=pointer]: planet Node
      - generic [ref=e58] [cursor=pointer]: Linked to First
      - generic [ref=e59] [cursor=pointer]: First Graph Note
      - generic [ref=e60] [cursor=pointer]: Loading Test Link 5
      - generic [ref=e61] [cursor=pointer]: Loading Test Link 4
      - generic [ref=e62] [cursor=pointer]: Loading Test Link 3
      - generic [ref=e63] [cursor=pointer]: Loading Test Link 1
      - generic [ref=e64] [cursor=pointer]: Loading Test Link 0
      - generic [ref=e65] [cursor=pointer]: Progressive Loadi...
      - generic [ref=e66] [cursor=pointer]: Linked Note 4
      - generic [ref=e67] [cursor=pointer]: Linked Note 3
      - generic [ref=e68] [cursor=pointer]: Linked Note 2
      - generic [ref=e69] [cursor=pointer]: Linked Note 1
      - generic [ref=e70] [cursor=pointer]: Горький Призыв~
```

# Test source

```ts
  156 |     // Navigate back to 3D graph
  157 |     await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
  158 |     await page.waitForLoadState('networkidle');
  159 |     await page.waitForTimeout(4000);
  160 |     
  161 |     // 3D graph should reload correctly
  162 |     const container3DAgain = page.locator('.graph-3d-container').first();
  163 |     await expect(container3DAgain).toBeVisible();
  164 |   });
  165 | 
  166 |   test('should transition from 2D to 3D graph maintaining context', async ({ page, request }) => {
  167 |     const note = await request.post('http://localhost:8080/notes', {
  168 |       data: { title: '2D to 3D Transition', content: 'Testing 2D to 3D', type: 'planet' }
  169 |     });
  170 |     const noteId = (await note.json()).id;
  171 |     
  172 |     // Start at 2D graph page
  173 |     await page.goto('http://localhost:5173/graph');
  174 |     await page.waitForLoadState('networkidle');
  175 |     await page.waitForTimeout(3000);
  176 |     
  177 |     // Verify 2D graph loads
  178 |     const graph2D = page.locator('.graph-container').first();
  179 |     await expect(graph2D).toBeVisible();
  180 |     
  181 |     // Navigate to 3D version
  182 |     await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
  183 |     await page.waitForLoadState('networkidle');
  184 |     await page.waitForTimeout(4000);
  185 |     
  186 |     // Verify 3D graph loads
  187 |     const graph3D = page.locator('.graph-3d-container').first();
  188 |     await expect(graph3D).toBeVisible();
  189 |     
  190 |     // Stats should show the specific note's local graph
  191 |     const statsBar = page.locator('.stats-bar').first();
  192 |     if (await statsBar.isVisible().catch(() => false)) {
  193 |       const statsText = await statsBar.textContent();
  194 |       expect(statsText).toMatch(/\d+\s*nodes/);
  195 |       expect(statsText).toMatch(/\d+\s*links/);
  196 |     }
  197 |   });
  198 | 
  199 |   test('should navigate from home page 3D button to full 3D graph', async ({ page }) => {
  200 |     // Start at home page
  201 |     await page.goto('http://localhost:5173/');
  202 |     await page.waitForLoadState('networkidle');
  203 |     await page.waitForTimeout(2000);
  204 |     
  205 |     // Find and click the 3D button
  206 |     const button3D = page.locator('.view-button-3d, button:has-text("3D")').first();
  207 |     if (await button3D.isVisible().catch(() => false)) {
  208 |       await button3D.click();
  209 |       await page.waitForTimeout(3000);
  210 |       
  211 |       // Should navigate to /graph/3d
  212 |       const currentUrl = page.url();
  213 |       expect(currentUrl).toContain('/graph/3d');
  214 |       
  215 |       // 3D graph should load
  216 |       const container = page.locator('.graph-3d-container, .lazy-loading').first();
  217 |       await expect(container).toBeVisible({ timeout: 10000 });
  218 |     }
  219 |   });
  220 | 
  221 |   test('should handle direct URL access to 3D graph routes', async ({ page, request }) => {
  222 |     const note = await request.post('http://localhost:8080/notes', {
  223 |       data: { title: 'Direct URL Test', content: 'Testing direct URL access', type: 'comet' }
  224 |     });
  225 |     const noteId = (await note.json()).id;
  226 |     
  227 |     // Test direct access to local 3D graph
  228 |     await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
  229 |     await page.waitForLoadState('networkidle');
  230 |     await page.waitForTimeout(4000);
  231 |     
  232 |     const localContainer = page.locator('.graph-3d-container').first();
  233 |     await expect(localContainer).toBeVisible();
  234 |     
  235 |     // Local mode should be indicated
  236 |     const statsLocal = page.locator('.stats-bar').first();
  237 |     if (await statsLocal.isVisible().catch(() => false)) {
  238 |       const statsText = await statsLocal.textContent();
  239 |       const hasLocalMode = statsText?.toLowerCase().includes('local');
  240 |       expect(hasLocalMode).toBe(true);
  241 |     }
  242 |     
  243 |     // Test direct access to full 3D graph
  244 |     await page.goto('http://localhost:5173/graph/3d');
  245 |     await page.waitForLoadState('networkidle');
  246 |     await page.waitForTimeout(4000);
  247 |     
  248 |     const fullContainer = page.locator('.graph-3d-container, .lazy-loading').first();
  249 |     await expect(fullContainer).toBeVisible();
  250 |     
  251 |     // Full mode should be indicated
  252 |     const statsFull = page.locator('.stats-bar').first();
  253 |     if (await statsFull.isVisible().catch(() => false)) {
  254 |       const statsText = await statsFull.textContent();
  255 |       const hasFullMode = statsText?.toLowerCase().includes('full');
> 256 |       expect(hasFullMode).toBe(true);
      |                           ^ Error: expect(received).toBe(expected) // Object.is equality
  257 |     }
  258 |   });
  259 | 
  260 |   test('should show empty state with appropriate camera position when no notes', async ({ page }) => {
  261 |     // Navigate to full 3D graph when no notes exist (or after clearing)
  262 |     await page.goto('http://localhost:5173/graph/3d');
  263 |     await page.waitForLoadState('networkidle');
  264 |     await page.waitForTimeout(3000);
  265 |     
  266 |     // Should show empty state or loading state, not error
  267 |     const emptyState = page.locator('.no-data-message, .empty-state, .lazy-loading').first();
  268 |     const graphContainer = page.locator('.graph-3d-container').first();
  269 |     
  270 |     const hasEmpty = await emptyState.isVisible().catch(() => false);
  271 |     const hasGraph = await graphContainer.isVisible().catch(() => false);
  272 |     
  273 |     expect(hasEmpty || hasGraph).toBe(true);
  274 |   });
  275 | 
  276 |   test('should position camera correctly for isolated single node', async ({ page, request }) => {
  277 |     // Create note with NO connections
  278 |     const isolatedNote = await request.post('http://localhost:8080/notes', {
  279 |       data: { title: 'Isolated Node', content: 'No connections', type: 'galaxy' }
  280 |     });
  281 |     const noteId = (await isolatedNote.json()).id;
  282 |     
  283 |     await page.goto(`http://localhost:5173/graph/3d/${noteId}`);
  284 |     await page.waitForLoadState('networkidle');
  285 |     await page.waitForTimeout(4000);
  286 |     
  287 |     // Should render without errors
  288 |     const container = page.locator('.graph-3d-container, .no-data-message').first();
  289 |     await expect(container).toBeVisible();
  290 |     
  291 |     // Stats should show 1 node (just the isolated note itself)
  292 |     const statsBar = page.locator('.stats-bar').first();
  293 |     if (await statsBar.isVisible().catch(() => false)) {
  294 |       const statsText = await statsBar.textContent();
  295 |       const nodeMatch = statsText?.match(/(\d+)\s*nodes/);
  296 |       if (nodeMatch) {
  297 |         const nodeCount = parseInt(nodeMatch[1], 10);
  298 |         expect(nodeCount).toBeGreaterThanOrEqual(1);
  299 |       }
  300 |     }
  301 |   });
  302 | });
  303 | 
```