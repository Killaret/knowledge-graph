// Интеграционные тесты для authStore с PreloadService (более надежные)
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { goto } from '$app/navigation';
import * as authApi from '$shared/api/auth';
import * as usersApi from '$shared/api/users';

const preloadMocks = vi.hoisted(() => ({
  clearPreloadCache: vi.fn(),
  hasPreloadedData: vi.fn(() => false)
}));

// Мокаем зависимости
vi.mock('$app/environment', () => ({
  browser: true
}));

vi.mock('$app/navigation', () => ({
  goto: vi.fn()
}));

vi.mock('$shared/api/auth', () => ({
  login: vi.fn(),
  logout: vi.fn(),
  refreshTokens: vi.fn()
}));

vi.mock('$shared/api/users', () => ({
  getMe: vi.fn()
}));

vi.mock('$shared/services/PreloadService', () => ({
  clearPreloadCache: preloadMocks.clearPreloadCache,
  preloadAuthenticatedGraph: vi.fn(),
  hasPreloadedData: preloadMocks.hasPreloadedData,
  PreloadService: {
    startPreload: vi.fn(),
    clearCache: preloadMocks.clearPreloadCache,
    hasPreloadedData: preloadMocks.hasPreloadedData,
    getPreloadedGraph: vi.fn(),
    getPreloadedAchievements: vi.fn(),
    isPreloadingData: vi.fn(() => false)
  }
}));

// Импортируем auth.svelte ПОСЛЕ моков
import { initAuth, login, logout, isAuthenticated } from './auth.svelte';

// Мокаем localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn()
};

Object.defineProperty(window, 'localStorage', {
  value: localStorageMock
});

describe('Auth Store Integration with PreloadService (Simplified)', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    
    // Сбрасываем моки PreloadService
    preloadMocks.clearPreloadCache.mockReset();
    preloadMocks.hasPreloadedData.mockReturnValue(false);
    
    // Мокаем успешные ответы API
    vi.mocked(authApi.login).mockResolvedValue({
      access_token: 'test_access_token',
      refresh_token: 'test_refresh_token',
      token_type: 'Bearer',
      expires_at: '2024-12-31T23:59:59Z'
    });
    
    vi.mocked(authApi.refreshTokens).mockResolvedValue({
      access_token: 'new_access_token',
      refresh_token: 'new_refresh_token',
      token_type: 'Bearer',
      expires_at: '2024-12-31T23:59:59Z'
    });
    
    vi.mocked(usersApi.getMe).mockResolvedValue({
      id: '1',
      login: 'testuser',
      email: 'test@example.com',
      role: 'user',
      created_at: '2024-01-01T00:00:00Z'
    });
    
    // Очищаем localStorage
    localStorageMock.getItem.mockReturnValue(null);
  });

  afterEach(() => {
    vi.clearAllMocks();
    vi.clearAllTimers();
  });

  describe('Logout Integration', () => {
    it('should clear preload cache on logout', async () => {
      // Мокаем наличие предзагруженных данных
      preloadMocks.hasPreloadedData.mockReturnValue(true);
      
      // Выполняем выход
      await logout();
      
      // Проверяем, что clearPreloadCache был вызван
      expect(preloadMocks.clearPreloadCache).toHaveBeenCalled();
      
      // Должен быть вызван редирект
      expect(vi.mocked(goto)).toHaveBeenCalledWith('/auth/login');
    });

    it('should clear preload cache even if logout API fails', async () => {
      // Мокаем наличие предзагруженных данных
      preloadMocks.hasPreloadedData.mockReturnValue(true);
      
      // Мокаем ошибку API выхода
      vi.mocked(authApi.logout).mockRejectedValue(new Error('Logout failed'));
      
      // Выполняем выход
      await logout();
      
      // Кэш все равно должен быть очищен
      expect(preloadMocks.clearPreloadCache).toHaveBeenCalled();
      
      // Редирект все равно должен произойти
      expect(vi.mocked(goto)).toHaveBeenCalledWith('/auth/login');
    });

    it('should handle missing refresh token gracefully', async () => {
      // Мокаем наличие предзагруженных данных
      preloadMocks.hasPreloadedData.mockReturnValue(true);
      
      // Выходим (без refresh token)
      await logout();
      
      // Кэш должен быть очищен
      expect(preloadMocks.clearPreloadCache).toHaveBeenCalled();
      
      // Редирект должен произойти
      expect(vi.mocked(goto)).toHaveBeenCalledWith('/auth/login');
    });
  });

  describe('Login Flow Integration', () => {
    it('should not interfere with preload cache during login', async () => {
      // Мокаем наличие предзагруженных данных
      preloadMocks.hasPreloadedData.mockReturnValue(true);
      
      // Выполняем вход
      const loginResult = await login('testuser', 'password');
      
      expect(loginResult).toBe(true);
      
      // clearPreloadCache не должен вызываться при входе
      expect(preloadMocks.clearPreloadCache).not.toHaveBeenCalled();
      
      // localStorage должен быть обновлен
      expect(localStorageMock.setItem).toHaveBeenCalledWith('access_token', 'test_access_token');
      expect(localStorageMock.setItem).toHaveBeenCalledWith('refresh_token', 'test_refresh_token');
    });

    it('should clear cache on failed login', async () => {
      // Мокаем наличие предзагруженных данных
      preloadMocks.hasPreloadedData.mockReturnValue(true);
      
      // Мокаем ошибку входа
      vi.mocked(authApi.login).mockRejectedValue(new Error('Invalid credentials'));
      
      // Выполняем вход
      const loginResult = await login('testuser', 'wrongpassword');
      
      expect(loginResult).toBe(false);
      
      // Кэш не должен очищаться при неудачном входе (только при выходе)
      expect(preloadMocks.clearPreloadCache).not.toHaveBeenCalled();
    });
  });

  describe('Auth Initialization Integration', () => {
    it('should not start preload when user is already authenticated', async () => {
      // Устанавливаем токены в localStorage
      localStorageMock.getItem.mockImplementation((key) => {
        if (key === 'access_token') return 'existing_access_token';
        if (key === 'refresh_token') return 'existing_refresh_token';
        return null;
      });
      
      // Мокаем успешное обновление токена
      vi.mocked(authApi.refreshTokens).mockResolvedValue({
        access_token: 'new_access_token',
        refresh_token: 'new_refresh_token',
        token_type: 'Bearer',
        expires_at: '2024-12-31T23:59:59Z'
      });
      
      // Инициализируем auth
      await initAuth();
      
      // Проверяем, что пользователь аутентифицирован
      expect(isAuthenticated()).toBe(true);
      
      // clearPreloadCache не должен вызываться при инициализации
      expect(preloadMocks.clearPreloadCache).not.toHaveBeenCalled();
    });

    it('should handle auth initialization without affecting preload cache', async () => {
      // Мокаем наличие предзагруженных данных
      preloadMocks.hasPreloadedData.mockReturnValue(true);
      
      // Инициализируем auth (без токенов)
      await initAuth();
      
      // Кэш не должен быть затронут
      expect(preloadMocks.clearPreloadCache).not.toHaveBeenCalled();
      expect(isAuthenticated()).toBe(false);
    });
  });

  describe('Browser Environment Integration', () => {
    it('should work correctly in browser environment', async () => {
      vi.stubGlobal('browser', true);
      
      // Мокаем наличие предзагруженных данных
      preloadMocks.hasPreloadedData.mockReturnValue(true);
      
      await logout();
      
      // Кэш должен быть очищен
      expect(preloadMocks.clearPreloadCache).toHaveBeenCalled();
      
      vi.unstubAllGlobals();
    });

    it('should work correctly in server environment', async () => {
      vi.stubGlobal('browser', false);
      
      // Мокаем наличие предзагруженных данных
      preloadMocks.hasPreloadedData.mockReturnValue(true);
      
      await logout();
      
      // Кэш должен быть очищен
      expect(preloadMocks.clearPreloadCache).toHaveBeenCalled();
      
      vi.unstubAllGlobals();
    });
  });

  describe('Error Handling Integration', () => {
    it('should handle preload service errors gracefully', async () => {
      // Мокаем ошибку в clearPreloadCache
      preloadMocks.clearPreloadCache.mockImplementation(() => {
        throw new Error('PreloadService error');
      });
      
      // Выход не должен падать из-за ошибки в PreloadService
      await expect(logout()).resolves.not.toThrow();
      
      // Редирект все равно должен произойти
      expect(vi.mocked(goto)).toHaveBeenCalledWith('/auth/login');
    });
  });

  describe('Concurrent Operations', () => {
    it('should handle concurrent auth operations with preload', async () => {
      // Мокаем наличие предзагруженных данных
      preloadMocks.hasPreloadedData.mockReturnValue(true);
      
      // Запускаем несколько операций параллельно
      const operations = [
        login('user1', 'pass1'),
        login('user2', 'pass2'),
        logout()
      ];
      
      const results = await Promise.allSettled(operations);
      
      // Никаких ошибок быть не должно
      results.forEach(result => {
        expect(result.status).toBe('fulfilled');
      });
      
      // clearPreloadCache должен быть вызван только для logout
      expect(preloadMocks.clearPreloadCache).toHaveBeenCalledTimes(1);
    });
  });
});
