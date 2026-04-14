<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import { goto } from '$app/navigation';
  import { getNotes, type Note } from '$lib/api/notes';
  import { getGraphData, getFullGraphData, type GraphData } from '$lib/api/graph';
  import GraphCanvas from '$lib/components/GraphCanvas.svelte';
  import NoteSidePanel from '$lib/components/NoteSidePanel.svelte';
  import BackButton from '$lib/components/BackButton.svelte';

  let notes: Note[] = $state([]);
  let graphData: GraphData = $state({ nodes: [], links: [] });
  let loading = $state(true);
  let error = $state('');
  let selectedNodeId: string | null = $state(null);
  let showFullGraph = $state(false); // По умолчанию локальный вид

  async function loadGraphData() {
    loading = true;
    error = '';
    try {
      if (showFullGraph) {
        // Загружаем полный граф всех заметок
        graphData = await getFullGraphData();
      } else {
        // Загружаем локальный граф
        notes = await getNotes();
        if (notes.length > 0) {
          const centerNote = notes[0];
          graphData = await getGraphData(centerNote.id, 3);
        } else {
          error = 'No notes found. Create some notes first.';
        }
      }
    } catch (e) {
      console.error('Failed to load graph:', e);
      error = 'Failed to load graph data';
    } finally {
      loading = false;
    }
  }

  onMount(async () => {
    if (!browser) return;
    await loadGraphData();
  });

  function handleNodeSelect(nodeId: string) {
    selectedNodeId = nodeId;
  }

  // Отслеживаем изменение showFullGraph и загружаем данные
  $effect(() => {
    if (browser) {
      showFullGraph; // Access for reactivity
      loadGraphData();
    }
  });
</script>

<div class="graph-page">
  <BackButton href="/" />
  
  <h1>Knowledge Graph</h1>
  
  <div class="controls">
    <label class="toggle">
      <input type="checkbox" bind:checked={showFullGraph} />
      <span>Показать все заметки ({showFullGraph ? 'включено' : 'выключено'})</span>
    </label>
  </div>
  
  <!-- Stats -->
  {#if !loading && !error}
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
    <div class="center">
      <div class="spinner"></div>
      <p>Loading graph...</p>
    </div>
  {:else if error}
    <div class="error">
      <p>{error}</p>
      <button onclick={() => goto('/')}>Go Home</button>
    </div>
  {:else}
    <div class="graph-container">
      {#if graphData.nodes.length > 0}
        {#key graphData.nodes.length + '-' + graphData.links.length}
          <GraphCanvas 
            nodes={graphData.nodes}
            links={graphData.links}
            onNodeClick={(node) => handleNodeSelect(node.id)}
          />
        {/key}
      {:else}
        <div class="empty">
          <p>No graph data available</p>
        </div>
      {/if}
    </div>
  {/if}
</div>

{#if selectedNodeId}
  <NoteSidePanel 
    nodeId={selectedNodeId}
    onClose={() => selectedNodeId = null}
    onEdit={(id) => goto(`/notes/${id}/edit`)}
    onDelete={() => {
      selectedNodeId = null;
      // Reload graph
      window.location.reload();
    }}
  />
{/if}

<style>
  .graph-page {
    height: 100vh;
    display: flex;
    flex-direction: column;
    padding: 20px;
    background: linear-gradient(135deg, #0a0a1a 0%, #1a1a2e 100%);
    color: white;
  }

  h1 {
    margin: 0 0 20px 0;
    font-size: 1.5rem;
  }

  .controls {
    position: absolute;
    top: 80px;
    right: 20px;
    z-index: 1000;
    background: rgba(0, 0, 0, 0.8);
    padding: 12px 16px;
    border-radius: 8px;
    color: white;
    border: 1px solid rgba(255, 255, 255, 0.2);
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

  .stats-bar {
    display: flex;
    align-items: center;
    gap: 16px;
    margin: 10px 0 20px 0;
    padding: 10px 16px;
    background: rgba(0, 0, 0, 0.6);
    border-radius: 8px;
    font-size: 14px;
    color: #94a3b8;
  }

  .stats-item {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .stats-item strong {
    color: #3b82f6;
    font-weight: 600;
  }

  .stats-mode {
    margin-left: auto;
    font-style: italic;
    color: #64748b;
  }

  .graph-container {
    flex: 1;
    position: relative;
    min-height: 0;
    border-radius: 12px;
    overflow: hidden;
  }

  .center {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    gap: 16px;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 3px solid rgba(255,255,255,0.2);
    border-top-color: #3b82f6;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .error {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    gap: 16px;
    color: #ef4444;
  }

  .error button {
    padding: 8px 16px;
    background: #3b82f6;
    color: white;
    border: none;
    border-radius: 6px;
    cursor: pointer;
  }

  .empty {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: #94a3b8;
  }
</style>
