<script lang="ts">
  /* eslint-disable prefer-const -- Svelte 5 bindable props require let, see: https://svelte.dev/docs/svelte/$bindable */
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';

  interface Props {
    open: boolean;
    title: string;
    onClose?: () => void;
    children?: () => any;
  }

  let {
    open = $bindable(false),
    title,
    onClose,
    children
  }: Props = $props();

  let modalRef = $state<HTMLDivElement | null>(null);

  // Отключаем анимации в тестовом окружении (где нет Web Animations API)
  const hasAnimations = browser && typeof Element !== 'undefined' && 
    (Element.prototype.animate || typeof document !== 'undefined');

  function handleClose() {
    open = false;
    onClose?.();
  }

  function handleOverlayClick(e: MouseEvent) {
    if (e.target === e.currentTarget) {
      handleClose();
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape' && open) {
      handleClose();
    }
  }

  onMount(() => {
    document.addEventListener('keydown', handleKeydown);
    return () => {
      document.removeEventListener('keydown', handleKeydown);
    };
  });
</script>

{#if open}
  <div
    class="modal-overlay"
    onclick={handleOverlayClick}
    class:no-transition={!hasAnimations}
    role="presentation"
  >
    <div
      class="modal-container"
      role="dialog"
      aria-modal="true"
      aria-labelledby="modal-title"
      bind:this={modalRef}
      class:no-transition={!hasAnimations}
    >
      <div class="modal-header">
        <h2 id="modal-title">{title}</h2>
        <button
          class="close-btn"
          onclick={handleClose}
          aria-label="Close"
          type="button"
        >
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="20"
            height="20"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </svg>
        </button>
      </div>
      <div class="modal-content">
        {@render children?.()}
      </div>
    </div>
  </div>
{/if}

<style>
  /* Отключаем анимации в тестовом окружении */
  :global(.no-transition) {
    animation: none !important;
    transition: none !important;
  }

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

  .modal-container {
    background: var(--color-surface, white);
    border-radius: 12px;
    width: 100%;
    max-width: 500px;
    max-height: 90vh;
    overflow: hidden;
    box-shadow: var(--shadow-xl);
    display: flex;
    flex-direction: column;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px 24px;
    border-bottom: 1px solid var(--color-border);
    flex-shrink: 0;
  }

  .modal-header h2 {
    font-size: 18px;
    font-weight: 600;
    color: var(--color-text);
    margin: 0;
  }

  .close-btn {
    padding: 8px;
    border: none;
    background: transparent;
    border-radius: 8px;
    cursor: pointer;
    color: var(--color-text-secondary);
    transition: all 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .close-btn:hover {
    background: var(--color-surface-elevated);
    color: var(--color-text);
  }

  .close-btn:focus {
    outline: none;
    box-shadow: 0 0 0 2px var(--color-border-focus);
  }

  .modal-content {
    padding: 24px;
    overflow-y: auto;
    flex: 1;
  }
</style>
