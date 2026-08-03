<script lang="ts">
  import { onMount } from "svelte";
  import { browser } from "$app/environment";
  import { goto } from "$app/navigation";
  import { initAuth, isAuthenticated } from "$shared/stores/auth.svelte";
  import {
    getNotes,
    getNote,
    createNote,
    deleteNote,
    restoreNote,
    type Note,
  } from "$shared/api/notes";
  import {
    getGraphData,
    getFullGraphData,
    type GraphData,
    type GraphNode,
    type GraphLink,
  } from "$shared/api/graph";

  const KNOWLEDGE_CORE_ID = "00000000-0000-0000-0000-000000000001";
  import { createLink, updateLink, deleteLink } from "$shared/api/links";
  import GraphCanvas from "$widgets/graph-canvas/GraphCanvas.svelte";
  import { PublicGraphTopBar } from "$features/graph-ui";
  import CosmicCockpitLayout from "$widgets/cosmic-cockpit/CosmicCockpitLayout.svelte";
  import EditNoteModal from "$widgets/notes/EditNoteModal.svelte";
  import CreateNoteModal from "$widgets/notes/CreateNoteModal.svelte";
  import ConfirmModal from "$widgets/confirm/ConfirmModal.svelte";
  import StateIllustration from "$components/atoms/StateIllustration.svelte";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
  import { CelestialBody } from "$entities";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  interface RawNode extends GraphNode {
    Id?: string;
    ID?: string;
    Title?: string;
    Type?: string;
    created_at?: string;
    createdAt?: string;
    CreatedAt?: string;
  }

  interface RawLink extends GraphLink {
    source_note_id?: string;
    target_note_id?: string;
    source_type?: string;
  }

  let notes: Note[] = $state([]);
  let graphData: GraphData = $state({ nodes: [], links: [] });
  let knowledgeCore: Note | null = $state(null);
  let loading = $state(true);
  let error = $state("");
  let selectedNodeId: string | null = $state(null);
  // Default to the full graph (same behavior as the main "/" page) unless the
  // caller explicitly asks for the local/centered view via ?full=false.
  let showFullGraph = $state(
    !browser || new URL(window.location.href).searchParams.get("full") !== "false"
  );
  let showEditModal = $state(false);
  let noteToEdit: string | null = $state(null);
  let showCreateModal = $state(false);
  let createChildParent: { id: string; title: string; type?: string } | null = $state(null);
  let showConfirmDelete = $state(false);
  let noteToDelete: string | null = $state(null);
  let canvasController: {
    focusMode: boolean;
    resetView: () => void;
    openSearch: () => void;
    toggleFocus: () => void;
  } | undefined = $state(undefined);

  const graphTypeFilters = [
    { id: "all", label: t("filter.all"), emoji: "🌌", description: t("filter.all.description") },
    ...["star", "planet", "moon", "comet", "galaxy", "nebula", "asteroid", "satellite", "blackhole", "dust", "unknown"].map(
      (id) => {
        const body = CelestialBody.fromString(id);
        return { id, label: body.label, emoji: body.emoji, description: body.description };
      }
    ),
  ];

  async function loadGraphData({ nocache = false }: { nocache?: boolean } = {}) {
    loading = true;
    error = "";
    try {
      let rawData: GraphData;
      if (showFullGraph) {
        // Load the full graph of all notes; bypass cache after mutations
        // limit=0 means no cap — a positive default_limit would hide new nodes
        // when the total exceeds the configured page size.
        rawData = await getFullGraphData(0, undefined, nocache);
      } else {
        // Load the local graph
        if (isAuthenticated()) {
          try {
            notes = await getNotes();
          } catch {
            notes = [];
          }
        } else {
          notes = [];
        }

        if (notes.length > 0) {
          const centerNote = notes[0];
          rawData = await getGraphData(centerNote.id, 3);
        } else {
          // If there are no notes, load the full graph (public notes)
          rawData = await getFullGraphData(0, undefined, nocache);
        }
      }

      // Fetch the Knowledge Core system note for in-app help.
      // Anonymous users cannot access individual notes, so skip this call
      // to avoid a 401 -> auth/refresh -> redirect-to-login cascade on the
      // public graph page.
      if (isAuthenticated()) {
        try {
          knowledgeCore = await getNote(KNOWLEDGE_CORE_ID);
        } catch {
          knowledgeCore = null;
        }
      } else {
        knowledgeCore = null;
      }

      // Transform nodes: backend might return Id/id/ID in different cases
      const transformedNodes = (rawData.nodes as RawNode[]).map((n) => ({
        id: n.id || n.Id || n.ID || "",
        title: n.title || n.Title || "",
        type: n.type || n.Type || "star",
        createdAt: n.created_at || n.createdAt || n.CreatedAt,
      }));

      // Ensure the Knowledge Core is always present on the canvas
      if (knowledgeCore && !transformedNodes.some((n) => n.id === knowledgeCore?.id)) {
        transformedNodes.push({
          id: knowledgeCore.id,
          title: knowledgeCore.title,
          type: "technical",
          createdAt: knowledgeCore.created_at,
        });
      }

      // Transform links: backend returns source_note_id/target_note_id, frontend expects source/target
      const transformedLinks = (rawData.links as RawLink[]).map((l) => ({
        source: l.source_note_id || l.source,
        target: l.target_note_id || l.target,
        weight: l.weight,
        link_type: l.link_type,
        source_type: l.source_type || "user",
      }));

      graphData = {
        nodes: transformedNodes,
        links: transformedLinks,
      };

      if (import.meta.env.DEV) {
        console.log(
          "[graph/+page] Graph loaded:",
          graphData.nodes.length,
          "nodes,",
          graphData.links.length,
          "links"
        );
        console.log("[graph/+page] Sample node:", transformedNodes[0]);
        console.log("[graph/+page] Sample link:", transformedLinks[0]);
      }
    } catch (e) {
      if (import.meta.env.DEV) {
        console.error("Failed to load graph:", e);
      }
      error = t("graph.loadDataError");
    } finally {
      loading = false;
    }
  }

  onMount(async () => {
    if (!browser) return;
    // Make sure auth is initialized before loading graph data so the token is
    // available for graph-service requests.
    await initAuth();
    // Allow tests/URLs to force a fresh graph load (bypass graph-service cache)
    const url = new URL(window.location.href);
    await loadGraphData({ nocache: url.searchParams.has("nocache") });

    // Expose flag for E2E tests to assert background sync is gated by auth.
    // This page never polls for delta updates, so the flag is always false.
    (window as unknown as Record<string, unknown>).__kgGraphPollingActive = false;
  });

  function handleNodeSelect(nodeId: string | null) {
    selectedNodeId = nodeId;
  }

  function handleDeleteRequest(nodeId: string) {
    noteToDelete = nodeId;
    showConfirmDelete = true;
  }

  async function handleConfirmDelete() {
    if (!noteToDelete) return;
    try {
      await deleteNote(noteToDelete);
      selectedNodeId = null;
      await loadGraphData({ nocache: true });
    } catch (e) {
      if (import.meta.env.DEV) {
        console.error("Failed to delete note:", e);
      }
    } finally {
      noteToDelete = null;
      showConfirmDelete = false;
    }
  }

  async function handleNoteRestore(nodeId: string) {
    try {
      await restoreNote(nodeId);
      await loadGraphData();
    } catch (e) {
      if (import.meta.env.DEV) {
        console.error("Failed to restore note:", e);
      }
    }
  }

  async function handleNoteCreate(data: { title: string; content: string; type: string }) {
    try {
      await createNote(data);
      await loadGraphData({ nocache: true });
    } catch (e) {
      if (import.meta.env.DEV) {
        console.error("Failed to create note:", e);
      }
    }
  }

  async function handleNoteCreatedSuccess(note: Note) {
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
    selectedNodeId = note.id;
    await loadGraphData({ nocache: true });
  }

  function handleCreateChildNote(parent: { id: string; title: string; type?: string }) {
    createChildParent = parent;
    showCreateModal = true;
  }

  function handleNoteEdit(id: string) {
    noteToEdit = id;
    showEditModal = true;
  }

  async function handleLinkCreate(link: {
    source: string;
    target: string;
    link_type: string;
    weight: number;
  }) {
    try {
      await createLink({
        source_note_id: link.source,
        target_note_id: link.target,
        link_type: link.link_type,
        weight: link.weight,
      });
      await loadGraphData({ nocache: true });
    } catch (e) {
      if (import.meta.env.DEV) {
        console.error("Failed to create link:", e);
      }
    }
  }

  async function handleLinkEdit(link: {
    id?: string;
    source: string;
    target: string;
    link_type: string;
    weight: number;
  }) {
    if (!link.id) {
      if (import.meta.env.DEV) {
        console.error("Cannot edit link without id");
      }
      return;
    }
    try {
      await updateLink(link.id, {
        link_type: link.link_type,
        weight: link.weight,
      });
      await loadGraphData({ nocache: true });
    } catch (e) {
      if (import.meta.env.DEV) {
        console.error("Failed to update link:", e);
      }
    }
  }

  async function handleLinkDelete(link: {
    id?: string;
    source: string;
    target: string;
    link_type: string;
  }) {
    if (!link.id) {
      if (import.meta.env.DEV) {
        console.error("Cannot delete link without id");
      }
      return;
    }
    try {
      await deleteLink(link.id);
      await loadGraphData({ nocache: true });
    } catch (e) {
      if (import.meta.env.DEV) {
        console.error("Failed to delete link:", e);
      }
    }
  }

  // Отслеживаем изменение showFullGraph и загружаем данные
  // (skip the initial run, onMount already loads once).
  let showFullGraphInitialized = false;
  $effect(() => {
    if (browser) {
      // eslint-disable-next-line @typescript-eslint/no-unused-expressions
      showFullGraph;
      if (!showFullGraphInitialized) {
        showFullGraphInitialized = true;
        return;
      }
      loadGraphData();
    }
  });
</script>

<CosmicCockpitLayout
  isAuthenticated={isAuthenticated()}
  currentView="graph"
  layoutProvider="d3"
  onToggleView={(view) => {
    if (view === "list") goto("/");
    else if (view === "3d") goto("/graph/3d");
  }}
  onNoteCreate={() => (showCreateModal = true)}
  onNoteDelete={handleDeleteRequest}
  onNoteEdit={handleNoteEdit}
  onNodeSelect={handleNodeSelect}
  onCreateChildNote={handleCreateChildNote}
  selectedNodeId={selectedNodeId ?? null}
  nodeCount={graphData.nodes.length}
  linkCount={graphData.links.length}
  notes={graphData.nodes.map((n) => ({ id: n.id, title: n.title, type: n.type }))}
  showFullGraph={showFullGraph}
  onToggleFullGraph={(value) => (showFullGraph = value)}
>
  <div class="graph-cockpit-content">
    {#if !isAuthenticated() && !loading && !error}
      <PublicGraphTopBar
        currentView="graph"
        layoutProvider="d3"
        searchQuery=""
        selectedType="all"
        typeFilters={graphTypeFilters}
        nodeCount={graphData.nodes.length}
        linkCount={graphData.links.length}
        onSearch={() => {}}
        onToggleView={(view) => {
          if (view === "list") goto("/");
          else if (view === "3d") goto("/graph/3d");
        }}
        onFilter={() => {}}
        onSignIn={() => goto("/auth/login")}
        onRegister={() => goto("/auth/register")}
        {canvasController}
      />
    {/if}

    {#if loading}
      <div class="loading-overlay" data-testid="loading-overlay">
        <div class="spinner"></div>
        <p>{t("graph.loading")}</p>
      </div>
    {:else if error}
      <div class="error">
        <p>{error}</p>
        <button onclick={() => goto("/")}>{t("graph.goHome")}</button>
      </div>
    {:else if graphData.nodes.length > 0}
      {#key graphData.nodes.length + "-" + graphData.links.length}
        <div class="graph-view-wrapper">
          <GraphCanvas
            nodes={graphData.nodes}
            links={graphData.links}
            onNodeClick={(node: { id: string }) => handleNodeSelect(node.id)}
            onNoteDelete={handleDeleteRequest}
            onNoteRestore={handleNoteRestore}
            onNoteCreate={handleNoteCreate}
            onCreateChildNote={handleCreateChildNote}
            onLinkCreate={handleLinkCreate}
            onLinkEdit={handleLinkEdit}
            onLinkDelete={handleLinkDelete}
            helpContent={knowledgeCore?.content}
            showLinkTypeLegend={isAuthenticated()}
            showTopBar={isAuthenticated()}
            bind:controller={canvasController}
          />
        </div>
      {/key}
    {:else}
      <div class="empty">
        <StateIllustration type="no-links" />
        <p>{t("graph.noData")}</p>
      </div>
    {/if}
  </div>
</CosmicCockpitLayout>

{#if noteToEdit}
  <EditNoteModal
    bind:open={showEditModal}
    noteId={noteToEdit}
    onSuccess={() => {
      showEditModal = false;
      noteToEdit = null;
      loadGraphData({ nocache: true });
    }}
  />
{/if}

{#if showCreateModal}
  <CreateNoteModal
    bind:open={showCreateModal}
    onSuccess={handleNoteCreatedSuccess}
    onClose={() => (createChildParent = null)}
    parentNote={createChildParent ?? undefined}
    defaultType={createChildParent
      ? CelestialBody.getChildSuggestion(createChildParent.type)
      : undefined}
  />
{/if}

{#if showConfirmDelete}
  <ConfirmModal
    bind:open={showConfirmDelete}
    title={t("confirmModal.title")}
    message={t("note.deleteConfirm")}
    danger={true}
    onConfirm={handleConfirmDelete}
    onCancel={() => {
      noteToDelete = null;
      showConfirmDelete = false;
    }}
  />
{/if}

<style>
  .graph-cockpit-content {
    display: flex;
    flex-direction: column;
    width: 100%;
    height: 100%;
    background: transparent;
    color: var(--color-text-dark);
    overflow: hidden;
  }

  .graph-view-wrapper {
    position: relative;
    flex: 1 1 auto;
    min-height: 0;
    width: 100%;
  }

  .loading-overlay {
    position: absolute;
    inset: 0;
    background: rgba(0, 0, 0, 0.8);
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 16px;
    z-index: 1000;
    color: white;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 3px solid rgba(255, 255, 255, 0.2);
    border-top-color: #3b82f6;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }

  .error {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 16px;
    color: #ef4444;
  }

  .error button {
    padding: 8px 16px;
    background: #3b82f6;
    color: white;
    border: none;
    border-radius: 6px;
    cursor: pointer;
  }

  .empty {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    color: #94a3b8;
  }
</style>
