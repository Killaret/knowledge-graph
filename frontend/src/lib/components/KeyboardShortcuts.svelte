<script lang="ts">
  import { goto } from '$app/navigation';
  import { browser } from '$app/environment';
  
  let { 
    onSearchFocus
  }: {
    onSearchFocus?: () => void;
  } = $props();
  
  function handleKeydown(event: KeyboardEvent) {
    // Ignore if user is typing in input/textarea
    const target = event.target as HTMLElement;
    if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable) {
      // Allow Escape even in inputs
      if (event.key !== 'Escape') return;
    }
    
    // Ctrl/Cmd + N - Create new note
    if ((event.ctrlKey || event.metaKey) && event.key === 'n') {
      event.preventDefault();
      goto('/notes/new');
    }
    
    // Ctrl/Cmd + F - Focus search
    if ((event.ctrlKey || event.metaKey) && event.key === 'f') {
      event.preventDefault();
      if (onSearchFocus) {
        onSearchFocus();
      } else if (browser) {
        window.dispatchEvent(new CustomEvent('app:focus-search'));
      }
    }
    
    // Escape - Close modals (handled by individual components)
    if (event.key === 'Escape') {
      // Dispatch custom event for components to listen
      if (browser) {
        window.dispatchEvent(new CustomEvent('app:escape'));
      }
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

{#if browser}
  <div class="shortcuts-hint">
    <span>Ctrl+N</span> — новая заметка
    <span>Ctrl+F</span> — поиск
    <span>Esc</span> — закрыть
  </div>
{/if}

<style>
  .shortcuts-hint {
    position: fixed;
    bottom: 8px;
    left: 50%;
    transform: translateX(-50%);
    font-size: 0.75rem;
    color: rgba(136, 170, 204, 0.5);
    display: flex;
    gap: 16px;
    pointer-events: none;
    user-select: none;
    z-index: 100;
  }

  .shortcuts-hint span {
    background: rgba(10, 26, 58, 0.6);
    padding: 2px 6px;
    border-radius: 4px;
    border: 1px solid rgba(136, 170, 204, 0.2);
  }

  @media (max-width: 768px) {
    .shortcuts-hint {
      display: none;
    }
  }
</style>
