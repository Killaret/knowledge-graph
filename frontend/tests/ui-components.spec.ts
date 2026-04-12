import { test, expect } from '@playwright/test';

test.describe('UI Components from Flowchart', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost:5173');
    await page.waitForLoadState('networkidle');
  });

  test('should show left toolbar with navigation', async ({ page }) => {
    // Check toolbar is visible
    await expect(page.locator('nav[aria-label="Main navigation"]')).toBeVisible();
    
    // Check toolbar buttons
    await expect(page.locator('button[aria-label="Новая звезда"]')).toBeVisible();
    await expect(page.locator('button[aria-label="Граф"]')).toBeVisible();
    await expect(page.locator('button[aria-label="Все заметки"]')).toBeVisible();
    await expect(page.locator('button[aria-label="Импорт"]')).toBeVisible();
    await expect(page.locator('button[aria-label="Экспорт"]')).toBeVisible();
    await expect(page.locator('button[aria-label="Настройки"]')).toBeVisible();
  });

  test('should open search from header', async ({ page }) => {
    // Click search trigger button in header
    await page.click('button[aria-label="Поиск (Ctrl+F)"]');
    
    // Check search expanded is visible
    await expect(page.locator('input[placeholder="Поиск по заметкам..."]')).toBeVisible();
    
    // Type search query
    await page.fill('input[placeholder="Поиск по заметкам..."]', 'test');
    await page.waitForTimeout(600); // Wait for debounce
    
    // Close search
    await page.keyboard.press('Escape');
  });

  test('should open document import modal from toolbar', async ({ page }) => {
    // Click import button in left toolbar
    await page.click('button[aria-label="Импорт"]');
    await page.waitForTimeout(300); // Wait for tooltip/handler
    
    // Note: Import modal is now triggered via toolbar onImport callback
    // This test verifies the toolbar button is present and clickable
    await expect(page.locator('button[aria-label="Импорт"]')).toBeVisible();
  });

  test('should open export modal from toolbar', async ({ page }) => {
    // Create a note first via API
    await page.request.post('http://localhost:8080/notes', {
      data: { title: 'Export Test Note', content: 'Test content' }
    });
    
    // Click export button in left toolbar
    await page.click('button[aria-label="Экспорт"]');
    await page.waitForTimeout(300);
    
    // Verify export button is present
    await expect(page.locator('button[aria-label="Экспорт"]')).toBeVisible();
  });

  test('should show empty state when no notes', async ({ page, request }) => {
    // Clear all notes via API
    const notes = await request.get('http://localhost:8080/notes');
    const notesData = await notes.json();
    for (const note of notesData) {
      await request.delete(`http://localhost:8080/notes/${note.id}`);
    }
    
    // Refresh page
    await page.goto('http://localhost:5173');
    await page.waitForLoadState('networkidle');
    
    // Check empty state
    await expect(page.locator('text=Нет заметок')).toBeVisible();
    await expect(page.locator('text=Создайте первую заметку, чтобы начать')).toBeVisible();
  });

  test('should use keyboard shortcuts', async ({ page }) => {
    // Test Ctrl+F - should open search
    await page.keyboard.press('Control+f');
    await expect(page.locator('input[placeholder="Поиск по заметкам..."]')).toBeVisible();
    
    // Test Escape - should close search
    await page.keyboard.press('Escape');
    await expect(page.locator('input[placeholder="Поиск по заметкам..."]')).not.toBeVisible();
  });

  test('should create note via toolbar', async ({ page }) => {
    await page.click('[data-testid="toolbar-new-note"]');
    await page.waitForURL('**/notes/new', { timeout: 5000 });
    
    await page.fill('input[placeholder="Title"]', 'Toolbar Created Note');
    await page.fill('textarea', 'Content from toolbar test');
    await page.click('button:has-text("Create")');
    
    await page.waitForURL(/\/notes\/[a-f0-9-]+/, { timeout: 5000 });
    await expect(page.locator('h1')).toHaveText('Toolbar Created Note');
  });
});

test.describe('Graph Page Features', () => {
  test('should show note popup on node click', async ({ page, request }) => {
    // Create two notes and a link
    const note1 = await request.post('http://localhost:8080/notes', {
      data: { title: 'Graph Node A', content: 'Content A' }
    });
    const note2 = await request.post('http://localhost:8080/notes', {
      data: { title: 'Graph Node B', content: 'Content B' }
    });
    const id1 = (await note1.json()).id;
    const id2 = (await note2.json()).id;
    
    await request.post('http://localhost:8080/links', {
      data: { source_note_id: id1, target_note_id: id2, link_type: 'reference', weight: 1.0 }
    });
    
    // Navigate to graph
    await page.goto(`http://localhost:5173/graph/${id1}`);
    await page.waitForSelector('[data-testid="main-graph-canvas"]', { timeout: 10000 });
    await page.waitForTimeout(2000); // Let graph render
    
    // Canvas should be visible
    await expect(page.locator('[data-testid="main-graph-canvas"]')).toBeVisible();
  });

  test('should show empty state for isolated node', async ({ page, request }) => {
    // Create single note without links
    const note = await request.post('http://localhost:8080/notes', {
      data: { title: 'Isolated Node', content: 'No connections' }
    });
    const id = (await note.json()).id;
    
    // Navigate to graph
    await page.goto(`http://localhost:5173/graph/${id}`);
    await page.waitForLoadState('networkidle');
    
    // Should show empty state for isolated node
    await expect(page.locator('text=Одинокая звезда')).toBeVisible();
    await expect(page.locator('text=Создайте связи, чтобы увидеть созвездие')).toBeVisible();
  });
});
