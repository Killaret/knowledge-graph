<script lang="ts">
  import { goto } from '$app/navigation';
  import { createNote } from '$lib/api/notes';

  let title = $state('');
  let content = $state('');
  let saving = $state(false);
  let error = $state('');

  async function handleSubmit() {
    if (!title.trim()) {
      error = 'Title is required';
      return;
    }
    saving = true;
    error = '';
    try {
      const note = await createNote({ title, content, metadata: {} });
      goto(`/notes/${note.id}`);
    } catch (e) {
      error = 'Failed to create note';
    } finally {
      saving = false;
    }
  }
</script>

<h1>New Note</h1>

{#if error}
  <p class="error">{error}</p>
{/if}

<form onsubmit={handleSubmit}>
  <input type="text" placeholder="Title" bind:value={title} required />
  <textarea placeholder="Content (supports [[wiki links]])" bind:value={content} rows="15"></textarea>
  <button type="submit" disabled={saving}>{saving ? 'Saving...' : 'Create'}</button>
</form>

<style>
  input, textarea { width: 100%; padding: 0.5rem; margin-bottom: 1rem; border: 1px solid #ccc; border-radius: 4px; font-family: inherit; }
  .error { color: red; }
</style>