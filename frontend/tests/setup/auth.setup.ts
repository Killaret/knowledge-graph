import { test as setup, expect } from "@playwright/test";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const TEST_USER = {
  login: "testuser",
  password: "TestPassword123!",
};

// Resolve relative to this file, not process.cwd(): launching Playwright from
// the repo root would otherwise write the state outside frontend/.
const STORAGE_STATE = resolve(dirname(fileURLToPath(import.meta.url)), ".auth/testuser.json");

/**
 * Real auth setup for visual regression.
 *
 * Logs in as the seeded `testuser`, waits for the home page, and persists the
 * resulting cookies (HttpOnly refresh token) and session hint to a storage
 * state file. The `visual-real-auth` project then uses this state so tests
 * enter the app already authenticated.
 *
 * Note: the saved state is one-shot. The backend rotates the refresh token,
 * so a reused file yields `refresh 401` and the app silently falls back to
 * the anonymous view. Always rerun this setup project (`visual-real-auth`
 * depends on it) instead of reusing a stale file.
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
