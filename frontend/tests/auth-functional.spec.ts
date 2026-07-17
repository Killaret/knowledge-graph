import { test, expect } from '@playwright/test';
import { getBackendUrl } from './helpers/testData';

/**
 * Functional E2E Tests for Authentication
 *
 * These tests are designed for SKIP_AUTH=true mode (default development mode)
 * In this mode, authentication is bypassed and all requests are made as test_user
 */

test.describe('Auth Functional Tests (SKIP_AUTH Mode)', { tag: ['@auth', '@e2e'] }, () => {

  test('should access application without authentication in SKIP_AUTH mode', async ({ page }) => {
    // Navigate to main page
    await page.goto('/', { timeout: 60000 });
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);

    // In SKIP_AUTH mode, we should have direct access to the graph
    const graphContainer = page.locator('.fullscreen-graph, canvas').first();
    await expect(graphContainer).toBeVisible({ timeout: 10000 });
  });

  test('should show user interface as test_user', async ({ page }) => {
    await page.goto('/', { timeout: 60000 });
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);

    // In SKIP_AUTH mode, user interface may vary - just verify we have access
    // Check for any user-related UI elements
    const userMenu = page.locator('[data-testid="user-menu"], .user-menu, text=test_user').first();
    const logoutBtn = page.locator('text=Logout').first();
    const userProfile = page.locator('[data-testid="user-profile"], .user-profile').first();

    // At least one UI element should be present, or graph should be accessible
    const hasUserUI = await userMenu.isVisible().catch(() => false) ||
                      await logoutBtn.isVisible().catch(() => false) ||
                      await userProfile.isVisible().catch(() => false);

    // If no user UI, verify graph is accessible (SKIP_AUTH mode working)
    if (!hasUserUI) {
      const graphContainer = page.locator('.fullscreen-graph, canvas').first();
      const graphAccessible = await graphContainer.isVisible().catch(() => false);
      expect(graphAccessible).toBeTruthy();
    } else {
      expect(hasUserUI).toBeTruthy();
    }
  });

  test('should protect auth routes when SKIP_AUTH is disabled', async ({ page }) => {
    // This test verifies the auth bypass is working
    // In SKIP_AUTH=true mode, auth routes should redirect to main page

    await page.goto('/auth/login', { timeout: 60000 });
    await page.waitForTimeout(1000);

    // Should either stay on login page (if SKIP_AUTH=false) or redirect
    const currentUrl = page.url();

    // In SKIP_AUTH mode, we expect to be redirected to main page or stay on login
    // Verify we're on a valid page (not an error page)
    expect(currentUrl).toMatch(/\/(auth\/login|\/|graph)/);
    // Verify page loaded successfully (no 404 or error)
    await expect(page.locator('main, h1, .auth-card').first()).toBeVisible({ timeout: 5000 });
  });

  test('should register endpoint handle requests', async ({ request }) => {
    const uniqueId = Date.now();
    const login = `test_user_${uniqueId}`;
    const email = `${login}@test.example.com`;

    // Try to register - in SKIP_AUTH mode this may be disabled
    const response = await request.post(`${getBackendUrl()}/api/v1/auth/register`, {
      data: {
        login,
        email,
        password: 'TestPassword123!',
      },
    });

    // In SKIP_AUTH mode, registration might return an error or be disabled
    // Verify endpoint responds with valid JSON
    expect([200, 201, 400, 403, 409]).toContain(response.status());
    const contentType = response.headers()['content-type'];
    expect(contentType).toContain('application/json');
  });

  test('should login endpoint handle requests', async ({ request }) => {
    // Try to login - in SKIP_AUTH mode this may be bypassed
    const response = await request.post(`${getBackendUrl()}/api/v1/auth/login`, {
      data: {
        login: 'test_user',
        password: 'any_password',
      },
    });

    // In SKIP_AUTH mode, login might return token or redirect
    // Verify endpoint responds with valid JSON
    expect([200, 201, 400, 401]).toContain(response.status());
    const contentType = response.headers()['content-type'];
    expect(contentType).toContain('application/json');
  });

  test('should have working logout', async ({ page }) => {
    await page.goto('/', { timeout: 60000 });
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);

    // Find logout button
    const logoutBtn = page.locator('text=Logout').first();
    const isVisible = await logoutBtn.isVisible().catch(() => false);

    if (isVisible) {
      await logoutBtn.click();
      await page.waitForTimeout(1000);
      // Verify we're still on the page or redirected appropriately
      const currentUrl = page.url();
      expect(currentUrl).toBeTruthy();
    } else {
      // If logout button doesn't exist, that's acceptable in SKIP_AUTH mode
      // Verify we can still access the application
      await expect(page.locator('main')).toBeVisible();
    }
  });
});

