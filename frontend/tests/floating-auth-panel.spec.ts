import { test, expect } from "@playwright/test";

const FRONTEND_URL = process.env.FRONTEND_URL || "http://127.0.0.1:3002";

test.describe("Floating auth panel on public graph @auth-real", () => {
  test.beforeEach(async ({ page }) => {
    await page.goto(`${FRONTEND_URL}/?nocache=1`);
    await page.waitForSelector('[data-testid="graph-2d-container"]', {
      timeout: 30000,
    });
  });

  test("opens and closes the floating login panel", async ({ page }) => {
    const loginBtn = page.locator('[data-testid="floating-login-button"]');
    await expect(loginBtn).toBeVisible();

    await loginBtn.click();

    const panel = page.locator('[data-testid="floating-auth-panel"]');
    await expect(panel).toBeVisible();

    const loginTab = page.locator('[data-testid="floating-auth-tab-login"]');
    await expect(loginTab).toHaveAttribute("aria-selected", "true");

    const closeBtn = page.locator('[data-testid="floating-auth-close"]');
    await closeBtn.click();
    await expect(panel).not.toBeVisible();
  });

  test("switches between login and register tabs", async ({ page }) => {
    await page.locator('[data-testid="floating-login-button"]').click();

    const panel = page.locator('[data-testid="floating-auth-panel"]');
    await expect(panel).toBeVisible();

    const registerTab = page.locator(
      '[data-testid="floating-auth-tab-register"]'
    );
    await registerTab.click();
    await expect(registerTab).toHaveAttribute("aria-selected", "true");

    const loginTab = page.locator('[data-testid="floating-auth-tab-login"]');
    await loginTab.click();
    await expect(loginTab).toHaveAttribute("aria-selected", "true");

    await page.locator('[data-testid="floating-auth-close"]').click();
    await expect(panel).not.toBeVisible();
  });

  test("opens the panel on the register tab from the register button", async ({
    page,
  }) => {
    const registerBtn = page.locator(
      '[data-testid="floating-register-button"]'
    );
    await expect(registerBtn).toBeVisible();

    await registerBtn.click();

    const panel = page.locator('[data-testid="floating-auth-panel"]');
    await expect(panel).toBeVisible();

    const registerTab = page.locator(
      '[data-testid="floating-auth-tab-register"]'
    );
    await expect(registerTab).toHaveAttribute("aria-selected", "true");
  });

  test("logs in through the floating panel and reloads the graph", async ({
    page,
  }) => {
    // Capture the public note count before login.
    const countBefore = await page
      .locator('[data-testid="filter-count-all"]')
      .textContent();

    await page.locator('[data-testid="floating-login-button"]').click();

    const panel = page.locator('[data-testid="floating-auth-panel"]');
    await expect(panel).toBeVisible();

    await panel.locator('input[name="login"]').fill("testuser");
    await panel.locator('input[name="password"]').fill("TestPassword123!");

    // Submit the login form inside the panel.
    await panel.locator('button[type="submit"]').click();

    // Panel should close.
    await expect(panel).not.toBeVisible({ timeout: 10000 });

    // Authenticated UI should appear (sidebar shows user info and log out).
    await expect(page.locator('.sidebar-footer .user-info')).toBeVisible({
      timeout: 15000,
    });
    await expect(
      page.locator('.sidebar-footer a:has-text("↩")')
    ).toBeVisible();

    // The graph should eventually reload and the note count should change.
    await expect
      .poll(
        async () => page.locator('[data-testid="filter-count-all"]').textContent(),
        { timeout: 20000 }
      )
      .not.toBe(countBefore);
  });

  test("drags the floating panel across the viewport", async ({ page }) => {
    await page.locator('[data-testid="floating-login-button"]').click();

    const panel = page.locator('[data-testid="floating-auth-panel"]');
    await expect(panel).toBeVisible();

    const handle = page.locator('[data-testid="floating-auth-drag-handle"]');
    const box = await panel.boundingBox();
    expect(box).not.toBeNull();

    const startX = box!.x + box!.width / 2;
    const startY = box!.y + 20;

    // Drag 120px left and 80px down.
    await handle.hover();
    await page.mouse.down();
    await page.mouse.move(startX - 120, startY + 80);
    await page.mouse.up();

    const newBox = await panel.boundingBox();
    expect(newBox).not.toBeNull();
    expect(Math.abs(newBox!.x - (box!.x - 120))).toBeLessThan(20);
    expect(Math.abs(newBox!.y - (box!.y + 80))).toBeLessThan(20);
  });
});
