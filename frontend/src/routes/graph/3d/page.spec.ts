import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, waitFor } from "@testing-library/svelte";
import Page from "./+page.svelte";

// Regression guard: the 3D graph page must wait for initAuth() before
// calling the layout provider. Without the await, the graph request fires
// while a session refresh is still in flight, isAuthenticated() is false,
// and the scene renders the anonymous public graph even though the full
// graph is fetched later (found via auth-setup review: graph/full returned
// 100 nodes but the scene stayed at the public 20).

const { loadMock, initAuthMock } = vi.hoisted(() => ({
  loadMock: vi.fn(),
  initAuthMock: vi.fn(),
}));

vi.mock("$features/graph-3d", () => ({
  createLayoutProvider: vi.fn(() => ({ load: loadMock })),
  toRuntimeConfig: vi.fn(() => ({ layoutProvider: "d3" })),
}));

vi.mock("$shared/stores/auth.svelte", async () => {
  const actual = await vi.importActual<typeof import("$shared/stores/auth.svelte")>(
    "$shared/stores/auth.svelte"
  );
  return {
    ...actual,
    initAuth: initAuthMock,
  };
});

vi.mock("$widgets/graph-3d-viewer/Graph3DViewer.svelte", () => ({
  default: null,
}));

describe("Graph 3D page - auth ordering", () => {
  beforeEach(() => {
    loadMock.mockReset();
    initAuthMock.mockReset();
    loadMock.mockResolvedValue({ nodes: [], links: [] });
  });

  it("does not call layoutProvider.load until initAuth resolves", async () => {
    let resolveInit!: () => void;
    initAuthMock.mockImplementation(
      () => new Promise<void>((resolve) => (resolveInit = resolve))
    );

    render(Page);

    // Let onMount run: initAuth is pending, so no graph request may fire.
    await Promise.resolve();
    expect(loadMock).not.toHaveBeenCalled();

    resolveInit();
    await waitFor(() => expect(loadMock).toHaveBeenCalledTimes(1));
    expect(loadMock).toHaveBeenCalledWith({});
  });

  it("loads immediately when initAuth is already resolved", async () => {
    initAuthMock.mockResolvedValue(undefined);

    render(Page);

    await waitFor(() => expect(loadMock).toHaveBeenCalledTimes(1));
  });
});
