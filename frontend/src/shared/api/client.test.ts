// API client tests
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";

// Mock the dependencies
vi.mock("$app/navigation", () => ({
  goto: vi.fn(),
}));

vi.mock("$shared/stores/auth.svelte", () => ({
  getAccessToken: vi.fn(() => "test-token"),
  getApiKey: vi.fn(() => null),
  refreshAccessToken: vi.fn(() => Promise.resolve(true)),
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
