import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, waitFor } from "@testing-library/svelte";
import { readable } from "svelte/store";
import Page from "./[id]/+page.svelte";
import { goto } from "$app/navigation";
import { isGraph3DEnabled } from "$shared/config/config";

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

vi.mock("$shared/config/config", async (importOriginal) => ({
  ...(await importOriginal<typeof import("$shared/config/config")>()),
  isGraph3DEnabled: vi.fn(() => true),
}));

describe("Graph 3D focused page - auth ordering", () => {
  beforeEach(() => {
    loadMock.mockReset();
    initAuthMock.mockReset();
    vi.mocked(isGraph3DEnabled).mockReturnValue(true);
    loadMock.mockResolvedValue({ nodes: [], links: [] });
  });

  it("FREEZE-3D-1: redirects to the 2D focused graph when 3D is disabled", async () => {
    vi.mocked(isGraph3DEnabled).mockReturnValue(false);
    initAuthMock.mockResolvedValue(undefined);

    render(Page);

    await waitFor(() =>
      expect(vi.mocked(goto)).toHaveBeenCalledWith("/graph/note-1", { replaceState: true })
    );
    await Promise.resolve();
    expect(loadMock).not.toHaveBeenCalled();
  });

  it("does not call layoutProvider.load until initAuth resolves", async () => {
    let resolveInit!: () => void;
    initAuthMock.mockImplementation(() => new Promise<void>((resolve) => (resolveInit = resolve)));

    render(Page);

    await Promise.resolve();
    expect(loadMock).not.toHaveBeenCalled();

    resolveInit();
    await waitFor(() => expect(loadMock).toHaveBeenCalledTimes(1));
    expect(loadMock).toHaveBeenCalledWith({ noteId: "note-1", depth: 3 });
  });
});
