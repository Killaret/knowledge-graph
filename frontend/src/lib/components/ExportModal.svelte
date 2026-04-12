<script lang="ts">
  import { fade, scale } from 'svelte/transition';
  
  const { 
    isOpen = false,
    notes = [],
    onClose,
    onExport
  }: {
    isOpen: boolean;
    notes: any[];
    onClose: () => void;
    onExport: (format: 'json' | 'csv' | 'md', selectedIds: string[]) => void;
  } = $props();
  
  let selectedFormat: 'json' | 'csv' | 'md' = $state('json');
  let selectedIds: string[] = $state([]);
  let selectAll = $state(true);
  
  $effect(() => {
    if (selectAll) {
      selectedIds = notes.map(n => n.id);
    } else {
      selectedIds = [];
    }
  });
  
  function toggleSelectAll() {
    selectAll = !selectAll;
  }
  
  function toggleNote(id: string) {
    if (selectedIds.includes(id)) {
      selectedIds = selectedIds.filter(i => i !== id);
      selectAll = false;
    } else {
      selectedIds = [...selectedIds, id];
      if (selectedIds.length === notes.length) {
        selectAll = true;
      }
    }
  }
  
  function handleExport() {
    onExport(selectedFormat, selectedIds);
    onClose();
  }
  
  function handleBackdropClick(event: MouseEvent) {
    if (event.target === event.currentTarget) {
      onClose();
    }
  }
</script>

{#if isOpen}
  <div 
    class="modal-backdrop" 
    onclick={handleBackdropClick}
    onkeydown={(e) => e.key === 'Escape' && onClose()}
    role="dialog"
    aria-modal="true"
    tabindex="-1"
    transition:fade={{ duration: 200 }}
  >
    <div class="modal-content" transition:scale={{ duration: 200, start: 0.95 }}>
      <button class="close-btn" onclick={onClose} aria-label="Закрыть">
        <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      </button>
      
      <h2 class="modal-title">Экспорт заметок</h2>
      
      <div class="format-section">
        <span class="section-label">Формат:</span>
        <div class="format-options">
          <button 
            class="format-btn"
            class:active={selectedFormat === 'json'}
            onclick={() => selectedFormat = 'json'}
          >
            <span class="format-icon">📄</span>
            JSON
          </button>
          <button 
            class="format-btn"
            class:active={selectedFormat === 'csv'}
            onclick={() => selectedFormat = 'csv'}
          >
            <span class="format-icon">📊</span>
            CSV
          </button>
          <button 
            class="format-btn"
            class:active={selectedFormat === 'md'}
            onclick={() => selectedFormat = 'md'}
          >
            <span class="format-icon">📝</span>
            Markdown
          </button>
        </div>
      </div>
      
      <div class="notes-section">
        <div class="select-all-row">
          <label class="checkbox-label">
            <input 
              type="checkbox" 
              checked={selectAll}
              onchange={toggleSelectAll}
            />
            <span>Выбрать все ({notes.length})</span>
          </label>
        </div>
        
        <div class="notes-list">
          {#each notes as note}
            <label class="note-item">
              <input 
                type="checkbox" 
                checked={selectedIds.includes(note.id)}
                onchange={() => toggleNote(note.id)}
              />
              <span class="note-title">{note.title}</span>
            </label>
          {/each}
        </div>
      </div>
      
      <div class="modal-actions">
        <button class="btn-cancel" onclick={onClose}>Отмена</button>
        <button 
          class="btn-export" 
          onclick={handleExport}
          disabled={selectedIds.length === 0}
        >
          Экспортировать {selectedIds.length > 0 ? `(${selectedIds.length})` : ''}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(8px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 2000;
  }

  .modal-content {
    position: relative;
    width: 90%;
    max-width: 500px;
    max-height: 80vh;
    background: linear-gradient(145deg, #0a1a3a, #1a2a4a);
    border: 1px solid rgba(255, 221, 136, 0.2);
    border-radius: 16px;
    padding: 32px;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
    overflow-y: auto;
  }

  .close-btn {
    position: absolute;
    top: 16px;
    right: 16px;
    background: none;
    border: none;
    color: #88aacc;
    cursor: pointer;
    padding: 8px;
    border-radius: 8px;
    transition: all 0.2s;
  }

  .close-btn:hover {
    color: #ffdd88;
    background: rgba(255, 221, 136, 0.1);
  }

  .modal-title {
    margin: 0 0 24px 0;
    color: #ffdd88;
    font-size: 1.5rem;
    font-weight: 600;
  }

  .section-label {
    display: block;
    color: #88aacc;
    font-size: 0.9rem;
    margin-bottom: 12px;
  }

  .format-options {
    display: flex;
    gap: 12px;
    margin-bottom: 24px;
  }

  .format-btn {
    flex: 1;
    padding: 16px;
    background: rgba(10, 26, 58, 0.6);
    border: 2px solid rgba(136, 170, 204, 0.2);
    border-radius: 12px;
    color: #e0e0e0;
    cursor: pointer;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    transition: all 0.2s;
  }

  .format-btn:hover {
    border-color: rgba(255, 221, 136, 0.4);
    background: rgba(255, 221, 136, 0.05);
  }

  .format-btn.active {
    border-color: #ffdd88;
    background: rgba(255, 221, 136, 0.15);
    color: #ffdd88;
  }

  .format-icon {
    font-size: 1.5rem;
  }

  .notes-section {
    margin-bottom: 24px;
  }

  .select-all-row {
    padding: 12px 0;
    border-bottom: 1px solid rgba(136, 170, 204, 0.2);
    margin-bottom: 8px;
  }

  .checkbox-label {
    display: flex;
    align-items: center;
    gap: 8px;
    color: #e0e0e0;
    cursor: pointer;
  }

  .checkbox-label input {
    width: 18px;
    height: 18px;
    accent-color: #ffdd88;
  }

  .notes-list {
    max-height: 200px;
    overflow-y: auto;
  }

  .note-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 0;
    border-bottom: 1px solid rgba(136, 170, 204, 0.1);
    cursor: pointer;
    color: #e0e0e0;
  }

  .note-item:last-child {
    border-bottom: none;
  }

  .note-item input {
    width: 18px;
    height: 18px;
    accent-color: #ffdd88;
  }

  .note-title {
    font-size: 0.95rem;
  }

  .modal-actions {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
  }

  .btn-cancel, .btn-export {
    padding: 12px 24px;
    border-radius: 8px;
    border: none;
    cursor: pointer;
    font-size: 0.95rem;
    font-weight: 600;
    transition: all 0.2s;
  }

  .btn-cancel {
    background: rgba(136, 170, 204, 0.2);
    color: #88aacc;
  }

  .btn-cancel:hover {
    background: rgba(136, 170, 204, 0.3);
  }

  .btn-export {
    background: linear-gradient(135deg, #ffdd88 0%, #ffaa44 100%);
    color: #0a1a3a;
  }

  .btn-export:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(255, 170, 68, 0.4);
  }

  .btn-export:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
