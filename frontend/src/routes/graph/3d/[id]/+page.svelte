<script lang="ts">
  import { page } from '$app/stores';
  import { browser } from '$app/environment';
  import { getGraphData, getFullGraphData } from '$lib/api/graph';
  import LazyGraph3D from '$lib/components/LazyGraph3D.svelte';
  import type { GraphData } from '$lib/api/graph';

  let graphData: GraphData | null = $state(null);
  let loading = $state(true);
  let error = $state('');
  let showFullGraph = $state(false); // По умолчанию локальный вид

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

  // Отслеживаем изменение showFullGraph и загружаем данные
  $effect(() => {
    if (browser) {
      const mode = showFullGraph;
      const id = $page.params.id;
      if (id) {
        loadGraph(id);
      }
    }
  });
</script>

<div class="page">
  <div class="controls">
    <label class="toggle">
      <input type="checkbox" bind:checked={showFullGraph} />
      <span>Показать все заметки ({showFullGraph ? 'включено' : 'выключено'})</span>
    </label>
  </div>
  
  <!-- Stats -->
  {#if !loading && !error && graphData}
    <div class="stats-bar">
      <span class="stats-item">
        <strong>{graphData.nodes.length}</strong> nodes
      </span>
      <span class="stats-item">
        <strong>{graphData.links.length}</strong> links
      </span>
      {#if showFullGraph}
        <span class="stats-mode">(Full graph)</span>
      {:else}
        <span class="stats-mode">(Local view)</span>
      {/if}
    </div>
  {/if}
  
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

  .stats-bar {
    position: absolute;
    top: 140px;
    right: 20px;
    z-index: 10000;
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 10px 16px;
    background: rgba(0, 0, 0, 0.85);
    border-radius: 8px;
    font-size: 14px;
    color: #94a3b8;
    border: 1px solid rgba(255, 255, 255, 0.2);
    backdrop-filter: blur(10px);
  }

  .stats-item {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .stats-item strong {
    color: #88aaff;
    font-weight: 600;
  }

  .stats-mode {
    margin-left: 8px;
    font-style: italic;
    color: #64748b;
    font-size: 12px;
  }
</style>
