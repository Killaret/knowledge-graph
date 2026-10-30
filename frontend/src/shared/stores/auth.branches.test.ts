import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";

const mockEnv = vi.hoisted(() => ({ browser: true }));

vi.mock("$app/environment", () => mockEnv);
vi.mock("$app/navigation", () => ({ goto: vi.fn() }));

vi.mock("$shared/utils/i18n", () => ({
  setLocale: vi.fn(),
}));

vi.mock("$shared/services/PreloadService", () => ({
  clearPreloadCache: vi.fn(),
  preloadAuthenticatedGraph: vi.fn(),
}));

vi.mock("$shared/api/auth", () => ({
  login: vi.fn(),
  register: vi.fn(),
  logout: vi.fn(),
  refreshTokens: vi.fn(),
  handleYandexCallback: vi.fn(),
}));

vi.mock("$shared/api/users", () => ({
  getMe: vi.fn(),
  getSettings: vi.fn(),
  getAllAchievements: vi.fn(),
}));

import * as authApi from "$shared/api/auth";
import * as usersApi from "$shared/api/users";

async function importAuth() {
  vi.resetModules();
  const mod = await import("./auth.svelte");
  return mod;
}

async function importAuthSession() {
  vi.resetModules();
  const mod = await import("./auth-session.svelte");
  return mod;
}

describe("auth.svelte branch coverage", () => {
  beforeEach(() => {
    mockEnv.browser = true;
    localStorage.clear();
    vi.clearAllMocks();
    vi.mocked(usersApi.getMe).mockResolvedValue({
      id: "u1",
      login: "user",
      email: "user@example.com",
      role: "user",
      created_at: "2024-01-01T00:00:00Z",
    });
    vi.mocked(usersApi.getSettings).mockResolvedValue({ settings: [] });
    vi.mocked(authApi.refreshTokens).mockResolvedValue({
      access_token: "ref",
      refresh_token: "refres",
      token_type: "Bearer",
      expires_at: "2024-12-31T23:59:59Z",
    });
    delete (window as { __ACCESS_TOKEN__?: string }).__ACCESS_TOKEN__;
    delete (window as { __SKIP_AUTH__?: boolean }).__SKIP_AUTH__;
    delete (import.meta.env as any).VITE_SKIP_AUTH;
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("initAuth returns immediately on server", async () => {
    mockEnv.browser = false;
    const { initAuth, isInitialized } = await importAuth();

    await initAuth();
    expect(isInitialized()).toBe(true);
  });

  it("initAuth sets a skip-auth user when SKIP_AUTH is enabled", async () => {
    (window as { __SKIP_AUTH__?: boolean }).__SKIP_AUTH__ = true;
    const { initAuth, isAuthenticated, currentUser } = await importAuth();

    await initAuth();
    expect(isAuthenticated()).toBe(true);
    expect(currentUser()?.login).toBe("testuser");

    delete (window as { __SKIP_AUTH__?: boolean }).__SKIP_AUTH__;
  });

  it("initAuth returns the same promise when already in flight", async () => {
    localStorage.setItem("kg_auth_session", "1");
    vi.mocked(authApi.refreshTokens).mockReturnValue(new Promise(() => {}));
    const { initAuth } = await importAuth();

    void initAuth();
    void initAuth();

    expect(authApi.refreshTokens).toHaveBeenCalledTimes(1);
  });

  it("initAuth returns early when already initialized with a user", async () => {
    const { initAuth, isAuthenticated } = await importAuth();
    const { authState } = await import("./auth-session.svelte");

    authState.isInitialized = true;
    authState.currentUser = { id: "u1", login: "user", role: "user" } as any;

    await initAuth();
    expect(isAuthenticated()).toBe(true);
    expect(authApi.refreshTokens).not.toHaveBeenCalled();
  });

  it("initAuth reads an injected access token", async () => {
    (window as { __ACCESS_TOKEN__?: string }).__ACCESS_TOKEN__ = "injected_token";
    const { initAuth, isAuthenticated, accessToken } = await importAuth();

    await initAuth();
    expect(accessToken()).toBe("injected_token");
    expect(isAuthenticated()).toBe(true);
  });

  it("initAuth validates an API key from localStorage", async () => {
    localStorage.setItem("api_key", "my-key");
    const { initAuth, isAuthenticated, getApiKey } = await importAuth();

    await initAuth();
    expect(getApiKey()).toBe("my-key");
    expect(isAuthenticated()).toBe(true);
  });

  it("initAuth clears state when the API key is rejected", async () => {
    localStorage.setItem("api_key", "bad-key");
    vi.mocked(usersApi.getMe).mockRejectedValue(new Error("nope"));
    const { initAuth, isAuthenticated, getApiKey } = await importAuth();

    await initAuth();
    expect(getApiKey()).toBeNull();
    expect(isAuthenticated()).toBe(false);
  });

  it("initAuth logs a warning in DEV when settings fail", async () => {
    (import.meta.env as any).DEV = true;
    localStorage.setItem("api_key", "key");
    vi.mocked(usersApi.getSettings).mockRejectedValue(new Error("settings boom"));
    const consoleWarn = vi.spyOn(console, "warn").mockImplementation(() => {});

    const { initAuth } = await importAuth();
    await initAuth();

    expect(consoleWarn).toHaveBeenCalledWith("Failed to load user settings:", expect.any(Error));
  });

  it("initAuth logs an error in DEV when refresh fails", async () => {
    (import.meta.env as any).DEV = true;
    localStorage.setItem("kg_auth_session", "1");
    vi.mocked(authApi.refreshTokens).mockRejectedValue(new Error("refresh boom"));
    const consoleError = vi.spyOn(console, "error").mockImplementation(() => {});

    const { initAuth } = await importAuth();
    await initAuth();

    expect(consoleError).toHaveBeenCalledWith("Failed to initialize auth:", expect.any(Error));
  });

  it("initAuth falls back to existing token when refresh succeeds", async () => {
    localStorage.setItem("kg_auth_session", "1");
    const { initAuth, isAuthenticated } = await importAuth();

    await initAuth();
    expect(isAuthenticated()).toBe(true);
    expect(authApi.refreshTokens).toHaveBeenCalled();
  });
});
