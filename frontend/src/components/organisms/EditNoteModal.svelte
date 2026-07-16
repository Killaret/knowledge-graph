<script lang="ts">
  import { updateNote, getNote, type Note } from '$shared/api/notes';
  import Button from '$components/atoms/Button.svelte';
import Modal from '$components/atoms/Modal.svelte';
import TypeSelector from '$components/molecules/TypeSelector.svelte';
  import ApiErrorDisplay from '$components/atoms/ApiErrorDisplay.svelte';
  import type { ErrorResponse } from '$shared/types/errors';
  import { getMessage, mode } from '$shared/stores/lexicon-settings';

  /* eslint-disable prefer-const -- Svelte 5 $bindable() requires let, not const, see: https://svelte.dev/docs/svelte/$bindable */
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
  let type = $state<'star' | 'planet' | 'comet' | 'galaxy' | 'asteroid' | 'satellite' | 'debris' | 'nebula' | 'dust'>('star');
  let loading = $state(false);
  let saving = $state(false);
  let apiError = $state<ErrorResponse | null>(null);
  
  let modalTitle = $state('Edit Note');
  let titleLabel = $state('Title *');
  let typeLabel = $state('Type');
  let contentLabel = $state('Content');
  let cancelText = $state('Cancel');
  let saveText = $state('Save Changes');
  let savingText = $state('Saving...');
  let loadingText = $state('Loading...');
  let titlePlaceholder = $state('Enter note title');
  let contentPlaceholder = $state('Enter note content');

  // Update labels based on galactic mode
  $effect(() => {
    let currentMode = 'standard';
    mode.subscribe(m => currentMode = m)();
    
    if (currentMode === 'galactic') {
      modalTitle = 'Recalibrate Orbit';
      titleLabel = 'Star Name *';
      typeLabel = 'Celestial Type';
      contentLabel = 'Star Data';
      cancelText = 'Abort Mission';
      saveText = 'Update Orbit';
      savingText = 'Recalibrating...';
      loadingText = 'Scanning star...';
      titlePlaceholder = 'Enter star name';
      contentPlaceholder = 'Enter star data';
    } else {
      modalTitle = 'Edit Note';
      titleLabel = 'Title *';
      typeLabel = 'Type';
      contentLabel = 'Content';
      cancelText = 'Cancel';
      saveText = 'Save Changes';
      savingText = 'Saving...';
      loadingText = 'Loading...';
      titlePlaceholder = 'Enter note title';
      contentPlaceholder = 'Enter note content';
    }
  });

  // Загрузка данных при открытии
  $effect(() => {
    if (open && noteId) {
      loadNote();
    }
  });

  async function loadNote() {
    loading = true;
    apiError = null;
    try {
      const note = await getNote(noteId);
      title = note.title;
      content = note.content || '';
      type = (note.type as typeof type) || 'star';
    } catch (err: any) {
      apiError = err?.response?.data || { code: 'API_ERROR', message: 'Failed to load note' };
    } finally {
      loading = false;
    }
  }

  async function handleSubmit(e: Event) {
    e.preventDefault();
    if (!title.trim()) {
      const msg = await getMessage('error', 'validation', 'title');
      apiError = { code: 'VALIDATION_ERROR', message: msg };
      return;
    }

    saving = true;
    apiError = null;

    try {
      const note = await updateNote(noteId, {
        title: title.trim(),
        content: content.trim(),
        type: type,
        metadata: {}
      });

      onSuccess?.(note);
      close();
    } catch (err: any) {
      apiError = err?.response?.data || { code: 'API_ERROR', message: 'Failed to update note' };
    } finally {
      saving = false;
    }
  }

  function close() {
    open = false;
    apiError = null;
  }
</script>

<Modal bind:open title={modalTitle} onClose={close}>
  {#if loading}
    <div class="loading-state">
      <div class="spinner"></div>
      <span>{loadingText}</span>
    </div>
  {:else}
    <form onsubmit={handleSubmit}>
      <ApiErrorDisplay error={apiError} onClose={() => apiError = null} />

      <div class="form-group">
        <label for="edit-note-title">{titleLabel}</label>
        <input
          id="edit-note-title"
          type="text"
          bind:value={title}
          placeholder={titlePlaceholder}
          disabled={saving}
          data-testid="edit-title-input"
        />
      </div>

      <div class="form-group">
        <label for="edit-note-type">{typeLabel}</label>
        <TypeSelector id="edit-note-type" bind:selected={type} />
      </div>

      <div class="form-group">
        <label for="edit-note-content">{contentLabel}</label>
        <textarea
          id="edit-note-content"
          bind:value={content}
          placeholder={contentPlaceholder}
          rows="6"
          disabled={saving}
          data-testid="edit-content-input"
        ></textarea>
      </div>

      <div class="modal-footer">
        <Button variant="secondary" onClick={close} disabled={saving}>
          {cancelText}
        </Button>
        <Button variant="primary" type="submit" disabled={saving}>
          {saving ? savingText : saveText}
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
