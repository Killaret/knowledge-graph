<script lang="ts">
  import { tick } from 'svelte';
  import type { Note } from '$lib/api/notes';
  import { searchNotes } from '$lib/api/notes';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  
  interface Props {
    onResultSelect?: (result: Note) => void;
  }
  
  const { onResultSelect }: Props = $props();
  
  let isOpen = $state(false);
  let query = $state('');
  let results = $state<Note[]>([]);
  let isLoading = $state(false);
  let inputRef = $state<HTMLInputElement | null>(null);
  let searchContainerRef = $state<HTMLDivElement | null>(null);
  let debounceTimer = $state<ReturnType<typeof setTimeout> | null>(null);
  
  const isGraphPage = $derived($page.url.pathname.startsWith('/graph'));
  
  // Keyboard shortcut Ctrl+F
  $effect(() => {
    const handleKeydown = (e: KeyboardEvent) => {
      if (e.ctrlKey && e.key === 'f') {
        e.preventDefault();
        openSearch();
      }
      if (e.key === 'Escape' && isOpen) {
        closeSearch();
      }
    };
    
    window.addEventListener('keydown', handleKeydown);
    return () => window.removeEventListener('keydown', handleKeydown);
  });
  
  // Click outside to close
  $effect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (isOpen && searchContainerRef && !searchContainerRef.contains(e.target as Node)) {
        closeSearch();
      }
    };
    
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  });
  
  async function openSearch() {
    isOpen = true;
    await tick();
    inputRef?.focus();
  }
  
  function closeSearch() {
    isOpen = false;
    query = '';
    results = [];
  }
  
  function handleInput() {
    if (debounceTimer) clearTimeout(debounceTimer);
    
    if (!query.trim()) {
      results = [];
      return;
    }
    
    isLoading = true;
    debounceTimer = setTimeout(async () => {
      try {
        const response = await searchNotes(query);
        results = response.data || [];
      } catch (e) {
        console.error('Search failed:', e);
        results = [];
      } finally {
        isLoading = false;
      }
    }, 500);
  }
  
  function selectResult(result: Note) {
    if (onResultSelect) {
      onResultSelect(result);
    } else if (isGraphPage) {
      // On graph page, just notify parent to focus on node
      // The graph page should handle this
    } else {
      // Navigate to note
      goto(`/notes/${result.id}`);
    }
    closeSearch();
  }
  
  function highlightMatch(text: string, query: string): string {
    if (!query) return text;
    const regex = new RegExp(`(${query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')})`, 'gi');
    return text.replace(regex, '<mark class="search-highlight">$1</mark>');
  }
</script>

<div class="search-container" class:open={isOpen} bind:this={searchContainerRef}>
  {#if !isOpen}
    <button
      class="search-trigger"
      onclick={openSearch}
      aria-label="Поиск (Ctrl+F)"
    >
      <svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="11" cy="11" r="8"/>
        <line x1="21" y1="21" x2="16.65" y2="16.65"/>
      </svg>
    </button>
  {:else}
    <div class="search-expanded">
      <div class="search-input-wrapper">
        <svg class="search-icon" viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="8"/>
          <line x1="21" y1="21" x2="16.65" y2="16.65"/>
        </svg>
        
        <input
          bind:this={inputRef}
          type="text"
          bind:value={query}
          oninput={handleInput}
          placeholder="Поиск по заметкам..."
          class="search-input"
        />
        
        {#if isLoading}
          <span class="loading-indicator">⌛</span>
        {/if}
        
        <button class="close-btn" onclick={closeSearch} aria-label="Закрыть">
          <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="18" y1="6" x2="6" y2="18"/>
            <line x1="6" y1="6" x2="18" y2="18"/>
          </svg>
        </button>
      </div>
      
      {#if results.length > 0}
        <div class="search-results">
          {#each results as result}
            <button
              class="result-item"
              onclick={() => selectResult(result)}
            >
              <div class="result-title">
                {@html highlightMatch(result.title, query)}
              </div>
              <div class="result-preview">
                {@html highlightMatch(result.content.slice(0, 80), query)}...
              </div>
            </button>
          {/each}
        </div>
      {:else if query && !isLoading}
        <div class="no-results">Ничего не найдено</div>
      {/if}
    </div>
  {/if}
</div>

<style>
  .search-container {
    position: relative;
    z-index: 50;
  }
  
  .search-trigger {
    width: 2.5rem;
    height: 2.5rem;
    display: flex;
    align-items: center;
    justify-content: center;
    
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 0.75rem;
    color: rgba(255, 255, 255, 0.7);
    cursor: pointer;
    transition: all 0.2s ease;
  }
  
  .search-trigger:hover {
    background: rgba(255, 255, 255, 0.15);
    color: white;
  }
  
  .search-container.open {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    padding: 1rem;
    background: rgba(0, 0, 0, 0.5);
    backdrop-filter: blur(12px);
  }
  
  @media (min-width: 768px) {
    .search-container.open {
      position: absolute;
      top: 50%;
      right: 0;
      left: auto;
      transform: translateY(-50%);
      width: 400px;
      padding: 0;
      background: transparent;
      backdrop-filter: none;
    }
  }
  
  .search-expanded {
    background: rgba(0, 0, 0, 0.6);
    backdrop-filter: blur(16px);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 1rem;
    overflow: hidden;
  }
  
  @media (max-width: 768px) {
    .search-expanded {
      max-height: calc(100vh - 2rem);
      display: flex;
      flex-direction: column;
    }
  }
  
  .search-input-wrapper {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }
  
  .search-icon {
    color: rgba(255, 255, 255, 0.5);
    flex-shrink: 0;
  }
  
  .search-input {
    flex: 1;
    background: transparent;
    border: none;
    color: white;
    font-size: 1rem;
    outline: none;
  }
  
  .search-input::placeholder {
    color: rgba(255, 255, 255, 0.4);
  }
  
  .loading-indicator {
    font-size: 0.875rem;
    animation: pulse 1s infinite;
  }
  
  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.5; }
  }
  
  .close-btn {
    width: 1.75rem;
    height: 1.75rem;
    display: flex;
    align-items: center;
    justify-content: center;
    
    background: rgba(255, 255, 255, 0.1);
    border: none;
    border-radius: 0.5rem;
    color: rgba(255, 255, 255, 0.6);
    cursor: pointer;
    transition: all 0.2s ease;
    flex-shrink: 0;
  }
  
  .close-btn:hover {
    background: rgba(255, 255, 255, 0.2);
    color: white;
  }
  
  .search-results {
    max-height: 300px;
    overflow-y: auto;
  }
  
  @media (max-width: 768px) {
    .search-results {
      max-height: calc(100vh - 200px);
    }
  }
  
  .result-item {
    width: 100%;
    padding: 0.875rem 1rem;
    text-align: left;
    background: transparent;
    border: none;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    color: white;
    cursor: pointer;
    transition: background 0.15s ease;
  }
  
  .result-item:hover {
    background: rgba(255, 255, 255, 0.05);
  }
  
  .result-item:last-child {
    border-bottom: none;
  }
  
  .result-title {
    font-weight: 500;
    font-size: 0.9375rem;
    margin-bottom: 0.25rem;
  }
  
  :global(.search-highlight) {
    background: rgba(99, 102, 241, 0.4);
    color: #c7d2fe;
    border-radius: 0.25rem;
    padding: 0 0.125rem;
  }
  
  .result-preview {
    font-size: 0.8125rem;
    color: rgba(255, 255, 255, 0.5);
    line-height: 1.4;
  }
  
  .no-results {
    padding: 2rem 1rem;
    text-align: center;
    color: rgba(255, 255, 255, 0.5);
    font-size: 0.875rem;
  }
</style>
