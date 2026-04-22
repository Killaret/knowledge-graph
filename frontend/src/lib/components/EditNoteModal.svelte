<script lang="ts">
  import { updateNote, getNote, type Note } from '$lib/api/notes';

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

  const types: Array<{ value: typeof type, label: string, color: string }> = [
    { value: 'star', label: '⭐ Star', color: '#fbbf24' },
    { value: 'planet', label: '🪐 Planet', color: '#60a5fa' },
    { value: 'comet', label: '☄️ Comet', color: '#f472b6' },
    { value: 'galaxy', label: '🌀 Galaxy', color: '#a78bfa' },
    { value: 'asteroid', label: '☁️ Asteroid', color: '#94a3b8' }
  ];

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

{#if open}
  <div class="modal-overlay" role="button" tabindex="0" onclick={close} onkeydown={(e) => e.key === 'Enter' && close()}>
    <div class="modal" role="dialog" tabindex="-1" onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.key === 'Escape' && close()}>
      <div class="modal-header">
        <h2>Edit Note</h2>
        <button class="close-btn" onclick={close} aria-label="Close">
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>

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
            <select id="edit-note-type" bind:value={type} disabled={saving} data-testid="edit-type-select">
              {#each types as t}
                <option value={t.value}>{t.label}</option>
              {/each}
            </select>
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
            <button type="button" class="btn-secondary" onclick={close} disabled={saving} data-testid="edit-cancel-btn">
              Cancel
            </button>
            <button type="submit" class="btn-primary" disabled={saving} data-testid="edit-save-btn">
              {saving ? 'Saving...' : 'Save Changes'}
            </button>
          </div>
        </form>
      {/if}
    </div>
  </div>
{/if}

<style>
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 1rem;
  }

  .modal {
    background: white;
    border-radius: 12px;
    width: 100%;
    max-width: 500px;
    max-height: 90vh;
    overflow-y: auto;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.5rem;
    border-bottom: 1px solid #e5e7eb;
  }

  .modal-header h2 {
    margin: 0;
    font-size: 1.25rem;
    font-weight: 600;
    color: #111827;
  }

  .close-btn {
    background: none;
    border: none;
    cursor: pointer;
    padding: 0.5rem;
    border-radius: 6px;
    color: #6b7280;
    transition: all 0.2s;
  }

  .close-btn:hover {
    background: #f3f4f6;
    color: #111827;
  }

  .loading-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem;
    gap: 1rem;
    color: #6b7280;
  }

  .spinner {
    width: 32px;
    height: 32px;
    border: 3px solid #e5e7eb;
    border-top-color: #3b82f6;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  form {
    padding: 1.5rem;
  }

  .error-message {
    background: #fee2e2;
    color: #dc2626;
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
    color: #374151;
    font-size: 0.875rem;
  }

  input, select, textarea {
    width: 100%;
    padding: 0.625rem 0.875rem;
    border: 1px solid #d1d5db;
    border-radius: 6px;
    font-size: 0.875rem;
    transition: border-color 0.2s, box-shadow 0.2s;
    font-family: inherit;
  }

  input:focus, select:focus, textarea:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  input:disabled, select:disabled, textarea:disabled {
    background: #f3f4f6;
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
    border-top: 1px solid #e5e7eb;
  }

  .btn-primary, .btn-secondary {
    padding: 0.625rem 1.25rem;
    border-radius: 6px;
    font-weight: 500;
    font-size: 0.875rem;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-primary {
    background: #3b82f6;
    color: white;
    border: none;
  }

  .btn-primary:hover:not(:disabled) {
    background: #2563eb;
  }

  .btn-secondary {
    background: white;
    color: #374151;
    border: 1px solid #d1d5db;
  }

  .btn-secondary:hover:not(:disabled) {
    background: #f9fafb;
  }

  .btn-primary:disabled, .btn-secondary:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
</style>
