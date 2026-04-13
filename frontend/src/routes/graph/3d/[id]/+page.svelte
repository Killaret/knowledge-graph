<script lang="ts">
  import { page } from '$app/stores';
  import { getGraphData, getFullGraphData } from '$lib/api/graph';
  import LazyGraph3D from '$lib/components/LazyGraph3D.svelte';
  import type { GraphData } from '$lib/api/graph';

  let graphData: GraphData | null = $state(null);
  let loading = $state(true);
  let error = $state('');
  let showFullGraph = $state(true); // По умолчанию показываем все заметки

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
      if (showFullGraph) {
        // Загружаем полный граф всех заметок
        graphData = await getFullGraphData();
      } else {
        // Загружаем локальный граф вокруг заметки
        graphData = await getGraphData(noteId, 2);
      }
    } catch (e) {
      error = 'Failed to load graph data';
      console.error(e);
    } finally {
      loading = false;
    }
  }

  function toggleGraphMode() {
    showFullGraph = !showFullGraph;
    const id = $page.params.id;
    if (id) {
      loadGraph(id);
    }
  }
</script>

<div class="page">
  <div class="controls">
    <label class="toggle">
      <input type="checkbox" bind:checked={showFullGraph} onchange={toggleGraphMode} />
      <span>Показать все заметки ({showFullGraph ? 'включено' : 'выключено'})</span>
    </label>
  </div>
  
  {#if loading}
    <div class="center">Загрузка...</div>
  {:else if error}
    <div class="center error">{error}</div>
  {:else if graphData}
    <LazyGraph3D data={graphData} />
  {/if}
</div>

<style>
  .page {
    position: relative;
    width: 100%;
    height: 100vh;
    overflow: hidden;
  }
  .controls {
    position: absolute;
    top: 80px;
    right: 20px;
    z-index: 10000;
    background: rgba(0, 0, 0, 0.85);
    padding: 12px 16px;
    border-radius: 8px;
    color: white;
    border: 1px solid rgba(255, 255, 255, 0.3);
    backdrop-filter: blur(10px);
  }
  .toggle {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;
    font-size: 14px;
  }
  .toggle input {
    cursor: pointer;
    width: 18px;
    height: 18px;
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
