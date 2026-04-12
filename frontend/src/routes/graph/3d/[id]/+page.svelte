<script lang="ts">
  import { page } from '$app/stores';
  import { getGraphData } from '$lib/api/graph';
  import LazyGraph3D from '$lib/components/LazyGraph3D.svelte';
  import type { GraphData } from '$lib/api/graph';

  let graphData: GraphData | null = $state(null);
  let centerNodeId: string = $state('');
  let loading = $state(true);
  let error = $state('');

  $effect(() => {
    const id = $page.params.id;
    if (id) {
      centerNodeId = id;
      loadGraph(id);
    }
  });

  async function loadGraph(noteId: string) {
    loading = true;
    error = '';
    try {
      // Use depth=2 for performance optimization
      graphData = await getGraphData(noteId, 2);
    } catch (e) {
      error = 'Failed to load graph data';
      console.error(e);
    } finally {
      loading = false;
    }
  }
</script>

<div class="page">
  {#if loading}
    <div class="center">Загрузка...</div>
  {:else if error}
    <div class="center error">{error}</div>
  {:else if graphData}
    <LazyGraph3D data={graphData} {centerNodeId} />
  {/if}
</div>

<style>
  .page {
    position: relative;
    width: 100%;
    height: 100vh;
    overflow: hidden;
  }
  .center {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: white;
  }
  .error {
    color: #ff6666;
  }
  </style>
