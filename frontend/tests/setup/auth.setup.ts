import { test as setup, expect } from "@playwright/test";

const TEST_USER = {
  login: "testuser",
  password: "TestPassword123!",
};

const STORAGE_STATE = "tests/setup/.auth/testuser.json";

/**
 * Real auth setup for visual regression.
 *
 * Logs in as the seeded `testuser`, waits for the home page, and persists the
 * resulting cookies (HttpOnly refresh token) and session hint to a storage
 * state file. The `visual-real-auth` project then uses this state so tests
 * enter the app already authenticated.
 */
setup("authenticate as testuser", async ({ page }) => {
  await page.goto("/auth/login");
  await page.waitForLoadState("networkidle");

  await page.fill('input[name="login"]', TEST_USER.login);
  await page.fill('input[name="password"]', TEST_USER.password);
  await page.click('button[type="submit"]');

  // After a successful login the user is redirected to the home page.
  await page.waitForURL("/", { timeout: 15000 });
  await expect(page.locator("main")).toBeVisible({ timeout: 15000 });

  await page.context().storageState({ path: STORAGE_STATE });
});
