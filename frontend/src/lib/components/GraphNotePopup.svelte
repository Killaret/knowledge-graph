<script lang="ts">
  import { fade, scale } from 'svelte/transition';
  import { goto } from '$app/navigation';
  import { deleteNote } from '$lib/api/notes';
  import ConfirmModal from './ConfirmModal.svelte';
  
  interface Note {
    id: string;
    title: string;
    content: string;
    created_at?: string;
  }
  
  let { 
    note,
    isOpen = false,
    position = { x: 0, y: 0 },
    onClose,
    onDelete
  }: {
    note: Note | null;
    isOpen: boolean;
    position?: { x: number; y: number };
    onClose: () => void;
    onDelete: (id: string) => void;
  } = $props();
  
  let showConfirmModal = $state(false);
  let isDeleting = $state(false);
  
  function handleEdit() {
    if (note) {
      goto(`/notes/${note.id}/edit`);
    }
  }
  
  function handleOpen() {
    if (note) {
      goto(`/notes/${note.id}`);
    }
  }
  
  function handleDeleteClick() {
    showConfirmModal = true;
  }
  
  async function handleConfirmDelete() {
    if (!note) return;
    
    isDeleting = true;
    try {
      await deleteNote(note.id);
      onDelete(note.id);
      onClose();
    } catch (e) {
      console.error('Failed to delete note:', e);
    } finally {
      isDeleting = false;
      showConfirmModal = false;
    }
  }
  
  function handleCancelDelete() {
    showConfirmModal = false;
  }
  
  function handleBackdropClick(event: MouseEvent) {
    if (event.target === event.currentTarget) {
      onClose();
    }
  }
  
  function truncateText(text: string, maxLength: number): string {
    if (text.length <= maxLength) return text;
    return text.slice(0, maxLength) + '...';
  }
</script>

{#if isOpen && note}
  <div 
    class="popup-backdrop" 
    onclick={handleBackdropClick}
    onkeydown={(e) => e.key === 'Escape' && onClose()}
    role="dialog"
    aria-modal="true"
    aria-labelledby="popup-title"
    tabindex="-1"
    transition:fade={{ duration: 200 }}
  >
    <div 
      class="popup-content" 
      style="left: {Math.min(position.x + 20, window.innerWidth - 320)}px; top: {Math.min(position.y, window.innerHeight - 300)}px;"
      transition:scale={{ duration: 200, start: 0.9 }}
    >
      <button class="close-btn" onclick={onClose} aria-label="Закрыть">
        <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="18" y1="6" x2="6" y2="18"></line>
          <line x1="6" y1="6" x2="18" y2="18"></line>
        </svg>
      </button>
      
      <h3 id="popup-title" class="popup-title">{note.title}</h3>
      
      <div class="popup-content-text">
        {truncateText(note.content, 200)}
      </div>
      
      <div class="popup-actions">
        <button class="btn-open" onclick={handleOpen}>
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"></path>
            <polyline points="15,3 21,3 21,9"></polyline>
            <line x1="10" y1="14" x2="21" y2="3"></line>
          </svg>
          Открыть
        </button>
        <button class="btn-edit" onclick={handleEdit}>
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
            <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
          </svg>
          Редактировать
        </button>
        <button class="btn-delete" onclick={handleDeleteClick} disabled={isDeleting}>
          {#if isDeleting}
            <span class="spinner"></span>
          {:else}
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="3,6 5,6 21,6"></polyline>
              <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
            </svg>
          {/if}
          Удалить
        </button>
      </div>
    </div>
  </div>
{/if}

<ConfirmModal
  isOpen={showConfirmModal}
  title="Удалить заметку?"
  message="Вы уверены, что хотите удалить «{note?.title || ''}»? Это действие нельзя отменить."
  confirmText="Удалить"
  cancelText="Отмена"
  onConfirm={handleConfirmDelete}
  onCancel={handleCancelDelete}
/>

<style>
  .popup-backdrop {
    position: fixed;
    inset: 0;
    z-index: 1500;
    pointer-events: none;
  }

  .popup-content {
    position: absolute;
    width: 300px;
    background: linear-gradient(145deg, #0a1a3a, #1a2a4a);
    border: 1px solid rgba(255, 221, 136, 0.3);
    border-radius: 12px;
    padding: 20px;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
    pointer-events: auto;
  }

  .close-btn {
    position: absolute;
    top: 12px;
    right: 12px;
    background: none;
    border: none;
    color: #88aacc;
    cursor: pointer;
    padding: 4px;
    border-radius: 4px;
    transition: all 0.2s;
  }

  .close-btn:hover {
    color: #ffdd88;
    background: rgba(255, 221, 136, 0.1);
  }

  .popup-title {
    margin: 0 0 12px 0;
    color: #ffdd88;
    font-size: 1.2rem;
    font-weight: 600;
    padding-right: 30px;
  }

  .popup-content-text {
    color: #e0e0e0;
    font-size: 0.95rem;
    line-height: 1.5;
    margin-bottom: 16px;
    max-height: 150px;
    overflow-y: auto;
  }

  .popup-actions {
    display: flex;
    gap: 8px;
    flex-wrap: wrap;
  }

  .popup-actions button {
    flex: 1;
    min-width: 80px;
    padding: 8px 12px;
    border-radius: 6px;
    border: none;
    cursor: pointer;
    font-size: 0.85rem;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    transition: all 0.2s;
  }

  .btn-open {
    background: rgba(68, 170, 255, 0.2);
    color: #44aaff;
  }

  .btn-open:hover {
    background: rgba(68, 170, 255, 0.3);
  }

  .btn-edit {
    background: rgba(255, 221, 136, 0.2);
    color: #ffdd88;
  }

  .btn-edit:hover {
    background: rgba(255, 221, 136, 0.3);
  }

  .btn-delete {
    background: rgba(255, 107, 107, 0.2);
    color: #ff6b6b;
  }

  .btn-delete:hover:not(:disabled) {
    background: rgba(255, 107, 107, 0.3);
  }

  .btn-delete:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .spinner {
    width: 14px;
    height: 14px;
    border: 2px solid rgba(255, 107, 107, 0.3);
    border-top-color: #ff6b6b;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }
</style>
