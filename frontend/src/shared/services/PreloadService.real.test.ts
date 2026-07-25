import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { mockGraphData } from "./__mocks__/PreloadService.mocks";

const mockEnv = vi.hoisted(() => ({ browser: true }));
vi.mock("$app/environment", () => mockEnv);

const mockAuth = vi.hoisted(() => ({
  isAuthenticated: vi.fn(() => false),
}));
vi.mock("$shared/stores/auth-session.svelte", () => mockAuth);

const mockGraphApi = vi.hoisted(() => ({
  getFullGraphData: vi.fn(),
  getGraphDelta: vi.fn(),
  normalizeNode: vi.fn((n) => n),
  normalizeLink: vi.fn((l) => l),
}));
vi.mock("$shared/api/graph", () => mockGraphApi);

describe("PreloadService (real)", () => {
  let PreloadService: any;

  beforeEach(async () => {
    vi.resetModules();
    vi.useFakeTimers({ shouldAdvanceTime: true });
    vi.setSystemTime(new Date("2026-01-01T00:00:00Z"));

    mockEnv.browser = true;
    mockAuth.isAuthenticated.mockReturnValue(false);
    mockGraphApi.getFullGraphData.mockResolvedValue({ ...mockGraphData, hash: "hash-1" });
    mockGraphApi.getGraphDelta.mockResolvedValue({ current_hash: "hash-2" });

    vi.spyOn(console, "log").mockImplementation(() => {});
    vi.spyOn(console, "warn").mockImplementation(() => {});
    vi.spyOn(console, "error").mockImplementation(() => {});

    const mod = await import("./PreloadService");
    PreloadService = mod.PreloadService;
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  it("preloads public graph when not authenticated", async () => {
    await PreloadService.startPreload();

    expect(mockGraphApi.getFullGraphData).toHaveBeenCalledTimes(1);
    expect(PreloadService.getPreloadedGraph()).toEqual({ ...mockGraphData, hash: "hash-1" });
    expect(PreloadService.hasPreloadedData()).toBe(true);
  });

  it("does not preload in non-browser environment", async () => {
    mockEnv.browser = false;

    await PreloadService.startPreload();

    expect(mockGraphApi.getFullGraphData).not.toHaveBeenCalled();
    expect(PreloadService.getPreloadedGraph()).toBeNull();
  });

  it("does not preload when authenticated", async () => {
    mockAuth.isAuthenticated.mockReturnValue(true);

    await PreloadService.startPreload();

    expect(mockGraphApi.getFullGraphData).not.toHaveBeenCalled();
    expect(PreloadService.getPreloadedGraph()).toBeNull();
  });

  it("preloads authenticated graph when user is logged in", async () => {
    mockAuth.isAuthenticated.mockReturnValue(true);

    await PreloadService.preloadAuthenticatedGraph();

    expect(mockGraphApi.getFullGraphData).toHaveBeenCalledTimes(1);
    expect(PreloadService.getPreloadedGraph()).toEqual({ ...mockGraphData, hash: "hash-1" });
    expect(PreloadService.getPreloadedGraphData()?.lastHash).toBe("hash-1");
  });

  it("deduplicates concurrent preload calls", async () => {
    let resolve: (value: any) => void = () => {};
    mockGraphApi.getFullGraphData.mockImplementation(() => new Promise((r) => (resolve = r)));

    const p1 = PreloadService.startPreload();
    const p2 = PreloadService.startPreload();

    resolve(mockGraphData);
    await Promise.all([p1, p2]);

    expect(mockGraphApi.getFullGraphData).toHaveBeenCalledTimes(1);
    expect(PreloadService.getPreloadedGraph()).toEqual(mockGraphData);
  });

  it("returns null for expired preloaded graph", async () => {
    await PreloadService.startPreload();
    expect(PreloadService.getPreloadedGraphData()).not.toBeNull();

    vi.advanceTimersByTime(6 * 60 * 1000); // 6 minutes, TTL is 5

    expect(PreloadService.getPreloadedGraphData()).toBeNull();
    expect(PreloadService.getPreloadedGraph()).toBeNull();
  });

  it("updates preloaded graph with delta", async () => {
    const delta = {
      added_nodes: [{ id: "n4", title: "New", type: "star" }],
      removed_nodes: [],
      updated_nodes: [],
      added_links: [],
      removed_links: [],
      current_hash: "hash-2",
    };
    mockGraphApi.getGraphDelta.mockResolvedValue(delta);

    await PreloadService.startPreload();
    expect(PreloadService.getPreloadedGraphData()?.lastHash).toBe("hash-1");

    const result = await PreloadService.updateWithDelta();

    expect(mockGraphApi.getGraphDelta).toHaveBeenCalledWith("hash-1");
    expect(result).toEqual(delta);
    expect(PreloadService.getPreloadedGraphDelta()).toEqual(delta);
    expect(PreloadService.getPreloadedGraphData()?.lastHash).toBe("hash-2");
  });

  it("handles public preload errors gracefully", async () => {
    mockGraphApi.getFullGraphData.mockRejectedValue(new Error("network"));

    await expect(PreloadService.startPreload()).resolves.toBeUndefined();
    expect(PreloadService.getPreloadedGraph()).toBeNull();
    expect(PreloadService.hasPreloadedData()).toBe(false);
  });

  it("exposes correct preload status and stats", async () => {
    expect(PreloadService.isPreloadingData()).toBe(false);
    expect(PreloadService.hasPreloadedData()).toBe(false);

    mockGraphApi.getFullGraphData.mockImplementation(
      () => new Promise((resolve) => setTimeout(() => resolve(mockGraphData), 100))
    );

    const promise = PreloadService.startPreload();
    expect(PreloadService.isPreloadingData()).toBe(true);

    await promise;

    const stats = PreloadService.getStats();
    expect(stats.hasGraph).toBe(true);
    expect(stats.hasAchievements).toBe(false);
    expect(stats.isPreloading).toBe(false);
    expect(stats.graphAge).toBe(0);
  });

  it("clears and invalidates caches", async () => {
    await PreloadService.startPreload();
    expect(PreloadService.hasPreloadedData()).toBe(true);

    PreloadService.clearCache();
    expect(PreloadService.hasPreloadedData()).toBe(false);
    expect(PreloadService.getPreloadedGraph()).toBeNull();

    await PreloadService.startPreload();
    expect(PreloadService.hasPreloadedData()).toBe(true);

    PreloadService.invalidateGraphCache();
    expect(PreloadService.getPreloadedGraph()).toBeNull();
    expect(PreloadService.hasPreloadedData()).toBe(false);
  });
});
