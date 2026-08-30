<script lang="ts">
  import { isAuthenticated } from "$shared/stores/auth.svelte";
  import { graphStore } from "$shared/stores/graph.svelte";
  import { createHomePageState } from "$features/home-page";
  import { GraphPageShell } from "$widgets/graph-page";
  import CreateNoteModal from "$widgets/notes/CreateNoteModal.svelte";
  import EditNoteModal from "$widgets/notes/EditNoteModal.svelte";
  import ConfirmModal from "$widgets/confirm/ConfirmModal.svelte";
  import NoteCard from "$widgets/notes/NoteCard.svelte";
  import ApiErrorDisplay from "$components/atoms/ApiErrorDisplay.svelte";
  import StateIllustration from "$components/atoms/StateIllustration.svelte";
  import GraphCanvas from "$widgets/graph-canvas/GraphCanvas.svelte";
  import FloatingAuthPanel from "$widgets/floating-auth-panel/FloatingAuthPanel.svelte";

  import SplashScreen from "$components/atoms/SplashScreen.svelte";

  const homePage = createHomePageState();
  const {
    t,
    openAuthPanel,
    closeAuthPanel,
    handleAuthSuccess,
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
    loadData,
  } = homePage;

  const Graph3DViewer = $derived(homePage.Graph3DViewer);

  let canvasController:
    | {
        focusMode: boolean;
        fogEnabled: boolean;
        resetView: () => void;
        openSearch: () => void;
        toggleFocus: () => void;
        toggleFog: () => void;
      }
    | undefined = $state(undefined);
</script>

<!-- Splash Screen on initial load -->
<SplashScreen />

<!-- Main page container - root element for the page layout -->
<!-- Functionality: Provides full viewport height/width container with hidden overflow -->
<GraphPageShell
  view={graphStore.currentView}
  layoutProvider={homePage.layoutProvider}
  searchQuery={homePage.searchQuery}
  selectedType={homePage.selectedType}
  typeFilters={homePage.typeFilters}
  notes={homePage.allNotes}
  nodeCount={homePage.filteredGraphData.nodes.length}
  linkCount={homePage.filteredGraphData.links.length}
  links={homePage.filteredGraphData.links}
  selectedNodeId={graphStore.selectedNodeId}
  {canvasController}
  onSearch={handleSearchQuery}
  onFilter={handleFilter}
  onToggleView={handleToggleView}
  onToggleLayoutProvider={handleToggleLayoutProvider}
  onNodeSelect={(id) => (graphStore.selectedNodeId = id)}
  onNoteCreate={() => (homePage.showCreateModal = true)}
  onNoteDelete={handleDeleteRequest}
  onNoteEdit={(id: string) => {
    homePage.noteToEdit = id;
    homePage.showEditModal = true;
  }}
  onCreateChildNote={handleCreateChildNote}
  onImport={handleImport}
  onSignIn={() => openAuthPanel("login")}
  onRegister={() => openAuthPanel("register")}
>
  <!-- Graph/List Container -->
  <div class="graph-content" data-testid="graph-2d-container">
    {#if homePage.loading}
      <div class="loading-overlay">
        <div class="spinner"></div>
        <p>{t("page.loadingNotes")}</p>
      </div>
    {:else if homePage.apiError}
      <ApiErrorDisplay error={homePage.apiError} onClose={clearApiError} />
      <button
        onclick={() => {
          clearApiError();
          loadData();
        }}>{t("page.retry")}</button
      >
    {:else if graphStore.currentView === "graph"}
      <!-- Debug info - remove in production -->
      {#if import.meta.env.DEV}
        <div
          style="position: fixed; top: 10px; left: 10px; background: rgba(0,0,0,0.8); color: #0f0; padding: 10px; font-family: monospace; font-size: 12px; z-index: 9999; max-width: 400px;"
        >
          <div>allNotes: {homePage.allNotes.length}</div>
          <div>graphData.nodes: {homePage.graphData.nodes.length}</div>
          <div>graphData.links: {homePage.graphData.links.length}</div>
          <div>filtered: {homePage.filteredGraphData.nodes.length}</div>
          <div>selectedType: {homePage.selectedType}</div>
          <div>loading: {homePage.loading}</div>
        </div>
      {/if}
      <!-- Fullscreen 2D Graph View -->
      <div class="graph-view-wrapper">
        <GraphCanvas
          nodes={homePage.filteredGraphData.nodes}
          links={homePage.filteredGraphData.links}
          onNodeClick={(node: { id: string }) => (graphStore.selectedNodeId = node.id)}
          onNoteCreate={handleNoteCreate}
          onNoteDelete={handleDeleteRequest}
          onCreateChildNote={handleCreateChildNote}
          showLinkTypeLegend={false}
          bind:controller={canvasController}
        />
      </div>
    {:else if graphStore.currentView === "3d" && Graph3DViewer}
      <!-- Fullscreen 3D Graph View -->
      <div class="graph-view-wrapper">
        <Graph3DViewer
          nodes={homePage.filteredGraphData.nodes}
          links={homePage.filteredGraphData.links}
          centerNodeId={graphStore.selectedNodeId}
          selectedNodeId={graphStore.selectedNodeId}
          onNodeClick={(node: { id: string }) => (graphStore.selectedNodeId = node.id)}
        />
      </div>
    {:else if graphStore.currentView === "list"}
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
              {homePage.selectionMode ? t("page.cancelSelection") : t("page.select")}
            </button>
            {#if homePage.selectionMode}
              <button
                class="list-control-btn"
                onclick={toggleSelectAll}
                aria-label={t("page.selectAllAria")}
              >
                {homePage.selectedNoteIds.size === homePage.filteredNotes.length
                  ? t("page.clearSelection")
                  : t("page.selectAll")}
              </button>
            {/if}
          </div>
          <div class="list-sort">
            <label for="sort-select" class="sort-label">{t("page.sortBy")}</label>
            <select
              id="sort-select"
              class="sort-select"
              value={homePage.sortBy}
              onchange={(e) => {
                handleSortChange(e.currentTarget.value as typeof homePage.sortBy);
              }}
              aria-label={t("page.sortAriaLabel")}
            >
              {#each homePage.sortOptions as opt}
                <option value={opt.id}>{opt.label}</option>
              {/each}
            </select>
          </div>
        </div>

        {#if homePage.filteredNotes.length === 0}
          <div class="empty-state" data-testid="empty-state">
            <StateIllustration
              type={!homePage.filterState.isTypeActive && !homePage.filterState.isSearchActive
                ? "empty"
                : "no-results"}
            />
            <h2>
              {!homePage.filterState.isTypeActive && !homePage.filterState.isSearchActive
                ? t("page.emptyListNoNotes")
                : t("page.emptyListNoSearch")}
            </h2>
            <p>
              {!homePage.filterState.isTypeActive && !homePage.filterState.isSearchActive
                ? t("page.emptyListPrompt")
                : homePage.filterState.isSearchActive
                  ? t("page.noSearchResults", {
                      query: homePage.filterState.searchQuery.value,
                    })
                  : t("page.noTypeResults", {
                      type:
                        homePage.filterState
                          .getSelectedTypeLabel(homePage.typeFilters)
                          ?.toLowerCase() ?? "",
                    })}
            </p>
            <button class="new-note-button" onclick={() => (homePage.showCreateModal = true)}>
              {t("page.createFirstNote")}
            </button>
          </div>
        {:else}
          <div class="notes-grid" data-testid="notes-grid">
            {#each homePage.filteredNotes as note, index (note.id)}
              <NoteCard
                {note}
                animationIndex={index}
                selected={homePage.selectedNoteIds.has(note.id)}
                selectMode={homePage.selectionMode}
                onSelect={handleNoteSelect}
                onEdit={handleNoteEdit}
                onDelete={handleNoteDelete}
                onClick={() => (graphStore.selectedNodeId = note.id)}
                highlightQuery={homePage.filterState.searchQuery.value}
                readonly={!isAuthenticated()}
              />
            {/each}
          </div>
        {/if}
      </div>

      <!-- Floating batch delete panel -->
      {#if homePage.selectionMode && homePage.selectedNoteIds.size > 0}
        <div class="batch-panel">
          <span class="batch-count"
            >{t("page.selectedCount", {
              count: homePage.selectedNoteIds.size.toString(),
            })}</span
          >
          <button
            class="batch-btn batch-btn--actions"
            onclick={() => (homePage.showBulkActionsMenu = !homePage.showBulkActionsMenu)}
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
              homePage.selectedNoteIds.clear();
              homePage.selectionMode = false;
            }}
            aria-label={t("page.cancelSelection")}
          >
            {t("modal.cancel")}
          </button>
        </div>

        <!-- Bulk actions menu -->
        {#if homePage.showBulkActionsMenu}
          <div class="bulk-actions-menu">
            <button
              class="bulk-action-item"
              onclick={() => {
                homePage.showBulkActionsMenu = false;
              }}
              aria-label={t("page.bulkActionsMoveType")}
            >
              <span class="bulk-action-icon">📂</span>
              {t("page.bulkActionsMoveType")}
            </button>
            <button
              class="bulk-action-item"
              onclick={() => {
                homePage.showBulkActionsMenu = false;
              }}
              aria-label={t("page.bulkActionsAddTags")}
            >
              <span class="bulk-action-icon">🏷️</span>
              {t("page.bulkActionsAddTags")}
            </button>
            <button
              class="bulk-action-item"
              onclick={() => {
                homePage.showBulkActionsMenu = false;
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
</GraphPageShell>

<!-- Public auth entry point -->
{#if !isAuthenticated()}
  <FloatingAuthPanel
    open={homePage.showAuthPanel}
    initialTab={homePage.authPanelTab}
    onClose={closeAuthPanel}
    onSuccess={handleAuthSuccess}
  />
{/if}

<!-- Create Note Modal -->
<CreateNoteModal
  bind:open={homePage.showCreateModal}
  onSuccess={handleNoteCreated}
  onClose={resetCreateChildParent}
  parentNote={homePage.createChildParent ?? undefined}
  defaultType={homePage.createChildDefaultType}
/>

<!-- Edit Note Modal -->
{#if homePage.noteToEdit}
  <EditNoteModal
    bind:open={homePage.showEditModal}
    noteId={homePage.noteToEdit}
    onSuccess={handleEditSuccess}
  />
{/if}

<!-- Confirm Modal for delete -->
<ConfirmModal
  bind:open={homePage.showConfirmDelete}
  title={t("modal.deleteTitle")}
  message={t("modal.deleteMessage")}
  confirmText={t("modal.delete")}
  cancelText={t("modal.cancel")}
  danger={true}
  onConfirm={handleDeleteConfirm}
  onCancel={cancelDelete}
/>

<!-- Undo toast -->
{#if homePage.showUndoToast}
  <div class="undo-toast" class:undo-toast--restore={homePage.undoToastStage === "restore"}>
    {#if homePage.undoToastStage === "done"}
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
  .graph-content {
    display: flex;
    flex-direction: column;
    width: 100%;
    height: 100%;
    overflow: hidden;
    background: transparent;
    color: var(--color-text-dark);
  }

  .graph-view-wrapper {
    position: relative;
    flex: 1 1 auto;
    min-height: 0;
    width: 100%;
  }

  .graph-view-wrapper :global(canvas) {
    width: 100% !important;
    height: 100% !important;
  }

  /* List Container */
  .list-container {
    flex: 1 1 auto;
    min-height: 0;
    max-width: 1400px;
    margin: 0 auto;
    padding: 24px;
    width: 100%;
    overflow-y: auto;
    box-sizing: border-box;
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

    .list-container {
      padding: 16px;
      height: 100%;
    }
  }
</style>
