<script lang="ts">
  import { browser } from '$app/environment';

  interface Props {
    open: boolean;
    title?: string;
    message: string;
    confirmText?: string;
    cancelText?: string;
    danger?: boolean;
    onConfirm: () => void;
    onCancel: () => void;
  }

  let {
    open = $bindable(false),
    title = 'Confirm',
    message,
    confirmText = 'Confirm',
    cancelText = 'Cancel',
    danger = false,
    onConfirm,
    onCancel
  }: Props = $props();

  function handleConfirm() {
    if (!browser) return;
    onConfirm();
  }

  function handleCancel() {
    onCancel();
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      handleCancel();
    } else if (e.key === 'Enter') {
      handleConfirm();
    }
  }
</script>

{#if open}
  <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
  <div
    class="modal-overlay"
    onclick={handleCancel}
    onkeydown={handleKeydown}
    role="dialog"
    aria-modal="true"
    aria-labelledby="confirm-title"
    tabindex="-1"
  >
    <!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
    <div class="modal-content" onclick={(e) => e.stopPropagation()}>
      <h2 id="confirm-title" class="modal-title">{title}</h2>
      <p class="modal-message">{message}</p>
      <div class="modal-actions">
        <button class="cancel-btn" onclick={handleCancel}>
          {cancelText}
        </button>
        <button
          class="confirm-btn"
          class:danger
          onclick={handleConfirm}
        >
          {confirmText}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .modal-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 1rem;
  }

  .modal-content {
    background: white;
    border-radius: 12px;
    padding: 24px;
    max-width: 400px;
    width: 100%;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
    animation: modalIn 0.2s ease-out;
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

  .modal-title {
    margin: 0 0 12px 0;
    font-size: 1.25rem;
    font-weight: 600;
    color: #1e293b;
  }

  .modal-message {
    margin: 0 0 24px 0;
    font-size: 1rem;
    color: #64748b;
    line-height: 1.5;
  }

  .modal-actions {
    display: flex;
    gap: 12px;
    justify-content: flex-end;
  }

  button {
    padding: 10px 20px;
    border-radius: 8px;
    font-size: 0.95rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
    border: none;
  }

  .cancel-btn {
    background: #f1f5f9;
    color: #64748b;
  }

  .cancel-btn:hover {
    background: #e2e8f0;
  }

  .confirm-btn {
    background: #3b82f6;
    color: white;
  }

  .confirm-btn:hover {
    background: #2563eb;
  }

  .confirm-btn.danger {
    background: #ef4444;
  }

  .confirm-btn.danger:hover {
    background: #dc2626;
  }

  @media (max-width: 480px) {
    .modal-content {
      padding: 20px;
    }

    .modal-actions {
      flex-direction: column-reverse;
    }

    button {
      width: 100%;
    }
  }
</style>
