import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, cleanup, waitFor } from "@testing-library/svelte";
import type { HomePageState } from "./";
import TestHomePageHost from "./__tests__/TestHomePageHost.svelte";
import { graphStore } from "$shared/stores/graph.svelte";
import { graphView } from "$shared/stores/graph-view.svelte";
import { goto } from "$app/navigation";
import { isAuthenticated } from "$shared/stores/auth.svelte";
import * as notesApi from "$shared/api/notes";
import * as linksApi from "$shared/api/links";
import { loadGraph } from "$shared/services/graphLoader";
import * as preloadHooks from "$features/preload/hooks/usePreloadedData";
import * as preloadService from "$shared/services/PreloadService";
import * as graph3d from "$features/graph-3d";
import { CelestialBody } from "$entities";

vi.mock("$app/environment", () => ({
  browser: true,
}));

vi.mock("$app/navigation", () => ({
  goto: vi.fn(),
}));

vi.mock("$shared/stores/auth.svelte", () => ({
  initAuth: vi.fn().mockResolvedValue(undefined),
  isAuthenticated: vi.fn().mockReturnValue(true),
}));

vi.mock("$shared/api/notes", () => ({
  getNotes: vi.fn(),
  getNote: vi.fn(),
  createNote: vi.fn(),
  deleteNote: vi.fn(),
  deleteNotesBatch: vi.fn(),
  restoreNote: vi.fn(),
}));

vi.mock("$shared/api/links", () => ({
  createLink: vi.fn(),
}));

vi.mock("$features/preload/hooks/usePreloadedData", () => ({
  getGraphWithPreload: vi.fn(),
}));

vi.mock("$shared/services/PreloadService", () => ({
  hasPreloadedData: vi.fn().mockReturnValue(false),
  updateGraphWithDelta: vi.fn(),
  getPreloadedGraph: vi.fn(),
}));

vi.mock("$shared/services/graphLoader", () => ({
  loadGraph: vi.fn(),
}));

vi.mock("$features/graph-3d", () => ({
  toRuntimeConfig: vi.fn().mockReturnValue({ layoutProvider: "d3" }),
  createLayoutProvider: vi.fn().mockReturnValue({
    load: vi.fn().mockResolvedValue({ nodes: [], links: [] }),
  }),
}));

const mockNotes = [
  {
    id: "n1",
    title: "Star Note",
    content: "content",
    metadata: { tags: [] },
    type: "star",
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-06-01T00:00:00Z",
  },
  {
    id: "n2",
    title: "Planet Note",
    content: "content",
    metadata: { tags: [] },
    type: "planet",
    created_at: "2024-02-01T00:00:00Z",
    updated_at: "2024-05-01T00:00:00Z",
  },
];

function setDefaultMocks() {
  vi.mocked(isAuthenticated).mockReturnValue(true);
  vi.mocked(notesApi.getNotes).mockResolvedValue(mockNotes as any);
  vi.mocked(notesApi.getNote).mockResolvedValue(mockNotes[0] as any);
  vi.mocked(notesApi.createNote).mockResolvedValue({ id: "n3" } as any);
  vi.mocked(notesApi.deleteNote).mockResolvedValue(undefined as any);
  vi.mocked(notesApi.deleteNotesBatch).mockResolvedValue(undefined as any);
  vi.mocked(notesApi.restoreNote).mockResolvedValue(undefined as any);
  vi.mocked(linksApi.createLink).mockResolvedValue(undefined as any);
  vi.mocked(preloadHooks.getGraphWithPreload).mockResolvedValue({
    nodes: [{ id: "n1", title: "Star Note", type: "star" }],
    links: [],
    hash: "hash1",
  } as any);
  vi.mocked(preloadService.hasPreloadedData).mockReturnValue(false);
  vi.mocked(preloadService.updateGraphWithDelta).mockResolvedValue(null);
  vi.mocked(preloadService.getPreloadedGraph).mockReturnValue({
    nodes: [{ id: "n1", title: "Star Note", type: "star" }],
    links: [],
    hash: "hash2",
  } as any);
  vi.mocked(loadGraph).mockResolvedValue({
    graph: { nodes: [{ id: "n1", title: "Star Note", type: "star" }], links: [], hash: "hash1" },
    notes: mockNotes,
    knowledgeCore: null,
  } as any);
  vi.mocked(graph3d.createLayoutProvider).mockReturnValue({
    load: vi.fn().mockResolvedValue({ nodes: [{ id: "n3" }], links: [] }),
  } as any);
}

describe("Home Page State", () => {
  beforeEach(() => {
    setDefaultMocks();
    (window as any).__SKIP_AUTH__ = true;
    localStorage.removeItem("graph-view-mode");
    graphView.restore();
    graphStore.currentView = "graph";
    graphStore.selectedNodeId = null;
    vi.useFakeTimers({ shouldAdvanceTime: true });
    window.confirm = vi.fn().mockReturnValue(true);
    window.alert = vi.fn();
  });

  afterEach(() => {
    cleanup();
    vi.useRealTimers();
    vi.clearAllMocks();
    delete (globalThis as unknown as { __TEST_HOME_PAGE?: HomePageState }).__TEST_HOME_PAGE;
  });

  async function getHomePage(): Promise<HomePageState> {
    render(TestHomePageHost);
    await waitFor(() => {
      const state = (globalThis as unknown as { __TEST_HOME_PAGE?: HomePageState })
        .__TEST_HOME_PAGE;
      expect(state).toBeDefined();
    });
    return (globalThis as unknown as { __TEST_HOME_PAGE: HomePageState }).__TEST_HOME_PAGE;
  }

  it("UI-LOAD-1: exposes the notes list before the graph resolves", async () => {
    let releaseGraph: ((v: any) => void) | null = null;
    vi.mocked(loadGraph).mockImplementation((options: any) => {
      // The loader hands the notes over as soon as they arrive — long before
      // the graph promise settles.
      options?.onNotesReady?.(mockNotes);
      return new Promise((resolve) => {
        releaseGraph = resolve;
      });
    });

    const homePage = await getHomePage();

    await waitFor(() => expect(homePage.allNotes.length).toBeGreaterThan(0));
    await waitFor(() => expect(homePage.filteredNotes.length).toBeGreaterThan(0));
    // The graph is still loading — but nothing blocks the notes anymore.
    expect(homePage.loading).toBe(true);
    expect(homePage.graphData.nodes).toHaveLength(0);

    releaseGraph!({
      graph: { nodes: [], links: [], hash: "hash1" },
      notes: mockNotes,
      knowledgeCore: null,
    });
    await waitFor(() => expect(homePage.loading).toBe(false));
  });

  it("loads data and exposes state", async () => {
    const homePage = await getHomePage();
    await waitFor(() => expect(homePage.loading).toBe(false));
    expect(homePage.allNotes.length).toBeGreaterThan(0);
    expect(homePage.typeFilters.length).toBeGreaterThan(0);
    expect(homePage.sortOptions.length).toBe(3);
  });

  it("filters, searches and sorts", async () => {
    const homePage = await getHomePage();
    await waitFor(() => expect(homePage.loading).toBe(false));

    homePage.handleFilter("star");
    expect(homePage.selectedType).toBe("star");
    expect(homePage.filteredNotes.every((n) => n.type === "star")).toBe(true);

    homePage.handleSearchQuery("Planet");
    expect(homePage.searchQuery).toBe("Planet");

    homePage.handleSortChange("updated");
    expect(homePage.sortBy).toBe("updated");
  });

  it("switches views and toggles layout provider", async () => {
    const homePage = await getHomePage();
    homePage.handleToggleView("list");
    expect(graphStore.currentView).toBe("list");

    homePage.handleToggleView("3d");
    expect(graphStore.currentView).toBe("3d");

    await homePage.handleToggleLayoutProvider("graph-service");
    expect(graph3d.createLayoutProvider).toHaveBeenCalled();
    expect(homePage.layoutProvider).toBe("graph-service");
  });

  it("manages selection and batch delete", async () => {
    const homePage = await getHomePage();
    await waitFor(() => expect(homePage.loading).toBe(false));

    homePage.toggleSelectionMode();
    expect(homePage.selectionMode).toBe(true);

    homePage.handleNoteSelect(homePage.allNotes[0], true);
    expect(homePage.selectedNoteIds.has(homePage.allNotes[0].id)).toBe(true);

    homePage.toggleSelectAll();
    expect(homePage.selectedNoteIds.size).toBe(homePage.filteredNotes.length);

    await homePage.handleBatchDelete();
    expect(notesApi.deleteNotesBatch).toHaveBeenCalled();
    expect(homePage.selectionMode).toBe(false);
  });

  it("creates and edits notes", async () => {
    const homePage = await getHomePage();

    homePage.handleCreateChildNote({ id: "p1", title: "Parent" });
    expect(homePage.showCreateModal).toBe(true);
    expect(homePage.createChildParent?.id).toBe("p1");

    await homePage.handleNoteCreate({ title: "Child", content: "body", type: "planet" });
    expect(notesApi.createNote).toHaveBeenCalled();

    await homePage.handleNoteCreated({ id: "new1", title: "Child" } as any);
    expect(homePage.showCreateModal).toBe(false);
    expect(linksApi.createLink).toHaveBeenCalled();

    homePage.handleNoteEdit(homePage.allNotes[0]);
    expect(homePage.showEditModal).toBe(true);

    homePage.handleEditSuccess();
    expect(homePage.showEditModal).toBe(false);
  });

  it("deletes and restores notes", async () => {
    const homePage = await getHomePage();
    await waitFor(() => expect(homePage.loading).toBe(false));

    homePage.handleDeleteRequest("n1");
    expect(homePage.showConfirmDelete).toBe(true);

    homePage.cancelDelete();
    expect(homePage.showConfirmDelete).toBe(false);

    const note = homePage.allNotes[0];
    await homePage.handleNoteDelete(note);
    expect(notesApi.deleteNote).toHaveBeenCalledWith(note.id);
    expect(homePage.showUndoToast).toBe(true);

    vi.advanceTimersByTime(1600);
    expect(homePage.undoToastStage).toBe("restore");

    homePage.handleUndoRestore();
    expect(notesApi.restoreNote).toHaveBeenCalledWith(note.id);
  });

  it("handles delete confirmation and reload", async () => {
    const homePage = await getHomePage();
    await waitFor(() => expect(homePage.loading).toBe(false));

    homePage.handleDeleteRequest("n1");
    await homePage.handleDeleteConfirm();
    expect(notesApi.deleteNote).toHaveBeenCalledWith("n1");
  });

  it("manages auth panel", async () => {
    const homePage = await getHomePage();

    homePage.openAuthPanel("register");
    expect(homePage.showAuthPanel).toBe(true);
    expect(homePage.authPanelTab).toBe("register");

    homePage.closeAuthPanel();
    expect(homePage.showAuthPanel).toBe(false);

    await homePage.handleAuthSuccess();
    expect(homePage.showAuthPanel).toBe(false);
  });

  it("imports, clears errors and resets child parent", async () => {
    const homePage = await getHomePage();

    homePage.handleImport();
    expect(goto).toHaveBeenCalledWith("/import");

    homePage.clearApiError();
    expect(homePage.apiError).toBeNull();

    homePage.handleCreateChildNote({ id: "p1", title: "Parent" });
    homePage.resetCreateChildParent();
    expect(homePage.createChildParent).toBeNull();
  });

  it("derives notes from public graph for anonymous users", async () => {
    vi.mocked(isAuthenticated).mockReturnValue(false);
    (window as any).__SKIP_AUTH__ = false;
    vi.mocked(loadGraph).mockResolvedValue({
      graph: { nodes: [{ id: "gn1", title: "Graph Note", type: "star" }], links: [], hash: "h" },
      notes: [],
      knowledgeCore: null,
    } as any);

    const homePage = await getHomePage();
    await waitFor(() => expect(homePage.loading).toBe(false));

    expect(homePage.allNotes.length).toBe(1);
    expect(homePage.allNotes[0].id).toBe("gn1");
  });

  it("sets api error when loading fails", async () => {
    vi.mocked(loadGraph).mockRejectedValue(new Error("boom"));

    const homePage = await getHomePage();
    await waitFor(() => expect(homePage.loading).toBe(false));

    expect(homePage.apiError).not.toBeNull();
  });

  it("applies graph delta on mutation refresh", async () => {
    vi.mocked(preloadService.hasPreloadedData).mockReturnValue(true);
    vi.mocked(preloadService.updateGraphWithDelta).mockResolvedValue({
      added_nodes: [{ id: "new" }],
      updated_nodes: [],
      removed_nodes: [],
      added_links: [],
      removed_links: [],
    } as any);
    vi.mocked(preloadService.getPreloadedGraph).mockReturnValue({
      nodes: [{ id: "n1" }, { id: "new" }],
      links: [],
      hash: "h2",
    } as any);
    vi.mocked(notesApi.getNotes).mockResolvedValue([
      ...mockNotes,
      { id: "new", title: "New", type: "star" } as any,
    ]);

    const homePage = await getHomePage();
    await waitFor(() => expect(homePage.loading).toBe(false));

    homePage.handleDeleteRequest("n1");
    await homePage.handleDeleteConfirm();

    expect(preloadService.updateGraphWithDelta).toHaveBeenCalled();
    expect(preloadService.getPreloadedGraph).toHaveBeenCalled();
    expect(notesApi.getNotes).toHaveBeenCalled();
    expect(homePage.graphData.nodes.length).toBe(2);
  });

  it("skips layout toggle when provider is unchanged", async () => {
    const homePage = await getHomePage();
    await homePage.handleToggleLayoutProvider("d3");
    expect(graph3d.createLayoutProvider).toHaveBeenCalledTimes(0);
  });

  it("shows api error when layout provider load fails", async () => {
    vi.mocked(graph3d.createLayoutProvider).mockReturnValue({
      load: vi.fn().mockRejectedValue(new Error("layout boom")),
    } as any);

    const homePage = await getHomePage();
    await homePage.handleToggleLayoutProvider("graph-service");

    expect(homePage.apiError).not.toBeNull();
  });

  it("cancels batch delete when user declines", async () => {
    const homePage = await getHomePage();
    await waitFor(() => expect(homePage.loading).toBe(false));

    homePage.toggleSelectionMode();
    homePage.handleNoteSelect(homePage.allNotes[0], true);

    window.confirm = vi.fn().mockReturnValue(false);
    await homePage.handleBatchDelete();

    expect(notesApi.deleteNotesBatch).not.toHaveBeenCalled();
  });

  it("creates note without child parent", async () => {
    const homePage = await getHomePage();

    await homePage.handleNoteCreate({ title: "Solo", content: "body", type: "star" });
    expect(notesApi.createNote).toHaveBeenCalled();

    homePage.resetCreateChildParent();
    await homePage.handleNoteCreated({ id: "solo", title: "Solo" } as any);
    expect(linksApi.createLink).not.toHaveBeenCalled();

    expect(homePage.createChildDefaultType).toBeUndefined();
  });

  it("suggests child default type from parent", async () => {
    const homePage = await getHomePage();

    homePage.handleCreateChildNote({ id: "p1", title: "Parent", type: "star" });
    expect(homePage.createChildDefaultType).toBeDefined();
  });

  it("handles note create error", async () => {
    vi.mocked(notesApi.createNote).mockRejectedValue(new Error("create fail"));

    const homePage = await getHomePage();

    await homePage.handleNoteCreate({ title: "Bad", content: "body", type: "star" });
    expect(notesApi.createNote).toHaveBeenCalled();
  });

  it("handles child link create error", async () => {
    vi.mocked(linksApi.createLink).mockRejectedValue(new Error("link fail"));

    const homePage = await getHomePage();

    homePage.handleCreateChildNote({ id: "p1", title: "Parent", type: "star" });
    await homePage.handleNoteCreated({ id: "child", title: "Child" } as any);

    expect(linksApi.createLink).toHaveBeenCalled();
    expect(homePage.createChildParent).toBeNull();
  });

  it("handles delete confirm error", async () => {
    vi.mocked(notesApi.deleteNote).mockRejectedValue(new Error("delete fail"));

    const homePage = await getHomePage();
    await waitFor(() => expect(homePage.loading).toBe(false));

    homePage.handleDeleteRequest("n1");
    await homePage.handleDeleteConfirm();

    expect(notesApi.deleteNote).toHaveBeenCalledWith("n1");
  });

  it("handles batch delete error", async () => {
    vi.mocked(notesApi.deleteNotesBatch).mockRejectedValue(new Error("batch fail"));

    const homePage = await getHomePage();
    await waitFor(() => expect(homePage.loading).toBe(false));

    homePage.toggleSelectionMode();
    homePage.handleNoteSelect(homePage.allNotes[0], true);
    await homePage.handleBatchDelete();

    expect(notesApi.deleteNotesBatch).toHaveBeenCalled();
  });

  it("handles undo restore error", async () => {
    vi.mocked(notesApi.restoreNote).mockRejectedValue(new Error("restore fail"));

    const homePage = await getHomePage();
    await waitFor(() => expect(homePage.loading).toBe(false));

    const note = homePage.allNotes[0];
    await homePage.handleNoteDelete(note);
    await homePage.handleUndoRestore();

    expect(notesApi.restoreNote).toHaveBeenCalledWith(note.id);
  });

  it("skips graph update when delta has no changes", async () => {
    vi.mocked(preloadService.hasPreloadedData).mockReturnValue(true);
    vi.mocked(preloadService.updateGraphWithDelta).mockResolvedValue({
      added_nodes: [],
      updated_nodes: [],
      removed_nodes: [],
      added_links: [],
      removed_links: [],
    } as any);
    vi.mocked(preloadService.getPreloadedGraph).mockReturnValue({
      nodes: [{ id: "n1" }],
      links: [],
      hash: "hash1",
    } as any);

    const homePage = await getHomePage();
    await waitFor(() => expect(homePage.loading).toBe(false));

    const graphBefore = homePage.graphData;
    homePage.handleDeleteRequest("n1");
    await homePage.handleDeleteConfirm();

    expect(preloadService.updateGraphWithDelta).toHaveBeenCalled();
    expect(homePage.graphData).toBe(graphBefore);
  });

  it("uses response.data from a structured API error", async () => {
    vi.mocked(loadGraph).mockRejectedValue({
      response: { data: { code: "CUSTOM", message: "custom error" } },
    });

    const homePage = await getHomePage();
    await waitFor(() => expect(homePage.loading).toBe(false));

    expect(homePage.apiError).toEqual({ code: "CUSTOM", message: "custom error" });
  });

  it("loads an empty graph without an error", async () => {
    vi.mocked(loadGraph).mockResolvedValue({
      graph: { nodes: [], links: [] },
      notes: [],
      knowledgeCore: null,
    } as any);

    const homePage = await getHomePage();
    await waitFor(() => expect(homePage.loading).toBe(false));

    expect(homePage.allNotes).toHaveLength(0);
    expect(homePage.graphData.nodes).toHaveLength(0);
  });

  it("silent refresh does not show loading overlay", async () => {
    const homePage = await getHomePage();
    await waitFor(() => expect(homePage.loading).toBe(false));

    vi.mocked(loadGraph).mockResolvedValue({
      graph: { nodes: [], links: [] },
      notes: [],
      knowledgeCore: null,
    } as any);

    await homePage.loadData({ silent: true });
    expect(homePage.loading).toBe(false);
  });

  it("keeps existing graph on silent refresh when new graph is empty", async () => {
    const homePage = await getHomePage();
    await waitFor(() => expect(homePage.loading).toBe(false));

    const existing = homePage.graphData;
    expect(existing.nodes.length).toBeGreaterThan(0);

    vi.mocked(loadGraph).mockResolvedValue({
      graph: { nodes: [], links: [] },
      notes: [],
      knowledgeCore: null,
    } as any);

    await homePage.loadData({ silent: true });
    expect(homePage.graphData).toBe(existing);
  });

  it("exposes UI type filters sorted from broad to narrow", async () => {
    const homePage = await getHomePage();
    const uiIds = homePage.typeFilters
      .filter((f) => CelestialBody.UI_TYPES.some((b) => b.type === f.id))
      .map((f) => f.id);
    expect(uiIds).toEqual([
      "galaxy",
      "nebula",
      "blackhole",
      "star",
      "planet",
      "moon",
      "comet",
      "satellite",
      "asteroid",
      "dust",
      "debris",
    ]);
    expect(homePage.typeFilters.some((f) => f.id === "unknown")).toBe(false);
  });

  it("hides undo toast and clears last deleted note after timeout", async () => {
    const homePage = await getHomePage();
    await waitFor(() => expect(homePage.loading).toBe(false));

    const note = homePage.allNotes[0];
    await homePage.handleNoteDelete(note);

    expect(homePage.showUndoToast).toBe(true);

    vi.advanceTimersByTime(6600);

    expect(homePage.showUndoToast).toBe(false);
    expect(homePage.lastDeletedNote).toBeNull();
  });

  it("skips background refresh when not authenticated", async () => {
    vi.mocked(isAuthenticated).mockReturnValue(false);
    vi.mocked(preloadService.hasPreloadedData).mockReturnValue(true);
    vi.mocked(preloadService.updateGraphWithDelta).mockResolvedValue({
      added_nodes: [{ id: "new" }],
    } as any);

    const homePage = await getHomePage();
    await waitFor(() => expect(homePage.loading).toBe(false));

    homePage.handleDeleteRequest("n1");
    await homePage.handleDeleteConfirm();

    expect(preloadService.updateGraphWithDelta).not.toHaveBeenCalled();
  });

  const scopedResult = (prefix: string, count: number) => ({
    graph: {
      nodes: Array.from({ length: count }, (_, i) => ({
        id: `${prefix}-${i}`,
        title: `${prefix} ${i}`,
        type: "star",
      })),
      links: [],
      hash: `${prefix}-hash`,
    },
    notes: Array.from({ length: count }, (_, i) => ({
      id: `${prefix}-${i}`,
      title: `${prefix} ${i}`,
      type: "star",
      content: "",
      metadata: {},
      created_at: "",
      updated_at: "",
    })),
    knowledgeCore: null,
  });

  it("PUB-2 switches scope and drops stale private data", async () => {
    vi.mocked(loadGraph).mockImplementation(({ viewMode }: { viewMode?: string }) =>
      Promise.resolve(
        scopedResult(
          viewMode === "community" ? "public" : "private",
          viewMode === "community" ? 10 : 20
        )
      )
    );

    const homePage = await getHomePage();
    await waitFor(() => expect(homePage.loading).toBe(false));
    expect(homePage.allNotes).toHaveLength(20);

    homePage.noteToEdit = "n1";
    homePage.showEditModal = true;
    homePage.showCreateModal = true;
    (homePage as any).lastDeletedNote = { id: "x" };
    (homePage as any).showUndoToast = true;

    graphView.mode = "community";

    await waitFor(() => expect(homePage.allNotes).toHaveLength(10));
    await waitFor(() => expect(homePage.loading).toBe(false));
    expect(homePage.allNotes.every((n: any) => n.id.startsWith("public"))).toBe(true);
    expect(homePage.noteToEdit).toBeNull();
    expect(homePage.showEditModal).toBe(false);
    expect(homePage.showCreateModal).toBe(false);
    expect((homePage as any).lastDeletedNote).toBeNull();
    expect((homePage as any).showUndoToast).toBe(false);

    const lastCall = vi.mocked(loadGraph).mock.calls.at(-1);
    expect((lastCall?.[0] as any).viewMode).toBe("community");
    expect((lastCall?.[0] as any).nocache).toBe(true);
  });

  it("PUB-2 does not let a deferred personal result overwrite a public switch", async () => {
    let resolveInitial: (value: any) => void = () => {};
    vi.mocked(loadGraph).mockImplementationOnce(() => new Promise((r) => (resolveInitial = r)));
    vi.mocked(loadGraph).mockImplementation(({ viewMode }: { viewMode?: string }) =>
      Promise.resolve(
        scopedResult(
          viewMode === "community" ? "public" : "private",
          viewMode === "community" ? 10 : 20
        )
      )
    );

    const homePage = await getHomePage();
    await waitFor(() => expect(vi.mocked(loadGraph)).toHaveBeenCalled());

    graphView.mode = "community";
    await waitFor(() => expect(homePage.allNotes).toHaveLength(10));
    await waitFor(() => expect(homePage.loading).toBe(false));

    resolveInitial(scopedResult("stale-private", 20));
    await waitFor(() => expect(homePage.allNotes).toHaveLength(10));
    expect(homePage.loading).toBe(false);
    expect(homePage.allNotes[0].id).toBe("public-0");
  });

  it("PUB-2 does not write old private notes over public data after focus", async () => {
    const private21 = Array.from({ length: 21 }, (_, i) => ({
      id: `private-${i}`,
      title: `Private ${i}`,
      type: "star",
      content: "",
      metadata: {},
      created_at: "",
      updated_at: "",
    }));
    let resolveNotes: (value: any) => void = () => {};

    vi.mocked(preloadService.hasPreloadedData).mockReturnValue(true);
    vi.mocked(preloadService.updateGraphWithDelta).mockResolvedValue({
      added_nodes: [{ id: "private-20", title: "Private 20", type: "star" }],
      updated_nodes: [],
      removed_nodes: [],
      added_links: [],
      removed_links: [],
    } as any);
    vi.mocked(preloadService.getPreloadedGraph).mockReturnValue({
      ...scopedResult("private", 21).graph,
      hash: "private-hash-v2",
    } as any);
    vi.mocked(notesApi.getNotes).mockImplementation(
      () =>
        new Promise((r) => {
          resolveNotes = r;
        })
    );

    vi.mocked(loadGraph).mockImplementation(({ viewMode }: { viewMode?: string }) =>
      Promise.resolve(
        scopedResult(
          viewMode === "community" ? "public" : "private",
          viewMode === "community" ? 10 : 20
        )
      )
    );

    const homePage = await getHomePage();
    await waitFor(() => expect(homePage.loading).toBe(false));
    expect(homePage.allNotes).toHaveLength(20);

    window.dispatchEvent(new Event("focus"));
    await waitFor(() => expect(notesApi.getNotes).toHaveBeenCalled());

    graphView.mode = "community";
    await waitFor(() => expect(homePage.allNotes).toHaveLength(10));
    await waitFor(() => expect(homePage.loading).toBe(false));
    expect(homePage.graphData.hash).toBe("public-hash");

    resolveNotes(private21);
    await Promise.resolve();
    await Promise.resolve();

    expect(homePage.allNotes).toHaveLength(10);
    expect(homePage.allNotes[0].id).toBe("public-0");
  });

  it("PUB-2 foreground load is not cancelled by a focus-driven refresh", async () => {
    let resolveInitial: (value: any) => void = () => {};
    vi.mocked(loadGraph).mockImplementationOnce(() => new Promise((r) => (resolveInitial = r)));

    const homePage = await getHomePage();
    expect(homePage.loading).toBe(true);

    window.dispatchEvent(new Event("focus"));
    resolveInitial(scopedResult("private", 20));

    await waitFor(() => expect(homePage.loading).toBe(false));
    expect(homePage.allNotes).toHaveLength(20);
    expect(homePage.graphData.hash).toBe("private-hash");
  });
});
