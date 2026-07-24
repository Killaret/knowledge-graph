<script lang="ts">
  import { onMount } from "svelte";
  import { browser } from "$app/environment";
  import FloatingControls from "$components/organisms/FloatingControls.svelte";
  import NoteSidePanel from "$components/organisms/NoteSidePanel.svelte";
  import CreateNoteModal from "$components/organisms/CreateNoteModal.svelte";
  import EditNoteModal from "$components/organisms/EditNoteModal.svelte";
  import ConfirmModal from "$components/organisms/ConfirmModal.svelte";
  import NoteCard from "$components/molecules/NoteCard.svelte";
  import ApiErrorDisplay from "$components/atoms/ApiErrorDisplay.svelte";
  import StateIllustration from "$components/atoms/StateIllustration.svelte";
  import {
    getNotes,
    createNote,
    deleteNote,
    deleteNotesBatch,
    restoreNote,
    searchNotes,
    type Note,
  } from "$shared/api/notes";
  import {
    getFullGraphData,
    getFreshGraph,
    type GraphData,
    type GraphDeltaData,
    type GraphNode,
    type GraphLink,
  } from "$shared/api/graph";
  import { apiConfig } from "$shared/config/config";
  import { getGraphWithPreload } from "$shared/hooks/usePreloadedData";
  import { isAuthenticated } from "$shared/stores/auth.svelte";
  import GraphCanvas from "$components/organisms/GraphCanvas.svelte";
  import type { ErrorResponse } from "$shared/types/errors";
  import SplashScreen from "$components/atoms/SplashScreen.svelte";
  import { CelestialBody, FilterState } from "$entities";
  import {
    formatMessage,
    getCurrentLocale,
    type MessageParams,
  } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string, params?: MessageParams) =>
    formatMessage(key, locale, params);

  // Raw API shapes that backend may return with alternative casing
  interface RawNode {
    id?: string;
    Id?: string;
    ID?: string;
    title?: string;
    Title?: string;
    type?: string;
    Type?: string;
    x?: number;
    y?: number;
    z?: number;
    size?: number;
  }

  interface RawLink {
    source?: string;
    source_note_id?: string;
    target?: string;
    target_note_id?: string;
    weight?: number;
    link_type?: string;
  }

  function normalizeNode(raw: RawNode): GraphNode {
    return {
      id: raw.id ?? raw.Id ?? raw.ID ?? "",
      title: raw.title ?? raw.Title ?? "",
      type: raw.type ?? raw.Type ?? "unknown",
      x: raw.x,
      y: raw.y,
      z: raw.z,
      size: raw.size,
    };
  }

  function normalizeLink(raw: RawLink): GraphLink {
    return {
      source: raw.source_note_id ?? raw.source ?? "",
      target: raw.target_note_id ?? raw.target ?? "",
      weight: raw.weight,
      link_type: raw.link_type,
    };
  }

  function toErrorResponse(e: unknown): ErrorResponse {
    if (e && typeof e === "object") {
      const err = e as { response?: { data?: ErrorResponse } };
      if (err.response?.data) {
        return err.response.data;
      }
    }
    return { code: "LOAD_ERROR", message: t("notes.loadError") };
  }

  // State
  let allNotes: Note[] = $state([]);
  let filteredNotes: Note[] = $state([]);
  let loading = $state(true);
  let apiError = $state<ErrorResponse | null>(null);
  let selectedNodeId: string | null = $state(null);
  let showCreateModal = $state(false);
  let showEditModal = $state(false);
  let noteToEdit: string | null = $state(null);
  let showConfirmDelete = $state(false);
  let noteToDelete: string | null = $state(null);
  let currentView: "graph" | "list" = $state("graph"); // Graph-first interface

  // Graph state - always show full graph on main page
  let graphData: GraphData = $state({ nodes: [], links: [] });
  let graphDelta: GraphDeltaData | undefined = $state(undefined);
  let graphLoading = $state(false);
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
    }),
  );
  $effect(() => {
    filterState = new FilterState({
      selectedType,
      sortBy,
      searchQuery,
      currentView,
    });
  });

  $effect(() => {
    filteredNotes = filterState.applyFiltersAndSort(allNotes, getNoteType);
  });
  const selectedNoteIds = $state<Set<string>>(new Set());
  let selectionMode = $state(false);
  let lastDeletedNote: Note | null = $state(null);
  let showUndoToast = $state(false);
  let undoToastStage = $state<"done" | "restore">("done");
  let showBulkActionsMenu = $state(false);

  function filterLabel(body: CelestialBody): string {
    const key = `filter.type.${body.type}`;
    const localized = t(key);
    return localized === key ? body.label : localized;
  }

  const typeFilters = [
    { id: "inbox", label: t("filter.inbox"), emoji: "📥" },
    { id: "all", label: t("filter.all"), emoji: "�" },
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
      return { id, label: filterLabel(body), emoji: body.emoji };
    }),
  ];

  const sortOptions = [
    { id: "created", label: t("sort.created") },
    { id: "updated", label: t("sort.updated") },
    { id: "type", label: t("sort.type") },
  ];

  // NOTE: The sortOptions constant was previously defined here but is not currently used.
  // These are the available sorting options for the notes list view:
  // - newest: Sort by creation date, newest first
  // - oldest: Sort by creation date, oldest first
  // - az: Alphabetical sorting A-Z
  // - za: Alphabetical sorting Z-A
  onMount(async () => {
    if (!browser) return;
    await loadDataParallel();
  });

  async function loadDataParallel() {
    try {
      // Reset error state before loading
      apiError = null;

      const isAuth = isAuthenticated();
      const graphTimeoutMs = 10000;

      // Start notes and graph loading in parallel
      const notesPromise = isAuth ? getNotes() : Promise.resolve([] as Note[]);
      const graphPromise = isAuth
        ? Promise.race([
            getFreshGraph(),
            new Promise<null>((resolve) =>
              setTimeout(() => resolve(null), graphTimeoutMs),
            ),
          ]).catch((e: unknown) => {
            if (import.meta.env.DEV) {
              console.error("[+page] Failed to load fresh graph:", e);
            }
            return null;
          })
        : Promise.race([
            getGraphWithPreload(),
            new Promise<null>((resolve) =>
              setTimeout(() => resolve(null), graphTimeoutMs),
            ),
          ]).catch((e: unknown) => {
            if (import.meta.env.DEV) {
              console.error("[+page] Failed to load public graph:", e);
            }
            return null;
          });

      if (isAuth) {
        // Authenticated: render notes and a note-based graph immediately so the
        // UI is responsive and tests can assert on filter counts right away.
        allNotes = await notesPromise;
        applyFiltersAndSort();
        graphData = {
          nodes: allNotes.map((n) => ({
            id: n.id,
            title: n.title,
            type: n.type || "unknown",
          })),
          links: [],
        };
        loading = false;
      }

      const freshGraphResult = await graphPromise;

      // For public view, derive the note list from the public graph.
      if (!isAuth && freshGraphResult && "nodes" in freshGraphResult) {
        allNotes = (freshGraphResult as GraphData).nodes.map((n) => {
          const raw = n as RawNode;
          return {
            id: raw.id ?? raw.Id ?? raw.ID ?? "",
            title: raw.title ?? raw.Title ?? "",
            content: "",
            metadata: {},
            type: raw.type ?? raw.Type ?? "unknown",
            created_at: "",
            updated_at: "",
          };
        });
        applyFiltersAndSort();
      }

      // Update graph data with fresh or preloaded graph when available.
      // We always keep all loaded notes represented in the graph so filter
      // counts and node visibility stay consistent with the note list.
      if (freshGraphResult) {
        let freshGraph: GraphData | null = null;
        if (isAuth && "fresh" in freshGraphResult) {
          freshGraph = freshGraphResult.fresh;
          graphDelta = freshGraphResult.delta ?? undefined;
        } else if (!isAuth) {
          freshGraph = freshGraphResult as GraphData;
          graphDelta = undefined;
        }

        if (freshGraph && freshGraph.nodes && freshGraph.nodes.length > 0) {
          // Merge fresh graph links with the note list so we never drop notes
          // that the graph service/backend cache hasn't picked up yet.
          const noteMap = new Map(allNotes.map((n) => [n.id, n]));
          const freshNodeMap = new Map(
            freshGraph.nodes.map((n) => [normalizeNode(n as RawNode).id, n as RawNode]),
          );

          // Use all current notes as nodes, preferring titles/types from the note list
          const mergedNodes = allNotes.map((n) => {
            const fresh = freshNodeMap.get(n.id) as RawNode | undefined;
            return normalizeNode(
              fresh
                ? { ...fresh, title: n.title, type: n.type || fresh.type || "unknown" }
                : ({
                    id: n.id,
                    title: n.title,
                    type: n.type || "unknown",
                  } as RawNode),
            );
          });

          // Add any extra graph nodes that are not in the note list (e.g. public graph)
          for (const node of freshGraph.nodes) {
            const normalized = normalizeNode(node as RawNode);
            if (!noteMap.has(normalized.id)) {
              mergedNodes.push(normalized);
            }
          }

          graphData = {
            nodes: mergedNodes,
            links: (freshGraph.links || []).map((l) => normalizeLink(l as RawLink)),
          };
        }
      }
    } catch (e: unknown) {
      apiError = toErrorResponse(e);
      if (import.meta.env.DEV) {
        console.error(e);
      }
    } finally {
      loading = false;
    }
  }

  async function loadNotes() {
    try {
      allNotes = await getNotes();
      applyFiltersAndSort();
      // Also load graph data when notes are loaded
      await loadGraphData();
    } catch (e: unknown) {
      apiError = toErrorResponse(e);
      if (import.meta.env.DEV) {
        console.error(e);
      }
    } finally {
      loading = false;
    }
  }

  async function loadGraphData() {
    if (allNotes.length === 0) {
      graphData = { nodes: [], links: [] };
      return;
    }

    graphLoading = true;
    try {
      // Always load full graph on main page
      const rawData = await getFullGraphData(apiConfig.default_limit, undefined, true);

      // Defensive check for API response structure (empty graphs also fall back)
      if (
        !rawData ||
        !rawData.nodes ||
        !Array.isArray(rawData.nodes) ||
        rawData.nodes.length === 0
      ) {
        if (import.meta.env.DEV) {
          console.warn(
            "[+page] Graph API returned empty or invalid data structure:",
            rawData,
          );
        }
        // Fallback: build simple graph from notes
        graphData = {
          nodes: allNotes.map((n) => ({
            id: n.id,
            title: n.title,
            type: n.type || "unknown",
          })),
          links: [],
        };
        return;
      }

      // Debug: check what types come from API
      const apiTypes = rawData.nodes.map(
        (n) => (n as RawNode).type ?? (n as RawNode).Type ?? "MISSING",
      );
      if (import.meta.env.DEV) {
        console.log(
          "[+page] API node types:",
          [...new Set(apiTypes)],
          "Total:",
          apiTypes.length,
        );
        console.log(
          "[+page] First 5 raw nodes:",
          rawData.nodes.slice(0, 5).map((n) => {
            const raw = n as RawNode;
            return { id: raw.id, type: raw.type, Type: raw.Type };
          }),
        );
      }

      // Transform nodes: backend might return Id/id/ID in different cases
      const transformedNodes = rawData.nodes.map((n) =>
        normalizeNode(n as RawNode),
      );

      // Transform links: backend returns source_note_id/target_note_id, frontend expects source/target
      const transformedLinks = (rawData.links || []).map((l) =>
        normalizeLink(l as RawLink),
      );

      graphData = {
        nodes: transformedNodes,
        links: transformedLinks,
      };

      // Defensive: ensure graph nodes correspond to loaded notes
      if (allNotes.length > 0 && transformedNodes.length > 0) {
        const noteIds = new Set(allNotes.map((n) => n.id));
        const hasIntersection = transformedNodes.some((n) =>
          noteIds.has(n.id),
        );
        if (!hasIntersection) {
          if (import.meta.env.DEV) {
            console.warn(
              "[+page] Loaded graph does not intersect with notes; rebuilding from notes",
            );
          }
          graphData = {
            nodes: allNotes.map((n) => ({
              id: n.id,
              title: n.title,
              type: n.type || "unknown",
            })),
            links: [],
          };
          return;
        }
      }

      if (import.meta.env.DEV) {
        console.log(
          "[+page] Transformed links:",
          transformedLinks.length,
          "links",
        );
        console.log(
          "[+page] First 3 transformed links:",
          transformedLinks.slice(0, 3),
        );

        // Check if links match node IDs
        const nodeIds = new Set(transformedNodes.map((n) => n.id));
        const validLinks = transformedLinks.filter(
          (l) => nodeIds.has(l.source) && nodeIds.has(l.target),
        );
        console.log("[+page] Node IDs:", Array.from(nodeIds).slice(0, 5));
        console.log(
          "[+page] Valid links (matching node IDs):",
          validLinks.length,
          "of",
          transformedLinks.length,
        );

        if (validLinks.length < transformedLinks.length) {
          console.warn(
            "[+page] Some links have invalid node IDs:",
            transformedLinks.filter(
              (l) => !nodeIds.has(l.source) || !nodeIds.has(l.target),
            ),
          );
        }
      }
    } catch (e) {
      if (import.meta.env.DEV) {
        console.error("[+page] Failed to load graph:", e);
      }
      // Fallback: build simple graph from notes
      graphData = {
        nodes: allNotes.map((n) => ({
          id: n.id,
          title: n.title,
          type: n.type || "unknown",
        })),
        links: [],
      };
    } finally {
      graphLoading = false;
    }
  }

  // Reload graph when allNotes changes (notes added/deleted)
  $effect(() => {
    if (browser && allNotes.length > 0) {
      loadGraphData();
    }
  });

  // Helper to get note type - unified with renderer.ts logic via CelestialBody
  function getNoteType(note: Note): string {
    return CelestialBody.fromString(note.type).type;
  }

  // Reactive filtered graph data based on filter state
  const filteredGraphData = $derived(
    filterState.filterGraphData(graphData, allNotes, getNoteType),
  );

  function applyFiltersAndSort() {
    filteredNotes = filterState.applyFiltersAndSort(allNotes, getNoteType);
  }

  async function handleSearch() {
    if (!filterState.isSearchActive) {
      // searchQuery cleared — filteredGraphData will reactively show all nodes
      applyFiltersAndSort();
      return;
    }

    try {
      // Use server search results only for the list view (filteredNotes)
      // The graph canvas uses local search via filteredGraphData derived
      const response = await searchNotes(filterState.searchQuery.value, 1, 20);
      filteredNotes = response.data;
    } catch (e) {
      // Fallback: local filter on allNotes
      applyFiltersAndSort();
      if (import.meta.env.DEV) {
        console.error("Search error:", e);
      }
    }
  }

  function handleDeleteRequest(id: string) {
    noteToDelete = id;
    showConfirmDelete = true;
  }

  async function handleDeleteConfirm() {
    if (!noteToDelete) return;

    try {
      await deleteNote(noteToDelete);
      selectedNodeId = null;
      // Remove deleted note from local arrays immediately
      allNotes = allNotes.filter((n) => n.id !== noteToDelete);
      filteredNotes = filteredNotes.filter((n) => n.id !== noteToDelete);
      // Then reload from server to ensure sync
      await loadNotes();
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
        `Delete ${selectedNoteIds.size} selected note${selectedNoteIds.size > 1 ? "s" : ""}? This action cannot be undone.`,
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
      await loadNotes();
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
    await loadNotes();
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
      await loadNotes();
    } catch {
      if (browser) {
        alert(t("page.restoreError"));
      }
    }
  }

  async function handleNoteCreate(data: {
    title: string;
    content: string;
    type: string;
  }) {
    try {
      await createNote({
        title: data.title,
        content: data.content,
        type: data.type,
      });
      await loadNotes();
    } catch {
      if (browser) {
        alert(t("note.createError"));
      }
    }
  }

  function handleNoteCreated(note: Note) {
    showCreateModal = false;
    selectedNodeId = note.id;
    loadNotes();
  }

  function handleToggleView(view: "graph" | "list") {
    currentView = view;
  }
</script>

<!-- Splash Screen on initial load -->
<SplashScreen />

<!-- Main page container - root element for the page layout -->
<!-- Functionality: Provides full viewport height/width container with hidden overflow -->
<div class="page-container">
  <!-- Floating Controls with Filters -->
  <FloatingControls
    onCreate={() => {
      showCreateModal = true;
    }}
    onSearch={(query: string) => {
      filterState = filterState.with({ searchQuery: query });
      searchQuery = query;
      handleSearch();
    }}
    onToggleView={handleToggleView}
    onFilter={(type: string) => {
      filterState = filterState.with({ selectedType: type });
      selectedType = type;
      applyFiltersAndSort();
    }}
    {typeFilters}
    {selectedType}
    {currentView}
    typeCounts={Object.fromEntries(
      typeFilters.map((f) => [
        f.id,
        f.id === "all"
          ? allNotes.length
          : allNotes.filter((n) => n.type === f.id).length,
      ]),
    )}
  />

  <!-- Fullscreen Graph Container -->
  <div class="fullscreen-graph" data-testid="graph-2d-container">
    {#if loading}
      <div class="loading-overlay">
        <div class="spinner"></div>
        <p>{t("page.loadingNotes")}</p>
      </div>
    {:else if apiError}
      <ApiErrorDisplay error={apiError} onClose={() => (apiError = null)} />
      <button
        onclick={() => {
          apiError = null;
          loadDataParallel();
        }}>{t("page.retry")}</button
      >
    {:else if currentView === "graph"}
      <!-- Debug info - remove in production -->
      {#if import.meta.env.DEV}
        <div
          style="position: fixed; top: 10px; left: 10px; background: rgba(0,0,0,0.8); color: #0f0; padding: 10px; font-family: monospace; font-size: 12px; z-index: 9999; max-width: 400px;"
        >
          <div>allNotes: {allNotes.length}</div>
          <div>graphData.nodes: {graphData.nodes.length}</div>
          <div>graphData.links: {graphData.links.length}</div>
          <div>filtered: {filteredGraphData.nodes.length}</div>
          <div>selectedType: {selectedType}</div>
          <div>loading: {loading}</div>
          <div>graphLoading: {graphLoading}</div>
          <div>
            apiError: {(apiError as ErrorResponse | null)?.message ?? "none"}
          </div>
        </div>
      {/if}
      <!-- Fullscreen 2D Graph View -->
      {#if graphLoading}
        <div class="center">
          <div class="spinner"></div>
          <p>{t("page.loadingGraph")}</p>
        </div>
      {:else}
        <GraphCanvas
          nodes={filteredGraphData.nodes}
          links={filteredGraphData.links}
          delta={graphDelta}
          onNodeClick={(node: { id: string }) => (selectedNodeId = node.id)}
          onNoteCreate={handleNoteCreate}
          onNoteDelete={handleDeleteRequest}
        />
      {/if}
    {:else if currentView === "list"}
      <!-- List View -->
      <div class="list-container" data-testid="list-container">
        <div class="list-header">
          <div class="list-controls">
            <button
              class="list-control-btn"
              data-testid="select-mode-toggle"
              onclick={toggleSelectionMode}
              aria-label={t("page.selectionToggle")}
            >
              {selectionMode ? t("page.cancelSelection") : t("page.select")}
            </button>
            {#if selectionMode}
              <button
                class="list-control-btn"
                onclick={toggleSelectAll}
                aria-label={t("page.selectAllAria")}
              >
                {selectedNoteIds.size === filteredNotes.length
                  ? t("page.clearSelection")
                  : t("page.selectAll")}
              </button>
            {/if}
          </div>
          <div class="list-sort">
            <label for="sort-select" class="sort-label"
              >{t("page.sortBy")}</label
            >
            <select
              id="sort-select"
              class="sort-select"
              value={sortBy}
              onchange={(e) => {
                sortBy = e.currentTarget.value as typeof sortBy;
                applyFiltersAndSort();
              }}
              aria-label={t("page.sortAriaLabel")}
            >
              {#each sortOptions as opt}
                <option value={opt.id}>{opt.label}</option>
              {/each}
            </select>
          </div>
        </div>

        {#if filteredNotes.length === 0}
          <div class="empty-state" data-testid="empty-state">
            <StateIllustration
              type={!filterState.isTypeActive && !filterState.isSearchActive
                ? "empty"
                : "no-results"}
            />
            <h2>
              {!filterState.isTypeActive && !filterState.isSearchActive
                ? t("page.emptyListNoNotes")
                : t("page.emptyListNoSearch")}
            </h2>
            <p>
              {!filterState.isTypeActive && !filterState.isSearchActive
                ? t("page.emptyListPrompt")
                : filterState.isSearchActive
                  ? t("page.noSearchResults", {
                      query: filterState.searchQuery.value,
                    })
                  : t("page.noTypeResults", {
                      type:
                        filterState
                          .getSelectedTypeLabel(typeFilters)
                          ?.toLowerCase() ?? "",
                    })}
            </p>
            <button
              class="new-note-button"
              onclick={() => (showCreateModal = true)}
            >
              {t("page.createFirstNote")}
            </button>
          </div>
        {:else}
          <div class="notes-grid" data-testid="notes-grid">
            {#each filteredNotes as note, index (note.id)}
              <NoteCard
                {note}
                animationIndex={index}
                selected={selectedNoteIds.has(note.id)}
                selectMode={selectionMode}
                onSelect={handleNoteSelect}
                onEdit={handleNoteEdit}
                onDelete={handleNoteDelete}
                onClick={() => (selectedNodeId = note.id)}
                highlightQuery={filterState.searchQuery.value}
              />
            {/each}
          </div>
        {/if}
      </div>

      <!-- Floating batch delete panel -->
      {#if selectionMode && selectedNoteIds.size > 0}
        <div class="batch-panel">
          <span class="batch-count"
            >{t("page.selectedCount", {
              count: selectedNoteIds.size.toString(),
            })}</span
          >
          <button
            class="batch-btn batch-btn--actions"
            onclick={() => (showBulkActionsMenu = !showBulkActionsMenu)}
            aria-label={t("page.bulkActionsToggle")}
          >
            {t("page.bulkActionsActions")}
          </button>
          <button
            class="batch-btn batch-btn--delete"
            onclick={handleBatchDelete}
            aria-label={t("page.bulkActionsDelete")}
          >
            {t("page.bulkActionsDeleteSelected")}
          </button>
          <button
            class="batch-btn batch-btn--cancel"
            onclick={() => {
              selectedNoteIds.clear();
              selectionMode = false;
            }}
            aria-label={t("page.cancelSelection")}
          >
            {t("modal.cancel")}
          </button>
        </div>

        <!-- Bulk actions menu -->
        {#if showBulkActionsMenu}
          <div class="bulk-actions-menu">
            <button
              class="bulk-action-item"
              onclick={() => {
                showBulkActionsMenu = false;
              }}
              aria-label={t("page.bulkActionsMoveType")}
            >
              <span class="bulk-action-icon">📂</span>
              {t("page.bulkActionsMoveType")}
            </button>
            <button
              class="bulk-action-item"
              onclick={() => {
                showBulkActionsMenu = false;
              }}
              aria-label={t("page.bulkActionsAddTags")}
            >
              <span class="bulk-action-icon">🏷️</span>
              {t("page.bulkActionsAddTags")}
            </button>
            <button
              class="bulk-action-item"
              onclick={() => {
                showBulkActionsMenu = false;
              }}
              aria-label={t("page.bulkActionsExport")}
            >
              <span class="bulk-action-icon">📤</span>
              {t("page.bulkActionsExport")}
            </button>
          </div>
        {/if}
      {/if}
    {/if}
  </div>
</div>

<!-- Side Panel for selected note -->
{#if selectedNodeId}
  <NoteSidePanel
    nodeId={selectedNodeId}
    onClose={() => (selectedNodeId = null)}
    onEdit={(id: string) => {
      noteToEdit = id;
      showEditModal = true;
    }}
    onDelete={handleDeleteRequest}
  />
{/if}

<!-- Create Note Modal -->
<CreateNoteModal bind:open={showCreateModal} onSuccess={handleNoteCreated} />

<!-- Edit Note Modal -->
{#if noteToEdit}
  <EditNoteModal
    bind:open={showEditModal}
    noteId={noteToEdit}
    onSuccess={() => {
      showEditModal = false;
      noteToEdit = null;
    }}
  />
{/if}

<!-- Confirm Modal for delete -->
<ConfirmModal
  bind:open={showConfirmDelete}
  title={t("modal.deleteTitle")}
  message={t("modal.deleteMessage")}
  confirmText={t("modal.delete")}
  cancelText={t("modal.cancel")}
  danger={true}
  onConfirm={handleDeleteConfirm}
  onCancel={() => {
    showConfirmDelete = false;
    noteToDelete = null;
  }}
/>

<!-- Undo toast -->
{#if showUndoToast}
  <div
    class="undo-toast"
    class:undo-toast--restore={undoToastStage === "restore"}
  >
    {#if undoToastStage === "done"}
      <span class="undo-toast-message">{t("toast.done")}</span>
    {:else}
      <span class="undo-toast-message">{t("toast.noteDeleted")}</span>
      <button
        class="undo-toast-btn"
        onclick={handleUndoRestore}
        aria-label={t("toast.restoreAriaLabel")}
      >
        {t("toast.restore")}
      </button>
    {/if}
  </div>
{/if}

<style>
  .page-container {
    height: 100vh;
    width: 100vw;
    overflow: hidden;
    position: relative;
    background: var(--gradient-cosmic-bg);
    color: var(--color-text-dark);
  }

  /* Fullscreen Graph Container */
  .fullscreen-graph {
    position: fixed;
    top: 60px; /* Space for floating controls (56px + margin) */
    left: 0;
    right: 0;
    bottom: 0;
    width: 100%;
    height: calc(100vh - 60px);
  }

  .fullscreen-graph :global(canvas) {
    width: 100% !important;
    height: 100% !important;
  }

  /* List Container */
  .list-container {
    max-width: 1400px;
    margin: 0 auto;
    padding: 24px;
    height: calc(100vh - 60px);
    overflow-y: auto;
  }

  .list-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 1.5rem;
    gap: 1rem;
  }

  .list-controls {
    display: flex;
    gap: 0.75rem;
  }

  .list-control-btn {
    padding: 0.5rem 1rem;
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: 6px;
    color: rgba(255, 255, 255, 0.85);
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition:
      background 0.2s ease,
      border-color 0.2s ease;
  }

  .list-control-btn:hover {
    background: rgba(255, 255, 255, 0.12);
    border-color: rgba(255, 255, 255, 0.2);
  }

  .list-sort {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .sort-label {
    font-size: 0.875rem;
    color: rgba(255, 255, 255, 0.6);
  }

  .sort-select {
    padding: 0.4rem 0.75rem;
    background: rgba(10, 14, 35, 0.85);
    border: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: 6px;
    color: rgba(255, 255, 255, 0.9);
    font-size: 0.875rem;
    cursor: pointer;
  }

  .sort-select option {
    background: rgba(10, 14, 35, 0.95);
    color: white;
  }

  .batch-panel {
    position: fixed;
    bottom: 24px;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 0.75rem 1.25rem;
    background: rgba(10, 14, 35, 0.95);
    border: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: 12px;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
    z-index: 100;
    animation: slide-up 0.3s ease;
  }

  @keyframes slide-up {
    from {
      opacity: 0;
      transform: translate(-50%, 20px);
    }
    to {
      opacity: 1;
      transform: translate(-50%, 0);
    }
  }

  .batch-count {
    font-size: 0.875rem;
    color: rgba(255, 255, 255, 0.7);
    font-weight: 500;
  }

  .batch-btn {
    padding: 0.5rem 1rem;
    border: none;
    border-radius: 6px;
    font-size: 0.875rem;
    font-weight: 600;
    cursor: pointer;
    transition: opacity 0.2s ease;
  }

  .batch-btn:hover {
    opacity: 0.85;
  }

  .batch-btn--delete {
    background: rgba(239, 68, 68, 0.85);
    color: white;
  }

  .batch-btn--cancel {
    background: rgba(255, 255, 255, 0.1);
    color: rgba(255, 255, 255, 0.85);
  }

  .batch-btn--actions {
    background: var(--color-primary, #8b5cf6);
    color: white;
  }

  .bulk-actions-menu {
    position: fixed;
    bottom: 100px;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 0.75rem;
    background: rgba(10, 14, 35, 0.95);
    border: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: 12px;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
    z-index: 101;
    animation: slide-up 0.3s ease;
    min-width: 200px;
  }

  .bulk-action-item {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.5rem 0.75rem;
    background: transparent;
    border: none;
    border-radius: 6px;
    color: rgba(255, 255, 255, 0.9);
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: background 0.2s ease;
    text-align: left;
  }

  .bulk-action-item:hover {
    background: rgba(255, 255, 255, 0.1);
  }

  .bulk-action-icon {
    font-size: 1.1rem;
  }

  .undo-toast {
    position: fixed;
    bottom: 24px;
    right: 24px;
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 0.75rem 1.25rem;
    background: rgba(10, 14, 35, 0.95);
    border: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: 12px;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
    z-index: 100;
    animation: slide-in 0.3s ease;
  }

  @keyframes slide-in {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .undo-toast-message {
    font-size: 0.875rem;
    color: rgba(255, 255, 255, 0.9);
  }

  .undo-toast-btn {
    padding: 0.4rem 0.8rem;
    background: var(--color-primary, #8b5cf6);
    border: none;
    border-radius: 4px;
    color: white;
    font-size: 0.8rem;
    font-weight: 600;
    cursor: pointer;
  }

  .undo-toast-btn:hover {
    opacity: 0.85;
  }

  .loading-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.8);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 16px;
    z-index: 1000;
    color: white;
  }

  .center {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 60px 20px;
    color: #64748b;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 3px solid #e2e8f0;
    border-top-color: #3b82f6;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-bottom: 16px;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  /* Empty State */
  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 80px 20px;
    text-align: center;
    background: linear-gradient(135deg, #0a1a3a 0%, #020617 100%);
    height: 100%;
  }

  .empty-state h2 {
    font-size: 24px;
    font-weight: 600;
    color: #f8fafc;
    margin: 0 0 8px;
  }

  .empty-state p {
    color: #94a3b8;
    margin: 0 0 24px;
  }

  .new-note-button {
    padding: 12px 24px;
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    color: white;
    border: none;
    border-radius: 8px;
    font-size: 16px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .new-note-button:hover {
    transform: translateY(-2px);
    box-shadow: 0 8px 20px rgba(59, 130, 246, 0.4);
  }

  .notes-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
    gap: 1.5rem;
  }

  @media (max-width: 768px) {
    .notes-grid {
      grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
      gap: 16px;
      padding: 16px 0;
    }

    .fullscreen-graph {
      top: 50px;
      height: calc(100vh - 50px);
    }

    .list-container {
      padding: 16px;
      height: calc(100vh - 50px);
    }
  }
</style>
