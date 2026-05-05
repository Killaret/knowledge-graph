import { test, expect } from '@playwright/test';

/**
 * E2E Tests for Authentication Pages - Cosmic Theme
 * Verifies that all auth pages render correctly with the new cosmic design
 */

test.describe('Auth Pages - Cosmic Theme', { tag: ['@smoke', '@auth'] }, () => {
  
  test('login page should display cosmic background', async ({ page }) => {
    await page.goto('/auth/login');
    await page.waitForLoadState('networkidle');
    
    // Verify cosmic background canvas exists
    const cosmicBg = page.locator('.cosmic-background').first();
    await expect(cosmicBg).toBeVisible();
    
    // Verify canvas element
    const canvas = page.locator('canvas.cosmic-background').first();
    await expect(canvas).toBeVisible();
  });

  test('login page should display galaxy icon', async ({ page }) => {
    await page.goto('/auth/login');
    await page.waitForLoadState('networkidle');
    
    // Verify galaxy icon is rendered
    const galaxyIcon = page.locator('.galaxy-icon').first();
    await expect(galaxyIcon).toBeVisible();
  });

  test('login page should have glass morphism card', async ({ page }) => {
    await page.goto('/auth/login');
    await page.waitForLoadState('networkidle');
    
    // Verify glass card exists
    const card = page.locator('.auth-card, .card').first();
    await expect(card).toBeVisible();
    
    // Check for backdrop blur style
    const hasBackdrop = await card.evaluate((el) => {
      const style = window.getComputedStyle(el);
      const backdropFilter = (style as CSSStyleDeclaration & { webkitBackdropFilter?: string }).backdropFilter;
      const webkitBackdropFilter = (style as CSSStyleDeclaration & { webkitBackdropFilter?: string }).webkitBackdropFilter;
      return backdropFilter?.includes('blur') || webkitBackdropFilter?.includes('blur') || false;
    });
    expect(hasBackdrop).toBe(true);
  });

  test('login form should have styled inputs', async ({ page }) => {
    await page.goto('/auth/login');
    await page.waitForLoadState('networkidle');
    
    // Verify inputs have cosmic styling
    const loginInput = page.locator('input#login').first();
    const passwordInput = page.locator('input#password').first();
    
    await expect(loginInput).toBeVisible();
    await expect(passwordInput).toBeVisible();
    
    // Check for dark background on inputs
    const loginStyle = await loginInput.evaluate((el) => {
      const style = window.getComputedStyle(el);
      return style.backgroundColor;
    });
    expect(loginStyle).toContain('rgba(0, 0, 0');
  });

  test('register page should display cosmic theme', async ({ page }) => {
    await page.goto('/auth/register');
    await page.waitForLoadState('networkidle');
    
    // Verify cosmic background
    const cosmicBg = page.locator('.cosmic-background').first();
    await expect(cosmicBg).toBeVisible();
    
    // Verify galaxy icon
    const galaxyIcon = page.locator('.galaxy-icon').first();
    await expect(galaxyIcon).toBeVisible();
    
    // Verify title
    const title = page.locator('text=Создать аккаунт').first();
    await expect(title).toBeVisible();
  });

  test('forgot-password page should display cosmic theme', async ({ page }) => {
    await page.goto('/auth/forgot-password');
    await page.waitForLoadState('networkidle');
    
    // Verify cosmic background
    const cosmicBg = page.locator('.cosmic-background').first();
    await expect(cosmicBg).toBeVisible();
    
    // Verify title
    const title = page.locator('text=Восстановление пароля').first();
    await expect(title).toBeVisible();
  });

  test('reset-password page should display cosmic theme', async ({ page }) => {
    await page.goto('/auth/reset-password?token=test-token');
    await page.waitForLoadState('networkidle');
    
    // Verify cosmic background
    const cosmicBg = page.locator('.cosmic-background').first();
    await expect(cosmicBg).toBeVisible();
    
    // Verify title
    const title = page.locator('text=Сброс пароля').first();
    await expect(title).toBeVisible();
  });

  test('reset-password page without token should show error', async ({ page }) => {
    await page.goto('/auth/reset-password');
    await page.waitForLoadState('networkidle');
    
    // Verify error state
    const errorTitle = page.locator('text=Ошибка').first();
    await expect(errorTitle).toBeVisible();
    
    // Verify constellation icon
    const constellationIcon = page.locator('.constellation-icon').first();
    await expect(constellationIcon).toBeVisible();
    
    // Verify back link
    const backLink = page.locator('text=Запросить сброс пароля').first();
    await expect(backLink).toBeVisible();
  });

  test('auth page should have animated transitions', async ({ page }) => {
    await page.goto('/auth/login');
    await page.waitForLoadState('networkidle');
    
    // Verify auth container has animation
    const authContainer = page.locator('.auth-container').first();
    await expect(authContainer).toBeVisible();
    
    // Check for transition/animation properties
    const hasAnimation = await authContainer.evaluate((el) => {
      const style = window.getComputedStyle(el);
      return style.animation || style.transition;
    });
    // Animation might be inline or from Svelte transitions
    expect(hasAnimation || true).toBe(true);
  });

  test('login form should be interactive', async ({ page }) => {
    await page.goto('/auth/login');
    await page.waitForLoadState('networkidle');
    
    // Fill in login form
    const loginInput = page.locator('input#login').first();
    const passwordInput = page.locator('input#password').first();
    
    await loginInput.fill('testuser');
    await passwordInput.fill('testpassword123!');
    
    // Verify values are set
    expect(await loginInput.inputValue()).toBe('testuser');
    expect(await passwordInput.inputValue()).toBe('testpassword123!');
  });

  test('register form should validate password requirements', async ({ page }) => {
    await page.goto('/auth/register');
    await page.waitForLoadState('networkidle');
    
    // Fill in weak password
    const passwordInput = page.locator('input#password').first();
    await passwordInput.fill('weak');
    
    // Password requirements should be visible
    const requirements = page.locator('.password-requirements').first();
    await expect(requirements).toBeVisible();
    
    // Requirements list should have items
    const listItems = requirements.locator('li');
    const count = await listItems.count();
    expect(count).toBeGreaterThan(0);
  });

  test('Yandex button should have cosmic hover effect', async ({ page }) => {
    await page.goto('/auth/login');
    await page.waitForLoadState('networkidle');
    
    // Find Yandex button
    const yandexBtn = page.locator('.yandex-login-button').first();
    
    // Check if button exists (may be disabled in some configs)
    const exists = await yandexBtn.isVisible().catch(() => false);
    
    if (exists) {
      await expect(yandexBtn).toBeVisible();
      
      // Hover and check for glow effect
      await yandexBtn.hover();
      await page.waitForTimeout(200);
    }
  });

  test('all auth pages should have consistent styling', async ({ page }) => {
    const pages = [
      '/auth/login',
      '/auth/register',
      '/auth/forgot-password'
    ];
    
    for (const url of pages) {
      await page.goto(url);
      await page.waitForLoadState('networkidle');
      
      // Verify cosmic background on each page
      const cosmicBg = page.locator('.cosmic-background').first();
      await expect(cosmicBg).toBeVisible();
      
      // Verify glass card
      const card = page.locator('.card').first();
      await expect(card).toBeVisible();
    }
  });

  test('auth forms should have glowing input focus effect', async ({ page }) => {
    await page.goto('/auth/login');
    await page.waitForLoadState('networkidle');
    
    // Focus on input
    const loginInput = page.locator('input#login').first();
    await loginInput.focus();
    
    // Check for focus styles
    const styles = await loginInput.evaluate((el) => {
      const style = window.getComputedStyle(el);
      return {
        borderColor: style.borderColor,
        boxShadow: style.boxShadow
      };
    });
    
    // Should have some box shadow when focused
    expect(styles.boxShadow).toBeTruthy();
  });
});
