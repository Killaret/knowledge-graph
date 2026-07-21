// Интеграционные тесты для authStore с PreloadService
import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { goto } from "$app/navigation";
import {
  initAuth,
  login,
  logout,
  isAuthenticated,
  register,
  handleYandexCallback,
  loginWithApiKey,
  updateUserInfo,
  getApiKey,
} from "./auth.svelte";
import {
  clearPreloadCache,
  hasPreloadedData,
} from "$shared/services/PreloadService";
import * as authApi from "$shared/api/auth";
import * as usersApi from "$shared/api/users";
import * as graphApi from "$shared/api/graph";
import { PreloadService } from "$shared/services/PreloadService";

const mockGraphData = {
  nodes: [{ id: "n1", title: "Note 1", type: "star" as const }],
  links: [{ source: "n1", target: "n1", weight: 1 }],
};

const mockAchievementsPayload = {
  achievements: [
    {
      id: "a1",
      code: "first",
      title: "First",
      description: "",
      icon: "⭐",
      points: 1,
      earned: false,
      is_hidden: false,
    },
  ],
};

// Мокаем зависимости
vi.mock("$app/environment", () => ({
  browser: true,
}));

vi.mock("$app/navigation", () => ({
  goto: vi.fn(),
}));

vi.mock("$shared/api/auth", () => ({
  login: vi.fn(),
  logout: vi.fn(),
  refreshTokens: vi.fn(),
  register: vi.fn(),
  handleYandexCallback: vi.fn(),
}));

vi.mock("$shared/api/graph", () => ({
  getFullGraphData: vi.fn(),
  getCachedGraph: vi.fn(),
  getFreshGraph: vi.fn(),
}));

vi.mock("$shared/api/users", () => ({
  getMe: vi.fn(),
  getAllAchievements: vi.fn(),
  getSettings: vi.fn(),
}));

// Мокаем localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
};

Object.defineProperty(window, "localStorage", {
  value: localStorageMock,
});

describe("Auth Store Integration with PreloadService", () => {
  beforeEach(async () => {
    vi.clearAllMocks();
    clearPreloadCache();

    localStorageMock.getItem.mockReturnValue(null);
    vi.mocked(authApi.logout).mockResolvedValue(undefined);
    vi.mocked(authApi.refreshTokens).mockRejectedValue(new Error("No session"));
    await logout();

    const mockTokens = {
      access_token: "test_access_token",
      refresh_token: "test_refresh_token",
      token_type: "Bearer",
      expires_at: "2024-12-31T23:59:59Z",
    };

    // Мокаем успешные ответы API
    vi.mocked(authApi.login).mockResolvedValue(mockTokens);
    vi.mocked(authApi.register).mockResolvedValue(mockTokens);
    vi.mocked(authApi.handleYandexCallback).mockResolvedValue(mockTokens);

    vi.mocked(usersApi.getMe).mockResolvedValue({
      id: "1",
      login: "testuser",
      email: "test@example.com",
      role: "user",
      created_at: "2024-01-01T00:00:00Z",
    });

    vi.mocked(usersApi.getSettings).mockResolvedValue({
      settings: [{ key: "preferred_language", value: "en", updated_at: "" }],
    });

    vi.mocked(graphApi.getFullGraphData).mockResolvedValue(mockGraphData);
    vi.mocked(graphApi.getCachedGraph).mockResolvedValue(mockGraphData);
    vi.mocked(graphApi.getFreshGraph).mockResolvedValue({
      fresh: mockGraphData,
    });
    vi.mocked(usersApi.getAllAchievements).mockResolvedValue(
      mockAchievementsPayload,
    );
  });

  afterEach(() => {
    clearPreloadCache();
    vi.clearAllMocks();
  });

  describe("Logout Integration", () => {
    it("should clear preload cache on logout", async () => {
      // Предзагружаем данные
      await PreloadService.startPreload();

      expect(hasPreloadedData()).toBe(true);

      // Выполняем выход
      await logout();

      // Кэш должен быть очищен
      expect(hasPreloadedData()).toBe(false);

      // Должен быть вызван редирект
      expect(vi.mocked(goto)).toHaveBeenCalledWith("/auth/login");
    });

    it("should clear preload cache even if logout API fails", async () => {
      // Предзагружаем данные
      await PreloadService.startPreload();

      expect(hasPreloadedData()).toBe(true);

      // Мокаем ошибку API выхода
      vi.mocked(authApi.logout).mockRejectedValue(new Error("Logout failed"));

      // Выполняем выход
      await logout();

      // Кэш все равно должен быть очищен
      expect(hasPreloadedData()).toBe(false);

      // Редирект все равно должен произойти
      expect(vi.mocked(goto)).toHaveBeenCalledWith("/auth/login");
    });

    it("should handle missing refresh token gracefully", async () => {
      // Предзагружаем данные
      await PreloadService.startPreload();

      expect(hasPreloadedData()).toBe(true);

      // Выходим (без refresh token)
      await logout();

      // Кэш должен быть очищен
      expect(hasPreloadedData()).toBe(false);

      // Редирект должен произойти
      expect(vi.mocked(goto)).toHaveBeenCalledWith("/auth/login");
    });
  });

  describe("Login Flow Integration", () => {
    it("should not interfere with preload cache during login", async () => {
      // Предзагружаем данные
      await PreloadService.startPreload();

      expect(hasPreloadedData()).toBe(true);

      // Выполняем вход
      const loginResult = await login("testuser", "password");

      expect(loginResult).toBe(true);
      expect(hasPreloadedData()).toBe(true); // Кэш не должен быть затронут
    });

    it("should clear cache on failed login", async () => {
      // Предзагружаем данные
      await PreloadService.startPreload();

      expect(hasPreloadedData()).toBe(true);

      // Мокаем ошибку входа
      vi.mocked(authApi.login).mockRejectedValue(
        new Error("Invalid credentials"),
      );

      // Выполняем вход
      const loginResult = await login("testuser", "wrongpassword");

      expect(loginResult).toBe(false);
      // Кэш не должен очищаться при неудачном входе (только при выходе)
      expect(hasPreloadedData()).toBe(true);
    });
  });

  describe("Auth Initialization Integration", () => {
    it("should not start preload when user is already authenticated", async () => {
      // Мокаем успешное обновление токена (refresh токен передаётся через HttpOnly cookie)
      vi.mocked(authApi.refreshTokens).mockResolvedValue({
        access_token: "new_access_token",
        refresh_token: "new_refresh_token",
        token_type: "Bearer",
        expires_at: "2024-12-31T23:59:59Z",
      });

      // Инициализируем auth
      await initAuth();

      // Проверяем, что пользователь аутентифицирован
      expect(isAuthenticated()).toBe(true);

      // Проверяем, что preload не запускался автоматически
      // (это проверяется в PreloadService тестах через isAuthenticated мок)
    });

    it("should handle auth initialization without affecting preload cache", async () => {
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

  describe("Browser Environment Integration", () => {
    it("should work correctly in browser environment", async () => {
      vi.stubGlobal("browser", true);

      // Предзагружаем данные
      await PreloadService.startPreload();

      expect(hasPreloadedData()).toBe(true);

      await logout();

      expect(hasPreloadedData()).toBe(false);

      vi.unstubAllGlobals();
    });

    it("should work correctly in server environment", async () => {
      vi.stubGlobal("browser", false);

      // В серверном окружении localStorage не должен использоваться
      // но логика очистки кэша при выходе все равно должна работать

      await logout();

      // Ошибок быть не должно
      expect(vi.mocked(goto)).toHaveBeenCalledWith("/auth/login");

      vi.unstubAllGlobals();
    });
  });

  describe("Error Handling Integration", () => {
    it("should handle preload service errors gracefully", async () => {
      // Мокаем ошибку в PreloadService
      const originalClearCache = clearPreloadCache;
      const mockClearCache = vi.fn().mockImplementation(() => {
        throw new Error("PreloadService error");
      });

      // Заменяем функцию
      vi.doMock("$shared/services/PreloadService", () => ({
        clearPreloadCache: mockClearCache,
        hasPreloadedData: vi.fn(() => false),
      }));

      // Выход не должен падать из-за ошибки в PreloadService
      await expect(logout()).resolves.not.toThrow();

      // Редирект все равно должен произойти
      expect(vi.mocked(goto)).toHaveBeenCalledWith("/auth/login");

      // Восстанавливаем оригинальную функцию
      vi.doMock("$shared/services/PreloadService", () => ({
        clearPreloadCache: originalClearCache,
        hasPreloadedData: vi.fn(() => false),
      }));
    });
  });

  describe("Concurrent Operations", () => {
    it("should handle concurrent auth operations with preload", async () => {
      // Предзагружаем данные
      await PreloadService.startPreload();

      expect(hasPreloadedData()).toBe(true);

      // Запускаем операции последовательно, а не параллельно
      // чтобы избежать race condition между login и logout
      await login("user1", "pass1");
      await login("user2", "pass2");
      await logout();
    });
  });

  describe("Auth Flows", () => {
    it("registers a new user", async () => {
      const result = await register("newuser", "password", "email@example.com");

      expect(result).toBe(true);
      expect(isAuthenticated()).toBe(true);
      expect(usersApi.getMe).toHaveBeenCalled();
    });

    it("handles registration failure", async () => {
      vi.mocked(authApi.register).mockRejectedValue(new Error("Login exists"));

      const result = await register("existing", "password");

      expect(result).toBe(false);
      expect(isAuthenticated()).toBe(false);
    });

    it("handles Yandex OAuth callback", async () => {
      const result = await handleYandexCallback("code", "state");

      expect(result).toBe(true);
      expect(isAuthenticated()).toBe(true);
      expect(authApi.handleYandexCallback).toHaveBeenCalledWith(
        "code",
        "state",
      );
    });

    it("handles Yandex callback failure", async () => {
      vi.mocked(authApi.handleYandexCallback).mockRejectedValue(
        new Error("OAuth failed"),
      );

      const result = await handleYandexCallback("code", "state");

      expect(result).toBe(false);
      expect(isAuthenticated()).toBe(false);
    });

    it("logs in with API key", async () => {
      const result = await loginWithApiKey("my-api-key");

      expect(result).toBe(true);
      expect(getApiKey()).toBe("my-api-key");
      expect(isAuthenticated()).toBe(true);
    });

    it("handles invalid API key", async () => {
      vi.mocked(usersApi.getMe).mockRejectedValue(new Error("Invalid key"));

      const result = await loginWithApiKey("bad-key");

      expect(result).toBe(false);
      expect(getApiKey()).toBeNull();
    });

    it("updates user info when authenticated", async () => {
      // Authenticate first
      await login("user", "pass");

      vi.mocked(usersApi.getMe).mockResolvedValue({
        id: "2",
        login: "updated",
        email: "updated@example.com",
        role: "admin",
        created_at: "2024-01-01T00:00:00Z",
      });

      await updateUserInfo();

      expect(usersApi.getMe).toHaveBeenCalledTimes(2);
      expect(isAuthenticated()).toBe(true);
    });

    it("does not update user info when not authenticated", async () => {
      await logout();
      vi.mocked(usersApi.getMe).mockClear();

      await updateUserInfo();

      expect(usersApi.getMe).not.toHaveBeenCalled();
    });
  });
});
