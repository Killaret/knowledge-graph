<script lang="ts">
  import { updateNote, getNote, type Note } from '$lib/api/notes';
  import { Modal, Button, TypeSelector } from './index';

  /* eslint-disable prefer-const */
  let {
    open = $bindable(false),
    noteId = $bindable(''),
    onSuccess
  }: {
    open: boolean;
    noteId: string;
    onSuccess?: (note: Note) => void;
  } = $props();

  let title = $state('');
  let content = $state('');
  let type = $state<'star' | 'planet' | 'comet' | 'galaxy' | 'asteroid'>('star');
  let loading = $state(false);
  let saving = $state(false);
  let error = $state('');

  // Загрузка данных при открытии
  $effect(() => {
    if (open && noteId) {
      loadNote();
    }
  });

  async function loadNote() {
    loading = true;
    error = '';
    try {
      const note = await getNote(noteId);
      title = note.title;
      content = note.content || '';
      type = (note.type as typeof type) || 'star';
    } catch {
      error = 'Failed to load note';
    } finally {
      loading = false;
    }
  }

  async function handleSubmit(e: Event) {
    e.preventDefault();
    if (!title.trim()) {
      error = 'Title is required';
      return;
    }

    saving = true;
    error = '';

    try {
      const note = await updateNote(noteId, {
        title: title.trim(),
        content: content.trim(),
        type: type,
        metadata: {}
      });

      onSuccess?.(note);
      close();
    } catch {
      error = 'Failed to update note';
    } finally {
      saving = false;
    }
  }

  function close() {
    open = false;
    error = '';
  }
</script>

<Modal bind:open title="Edit Note" onClose={close}>
  {#if loading}
    <div class="loading-state">
      <div class="spinner"></div>
      <span>Loading...</span>
    </div>
  {:else}
    <form onsubmit={handleSubmit}>
      {#if error}
        <div class="error-message" role="alert">
          {error}
        </div>
      {/if}

      <div class="form-group">
        <label for="edit-note-title">Title *</label>
        <input
          id="edit-note-title"
          type="text"
          bind:value={title}
          placeholder="Enter note title"
          disabled={saving}
          data-testid="edit-title-input"
        />
      </div>

      <div class="form-group">
        <label for="edit-note-type">Type</label>
        <TypeSelector id="edit-note-type" bind:selected={type} />
      </div>

      <div class="form-group">
        <label for="edit-note-content">Content</label>
        <textarea
          id="edit-note-content"
          bind:value={content}
          placeholder="Enter note content"
          rows="6"
          disabled={saving}
          data-testid="edit-content-input"
        ></textarea>
      </div>

      <div class="modal-footer">
        <Button variant="secondary" onClick={close} disabled={saving}>
          Cancel
        </Button>
        <Button variant="primary" type="submit" disabled={saving}>
          {saving ? 'Saving...' : 'Save Changes'}
        </Button>
      </div>
    </form>
  {/if}
</Modal>

<style>
  .loading-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem;
    gap: 1rem;
    color: var(--color-text-secondary, #6b7280);
  }

  .spinner {
    width: 32px;
    height: 32px;
    border: 3px solid var(--color-border, #e5e7eb);
    border-top-color: var(--color-primary, #3b82f6);
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .error-message {
    background: var(--color-danger-light, #fee2e2);
    color: var(--color-danger, #dc2626);
    padding: 0.75rem 1rem;
    border-radius: 6px;
    margin-bottom: 1rem;
    font-size: 0.875rem;
  }

  .form-group {
    margin-bottom: 1rem;
  }

  label {
    display: block;
    margin-bottom: 0.5rem;
    font-weight: 500;
    color: var(--color-text, #374151);
    font-size: 0.875rem;
  }

  input, textarea {
    width: 100%;
    padding: 0.625rem 0.875rem;
    border: 1px solid var(--color-border, #d1d5db);
    border-radius: 6px;
    font-size: 0.875rem;
    transition: border-color 0.2s, box-shadow 0.2s;
    font-family: inherit;
    color: var(--color-text, #1f2937);
    background: var(--color-surface, white);
  }

  input:focus, textarea:focus {
    outline: none;
    border-color: var(--color-primary, #3b82f6);
    box-shadow: 0 0 0 3px var(--color-primary-light, rgba(59, 130, 246, 0.1));
  }

  input:disabled, textarea:disabled {
    background: var(--color-surface-elevated, #f3f4f6);
    cursor: not-allowed;
  }

  textarea {
    resize: vertical;
    min-height: 120px;
  }

  .modal-footer {
    display: flex;
    justify-content: flex-end;
    gap: 0.75rem;
    margin-top: 1.5rem;
    padding-top: 1rem;
    border-top: 1px solid var(--color-border, #e5e7eb);
  }
</style>
