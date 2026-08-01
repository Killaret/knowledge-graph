import { test, expect } from "@playwright/test";
import { getBackendUrl, loginOrCreateTestUser } from "./helpers/auth";

/**
 * Regression tests for anonymous/public graph behavior in real-auth mode.
 *
 * Historically the public graph caused two related bugs:
 * 1. docs/MANUAL_TEST_ISSUES.md #6: "Пустой публичный граф" — seed data and
 *    user notes were private by default, so anonymous visitors saw an empty
 *    graph even though the pipeline was fine.
 * 2. The main page polled /v1/graph/delta every 30s and on window focus even
 *    for anonymous users. /v1/graph/delta requires auth, so it returned 401,
 *    the ky afterResponse hook tried /v1/auth/refresh (400), and the resulting
 *    cascade either re-rendered the graph or redirected the public page to
 *    /auth/login.
 */
test.describe("Public graph in real-auth mode @auth-real", () => {
  test("shows published notes to anonymous users even when the owner has private notes", async ({
    page,
    request,
  }) => {
    const backendUrl = getBackendUrl();
    const token = await loginOrCreateTestUser(request);
    const authHeaders = { Authorization: `Bearer ${token}` };
    const uniqueSuffix = Date.now();

    // A private note must never show up in the public/anonymous graph.
    const privateResp = await request.post(`${backendUrl}/api/v1/notes`, {
      headers: authHeaders,
      data: {
        title: `Private Regression Note ${uniqueSuffix}`,
        content: "This note must stay private",
        type: "star",
      },
    });
    expect(privateResp.ok(), await privateResp.text()).toBeTruthy();

    // A published note must be visible to anonymous users.
    const publicTitle = `Public Regression Note ${uniqueSuffix}`;
    const publicNoteResp = await request.post(`${backendUrl}/api/v1/notes`, {
      headers: authHeaders,
      data: {
        title: publicTitle,
        content: "This note should be publicly visible",
        type: "star",
      },
    });
    expect(publicNoteResp.ok(), await publicNoteResp.text()).toBeTruthy();
    const publicNote = await publicNoteResp.json();

    const publishResp = await request.post(
      `${backendUrl}/api/v1/notes/${publicNote.data.id}/publish`,
      { headers: authHeaders }
    );
    expect(publishResp.ok(), await publishResp.text()).toBeTruthy();

    // API-level assertion first: fast and unambiguous about what the
    // anonymous graph-service endpoint actually returns.
    const publicGraphResp = await request.get(
      `${backendUrl}/api/v1/graph/all?nocache=1`
    );
    expect(publicGraphResp.ok(), await publicGraphResp.text()).toBeTruthy();
    const publicGraph = await publicGraphResp.json();
    const publicNodeTitles: string[] = (publicGraph.data?.nodes ?? []).map(
      (n: { title: string }) => n.title
    );
    expect(publicNodeTitles.length).toBeGreaterThan(0);
    expect(publicNodeTitles).toContain(publicTitle);

    // UI-level assertion: a genuinely anonymous page (no access token
    // injected anywhere) must render at least one node.
    await page.goto("/graph?nocache=1&full=1", {
      timeout: 60000,
      waitUntil: "domcontentloaded",
    });
    await page.waitForLoadState("networkidle");

    const canvas = page.locator('[data-testid="graph-canvas"]');
    await expect(canvas).toBeVisible({ timeout: 20000 });

    const stats = page.locator('[data-testid="graph-stats"]');
    await expect(stats).toBeVisible({ timeout: 20000 });
    // Require a non-zero node count explicitly — a naive /\d+ nodes?/ regex
    // would incorrectly pass on "0 nodes".
    await expect(stats).toContainText(
      /[1-9]\d*\s*(?:nodes?|уз(?:лов|ел|ла|ьев)?)/i,
      {
        timeout: 20000,
      }
    );
  });

  test("anonymous /graph page does not poll delta or refresh", async ({
    page,
  }) => {
    const trackedUrls: string[] = [];
    page.on("request", (req) => {
      trackedUrls.push(req.url());
    });

    await page.goto("/graph?nocache=1&full=1", {
      timeout: 60000,
      waitUntil: "networkidle",
    });

    // The page should stay on the public graph and not redirect to login.
    expect(page.url()).toMatch(/\/graph/);

    const canvas = page.locator('[data-testid="graph-canvas"]');
    await expect(canvas).toBeVisible({ timeout: 20000 });

    // Trigger focus — the main page used to call refreshAfterMutation() on
    // focus, which for anonymous users produced /v1/graph/delta (401).
    await page.evaluate(() => window.dispatchEvent(new Event("focus")));
    await page.waitForTimeout(1500);

    // Wait long enough for any immediately-registered setInterval callbacks
    // to fire (the actual 30s interval won't fire, but we still assert
    // absence of delta/refresh after a realistic window).
    await page.waitForTimeout(2000);

    const deltaRequests = trackedUrls.filter((url) =>
      url.includes("/v1/graph/delta")
    );
    const refreshRequests = trackedUrls.filter((url) =>
      url.includes("/v1/auth/refresh")
    );

    expect(
      deltaRequests,
      "Anonymous public graph should not request /v1/graph/delta"
    ).toHaveLength(0);

    // Exactly one /v1/auth/refresh from initAuth is acceptable (the session
    // bootstrap tries to restore from a refresh cookie), but there must not be
    // the cascade of repeated 400s caused by 401s from delta/protected calls.
    expect(
      refreshRequests.length,
      "Anonymous public graph should not fire repeated /v1/auth/refresh"
    ).toBeLessThanOrEqual(1);

    // Main page sets this flag in onMount to expose whether it registered the
    // background sync interval.
    const pollingActive = await page.evaluate(
      () => (window as any).__kgGraphPollingActive
    );
    expect(pollingActive, "Background graph sync must be off for anonymous").toBe(false);
  });

  test("anonymous main page does not poll delta or refresh", async ({ page }) => {
    const trackedUrls: string[] = [];
    page.on("request", (req) => {
      trackedUrls.push(req.url());
    });

    await page.goto("/?nocache=1", {
      timeout: 60000,
      waitUntil: "networkidle",
    });

    // Main page should not redirect away for anonymous users (URL may keep
    // query params like ?nocache=1).
    expect(page.url()).not.toContain("/auth/login");

    const canvas = page.locator('[data-testid="graph-canvas"]');
    await expect(canvas).toBeVisible({ timeout: 20000 });

    // Trigger focus and wait for any immediate polling to fire.
    await page.evaluate(() => window.dispatchEvent(new Event("focus")));
    await page.waitForTimeout(1500);
    await page.waitForTimeout(2000);

    const deltaRequests = trackedUrls.filter((url) =>
      url.includes("/v1/graph/delta")
    );
    const refreshRequests = trackedUrls.filter((url) =>
      url.includes("/v1/auth/refresh")
    );

    expect(
      deltaRequests,
      "Anonymous main page should not request /v1/graph/delta"
    ).toHaveLength(0);
    expect(
      refreshRequests.length,
      "Anonymous main page should not fire repeated /v1/auth/refresh"
    ).toBeLessThanOrEqual(1);

    const pollingActive = await page.evaluate(
      () => (window as any).__kgGraphPollingActive
    );
    expect(pollingActive, "Background graph sync must be off for anonymous").toBe(false);
  });
});
