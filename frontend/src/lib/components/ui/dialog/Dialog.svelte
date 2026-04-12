<script lang="ts" module>
  export interface Props {
    open?: boolean;
    className?: string;
    onClose?: () => void;
    children?: import('svelte').Snippet;
  }
</script>

<script lang="ts">
  import { fade, fly } from 'svelte/transition';
  import { cn } from '$lib/utils.js';

  const props = $props();
  let open = props.open ?? false;
  const { className, onClose, children } = props;

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      open = false;
      onClose?.();
    }
  }

  function handleOverlayClick() {
    open = false;
    onClose?.();
  }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if open}
  <div class="dialog-overlay" role="button" tabindex="0" onclick={handleOverlayClick} onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && handleOverlayClick()} transition:fade={{ duration: 150 }}>
    <div
      class={cn(
        "dialog-content bg-white rounded-lg shadow-lg border p-6 max-w-lg w-full mx-4",
        className
      )}
      onclick={(e) => e.stopPropagation()}
      transition:fly={{ y: 10, duration: 200 }}
      role="dialog"
      aria-modal="true"
      tabindex="-1"
    >
      {@render children?.()}
    </div>
  </div>
{/if}

<style>
  .dialog-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(4px);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1500;
  }

  .dialog-content {
    position: relative;
    max-height: 90vh;
    overflow-y: auto;
  }
</style>
