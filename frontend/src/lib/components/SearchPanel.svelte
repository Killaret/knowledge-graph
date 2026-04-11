<script lang="ts">
  import { fade, fly } from 'svelte/transition';
  import { tick } from 'svelte';
  
  let { 
    isOpen = false,
    onClose,
    onSearch,
    onResultClick,
    results = []
  }: {
    isOpen: boolean;
    onClose: () => void;
    onSearch: (query: string) => void;
    onResultClick: (result: any) => void;
    results: any[];
  } = $props();
  
  let searchQuery = $state('');
  let searchInput: HTMLInputElement | undefined = $state();
  let debounceTimer: ReturnType<typeof setTimeout>;
  
  $effect(() => {
    if (isOpen) {
      tick().then(() => {
        searchInput?.focus();
      });
    }
  });
  
  function handleInput() {
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      if (searchQuery.trim()) {
        onSearch(searchQuery.trim());
      }
    }, 500);
  }
  
  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      onClose();
    }
  }
  
  function handleBackdropClick(event: MouseEvent) {
    if (event.target === event.currentTarget) {
      onClose();
    }
  }
  
  function highlightMatch(text: string, query: string): string {
    if (!query) return text;
    const regex = new RegExp(`(${query})`, 'gi');
    return text.replace(regex, '<mark>$1</mark>');
  }
</script>

{#if isOpen}
  <div 
    class="search-overlay"
    onclick={handleBackdropClick}
    onkeydown={handleKeydown}
    role="dialog"
    aria-modal="true"
    aria-label="Поиск"
    tabindex="-1"
    transition:fade={{ duration: 200 }}
  >
    <div 
      class="search-container"
      transition:fly={{ y: -20, duration: 300 }}
    >
      <div class="search-header">
        <div class="search-input-wrapper">
          <svg class="search-icon" viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"></circle>
            <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
          </svg>
          <input
            bind:this={searchInput}
            bind:value={searchQuery}
            oninput={handleInput}
            type="text"
            class="search-input"
            placeholder="Поиск по заметкам..."
            aria-label="Поисковый запрос"
          />
          {#if searchQuery}
            <button 
              class="clear-btn"
              onclick={() => { searchQuery = ''; onSearch(''); }}
              aria-label="Очистить поиск"
            >
              <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="18" y1="6" x2="6" y2="18"></line>
                <line x1="6" y1="6" x2="18" y2="18"></line>
              </svg>
            </button>
          {/if}
        </div>
        <button class="close-btn" onclick={onClose} aria-label="Закрыть поиск">
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"></line>
            <line x1="6" y1="6" x2="18" y2="18"></line>
          </svg>
        </button>
      </div>
      
      {#if results.length > 0}
        <div class="results-list" role="list">
          {#each results as result}
            <div 
              class="result-item"
              onclick={() => onResultClick(result)}
              role="button"
              tabindex="0"
              onkeydown={(e) => e.key === 'Enter' && onResultClick(result)}
            >
              <div class="result-icon">
                {#if result.type === 'star'}⭐
                {:else if result.type === 'planet'}🪐
                {:else if result.type === 'comet'}☄️
                {:else if result.type === 'galaxy'}🌌
                {:else}📝
                {/if}
              </div>
              <div class="result-content">
                <h4 class="result-title">
                  {@html highlightMatch(result.title, searchQuery)}
                </h4>
                <p class="result-preview">
                  {@html highlightMatch(result.content?.slice(0, 100) || '', searchQuery)}...
                </p>
              </div>
            </div>
          {/each}
        </div>
      {:else if searchQuery.trim()}
        <div class="no-results">
          <p>Ничего не найдено для "{searchQuery}"</p>
        </div>
      {/if}
      
      <div class="search-footer">
        <span class="shortcut-hint">
          <kbd>Esc</kbd> — закрыть • <kbd>↑↓</kbd> — навигация • <kbd>Enter</kbd> — выбор
        </span>
      </div>
    </div>
  </div>
{/if}

<style>
  .search-overlay {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.7);
    backdrop-filter: blur(8px);
    display: flex;
    justify-content: center;
    align-items: flex-start;
    padding-top: 80px;
    z-index: 1600;
  }

  .search-container {
    width: 90%;
    max-width: 600px;
    background: linear-gradient(145deg, #0a1a3a, #1a2a4a);
    border: 1px solid rgba(255, 221, 136, 0.2);
    border-radius: 16px;
    box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
    overflow: hidden;
  }

  .search-header {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 16px 20px;
    border-bottom: 1px solid rgba(136, 170, 204, 0.2);
  }

  .search-input-wrapper {
    flex: 1;
    display: flex;
    align-items: center;
    gap: 12px;
    background: rgba(10, 26, 58, 0.6);
    border: 1px solid rgba(136, 170, 204, 0.2);
    border-radius: 10px;
    padding: 0 12px;
  }

  .search-icon {
    color: #88aacc;
    flex-shrink: 0;
  }

  .search-input {
    flex: 1;
    background: none;
    border: none;
    color: #e0e0e0;
    font-size: 1rem;
    padding: 12px 0;
    outline: none;
  }

  .search-input::placeholder {
    color: rgba(136, 170, 204, 0.5);
  }

  .clear-btn {
    background: none;
    border: none;
    color: #88aacc;
    cursor: pointer;
    padding: 4px;
    border-radius: 4px;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .clear-btn:hover {
    color: #ff6b6b;
  }

  .close-btn {
    background: none;
    border: none;
    color: #88aacc;
    cursor: pointer;
    padding: 8px;
    border-radius: 8px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s;
  }

  .close-btn:hover {
    color: #ffdd88;
    background: rgba(255, 221, 136, 0.1);
  }

  .results-list {
    max-height: 400px;
    overflow-y: auto;
    padding: 8px;
  }

  .result-item {
    width: 100%;
    padding: 12px 16px;
    background: rgba(10, 26, 58, 0.4);
    border: 1px solid rgba(136, 170, 204, 0.1);
    border-radius: 10px;
    margin-bottom: 8px;
    display: flex;
    align-items: flex-start;
    gap: 12px;
    cursor: pointer;
    transition: all 0.2s;
    text-align: left;
  }

  .result-item:hover {
    background: rgba(255, 221, 136, 0.1);
    border-color: rgba(255, 221, 136, 0.3);
  }

  .result-item:last-child {
    margin-bottom: 0;
  }

  .result-icon {
    font-size: 1.5rem;
    flex-shrink: 0;
  }

  .result-content {
    flex: 1;
    min-width: 0;
  }

  .result-title {
    color: #ffdd88;
    font-size: 1rem;
    font-weight: 600;
    margin: 0 0 4px 0;
  }

  :global(.result-title mark) {
    background: rgba(255, 221, 136, 0.3);
    color: #ffdd88;
    border-radius: 2px;
    padding: 0 2px;
  }

  .result-preview {
    color: #88aacc;
    font-size: 0.85rem;
    margin: 0;
    line-height: 1.4;
  }

  :global(.result-preview mark) {
    background: rgba(68, 170, 255, 0.3);
    color: #44aaff;
    border-radius: 2px;
    padding: 0 2px;
  }

  .no-results {
    padding: 40px 20px;
    text-align: center;
    color: #88aacc;
  }

  .search-footer {
    padding: 12px 20px;
    border-top: 1px solid rgba(136, 170, 204, 0.2);
    background: rgba(10, 26, 58, 0.4);
  }

  .shortcut-hint {
    color: rgba(136, 170, 204, 0.6);
    font-size: 0.8rem;
  }

  .shortcut-hint kbd {
    background: rgba(136, 170, 204, 0.2);
    padding: 2px 6px;
    border-radius: 4px;
    font-family: inherit;
  }

  @media (max-width: 768px) {
    .search-overlay {
      padding-top: 20px;
    }

    .search-container {
      width: 95%;
      max-height: 80vh;
    }
  }
</style>
