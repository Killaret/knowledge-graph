// Authentication API tests
import { describe, it, expect, vi, beforeEach } from "vitest";

// Create mock functions using vi.hoisted
const { mockPost, mockGet } = vi.hoisted(() => {
  const mockPost = vi.fn();
  const mockGet = vi.fn();
  return { mockPost, mockGet };
});

// Mock API client with proper chain structure
vi.mock("./client", () => ({
  default: {
    post: mockPost,
    get: mockGet,
  },
}));

// Import after mock
import {
  login,
  register,
  refreshTokens,
  logout,
  forgotPassword,
  resetPassword,
  getYandexLoginUrl,
  handleYandexCallback,
} from "./auth";

describe("auth API", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("login", () => {
    it("should make POST request to login endpoint", async () => {
      const mockResponse = {
        access_token: "access123",
        refresh_token: "refresh123",
      };
      const mockJson = vi.fn().mockResolvedValue(mockResponse);
      mockPost.mockReturnValue({ json: mockJson });

      const result = await login("testuser", "password123");

      expect(mockPost).toHaveBeenCalledWith("v1/auth/login", {
        json: { login: "testuser", password: "password123" },
      });
      expect(result).toEqual(mockResponse);
    });
  });

  describe("register", () => {
    it("should make POST request to register endpoint without email", async () => {
      const mockResponse = {
        access_token: "access123",
        refresh_token: "refresh123",
      };
      const mockJson = vi.fn().mockResolvedValue(mockResponse);
      mockPost.mockReturnValue({ json: mockJson });

      const result = await register("testuser", "password123");

      expect(mockPost).toHaveBeenCalledWith("v1/auth/register", {
        json: { login: "testuser", password: "password123" },
      });
      expect(result).toEqual(mockResponse);
    });

    it("should make POST request to register endpoint with email", async () => {
      const mockResponse = {
        access_token: "access123",
        refresh_token: "refresh123",
      };
      const mockJson = vi.fn().mockResolvedValue(mockResponse);
      mockPost.mockReturnValue({ json: mockJson });

      const result = await register(
        "testuser",
        "password123",
        "test@example.com",
      );

      expect(mockPost).toHaveBeenCalledWith("v1/auth/register", {
        json: {
          login: "testuser",
          password: "password123",
          email: "test@example.com",
        },
      });
      expect(result).toEqual(mockResponse);
    });
  });

  describe("refreshTokens", () => {
    it("should make POST request to refresh endpoint without a body", async () => {
      const mockResponse = {
        access_token: "newaccess123",
        refresh_token: "newrefresh123",
      };
      const mockJson = vi.fn().mockResolvedValue(mockResponse);
      mockPost.mockReturnValue({ json: mockJson });

      const result = await refreshTokens();

      expect(mockPost).toHaveBeenCalledWith("v1/auth/refresh");
      expect(result).toEqual(mockResponse);
    });
  });

  describe("logout", () => {
    it("should make POST request to logout endpoint without a body", async () => {
      const mockJson = vi.fn().mockResolvedValue({});
      mockPost.mockReturnValue({ json: mockJson });

      await logout();

      expect(mockPost).toHaveBeenCalledWith("v1/auth/logout");
    });
  });

  describe("forgotPassword", () => {
    it("should make POST request to forgot-password endpoint", async () => {
      const mockJson = vi.fn().mockResolvedValue({});
      mockPost.mockReturnValue({ json: mockJson });

      await forgotPassword("test@example.com");

      expect(mockPost).toHaveBeenCalledWith("v1/auth/forgot-password", {
        json: { email: "test@example.com" },
      });
    });
  });

  describe("resetPassword", () => {
    it("should make POST request to reset-password endpoint", async () => {
      const mockJson = vi.fn().mockResolvedValue({});
      mockPost.mockReturnValue({ json: mockJson });

      await resetPassword("token123", "newpassword123");

      expect(mockPost).toHaveBeenCalledWith("v1/auth/reset-password", {
        json: { token: "token123", new_password: "newpassword123" },
      });
    });
  });

  describe("getYandexLoginUrl", () => {
    it("should make GET request to yandex login endpoint", async () => {
      const mockResponse = { url: "https://oauth.yandex.ru/authorize?..." };
      const mockJson = vi.fn().mockResolvedValue(mockResponse);
      mockGet.mockReturnValue({ json: mockJson });

      const result = await getYandexLoginUrl();

      expect(mockGet).toHaveBeenCalledWith("v1/auth/yandex/login");
      expect(result).toEqual(mockResponse);
    });
  });

  describe("handleYandexCallback", () => {
    it("should make GET request to yandex callback endpoint", async () => {
      const mockResponse = {
        access_token: "yandexaccess123",
        refresh_token: "yandexrefresh123",
      };
      const mockJson = vi.fn().mockResolvedValue(mockResponse);
      mockGet.mockReturnValue({ json: mockJson });

      const result = await handleYandexCallback("code123", "state123");

      expect(mockGet).toHaveBeenCalledWith("v1/auth/yandex/callback", {
        searchParams: { code: "code123", state: "state123" },
      });
      expect(result).toEqual(mockResponse);
    });
  });
});
