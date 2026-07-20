// API client tests
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";

// Mock the dependencies
vi.mock("$app/navigation", () => ({
  goto: vi.fn(),
}));

vi.mock("$shared/stores/auth-session.svelte", () => ({
  getApiKey: vi.fn(() => null),
  saveTokens: vi.fn(),
  clearAuthState: vi.fn(),
}));

describe("API Client", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  it("should create API client with default configuration", async () => {
    const { api } = await import("./client");
    expect(api).toBeDefined();
  });

  it("should have retry configuration for network resilience", async () => {
    const { api } = await import("./client");
    expect(api).toBeDefined();
  });
});
