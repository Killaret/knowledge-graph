import { describe, it, expect, vi } from "vitest";
import {
  createLayoutProvider,
  D3ForceLayoutProvider,
  GraphServiceLayoutProvider,
} from "./layout-provider";
import type { Graph3DRuntimeConfig } from "../config";

vi.mock("$shared/api/graph", () => ({
  getFullGraphData: vi.fn(),
  getGraphData: vi.fn(),
}));

describe("createLayoutProvider", () => {
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
});
