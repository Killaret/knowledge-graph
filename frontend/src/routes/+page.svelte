<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import { goto } from '$app/navigation';
  import FloatingControls from '$lib/components/FloatingControls.svelte';
  import NoteSidePanel from '$lib/components/NoteSidePanel.svelte';
  import CreateNoteModal from '$lib/components/CreateNoteModal.svelte';
  import ConfirmModal from '$lib/components/ConfirmModal.svelte';
  import NoteCard from '$lib/components/NoteCard.svelte';
  import { getNotes, deleteNote, searchNotes, type Note } from '$lib/api/notes';
  import { getGraphData, getFullGraphData, type GraphData } from '$lib/api/graph';
  import GraphCanvas from '$lib/components/GraphCanvas.svelte';

  // State
  let allNotes: Note[] = $state([]);
  let filteredNotes: Note[] = $state([]);
  let loading = $state(true);
  let error = $state('');
  let selectedNodeId: string | null = $state(null);
  let showCreateModal = $state(false);
  let showConfirmDelete = $state(false);
  let noteToDelete: string | null = $state(null);
  let currentView: 'graph' | 'list' = $state('graph');  // Graph-first interface
  
  // Graph state
  let graphData: GraphData = $state({ nodes: [], links: [] });
  let graphLoading = $state(false);
  let searchQuery = $state('');
  let showFullGraph = $state(false); // Toggle between full graph and local view
  
  // Filter and sort state
  let selectedType = $state<string>('all');
  let sortOption = $state<string>('newest');

  const typeFilters = [
    { id: 'all', label: 'All', emoji: '🌌' },
    { id: 'star', label: 'Stars', emoji: '⭐' },
    { id: 'planet', label: 'Planets', emoji: '🪐' },
    { id: 'comet', label: 'Comets', emoji: '☄️' },
    { id: 'galaxy', label: 'Galaxies', emoji: '🌀' }
  ];

  const sortOptions = [
    { id: 'newest', label: 'Newest first' },
    { id: 'oldest', label: 'Oldest first' },
    { id: 'az', label: 'Alphabetical (A-Z)' },
    { id: 'za', label: 'Alphabetical (Z-A)' }
  ];

  onMount(async () => {
    if (!browser) return;
    await loadNotes();
  });

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
    if (allNotes.length === 0) return;
    
    graphLoading = true;
    try {
      if (showFullGraph) {
        // Load full graph of all notes
        graphData = await getFullGraphData();
      } else {
        // Build local graph from all notes - use first note as center
        const centerNote = allNotes[0];
        const data = await getGraphData(centerNote.id, 2);
        graphData = data;
      }
    } catch (e) {
      console.error('Failed to load graph:', e);
      // Fallback: build simple graph from notes
      graphData = {
        nodes: allNotes.map(n => ({ id: n.id, title: n.title, type: n.type || 'star' })),
        links: []
      };
    } finally {
      graphLoading = false;
    }
  }
  
  function toggleGraphMode() {
    showFullGraph = !showFullGraph;
    loadGraphData();
  }

  function applyFiltersAndSort() {
    let result = [...allNotes];

    // Apply type filter
    if (selectedType !== 'all') {
      result = result.filter(n => n.type === selectedType);
    }

    // Apply search
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      result = result.filter(n => 
        n.title.toLowerCase().includes(query) || 
        n.content.toLowerCase().includes(query)
      );
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

  function handleToggle3D() {
    goto('/graph');
  }

  function getPluralForm(count: number, one: string, few: string, many: string): string {
    const lastDigit = count % 10;
    const lastTwoDigits = count % 100;
    
    if (lastTwoDigits >= 11 && lastTwoDigits <= 14) {
      return many;
    }
    if (lastDigit === 1) {
      return one;
    }
    if (lastDigit >= 2 && lastDigit <= 4) {
      return few;
    }
    return many;
  }
</script>

<div class="page-container">
  <!-- Floating Controls -->
  <FloatingControls
    onCreate={() => { showCreateModal = true; }}
    onSearch={(query: string) => { searchQuery = query; handleSearch(); }}
    onToggleView={handleToggleView}
    onToggle3D={handleToggle3D}
  />

  <!-- Main Content -->
  <div class="content-area">
    {#if loading}
      <div class="center">
        <div class="spinner"></div>
        <p>Loading notes...</p>
      </div>
    {:else if error}
      <p class="error">{error}</p>
    {:else}
      <!-- Filters and Sort -->
      <div class="controls-bar">
        <div class="filter-group">
          <span class="filter-label">Filter:</span>
          <div class="filter-buttons">
            {#each typeFilters as filter}
              <button
                class="filter-button"
                class:active={selectedType === filter.id}
                onclick={() => { selectedType = filter.id; applyFiltersAndSort(); }}
              >
                <span class="emoji">{filter.emoji}</span>
                <span>{filter.label}</span>
              </button>
            {/each}
          </div>
        </div>

        <div class="sort-group">
          <label for="sort-select" class="sort-label">Sort:</label>
          <select
            id="sort-select"
            class="sort-select"
            bind:value={sortOption}
            onchange={applyFiltersAndSort}
          >
            {#each sortOptions as option}
              <option value={option.id}>{option.label}</option>
            {/each}
          </select>
        </div>
      </div>

      <!-- Stats -->
      <div class="stats-bar">
        <span class="stats-total">{filteredNotes.length} {getPluralForm(filteredNotes.length, 'note', 'notes', 'notes')}</span>
        {#if filteredNotes.length !== allNotes.length}
          <span class="stats-of">of {allNotes.length} total</span>
        {/if}
        {#if selectedType !== 'all'}
          <span class="stats-filter">(filtered by: {typeFilters.find(f => f.id === selectedType)?.label})</span>
        {/if}
        {#if searchQuery}
          <span class="stats-filter">(search: "{searchQuery}")</span>
        {/if}
      </div>

      <!-- Graph Mode Toggle (Full Graph / Local View) -->
      {#if currentView === 'graph'}
        <div class="graph-mode-toggle">
          <label class="toggle-label">
            <input 
              type="checkbox" 
              checked={showFullGraph} 
              onchange={toggleGraphMode}
            />
            <span class="toggle-text">Показать все заметки ({showFullGraph ? 'включено' : 'выключено'})</span>
          </label>
        </div>
      {/if}

      <!-- Graph View (Primary) -->
      {#if currentView === 'graph' && !loading}
        <div class="graph-container" style="height: 600px; margin-bottom: 40px;">
          {#if graphLoading}
            <div class="center">
              <div class="spinner"></div>
              <p>Loading graph...</p>
            </div>
          {:else if graphData.nodes.length > 0}
            <GraphCanvas 
              nodes={graphData.nodes}
              links={graphData.links}
              onNodeClick={(node) => selectedNodeId = node.id}
            />
          {:else}
            <div class="empty-state">
              <div class="empty-icon">🌌</div>
              <h2>No graph data</h2>
              <p>Create some notes to see the knowledge graph</p>
            </div>
          {/if}
        </div>
      {/if}

      <!-- Notes Grid (List View Only) -->
      {#if currentView === 'list'}
        {#if filteredNotes.length === 0}
          <div class="empty-state">
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
          <div class="notes-grid">
            {#each filteredNotes as note (note.id)}
              <NoteCard
                {note}
                onClick={() => selectedNodeId = note.id}
                highlightQuery={searchQuery}
              />
            {/each}
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
    onClose={() => selectedNodeId = null}
    onEdit={(id) => goto(`/notes/${id}/edit`)}
    onDelete={handleDeleteRequest}
  />
{/if}

<!-- Create Note Modal -->
<CreateNoteModal 
  bind:open={showCreateModal}
  onSuccess={handleNoteCreated}
/>

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
    min-height: 100vh;
    padding-top: 100px;
    background: linear-gradient(135deg, #f8fafc 0%, #e2e8f0 100%);
  }

  .content-area {
    max-width: 1400px;
    margin: 0 auto;
    padding: 0 24px 40px;
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
    margin: 20px 0;
  }

  .controls-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    flex-wrap: wrap;
    gap: 16px;
  }

  .filter-group {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .filter-label {
    font-size: 14px;
    font-weight: 500;
    color: #64748b;
  }

  .filter-buttons {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .filter-button {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 16px;
    background: white;
    border: 1px solid #e2e8f0;
    border-radius: 20px;
    font-size: 14px;
    cursor: pointer;
    transition: all 0.2s;
  }

  .filter-button:hover {
    background: #f8fafc;
    border-color: #cbd5e1;
  }

  .filter-button.active {
    background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
    color: white;
    border-color: transparent;
  }

  .emoji {
    font-size: 16px;
  }

  .sort-group {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .sort-label {
    font-size: 14px;
    color: #64748b;
  }

  .sort-select {
    padding: 8px 12px;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    font-size: 14px;
    background: white;
    cursor: pointer;
  }

  .stats-bar {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 24px;
    font-size: 14px;
  }

  .stats-total {
    font-weight: 600;
    color: #1e293b;
  }

  .stats-filter {
    color: #64748b;
  }

  .stats-of {
    color: #94a3b8;
    font-size: 13px;
  }

  .graph-mode-toggle {
    display: flex;
    justify-content: center;
    margin-bottom: 16px;
    padding: 12px 16px;
    background: linear-gradient(135deg, rgba(59, 130, 246, 0.1) 0%, rgba(139, 92, 246, 0.1) 100%);
    border-radius: 8px;
    border: 1px solid rgba(59, 130, 246, 0.2);
  }

  .toggle-label {
    display: flex;
    align-items: center;
    gap: 10px;
    cursor: pointer;
    font-size: 14px;
    font-weight: 500;
    color: #1e293b;
  }

  .toggle-label input {
    width: 18px;
    height: 18px;
    cursor: pointer;
    accent-color: #3b82f6;
  }

  .toggle-text {
    user-select: none;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 80px 20px;
    text-align: center;
  }

  .empty-icon {
    font-size: 64px;
    margin-bottom: 16px;
  }

  .empty-state h2 {
    font-size: 24px;
    font-weight: 600;
    color: #1e293b;
    margin: 0 0 8px;
  }

  .empty-state p {
    color: #64748b;
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
    gap: 24px;
  }

  @media (max-width: 768px) {
    .content-area {
      padding: 0 16px 32px;
    }

    .controls-bar {
      flex-direction: column;
      align-items: stretch;
    }

    .filter-buttons {
      justify-content: center;
    }

    .notes-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
