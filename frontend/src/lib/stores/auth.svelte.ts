// Auth store with Svelte 5 runes
import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import * as authApi from '$lib/api/auth';
import * as usersApi from '$lib/api/users';
import type { User, AuthTokens } from '$lib/types';

// Global reactive state
export let currentUser = $state<User | null>(null);
export let accessToken = $state<string | null>(null);
export let refreshToken = $state<string | null>(null);
export let isInitialized = $state(false);
export let isLoading = $state(false);
export let error = $state<string | null>(null);

// API Key support
export let apiKey = $state<string | null>(null);

// LocalStorage keys
const ACCESS_TOKEN_KEY = 'access_token';
const REFRESH_TOKEN_KEY = 'refresh_token';
const API_KEY = 'api_key';

/**
 * Initialize auth state from localStorage
 */
export async function initAuth(): Promise<void> {
  if (!browser) {
    isInitialized = true;
    return;
  }

  try {
    // Try to load tokens from localStorage
    const storedAccessToken = localStorage.getItem(ACCESS_TOKEN_KEY);
    const storedRefreshToken = localStorage.getItem(REFRESH_TOKEN_KEY);
    const storedApiKey = localStorage.getItem(API_KEY);

    if (storedApiKey) {
      apiKey = storedApiKey;
    }

    if (storedAccessToken) {
      accessToken = storedAccessToken;
    }

    if (storedRefreshToken) {
      refreshToken = storedRefreshToken;
      
      // Try to refresh the token and get user info
      const refreshed = await refreshAccessToken();
      
      if (refreshed) {
        try {
          const user = await usersApi.getMe();
          currentUser = user;
        } catch (e) {
          // If getting user fails, clear auth state
          clearAuthState();
        }
      }
    }
  } catch (e) {
    console.error('Failed to initialize auth:', e);
    clearAuthState();
  } finally {
    isInitialized = true;
  }
}

/**
 * Login with credentials
 */
export async function login(login: string, password: string): Promise<boolean> {
  isLoading = true;
  error = null;

  try {
    const tokens = await authApi.login(login, password);
    
    // Save tokens
    saveTokens(tokens);
    
    // Get user info
    const user = await usersApi.getMe();
    currentUser = user;
    
    return true;
  } catch (e) {
    error = e instanceof Error ? e.message : 'Login failed';
    clearAuthState();
    return false;
  } finally {
    isLoading = false;
  }
}

/**
 * Register new user
 */
export async function register(login: string, password: string, email?: string): Promise<boolean> {
  isLoading = true;
  error = null;

  try {
    const tokens = await authApi.register(login, password, email);
    
    // Save tokens
    saveTokens(tokens);
    
    // Get user info
    const user = await usersApi.getMe();
    currentUser = user;
    
    return true;
  } catch (e) {
    error = e instanceof Error ? e.message : 'Registration failed';
    clearAuthState();
    return false;
  } finally {
    isLoading = false;
  }
}

/**
 * Logout user
 */
export async function logout(): Promise<void> {
  if (refreshToken) {
    try {
      await authApi.logout(refreshToken);
    } catch (e) {
      // Ignore errors during logout
      console.error('Logout error:', e);
    }
  }
  
  clearAuthState();
  goto('/auth/login');
}

/**
 * Refresh access token
 */
export async function refreshAccessToken(): Promise<boolean> {
  if (!refreshToken) {
    return false;
  }

  try {
    const tokens = await authApi.refreshTokens(refreshToken);
    saveTokens(tokens);
    return true;
  } catch (e) {
    console.error('Token refresh failed:', e);
    clearAuthState();
    return false;
  }
}

/**
 * Handle Yandex OAuth callback
 */
export async function handleYandexCallback(code: string, state: string): Promise<boolean> {
  isLoading = true;
  error = null;

  try {
    const tokens = await authApi.handleYandexCallback(code, state);
    
    // Save tokens
    saveTokens(tokens);
    
    // Get user info
    const user = await usersApi.getMe();
    currentUser = user;
    
    return true;
  } catch (e) {
    error = e instanceof Error ? e.message : 'Yandex authentication failed';
    clearAuthState();
    return false;
  } finally {
    isLoading = false;
  }
}

/**
 * Login with API Key
 */
export async function loginWithApiKey(key: string): Promise<boolean> {
  isLoading = true;
  error = null;

  try {
    // Save API key
    apiKey = key;
    if (browser) {
      localStorage.setItem(API_KEY, key);
    }
    
    // Try to get user info with API key
    const user = await usersApi.getMe();
    currentUser = user;
    
    return true;
  } catch (e) {
    error = e instanceof Error ? e.message : 'Invalid API key';
    apiKey = null;
    if (browser) {
      localStorage.removeItem(API_KEY);
    }
    return false;
  } finally {
    isLoading = false;
  }
}

/**
 * Get current access token
 */
export function getAccessToken(): string | null {
  return accessToken;
}

/**
 * Get API key
 */
export function getApiKey(): string | null {
  return apiKey;
}

/**
 * Check if user is authenticated
 */
export function isAuthenticated(): boolean {
  return !!accessToken || !!apiKey;
}

/**
 * Check if user is admin
 */
export function isAdmin(): boolean {
  return currentUser?.role === 'admin';
}

/**
 * Save tokens to state and localStorage
 */
function saveTokens(tokens: AuthTokens): void {
  accessToken = tokens.access_token;
  refreshToken = tokens.refresh_token;
  
  if (browser) {
    localStorage.setItem(ACCESS_TOKEN_KEY, tokens.access_token);
    localStorage.setItem(REFRESH_TOKEN_KEY, tokens.refresh_token);
  }
}

/**
 * Clear all auth state
 */
function clearAuthState(): void {
  accessToken = null;
  refreshToken = null;
  currentUser = null;
  apiKey = null;
  error = null;
  
  if (browser) {
    localStorage.removeItem(ACCESS_TOKEN_KEY);
    localStorage.removeItem(REFRESH_TOKEN_KEY);
    localStorage.removeItem(API_KEY);
  }
}

/**
 * Update user info
 */
export async function updateUserInfo(): Promise<void> {
  if (!isAuthenticated()) {
    return;
  }

  try {
    const user = await usersApi.getMe();
    currentUser = user;
  } catch (e) {
    console.error('Failed to update user info:', e);
  }
}
