<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { searchNotes, type Note } from '$lib/api/notes';
  import SearchBar from '$lib/components/SearchBar.svelte';
  import NoteCard from '$lib/components/NoteCard.svelte';

  let notes = $state<Note[]>([]);
  let loading = $state(false);
  let error = $state('');
  let total = $state(0);
  let currentPage = $state(1);
  let totalPages = $state(1);
  const size = 20;

  // Get parameters from URL and perform search
  $effect(() => {
    const q = $page.url.searchParams.get('q') || '';
    const pageParam = $page.url.searchParams.get('page');
    const pageNum = pageParam ? parseInt(pageParam, 10) : 1;
    currentPage = pageNum > 0 ? pageNum : 1;

    if (q) {
      performSearch(q, currentPage);
    } else {
      // If query is empty, show empty state
      notes = [];
      total = 0;
    }
  });

  async function performSearch(query: string, pageNum: number) {
    loading = true;
    error = '';
    try {
      const response = await searchNotes(query, pageNum, size);
      notes = response.data;
      total = response.total;
      totalPages = response.totalPages;
    } catch (e) {
      error = 'Failed to perform search. Please try again later.';
      console.error('Search error:', e);
    } finally {
      loading = false;
    }
  }

  // Navigate to different page
  function goToPage(newPage: number) {
    if (newPage < 1 || newPage > totalPages) return;
    const q = $page.url.searchParams.get('q') || '';
    goto(`/search?q=${encodeURIComponent(q)}&page=${newPage}`);
  }

  function getPluralForm(count: number, one: string, few: string, many: string): string {
    const lastDigit = count % 10;
    const lastTwoDigits = count % 100;
    
    if (lastTwoDigits >= 11 && lastTwoDigits <= 14) {
      return many;
    }
    if (lastDigit === 1) {
      return one;
    }
    if (lastDigit >= 2 && lastDigit <= 4) {
      return few;
    }
    return many;
  }
</script>

<div class="search-page">
  <div class="search-header">
    <h1>Search Notes</h1>
    <SearchBar placeholder="Enter your search query..." autoFocus={true} />
  </div>

  {#if loading}
    <div class="center">
      <div class="spinner"></div>
      <p>Searching...</p>
    </div>
  {:else if error}
    <div class="error-message">{error}</div>
  {:else if $page.url.searchParams.get('q') && notes.length === 0}
    <div class="no-results">
      <div class="no-results-icon">No results found</div>
      <h2>No results found</h2>
      <p>No notes found for query "{$page.url.searchParams.get('q')}".</p>
      <p>Try using different keywords or check spelling.</p>
    </div>
  {:else if $page.url.searchParams.get('q')}
    <div class="search-stats">
      Found: {total} {getPluralForm(total, 'note', 'notes', 'notes')}
    </div>

    <div class="notes-grid">
      {#each notes as note}
        <NoteCard {note} />
      {/each}
    </div>

    <!-- Pagination -->
    {#if totalPages > 1}
      <div class="pagination">
        <button 
          onclick={() => goToPage(currentPage - 1)} 
          disabled={currentPage === 1}
          class="pagination-button"
        >
          Previous
        </button>
        
        <span class="pagination-info">
          Page {currentPage} of {totalPages}
        </span>
        
        <button 
          onclick={() => goToPage(currentPage + 1)} 
          disabled={currentPage === totalPages}
          class="pagination-button"
        >
          Next
        </button>
      </div>
    {/if}
  {:else}
    <div class="empty-state">
      <div class="empty-icon">Search</div>
      <h2>Search Notes</h2>
      <p>Enter keywords above to search through your notes.</p>
      <p>Search supports both Russian and English languages with automatic stemming.</p>
    </div>
  {/if}
</div>

<style>
  .search-page {
    max-width: 1200px;
    margin: 0 auto;
    padding: 1.5rem;
    min-height: 100vh;
  }

  .search-header {
    text-align: center;
    margin-bottom: 2rem;
  }

  h1 {
    margin-bottom: 2rem;
    color: var(--color-text, #1f2937);
  }

  .center {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 4px solid var(--color-border, #e2e8f0);
    border-top: 4px solid var(--color-primary, #3b82f6);
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-bottom: 1rem;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  .error-message {
    padding: 1rem;
    background: #fee2e2;
    color: #b91c1c;
    border-radius: 0.5rem;
    text-align: center;
    margin: 2rem 0;
  }

  .no-results, .empty-state {
    text-align: center;
    padding: 3rem;
    color: var(--color-text-secondary, #6b7280);
  }

  .no-results-icon, .empty-icon {
    font-size: 3rem;
    margin-bottom: 1rem;
    opacity: 0.5;
  }

  .no-results h2, .empty-state h2 {
    color: var(--color-text, #1f2937);
    margin-bottom: 1rem;
  }

  .search-stats {
    margin: 1rem 0 2rem 0;
    color: var(--color-text-secondary, #475569);
    font-size: 0.9rem;
    text-align: center;
  }

  .notes-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
    gap: 1.5rem;
    margin-top: 1.5rem;
  }

  .pagination {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 1rem;
    margin-top: 3rem;
  }

  .pagination-button {
    padding: 0.5rem 1rem;
    background: var(--color-primary, #3b82f6);
    color: white;
    border: none;
    border-radius: 0.25rem;
    cursor: pointer;
    font-weight: 500;
    transition: background 0.2s;
  }

  .pagination-button:hover:not(:disabled) {
    background: var(--color-primary-dark, #2563eb);
  }

  .pagination-button:disabled {
    background: var(--color-border, #e2e8f0);
    color: var(--color-text-secondary, #6b7280);
    cursor: not-allowed;
  }

  .pagination-info {
    color: var(--color-text-secondary, #6b7280);
    font-weight: 500;
  }

  @media (max-width: 768px) {
    .search-page {
      padding: 1rem;
    }

    .notes-grid {
      grid-template-columns: 1fr;
    }

    .pagination {
      flex-direction: column;
      gap: 0.5rem;
    }

    .pagination-button {
      width: 100%;
    }
  }
</style>
