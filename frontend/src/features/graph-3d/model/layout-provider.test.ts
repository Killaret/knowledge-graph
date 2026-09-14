import { describe, it, expect, vi, beforeEach } from "vitest";
import {
  createLayoutProvider,
  D3ForceLayoutProvider,
  GraphServiceLayoutProvider,
} from "./layout-provider";
import { getFullGraphData, getGraphData } from "$shared/api/graph";
import type { Graph3DRuntimeConfig } from "../config";

vi.mock("$shared/api/graph", () => ({
  getFullGraphData: vi.fn(),
  getGraphData: vi.fn(),
}));

describe("createLayoutProvider", () => {
  beforeEach(() => {
    vi.mocked(getFullGraphData).mockReset();
    vi.mocked(getGraphData).mockReset();
  });

  it("returns GraphServiceLayoutProvider when layoutProvider is graph-service", () => {
    const runtime: Graph3DRuntimeConfig = {
      useGraphServiceLayout: true,
      warmStartTicks: 80,
      layoutProvider: "graph-service",
    };
    const provider = createLayoutProvider(runtime);
    expect(provider).toBeInstanceOf(GraphServiceLayoutProvider);
  });

  it("returns D3ForceLayoutProvider when layoutProvider is d3", () => {
    const runtime: Graph3DRuntimeConfig = {
      useGraphServiceLayout: false,
      warmStartTicks: 80,
      layoutProvider: "d3",
    };
    const provider = createLayoutProvider(runtime);
    expect(provider).toBeInstanceOf(D3ForceLayoutProvider);
  });

  it.each([
    ["d3", D3ForceLayoutProvider],
    ["graph-service", GraphServiceLayoutProvider],
  ] as const)(
    "PUB-2 %s provider forwards viewMode/nocache to getFullGraphData",
    async (layout, Provider) => {
      const provider = new Provider();
      vi.mocked(getFullGraphData).mockResolvedValue({ nodes: [], links: [] });
      await provider.load({ viewMode: "community", nocache: true, limit: 20, userId: "u1" });
      expect(getFullGraphData).toHaveBeenCalledWith(20, "u1", true, "community");
      expect(getGraphData).not.toHaveBeenCalled();
    }
  );

  it.each([
    ["d3", D3ForceLayoutProvider],
    ["graph-service", GraphServiceLayoutProvider],
  ] as const)("PUB-2 %s provider still calls getGraphData for noteId", async (layout, Provider) => {
    const provider = new Provider();
    vi.mocked(getGraphData).mockResolvedValue({ nodes: [], links: [] });
    await provider.load({ noteId: "n1", depth: 3, userId: "u1" });
    if (layout === "graph-service") {
      expect(getGraphData).toHaveBeenCalledWith("n1", 3, "u1", "3d");
    } else {
      expect(getGraphData).toHaveBeenCalledWith("n1", 3, "u1");
    }
    expect(getFullGraphData).not.toHaveBeenCalled();
  });
});
