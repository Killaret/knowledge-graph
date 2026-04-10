<script lang="ts">
  import { onMount } from 'svelte';
  import type { Note } from '$lib/api/notes';
  import { getNotes, deleteNote } from '$lib/api/notes';
  import SearchBar from '$lib/components/SearchBar.svelte';

  // Реактивные переменные (Svelte 5 runes)
  let notes: Note[] = $state([]);
  let loading = $state(true);
  let error = $state('');

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
    } catch (e) {
      alert('Delete failed');
    }
  }
</script>

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
  <a href="/notes/new" class="new-note">+ New Note</a>
{/if}

<style>
  .notes-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 1.5rem; margin: 2rem 0; }
  .note-card { border: 1px solid #ddd; border-radius: 8px; padding: 1rem; background: white; transition: box-shadow 0.2s; }
  .note-card:hover { box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
  .note-card a { text-decoration: none; color: inherit; }
  .actions { margin-top: 0.5rem; display: flex; gap: 0.5rem; }
  .new-note { display: inline-block; background: #3b82f6; color: white; padding: 0.5rem 1rem; border-radius: 4px; text-decoration: none; }
  .error { color: red; }
</style>