<script lang="ts">
  import '../app.css';
  import { setContext } from 'svelte';
  import LeftToolbar from '$lib/components/LeftToolbar.svelte';
  import TopSearchBar from '$lib/components/TopSearchBar.svelte';
  import RightPanel from '$lib/components/RightPanel.svelte';
  import KeyboardShortcuts from '$lib/components/KeyboardShortcuts.svelte';
  
  let { children } = $props();

  // Search context for components
  let searchFocusHandler: (() => void) | null = null;
  function registerSearchFocus(handler: () => void) {
    searchFocusHandler = handler;
  }
  setContext('registerSearchFocus', registerSearchFocus);

  function handleSearchFocus() {
    searchFocusHandler?.();
  }
</script>

<div class="app-layout">
  <!-- Header with search -->
  <header class="app-header">
    <div class="header-brand">
      <span class="brand-icon">🌌</span>
      <span class="brand-text">Knowledge Graph</span>
    </div>
    <TopSearchBar />
  </header>
  
  <!-- Main content -->
  <main class="app-main">
    {@render children()}
  </main>
  
  <!-- Floating UI elements -->
  <LeftToolbar />
  <RightPanel />
  <KeyboardShortcuts onSearchFocus={handleSearchFocus} />
</div>

<style>
  .app-layout {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
  }
  
  .app-header {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    height: 4rem;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 1.5rem;
    background: rgba(0, 0, 0, 0.3);
    backdrop-filter: blur(8px);
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    z-index: 40;
  }
  
  .header-brand {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }
  
  .brand-icon {
    font-size: 1.25rem;
  }
  
  .brand-text {
    font-weight: 600;
    font-size: 1rem;
    color: rgba(255, 255, 255, 0.9);
  }
  
  .app-main {
    flex: 1;
    padding-top: 4rem;
  }
  
  @media (max-width: 768px) {
    .app-header {
      padding: 0 1rem;
    }
    
    .brand-text {
      display: none;
    }
    
    .app-main {
      padding-bottom: 5rem;
    }
  }
</style>