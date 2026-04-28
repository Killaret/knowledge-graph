<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { browser } from '$app/environment';
  import { onMount, untrack } from 'svelte';

  // Props
  const { placeholder = 'Search notes (Russian & English)...', autoFocus = false } = $props();

  let query = $state('');
  let inputElement: HTMLInputElement;

  // Sync with URL parameter when URL changes (not when query changes)
  // Use untrack to prevent this effect from re-running when query changes
  $effect(() => {
    const q = $page.url.searchParams.get('q') || '';
    if (q !== untrack(() => query)) {
      query = q;
    }
  });

  // Perform search (navigate to search page)
  function doSearch() {
    const trimmed = query.trim();
    if (trimmed) {
      goto(`/search?q=${encodeURIComponent(trimmed)}`);
    }
    // Empty query - don't navigate
  }

  // Debounce for automatic search while typing (optional)
  let debounceTimer: ReturnType<typeof setTimeout> | null = null;
  function handleInput() {
    if (!browser) return;
    if (debounceTimer) clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      doSearch();
    }, 500); // search after 500ms of no typing
  }

  // Handle Enter key
  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      if (debounceTimer) clearTimeout(debounceTimer);
      doSearch();
    }
  }

  onMount(() => {
    if (autoFocus) inputElement?.focus();
    return () => {
      if (debounceTimer) clearTimeout(debounceTimer);
    };
  });
</script>

<div class="search-bar">
  <input
    bind:this={inputElement}
    type="search"
    bind:value={query}
    oninput={handleInput}
    onkeydown={handleKeyDown}
    placeholder={placeholder}
    aria-label="Search"
    class="search-input"
  />
  <button type="button" onclick={doSearch} class="search-button" aria-label="Search">
    Search
  </button>
</div>

<style>
  .search-bar {
    display: flex;
    gap: 0.5rem;
    width: 100%;
    max-width: 600px;
    margin: 0 auto;
  }
  
  .search-input {
    flex: 1;
    padding: 0.75rem 1rem;
    border: 2px solid var(--color-border, #e2e8f0);
    border-radius: 0.5rem;
    font-size: 1rem;
    background: white;
    transition: border-color 0.2s;
  }
  
  .search-input:focus {
    outline: none;
    border-color: var(--color-primary, #3b82f6);
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
  }
  
  .search-button {
    padding: 0.75rem 1.5rem;
    background: var(--color-primary, #3b82f6);
    color: white;
    border: none;
    border-radius: 0.5rem;
    cursor: pointer;
    font-size: 1rem;
    font-weight: 500;
    transition: background 0.2s;
  }
  
  .search-button:hover {
    background: var(--color-primary-dark, #2563eb);
  }
  
  .search-button:active {
    transform: translateY(0);
  }
</style>
