<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import GraphCanvas from '$lib/components/GraphCanvas.svelte';
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
      nodes = data.nodes;
      links = data.links;
    } catch (e) {
      error = 'Failed to load graph data';
      console.error(e);
    } finally {
      loading = false;
    }
  });
</script>

<BackButton href="/" />
<h1>Note Graph</h1>

{#if loading}
  <p>Loading graph...</p>
{:else if error}
  <p class="error">{error}</p>
{:else}
  <div class="graph-container">
    <GraphCanvas {nodes} {links} />
  </div>
{/if}

<style>
  .graph-container { width: 100%; height: 70vh; border: 1px solid #ccc; border-radius: 8px; overflow: hidden; background: #f0f2f5; }
  .error { color: red; }
</style>