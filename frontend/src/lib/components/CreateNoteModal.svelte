<script lang="ts">
  import { createNote, type Note } from '$lib/api/notes';
  
  let { 
    open = $bindable(false),
    onSuccess 
  }: { 
    open: boolean;
    onSuccess?: (note: Note) => void;
  } = $props();
  
  let title = $state('');
  let content = $state('');
  let type = $state<'star' | 'planet' | 'comet' | 'galaxy' | 'default'>('default');
  let loading = $state(false);
  let error = $state('');
  
  const types = [
    { value: 'star', label: '⭐ Star', color: '#fbbf24' },
    { value: 'planet', label: '🪐 Planet', color: '#60a5fa' },
    { value: 'comet', label: '☄️ Comet', color: '#f472b6' },
    { value: 'galaxy', label: '🌀 Galaxy', color: '#a78bfa' },
    { value: 'default', label: '📄 Default', color: '#94a3b8' }
  ];
  
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
    } catch (e) {
      error = 'Failed to create note';
    } finally {
      loading = false;
    }
  }
  
  function close() {
    open = false;
    title = '';
    content = '';
    type = 'default';
    error = '';
  }
</script>

{#if open}
  <div class="modal-overlay" onclick={close}>
    <div class="modal" onclick={(e) => e.stopPropagation()}>
      <div class="modal-header">
        <h2>Create New Note</h2>
        <button class="close-btn" onclick={close} aria-label="Close">
          <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>
      
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
          <label for="note-type">Type</label>
          <div class="type-selector">
            {#each types as t}
              <button
                type="button"
                class="type-btn {type === t.value ? 'active' : ''}"
                onclick={() => type = t.value}
                style="--type-color: {t.color}"
              >
                {t.label}
              </button>
            {/each}
          </div>
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
          <button type="button" class="btn-secondary" onclick={close} disabled={loading}>
            Cancel
          </button>
          <button type="submit" class="btn-primary" disabled={loading}>
            {loading ? 'Creating...' : 'Create Note'}
          </button>
        </div>
      </form>
    </div>
  </div>
{/if}

<style>
  .modal-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(4px);
    z-index: 1000;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
  }

  .modal {
    background: white;
    border-radius: 16px;
    width: 100%;
    max-width: 500px;
    max-height: 90vh;
    overflow: hidden;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
    animation: modalIn 0.2s ease;
  }

  @keyframes modalIn {
    from {
      opacity: 0;
      transform: scale(0.95);
    }
    to {
      opacity: 1;
      transform: scale(1);
    }
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px 24px;
    border-bottom: 1px solid #e2e8f0;
  }

  .modal-header h2 {
    font-size: 18px;
    font-weight: 600;
    color: #1e293b;
    margin: 0;
  }

  .close-btn {
    padding: 8px;
    border: none;
    background: transparent;
    border-radius: 8px;
    cursor: pointer;
    color: #64748b;
    transition: all 0.2s;
  }

  .close-btn:hover {
    background: #f1f5f9;
    color: #334155;
  }

  form {
    padding: 24px;
  }

  .form-group {
    margin-bottom: 20px;
  }

  label {
    display: block;
    font-size: 14px;
    font-weight: 500;
    color: #374151;
    margin-bottom: 6px;
  }

  input, textarea {
    width: 100%;
    padding: 10px 14px;
    border: 1px solid #d1d5db;
    border-radius: 8px;
    font-size: 15px;
    color: #1f2937;
    background: white;
    transition: border-color 0.2s;
  }

  input:focus, textarea:focus {
    outline: none;
    border-color: #3b82f6;
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }

  textarea {
    resize: vertical;
    font-family: inherit;
  }

  .type-selector {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .type-btn {
    padding: 8px 14px;
    border: 1px solid #e5e7eb;
    background: white;
    border-radius: 20px;
    cursor: pointer;
    font-size: 13px;
    transition: all 0.2s;
  }

  .type-btn:hover {
    border-color: var(--type-color);
    background: rgba(0, 0, 0, 0.02);
  }

  .type-btn.active {
    border-color: var(--type-color);
    background: var(--type-color);
    color: white;
  }

  .error {
    padding: 12px;
    background: #fee2e2;
    color: #dc2626;
    border-radius: 8px;
    font-size: 14px;
    margin-bottom: 20px;
  }

  .form-actions {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }

  .btn-secondary {
    padding: 10px 20px;
    border: 1px solid #e5e7eb;
    background: white;
    border-radius: 8px;
    cursor: pointer;
    font-size: 14px;
    font-weight: 500;
    color: #374151;
    transition: all 0.2s;
  }

  .btn-secondary:hover {
    background: #f9fafb;
  }

  .btn-primary {
    padding: 10px 20px;
    border: none;
    background: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
    border-radius: 8px;
    cursor: pointer;
    font-size: 14px;
    font-weight: 500;
    color: white;
    transition: all 0.2s;
  }

  .btn-primary:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
  }

  button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
</style>
