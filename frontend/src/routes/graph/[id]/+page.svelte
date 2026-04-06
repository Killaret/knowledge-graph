<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import GraphCanvas from '$lib/components/GraphCanvas.svelte';
  import { getGraphData } from '$lib/api/graph';

  let nodes: Array<{ id: string; title: string }> = $state([]);
  let links: Array<{ source: string; target: string; weight: number }> = $state([]);
  let loading = $state(true);
  let error = $state('');

  const id = $page.params.id;

  onMount(async () => {
    try {
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

<h1>Созвездие заметок</h1>

{#if loading}
  <p>Загрузка графа...</p>
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