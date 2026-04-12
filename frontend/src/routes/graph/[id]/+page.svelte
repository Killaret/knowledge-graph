<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import Graph3D from '$lib/components/Graph3D.svelte';
  import { getGraphData } from '$lib/api/graph';
  import BackButton from '$lib/components/BackButton.svelte';
  import NodeContextMenu from '$lib/components/NodeContextMenu.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import { Skeleton } from 'flowbite-svelte';

  let graphData = $state<{ nodes: any[], links: any[] }>({ nodes: [], links: [] });
  let loading = $state(true);
  let error = $state('');

  // Context menu state
  let selectedNote = $state<any>(null);
  let contextMenuPosition = $state({ x: 0, y: 0 });
  let isContextMenuOpen = $state(false);

  function getRouteId(): string {
    const id = $page.params.id;
    if (!id) throw new Error('Missing route parameter: id');
    return id;
  }

  onMount(async () => {
    try {
      const id = getRouteId();
      const data = await getGraphData(id);
      graphData = { nodes: data.nodes, links: data.links };
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
    <div class="loading-skeleton">
      <div class="skeleton-header">
        <Skeleton size="lg" class="w-64 mb-4" />
      </div>
      <div class="skeleton-canvas">
        <Skeleton size="xl" class="w-full h-96" />
      </div>
    </div>
  {:else if error}
    <div class="error-state">
      <p class="error">{error}</p>
    </div>
  {:else}
    <div class="graph-container">
      {#if graphData.nodes.length <= 1}
        <EmptyState
          icon="⭐"
          title="Это одинокая звезда"
          message="Создайте связи, чтобы увидеть созвездие"
          actionText="Создать связь"
          onAction={() => goto('/notes/new')}
        />
      {:else}
        <Graph3D 
          data={graphData} 
          onNodeClick={(node) => {
            goto(`/notes/${node.id}`);
          }}
          onNodeRightClick={(node, event) => {
            selectedNote = node;
            contextMenuPosition = { x: event.clientX, y: event.clientY };
            isContextMenuOpen = true;
          }}
        />
      {/if}
    </div>
    
    <NodeContextMenu
      note={selectedNote}
      isOpen={isContextMenuOpen}
      position={contextMenuPosition}
      onClose={() => isContextMenuOpen = false}
      onEdit={(id) => goto(`/notes/${id}/edit`)}
      onCreateLink={(id) => console.log('Create link for', id)}
      onCopyLink={(id) => console.log('Copy link for', id)}
      onDelete={(id) => {
        graphData = {
          ...graphData,
          nodes: graphData.nodes.filter(n => n.id !== id)
        };
      }}
      onChangeType={(id, type) => console.log('Change type', id, type)}
    />
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

  .loading-skeleton {
    flex: 1;
    padding: 2rem;
    max-width: 1200px;
    margin: 0 auto;
    width: 100%;
  }

  .skeleton-header {
    margin-bottom: 2rem;
  }

  .skeleton-canvas {
    height: 500px;
  }

  .error-state {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 1rem;
  }

  .error {
    color: #ff6b6b;
    padding: 1rem;
    background: rgba(255, 100, 100, 0.1);
    border-radius: 8px;
    border: 1px solid rgba(255, 100, 100, 0.3);
  }
</style>