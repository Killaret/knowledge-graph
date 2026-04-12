<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import type { Note } from '$lib/api/notes';
  import { getNotes, deleteNote, searchNotes } from '$lib/api/notes';
  import { Skeleton } from 'flowbite-svelte';
  import ConfirmDialog from '$lib/components/ui/ConfirmDialog.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';

  // Data
  let notes = $state<Note[]>([]);
  let filteredNotes = $state<Note[]>([]);
  let loading = $state(true);
  let error = $state('');

  // Filters
  let searchQuery = $state('');
  let selectedType = $state<string>('all');
  let selectedTag = $state<string>('all');
  let dateFrom = $state<string>('');
  let dateTo = $state<string>('');
  
  // Sorting
  let sortBy = $state<'date' | 'alpha' | 'links'>('date');
  let sortOrder = $state<'asc' | 'desc'>('desc');
  
  // Grouping
  let groupBy = $state<'none' | 'type' | 'date'>('none');
  
  // Pagination
  let currentPage = $state(1);
  let pageSize = $state(12);
  let totalPages = $state(1);
  
  // Confirm dialog
  let showConfirmDialog = $state(false);
  let noteToDelete = $state<string | null>(null);

  // Available types and tags
  const noteTypes = ['star', 'planet', 'comet', 'galaxy'];
  let availableTags = $state<string[]>([]);

  onMount(async () => {
    try {
      notes = await getNotes();
      applyFilters();
      
      // Extract unique tags from metadata
      const tags = new Set<string>();
      notes.forEach(note => {
        if (note.metadata?.tags) {
          note.metadata.tags.forEach((tag: string) => tags.add(tag));
        }
      });
      availableTags = Array.from(tags);
    } catch (e) {
      error = 'Failed to load notes';
      console.error(e);
    } finally {
      loading = false;
    }
  });

  // Apply all filters
  function applyFilters() {
    let result = [...notes];
    
    // Search filter
    if (searchQuery.trim()) {
      const query = searchQuery.toLowerCase();
      result = result.filter(n => 
        n.title.toLowerCase().includes(query) ||
        n.content.toLowerCase().includes(query)
      );
    }
    
    // Type filter
    if (selectedType !== 'all') {
      result = result.filter(n => n.metadata?.type === selectedType);
    }
    
    // Tag filter
    if (selectedTag !== 'all') {
      result = result.filter(n => 
        n.metadata?.tags?.includes(selectedTag)
      );
    }
    
    // Date range filter
    if (dateFrom) {
      result = result.filter(n => new Date(n.created_at) >= new Date(dateFrom));
    }
    if (dateTo) {
      result = result.filter(n => new Date(n.created_at) <= new Date(dateTo));
    }
    
    // Sorting
    result.sort((a, b) => {
      let comparison = 0;
      switch (sortBy) {
        case 'date':
          comparison = new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
          break;
        case 'alpha':
          comparison = a.title.localeCompare(b.title);
          break;
        case 'links':
          const linksA = a.metadata?.links?.length || 0;
          const linksB = b.metadata?.links?.length || 0;
          comparison = linksA - linksB;
          break;
      }
      return sortOrder === 'asc' ? comparison : -comparison;
    });
    
    // Calculate pagination
    totalPages = Math.ceil(result.length / pageSize) || 1;
    if (currentPage > totalPages) currentPage = totalPages;
    
    // Apply pagination
    const start = (currentPage - 1) * pageSize;
    filteredNotes = result.slice(start, start + pageSize);
  }

  // Group notes for display
  function getGroupedNotes() {
    if (groupBy === 'none') {
      return { 'Все заметки': filteredNotes };
    }
    
    const groups: Record<string, Note[]> = {};
    
    filteredNotes.forEach(note => {
      let key = '';
      switch (groupBy) {
        case 'type':
          key = note.metadata?.type || 'Без типа';
          break;
        case 'date':
          const date = new Date(note.created_at);
          key = date.toLocaleDateString('ru-RU', { month: 'long', year: 'numeric' });
          break;
      }
      if (!groups[key]) groups[key] = [];
      groups[key].push(note);
    });
    
    return groups;
  }

  function handleDelete(id: string) {
    noteToDelete = id;
    showConfirmDialog = true;
  }

  async function confirmDelete() {
    if (!noteToDelete) return;
    try {
      await deleteNote(noteToDelete);
      notes = notes.filter(n => n.id !== noteToDelete);
      applyFilters();
    } catch {
      alert('Delete failed');
    } finally {
      showConfirmDialog = false;
      noteToDelete = null;
    }
  }

  function cancelDelete() {
    showConfirmDialog = false;
    noteToDelete = null;
  }

  function getTypeIcon(type: string): string {
    const icons: Record<string, string> = {
      star: '⭐',
      planet: '🪐',
      comet: '☄️',
      galaxy: '🌌',
    };
    return icons[type] || '📝';
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString('ru-RU');
  }

  // Re-apply filters when any filter changes
  $effect(() => {
    searchQuery;
    selectedType;
    selectedTag;
    dateFrom;
    dateTo;
    sortBy;
    sortOrder;
    currentPage;
    pageSize;
    
    if (!loading) {
      applyFilters();
    }
  });
</script>

<div class="notes-page">
  <header class="page-header">
    <h1>Все звёзды</h1>
    <button class="btn-new" onclick={() => goto('/notes/new')}>
      <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
        <line x1="12" y1="5" x2="12" y2="19"/>
        <line x1="5" y1="12" x2="19" y2="12"/>
      </svg>
      Новая звезда
    </button>
  </header>

  <!-- Filters -->
  <div class="filters-bar">
    <div class="filter-group">
      <input
        type="text"
        bind:value={searchQuery}
        placeholder="Поиск по названию или содержимому..."
        class="search-input"
      />
    </div>
    
    <div class="filter-row">
      <select bind:value={selectedType} class="filter-select">
        <option value="all">Все типы</option>
        {#each noteTypes as type}
          <option value={type}>{getTypeIcon(type)} {type}</option>
        {/each}
      </select>
      
      <select bind:value={selectedTag} class="filter-select">
        <option value="all">Все теги</option>
        {#each availableTags as tag}
          <option value={tag}>{tag}</option>
        {/each}
      </select>
      
      <input
        type="date"
        bind:value={dateFrom}
        class="filter-date"
        title="От"
      />
      <input
        type="date"
        bind:value={dateTo}
        class="filter-date"
        title="До"
      />
      
      <select bind:value={sortBy} class="filter-select">
        <option value="date">По дате</option>
        <option value="alpha">По алфавиту</option>
        <option value="links">По связям</option>
      </select>
      
      <button
        class="sort-order-btn"
        onclick={() => sortOrder = sortOrder === 'asc' ? 'desc' : 'asc'}
        title={sortOrder === 'asc' ? 'По возрастанию' : 'По убыванию'}
      >
        {sortOrder === 'asc' ? '↑' : '↓'}
      </button>
      
      <select bind:value={groupBy} class="filter-select">
        <option value="none">Без группировки</option>
        <option value="type">По типу</option>
        <option value="date">По дате</option>
      </select>
    </div>
  </div>

  <!-- Results count -->
  <div class="results-info">
    Показано {filteredNotes.length} из {notes.length} заметок
  </div>

  <!-- Notes list -->
  {#if loading}
    <div class="loading-grid">
      {#each Array(6) as _}
        <div class="skeleton-card">
          <Skeleton size="md" class="mb-2" />
          <Skeleton size="sm" class="w-3/4" />
        </div>
      {/each}
    </div>
  {:else if error}
    <p class="error">{error}</p>
  {:else if filteredNotes.length === 0}
    <EmptyState
      icon="🔭"
      title="Ничего не найдено"
      message="Попробуйте изменить параметры фильтра или создайте новую заметку"
      actionText="Создать заметку"
      onAction={() => goto('/notes/new')}
    />
  {:else}
    <div class="notes-content">
      {#each Object.entries(getGroupedNotes()) as [groupName, groupNotes]}
        {#if groupBy !== 'none'}
          <h2 class="group-header">{groupName}</h2>
        {/if}
        
        <div class="notes-grid">
          {#each groupNotes as note}
            <div class="note-card" onclick={() => goto(`/notes/${note.id}`)}>
              <div class="note-header">
                <span class="note-type">{getTypeIcon(note.metadata?.type)}</span>
                <h3 class="note-title">{note.title}</h3>
              </div>
              <p class="note-preview">{note.content.slice(0, 120)}...</p>
              <div class="note-meta">
                <span class="note-date">{formatDate(note.created_at)}</span>
                {#if note.metadata?.tags}
                  <div class="note-tags">
                    {#each note.metadata.tags.slice(0, 3) as tag}
                      <span class="tag">{tag}</span>
                    {/each}
                  </div>
                {/if}
              </div>
              <div class="note-actions" onclick={(e) => e.stopPropagation()}>
                <button onclick={() => goto(`/notes/${note.id}/edit`)} class="btn-edit">
                  Редактировать
                </button>
                <button onclick={() => handleDelete(note.id)} class="btn-delete">
                  Удалить
                </button>
              </div>
            </div>
          {/each}
        </div>
      {/each}
    </div>
    
    <!-- Pagination -->
    {#if totalPages > 1}
      <div class="pagination">
        <button
          onclick={() => currentPage--}
          disabled={currentPage === 1}
          class="page-btn"
        >
          ←
        </button>
        
        {#each Array(totalPages) as _, i}
          <button
            onclick={() => currentPage = i + 1}
            class="page-btn"
            class:active={currentPage === i + 1}
          >
            {i + 1}
          </button>
        {/each}
        
        <button
          onclick={() => currentPage++}
          disabled={currentPage === totalPages}
          class="page-btn"
        >
          →
        </button>
      </div>
    {/if}
  {/if}
</div>

<ConfirmDialog
  open={showConfirmDialog}
  title="Удалить заметку?"
  message="Вы уверены, что хотите удалить эту заметку?"
  confirmText="Удалить"
  cancelText="Отмена"
  onConfirm={confirmDelete}
  onCancel={cancelDelete}
/>

<style>
  .notes-page {
    padding: 1.5rem;
    max-width: 1400px;
    margin: 0 auto;
  }

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
  }

  .page-header h1 {
    font-size: 1.75rem;
    font-weight: 600;
    color: white;
    margin: 0;
  }

  .btn-new {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.625rem 1.25rem;
    background: rgba(99, 102, 241, 0.8);
    border: none;
    border-radius: 0.75rem;
    color: white;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-new:hover {
    background: rgba(99, 102, 241, 1);
    transform: translateY(-1px);
  }

  .filters-bar {
    background: rgba(0, 0, 0, 0.3);
    backdrop-filter: blur(8px);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 1rem;
    padding: 1rem;
    margin-bottom: 1rem;
  }

  .filter-group {
    margin-bottom: 0.75rem;
  }

  .search-input {
    width: 100%;
    padding: 0.625rem 1rem;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 0.5rem;
    color: white;
    font-size: 0.9375rem;
  }

  .search-input::placeholder {
    color: rgba(255, 255, 255, 0.4);
  }

  .filter-row {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    align-items: center;
  }

  .filter-select,
  .filter-date {
    padding: 0.5rem 0.75rem;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 0.5rem;
    color: white;
    font-size: 0.875rem;
  }

  .sort-order-btn {
    width: 2rem;
    height: 2rem;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 0.5rem;
    color: white;
    cursor: pointer;
  }

  .results-info {
    color: rgba(255, 255, 255, 0.5);
    font-size: 0.875rem;
    margin-bottom: 1rem;
  }

  .loading-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 1rem;
  }

  .skeleton-card {
    padding: 1rem;
    background: rgba(255, 255, 255, 0.05);
    border-radius: 0.75rem;
  }

  .group-header {
    font-size: 1.125rem;
    font-weight: 600;
    color: rgba(255, 255, 255, 0.8);
    margin: 1.5rem 0 0.75rem;
    padding-bottom: 0.5rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .notes-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 1rem;
  }

  .note-card {
    background: rgba(0, 0, 0, 0.3);
    backdrop-filter: blur(8px);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 1rem;
    padding: 1.25rem;
    cursor: pointer;
    transition: all 0.2s;
  }

  .note-card:hover {
    background: rgba(255, 255, 255, 0.05);
    border-color: rgba(255, 255, 255, 0.2);
    transform: translateY(-2px);
  }

  .note-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.75rem;
  }

  .note-type {
    font-size: 1.25rem;
  }

  .note-title {
    font-size: 1.0625rem;
    font-weight: 600;
    color: white;
    margin: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .note-preview {
    font-size: 0.875rem;
    color: rgba(255, 255, 255, 0.6);
    line-height: 1.5;
    margin: 0 0 0.75rem;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .note-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    align-items: center;
    margin-bottom: 0.75rem;
  }

  .note-date {
    font-size: 0.75rem;
    color: rgba(255, 255, 255, 0.4);
  }

  .note-tags {
    display: flex;
    gap: 0.375rem;
    flex-wrap: wrap;
  }

  .tag {
    font-size: 0.6875rem;
    padding: 0.25rem 0.5rem;
    background: rgba(255, 255, 255, 0.1);
    border-radius: 0.25rem;
    color: rgba(255, 255, 255, 0.7);
  }

  .note-actions {
    display: flex;
    gap: 0.5rem;
  }

  .btn-edit,
  .btn-delete {
    flex: 1;
    padding: 0.5rem;
    font-size: 0.8125rem;
    border: none;
    border-radius: 0.5rem;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-edit {
    background: rgba(255, 255, 255, 0.1);
    color: rgba(255, 255, 255, 0.8);
  }

  .btn-edit:hover {
    background: rgba(255, 255, 255, 0.2);
  }

  .btn-delete {
    background: rgba(239, 68, 68, 0.2);
    color: #ef4444;
  }

  .btn-delete:hover {
    background: rgba(239, 68, 68, 0.3);
  }

  .pagination {
    display: flex;
    justify-content: center;
    gap: 0.375rem;
    margin-top: 2rem;
  }

  .page-btn {
    width: 2.5rem;
    height: 2.5rem;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 0.5rem;
    color: rgba(255, 255, 255, 0.7);
    cursor: pointer;
    transition: all 0.2s;
  }

  .page-btn:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.15);
  }

  .page-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }

  .page-btn.active {
    background: rgba(99, 102, 241, 0.8);
    border-color: rgba(99, 102, 241, 0.5);
    color: white;
  }

  .error {
    color: #ef4444;
    text-align: center;
    padding: 2rem;
  }

  @media (max-width: 768px) {
    .notes-page {
      padding: 1rem;
    }

    .filter-row {
      flex-direction: column;
      align-items: stretch;
    }

    .filter-select,
    .filter-date {
      width: 100%;
    }

    .notes-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
