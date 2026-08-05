<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { browser } from "$app/environment";
  import { onMount, untrack } from "svelte";
  import { SearchQuery } from "$entities";
  import { addJitter } from "$shared/utils/jitter";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  // Props
  const { placeholder = t("search.placeholder"), autoFocus = false } = $props();

  let query = $state("");
  let inputElement: HTMLInputElement;

  // Sync with URL parameter when URL changes (not when query changes)
  // Use untrack to prevent this effect from re-running when query changes
  $effect(() => {
    const q = SearchQuery.fromURL($page.url.searchParams.get("q")).value;
    if (q !== untrack(() => query)) {
      query = q;
    }
  });

  // Perform search (navigate to search page)
  function doSearch() {
    const searchQuery = new SearchQuery(query);
    if (!searchQuery.isEmpty()) {
      goto(`/search?q=${searchQuery.toURL()}`);
    }
  }

  // Debounce for automatic search while typing (optional)
  let debounceTimer: ReturnType<typeof setTimeout> | null = null;
  function handleInput() {
    if (!browser) return;
    if (debounceTimer) clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      doSearch();
    }, addJitter(500)); // search after ~500ms of no typing (with jitter)
  }

  // Handle Enter key
  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === "Enter") {
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
    {placeholder}
    aria-label={t("search.label")}
    class="search-input"
  />
  <button type="button" onclick={doSearch} class="search-button" aria-label={t("search.label")}>
    {t("search.label")}
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
    border: 1px solid var(--carbon-border, #2d2d3d);
    border-radius: 0.5rem;
    font-size: 1rem;
    background: var(--carbon-black, #050508);
    color: var(--carbon-text, #f0f0f5);
    transition:
      border-color 0.2s,
      box-shadow 0.2s;
  }

  .search-input::placeholder {
    color: var(--carbon-text-dim, #5a5a6e);
  }

  .search-input:focus {
    outline: none;
    border-color: var(--carbon-glow-cyan, #22d3ee);
    box-shadow: var(--carbon-focus-ring, 0 0 0 3px rgba(34, 211, 238, 0.15));
  }

  .search-button {
    padding: 0.75rem 1.5rem;
    background: var(--carbon-gradient-primary, linear-gradient(135deg, #22d3ee 0%, #8b5cf6 100%));
    color: white;
    border: none;
    border-radius: 0.5rem;
    cursor: pointer;
    font-size: 1rem;
    font-weight: 600;
    transition: all var(--carbon-transition, 0.25s ease);
    box-shadow: var(--carbon-glow-primary, 0 4px 15px rgba(34, 211, 238, 0.3));
  }

  .search-button:hover {
    filter: brightness(1.08);
    transform: translateY(-1px);
    box-shadow:
      0 8px 25px rgba(34, 211, 238, 0.35),
      0 0 30px rgba(139, 92, 246, 0.25);
  }

  .search-button:active {
    transform: translateY(0);
  }
</style>
