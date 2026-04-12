<script lang="ts">
  import { fade, scale } from 'svelte/transition';
  
  const { 
    isOpen = false,
    title = 'Подтверждение',
    message = 'Вы уверены?',
    confirmText = 'Да',
    cancelText = 'Отмена',
    onConfirm,
    onCancel
  }: {
    isOpen: boolean;
    title?: string;
    message?: string;
    confirmText?: string;
    cancelText?: string;
    onConfirm: () => void;
    onCancel: () => void;
  } = $props();
  
  function handleConfirm() {
    onConfirm();
  }
  
  function handleCancel() {
    onCancel();
  }
  
  function handleBackdropClick(event: MouseEvent) {
    if (event.target === event.currentTarget) {
      handleCancel();
    }
  }
</script>

{#if isOpen}
  <div 
    class="modal-backdrop" 
    onclick={handleBackdropClick}
    transition:fade={{ duration: 200 }}
  >
    <div class="modal-content" transition:scale={{ duration: 200, start: 0.9 }}>
      <h3 class="modal-title">{title}</h3>
      <p class="modal-message">{message}</p>
      <div class="modal-actions">
        <button class="btn-cancel" onclick={handleCancel}>
          {cancelText}
        </button>
        <button class="btn-confirm" onclick={handleConfirm}>
          {confirmText}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 2000;
  }

  .modal-content {
    background: linear-gradient(145deg, #0a1a3a, #1a2a4a);
    border: 1px solid rgba(255, 221, 136, 0.3);
    border-radius: 12px;
    padding: 24px;
    max-width: 400px;
    width: 90%;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4);
  }

  .modal-title {
    margin: 0 0 12px 0;
    color: #ffdd88;
    font-size: 1.25rem;
    font-weight: 600;
  }

  .modal-message {
    margin: 0 0 24px 0;
    color: #e0e0e0;
    font-size: 1rem;
    line-height: 1.5;
  }

  .modal-actions {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
  }

  .btn-cancel, .btn-confirm {
    padding: 10px 20px;
    border-radius: 6px;
    border: none;
    cursor: pointer;
    font-size: 0.95rem;
    transition: all 0.2s;
  }

  .btn-cancel {
    background: rgba(136, 170, 204, 0.2);
    color: #88aacc;
  }

  .btn-cancel:hover {
    background: rgba(136, 170, 204, 0.3);
  }

  .btn-confirm {
    background: linear-gradient(135deg, #ff6b6b 0%, #ff4444 100%);
    color: white;
  }

  .btn-confirm:hover {
    background: linear-gradient(135deg, #ff7777 0%, #ff5555 100%);
    transform: translateY(-1px);
  }
</style>
