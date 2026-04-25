<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import FloatingControls from '$lib/components/FloatingControls.svelte';
  import NoteSidePanel from '$lib/components/NoteSidePanel.svelte';
  import CreateNoteModal from '$lib/components/CreateNoteModal.svelte';
  import EditNoteModal from '$lib/components/EditNoteModal.svelte';
  import ConfirmModal from '$lib/components/ConfirmModal.svelte';
  import NoteCard from '$lib/components/NoteCard.svelte';
  import { getNotes, deleteNote, searchNotes, type Note } from '$lib/api/notes';
  import { getFullGraphData, type GraphData } from '$lib/api/graph';
  import GraphCanvas from '$lib/components/GraphCanvas.svelte';

  // State
  let allNotes: Note[] = $state([]);
  let filteredNotes: Note[] = $state([]);
  let loading = $state(true);
  let error = $state('');
  let selectedNodeId: string | null = $state(null);
  let showCreateModal = $state(false);
  let showEditModal = $state(false);
  let noteToEdit: string | null = $state(null);
  let showConfirmDelete = $state(false);
  let noteToDelete: string | null = $state(null);
  let currentView: 'graph' | 'list' = $state('graph');  // Graph-first interface
  
  // Graph state - always show full graph on main page
  let graphData: GraphData = $state({ nodes: [], links: [] });
  let graphLoading = $state(false);
  let searchQuery = $state('');
  
  // Filter and sort state
  let selectedType = $state<string>('all');
  const sortOption = $state<string>('newest');

  const typeFilters = [
    { id: 'all', label: 'All', emoji: '🌌' },
    { id: 'star', label: 'Stars', emoji: '⭐' },
    { id: 'planet', label: 'Planets', emoji: '🪐' },
    { id: 'comet', label: 'Comets', emoji: '☄️' },
    { id: 'galaxy', label: 'Galaxies', emoji: '🌀' }
  ];

  // NOTE: The sortOptions constant was previously defined here but is not currently used.
  // These are the available sorting options for the notes list view:
  // - newest: Sort by creation date, newest first
  // - oldest: Sort by creation date, oldest first  
  // - az: Alphabetical sorting A-Z
  // - za: Alphabetical sorting Z-A
  // Functionality: Provides sorting options for the notes list view UI
  /*
  const sortOptions = [
    { id: 'newest', label: 'Newest first' },
    { id: 'oldest', label: 'Oldest first' },
    { id: 'az', label: 'Alphabetical (A-Z)' },
    { id: 'za', label: 'Alphabetical (Z-A)' }
  ];
  */
  onMount(async () => {
    if (!browser) return;
    await loadDataParallel();
  });

  async function loadDataParallel() {
    try {
      // Load notes and graph data in parallel
      const [notesResult, graphResult] = await Promise.all([
        getNotes(),
        getFullGraphData().catch(e => {
          console.error('[+page] Failed to load graph:', e);
          return null;
        })
      ]);
      
      allNotes = notesResult;
      applyFiltersAndSort();
      
      // Set graph data if successful
      if (graphResult) {
        graphData = graphResult;
        console.log('[+page] Full graph loaded:', graphData.nodes.length, 'nodes,', graphData.links.length, 'links');
      } else {
        // Fallback: build simple graph from notes
        graphData = {
          nodes: allNotes.map(n => ({ id: n.id, title: n.title, type: n.type || 'star' })),
          links: []
        };
      }
    } catch (e) {
      error = 'Failed to load notes';
      console.error(e);
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
    } catch (e) {
      error = 'Failed to load notes';
      console.error(e);
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
      const rawData = await getFullGraphData();

      // Transform nodes: backend might return Id/id/ID in different cases
      const transformedNodes = rawData.nodes.map((n: any) => ({
        id: n.id || n.Id || n.ID,
        title: n.title || n.Title,
        type: n.type ?? n.Type ?? 'star'
      }));

      // Transform links: backend returns source_note_id/target_note_id, frontend expects source/target
      const transformedLinks = rawData.links.map((l: any) => ({
        source: l.source_note_id || l.source,
        target: l.target_note_id || l.target,
        weight: l.weight,
        link_type: l.link_type
      }));

      graphData = {
        nodes: transformedNodes,
        links: transformedLinks
      };

    } catch (e) {
      console.error('[+page] Failed to load graph:', e);
      // Fallback: build simple graph from notes
      graphData = {
        nodes: allNotes.map(n => ({ id: n.id, title: n.title, type: n.type || 'star' })),
        links: []
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

  // Helper to get note type from type field or metadata
  function getNoteType(note: Note): string {
    return note.type ?? (note.metadata?.type as string) ?? 'star';
  }

  // Reactive filtered graph data based on selected type
  const filteredGraphData = $derived(() => {
    if (selectedType === 'all' || !graphData.nodes.length) {
      return graphData;
    }
    
    const allowedNodeIds = new Set(
      allNotes.filter(n => getNoteType(n) === selectedType).map(n => n.id)
    );
    
    const filteredNodes = graphData.nodes.filter(n => allowedNodeIds.has(n.id));
    const filteredNodeIds = new Set(filteredNodes.map(n => n.id));
    const filteredLinks = graphData.links.filter(l => 
      filteredNodeIds.has(l.source) && filteredNodeIds.has(l.target)
    );
    
    console.log(`[FilteredGraph] Type: ${selectedType}, nodes: ${filteredNodes.length}, links: ${filteredLinks.length}`);
    
    return {
      nodes: filteredNodes,
      links: filteredLinks
    };
  });

  function applyFiltersAndSort() {
    let result = [...allNotes];

    // Apply type filter
    if (selectedType !== 'all') {
      result = result.filter(n => getNoteType(n) === selectedType);
    }

    // Apply search
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      result = result.filter(n => 
        n.title.toLowerCase().includes(query) || 
        n.content.toLowerCase().includes(query)
      );
    }
    
    // Expose to window for tests
    if (browser) {
      (window as any).filteredNotes = result;
    }

    // Apply sorting
    switch (sortOption) {
      case 'newest':
        result.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
        break;
      case 'oldest':
        result.sort((a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime());
        break;
      case 'az':
        result.sort((a, b) => a.title.localeCompare(b.title));
        break;
      case 'za':
        result.sort((a, b) => b.title.localeCompare(a.title));
        break;
    }

    filteredNotes = result;
  }

  async function handleSearch() {
    if (!searchQuery.trim()) {
      await loadNotes();
      return;
    }
    
    try {
      const response = await searchNotes(searchQuery, 1, 20);
      allNotes = response.data;
      applyFiltersAndSort();
    } catch (e) {
      console.error('Search error:', e);
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
      allNotes = allNotes.filter(n => n.id !== noteToDelete);
      filteredNotes = filteredNotes.filter(n => n.id !== noteToDelete);
      // Then reload from server to ensure sync
      await loadNotes();
    } catch {
      if (browser) {
        alert('Failed to delete note');
      }
    } finally {
      noteToDelete = null;
      showConfirmDelete = false;
    }
  }

  function handleNoteCreated(note: Note) {
    showCreateModal = false;
    selectedNodeId = note.id;
    loadNotes();
  }

  function handleToggleView() {
    currentView = currentView === 'graph' ? 'list' : 'graph';
  }
</script>

<!-- Main page container - root element for the page layout -->
<!-- Functionality: Provides full viewport height/width container with hidden overflow -->
<div class="page-container">

<!-- Floating Controls with Filters -->
  <FloatingControls
    onCreate={() => { showCreateModal = true; }}
    onSearch={(query: string) => { searchQuery = query; handleSearch(); }}
    onToggleView={handleToggleView}
    onFilter={(type: string) => { selectedType = type; applyFiltersAndSort(); }}
    noteId={selectedNodeId ?? undefined}
    typeFilters={typeFilters}
    selectedType={selectedType}
    currentView={currentView}
    typeCounts={Object.fromEntries(typeFilters.map(f => [f.id, f.id === 'all' ? allNotes.length : allNotes.filter(n => getNoteType(n) === f.id).length]))}
  />

  <!-- Fullscreen Graph Container -->
  <div class="fullscreen-graph" data-testid="graph-2d-container">
    {#if loading}
      <div class="center">
        <div class="spinner"></div>
        <p>Loading notes...</p>
      </div>
    {:else if error}
      <p class="error">{error}</p>
    {:else if currentView === 'graph'}
      <!-- Fullscreen 2D Graph View -->
      {#if graphLoading}
        <div class="center">
          <div class="spinner"></div>
          <p>Loading graph...</p>
        </div>
      {:else if filteredGraphData().nodes.length > 0}
        <GraphCanvas 
          nodes={filteredGraphData().nodes}
          links={filteredGraphData().links}
          onNodeClick={(node) => selectedNodeId = node.id}
        />
        <!-- Stats Overlay -->
        <div class="graph-stats-overlay" data-testid="graph-stats">
          <span class="stat-item"><strong>{filteredGraphData().nodes.length}</strong> nodes</span>
          <span class="stat-item"><strong>{filteredGraphData().links.length}</strong> links</span>
          {#if selectedType !== 'all'}
            <span class="stat-filter">{typeFilters.find(f => f.id === selectedType)?.label}</span>
          {/if}
        </div>
      {:else}
        <div class="empty-state">
          <div class="empty-icon">🌌</div>
          <h2>No graph data</h2>
          <p>{selectedType === 'all' 
            ? "Create some notes to see the knowledge graph" 
            : `No ${typeFilters.find(f => f.id === selectedType)?.label.toLowerCase()} in the graph. Try selecting a different type.`}
          </p>
        </div>
      {/if}

    {:else if currentView === 'list'}
      <!-- List View -->
      <div class="list-container" data-testid="list-container">
        {#if filteredNotes.length === 0}
          <div class="empty-state" data-testid="empty-state">
            <div class="empty-icon">🌌</div>
            <h2>No notes found</h2>
            <p>
              {selectedType === 'all' && !searchQuery
                ? "You haven't created any notes yet."
                : searchQuery
                  ? `No notes match "${searchQuery}".`
                  : `No ${typeFilters.find(f => f.id === selectedType)?.label.toLowerCase()} found.`}
            </p>
            <button class="new-note-button" onclick={() => showCreateModal = true}>
              Create your first note
            </button>
          </div>
        {:else}
          <div class="notes-grid" data-testid="notes-grid">
            {#each filteredNotes as note (note.id)}
              <NoteCard
                {note}
                onClick={() => selectedNodeId = note.id}
                highlightQuery={searchQuery}
              />
            {/each}
          </div>
        {/if}
      </div>
    {/if}
  </div>
</div>

<!-- Side Panel for selected note -->
{#if selectedNodeId}
  <NoteSidePanel 
    nodeId={selectedNodeId} 
    onClose={() => selectedNodeId = null}
    onEdit={(id) => { noteToEdit = id; showEditModal = true; }}
    onDelete={handleDeleteRequest}
  />
{/if}

<!-- Create Note Modal -->
<CreateNoteModal 
  bind:open={showCreateModal}
  onSuccess={handleNoteCreated}
/>

<!-- Edit Note Modal -->
{#if noteToEdit}
  <EditNoteModal 
    bind:open={showEditModal}
    noteId={noteToEdit}
    onSuccess={() => { showEditModal = false; noteToEdit = null; }}
  />
{/if}

<!-- Confirm Modal for delete -->
<ConfirmModal
  bind:open={showConfirmDelete}
  title="Delete Note?"
  message="Are you sure you want to delete this note? This action cannot be undone."
  confirmText="Delete"
  cancelText="Cancel"
  danger={true}
  onConfirm={handleDeleteConfirm}
  onCancel={() => { showConfirmDelete = false; noteToDelete = null; }}
/>

<style>
  .page-container {
    height: 100vh;
    width: 100vw;
    overflow: hidden;
    position: relative;
  }

  /* Fullscreen Graph Container */
  .fullscreen-graph {
    position: fixed;
    top: 80px; /* Space for floating controls */
    left: 0;
    right: 0;
    bottom: 0;
    width: 100%;
    height: calc(100vh - 80px);
  }

  .fullscreen-graph :global(canvas) {
    width: 100% !important;
    height: 100% !important;
  }

  /* Stats Overlay on Graph */
  .graph-stats-overlay {
    position: absolute;
    bottom: 20px;
    left: 20px;
    display: flex;
    gap: 16px;
    padding: 10px 16px;
    background: rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(10px);
    border-radius: 8px;
    color: #94a3b8;
    font-size: 14px;
    z-index: 10;
  }

  .stat-item strong {
    color: #88aaff;
    font-weight: 600;
  }

  .stat-filter {
    color: #64748b;
    font-style: italic;
  }

  /* List Container */
  .list-container {
    max-width: 1400px;
    margin: 0 auto;
    padding: 24px;
    height: calc(100vh - 80px);
    overflow-y: auto;
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
    to { transform: rotate(360deg); }
  }

  .error {
    padding: 20px;
    background: #fee2e2;
    color: #dc2626;
    border-radius: 8px;
    text-align: center;
    margin: 20px auto;
    max-width: 600px;
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

  .empty-icon {
    font-size: 64px;
    margin-bottom: 16px;
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
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 16px;
    padding: 16px 0;
  }

  @media (max-width: 768px) {
    .fullscreen-graph {
      top: 70px;
      height: calc(100vh - 70px);
    }

    .list-container {
      padding: 16px;
    }

    .graph-stats-overlay {
      bottom: 10px;
      left: 10px;
      right: 10px;
      flex-wrap: wrap;
    }
  }
</style>
