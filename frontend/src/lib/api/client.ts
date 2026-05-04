// API client wrapper для ky
import ky from 'ky';
import { goto } from '$app/navigation';
import { getAccessToken, getApiKey, refreshAccessToken } from '$lib/stores/auth.svelte';

// Определяем базовый URL в зависимости от среды
// import.meta.env доступен в Vite/SvelteKit
const isDev = import.meta.env.DEV;

// Проверяем тестовое окружение через process.env.VITEST
// Vitest устанавливает эту переменную автоматически
const isTest = typeof process !== 'undefined' && process.env?.VITEST === 'true';

// Получаем backend URL из env (для Docker) или используем default
let backendUrl = 'http://localhost:8080';
try {
  // @ts-ignore - Vite env vars
  const envUrl = (import.meta as any).env?.VITE_BACKEND_URL;
  if (envUrl) backendUrl = envUrl;
} catch {
  // Fallback to default
}

// В dev режиме (Vite) используем относительный путь (проксируется на backend)
// Production использует прямой backend URL
const prefixUrl = isDev && !isTest
  ? '' 
  : `${backendUrl}/api`;

// Flag to prevent infinite refresh loops
let isRefreshing = false;
let refreshPromise: Promise<boolean> | null = null;

// Базовый URL с прокси /api → http://localhost:8080
// Retry настроен для устойчивости к временным сетевым сбоям:
// - limit: 3 попытки (4 total: initial + 3 retries)
// - methods: все HTTP методы кроме DELETE (DELETE не идемпотентен)
// - statusCodes: сетевые и серверные ошибки, rate limiting
// Ky использует встроенный exponential backoff с jitter между попытками
export const api = ky.create({
  prefixUrl,
  timeout: 30000,
  retry: {
    limit: 3,
    methods: ['get', 'post', 'put', 'patch'],
    statusCodes: [408, 413, 429, 500, 502, 503, 504]
  },
  hooks: {
    beforeRequest: [
      async (request) => {
        // Add Authorization header if access token exists
        const token = getAccessToken();
        if (token) {
          request.headers.set('Authorization', `Bearer ${token}`);
        }
        
        // Add API Key header if exists (for API key auth)
        const key = getApiKey();
        if (key) {
          request.headers.set('X-API-Key', key);
        }
      }
    ],
    afterResponse: [
      async (request, options, response) => {
        // Handle 401 Unauthorized
        if (response.status === 401) {
          // Prevent infinite loops - if we're already refreshing, redirect to login
          if (isRefreshing) {
            // Wait for the current refresh to complete
            if (refreshPromise) {
              const refreshed = await refreshPromise;
              if (!refreshed) {
                // Refresh failed, clear auth and redirect
                goto('/auth/login');
                return response;
              }
            }
          } else {
            // Try to refresh the token
            isRefreshing = true;
            refreshPromise = refreshAccessToken();
            
            try {
              const refreshed = await refreshPromise;
              
              if (refreshed) {
                // Retry the original request with new token
                const newToken = getAccessToken();
                if (newToken) {
                  request.headers.set('Authorization', `Bearer ${newToken}`);
                  // Retry the request
                  return ky(request);
                }
              }
              
              // Refresh failed, redirect to login
              goto('/auth/login');
            } catch (error) {
              console.error('Token refresh error:', error);
              goto('/auth/login');
            } finally {
              isRefreshing = false;
              refreshPromise = null;
            }
          }
        }
        
        return response;
      }
    ]
  }
});

export default api;
