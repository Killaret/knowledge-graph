import { onMount, type Component } from "svelte";
import { goto } from "$app/navigation";
import { browser } from "$app/environment";
import {
  getNotes,
  createNote,
  deleteNote,
  deleteNotesBatch,
  restoreNote,
  type Note,
} from "$shared/api/notes";
import { createLink } from "$shared/api/links";
import { type GraphData } from "$shared/api/graph";
import { getGraphWithPreload } from "$features/preload/hooks/usePreloadedData";
import {
  hasPreloadedData,
  updateGraphWithDelta,
  getPreloadedGraph,
} from "$shared/services/PreloadService";
import { isAuthenticated, initAuth } from "$shared/stores/auth.svelte";
import { graphStore } from "$shared/stores/graph.svelte";
import { createLayoutProvider, toRuntimeConfig } from "$features/graph-3d";
import type { ErrorResponse } from "$shared/types/errors";
import { CelestialBody, FilterState } from "$entities";
import { formatMessage, getCurrentLocale, type MessageParams } from "$shared/utils/i18n";
import type { GraphNode, GraphLink } from "$shared/api/graph";

interface Graph3DViewerProps {
  nodes: GraphNode[];
  links: GraphLink[];
  centerNodeId?: string | null;
  selectedNodeId?: string | null;
  onNodeClick?: (node: { id: string; title: string; type?: string }) => void;
  onNodeDoubleClick?: (node: { id: string; title: string; type?: string }) => void;
}

const locale = getCurrentLocale();
const t = (key: string, params?: MessageParams) => formatMessage(key, locale, params);

export function createHomePageState() {
  function toErrorResponse(e: unknown): ErrorResponse {
    if (e && typeof e === "object") {
      const err = e as { response?: { data?: ErrorResponse } };
      if (err.response?.data) {
        return err.response.data;
      }
    }
    return { code: "LOAD_ERROR", message: t("notes.loadError") };
  }

  // State module
  let allNotes = $state<Note[]>([]);
  let filteredNotes = $state<Note[]>([]);
  let loading = $state(true);
  let apiError = $state<ErrorResponse | null>(null);
  let showCreateModal = $state(false);
  let createChildParent = $state<{ id: string; title: string; type?: string } | null>(null);
  let showEditModal = $state(false);
  let noteToEdit = $state<string | null>(null);
  let showConfirmDelete = $state(false);
  let noteToDelete = $state<string | null>(null);
  let Graph3DViewer = $state<Component<Graph3DViewerProps> | null>(null);

  $effect(() => {
    if (graphStore.currentView === "3d" && !Graph3DViewer) {
      import("$widgets/graph-3d-viewer/Graph3DViewer.svelte")
        .then((mod) => {
          Graph3DViewer = mod.default;
        })
        .catch((e) => {
          if (import.meta.env.DEV) {
            console.error("[HomePage] Failed to load Graph3DViewer:", e);
          }
        });
    }
  });

  // Graph state - always show full graph on main page
  let graphData = $state<GraphData>({ nodes: [], links: [] });
  let layoutProvider = $state<"d3" | "graph-service">(toRuntimeConfig().layoutProvider);
  let searchQuery = $state("");

  // Filter and sort state
  let selectedType = $state<string>("all");
  let sortBy = $state<"created" | "updated" | "type">("created");
  let filterState = $state(
    new FilterState({
      selectedType: "all",
      sortBy: "created",
      searchQuery: "",
      currentView: "graph",
    })
  );

  $effect(() => {
    filterState = new FilterState({
      selectedType,
      sortBy,
      searchQuery,
      currentView: graphStore.currentView,
    });
  });

  $effect(() => {
    filteredNotes = filterState.applyFiltersAndSort(allNotes, getNoteType);
  });

  const selectedNoteIds = $state<Set<string>>(new Set());
  let selectionMode = $state(false);
  let lastDeletedNote = $state<Note | null>(null);
  let showUndoToast = $state(false);
  let undoToastStage = $state<"done" | "restore">("done");
  let showBulkActionsMenu = $state(false);
  let showAuthPanel = $state(false);
  let authPanelTab = $state<"login" | "register">("login");

  function openAuthPanel(tab: "login" | "register") {
    authPanelTab = tab;
    showAuthPanel = true;
  }

  function closeAuthPanel() {
    showAuthPanel = false;
  }

  function handleAuthSuccess() {
    showAuthPanel = false;
    void loadData({ silent: true });
  }

  function filterLabel(body: CelestialBody): string {
    const key = `filter.type.${body.type}`;
    const localized = t(key);
    return localized === key ? body.label : localized;
  }

  const typeFilters = [
    {
      id: "inbox",
      label: t("filter.inbox"),
      emoji: "📥",
      description: t("filter.inbox.description"),
    },
    {
      id: "all",
      label: t("filter.all"),
      emoji: "🌌",
      description: t("filter.all.description"),
    },
    ...[
      "star",
      "planet",
      "moon",
      "comet",
      "galaxy",
      "nebula",
      "asteroid",
      "satellite",
      "blackhole",
      "dust",
      "unknown",
    ].map((id) => {
      const body = CelestialBody.fromString(id);
      return {
        id,
        label: filterLabel(body),
        emoji: body.emoji,
        description: body.description,
        example: body.example,
      };
    }),
  ];

  const sortOptions = [
    { id: "created", label: t("sort.created") },
    { id: "updated", label: t("sort.updated") },
    { id: "type", label: t("sort.type") },
  ];

  onMount(() => {
    if (!browser) return;

    let deltaInterval: ReturnType<typeof setInterval> | undefined;
    let handleFocus: (() => void) | undefined;

    (async () => {
      await initAuth();
      void loadData();

      // Periodically sync the graph via delta updates from graph-service.
      // Public/anonymous users have no writeable graph to sync and the
      // /v1/graph/delta endpoint requires authentication, so polling it would
      // fire 401 -> auth/refresh 400 cycles and trigger unnecessary re-renders.
      if (isAuthenticated()) {
        deltaInterval = setInterval(() => {
          void refreshAfterMutation();
        }, 30000);

        handleFocus = () => {
          void refreshAfterMutation();
        };
        window.addEventListener("focus", handleFocus);
      }

      // Expose flag for E2E tests to assert background sync is gated by auth.
      if (browser) {
        (window as unknown as Record<string, unknown>).__kgGraphPollingActive = !!isAuthenticated();
      }
    })();

    return () => {
      if (deltaInterval) clearInterval(deltaInterval);
      if (handleFocus) window.removeEventListener("focus", handleFocus);
    };
  });

  /**
   * Single source of truth load: graph-service full graph + notes.
   * For unauthenticated users the note list is derived from the public graph.
   */
  async function loadData({ silent = false }: { silent?: boolean } = {}) {
    try {
      apiError = null;
      if (!silent) {
        loading = true;
      }

      const isAuth = isAuthenticated();
      let graphResult: GraphData | null = null;

      if (isAuth) {
        const [notesResult, loadedGraph] = await Promise.all([getNotes(), getGraphWithPreload()]);
        allNotes = notesResult;
        graphResult = loadedGraph;

        // The graph-service layout cache can lag behind note mutations (especially in test stacks).
        // Make sure every note returned by the backend is represented in the graph so filters and
        // counts stay consistent.
        if (graphResult) {
          const nodeIds = new Set(graphResult.nodes.map((n) => n.id));
          for (const note of allNotes) {
            if (!nodeIds.has(note.id)) {
              graphResult.nodes.push({
                id: note.id,
                title: note.title,
                type: note.type || "unknown",
                x: 0,
                y: 0,
                z: 0,
                size: 1,
              });
              nodeIds.add(note.id);
            }
          }
        }
      } else {
        graphResult = await getGraphWithPreload();
        // Public graph is the single source of notes for anonymous users.
        if (graphResult && graphResult.nodes.length > 0) {
          allNotes = graphResult.nodes.map((n) => ({
            id: n.id,
            title: n.title,
            content: "",
            metadata: {},
            type: n.type || "unknown",
            created_at: "",
            updated_at: "",
          }));
        } else {
          allNotes = [];
        }
      }

      applyFiltersAndSort();

      if (graphResult && graphResult.nodes.length > 0) {
        if (import.meta.env.DEV) {
          const apiTypes = [...new Set(graphResult.nodes.map((n) => n.type))];
          console.log(
            "[+page] Full graph loaded:",
            graphResult.nodes.length,
            "nodes,",
            graphResult.links.length,
            "links"
          );
          console.log("[+page] API node types:", apiTypes);
        }

        graphData = graphResult;
      } else if (allNotes.length > 0) {
        // Fallback when graph-service returns empty but notes exist. This keeps
        // the UI usable and makes E2E tests stable regardless of graph-service state.
        graphData = {
          nodes: allNotes.map((n) => ({
            id: n.id,
            title: n.title,
            type: n.type || "unknown",
          })),
          links: [],
        };
      } else if (!silent || graphData.nodes.length === 0) {
        // Minimal fallback for empty/invalid graph-service result. When
        // refreshing silently in the background, keep showing the last known
        // graph instead of blanking it out on a transient empty response.
        graphData = { nodes: [], links: [] };
      }
    } catch (e: unknown) {
      // Background refreshes shouldn't surface a full-page error and wipe
      // out an already-rendered graph; just log and keep the current state.
      if (!silent) {
        apiError = toErrorResponse(e);
      }
      if (import.meta.env.DEV) {
        console.error(e);
      }
    } finally {
      if (!silent) {
        loading = false;
      }
    }
  }

  /**
   * After mutations try to apply a graph delta first; if no hash/preloaded data
   * or delta fails, fall back to a full reload. This always runs silently so
   * the graph doesn't flash a loading overlay on periodic background syncs.
   *
   * Anonymous users have no private graph to sync; /v1/graph/delta requires
   * auth and would return 401, so we skip background sync entirely for them.
   */
  async function refreshAfterMutation() {
    if (!isAuthenticated()) {
      return;
    }

    if (graphData.hash && hasPreloadedData()) {
      const previousHash = graphData.hash;
      const delta = await updateGraphWithDelta();
      if (delta) {
        const updated = getPreloadedGraph();
        const hasChanges =
          (delta.added_nodes?.length ?? 0) > 0 ||
          (delta.updated_nodes?.length ?? 0) > 0 ||
          (delta.removed_nodes?.length ?? 0) > 0 ||
          (delta.added_links?.length ?? 0) > 0 ||
          (delta.removed_links?.length ?? 0) > 0;
        if (updated && (hasChanges || updated.hash !== previousHash)) {
          graphData = updated;
          // The list view derives from allNotes, not graphData, so we must
          // refresh the notes list after a delta update.
          allNotes = await getNotes();
          applyFiltersAndSort();
        }
        return;
      }
    }
    await loadData({ silent: true });
  }

  // Helper to get note type - unified with renderer.ts logic via CelestialBody
  function getNoteType(note: Note): string {
    return CelestialBody.fromString(note.type).type;
  }

  // Reactive filtered graph data based on filter state
  const filteredGraphData = $derived(filterState.filterGraphData(graphData, allNotes, getNoteType));

  function applyFiltersAndSort() {
    filteredNotes = filterState.applyFiltersAndSort(allNotes, getNoteType);
  }

  function handleSearch() {
    // Keep list and graph search consistent with client-side substring filtering.
    applyFiltersAndSort();
  }

  function handleSearchQuery(query: string) {
    filterState = filterState.with({ searchQuery: query });
    searchQuery = query;
    handleSearch();
  }

  function handleFilter(type: string) {
    filterState = filterState.with({ selectedType: type });
    selectedType = type;
    applyFiltersAndSort();
  }

  function handleSortChange(next: "created" | "updated" | "type") {
    sortBy = next;
    applyFiltersAndSort();
  }

  function handleDeleteRequest(id: string) {
    noteToDelete = id;
    showConfirmDelete = true;
  }

  async function handleDeleteConfirm() {
    if (!noteToDelete) return;

    try {
      await deleteNote(noteToDelete);
      graphStore.selectedNodeId = null;
      // Remove deleted note from local arrays immediately
      allNotes = allNotes.filter((n) => n.id !== noteToDelete);
      filteredNotes = filteredNotes.filter((n) => n.id !== noteToDelete);
      // Then reload from server to ensure sync
      await refreshAfterMutation();
    } catch {
      if (browser) {
        alert(t("page.deleteError"));
      }
    } finally {
      noteToDelete = null;
      showConfirmDelete = false;
    }
  }

  function toggleSelectionMode() {
    selectionMode = !selectionMode;
    if (!selectionMode) {
      selectedNoteIds.clear();
    }
  }

  function toggleSelectAll() {
    if (selectedNoteIds.size === filteredNotes.length) {
      selectedNoteIds.clear();
    } else {
      filteredNotes.forEach((n) => selectedNoteIds.add(n.id));
    }
  }

  function handleNoteSelect(note: Note, selected: boolean) {
    if (selected) {
      selectedNoteIds.add(note.id);
    } else {
      selectedNoteIds.delete(note.id);
    }
  }

  async function handleBatchDelete() {
    if (selectedNoteIds.size === 0) return;

    // Confirmation dialog
    if (browser) {
      const confirmed = confirm(
        `Delete ${selectedNoteIds.size} selected note${selectedNoteIds.size > 1 ? "s" : ""}? This action cannot be undone.`
      );
      if (!confirmed) return;
    }

    try {
      await deleteNotesBatch(Array.from(selectedNoteIds));
      // Remove deleted notes from local arrays
      allNotes = allNotes.filter((n) => !selectedNoteIds.has(n.id));
      filteredNotes = filteredNotes.filter((n) => !selectedNoteIds.has(n.id));
      selectedNoteIds.clear();
      selectionMode = false;
      showBulkActionsMenu = false;
      // Reload to ensure sync
      await refreshAfterMutation();
    } catch {
      if (browser) {
        alert(t("page.batchDeleteError"));
      }
    }
  }

  async function handleNoteEdit(note: Note) {
    noteToEdit = note.id;
    showEditModal = true;
  }

  async function handleNoteDelete(note: Note) {
    lastDeletedNote = note;
    await deleteNote(note.id);
    allNotes = allNotes.filter((n) => n.id !== note.id);
    filteredNotes = filteredNotes.filter((n) => n.id !== note.id);
    await refreshAfterMutation();
    showUndoToast = true;
    undoToastStage = "done";
    setTimeout(() => {
      undoToastStage = "restore";
    }, 1500);
    setTimeout(() => {
      showUndoToast = false;
      lastDeletedNote = null;
    }, 6500);
  }

  async function handleUndoRestore() {
    if (!lastDeletedNote) return;
    try {
      await restoreNote(lastDeletedNote.id);
      showUndoToast = false;
      lastDeletedNote = null;
      await refreshAfterMutation();
    } catch {
      if (browser) {
        alert(t("page.restoreError"));
      }
    }
  }

  async function handleNoteCreate(data: { title: string; content: string; type: string }) {
    try {
      await createNote({
        title: data.title,
        content: data.content,
        type: data.type,
      });
      await refreshAfterMutation();
    } catch {
      if (browser) {
        alert(t("note.createError"));
      }
    }
  }

  async function handleNoteCreated(note: Note) {
    showCreateModal = false;
    if (createChildParent) {
      try {
        await createLink({
          source_note_id: createChildParent.id,
          target_note_id: note.id,
          link_type: "parent",
          weight: 0.9,
        });
      } catch {
        if (browser) {
          alert(t("note.createChildLinkError"));
        }
      }
      createChildParent = null;
    }
    graphStore.selectedNodeId = note.id;
    await refreshAfterMutation();
  }

  function handleCreateChildNote(parent: { id: string; title: string; type?: string }) {
    createChildParent = parent;
    showCreateModal = true;
  }

  function resetCreateChildParent() {
    createChildParent = null;
  }

  function handleToggleView(view: "graph" | "list" | "3d") {
    graphStore.currentView = view;
  }

  async function handleToggleLayoutProvider(provider: "d3" | "graph-service") {
    if (provider === layoutProvider) return;
    layoutProvider = provider;
    try {
      const runtime = { ...toRuntimeConfig(), layoutProvider: provider };
      graphData = await createLayoutProvider(runtime).load({ limit: 100 });
    } catch (e) {
      apiError = toErrorResponse(e);
      if (import.meta.env.DEV) {
        console.error("Failed to load 3D graph with layout provider:", provider, e);
      }
    }
  }

  function handleImport() {
    void goto("/import");
  }

  function clearApiError() {
    apiError = null;
  }

  function cancelDelete() {
    showConfirmDelete = false;
    noteToDelete = null;
  }

  function handleEditSuccess() {
    showEditModal = false;
    noteToEdit = null;
  }

  return {
    get allNotes() {
      return allNotes;
    },
    get filteredNotes() {
      return filteredNotes;
    },
    get loading() {
      return loading;
    },
    set loading(value) {
      loading = value;
    },
    get apiError(): ErrorResponse | null {
      return apiError;
    },
    get showCreateModal() {
      return showCreateModal;
    },
    set showCreateModal(value) {
      showCreateModal = value;
    },
    get createChildParent() {
      return createChildParent;
    },
    set createChildParent(value) {
      createChildParent = value;
    },
    get showEditModal() {
      return showEditModal;
    },
    set showEditModal(value) {
      showEditModal = value;
    },
    get noteToEdit() {
      return noteToEdit;
    },
    set noteToEdit(value) {
      noteToEdit = value;
    },
    get showConfirmDelete() {
      return showConfirmDelete;
    },
    set showConfirmDelete(value) {
      showConfirmDelete = value;
    },
    get noteToDelete() {
      return noteToDelete;
    },
    set noteToDelete(value) {
      noteToDelete = value;
    },
    get Graph3DViewer() {
      return Graph3DViewer;
    },
    get graphData() {
      return graphData;
    },
    get layoutProvider() {
      return layoutProvider;
    },
    get searchQuery() {
      return searchQuery;
    },
    set searchQuery(value) {
      searchQuery = value;
    },
    get selectedType() {
      return selectedType;
    },
    set selectedType(value) {
      selectedType = value;
    },
    get sortBy() {
      return sortBy;
    },
    set sortBy(value) {
      sortBy = value;
    },
    get filterState() {
      return filterState;
    },
    get selectedNoteIds() {
      return selectedNoteIds;
    },
    get selectionMode() {
      return selectionMode;
    },
    set selectionMode(value) {
      selectionMode = value;
    },
    get lastDeletedNote() {
      return lastDeletedNote;
    },
    set lastDeletedNote(value) {
      lastDeletedNote = value;
    },
    get showUndoToast() {
      return showUndoToast;
    },
    set showUndoToast(value) {
      showUndoToast = value;
    },
    get undoToastStage() {
      return undoToastStage;
    },
    set undoToastStage(value) {
      undoToastStage = value;
    },
    get showBulkActionsMenu() {
      return showBulkActionsMenu;
    },
    set showBulkActionsMenu(value) {
      showBulkActionsMenu = value;
    },
    get showAuthPanel() {
      return showAuthPanel;
    },
    set showAuthPanel(value) {
      showAuthPanel = value;
    },
    get authPanelTab() {
      return authPanelTab;
    },
    set authPanelTab(value) {
      authPanelTab = value;
    },
    get filteredGraphData() {
      return filteredGraphData;
    },
    get typeFilters() {
      return typeFilters;
    },
    get sortOptions() {
      return sortOptions;
    },
    get createChildDefaultType() {
      return createChildParent
        ? CelestialBody.getChildSuggestion(createChildParent.type)
        : undefined;
    },

    t,

    openAuthPanel,
    closeAuthPanel,
    handleAuthSuccess,
    loadData,
    applyFiltersAndSort,
    handleSearch,
    handleSearchQuery,
    handleFilter,
    handleSortChange,
    handleDeleteRequest,
    handleDeleteConfirm,
    toggleSelectionMode,
    toggleSelectAll,
    handleNoteSelect,
    handleBatchDelete,
    handleNoteEdit,
    handleNoteDelete,
    handleUndoRestore,
    handleNoteCreate,
    handleNoteCreated,
    handleCreateChildNote,
    resetCreateChildParent,
    handleToggleView,
    handleToggleLayoutProvider,
    handleImport,
    clearApiError,
    cancelDelete,
    handleEditSuccess,
  };
}

export type HomePageState = ReturnType<typeof createHomePageState>;
