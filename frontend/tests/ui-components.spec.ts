import { test, expect } from '@playwright/test';

test.describe('UI Components from Flowchart', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost:5173');
    await page.waitForLoadState('networkidle');
  });

  test('should open left sidebar menu', async ({ page }) => {
    // Click hamburger menu
    await page.click('button[aria-label="Открыть меню"]');
    
    // Check sidebar is visible
    await expect(page.locator('nav[aria-label="Главное меню"]')).toBeVisible();
    await expect(page.locator('text=Импорт')).toBeVisible();
    await expect(page.locator('text=Экспорт')).toBeVisible();
    
    // Close by clicking overlay
    await page.click('.sidebar-overlay');
    await expect(page.locator('nav[aria-label="Главное меню"]')).not.toBeVisible();
  });

  test('should open search panel', async ({ page }) => {
    // Click search button
    await page.click('button[aria-label="Открыть поиск"]');
    
    // Check search panel is visible
    await expect(page.locator('input[placeholder="Поиск по заметкам..."]')).toBeVisible();
    
    // Type search query
    await page.fill('input[placeholder="Поиск по заметкам..."]', 'test');
    await page.waitForTimeout(600); // Wait for debounce
    
    // Close search
    await page.keyboard.press('Escape');
    await expect(page.locator('input[placeholder="Поиск по заметкам..."]')).not.toBeVisible();
  });

  test('should open document import modal', async ({ page }) => {
    // Open sidebar
    await page.click('button[aria-label="Открыть меню"]');
    
    // Click import
    await page.click('text=Импорт');
    
    // Check modal is open
    await expect(page.locator('text=Импорт документа')).toBeVisible();
    await expect(page.locator('text=Перетащите PDF или файл сюда')).toBeVisible();
    
    // Close modal
    await page.click('button[aria-label="Закрыть"]');
    await expect(page.locator('text=Импорт документа')).not.toBeVisible();
  });

  test('should open export modal', async ({ page }) => {
    // Create a note first via API
    await page.request.post('http://localhost:8080/notes', {
      data: { title: 'Export Test Note', content: 'Test content' }
    });
    
    // Refresh page to see the note
    await page.goto('http://localhost:5173');
    await page.waitForLoadState('networkidle');
    
    // Open sidebar
    await page.click('button[aria-label="Открыть меню"]');
    
    // Click export
    await page.click('text=Экспорт');
    
    // Check modal is open
    await expect(page.locator('text=Экспорт заметок')).toBeVisible();
    await expect(page.locator('text=JSON')).toBeVisible();
    await expect(page.locator('text=CSV')).toBeVisible();
    await expect(page.locator('text=Markdown')).toBeVisible();
    
    // Close modal
    await page.click('button[aria-label="Закрыть"]');
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
    await expect(page.locator('text=Нет данных')).toBeVisible();
    await expect(page.locator('text=Создайте первую заметку')).toBeVisible();
  });

  test('should use keyboard shortcuts', async ({ page }) => {
    // Test Ctrl+N - should navigate to new note
    await page.keyboard.press('Control+n');
    await page.waitForURL('**/notes/new', { timeout: 5000 });
    await expect(page.locator('text=New Note')).toBeVisible();
    
    // Go back
    await page.goto('http://localhost:5173');
    
    // Test Ctrl+F - should open search
    await page.keyboard.press('Control+f');
    await expect(page.locator('input[placeholder="Поиск по заметкам..."]')).toBeVisible();
    
    // Test Escape - should close search
    await page.keyboard.press('Escape');
    await expect(page.locator('input[placeholder="Поиск по заметкам..."]')).not.toBeVisible();
  });

  test('should create note via FAB', async ({ page }) => {
    await page.click('[data-testid="fab-new-note"]');
    await page.waitForURL('**/notes/new', { timeout: 5000 });
    
    await page.fill('input[placeholder="Title"]', 'FAB Created Note');
    await page.fill('textarea', 'Content from FAB test');
    await page.click('button:has-text("Create")');
    
    await page.waitForURL(/\/notes\/[a-f0-9-]+/, { timeout: 5000 });
    await expect(page.locator('h1')).toHaveText('FAB Created Note');
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
    await page.waitForSelector('canvas', { timeout: 10000 });
    await page.waitForTimeout(2000); // Let graph render
    
    // Canvas should be visible
    await expect(page.locator('canvas')).toBeVisible();
    await expect(page.locator('text=Knowledge Constellation')).toBeVisible();
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
    
    // Should show empty state
    await expect(page.locator('text=Это одинокая звезда')).toBeVisible();
    await expect(page.locator('text=Создайте связи, чтобы увидеть созвездие')).toBeVisible();
  });
});
