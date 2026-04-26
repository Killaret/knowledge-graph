<script lang="ts">
  import { page } from '$app/stores';
  import { browser } from '$app/environment';
  import { goto } from '$app/navigation';
  import { getGraphData, getFullGraphData } from '$lib/api/graph';
  import { getNote } from '$lib/api/notes';
  import Graph3D from '$lib/components/Graph3D.svelte';
  import type { GraphData } from '$lib/api/graph';

  let graphData: GraphData | null = $state(null);
  let error = $state('');
  let showFullGraph = $state(false); // По умолчанию локальный вид
  let currentNoteId: string | null = $state(null);
  let graphRef: ReturnType<typeof Graph3D> | null = $state(null);
  let isLoadingFull = $state(false); // Background loading indicator only

  // Единый эффект для загрузки графа при изменении ID или режима
  $effect(() => {
    if (!browser) return;

    const id = $page.params.id;
    // Explicitly track showFullGraph as a dependency
    const showFull = showFullGraph;

    if (id && id !== currentNoteId) {
      currentNoteId = id;
    }

    if (currentNoteId) {
      console.log('[3D Page] Effect triggered:', { id: currentNoteId, showFullGraph: showFull });
      // Expose centerNodeId to window for tests
      (window as any).centerNodeId = currentNoteId;
      loadGraphProgressive(currentNoteId);
    }
  });

  async function loadGraphProgressive(noteId: string) {
    console.log('[3D Page] loadGraphProgressive called:', { noteId, showFullGraph });
    // Progressive loading: no spinner, show graph immediately
    error = '';
    
    try {
      let initialData: GraphData;
      
      if (showFullGraph) {
        // For full graph, still load progressively
        initialData = await getFullGraphData(50); // Load first 50 nodes initially
        console.log('[3D Page] Loaded initial full graph (limited):', initialData.nodes.length, 'nodes');
        
        graphData = initialData;
        // Graph is shown immediately without waiting
        
        // Load remaining full graph in background
        isLoadingFull = true;
        getFullGraphData().then((fullData: GraphData) => {
          console.log('[3D Page] Background loaded full graph:', fullData.nodes.length, 'nodes');
          if (graphRef && fullData.nodes.length > initialData.nodes.length) {
            graphRef.addData(fullData);
          }
          isLoadingFull = false;
        });
      } else {
        // TWO-PHASE LOADING for local graph
        // Phase 1: Load initial data (depth 1) - immediate display
        initialData = await getGraphData(noteId, 1);
        console.log('[3D Page] Phase 1 - Loaded initial graph (depth 1):', initialData.nodes.length, 'nodes');
        
        // Check if start node is in the data - backend may not include it
        const hasStartNode = initialData.nodes.some((n: { id: string }) => n.id === noteId);
        if (!hasStartNode) {
          console.log('[3D Page] Start node not in response, fetching separately...');
          try {
            const startNote = await getNote(noteId);
            initialData.nodes.unshift({
              id: startNote.id,
              title: startNote.title,
              type: startNote.type ?? (startNote.metadata?.type as string) ?? 'star'
            });
            console.log('[3D Page] Added start node to graph:', startNote.id);
          } catch (noteErr) {
            console.error('[3D Page] Failed to fetch start note:', noteErr);
          }
        }
        
        // Immediately display initial data
        graphData = initialData;
        // Graph is shown immediately without waiting
        
        // Phase 2: Background load full graph (depth 3)
        isLoadingFull = true;
        console.log('[3D Page] Starting Phase 2 - loading full graph (depth 3)...');
        
        getGraphData(noteId, 3).then((fullData: GraphData) => {
          console.log('[3D Page] Phase 2 - Loaded full graph:', fullData.nodes.length, 'nodes');
          
          // Ensure start node is in the full data too
          const hasStartNodeFull = fullData.nodes.some((n: { id: string }) => n.id === noteId);
          if (!hasStartNodeFull) {
            // Start node will be added when we merge, but let's verify
            console.log('[3D Page] Start node not in full data, it will be merged from current data');
          }
          
          // Call addData on the Graph3D component to add new nodes incrementally
          if (graphRef && fullData.nodes.length > initialData.nodes.length) {
            graphRef.addData(fullData);
          }
          isLoadingFull = false;
        }).catch((e: unknown) => {
          console.error('[3D Page] Failed to load full graph:', e);
          isLoadingFull = false;
        });
      }
    } catch (e: any) {
      if (e.response?.status === 404) {
        error = 'Заметка не найдена';
        setTimeout(() => goto('/'), 3000);
      } else {
        error = 'Failed to load graph data';
      }
      console.error(e);
    }
  }
</script>

<div class="page">
  <div class="controls">
    <label class="toggle">
      <input type="checkbox" bind:checked={showFullGraph} />
      <span>Показать все заметки ({showFullGraph ? 'включено' : 'выключено'})</span>
    </label>
  </div>
  
  <!-- Stats with background loading indicator -->
  {#if graphData}
    <div class="stats-bar" data-testid="graph-stats">
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
      {#if isLoadingFull}
        <span class="loading-indicator" title="Loading more data...">
          <span class="mini-spinner"></span>
        </span>
      {/if}
    </div>
  {/if}
  
  {#if error}
    <div class="center error">{error}</div>
  {:else if graphData}
    <div class="graph-wrapper">
      <Graph3D 
        bind:this={graphRef} 
        data={graphData} 
        centerNodeId={currentNoteId}
        onNodeDoubleClick={(node: { id: string }) => console.log('Double-clicked node:', node.id)}
      />
    </div>
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

  .graph-wrapper {
    position: relative;
    width: 100%;
    height: 100%;
  }

  .loading-indicator {
    display: inline-flex;
    align-items: center;
    margin-left: 8px;
  }

  .mini-spinner {
    width: 14px;
    height: 14px;
    border: 2px solid rgba(255, 255, 255, 0.2);
    border-top-color: #88aaff;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }
</style>
