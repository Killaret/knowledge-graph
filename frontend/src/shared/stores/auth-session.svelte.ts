// Low-level auth session state. This module must not import the API client
// or PreloadService to avoid circular dependencies.
import { browser } from "$app/environment";
import type { User, AuthTokens } from "$shared/types";

const API_KEY = "api_key";

// Global reactive state - wrapped in object for export
export const authState = $state({
  currentUser: null as User | null,
  accessToken: null as string | null,
  refreshToken: null as string | null,
  isInitialized: false,
  isLoading: false,
  error: null as string | null,
  apiKey: null as string | null,
});

// Export reactive state through getter functions
export function currentUser(): User | null {
  return authState.currentUser;
}
export function accessToken(): string | null {
  return authState.accessToken;
}
export function refreshToken(): string | null {
  return authState.refreshToken;
}
export function isInitialized(): boolean {
  return authState.isInitialized;
}
export function isLoading(): boolean {
  return authState.isLoading;
}
export function error(): string | null {
  return authState.error;
}
export function apiKey(): string | null {
  return authState.apiKey;
}
export function getApiKey(): string | null {
  return authState.apiKey;
}

/**
 * Set the API key and persist it to localStorage (browser only).
 */
export function setApiKey(key: string | null): void {
  authState.apiKey = key;
  if (browser) {
    if (key) {
      localStorage.setItem(API_KEY, key);
    } else {
      localStorage.removeItem(API_KEY);
    }
  }
}

/**
 * Check if SKIP_AUTH mode is enabled.
 * The window flag __SKIP_AUTH__ is respected in all builds to support isolated visual regression tests.
 * LocalStorage and query params remain dev-only to avoid accidental production bypass.
 */
export function skipAuthMode(): boolean {
  if (!browser) return false;
  if (window.__SKIP_AUTH__ === true) return true;
  if (!import.meta.env.DEV) return false;
  return localStorage.getItem("__SKIP_AUTH__") === "true";
}

/**
 * Check if user is authenticated.
 * SKIP_AUTH bypasses auth for testing via the window flag, localStorage, or query param.
 */
export function isAuthenticated(): boolean {
  if (browser) {
    if (window.__SKIP_AUTH__ === true) {
      return true;
    }
    if (import.meta.env.DEV) {
      if (localStorage.getItem("__SKIP_AUTH__") === "true") {
        return true;
      }
      const url = new URL(window.location.href);
      if (url.searchParams.get("skip_auth") === "true") {
        localStorage.setItem("__SKIP_AUTH__", "true");
        return true;
      }
    }
  }
  return !!authState.currentUser || !!authState.apiKey;
}

/**
 * Check if user is admin
 */
export function isAdmin(): boolean {
  return authState.currentUser?.role === "admin";
}

/**
 * Save tokens to in-memory state. Tokens are also stored as HttpOnly cookies
 * by the backend, so JavaScript never persists them.
 */
export function saveTokens(tokens: AuthTokens): void {
  authState.accessToken = tokens.access_token;
  authState.refreshToken = tokens.refresh_token;
}

/**
 * Clear all auth state
 */
export function clearAuthState(): void {
  authState.accessToken = null;
  authState.refreshToken = null;
  authState.currentUser = null;
  authState.apiKey = null;
  authState.error = null;

  if (browser) {
    localStorage.removeItem(API_KEY);
  }
}
