import { test, expect } from '@playwright/test';

const BACKEND_URL = process.env.BACKEND_URL || 'http://127.0.0.1:8083';
const FRONTEND_URL = process.env.FRONTEND_URL || 'http://localhost:3002';

test.describe('Smoke tests - real auth flow', { tag: ['@smoke', '@auth-real'] }, () => {
  test('register, login, profile, graph, create note, logout', async ({ page, request }) => {
    const id = crypto.randomUUID();
    const loginName = `smoke_${id.replace(/-/g, '').slice(0, 12)}`;
    const password = 'TestPassword123!';
    const email = `${loginName}@test.example.com`;

    // Register user via API
    const registerResponse = await request.post(`${BACKEND_URL}/api/v1/auth/register`, {
      data: { login: loginName, email, password }
    });
    expect([200, 201]).toContain(registerResponse.status());

    // Login via UI
    await page.goto(`${FRONTEND_URL}/auth/login`);
    await page.waitForLoadState('networkidle');
    await page.fill('input[name="login"]', loginName);
    await page.fill('input[name="password"]', password);
    await page.click('button[type="submit"]');
    await page.waitForURL(`${FRONTEND_URL}/`, { timeout: 15000 });

    // Main page: public/authenticated graph canvas visible
    const graphCanvas = page.locator('[data-testid="graph-canvas"]');
    await expect(graphCanvas).toBeVisible({ timeout: 15000 });

    const graphStats = page.locator('[data-testid="graph-stats"]');
    await expect(graphStats).toContainText(/\d+\s+nodes?/i, { timeout: 10000 });

    // Profile page is accessible after authentication
    await page.goto(`${FRONTEND_URL}/profile`);
    await page.waitForLoadState('domcontentloaded');
    const profileContent = page.locator('[data-testid="profile-content"], .profile-container').first();
    await expect(profileContent).toBeVisible({ timeout: 10000 });

    // Graph view is accessible
    await page.goto(`${FRONTEND_URL}/graph`);
    await page.waitForLoadState('domcontentloaded');
    await expect(page.locator('[data-testid="graph-canvas"]')).toBeVisible({ timeout: 10000 });

    // Create a note using the ghost node form
    await page.goto(`${FRONTEND_URL}/`);
    await page.waitForLoadState('networkidle');
    await page.keyboard.press('n');

    const titleInput = page.locator('input[placeholder="Title"]').first();
    await expect(titleInput).toBeVisible({ timeout: 5000 });

    const noteTitle = `Smoke Note ${loginName}`;
    await titleInput.fill(noteTitle);
    await page.locator('button:has-text("Create")').first().click();

    // Form should close
    await expect(titleInput).not.toBeVisible({ timeout: 5000 });
    await expect(graphCanvas).toBeVisible();

    // Verify note was persisted via API using a fresh API login
    const apiLoginResponse = await request.post(`${BACKEND_URL}/api/v1/auth/login`, {
      data: { login: loginName, password }
    });
    expect(apiLoginResponse.ok()).toBe(true);
    const tokens = await apiLoginResponse.json();
    expect(tokens.access_token).toBeTruthy();

    const notesResponse = await request.get(`${BACKEND_URL}/api/v1/notes`, {
      headers: { Authorization: `Bearer ${tokens.access_token}` }
    });
    expect(notesResponse.ok()).toBe(true);
    const notesData = await notesResponse.json();
    const notes = Array.isArray(notesData.notes) ? notesData.notes : notesData;
    const createdNote = notes.find((n: any) => n.title === noteTitle);
    expect(createdNote).toBeTruthy();

    // Logout and verify redirect to login
    const logoutButton = page.locator('[data-testid="logout-button"], text=Logout, text=Выйти').first();
    const hasLogout = await logoutButton.isVisible().catch(() => false);
    if (hasLogout) {
      await logoutButton.click();
      await page.waitForURL(/.*auth\/login.*/, { timeout: 10000 });
      expect(page.url()).toContain('/auth/login');
    }
  });
});
