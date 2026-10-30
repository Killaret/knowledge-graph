import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, waitFor, fireEvent, cleanup } from "@testing-library/svelte";
import Page from "./+page.svelte";
import { graphView } from "$shared/stores/graph-view.svelte";
import { authState } from "$shared/stores/auth-session.svelte";

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
    (window as any).__SKIP_AUTH__ = false;
    authState.currentUser = null;
    authState.accessToken = null;
    authState.apiKey = null;
    graphView.clear();
    localStorage.removeItem("graph-view-mode");
    loadMock.mockReset();
    initAuthMock.mockReset();
    loadMock.mockResolvedValue({ nodes: [], links: [] });
  });

  afterEach(() => {
    cleanup();
  });

  it("does not call layoutProvider.load until initAuth resolves", async () => {
    let resolveInit!: () => void;
    initAuthMock.mockImplementation(() => new Promise<void>((resolve) => (resolveInit = resolve)));

    render(Page);

    // Let onMount run: initAuth is pending, so no graph request may fire.
    await Promise.resolve();
    expect(loadMock).not.toHaveBeenCalled();

    resolveInit();
    await waitFor(() => expect(loadMock).toHaveBeenCalledTimes(1));
    expect(loadMock).toHaveBeenCalledWith({ nocache: false, viewMode: "community" });
  });

  it("loads immediately when initAuth is already resolved", async () => {
    initAuthMock.mockResolvedValue(undefined);

    render(Page);

    await waitFor(() => expect(loadMock).toHaveBeenCalledTimes(1));
  });

  it("PUB-2 switches to community with nocache and updates stats", async () => {
    (window as any).__SKIP_AUTH__ = true;
    authState.accessToken = "tok1";
    loadMock.mockImplementation(({ viewMode }: { viewMode?: string }) =>
      Promise.resolve(
        viewMode === "community"
          ? { nodes: [{ id: "public-node", title: "Public" }], links: [], hash: "public-hash" }
          : { nodes: [{ id: "private-node", title: "Private" }], links: [], hash: "private-hash" }
      )
    );

    initAuthMock.mockResolvedValue(undefined);
    render(Page);

    await waitFor(() => expect(loadMock).toHaveBeenCalledTimes(1), { timeout: 2000 });
    expect(loadMock).toHaveBeenLastCalledWith(expect.objectContaining({ viewMode: "personal" }));

    const community = screen.getByTestId("graph-view-community");
    await fireEvent.click(community);

    await waitFor(() => expect(screen.getByTestId("graph-stats")).toHaveTextContent("1"), {
      timeout: 2000,
    });
    expect(loadMock).toHaveBeenLastCalledWith(
      expect.objectContaining({ viewMode: "community", nocache: true })
    );
  });

  it("PUB-2 does not overwrite community scene with a stale personal response", async () => {
    (window as any).__SKIP_AUTH__ = true;
    authState.accessToken = "tok1";
    let resolvePersonal: (value: any) => void = () => {};
    loadMock
      .mockImplementationOnce(
        () =>
          new Promise((r) => {
            resolvePersonal = r;
          })
      )
      .mockImplementation(({ viewMode }: { viewMode?: string }) =>
        Promise.resolve(
          viewMode === "community"
            ? { nodes: [{ id: "public-node", title: "Public" }], links: [], hash: "public-hash" }
            : { nodes: [{ id: "private-node", title: "Private" }], links: [], hash: "private-hash" }
        )
      );

    initAuthMock.mockResolvedValue(undefined);
    render(Page);
    await waitFor(() => expect(loadMock).toHaveBeenCalled(), { timeout: 2000 });

    const community = screen.getByTestId("graph-view-community");
    await fireEvent.click(community);

    await waitFor(() => expect(screen.getByTestId("graph-stats")).toHaveTextContent("1"), {
      timeout: 2000,
    });
    resolvePersonal({ nodes: [{ id: "private-stale", title: "Stale" }], links: [], hash: "stale" });

    await waitFor(() => expect(loadMock).toHaveBeenCalledTimes(2), { timeout: 2000 });
    expect(screen.getByTestId("graph-stats")).toHaveTextContent("1");
    expect(screen.queryByText("Private")).not.toBeInTheDocument();
  });
});
