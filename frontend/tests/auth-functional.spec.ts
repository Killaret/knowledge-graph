import { test, expect } from '@playwright/test';
import { getBackendUrl } from './helpers/testData';

/**
 * Functional E2E Tests for Authentication
 * Tests actual registration and login flows with API
 */

const TEST_USER_PREFIX = 'test_user_';
const TEST_PASSWORD = 'TestPassword123!';

test.describe('Auth Functional Tests', { tag: ['@auth', '@e2e'] }, () => {
  
  test('should register new user successfully', async ({ page, request }) => {
    const uniqueId = Date.now();
    const login = `${TEST_USER_PREFIX}${uniqueId}`;
    
    // Navigate to register page
    await page.goto('/auth/register');
    await page.waitForSelector('form', { timeout: 10000 });
    
    // Fill registration form
    await page.fill('input#login', login);
    await page.fill('input#password', TEST_PASSWORD);
    await page.fill('input#confirmPassword', TEST_PASSWORD);
    
    // Submit form
    await page.click('button[type="submit"]');
    
    // Should redirect to login or graph after successful registration
    await page.waitForURL(/\/(auth\/login|graph)/, { timeout: 10000 });
    
    // Verify we're logged in (check for logout button or user info)
    const logoutBtn = page.locator('text=Выйти').first();
    const userMenu = page.locator('[data-testid="user-menu"]').first();
    expect(await logoutBtn.isVisible().catch(() => false) || 
           await userMenu.isVisible().catch(() => false)).toBeTruthy();
  });

  test('should login with existing user', async ({ page }) => {
    // First register a user
    const uniqueId = Date.now();
    const login = `${TEST_USER_PREFIX}${uniqueId}`;
    
    // Register via API
    await page.request.post(`${getBackendUrl()}/auth/register`, {
      data: { login, password: TEST_PASSWORD }
    });
    
    // Navigate to login
    await page.goto('/auth/login');
    await page.waitForSelector('form', { timeout: 10000 });
    
    // Fill login form
    await page.fill('input#login', login);
    await page.fill('input#password', TEST_PASSWORD);
    
    // Submit
    await page.click('button[type="submit"]');
    
    // Should redirect to graph page
    await page.waitForURL('/graph', { timeout: 10000 });
    
    // Verify we're on graph page
    await expect(page.locator('canvas').first()).toBeVisible({ timeout: 10000 });
  });

  test('should show error for invalid login credentials', async ({ page }) => {
    await page.goto('/auth/login');
    await page.waitForSelector('form', { timeout: 10000 });
    
    // Fill with non-existent user
    await page.fill('input#login', 'nonexistent_user_12345');
    await page.fill('input#password', 'wrongpassword');
    
    // Submit
    await page.click('button[type="submit"]');
    
    // Should show error message
    const errorMessage = page.locator('.error, [role="alert"], .alert').first();
    await expect(errorMessage).toBeVisible({ timeout: 5000 });
  });

  test('should show validation errors for weak password', async ({ page }) => {
    await page.goto('/auth/register');
    await page.waitForSelector('form', { timeout: 10000 });
    
    // Fill with weak password
    await page.fill('input#login', `test_${Date.now()}`);
    await page.fill('input#password', '123');
    await page.fill('input#confirmPassword', '123');
    
    // Submit
    await page.click('button[type="submit"]');
    
    // Should stay on register page and show validation errors
    await expect(page).toHaveURL(/.*register.*/);
    
    // Check for password requirements
    const requirements = page.locator('.password-requirements li').first();
    await expect(requirements).toBeVisible({ timeout: 5000 });
  });

  test('should logout successfully', async ({ page }) => {
    // Login first
    const uniqueId = Date.now();
    const login = `${TEST_USER_PREFIX}${uniqueId}`;
    
    await page.request.post(`${getBackendUrl()}/auth/register`, {
      data: { login, password: TEST_PASSWORD }
    });
    
    await page.goto('/auth/login');
    await page.fill('input#login', login);
    await page.fill('input#password', TEST_PASSWORD);
    await page.click('button[type="submit"]');
    await page.waitForURL('/graph', { timeout: 10000 });
    
    // Now logout
    const logoutBtn = page.locator('text=Выйти').first();
    await logoutBtn.click();
    
    // Should redirect to login
    await page.waitForURL('/auth/login', { timeout: 10000 });
    
    // Verify login form is visible
    await expect(page.locator('input#login')).toBeVisible();
  });

  test('should protect routes when not authenticated', async ({ page }) => {
    // Try to access graph without auth
    await page.goto('/graph');
    
    // Should redirect to login
    await page.waitForURL('/auth/login', { timeout: 10000 });
  });

  test('should stay on page when accessing public routes', async ({ page }) => {
    // These should work without auth
    const publicRoutes = ['/auth/login', '/auth/register', '/auth/forgot-password'];
    
    for (const route of publicRoutes) {
      await page.goto(route);
      await expect(page).toHaveURL(route);
      
      // Verify form is visible
      const form = page.locator('form').first();
      await expect(form).toBeVisible({ timeout: 5000 });
    }
  });
});
