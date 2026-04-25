<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import { page } from '$app/stores';
  import { getNote, getSuggestions, deleteNote } from '$lib/api/notes';
  import type { Note, Suggestion } from '$lib/api/notes';
  import { goto } from '$app/navigation';
  import { formatDate, formatDateTime } from '$lib/utils/date';
  import BackButton from '$lib/components/BackButton.svelte';
  import EditNoteModal from '$lib/components/EditNoteModal.svelte';

  let note: Note | null = $state(null);
  let suggestions: Suggestion[] = $state([]);
  let loading = $state(true);
  let error = $state('');
  let editModalOpen = $state(false);

  function getRouteId(): string {
    const id = $page.params.id;
    if (!id) throw new Error('Missing route parameter: id');
    return id;
  }

  onMount(async () => {
    try {
      const id = getRouteId();
      note = await getNote(id);
      suggestions = await getSuggestions(id, 5);
    } catch (e: any) {
      if (e.response?.status === 404) {
        error = 'Заметка не найдена';
        setTimeout(() => goto('/'), 3000);
      } else {
        error = 'Note not found';
      }
    } finally {
      loading = false;
    }
  });

  async function handleDelete() {
    if (!browser) return;
    if (!confirm('Delete this note?')) return;
    const id = getRouteId();
    await deleteNote(id);
    await goto('/');
  }
</script>

{#if loading}
  <p>Loading...</p>
{:else if error}
  <p class="error">{error}</p>
{:else if note}
  <div class="note-container">
    <BackButton href="/" />
    <h1>{note.title}</h1>
    <div class="meta">Создано: {formatDateTime(note.created_at)}</div>
    <div class="content">{note.content}</div>
    <div class="actions">
      <button onclick={() => editModalOpen = true} class="edit-btn">Edit</button>
      <button onclick={handleDelete}>Delete</button>
      <a href={`/graph/3d/${note.id}`} class="graph-link">✨ Show constellation</a>
    </div>

    <EditNoteModal
      bind:open={editModalOpen}
      noteId={note.id}
      onSuccess={(updatedNote) => note = updatedNote}
    />

    {#if suggestions.length}
      <h2>Similar notes</h2>
      <ul class="suggestions">
        {#each suggestions as s}
          <li><a href={`/notes/${s.note_id}`}>{s.title}</a> <span class="score">score: {s.score.toFixed(3)}</span></li>
        {/each}
      </ul>
    {/if}
  </div>
{/if}

<style>
  .note-container { max-width: 800px; margin: 0 auto; }
  .content { white-space: pre-wrap; margin: 1rem 0; }
  .actions { display: flex; gap: 1rem; margin: 1rem 0; }
  .actions button {
    padding: 0.5rem 1rem;
    border: 1px solid var(--color-border);
    border-radius: 4px;
    cursor: pointer;
    background: var(--color-surface-elevated);
  }
  .actions button:hover {
    background: var(--color-background);
  }
  .edit-btn {
    background: var(--color-primary) !important;
    color: white;
    border-color: var(--color-primary) !important;
  }
  .edit-btn:hover {
    background: var(--color-primary-hover) !important;
  }
  .suggestions li { margin-bottom: 0.5rem; }
  .score { margin-left: 1rem; color: var(--color-text-secondary); font-size: 0.9rem; }
  .error { color: var(--color-danger); }
  .graph-link { background: var(--color-galaxy); color: white; padding: 0.25rem 0.75rem; border-radius: 4px; text-decoration: none; }
</style>