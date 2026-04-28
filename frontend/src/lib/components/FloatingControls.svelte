<script lang="ts">
  import { goto } from '$app/navigation';
  
  const {
    onCreate,
    onSearch,
    onToggleView,
    onFilter,
    onImport,
    onExport,
    noteId = '',
    typeFilters = [],
    selectedType = 'all',
    typeCounts = {},
    currentView = 'graph'
  }: {
    onCreate?: () => void;
    onSearch?: (query: string) => void;
    onToggleView?: () => void;
    onFilter?: (type: string) => void;
    onImport?: () => void;
    onExport?: () => void;
    noteId?: string;
    typeFilters?: Array<{ id: string; label: string; emoji: string }>;
    selectedType?: string;
    typeCounts?: Record<string, number>;
    currentView?: 'graph' | 'list';
  } = $props();
  
  let searchQuery = $state('');
  let showMenu = $state(false);
  
  function handleSearch() {
    onSearch?.(searchQuery);
  }
  
  function handleToggle3D() {
    // Если есть noteId, переходим на 3D граф этой заметки, иначе на общий 3D граф
    const targetUrl = noteId ? `/graph/3d/${noteId}` : '/graph/3d';
    goto(targetUrl);
  }
  
  function toggleView() {
    onToggleView?.();
  }

  function handleFilter(typeId: string) {
    onFilter?.(typeId);
  }
</script>

<div class="floating-controls">
  <!-- View Toggle: 2D / 3D / List -->
  <div class="view-toggle">
    <button
      type="button"
      class="toggle-btn {currentView === 'graph' ? 'active' : ''}"
      onclick={toggleView}
      title="2D Graph"
      data-testid="view-toggle-graph"
      aria-pressed={currentView === 'graph'}
      aria-label="Switch to 2D graph view"
    >
      <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="3"/>
        <circle cx="5" cy="5" r="2"/>
        <circle cx="19" cy="5" r="2"/>
        <circle cx="5" cy="19" r="2"/>
        <circle cx="19" cy="19" r="2"/>
        <line x1="7" y1="7" x2="10" y2="10"/>
        <line x1="14" y1="10" x2="17" y2="7"/>
        <line x1="7" y1="17" x2="10" y2="14"/>
        <line x1="14" y1="14" x2="17" y2="17"/>
      </svg>
      <span class="btn-label">2D</span>
    </button>
    <button
      type="button"
      class="toggle-btn"
      onclick={handleToggle3D}
      title={noteId ? "3D Graph for selected note" : "Full 3D Graph"}
      data-testid="view-toggle-3d"
      aria-label={noteId ? "Open 3D graph for selected note" : "Open full 3D graph"}
    >
      <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M12 2L2 7l10 5 10-5-10-5z"/>
        <path d="M2 17l10 5 10-5"/>
        <path d="M2 12l10 5 10-5"/>
      </svg>
      <span class="btn-label">3D</span>
    </button>
    <button
      type="button"
      class="toggle-btn {currentView === 'list' ? 'active' : ''}"
      onclick={toggleView}
      title="List View"
      data-testid="view-toggle-list"
      aria-pressed={currentView === 'list'}
      aria-label="Switch to list view"
    >
      <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <line x1="3" y1="6" x2="21" y2="6"/>
        <line x1="3" y1="12" x2="21" y2="12"/>
        <line x1="3" y1="18" x2="21" y2="18"/>
      </svg>
      <span class="btn-label">List</span>
    </button>
  </div>

  <!-- Type Filters -->
  {#if typeFilters.length > 0}
    <div class="type-filters">
      {#each typeFilters as filter}
        <button
          type="button"
          class="filter-chip {selectedType === filter.id ? 'active' : ''}"
          onclick={() => handleFilter(filter.id)}
          title={filter.label}
          data-testid="filter-chip-{filter.id}"
          aria-pressed={selectedType === filter.id}
          aria-label={`Filter by ${filter.label}`}
        >
          <span class="filter-emoji">{filter.emoji}</span>
          <span class="filter-label">{filter.label}</span>
          {#if typeCounts[filter.id] !== undefined}
            <span class="filter-count" data-testid="filter-count-{filter.id}">{typeCounts[filter.id]}</span>
          {/if}
        </button>
      {/each}
    </div>
  {/if}

  <!-- Search -->
  <div class="search-container">
    <input
      type="text"
      bind:value={searchQuery}
      placeholder="Search notes..."
      onkeyup={(e) => e.key === 'Enter' && handleSearch()}
      class="search-input"
      data-testid="search-input"
      aria-label="Search notes"
    />
    <button type="button" class="search-btn" onclick={handleSearch} aria-label="Search">
      <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="11" cy="11" r="8"/>
        <line x1="21" y1="21" x2="16.65" y2="16.65"/>
      </svg>
    </button>
  </div>

  <!-- Menu -->
  <div class="menu-container">
    <button 
      type="button"
      class="menu-btn"
      onclick={() => showMenu = !showMenu}
      title="Menu"
      aria-expanded={showMenu}
      aria-haspopup="true"
      aria-label="Open menu"
    >
      <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <line x1="3" y1="12" x2="21" y2="12"/>
        <line x1="3" y1="6" x2="21" y2="6"/>
        <line x1="3" y1="18" x2="21" y2="18"/>
      </svg>
    </button>
    
    {#if showMenu}
      <div class="dropdown-menu" role="menu">
        <button type="button" class="menu-item" role="menuitem" onclick={() => { onImport?.(); showMenu = false; }}>
          Import
        </button>
        <button type="button" class="menu-item" role="menuitem" onclick={() => { onExport?.(); showMenu = false; }}>
          Export
        </button>
        <button type="button" class="menu-item" role="menuitem" onclick={() => { handleToggle3D(); showMenu = false; }}>
          {noteId ? '3D View for Note' : 'Full 3D View'}
        </button>
      </div>
    {/if}
  </div>

  <!-- Create Button -->
  <button type="button" class="create-btn" onclick={() => onCreate?.()} title="Create new note" data-testid="create-note-button" aria-label="Create new note">
    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
      <line x1="12" y1="5" x2="12" y2="19"/>
      <line x1="5" y1="12" x2="19" y2="12"/>
    </svg>
  </button>
</div>

<style>
  .floating-controls {
    position: fixed;
    top: 20px;
    left: 50%;
    transform: translateX(-50%);
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 20px;
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(10px);
    border-radius: 50px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
    z-index: 100;
  }

  .view-toggle {
    display: flex;
    gap: 4px;
    padding: 4px;
    background: #f1f5f9;
    border-radius: 25px;
  }

  .toggle-btn {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 8px 12px;
    border: none;
    background: transparent;
    border-radius: 20px;
    cursor: pointer;
    transition: all 0.2s;
    color: #64748b;
    font-size: 12px;
  }

  .toggle-btn:hover {
    background: rgba(255, 255, 255, 0.5);
  }

  .toggle-btn.active {
    background: white;
    color: #3b82f6;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .btn-label {
    font-weight: 500;
  }

  .search-container {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 4px 4px 4px 16px;
    background: #f8fafc;
    border-radius: 25px;
    border: 1px solid #e2e8f0;
  }

  .search-input {
    border: none;
    background: transparent;
    outline: none;
    font-size: 14px;
    width: 180px;
    color: #334155;
  }

  .search-input::placeholder {
    color: #94a3b8;
  }

  .search-btn {
    padding: 8px;
    border: none;
    background: transparent;
    cursor: pointer;
    color: #64748b;
    border-radius: 50%;
    transition: all 0.2s;
  }

  .search-btn:hover {
    background: #e2e8f0;
    color: #334155;
  }

  .menu-container {
    position: relative;
  }

  .menu-btn {
    padding: 10px;
    border: none;
    background: transparent;
    cursor: pointer;
    color: #64748b;
    border-radius: 50%;
    transition: all 0.2s;
  }

  .menu-btn:hover {
    background: #f1f5f9;
    color: #334155;
  }

  .dropdown-menu {
    position: absolute;
    top: 100%;
    right: 0;
    margin-top: 8px;
    background: white;
    border-radius: 12px;
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.15);
    padding: 8px;
    min-width: 140px;
  }

  .menu-item {
    display: block;
    width: 100%;
    padding: 10px 16px;
    border: none;
    background: transparent;
    text-align: left;
    cursor: pointer;
    border-radius: 8px;
    font-size: 14px;
    color: #334155;
    transition: background 0.2s;
  }

  .menu-item:hover {
    background: #f1f5f9;
  }

  .create-btn {
    padding: 12px;
    border: none;
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    color: white;
    border-radius: 50%;
    cursor: pointer;
    transition: all 0.2s;
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
  }

  .create-btn:hover {
    transform: scale(1.05);
    box-shadow: 0 6px 16px rgba(59, 130, 246, 0.5);
  }

  /* Type Filters */
  .type-filters {
    display: flex;
    gap: 6px;
    align-items: center;
  }

  .filter-chip {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 6px 10px;
    border: 1px solid #e2e8f0;
    background: white;
    border-radius: 16px;
    cursor: pointer;
    transition: all 0.2s;
    font-size: 12px;
    color: #64748b;
  }

  .filter-chip:hover {
    background: #f8fafc;
    border-color: #cbd5e1;
  }

  .filter-chip.active {
    background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
    color: white;
    border-color: transparent;
  }

  .filter-emoji {
    font-size: 14px;
  }

  .filter-label {
    font-weight: 500;
  }

  .filter-count {
    font-size: 11px;
    background: rgba(0, 0, 0, 0.1);
    padding: 2px 6px;
    border-radius: 10px;
    margin-left: 2px;
    font-weight: 600;
  }

  .filter-chip.active .filter-count {
    background: rgba(255, 255, 255, 0.2);
  }
</style>
