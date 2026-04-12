<script lang="ts">
  import { computePosition, flip, shift, offset } from '@floating-ui/dom';
  import { fade } from 'svelte/transition';
  import ConfirmDialog from './ui/ConfirmDialog.svelte';
  
  interface Note {
    id: string;
    title: string;
    content: string;
    type?: string;
  }
  
  interface Props {
    note: Note | null;
    isOpen: boolean;
    position: { x: number; y: number };
    onClose: () => void;
    onEdit?: (id: string) => void;
    onCreateLink?: (id: string) => void;
    onCopyLink?: (id: string) => void;
    onDelete?: (id: string) => void;
    onChangeType?: (id: string, type: string) => void;
  }
  
  let { 
    note, 
    isOpen, 
    position, 
    onClose, 
    onEdit, 
    onCreateLink, 
    onCopyLink, 
    onDelete, 
    onChangeType 
  }: Props = $props();
  
  let menuRef = $state<HTMLDivElement | null>(null);
  let menuPosition = $state({ x: 0, y: 0 });
  let showDeleteConfirm = $state(false);
  
  // Update initial position when opened
  $effect(() => {
    if (isOpen) {
      menuPosition = { x: position.x, y: position.y };
    }
  });
  
  // Update position using floating-ui
  $effect(() => {
    if (!isOpen || !menuRef) return;
    
    const virtualElement = {
      getBoundingClientRect() {
        return {
          width: 0,
          height: 0,
          x: position.x,
          y: position.y,
          top: position.y,
          left: position.x,
          right: position.x,
          bottom: position.y,
        };
      },
    };
    
    computePosition(virtualElement, menuRef, {
      placement: 'bottom-start',
      middleware: [offset(8), flip(), shift({ padding: 8 })],
    }).then(({ x, y }) => {
      menuPosition = { x, y };
    });
  });
  
  // Close on click outside
  $effect(() => {
    if (!isOpen) return;
    
    const handleClick = (e: MouseEvent) => {
      if (menuRef && !menuRef.contains(e.target as Node)) {
        onClose();
      }
    };
    
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  });
  
  // Close on Escape
  $effect(() => {
    if (!isOpen) return;
    
    const handleKeydown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    
    window.addEventListener('keydown', handleKeydown);
    return () => window.removeEventListener('keydown', handleKeydown);
  });
  
  function handleEdit() {
    if (note) onEdit?.(note.id);
    onClose();
  }
  
  function handleCreateLink() {
    if (note) onCreateLink?.(note.id);
    onClose();
  }
  
  function handleCopyLink() {
    if (note) {
      const url = `${window.location.origin}/notes/${note.id}`;
      navigator.clipboard.writeText(url);
      onCopyLink?.(note.id);
    }
    onClose();
  }
  
  function handleDeleteClick() {
    showDeleteConfirm = true;
  }
  
  function confirmDelete() {
    if (note) onDelete?.(note.id);
    showDeleteConfirm = false;
    onClose();
  }
  
  function cancelDelete() {
    showDeleteConfirm = false;
  }
  
  function handleChangeType(type: string) {
    if (note) onChangeType?.(note.id, type);
    onClose();
  }
  
  const nodeTypes = [
    { type: 'star', label: 'Звезда', color: '#ffaa00' },
    { type: 'planet', label: 'Планета', color: '#44aaff' },
    { type: 'comet', label: 'Комета', color: '#aaffdd' },
    { type: 'galaxy', label: 'Галактика', color: '#aa44ff' },
  ];
</script>

{#if isOpen && note}
  <div
    bind:this={menuRef}
    class="context-menu"
    style="left: {menuPosition.x}px; top: {menuPosition.y}px;"
    transition:fade={{ duration: 100 }}
    role="menu"
  >
    <div class="menu-header">
      <span class="note-title">{note.title}</span>
    </div>
    
    <div class="menu-items">
      <button class="menu-item" onclick={handleEdit} role="menuitem">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/>
          <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/>
        </svg>
        <span>Редактировать</span>
      </button>
      
      <button class="menu-item" onclick={handleCreateLink} role="menuitem">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/>
          <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/>
        </svg>
        <span>Создать связь</span>
      </button>
      
      <button class="menu-item" onclick={handleCopyLink} role="menuitem">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="9" y="9" width="13" height="13" rx="2" ry="2"/>
          <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
        </svg>
        <span>Копировать ссылку</span>
      </button>
      
      <div class="menu-divider"></div>
      
      <div class="menu-section">
        <span class="section-label">Тип узла</span>
        <div class="type-options">
          {#each nodeTypes as type}
            <button
              class="type-option"
              class:active={note.type === type.type}
              onclick={() => handleChangeType(type.type)}
              title={type.label}
            >
              <span class="type-dot" style="background-color: {type.color}"></span>
            </button>
          {/each}
        </div>
      </div>
      
      <div class="menu-divider"></div>
      
      <button class="menu-item danger" onclick={handleDeleteClick} role="menuitem">
        <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="3,6 5,6 21,6"/>
          <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/>
        </svg>
        <span>Удалить</span>
      </button>
    </div>
  </div>
{/if}

<ConfirmDialog
  open={showDeleteConfirm}
  title="Удалить заметку?"
  message="Вы уверены, что хотите удалить «{note?.title || ''}»? Это действие нельзя отменить."
  confirmText="Удалить"
  cancelText="Отмена"
  onConfirm={confirmDelete}
  onCancel={cancelDelete}
/>

<style>
  .context-menu {
    position: fixed;
    min-width: 14rem;
    background: rgba(0, 0, 0, 0.9);
    backdrop-filter: blur(16px);
    border: 1px solid rgba(255, 255, 255, 0.15);
    border-radius: 0.75rem;
    padding: 0.5rem 0;
    z-index: 2000;
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.4);
  }
  
  .menu-header {
    padding: 0.75rem 1rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    margin-bottom: 0.5rem;
  }
  
  .note-title {
    font-weight: 600;
    font-size: 0.875rem;
    color: white;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    display: block;
  }
  
  .menu-items {
    display: flex;
    flex-direction: column;
    padding: 0 0.5rem;
  }
  
  .menu-item {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.625rem 0.75rem;
    background: transparent;
    border: none;
    border-radius: 0.5rem;
    color: rgba(255, 255, 255, 0.85);
    font-size: 0.875rem;
    cursor: pointer;
    transition: all 0.15s ease;
    text-align: left;
  }
  
  .menu-item:hover {
    background: rgba(255, 255, 255, 0.1);
    color: white;
  }
  
  .menu-item.danger {
    color: #ef4444;
  }
  
  .menu-item.danger:hover {
    background: rgba(239, 68, 68, 0.2);
  }
  
  .menu-item svg {
    flex-shrink: 0;
    opacity: 0.7;
  }
  
  .menu-divider {
    height: 1px;
    background: rgba(255, 255, 255, 0.1);
    margin: 0.5rem 0.75rem;
  }
  
  .menu-section {
    padding: 0.5rem 0.75rem;
  }
  
  .section-label {
    display: block;
    font-size: 0.6875rem;
    color: rgba(255, 255, 255, 0.5);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: 0.5rem;
  }
  
  .type-options {
    display: flex;
    gap: 0.5rem;
  }
  
  .type-option {
    width: 1.75rem;
    height: 1.75rem;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 0.5rem;
    cursor: pointer;
    transition: all 0.15s ease;
    padding: 0;
  }
  
  .type-option:hover {
    background: rgba(255, 255, 255, 0.15);
    transform: scale(1.1);
  }
  
  .type-option.active {
    border-color: rgba(255, 255, 255, 0.5);
    background: rgba(255, 255, 255, 0.2);
  }
  
  .type-dot {
    width: 0.75rem;
    height: 0.75rem;
    border-radius: 50%;
  }
</style>
