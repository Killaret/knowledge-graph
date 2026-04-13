<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { getNote, updateNote } from '$lib/api/notes';
  import type { Note } from '$lib/api/notes';

  let note: Note | null = $state(null);
  let title = $state('');
  let content = $state('');
  let saving = $state(false);
  let error = $state('');
  let loading = $state(true);

  function getRouteId(): string {
    const id = $page.params.id;
    if (!id) throw new Error('Missing route parameter: id');
    return id;
  }

  onMount(async () => {
    try {
      const id = getRouteId();
      note = await getNote(id);
      title = note.title;
      content = note.content;
    } catch {
      error = 'Note not found';
    } finally {
      loading = false;
    }
  });

async function handleSubmit(event: Event) {
    event.preventDefault();
    const id = getRouteId();
    if (!title.trim()) {
        error = 'Title is required';
        return;
    }
    saving = true;
    error = '';
    try {
        await updateNote(id, { title, content });
        goto(`/notes/${id}`);
    } catch {
        error = 'Failed to update note';
    } finally {
        saving = false;
    }
}
</script>

<h1>Edit Note</h1>

{#if loading}
  <p>Loading...</p>
{:else if error}
  <p class="error">{error}</p>
{:else}
  <form onsubmit={handleSubmit}>
    <input type="text" name="title" placeholder="Title" bind:value={title} required />
    <textarea name="content" bind:value={content} rows="15"></textarea>
    <button type="submit" disabled={saving}>{saving ? 'Saving...' : 'Update'}</button>
  </form>
{/if}

<style>
  input, textarea { width: 100%; padding: 0.5rem; margin-bottom: 1rem; border: 1px solid #ccc; border-radius: 4px; font-family: inherit; }
  .error { color: red; }
</style>