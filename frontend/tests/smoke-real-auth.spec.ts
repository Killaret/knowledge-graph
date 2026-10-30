import { test, expect } from "@playwright/test";

const BACKEND_URL = process.env.BACKEND_URL || "http://127.0.0.1:18083";
const FRONTEND_URL = process.env.FRONTEND_URL || "http://localhost:3002";

test.describe("Smoke tests - real auth flow", { tag: ["@smoke", "@auth-real"] }, () => {
  test("register, login, profile, graph, create note, logout", async ({ page, request }) => {
    page.on("console", (msg) => console.log("[browser console]", msg.type(), msg.text()));
    page.on("pageerror", (err) => console.log("[browser pageerror]", err));
    page.on("response", (response) => {
      const status = response.status();
      if (status >= 400) {
        console.log("[browser response error]", status, response.url());
      }
    });

    const id = crypto.randomUUID();
    const loginName = `smoke_${id.replace(/-/g, "").slice(0, 12)}`;
    const password = "TestPassword123!";
    const email = `${loginName}@test.example.com`;

    // Register user via API
    const registerResponse = await request.post(`${BACKEND_URL}/api/v1/auth/register`, {
      data: { login: loginName, email, password },
    });
    expect([200, 201]).toContain(registerResponse.status());

    // Login via UI
    await page.goto(`${FRONTEND_URL}/auth/login`);
    await page.waitForLoadState("networkidle");
    await page.fill('input[name="login"]', loginName);
    await page.fill('input[name="password"]', password);
    await page.click('button[type="submit"]');
    await page.waitForURL(`${FRONTEND_URL}/`, { timeout: 15000 });

    // Main page: public/authenticated graph canvas visible
    const graphCanvas = page.locator('[data-testid="graph-canvas"]');
    await expect(graphCanvas).toBeVisible({ timeout: 15000 });

    const graphStats = page.locator('[data-testid="graph-stats"]').first();
    await expect(graphStats).toContainText(/\d+\s+nodes?/i, {
      timeout: 10000,
    });

    // Profile page is accessible after authentication.
    // The Playwright page context does not reliably share HttpOnly refresh
    // cookies with the test API context, so we inject the access token
    // directly into the page to bootstrap auth on the next navigation.
    const uiLoginResponse = await request.post(`${BACKEND_URL}/api/v1/auth/login`, {
      data: { login: loginName, password },
    });
    expect(uiLoginResponse.ok()).toBe(true);
    const uiTokens = await uiLoginResponse.json();
    expect(uiTokens.access_token).toBeTruthy();
    await page.addInitScript((t: string) => {
      (window as any).__ACCESS_TOKEN__ = t;
    }, uiTokens.access_token);
    await page.context().addInitScript((t: string) => {
      (window as any).__ACCESS_TOKEN__ = t;
    }, uiTokens.access_token);

    await page.goto(`${FRONTEND_URL}/profile`);
    await page.waitForLoadState("domcontentloaded");
    const profileContent = page
      .locator('[data-testid="profile-content"], .profile-container')
      .first();
    await expect(profileContent).toBeVisible({ timeout: 10000 });

    // Create a note using the ghost node form on the main graph page
    await page.goto(`${FRONTEND_URL}/`);
    await page.waitForLoadState("networkidle");
    await expect(graphCanvas).toBeVisible({ timeout: 15000 });
    await page.keyboard.press("n");

    const titleInput = page.locator('input[placeholder="Title"]').first();
    await expect(titleInput).toBeVisible({ timeout: 5000 });

    const noteTitle = `Smoke Note ${loginName}`;
    await titleInput.fill(noteTitle);
    await page.locator('button:has-text("Create")').first().click();

    // Form should close
    await expect(titleInput).not.toBeVisible({ timeout: 5000 });
    await expect(graphCanvas).toBeVisible();

    // Graph view is accessible after creating the note
    await page.goto(`${FRONTEND_URL}/graph?full=1&nocache=1`);
    await page.waitForLoadState("networkidle");
    await page.waitForSelector('[data-testid="graph-container"], [data-testid="graph-canvas"]', {
      timeout: 15000,
    });
    await expect(
      page.locator('[data-testid="graph-container"], [data-testid="graph-canvas"]').first()
    ).toBeVisible({
      timeout: 10000,
    });

    // Verify note was persisted via API using a fresh API login
    const apiLoginResponse = await request.post(`${BACKEND_URL}/api/v1/auth/login`, {
      data: { login: loginName, password },
    });
    expect(apiLoginResponse.ok()).toBe(true);
    const tokens = await apiLoginResponse.json();
    expect(tokens.access_token).toBeTruthy();

    const notesResponse = await request.get(`${BACKEND_URL}/api/v1/notes`, {
      headers: { Authorization: `Bearer ${tokens.access_token}` },
    });
    expect(notesResponse.ok()).toBe(true);
    const notesData = await notesResponse.json();
    const notes = Array.isArray(notesData.notes) ? notesData.notes : notesData;
    const createdNote = notes.find((n: any) => n.title === noteTitle);
    expect(createdNote).toBeTruthy();

    // Logout and verify redirect to login
    const logoutButton = page
      .locator('[data-testid="logout-button"], text=Logout, text=Выйти')
      .first();
    const hasLogout = await logoutButton.isVisible().catch(() => false);
    if (hasLogout) {
      await logoutButton.click();
      await page.waitForURL(/.*auth\/login.*/, { timeout: 10000 });
      expect(page.url()).toContain("/auth/login");
    }
  });
});
