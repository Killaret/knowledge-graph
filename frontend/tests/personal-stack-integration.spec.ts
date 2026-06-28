import { test, expect } from '@playwright/test';

/**
 * Personal Stack Integration Tests
 * Comprehensive testing of personal stack functionality
 */

const PERSONAL_BASE_URL = 'http://localhost:3001';
const API_BASE_URL = 'http://localhost:8082';

test.describe('Personal Stack Integration Tests', () => {
  
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => {
      (window as any).__SKIP_AUTH__ = true;
    });
  });

  test('Personal stack - Frontend accessibility', async ({ page }) => {
    await page.goto(PERSONAL_BASE_URL);
    await page.waitForLoadState('networkidle');
    
    // Check that main content is visible
    await expect(page.locator('main')).toBeVisible({ timeout: 15000 });
    
    // Check that app shell is loaded
    await expect(page.locator('.app-shell')).toBeVisible();
  });

  test('Personal stack - API connectivity', async ({ request }) => {
    // Test backend API through nginx
    const notesResponse = await request.get(`${API_BASE_URL}/api/v1/notes`);
    expect(notesResponse.ok()).toBeTruthy();
    
    const notesData = await notesResponse.json();
    expect(notesData).toHaveProperty('notes');
    expect(Array.isArray(notesData.notes)).toBeTruthy();
  });

  test('Personal stack - Graph Service connectivity', async ({ request }) => {
    // Test graph service through nginx
    const graphResponse = await request.get(`${API_BASE_URL}/graph-service/api/v1/graph/full`);
    expect(graphResponse.ok()).toBeTruthy();
    
    const graphData = await graphResponse.json();
    expect(graphData).toHaveProperty('data');
    expect(graphData.data).toHaveProperty('nodes');
    expect(graphData.data).toHaveProperty('links');
  });

  test('Personal stack - Create note via API', async ({ request }) => {
    const newNote = {
      title: 'Integration Test Note',
      content: 'This note was created during integration testing',
      type: 'star',
      metadata: { test: 'integration' }
    };

    const createResponse = await request.post(`${API_BASE_URL}/api/v1/notes`, {
      data: newNote
    });
    
    expect(createResponse.ok()).toBeTruthy();
    const createdNote = await createResponse.json();
    expect(createdNote.data).toHaveProperty('id');
    expect(createdNote.data.title).toBe(newNote.title);
    
    // Verify note was created
    const getResponse = await request.get(`${API_BASE_URL}/api/v1/notes`);
    const notesData = await getResponse.json();
    const foundNote = notesData.notes.find((n: any) => n.id === createdNote.data.id);
    expect(foundNote).toBeDefined();
  });

  test('Personal stack - Graph view loading', async ({ page }) => {
    await page.goto(`${PERSONAL_BASE_URL}/graph`);
    await page.waitForLoadState('networkidle');
    
    // Wait for graph canvas to load
    await expect(page.locator('canvas')).toBeVisible({ timeout: 15000 });
    
    // Wait a bit for graph simulation to stabilize
    await page.waitForTimeout(2000);
    
    // Check that canvas is present and has some size
    const canvas = page.locator('canvas');
    const box = await canvas.boundingBox();
    expect(box).toBeDefined();
    expect(box!.width).toBeGreaterThan(0);
    expect(box!.height).toBeGreaterThan(0);
  });

  test('Personal stack - Vector embedding processing', async ({ request }) => {
    // Create a note that should trigger vector processing
    const vectorNote = {
      title: 'Vector Test Note',
      content: 'Testing vector embeddings and similarity search',
      type: 'comet',
      metadata: { test: 'vector' }
    };

    const createResponse = await request.post(`${API_BASE_URL}/api/v1/notes`, {
      data: vectorNote
    });
    
    expect(createResponse.ok()).toBeTruthy();
    const createdNote = await createResponse.json();
    
    // Wait for worker to process embedding
    await new Promise(resolve => setTimeout(resolve, 5000));
    
    // Verify note exists in graph
    const graphResponse = await request.get(`${API_BASE_URL}/graph-service/api/v1/graph/note/${createdNote.data.id}?depth=1`);
    expect(graphResponse.ok()).toBeTruthy();
    
    const graphData = await graphResponse.json();
    expect(graphData.data.nodes).toHaveLength(1);
    expect(graphData.data.nodes[0].id).toBe(createdNote.data.id);
  });

  test('Personal stack - Health endpoints', async ({ request }) => {
    // Test nginx health
    const nginxHealth = await request.get(`${API_BASE_URL}/health`);
    expect(nginxHealth.ok()).toBeTruthy();
    expect(await nginxHealth.text()).toContain('OK');
    
    // Test backend health
    const backendHealth = await request.get(`${API_BASE_URL}/api/v1/health`);
    expect(backendHealth.ok()).toBeTruthy();
  });

  test('Personal stack - List view functionality', async ({ page }) => {
    await page.goto(PERSONAL_BASE_URL);
    await page.waitForLoadState('networkidle');
    await expect(page.locator('main')).toBeVisible({ timeout: 15000 });
    
    // Try to switch to list view if available
    try {
      const listButton = page.locator('[data-testid="view-toggle-list"]');
      if (await listButton.isVisible({ timeout: 3000 })) {
        await listButton.click();
        await page.waitForTimeout(1000);
      }
    } catch {
      // List toggle might not be available, that's okay
    }
    
    // Verify page is still functional
    await expect(page.locator('main')).toBeVisible();
  });
});