import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, cleanup } from "@testing-library/svelte";
import { tick } from "svelte";

const { isPreloadingData, hasPreloadedData, getPreloadedGraph } = vi.hoisted(() => ({
  isPreloadingData: vi.fn<() => boolean>(() => false),
  hasPreloadedData: vi.fn<() => boolean>(() => false),
  getPreloadedGraph: vi.fn<() => { nodes: unknown[]; links: unknown[] } | null>(() => null),
}));

vi.mock("$shared/services/PreloadService", () => ({
  isPreloadingData,
  hasPreloadedData,
  getPreloadedGraph,
}));

import PreloadIndicator from "./PreloadIndicator.svelte";

describe("PreloadIndicator Component", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.mocked(isPreloadingData).mockReturnValue(false);
    vi.mocked(hasPreloadedData).mockReturnValue(false);
    vi.mocked(getPreloadedGraph).mockReturnValue(null);
  });

  afterEach(() => {
    cleanup();
    vi.useRealTimers();
    vi.clearAllMocks();
  });

  it("shows loading state when preloading", async () => {
    vi.mocked(isPreloadingData).mockReturnValue(true);

    const { getByText } = render(PreloadIndicator);
    await tick();

    expect(getByText("Preparing the universe...")).toBeTruthy();
  });

  it("shows ready state with preloaded star count", async () => {
    vi.mocked(hasPreloadedData).mockReturnValue(true);
    vi.mocked(getPreloadedGraph).mockReturnValue({
      nodes: [{ id: "1" }, { id: "2" }, { id: "3" }],
      links: [],
    });

    const { getByText } = render(PreloadIndicator);
    await tick();

    expect(getByText("Universe ready: 3 stars")).toBeTruthy();
  });

  it("renders nothing when no preload activity", async () => {
    const { container } = render(PreloadIndicator);
    await tick();

    expect(container.textContent?.trim()).toBe("");
  });
});
