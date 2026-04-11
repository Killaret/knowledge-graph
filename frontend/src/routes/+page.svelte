<script lang="ts">
  import { onMount } from 'svelte';
  import type { Note } from '$lib/api/notes';
  import { getNotes, deleteNote } from '$lib/api/notes';
  import SearchBar from '$lib/components/SearchBar.svelte';
  import FloatingActionButton from '$lib/components/FloatingActionButton.svelte';
  import LeftSidebar from '$lib/components/LeftSidebar.svelte';
  import SearchPanel from '$lib/components/SearchPanel.svelte';
  import DocumentImport from '$lib/components/DocumentImport.svelte';
  import ExportModal from '$lib/components/ExportModal.svelte';

  // Реактивные переменные (Svelte 5 runes)
  let notes: Note[] = $state([]);
  let loading = $state(true);
  let error = $state('');
  
  // UI states
  let sidebarOpen = $state(false);
  let searchOpen = $state(false);
  let importOpen = $state(false);
  let exportOpen = $state(false);
  let searchResults: any[] = $state([]);

  onMount(async () => {
    try {
      notes = await getNotes();
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
      // Reload notes from server to ensure consistency
      notes = await getNotes();
    } catch {
      alert('Delete failed');
    }
  }
</script>

<LeftSidebar
  isOpen={sidebarOpen}
  onToggle={() => sidebarOpen = !sidebarOpen}
  onImportClick={() => { sidebarOpen = false; importOpen = true; }}
  onExportClick={() => { sidebarOpen = false; exportOpen = true; }}
/>

<button 
  class="search-toggle-btn"
  onclick={() => searchOpen = true}
  aria-label="Открыть поиск"
>
  <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
    <circle cx="11" cy="11" r="8"></circle>
    <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
  </svg>
  Поиск
</button>

<h1>My Notes</h1>

<SearchBar placeholder="Search through your notes..." />

{#if loading}
  <p>Loading...</p>
{:else if error}
  <p class="error">{error}</p>
{:else}
  <div class="notes-grid">
    {#each notes as note}
      <div class="note-card">
        <a href={`/notes/${note.id}`}>
          <h3>{note.title}</h3>
          <p>{note.content.slice(0, 100)}...</p>
        </a>
        <div class="actions">
          <a href={`/notes/${note.id}/edit`}>Edit</a>
          <button onclick={() => handleDelete(note.id)}>Delete</button>
        </div>
      </div>
    {/each}
  </div>
{/if}

<FloatingActionButton />

<SearchPanel
  isOpen={searchOpen}
  onClose={() => searchOpen = false}
  onSearch={(query) => {
    // Filter notes locally for demo
    searchResults = notes.filter(n => 
      n.title.toLowerCase().includes(query.toLowerCase()) ||
      n.content.toLowerCase().includes(query.toLowerCase())
    );
  }}
  onResultClick={(result) => {
    window.location.href = `/notes/${result.id}`;
  }}
  results={searchResults}
/>

<DocumentImport
  isOpen={importOpen}
  onClose={() => importOpen = false}
  onImport={(files, url) => {
    console.log('Import:', { files, url });
    // Mock: add dummy note after import
    const newNote: Note = {
      id: 'imported-' + Date.now(),
      title: 'Импортированная заметка',
      content: 'Содержимое из импортированного документа...',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      metadata: {}
    };
    notes = [...notes, newNote];
  }}
/>

<ExportModal
  isOpen={exportOpen}
  notes={notes}
  onClose={() => exportOpen = false}
  onExport={(format, ids) => {
    console.log('Export:', { format, ids });
    alert(`Экспортировано ${ids.length} заметок в формате ${format.toUpperCase()}`);
  }}
/>

<style>
  .notes-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 1.5rem; margin: 2rem 0; }
  .note-card { border: 1px solid #ddd; border-radius: 8px; padding: 1rem; background: white; transition: box-shadow 0.2s; }
  .note-card:hover { box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
  .note-card a { text-decoration: none; color: inherit; }
  .actions { margin-top: 0.5rem; display: flex; gap: 0.5rem; }
  .error { color: red; }
  
  .search-toggle-btn {
    position: fixed;
    top: 16px;
    right: 16px;
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 16px;
    background: rgba(10, 26, 58, 0.8);
    border: 1px solid rgba(255, 221, 136, 0.2);
    border-radius: 10px;
    color: #ffdd88;
    cursor: pointer;
    font-size: 0.9rem;
    z-index: 1400;
    transition: all 0.2s;
  }
  
  .search-toggle-btn:hover {
    background: rgba(255, 221, 136, 0.1);
    border-color: rgba(255, 221, 136, 0.4);
  }
</style>