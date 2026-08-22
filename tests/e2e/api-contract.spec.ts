/**
 * E2E API Contract Tests
 * Verifies that backend API responses match frontend expectations
 */
import { test, expect } from '@playwright/test';

const API_BASE = process.env.BACKEND_URL || 'http://localhost:8080/api';

test.describe('API Contract - Graph Endpoints', () => {
  
  test('GET /v1/graph/all returns wrapped response with data property', async ({ request }) => {
    const response = await request.get(`${API_BASE}/v1/graph/all?limit=5`);
    
    expect(response.ok()).toBeTruthy();
    
    const body = await response.json();
    
    // Contract: Response must have "data" wrapper
    expect(body).toHaveProperty('data');
    expect(body.data).toHaveProperty('nodes');
    expect(body.data).toHaveProperty('links');
    expect(Array.isArray(body.data.nodes)).toBe(true);
    expect(Array.isArray(body.data.links)).toBe(true);
    
    // Optional meta field
    if (body.meta) {
      expect(body.meta).toHaveProperty('total_nodes');
      expect(body.meta).toHaveProperty('total_links');
    }
  });

  test('GET /v1/notes/:id/graph returns wrapped response', async ({ request }) => {
    // First create a test note
    const createRes = await request.post(`${API_BASE}/v1/notes`, {
      data: {
        title: 'Contract Test Note',
        content: 'Test content',
        type: 'star'
      }
    });
    expect(createRes.ok()).toBeTruthy();
    const note = await createRes.json();
    
    // Test graph endpoint structure
    const graphRes = await request.get(`${API_BASE}/v1/notes/${note.id}/graph?depth=1`);
    expect(graphRes.ok()).toBeTruthy();
    
    const body = await graphRes.json();
    
    // Contract: Must have data wrapper with nodes and links
    expect(body).toHaveProperty('data');
    expect(body.data).toHaveProperty('nodes');
    expect(body.data).toHaveProperty('links');
    expect(Array.isArray(body.data.nodes)).toBe(true);
    expect(Array.isArray(body.data.links)).toBe(true);
    
    // Cleanup
    await request.delete(`${API_BASE}/v1/notes/${note.id}`);
  });

  test('Graph nodes have required fields', async ({ request }) => {
    const response = await request.get(`${API_BASE}/v1/graph/all?limit=3`);
    const body = await response.json();
    
    if (body.data.nodes.length > 0) {
      const node = body.data.nodes[0];
      
      // Contract: Each node must have these fields
      expect(node).toHaveProperty('id');
      expect(node).toHaveProperty('title');
      expect(node).toHaveProperty('type');
      expect(typeof node.id).toBe('string');
      expect(typeof node.title).toBe('string');
      expect(typeof node.type).toBe('string');
    }
  });

  test('Graph links have required fields', async ({ request }) => {
    const response = await request.get(`${API_BASE}/v1/graph/all?limit=50`);
    const body = await response.json();
    
    if (body.data.links && body.data.links.length > 0) {
      const link = body.data.links[0];
      
      // Contract: Each link must have source and target
      expect(link).toHaveProperty('source');
      expect(link).toHaveProperty('target');
      expect(typeof link.source).toBe('string');
      expect(typeof link.target).toBe('string');
    }
  });
});

test.describe('API Contract - Notes Endpoints', () => {
  
  test('GET /v1/notes returns paginated response', async ({ request }) => {
    const response = await request.get(`${API_BASE}/v1/notes?limit=5&offset=0`);
    
    expect(response.ok()).toBeTruthy();
    
    const body = await response.json();
    
    // Contract: Paginated response structure
    expect(body).toHaveProperty('notes');
    expect(body).toHaveProperty('total');
    expect(Array.isArray(body.notes)).toBe(true);
    expect(typeof body.total).toBe('number');
  });

  test('POST /v1/notes creates note and returns note object', async ({ request }) => {
    const response = await request.post(`${API_BASE}/v1/notes`, {
      data: {
        title: 'API Contract Test',
        content: 'Testing contract',
        type: 'star'
      }
    });
    
    expect(response.ok()).toBeTruthy();
    
    const body = await response.json();
    
    // Contract: Created note must have these fields
    expect(body).toHaveProperty('id');
    expect(body).toHaveProperty('title');
    expect(body).toHaveProperty('type');
    expect(body.title).toBe('API Contract Test');
    expect(body.type).toBe('star');
    
    // Cleanup
    await request.delete(`${API_BASE}/v1/notes/${body.id}`);
  });
});

test.describe('Frontend Integration via Vite Proxy', () => {
  
  test('Frontend proxy correctly forwards /api requests to backend', async ({ page }) => {
    // Navigate to frontend dev server
    await page.goto('http://localhost:5173');
    
    // Wait for page to load
    await page.waitForLoadState('networkidle');
    
    // Check that API calls succeed (look for successful network requests)
    const apiResponses = await page.evaluate(async () => {
      const response = await fetch('/api/v1/notes?limit=1');
      return {
        ok: response.ok,
        status: response.status,
        body: await response.json()
      };
    });
    
    expect(apiResponses.ok).toBe(true);
    expect(apiResponses.status).toBe(200);
    expect(apiResponses.body).toHaveProperty('notes');
  });
});
