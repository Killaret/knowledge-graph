<script lang="ts">
  import { createNote, type Note } from '$lib/api/notes';
  import { Modal, Button, TypeSelector } from './index';
  import ApiErrorDisplay from './ApiErrorDisplay.svelte';
  import type { ErrorResponse } from '$lib/types/errors';
  import { getMessage, mode } from '$lib/stores/lexicon-settings';

  /* eslint-disable prefer-const -- Svelte 5 $bindable() requires let, not const, see: https://svelte.dev/docs/svelte/$bindable */
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
  let apiError = $state<ErrorResponse | null>(null);
  let modalTitle = $state('Create New Note');
  let titleLabel = $state('Title *');
  let typeLabel = $state('Type');
  let contentLabel = $state('Content');
  let cancelText = $state('Cancel');
  let createText = $state('Create Note');
  let creatingText = $state('Creating...');
  let titlePlaceholder = $state('Enter note title...');
  let contentPlaceholder = $state('Enter note content...');

  // Update labels based on galactic mode
  $effect(async () => {
    let currentMode = 'standard';
    mode.subscribe(m => currentMode = m)();
    
    if (currentMode === 'galactic') {
      modalTitle = 'Ignite New Star';
      titleLabel = 'Star Name *';
      typeLabel = 'Celestial Type';
      contentLabel = 'Star Data';
      cancelText = 'Abort Mission';
      createText = 'Ignite Star';
      creatingText = 'Igniting...';
      titlePlaceholder = 'Enter star name...';
      contentPlaceholder = 'Enter star data...';
    } else {
      modalTitle = 'Create New Note';
      titleLabel = 'Title *';
      typeLabel = 'Type';
      contentLabel = 'Content';
      cancelText = 'Cancel';
      createText = 'Create Note';
      creatingText = 'Creating...';
      titlePlaceholder = 'Enter note title...';
      contentPlaceholder = 'Enter note content...';
    }
  });
  
  async function handleSubmit(e: Event) {
    e.preventDefault();
    if (!title.trim()) {
      const msg = await getMessage('error', 'validation', 'title');
      apiError = { code: 'VALIDATION_ERROR', message: msg };
      return;
    }
    
    loading = true;
    apiError = null;
    
    try {
      const note = await createNote({ 
        title: title.trim(), 
        content: content.trim(),
        type: type,
        metadata: {}
      });
      
      onSuccess?.(note);
      close();
    } catch (err: any) {
      apiError = err?.response?.data || { code: 'API_ERROR', message: 'Failed to create note' };
    } finally {
      loading = false;
    }
  }
  
  function close() {
    open = false;
    title = '';
    content = '';
    type = 'star';
    apiError = null;
  }
</script>

<Modal bind:open title={modalTitle} onClose={close}>
  <form onsubmit={handleSubmit}>
    <div class="form-group">
      <label for="note-title">{titleLabel}</label>
      <input
        id="note-title"
        name="title"
        type="text"
        bind:value={title}
        placeholder={titlePlaceholder}
        disabled={loading}
        data-testid="create-note-title"
      />
    </div>
    
    <div class="form-group">
      <label for="note-type">{typeLabel}</label>
      <TypeSelector id="note-type" bind:selected={type} />
    </div>
    
    <div class="form-group">
      <label for="note-content">{contentLabel}</label>
      <textarea
        id="note-content"
        name="content"
        bind:value={content}
        placeholder={contentPlaceholder}
        rows={6}
        disabled={loading}
        data-testid="create-note-content"
      ></textarea>
    </div>
    
    <ApiErrorDisplay error={apiError} onClose={() => apiError = null} />
    
    <div class="form-actions">
      <Button variant="secondary" onClick={close} disabled={loading}>
        {cancelText}
      </Button>
      <Button variant="primary" type="submit" disabled={loading}>
        {loading ? creatingText : createText}
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

  .form-actions {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }
</style>
