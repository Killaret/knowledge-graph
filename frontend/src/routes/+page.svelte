<script lang="ts">
  import { onMount } from 'svelte';
  import type { Note } from '$lib/api/notes';
  import { getNotes, deleteNote } from '$lib/api/notes';
  import SearchBar from '$lib/components/SearchBar.svelte';
  import NoteCard from '$lib/components/NoteCard.svelte';

  // Реактивные переменные (Svelte 5 runes)
  let allNotes: Note[] = $state([]);
  let filteredNotes: Note[] = $state([]);
  let loading = $state(true);
  let error = $state('');

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

  function applyFiltersAndSort() {
    let result = [...allNotes];

    // Apply type filter
    if (selectedType !== 'all') {
      result = result.filter(n => n.type === selectedType);
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

  $effect(() => {
    if (allNotes.length > 0) {
      applyFiltersAndSort();
    }
  });

  onMount(async () => {
    try {
      allNotes = await getNotes();
      filteredNotes = allNotes;
    } catch (e) {
      error = 'Failed to load notes';
      console.error(e);
    } finally {
      loading = false;
    }
  });

  async function handleDelete(id: string) {
    if (!confirm('Delete this note?')) return;
    try {
      await deleteNote(id);
      allNotes = await getNotes();
      applyFiltersAndSort();
    } catch {
      alert('Delete failed');
    }
  }
</script>

<div class="page-header">
  <h1>My Knowledge Graph</h1>
  <SearchBar placeholder="Search through your notes..." />
</div>

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
    <!-- Type Filters -->
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

    <!-- Sort Dropdown -->
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
    <span class="stats-total">{filteredNotes.length} notes</span>
    {#if selectedType !== 'all'}
      <span class="stats-filter">(filtered by: {typeFilters.find(f => f.id === selectedType)?.label})</span>
    {/if}
  </div>

  <!-- Notes Grid -->
  {#if filteredNotes.length === 0}
    <div class="empty-state">
      <div class="empty-icon">🌌</div>
      <h2>No notes found</h2>
      <p>
        {selectedType === 'all'
          ? "You haven't created any notes yet."
          : `No ${typeFilters.find(f => f.id === selectedType)?.label.toLowerCase()} found.`}
      </p>
      <a href="/notes/new" class="new-note-button">Create your first note</a>
    </div>
  {:else}
    <div class="notes-grid">
      {#each filteredNotes as note (note.id)}
        <NoteCard {note} onDelete={handleDelete} />
      {/each}
    </div>
  {/if}
{/if}

<style>
  .notes-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 1.5rem; margin: 2rem 0; }
  .note-card { border: 1px solid #ddd; border-radius: 8px; padding: 1rem; background: white; transition: box-shadow 0.2s; }
  .controls-bar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem; }
  .filter-group { display: flex; align-items: center; }
  .filter-label { margin-right: 0.5rem; }
  .filter-buttons { display: flex; }
  .filter-button { background: none; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; }
  .filter-button:hover { background: #f0f0f0; }
  .filter-button.active { background: #3b82f6; color: white; }
  .sort-group { display: flex; align-items: center; }
  .sort-label { margin-right: 0.5rem; }
  .sort-select { padding: 0.5rem; border: 1px solid #ddd; border-radius: 4px; }
  .stats-bar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem; }
  .stats-total { font-weight: bold; }
  .stats-filter { color: #666; }
  .empty-state { display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100vh; }
  .empty-icon { font-size: 48px; margin-bottom: 1rem; }
  .new-note-button { background: #3b82f6; color: white; padding: 0.5rem 1rem; border-radius: 4px; text-decoration: none; }
  .note-card a { text-decoration: none; color: inherit; }
  .actions { margin-top: 0.5rem; display: flex; gap: 0.5rem; }
  .new-note { display: inline-block; background: #3b82f6; color: white; padding: 0.5rem 1rem; border-radius: 4px; text-decoration: none; }
  .error { color: red; }
</style>