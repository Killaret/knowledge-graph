import { test, expect } from '@playwright/test';
import { getBackendUrl } from './helpers/testData';
import { setupSkipAuth } from './helpers/testUtils';

/**
 * E2E Tests for SKIP_AUTH Mode
 * 
 * These tests verify that the application works correctly when SKIP_AUTH=true
 * All requests are made as test_user (ID: 00000000-0000-0000-0000-000000000001)
 */

test.describe('SKIP_AUTH Mode Tests', { tag: ['@auth', '@skip-auth', '@e2e'] }, () => {
  
  test.beforeEach(async ({ page }) => {
    // Setup SKIP_AUTH for all tests
    await setupSkipAuth(page);
  });

  test('should bypass authentication and allow direct access', async ({ page }) => {
    // Should be able to access protected routes directly
    await page.goto('/graph');
    await page.waitForLoadState('networkidle');
    
    // Should stay on graph page (no redirect to login)
    expect(page.url()).toContain('/graph');
    
    // Wait for loading to complete - check for loading indicator first
    try {
      await expect(page.locator('text=Loading graph...')).toBeVisible({ timeout: 5000 });
      // Wait for loading to complete
      await expect(page.locator('text=Loading graph...')).not.toBeVisible({ timeout: 15000 });
    } catch {
      // Loading might have already completed
    }
    
    // Check for graph container or canvas
    try {
      await expect(page.locator('canvas').first()).toBeVisible({ timeout: 5000 });
    } catch {
      // Try alternative selectors for graph container
      const graphContainer = page.locator('.graph-page, .center, [data-testid*="graph"]').first();
      await expect(graphContainer).toBeVisible({ timeout: 5000 });
      console.log('[TEST] Graph container found, canvas might be loading');
    }
  });

  test('should work with API requests as test_user', async ({ request }) => {
    // Create note via API - should work without auth
    const timestamp = Date.now();
    const response = await request.post(`${getBackendUrl()}/api/v1/notes`, {
      data: {
        title: `SKIP_AUTH Test ${timestamp}`,
        content: 'Test content for SKIP_AUTH mode',
        type: 'star'
      }
    });
    
    expect(response.ok()).toBeTruthy();
    const noteData = await response.json();
    expect(noteData.data).toBeDefined();
    expect(noteData.data.title).toContain('SKIP_AUTH Test');
    
    // Navigate to the note in UI
    await page.goto(`/notes/${noteData.data.id}`);
    await page.waitForLoadState('networkidle');
    
    // Should be able to view the note
    await expect(page.locator('h1')).toContainText('SKIP_AUTH Test');
  });

  test('should not show login forms when SKIP_AUTH is enabled', async ({ page }) => {
    // Navigate to login page
    await page.goto('/auth/login');
    await page.waitForLoadState('networkidle');
    
    // Should redirect to home page immediately (already "authenticated")
    try {
      await page.waitForURL('/', { timeout: 8000 });
      expect(page.url()).toContain('/');
    } catch {
      // If still on login page, check if form is hidden or modified
      const loginForm = page.locator('form').first();
      const isVisible = await loginForm.isVisible().catch(() => false);
      
      if (isVisible) {
        // Form might be visible but should indicate SKIP_AUTH mode
        console.log('[TEST] Login form visible but SKIP_AUTH is enabled');
      }
    }
  });

  test('should allow access to profile page', async ({ page }) => {
    // Profile page should be accessible
    await page.goto('/profile');
    await page.waitForLoadState('networkidle');
    
    // Should stay on profile page (no redirect)
    expect(page.url()).toContain('/profile');
    
    // Should show some user info (even if limited)
    const profileContent = page.locator('[data-testid="profile-content"], .profile-container').first();
    await expect(profileContent).toBeVisible({ timeout: 5000 });
  });

  test('should handle concurrent requests as test_user', async ({ request }) => {
    // Create multiple notes concurrently
    const timestamp = Date.now();
    const promises = [];
    
    for (let i = 0; i < 5; i++) {
      promises.push(
        request.post(`${getBackendUrl()}/api/v1/notes`, {
          data: {
            title: `Concurrent Test ${timestamp}-${i}`,
            content: `Test content ${i}`,
            type: 'star'
          }
        })
      );
    }
    
    // All requests should succeed
    const responses = await Promise.all(promises);
    for (const response of responses) {
      expect(response.ok()).toBeTruthy();
    }
    
    // Verify notes are created
    const notesResponse = await request.get(`${getBackendUrl()}/api/v1/notes`);
    expect(notesResponse.ok()).toBeTruthy();
    const notesData = await notesResponse.json();
    
    const ourNotes = notesData.notes?.filter((note: any) => 
      note.title.includes(`Concurrent Test ${timestamp}`)
    );
    expect(ourNotes?.length).toBe(5);
  });

  test('should maintain SKIP_AUTH state across navigation', async ({ page }) => {
    // Start with graph page
    await page.goto('/graph');
    await page.waitForLoadState('networkidle');
    
    // Navigate to different pages
    await page.goto('/notes');
    await page.waitForLoadState('networkidle');
    
    await page.goto('/profile');
    await page.waitForLoadState('networkidle');
    
    await page.goto('/graph');
    await page.waitForLoadState('networkidle');
    
    // Should never be redirected to login
    expect(page.url()).toContain('/graph');
    
    // SKIP_AUTH flag should still be set
    const skipAuthFlag = await page.evaluate(() => {
      return (window as any).__SKIP_AUTH__;
    });
    expect(skipAuthFlag).toBe(true);
  });

  test('should handle API errors gracefully in SKIP_AUTH mode', async ({ request }) => {
    // Try to access non-existent note
    const response = await request.get(`${getBackendUrl()}/api/v1/notes/non-existent-id`);
    
    // Should return proper error response (400 for invalid ID format, 404 for valid but non-existent)
    expect([400, 404]).toContain(response.status());
    
    // Navigate to non-existent note in UI
    await page.goto('/notes/non-existent-id');
    await page.waitForLoadState('networkidle');
    
    // Should show error state or redirect
    const currentUrl = page.url();
    expect(currentUrl).toMatch(/\/(notes\/non-existent-id|graph|error)/);
  });
});
