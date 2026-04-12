<script lang="ts">
  import { page } from '$app/stores';
  import { getGraphData } from '$lib/api/graph';
  import Graph3D from '$lib/components/Graph3D.svelte';
  import type { GraphData } from '$lib/api/graph';

  let graphData: GraphData | null = $state(null);
  let loading = $state(true);
  let error = $state('');

  $effect(() => {
    const id = $page.params.id;
    if (id) {
      loadGraph(id);
    }
  });

  async function loadGraph(noteId: string) {
    loading = true;
    error = '';
    try {
      graphData = await getGraphData(noteId);
    } catch (e) {
      error = 'Не удалось загрузить данные графа';
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
    <Graph3D data={graphData} />
  {/if}

  <a href="/notes/{$page.params.id}" class="back-button">← Назад к заметке</a>
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
  .back-button {
    position: absolute;
    top: 20px;
    left: 20px;
    padding: 8px 16px;
    background: rgba(20,30,50,0.8);
    color: white;
    border-radius: 8px;
    text-decoration: none;
    backdrop-filter: blur(4px);
    border: 1px solid rgba(255,255,255,0.2);
    z-index: 5;
    transition: background 0.2s;
  }
  .back-button:hover {
    background: rgba(40,60,100,0.9);
  }
</style>
