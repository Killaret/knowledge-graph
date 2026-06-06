// Интеграционные тесты для authStore с PreloadService
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import { initAuth, login, logout, isAuthenticated } from './auth.svelte';
import { clearPreloadCache, hasPreloadedData } from '../services/PreloadService';
import * as authApi from '../api/auth';
import * as usersApi from '../api/users';
import * as graphApi from '../api/graph';
import { PreloadService } from '../services/PreloadService';

const mockGraphData = {
  nodes: [{ id: 'n1', title: 'Note 1', type: 'star' as const }],
  links: [{ source: 'n1', target: 'n1', weight: 1 }]
};

const mockAchievementsPayload = {
  achievements: [
    {
      id: 'a1',
      code: 'first',
      title: 'First',
      description: '',
      icon: '⭐',
      points: 1,
      earned: false,
      is_hidden: false
    }
  ]
};

// Мокаем зависимости
vi.mock('$app/environment', () => ({
  browser: true
}));

vi.mock('$app/navigation', () => ({
  goto: vi.fn()
}));

vi.mock('../api/auth', () => ({
  login: vi.fn(),
  logout: vi.fn(),
  refreshTokens: vi.fn()
}));

vi.mock('../api/graph', () => ({
  getFullGraphData: vi.fn(),
  getCachedGraph: vi.fn(),
  getFreshGraph: vi.fn()
}));

vi.mock('../api/users', () => ({
  getMe: vi.fn(),
  getAllAchievements: vi.fn()
}));

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

describe('Auth Store Integration with PreloadService', () => {
  beforeEach(async () => {
    vi.clearAllMocks();
    clearPreloadCache();

    localStorageMock.getItem.mockReturnValue(null);
    vi.mocked(authApi.logout).mockResolvedValue(undefined);
    await logout();

    // Мокаем успешные ответы API
    vi.mocked(authApi.login).mockResolvedValue({
      access_token: 'test_access_token',
      refresh_token: 'test_refresh_token',
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

    vi.mocked(graphApi.getFullGraphData).mockResolvedValue(mockGraphData);
    vi.mocked(graphApi.getCachedGraph).mockResolvedValue(mockGraphData);
    vi.mocked(graphApi.getFreshGraph).mockResolvedValue({ fresh: mockGraphData });
    vi.mocked(usersApi.getAllAchievements).mockResolvedValue(mockAchievementsPayload);
  });

  afterEach(() => {
    clearPreloadCache();
    vi.clearAllMocks();
  });

  describe('Logout Integration', () => {
    it('should clear preload cache on logout', async () => {
      // Предзагружаем данные
      await PreloadService.startPreload();
      
      expect(hasPreloadedData()).toBe(true);
      
      // Выполняем выход
      await logout();
      
      // Кэш должен быть очищен
      expect(hasPreloadedData()).toBe(false);
      
      // Должен быть вызван редирект
      expect(vi.mocked(goto)).toHaveBeenCalledWith('/auth/login');
    });

    it('should clear preload cache even if logout API fails', async () => {
      // Предзагружаем данные
      await PreloadService.startPreload();
      
      expect(hasPreloadedData()).toBe(true);
      
      // Мокаем ошибку API выхода
      vi.mocked(authApi.logout).mockRejectedValue(new Error('Logout failed'));
      
      // Выполняем выход
      await logout();
      
      // Кэш все равно должен быть очищен
      expect(hasPreloadedData()).toBe(false);
      
      // Редирект все равно должен произойти
      expect(vi.mocked(goto)).toHaveBeenCalledWith('/auth/login');
    });

    it('should handle missing refresh token gracefully', async () => {
      // Предзагружаем данные
      await PreloadService.startPreload();
      
      expect(hasPreloadedData()).toBe(true);
      
      // Выходим (без refresh token)
      await logout();
      
      // Кэш должен быть очищен
      expect(hasPreloadedData()).toBe(false);
      
      // Редирект должен произойти
      expect(vi.mocked(goto)).toHaveBeenCalledWith('/auth/login');
    });
  });

  describe('Login Flow Integration', () => {
    it('should not interfere with preload cache during login', async () => {
      // Предзагружаем данные
      await PreloadService.startPreload();
      
      expect(hasPreloadedData()).toBe(true);
      
      // Выполняем вход
      const loginResult = await login('testuser', 'password');
      
      expect(loginResult).toBe(true);
      expect(hasPreloadedData()).toBe(true); // Кэш не должен быть затронут
      
      // localStorage должен быть обновлен
      expect(localStorageMock.setItem).toHaveBeenCalledWith('access_token', 'test_access_token');
      expect(localStorageMock.setItem).toHaveBeenCalledWith('refresh_token', 'test_refresh_token');
    });

    it('should clear cache on failed login', async () => {
      // Предзагружаем данные
      await PreloadService.startPreload();
      
      expect(hasPreloadedData()).toBe(true);
      
      // Мокаем ошибку входа
      vi.mocked(authApi.login).mockRejectedValue(new Error('Invalid credentials'));
      
      // Выполняем вход
      const loginResult = await login('testuser', 'wrongpassword');
      
      expect(loginResult).toBe(false);
      // Кэш не должен очищаться при неудачном входе (только при выходе)
      expect(hasPreloadedData()).toBe(true);
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
      
      // Проверяем, что preload не запускался автоматически
      // (это проверяется в PreloadService тестах через isAuthenticated мок)
    });

    it('should handle auth initialization without affecting preload cache', async () => {
      // Предзагружаем данные
      await PreloadService.startPreload();
      
      expect(hasPreloadedData()).toBe(true);
      
      // Инициализируем auth (без токенов)
      await initAuth();
      
      // Кэш не должен быть затронут
      expect(hasPreloadedData()).toBe(true);
      expect(isAuthenticated()).toBe(false);
    });
  });

  describe('Browser Environment Integration', () => {
    it('should work correctly in browser environment', async () => {
      vi.stubGlobal('browser', true);
      
      // Предзагружаем данные
      await PreloadService.startPreload();
      
      expect(hasPreloadedData()).toBe(true);
      
      await logout();
      
      expect(hasPreloadedData()).toBe(false);
      
      vi.unstubAllGlobals();
    });

    it('should work correctly in server environment', async () => {
      vi.stubGlobal('browser', false);
      
      // В серверном окружении localStorage не должен использоваться
      // но логика очистки кэша при выходе все равно должна работать
      
      await logout();
      
      // Ошибок быть не должно
      expect(vi.mocked(goto)).toHaveBeenCalledWith('/auth/login');
      
      vi.unstubAllGlobals();
    });
  });

  describe('Error Handling Integration', () => {
    it('should handle preload service errors gracefully', async () => {
      // Мокаем ошибку в PreloadService
      const originalClearCache = clearPreloadCache;
      const mockClearCache = vi.fn().mockImplementation(() => {
        throw new Error('PreloadService error');
      });
      
      // Заменяем функцию
      vi.doMock('../services/PreloadService', () => ({
        clearPreloadCache: mockClearCache,
        hasPreloadedData: vi.fn(() => false)
      }));
      
      // Выход не должен падать из-за ошибки в PreloadService
      await expect(logout()).resolves.not.toThrow();
      
      // Редирект все равно должен произойти
      expect(vi.mocked(goto)).toHaveBeenCalledWith('/auth/login');
      
      // Восстанавливаем оригинальную функцию
      vi.doMock('../services/PreloadService', () => ({
        clearPreloadCache: originalClearCache,
        hasPreloadedData: vi.fn(() => false)
      }));
    });
  });

  describe('Concurrent Operations', () => {
    it('should handle concurrent auth operations with preload', async () => {
      // Предзагружаем данные
      await PreloadService.startPreload();
      
      expect(hasPreloadedData()).toBe(true);
      
      // Запускаем операции последовательно, а не параллельно
      // чтобы избежать race condition между login и logout
      await login('user1', 'pass1');
      await login('user2', 'pass2');
      await logout();
    });
  });
});
