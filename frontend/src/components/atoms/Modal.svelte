<script lang="ts">
  /* eslint-disable prefer-const -- Svelte 5 bindable props require let, see: https://svelte.dev/docs/svelte/$bindable */
  import { onMount, type Snippet } from "svelte";
  import { browser } from "$app/environment";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const t = (key: string) => formatMessage(key, getCurrentLocale());

  interface Props {
    open: boolean;
    title: string;
    onClose?: () => void;
    children?: Snippet;
  }

  let { open = $bindable(false), title, onClose, children }: Props = $props();

  let modalRef = $state<HTMLDivElement | null>(null);

  // Отключаем анимации в тестовом окружении (где нет Web Animations API)
  const hasAnimations =
    browser &&
    typeof Element !== "undefined" &&
    (Element.prototype.animate || typeof document !== "undefined");

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
    if (e.key === "Escape" && open) {
      handleClose();
    }
  }

  onMount(() => {
    document.addEventListener("keydown", handleKeydown);
    return () => {
      document.removeEventListener("keydown", handleKeydown);
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
        <button class="close-btn" onclick={handleClose} aria-label={t("modal.close")} type="button">
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
    background: rgba(5, 5, 8, 0.75);
    backdrop-filter: blur(14px);
    -webkit-backdrop-filter: blur(14px);
    z-index: 1000;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
    animation: fadeIn 0.3s ease;
  }

  .modal-container {
    position: relative;
    background: var(--carbon-gradient-card, linear-gradient(145deg, rgba(30, 30, 42, 0.75) 0%, rgba(18, 18, 26, 0.92) 100%));
    border-radius: 16px;
    width: 100%;
    max-width: 520px;
    max-height: 90vh;
    overflow: hidden;
    box-shadow:
      0 25px 50px -12px rgba(0, 0, 0, 0.6),
      var(--carbon-glow-primary, 0 0 30px rgba(139, 92, 246, 0.15));
    border: 1px solid var(--carbon-border, #2d2d3d);
    display: flex;
    flex-direction: column;
    animation: slideIn 0.3s ease;
  }

  .modal-container::before {
    content: "";
    position: absolute;
    inset: 0;
    border-radius: 16px;
    padding: 1px;
    background: linear-gradient(135deg, rgba(34, 211, 238, 0.25), rgba(139, 92, 246, 0.25), transparent 60%);
    -webkit-mask:
      linear-gradient(#fff 0 0) content-box,
      linear-gradient(#fff 0 0);
    mask:
      linear-gradient(#fff 0 0) content-box,
      linear-gradient(#fff 0 0);
    -webkit-mask-composite: xor;
    mask-composite: exclude;
    pointer-events: none;
  }

  .modal-header {
    position: relative;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px 24px;
    border-bottom: 1px solid var(--carbon-border, #2d2d3d);
    flex-shrink: 0;
  }

  .modal-header h2 {
    font-size: 18px;
    font-weight: 600;
    color: var(--carbon-text, #f0f0f5);
    margin: 0;
    text-shadow: 0 0 12px rgba(34, 211, 238, 0.2);
  }

  .close-btn {
    padding: 8px;
    border: 1px solid var(--carbon-border, #2d2d3d);
    background: transparent;
    border-radius: 8px;
    cursor: pointer;
    color: var(--carbon-text-muted, #8b8b9e);
    transition: all var(--carbon-transition, 0.25s ease);
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .close-btn:hover {
    background: rgba(245, 158, 11, 0.1);
    border-color: rgba(245, 158, 11, 0.4);
    color: var(--carbon-glow-amber, #f59e0b);
    box-shadow: var(--carbon-glow-amber, 0 0 10px rgba(245, 158, 11, 0.2));
  }

  .close-btn:focus-visible {
    outline: none;
    box-shadow: var(--carbon-focus-ring, 0 0 0 3px rgba(34, 211, 238, 0.15));
  }

  .modal-content {
    position: relative;
    padding: 24px;
    overflow-y: auto;
    flex: 1;
    color: var(--carbon-text, #f0f0f5);
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

  /* Disable animations for tests and for users who prefer reduced motion */
  @media (prefers-reduced-motion: reduce) {
    .modal-overlay,
    .modal-container {
      animation: none;
    }
  }

  :global(.no-transition) .modal-overlay {
    animation: none;
  }

  :global(.no-transition) .modal-container {
    animation: none;
  }
</style>
