<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import SmartGraph from '$lib/components/SmartGraph.svelte';
  import { getGraphData } from '$lib/api/graph';
  import BackButton from '$lib/components/BackButton.svelte';
  import type { GraphNode, GraphLink } from '$lib/api/graph';

  let nodes: GraphNode[] = $state([]);
  let links: GraphLink[] = $state([]);
  let loading = $state(true);
  let error = $state('');

  function getRouteId(): string {
    const id = $page.params.id;
    if (!id) throw new Error('Missing route parameter: id');
    return id;
  }

  onMount(async () => {
    try {
      const id = getRouteId();
      const data = await getGraphData(id);
      // Map nodes with defaults for SmartGraph compatibility
      nodes = data.nodes.map((n: GraphNode) => ({
        ...n,
        type: n.type ?? 'default',
        size: n.size ?? 5
      }));
      links = data.links.map((l: GraphLink) => ({
        ...l,
        weight: l.weight ?? 1
      }));
    } catch (e) {
      error = 'Failed to load graph data';
      console.error(e);
    } finally {
      loading = false;
    }
  });
</script>

<div class="graph-page">
  <header class="graph-header">
    <BackButton href="/" />
    <h1>Knowledge Constellation</h1>
    <span class="hint">Drag to rotate/pan • Scroll to zoom • Click node to open</span>
  </header>

  {#if loading}
    <div class="loading-state">
      <div class="spinner"></div>
      <p>Loading constellation...</p>
    </div>
  {:else if error}
    <div class="error-state">
      <p class="error">{error}</p>
    </div>
  {:else}
    <div class="graph-container">
      <SmartGraph {nodes} {links} />
    </div>
  {/if}
</div>

<style>
  .graph-page {
    display: flex;
    flex-direction: column;
    height: 100vh;
    background: linear-gradient(145deg, #0a1a3a, #020617);
    color: #e0e0e0;
  }

  .graph-header {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1rem 1.5rem;
    border-bottom: 1px solid rgba(100, 150, 200, 0.2);
    background: rgba(10, 26, 58, 0.8);
    backdrop-filter: blur(10px);
  }

  .graph-header h1 {
    margin: 0;
    font-size: 1.5rem;
    font-weight: 600;
    color: #ffdd88;
    text-shadow: 0 0 10px rgba(255, 200, 100, 0.3);
  }

  .hint {
    margin-left: auto;
    font-size: 0.85rem;
    color: #88aacc;
    opacity: 0.8;
  }

  .graph-container {
    flex: 1;
    overflow: hidden;
    position: relative;
  }

  .loading-state,
  .error-state {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 1rem;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 3px solid rgba(100, 150, 200, 0.3);
    border-top-color: #ffdd88;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .error {
    color: #ff6b6b;
    padding: 1rem;
    background: rgba(255, 100, 100, 0.1);
    border-radius: 8px;
    border: 1px solid rgba(255, 100, 100, 0.3);
  }
</style>