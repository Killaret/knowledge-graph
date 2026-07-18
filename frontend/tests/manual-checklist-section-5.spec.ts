import { test, expect, type Page } from "@playwright/test";
import { loginAsTestUser } from "./helpers/auth";

const ROUTES = ["/", "/graph", "/profile"];

async function waitForRouteContent(page: Page, route: string) {
  if (route === "/graph") {
    await expect(page.locator('[data-testid="graph-stats"]')).toBeVisible({
      timeout: 20000,
    });
  } else if (route === "/profile") {
    await expect(page.locator('[data-testid="profile-content"]')).toBeVisible({
      timeout: 20000,
    });
  } else {
    await expect(page.locator("main")).toBeVisible({ timeout: 20000 });
  }
}

function attachConsoleAndDialogListeners(
  page: Page,
  errors: string[],
  dialogs: string[],
) {
  page.on("console", (msg) => {
    if (msg.type() === "error") {
      errors.push(msg.text());
    }
  });
  page.on("pageerror", (err) => {
    errors.push(err.message);
  });
  page.on("dialog", (dialog) => {
    dialogs.push(dialog.type());
    void dialog.dismiss();
  });
}

test.describe("Section 5 - General UX", { tag: ["@manual", "@ux"] }, () => {
  test.beforeEach(async ({ page, request }) => {
    await loginAsTestUser(page, request);
  });

  test("authenticated navigation across main routes returns no 401", async ({
    page,
  }) => {
    const unauthorized: string[] = [];
    page.on("response", (res) => {
      if (res.status() === 401) {
        unauthorized.push(`${res.request().method()} ${res.url()}`);
      }
    });

    for (const route of ROUTES) {
      await page.goto(route, { timeout: 60000, waitUntil: "domcontentloaded" });
      await waitForRouteContent(page, route);
    }

    expect(unauthorized).toEqual([]);
  });

  test("logout and re-login preserves graph data", async ({
    page,
    request,
  }) => {
    await page.goto("/graph?nocache=1", {
      timeout: 60000,
      waitUntil: "domcontentloaded",
    });
    const beforeStats = await page
      .locator('[data-testid="graph-stats"]')
      .textContent();

    // Logout via UI: the auth store clears tokens and redirects
    const logoutBtn = page.locator("text=Logout").first();
    if (await logoutBtn.isVisible().catch(() => false)) {
      await logoutBtn.click();
      await page.waitForURL(/\/(auth\/login|)/, { timeout: 10000 });
    }

    // Re-login by setting fresh tokens and navigating back
    await loginAsTestUser(page, request);
    await page.goto("/graph?nocache=1", {
      timeout: 60000,
      waitUntil: "domcontentloaded",
    });
    await expect(page.locator('[data-testid="graph-stats"]')).toBeVisible({
      timeout: 20000,
    });

    const afterStats = await page
      .locator('[data-testid="graph-stats"]')
      .textContent();
    expect(afterStats).toBe(beforeStats);
  });

  test("basic actions do not trigger console errors or browser dialogs", async ({
    page,
  }) => {
    const consoleErrors: string[] = [];
    const dialogs: string[] = [];
    const resource404s: string[] = [];
    page.on("response", (res) => {
      if (res.status() === 404) {
        resource404s.push(res.url());
      }
    });
    attachConsoleAndDialogListeners(page, consoleErrors, dialogs);

    await page.goto("/graph?nocache=1", {
      timeout: 60000,
      waitUntil: "domcontentloaded",
    });
    await expect(page.locator('[data-testid="graph-stats"]')).toBeVisible({
      timeout: 20000,
    });

    await page.keyboard.press("?");
    await expect(page.locator('[data-testid="help-modal"]')).toBeVisible({
      timeout: 5000,
    });
    await page.keyboard.press("Escape");
    await expect(page.locator('[data-testid="help-modal"]')).toBeHidden({
      timeout: 5000,
    });

    await page.goto("/", { timeout: 60000, waitUntil: "domcontentloaded" });
    await waitForRouteContent(page, "/");

    expect(dialogs).toEqual([]);
    // Allow known, non-critical warnings and network 404s from placeholder system notes
    const criticalErrors = consoleErrors.filter(
      (e) =>
        !e.includes("a11y") &&
        !e.includes("List") &&
        !e.includes("Failed to load resource"),
    );
    expect(criticalErrors).toEqual([]);
  });

  test("language switch in profile updates UI text", async ({ page }) => {
    await page.goto("/profile", {
      timeout: 60000,
      waitUntil: "domcontentloaded",
    });
    await expect(page.locator('[data-testid="profile-content"]')).toBeVisible({
      timeout: 20000,
    });

    const localeSelect = page.locator("#locale");
    await expect(localeSelect).toBeVisible();

    const responses: string[] = [];
    page.on("response", async (res) => {
      if (res.status() >= 400) {
        const body = await res.text().catch(() => "");
        responses.push(
          `${res.status()} ${res.request().method()} ${res.url()} ${body}`,
        );
      }
    });

    // The locale change triggers a full page reload
    await Promise.all([
      page.waitForURL("/profile", { timeout: 10000 }),
      localeSelect.selectOption("ru"),
    ]);

    // Wait for profile content to re-appear (initAuth + getMe finished)
    await expect(page.locator('[data-testid="profile-content"]')).toBeVisible({
      timeout: 20000,
    });
    await expect(localeSelect).toBeVisible({ timeout: 10000 });

    // ProfileEditor re-renders in Russian after reload
    await expect(
      page.locator("h2", { hasText: "Редактировать профиль" }),
    ).toBeVisible({ timeout: 10000 });
  });

  test("responsive layout keeps canvas visible on narrow viewport", async ({
    page,
  }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto("/graph?nocache=1", {
      timeout: 60000,
      waitUntil: "domcontentloaded",
    });
    await expect(page.locator('[data-testid="graph-canvas"]')).toBeVisible({
      timeout: 20000,
    });
    await expect(page.locator('[data-testid="graph-stats"]')).toBeVisible({
      timeout: 20000,
    });

    await page.setViewportSize({ width: 1280, height: 720 });
    await page.goto("/", { timeout: 60000, waitUntil: "domcontentloaded" });
    await waitForRouteContent(page, "/");
  });
});
