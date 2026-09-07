import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, waitFor } from "@testing-library/svelte";
import { readable } from "svelte/store";
import Page from "./[id]/+page.svelte";

// Same regression guard as ../page.spec.ts: the focused 3D page must wait
// for initAuth() before loading, otherwise an in-flight session refresh
// leaves isAuthenticated() false and the scene renders the public graph.

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

vi.mock("$app/stores", () => ({
  page: readable({
    url: new URL("http://localhost/graph/3d/note-1"),
    params: { id: "note-1" },
    route: { id: "/graph/3d/[id]" },
    status: 200,
    error: null,
    data: {},
    form: undefined,
  }),
  updated: { subscribe: vi.fn(() => () => {}), check: vi.fn() },
}));

vi.mock("$widgets/graph-3d-viewer/Graph3DViewer.svelte", () => ({
  default: null,
}));

describe("Graph 3D focused page - auth ordering", () => {
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

    await Promise.resolve();
    expect(loadMock).not.toHaveBeenCalled();

    resolveInit();
    await waitFor(() => expect(loadMock).toHaveBeenCalledTimes(1));
    expect(loadMock).toHaveBeenCalledWith({ noteId: "note-1", depth: 3 });
  });
});
