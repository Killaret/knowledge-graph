import { describe, it, expect, beforeEach, afterEach } from "vitest";
import {
  authState,
  currentUser,
  accessToken,
  refreshToken,
  isInitialized,
  isLoading,
  error,
  apiKey,
  getApiKey,
  setApiKey,
  skipAuthMode,
  isAuthenticated,
  isAdmin,
  saveTokens,
  clearAuthState,
} from "./auth-session.svelte";
import type { User, AuthTokens } from "$shared/types";

describe("auth-session store", () => {
  beforeEach(() => {
    clearAuthState();
    authState.isInitialized = false;
    authState.isLoading = false;
    authState.error = null;
    localStorage.clear();
    delete (window as { __SKIP_AUTH__?: boolean }).__SKIP_AUTH__;
  });

  afterEach(() => {
    delete (window as { __SKIP_AUTH__?: boolean }).__SKIP_AUTH__;
  });

  it("getters reflect initial state", () => {
    expect(currentUser()).toBeNull();
    expect(accessToken()).toBeNull();
    expect(refreshToken()).toBeNull();
    expect(isInitialized()).toBe(false);
    expect(isLoading()).toBe(false);
    expect(error()).toBeNull();
    expect(apiKey()).toBeNull();
    expect(getApiKey()).toBeNull();
    expect(isAuthenticated()).toBe(false);
    expect(isAdmin()).toBe(false);
  });

  it("saveTokens updates token getters", () => {
    const tokens: AuthTokens = {
      access_token: "acc",
      refresh_token: "ref",
      token_type: "Bearer",
      expires_at: "2024-12-31T23:59:59Z",
    };
    saveTokens(tokens);
    expect(accessToken()).toBe("acc");
    expect(refreshToken()).toBe("ref");
  });

  it("setApiKey persists the key and clearApiKey removes it", () => {
    setApiKey("secret-key");
    expect(apiKey()).toBe("secret-key");
    expect(localStorage.getItem("api_key")).toBe("secret-key");

    setApiKey(null);
    expect(apiKey()).toBeNull();
    expect(localStorage.getItem("api_key")).toBeNull();
  });

  it("clearAuthState resets all auth fields", () => {
    authState.currentUser = { id: "u1" } as User;
    authState.accessToken = "acc";
    authState.apiKey = "key";
    authState.error = "boom";

    clearAuthState();

    expect(currentUser()).toBeNull();
    expect(accessToken()).toBeNull();
    expect(apiKey()).toBeNull();
    expect(error()).toBeNull();
  });

  it("skipAuthMode respects window flag", () => {
    expect(skipAuthMode()).toBe(false);
    (window as { __SKIP_AUTH__?: boolean }).__SKIP_AUTH__ = true;
    expect(skipAuthMode()).toBe(true);
  });

  it("isAuthenticated returns true for user, admin key or API key", () => {
    expect(isAuthenticated()).toBe(false);

    authState.currentUser = { id: "u1", role: "user" } as User;
    expect(isAuthenticated()).toBe(true);
    expect(isAdmin()).toBe(false);

    authState.currentUser = { id: "u1", role: "admin" } as User;
    expect(isAdmin()).toBe(true);

    authState.currentUser = null;
    authState.apiKey = "key";
    expect(isAuthenticated()).toBe(true);
  });
});
