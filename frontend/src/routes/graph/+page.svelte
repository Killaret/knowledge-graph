<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import { goto } from '$app/navigation';
  import { getNotes, type Note } from '$lib/api/notes';
  import { getGraphData, getFullGraphData, type GraphData } from '$lib/api/graph';
  import Graph3D from '$lib/components/Graph3D.svelte';
  import NoteSidePanel from '$lib/components/NoteSidePanel.svelte';
  import BackButton from '$lib/components/BackButton.svelte';

  let notes: Note[] = $state([]);
  let graphData: GraphData = $state({ nodes: [], links: [] });
  let loading = $state(true);
  let error = $state('');
  let selectedNodeId: string | null = $state(null);
  let showFullGraph = $state(true); // По умолчанию показываем все заметки

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

  function toggleGraphMode() {
    showFullGraph = !showFullGraph;
    loadGraphData();
  }
</script>

<div class="graph-page">
  <BackButton href="/" />
  
  <h1>Knowledge Graph</h1>
  
  <div class="controls">
    <label class="toggle">
      <input type="checkbox" bind:checked={showFullGraph} onchange={toggleGraphMode} />
      <span>Показать все заметки ({showFullGraph ? 'включено' : 'выключено'})</span>
    </label>
  </div>
  
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
        <Graph3D 
          data={graphData} 
          centerNodeId={graphData.nodes[0]?.id || ''}
          onNodeClick={(node) => handleNodeSelect(node.id)}
        />
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
    overflow: hidden;
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
  }
  .toggle input {
    cursor: pointer;
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
