import { type Page, type APIRequestContext, expect } from "@playwright/test";

export const BACKEND_URL = process.env.BACKEND_URL || "http://localhost:8083";

export const TEST_USER = {
  login: "testuser",
  password: "TestPassword123!",
};

export async function loginAsTestUser(page: Page, request: APIRequestContext) {
  // When the backend runs in SKIP_AUTH mode the frontend should bypass the
  // real login/refresh flow entirely. initAuth will create a fake user and
  // tokens when it sees __SKIP_AUTH__ === true.
  if (process.env.SKIP_AUTH === "true") {
    await page.addInitScript(() => {
      (window as any).__SKIP_AUTH__ = true;
    });
    return;
  }

  const response = await request.post(`${BACKEND_URL}/api/v1/auth/login`, {
    data: TEST_USER,
    headers: { "Content-Type": "application/json" },
  });
  expect(response.ok()).toBe(true);
  const { access_token, refresh_token } = (await response.json()) as {
    access_token: string;
    refresh_token: string;
  };

  await page.addInitScript(
    (tokens: { access: string; refresh: string }) => {
      // Only seed tokens on first load; after a refresh the auth store may have
      // rotated to a newer refresh token, and overwriting it would revoke the
      // active session.
      if (!localStorage.getItem("refresh_token")) {
        localStorage.setItem("access_token", tokens.access);
        localStorage.setItem("refresh_token", tokens.refresh);
      }
      (window as any).__SKIP_AUTH__ = false;
    },
    { access: access_token, refresh: refresh_token },
  );
}

export async function getAuthToken(
  request: APIRequestContext,
): Promise<string> {
  const response = await request.post(`${BACKEND_URL}/api/v1/auth/login`, {
    data: TEST_USER,
    headers: { "Content-Type": "application/json" },
  });
  expect(response.ok()).toBe(true);
  const { access_token } = (await response.json()) as { access_token: string };
  return access_token;
}
