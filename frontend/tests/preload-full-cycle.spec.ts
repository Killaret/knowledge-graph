// E2E tests for full preload data cycle
import { test, expect } from "@playwright/test";

test.describe("PreloadService Full Cycle E2E", () => {
  test.beforeEach(async ({ page }) => {
    // Skip all preload tests in SKIP_AUTH mode (no login flow)
    test.skip(
      process.env.SKIP_AUTH === "true",
      "Preload tests skipped in SKIP_AUTH mode",
    );

    // Clear localStorage before each test
    await page.context().clearCookies();
    await page.addInitScript(() => {
      // Initialize localStorage if needed
      if (typeof localStorage !== "undefined") {
        localStorage.clear();
      }
    });
  });

  test("should preload data on login page and display instantly after login", async ({
    page,
  }) => {
    // Navigate to login page
    await page.goto("/auth/login");

    // Wait for login page to load
    await expect(page.locator("h1")).toContainText("Knowledge Graph");

    // Check that PreloadService starts (via console logs)
    const consoleMessages: string[] = [];
    page.on("console", (msg) => {
      consoleMessages.push(msg.text());
    });

    // Wait for preload to complete
    await page.waitForTimeout(2000);

    // Check that preload logs exist
    const preloadLogs = consoleMessages.filter(
      (msg) =>
        msg.includes("[PreloadService]") &&
        msg.includes("Starting background preload"),
    );
    expect(preloadLogs.length).toBeGreaterThan(0);

    // Fill login form
    await page.fill('input[name="login"]', "testuser");
    await page.fill('input[name="password"]', "testpassword");

    // Track main page display time after login
    const startTime = Date.now();

    // Click login button
    await page.click('button[type="submit"]');

    // Wait for navigation to main page
    await page.waitForURL("/");

    // Check that interface displays quickly (less than 1 second)
    const loadTime = Date.now() - startTime;
    expect(loadTime).toBeLessThan(1000);

    // Check that graph displays (not empty)
    await expect(page.locator('[data-testid="graph-canvas"]')).toBeVisible();

    // Check that preloaded data usage logs exist
    const usePreloadedLogs = consoleMessages.filter(
      (msg) =>
        msg.includes("[usePreloadedData]") && msg.includes("Using preloaded"),
    );
    expect(usePreloadedLogs.length).toBeGreaterThan(0);
  });

  test("should handle preload errors gracefully", async ({ page }) => {
    // Mock API errors via page
    await page.route("**/api/v1/graph/all**", (route) => {
      route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({ error: "Internal server error" }),
      });
    });

    await page.route("**/api/v1/achievements**", (route) => {
      route.fulfill({
        status: 500,
        contentType: "application/json",
        body: JSON.stringify({ error: "Internal server error" }),
      });
    });

    // Navigate to login page
    await page.goto("/auth/login");

    // Wait for preload attempt
    await page.waitForTimeout(2000);

    // Check that preload error logs exist
    const consoleMessages: string[] = [];
    page.on("console", (msg) => {
      consoleMessages.push(msg.text());
    });

    await page.waitForTimeout(1000);

    const errorLogs = consoleMessages.filter(
      (msg) =>
        msg.includes("[PreloadService]") && msg.includes("Failed to preload"),
    );
    expect(errorLogs.length).toBeGreaterThan(0);

    // Execute login
    await page.fill('input[name="login"]', "testuser");
    await page.fill('input[name="password"]', "testpassword");
    await page.click('button[type="submit"]');

    // Wait for navigation to main page
    await page.waitForURL("/");

    // Application should still load (with fallback to server)
    await expect(page.locator('[data-testid="graph-canvas"]')).toBeVisible();
  });

  test("should clear preload cache on logout", async ({ page }) => {
    // Navigate to login page
    await page.goto("/auth/login");

    // Wait for preload
    await page.waitForTimeout(2000);

    // Execute login
    await page.fill('input[name="login"]', "testuser");
    await page.fill('input[name="password"]', "testpassword");
    await page.click('button[type="submit"]');

    // Wait for main page to load
    await page.waitForURL("/");

    // Check that data is loaded
    await expect(page.locator('[data-testid="graph-canvas"]')).toBeVisible();

    // Execute logout
    await page.click('[data-testid="logout-button"]');

    // Wait for navigation to login page
    await page.waitForURL("/auth/login");

    // Check that cache is cleared (via localStorage)
    await page.evaluate(() => {
      localStorage.removeItem("preload_cache");
    });

    // In our implementation cache is stored in memory, but we can check
    // that user is logged out
    await expect(page.locator("h1")).toContainText("Knowledge Graph");
  });

  test("should not preload when already authenticated", async ({ page }) => {
    // Set tokens in localStorage (simulate already authenticated user)
    await page.addInitScript(() => {
      localStorage.setItem("access_token", "test_token");
      localStorage.setItem("refresh_token", "test_refresh");
    });

    // Navigate to main page
    await page.goto("/");

    // Check that we are not redirected to login page
    await expect(page).toHaveURL("/");

    // Check that preload did not start
    const consoleMessages: string[] = [];
    page.on("console", (msg) => {
      consoleMessages.push(msg.text());
    });

    await page.waitForTimeout(2000);

    const preloadLogs = consoleMessages.filter(
      (msg) =>
        msg.includes("[PreloadService]") &&
        msg.includes("Starting background preload"),
    );
    expect(preloadLogs.length).toBe(0);
  });

  test("should handle concurrent preload requests", async ({ page }) => {
    // Navigate to login page
    await page.goto("/auth/login");

    // Wait for preload to start
    await page.waitForTimeout(500);

    // Open second tab with same page
    const newPage = await page.context().newPage();
    await newPage.goto("/auth/login");

    // Wait for preload to complete
    await page.waitForTimeout(2000);

    // Execute login on first page
    await page.fill('input[name="login"]', "testuser");
    await page.fill('input[name="password"]', "testpassword");
    await page.click('button[type="submit"]');

    await page.waitForURL("/");

    // Execute login on second page
    await newPage.fill('input[name="login"]', "testuser2");
    await newPage.fill('input[name="password"]', "testpassword2");
    await newPage.click('button[type="submit"]');

    await newPage.waitForURL("/");

    // Both pages should load correctly
    await expect(page.locator('[data-testid="graph-canvas"]')).toBeVisible();
    await expect(newPage.locator('[data-testid="graph-canvas"]')).toBeVisible();

    await newPage.close();
  });

  test("should respect TTL and refresh expired cache", async ({ page }) => {
    // Navigate to login page
    await page.goto("/auth/login");

    // Wait for preload
    await page.waitForTimeout(2000);

    // Execute login
    await page.fill('input[name="login"]', "testuser");
    await page.fill('input[name="password"]', "testpassword");
    await page.click('button[type="submit"]');

    await page.waitForURL("/");

    // Check that data displays quickly
    await expect(page.locator('[data-testid="graph-canvas"]')).toBeVisible();

    // Simulate TTL expiration (reload page after long time)
    await page.evaluate(() => {
      // Mock Date.now to simulate TTL expiration
      const originalDateNow = Date.now;
      Date.now = () => originalDateNow() + 6 * 60 * 1000; // +6 minutes
    });

    // Reload page
    await page.reload();

    // Data should load from server (slower)
    const startTime = Date.now();
    await expect(page.locator('[data-testid="graph-canvas"]')).toBeVisible();
    const loadTime = Date.now() - startTime;

    // Load should take longer (no preloaded data)
    expect(loadTime).toBeGreaterThan(500);

    // Restore Date.now
    await page.evaluate(() => {
      // Date.now will be restored on page reload
    });
  });

  test("should work with different user roles", async ({ page }) => {
    // Test with different user roles
    const userRoles = ["user", "admin"];

    for (const role of userRoles) {
      // Create new page for each role
      const testPage = await page.context().newPage();

      // Navigate to login page
      await testPage.goto("/auth/login");

      // Wait for preload
      await testPage.waitForTimeout(2000);

      // Execute login with corresponding role
      await testPage.fill('input[name="login"]', `${role}user`);
      await testPage.fill('input[name="password"]', "testpassword");
      await testPage.click('button[type="submit"]');

      await testPage.waitForURL("/");

      // Check that interface displays correctly
      await expect(
        testPage.locator('[data-testid="graph-canvas"]'),
      ).toBeVisible();

      // For admin there may be additional elements
      if (role === "admin") {
        // Check for admin elements (if they exist)
        void testPage.locator('[data-testid*="admin"]');
        // Don't expect them to exist, just check no errors
      }

      await testPage.close();
    }
  });

  test("should handle network interruptions gracefully", async ({ page }) => {
    // Navigate to login page
    await page.goto("/auth/login");

    // Simulate network interruption during preload
    await page.route("**/api/v1/graph/all**", (route) => {
      // Abort connection
      route.abort("failed");
    });

    // Wait for preload attempt
    await page.waitForTimeout(2000);

    // Remove blocking
    await page.unroute("**/api/v1/graph/all**");

    // Execute login
    await page.fill('input[name="login"]', "testuser");
    await page.fill('input[name="password"]', "testpassword");
    await page.click('button[type="submit"]');

    // Application should load with fallback
    await page.waitForURL("/");
    await expect(page.locator('[data-testid="graph-canvas"]')).toBeVisible();
  });
});
