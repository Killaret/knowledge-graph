<script lang="ts">
  import { createNote, type Note } from '$lib/api/notes';
  import { Modal, Button, TypeSelector } from './index';

  /* eslint-disable prefer-const */
  let {
    open = $bindable(false),
    onSuccess
  }: {
    open: boolean;
    onSuccess?: (note: Note) => void;
  } = $props();
  
  let title = $state('');
  let content = $state('');
  let type = $state<'star' | 'planet' | 'comet' | 'galaxy' | 'asteroid'>('star');
  let loading = $state(false);
  let error = $state('');
  
  async function handleSubmit(e: Event) {
    e.preventDefault();
    if (!title.trim()) {
      error = 'Title is required';
      return;
    }
    
    loading = true;
    error = '';
    
    try {
      const note = await createNote({ 
        title: title.trim(), 
        content: content.trim(),
        type: type,
        metadata: {}
      });
      
      onSuccess?.(note);
      close();
    } catch {
      error = 'Failed to create note';
    } finally {
      loading = false;
    }
  }
  
  function close() {
    open = false;
    title = '';
    content = '';
    type = 'star';
    error = '';
  }
</script>

<Modal bind:open title="Create New Note" onClose={close}>
  <form onsubmit={handleSubmit}>
    <div class="form-group">
      <label for="note-title">Title *</label>
      <input
        id="note-title"
        name="title"
        type="text"
        bind:value={title}
        placeholder="Enter note title..."
        disabled={loading}
      />
    </div>
    
    <div class="form-group">
      <label>Type</label>
      <TypeSelector bind:selected={type} />
    </div>
    
    <div class="form-group">
      <label for="note-content">Content</label>
      <textarea
        id="note-content"
        name="content"
        bind:value={content}
        placeholder="Enter note content..."
        rows={6}
        disabled={loading}
      ></textarea>
    </div>
    
    {#if error}
      <div class="error">{error}</div>
    {/if}
    
    <div class="form-actions">
      <Button variant="secondary" onClick={close} disabled={loading}>
        Cancel
      </Button>
      <Button variant="primary" type="submit" disabled={loading}>
        {loading ? 'Creating...' : 'Create Note'}
      </Button>
    </div>
  </form>
</Modal>

<style>
  .form-group {
    margin-bottom: 20px;
  }

  label {
    display: block;
    font-size: 14px;
    font-weight: 500;
    color: var(--color-text, #374151);
    margin-bottom: 6px;
  }

  input, textarea {
    width: 100%;
    padding: 10px 14px;
    border: 1px solid var(--color-border, #d1d5db);
    border-radius: 8px;
    font-size: 15px;
    color: var(--color-text, #1f2937);
    background: var(--color-surface, white);
    transition: border-color 0.2s;
  }

  input:focus, textarea:focus {
    outline: none;
    border-color: var(--color-primary, #3b82f6);
    box-shadow: 0 0 0 3px var(--color-primary-light, rgba(59, 130, 246, 0.1));
  }

  textarea {
    resize: vertical;
    font-family: inherit;
  }

  .error {
    padding: 12px;
    background: var(--color-danger-light, #fee2e2);
    color: var(--color-danger, #dc2626);
    border-radius: 8px;
    font-size: 14px;
    margin-bottom: 20px;
  }

  .form-actions {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }
</style>
