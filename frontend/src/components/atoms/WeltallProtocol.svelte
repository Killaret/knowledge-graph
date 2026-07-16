<script lang="ts">
  import { onMount } from 'svelte';
  import { fade, fly } from 'svelte/transition';
  import { browser } from '$app/environment';

  interface Props {
    show?: boolean;
  }

  const { show = false }: Props = $props();

  let modalRef = $state<HTMLDivElement | null>(null);
  let closeButtonRef = $state<HTMLButtonElement | null>(null);

  // Handle escape key
  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape' && show) {
      dispatchClose();
    }
  }

  // Focus trap
  function handleFocusTrap(e: KeyboardEvent) {
    if (e.key !== 'Tab' || !modalRef) return;

    const focusableElements = modalRef.querySelectorAll(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    );
    const firstElement = focusableElements[0] as HTMLElement;
    const lastElement = focusableElements[focusableElements.length - 1] as HTMLElement;

    if (e.shiftKey && document.activeElement === firstElement) {
      e.preventDefault();
      lastElement?.focus();
    } else if (!e.shiftKey && document.activeElement === lastElement) {
      e.preventDefault();
      firstElement?.focus();
    }
  }

  // Custom event dispatcher
  function dispatchClose() {
    const event = new CustomEvent('close', { bubbles: true });
    modalRef?.dispatchEvent(event);
  }

  // Focus close button when shown
  $effect(() => {
    if (show && closeButtonRef) {
      setTimeout(() => closeButtonRef?.focus(), 100);
    }
  });

  onMount(() => {
    if (browser) {
      document.addEventListener('keydown', handleKeydown);
      return () => {
        document.removeEventListener('keydown', handleKeydown);
      };
    }
  });
</script>

{#if show}
  <div
    class="protocol-overlay"
    onclick={(e) => { if (e.target === e.currentTarget) dispatchClose(); }}
    role="presentation"
    transition:fade={{ duration: 300 }}
  >
    <div
      class="protocol-modal"
      bind:this={modalRef}
      role="dialog"
      tabindex="-1"
      aria-modal="true"
      aria-labelledby="protocol-title"
      transition:fly={{ y: 20, duration: 400, delay: 100 }}
      onkeydown={handleFocusTrap}
    >
      <!-- Corner decorations -->
      <div class="corner top-left"></div>
      <div class="corner top-right"></div>
      <div class="corner bottom-left"></div>
      <div class="corner bottom-right"></div>

      <!-- Content -->
      <div class="protocol-content">
        <h2 id="protocol-title" class="protocol-title">
          <span class="protocol-prefix">WELTALL</span>
          <span class="protocol-suffix">PROTOCOL v4.0</span>
        </h2>

        <div class="protocol-body">
          <p class="protocol-message">WELTALL PROTOCOL will be later ...</p>

          <div class="protocol-table">
            <div class="table-header">
              <span>PARAMETER</span>
              <span>STATUS</span>
              <span>VALUE</span>
            </div>
            <div class="table-row">
              <span>System Status</span>
              <span class="status-pending">PENDING</span>
              <span>--</span>
            </div>
            <div class="table-row">
              <span>Connection</span>
              <span class="status-active">ACTIVE</span>
              <span>ONLINE</span>
            </div>
            <div class="table-row">
              <span>Data Stream</span>
              <span class="status-pending">PENDING</span>
              <span>--</span>
            </div>
          </div>
        </div>

        <button
          class="close-button"
          bind:this={closeButtonRef}
          onclick={dispatchClose}
          aria-label="Close protocol panel"
        >
          <span class="button-text">Close</span>
          <span class="button-glow"></span>
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .protocol-overlay {
    position: fixed;
    inset: 0;
    z-index: 9998;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
    background: rgba(5, 5, 10, 0.9);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
  }

  .protocol-modal {
    position: relative;
    width: 100%;
    max-width: 500px;
    background: rgba(10, 10, 26, 0.95);
    border: 1px solid rgba(255, 204, 0, 0.3);
    border-radius: 4px;
    box-shadow:
      0 0 60px rgba(255, 51, 51, 0.2),
      inset 0 0 40px rgba(255, 204, 0, 0.05);
  }

  /* Corner decorations */
  .corner {
    position: absolute;
    width: 20px;
    height: 20px;
    border: 2px solid rgba(255, 204, 0, 0.5);
  }

  .corner.top-left {
    top: -2px;
    left: -2px;
    border-right: none;
    border-bottom: none;
  }

  .corner.top-right {
    top: -2px;
    right: -2px;
    border-left: none;
    border-bottom: none;
  }

  .corner.bottom-left {
    bottom: -2px;
    left: -2px;
    border-right: none;
    border-top: none;
  }

  .corner.bottom-right {
    bottom: -2px;
    right: -2px;
    border-left: none;
    border-top: none;
  }

  .protocol-content {
    padding: 2rem;
  }

  .protocol-title {
    margin: 0 0 1.5rem;
    font-family: 'Courier New', monospace;
    font-size: 1.5rem;
    text-align: center;
    letter-spacing: 0.15em;
  }

  .protocol-prefix {
    color: #ffcc00;
    text-shadow: 0 0 10px rgba(255, 204, 0, 0.5);
  }

  .protocol-suffix {
    color: rgba(255, 255, 255, 0.6);
    font-size: 0.7em;
  }

  .protocol-body {
    margin-bottom: 1.5rem;
  }

  .protocol-message {
    color: rgba(255, 255, 255, 0.7);
    font-family: 'Courier New', monospace;
    font-size: 0.9rem;
    text-align: center;
    margin-bottom: 1.5rem;
  }

  /* Protocol table */
  .protocol-table {
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 4px;
    overflow: hidden;
    font-family: 'Courier New', monospace;
    font-size: 0.75rem;
  }

  .table-header {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    gap: 0.5rem;
    padding: 0.75rem;
    background: rgba(255, 204, 0, 0.1);
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    color: #ffcc00;
    font-weight: 600;
  }

  .table-row {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    gap: 0.5rem;
    padding: 0.75rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    color: rgba(255, 255, 255, 0.7);
  }

  .table-row:last-child {
    border-bottom: none;
  }

  .status-active {
    color: #60a5fa;
  }

  .status-pending {
    color: #fbbf24;
  }

  /* Close button */
  .close-button {
    position: relative;
    width: 100%;
    padding: 0.875rem;
    background: transparent;
    border: 1px solid rgba(255, 204, 0, 0.3);
    border-radius: 4px;
    color: #ffcc00;
    font-family: 'Courier New', monospace;
    font-size: 0.875rem;
    cursor: pointer;
    overflow: hidden;
    transition: all 0.3s ease;
  }

  .close-button:hover {
    background: rgba(255, 204, 0, 0.1);
    border-color: rgba(255, 204, 0, 0.6);
    box-shadow: 0 0 20px rgba(255, 204, 0, 0.2);
  }

  .close-button:focus {
    outline: none;
    box-shadow: 0 0 0 2px rgba(255, 204, 0, 0.3);
  }

  .button-text {
    position: relative;
    z-index: 1;
  }

  .button-glow {
    position: absolute;
    inset: 0;
    background: radial-gradient(circle at center, rgba(255, 204, 0, 0.2) 0%, transparent 70%);
    opacity: 0;
    transition: opacity 0.3s ease;
  }

  .close-button:hover .button-glow {
    opacity: 1;
  }

  /* Reduced motion support */
  @media (prefers-reduced-motion: reduce) {
    .corner {
      animation: none;
    }
  }
</style>
