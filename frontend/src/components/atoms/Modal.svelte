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
    background: rgba(5, 5, 10, 0.85);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    z-index: 1000;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
    animation: fadeIn 0.3s ease;
  }

  .modal-container {
    background: rgba(10, 10, 26, 0.95);
    border-radius: 12px;
    width: 100%;
    max-width: 500px;
    max-height: 90vh;
    overflow: hidden;
    box-shadow:
      0 25px 50px -12px rgba(0, 0, 0, 0.5),
      0 0 30px rgba(255, 204, 0, 0.1);
    border: 1px solid rgba(255, 255, 255, 0.1);
    display: flex;
    flex-direction: column;
    animation: slideIn 0.3s ease;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px 24px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    flex-shrink: 0;
  }

  .modal-header h2 {
    font-size: 18px;
    font-weight: 600;
    color: var(--color-text-dark, #e0e0e0);
    margin: 0;
    text-shadow: 0 0 10px rgba(255, 204, 0, 0.2);
  }

  .close-btn {
    padding: 8px;
    border: none;
    background: transparent;
    border-radius: 8px;
    cursor: pointer;
    color: rgba(255, 255, 255, 0.6);
    transition: all 0.2s ease;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .close-btn:hover {
    background: rgba(255, 255, 255, 0.1);
    color: var(--color-glow, #ffcc00);
    box-shadow: 0 0 10px rgba(255, 204, 0, 0.2);
  }

  .close-btn:focus {
    outline: none;
    box-shadow: 0 0 0 2px var(--color-glow-subtle, rgba(255, 204, 0, 0.3));
  }

  .modal-content {
    padding: 24px;
    overflow-y: auto;
    flex: 1;
    color: var(--color-text-dark, #e0e0e0);
  }

  /* Animations */
  @keyframes fadeIn {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }

  @keyframes slideIn {
    from {
      opacity: 0;
      transform: translateY(-20px) scale(0.95);
    }
    to {
      opacity: 1;
      transform: translateY(0) scale(1);
    }
  }

  /* Disable animations for tests */
  :global(.no-transition) .modal-overlay {
    animation: none;
  }

  :global(.no-transition) .modal-container {
    animation: none;
  }
</style>
