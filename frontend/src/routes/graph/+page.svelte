<script lang="ts">
  import { onMount } from "svelte";
  import { browser } from "$app/environment";
  import { goto } from "$app/navigation";
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
  } from "$shared/api/graph";

  const KNOWLEDGE_CORE_ID = "00000000-0000-0000-0000-000000000001";
  import { createLink } from "$shared/api/links";
  import GraphCanvas from "$components/organisms/GraphCanvas.svelte";
  import NoteSidePanel from "$components/organisms/NoteSidePanel.svelte";
  import EditNoteModal from "$components/organisms/EditNoteModal.svelte";
  import BackButton from "$components/atoms/BackButton.svelte";
  import WeltallBackground from "$components/atoms/WeltallBackground.svelte";
  import StateIllustration from "$components/atoms/StateIllustration.svelte";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  let notes: Note[] = $state([]);
  let graphData: GraphData = $state({ nodes: [], links: [] });
  let knowledgeCore: Note | null = $state(null);
  let loading = $state(true);
  let error = $state("");
  let selectedNodeId: string | null = $state(null);
  let showFullGraph = $state(
    browser && new URL(window.location.href).searchParams.has("full"),
  );
  let showEditModal = $state(false);
  let noteToEdit: string | null = $state(null);

  async function loadGraphData({
    nocache = false,
  }: { nocache?: boolean } = {}) {
    loading = true;
    error = "";
    try {
      let rawData: GraphData;
      if (showFullGraph) {
        // Load the full graph of all notes; bypass cache after mutations
        rawData = await getFullGraphData(0, undefined, nocache);
      } else {
        // Load the local graph
        try {
          notes = await getNotes();
        } catch {
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

      // Fetch the Knowledge Core system note for in-app help
      try {
        knowledgeCore = await getNote(KNOWLEDGE_CORE_ID);
      } catch {
        knowledgeCore = null;
      }

      // Transform nodes: backend might return Id/id/ID in different cases
      const transformedNodes = rawData.nodes.map((n: any) => ({
        id: n.id || n.Id || n.ID,
        title: n.title || n.Title,
        type: n.type || n.Type || "star",
        createdAt: n.created_at || n.createdAt || n.CreatedAt,
      }));

      // Ensure the Knowledge Core is always present on the canvas
      if (
        knowledgeCore &&
        !transformedNodes.some((n) => n.id === knowledgeCore?.id)
      ) {
        transformedNodes.push({
          id: knowledgeCore.id,
          title: knowledgeCore.title,
          type: "technical",
          createdAt: knowledgeCore.created_at,
        });
      }

      // Transform links: backend returns source_note_id/target_note_id, frontend expects source/target
      const transformedLinks = rawData.links.map((l: any) => ({
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
          "links",
        );
        console.log("[graph/+page] Sample node:", transformedNodes[0]);
        console.log("[graph/+page] Sample link:", transformedLinks[0]);
      }
    } catch (e) {
      console.error("Failed to load graph:", e);
      error = t("graph.loadDataError");
    } finally {
      loading = false;
    }
  }

  onMount(async () => {
    if (!browser) return;
    // Allow tests/URLs to force a fresh graph load (bypass graph-service cache)
    const url = new URL(window.location.href);
    await loadGraphData({ nocache: url.searchParams.has("nocache") });
  });

  function handleNodeSelect(nodeId: string) {
    selectedNodeId = nodeId;
  }

  async function handleNoteDelete(nodeId: string) {
    try {
      await deleteNote(nodeId);
      await loadGraphData({ nocache: true });
    } catch (e) {
      console.error("Failed to delete note:", e);
    }
  }

  async function handleNoteRestore(nodeId: string) {
    try {
      await restoreNote(nodeId);
      await loadGraphData();
    } catch (e) {
      console.error("Failed to restore note:", e);
    }
  }

  async function handleNoteCreate(data: {
    title: string;
    content: string;
    type: string;
  }) {
    try {
      await createNote(data);
      await loadGraphData({ nocache: true });
    } catch (e) {
      console.error("Failed to create note:", e);
    }
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
      console.error("Failed to create link:", e);
    }
  }

  // Отслеживаем изменение showFullGraph и загружаем данные
  $effect(() => {
    if (browser) {
      // eslint-disable-next-line @typescript-eslint/no-unused-expressions
      showFullGraph;
      loadGraphData();
    }
  });
</script>

<!-- Cosmic Background -->
<WeltallBackground />

<div class="graph-page">
  <BackButton href="/" />

  <div class="top-right-controls">
    <button
      class="login-btn"
      onclick={() => goto("/auth/login")}
      title={t("graph.loginTitle")}
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="20"
        height="20"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
      >
        <path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4" />
        <polyline points="10 17 15 12 10 7" />
        <line x1="15" y1="12" x2="3" y2="12" />
      </svg>
    </button>
  </div>

  <h1>{t("graph.title")}</h1>

  <div class="controls">
    <label class="toggle">
      <input
        type="checkbox"
        bind:checked={showFullGraph}
        data-testid="full-graph-toggle"
      />
      <span
        >{t("graph.showAllNotes")} ({showFullGraph
          ? t("graph.enabled")
          : t("graph.disabled")})</span
      >
    </label>
  </div>

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
  {:else}
    <div
      class="graph-container graph-3d-container"
      data-testid="graph-container"
    >
      {#if graphData.nodes.length > 0}
        {#key graphData.nodes.length + "-" + graphData.links.length}
          <GraphCanvas
            nodes={graphData.nodes}
            links={graphData.links}
            onNodeClick={(node: { id: string }) => handleNodeSelect(node.id)}
            onNoteDelete={handleNoteDelete}
            onNoteRestore={handleNoteRestore}
            onNoteCreate={handleNoteCreate}
            onLinkCreate={handleLinkCreate}
            helpContent={knowledgeCore?.content}
          />
        {/key}
      {:else}
        <div class="empty">
          <StateIllustration type="no-links" />
          <p>{t("graph.noData")}</p>
        </div>
      {/if}
    </div>
  {/if}
</div>

{#if selectedNodeId}
  <NoteSidePanel
    nodeId={selectedNodeId}
    onClose={() => (selectedNodeId = null)}
    onEdit={(id: string) => {
      noteToEdit = id;
      showEditModal = true;
    }}
    onDelete={() => {
      selectedNodeId = null;
      // Reload graph
      window.location.reload();
    }}
  />
{/if}

{#if noteToEdit}
  <EditNoteModal
    bind:open={showEditModal}
    noteId={noteToEdit}
    onSuccess={() => {
      showEditModal = false;
      noteToEdit = null;
      window.location.reload();
    }}
  />
{/if}

<style>
  .graph-page {
    padding: 20px;
    width: 100%;
    max-width: 100%;
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    flex: 1;
    min-height: 100vh;
    background: var(--gradient-cosmic-bg);
    color: var(--color-text-dark);
  }

  .top-right-controls {
    position: absolute;
    top: 20px;
    right: 20px;
    z-index: 1000;
  }

  .login-btn {
    padding: 10px;
    border: none;
    background: linear-gradient(135deg, #10b981 0%, #059669 100%);
    color: white;
    border-radius: 50%;
    cursor: pointer;
    transition: all 0.2s;
    box-shadow: 0 4px 12px rgba(16, 185, 129, 0.4);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .login-btn:hover {
    transform: scale(1.05);
    box-shadow: 0 6px 16px rgba(16, 185, 129, 0.5);
  }

  h1 {
    margin: 0 0 20px 0;
    font-size: 1.5rem;
  }

  .controls {
    position: absolute;
    top: 80px;
    right: 20px;
    z-index: 1000;
    background: rgba(0, 0, 0, 0.8);
    padding: 12px 16px;
    border-radius: 8px;
    color: white;
    border: 1px solid rgba(255, 255, 255, 0.2);
    backdrop-filter: blur(10px);
  }

  .toggle {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
    font-size: 14px;
  }

  .toggle input {
    cursor: pointer;
    width: 18px;
    height: 18px;
  }

  .graph-container {
    flex: 1;
    position: relative;
    min-height: 0;
    height: 100%;
    border-radius: 12px;
    overflow: hidden;
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
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
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
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: #94a3b8;
  }
</style>
