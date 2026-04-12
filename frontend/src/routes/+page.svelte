<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import Graph3D from '$lib/components/Graph3D.svelte';
  import NodeContextMenu from '$lib/components/NodeContextMenu.svelte';
  import { Skeleton } from 'flowbite-svelte';
  import { getGlobalGraphData, type GraphData } from '$lib/api/graph';
  import { getNotes } from '$lib/api/notes';

  // Graph data
  let graphData = $state<GraphData>({ nodes: [], links: [] });
  let loading = $state(true);
  let hasNotes = $state(false);
  
  // Context menu state
  let selectedNode = $state<any>(null);
  let contextMenuPosition = $state({ x: 0, y: 0 });
  let isContextMenuOpen = $state(false);

  onMount(async () => {
    try {
      // Check if there are any notes
      const notes = await getNotes();
      hasNotes = notes.length > 0;
      
      if (hasNotes) {
        // Try to load global graph
        graphData = await getGlobalGraphData();
      }
    } catch (e) {
      console.error('Failed to load graph:', e);
    } finally {
      loading = false;
    }
  });

  function handleNodeClick(node: any) {
    goto(`/notes/${node.id}`);
  }

  function handleNodeRightClick(node: any, event: MouseEvent) {
    selectedNode = node;
    contextMenuPosition = { x: event.clientX, y: event.clientY };
    isContextMenuOpen = true;
  }

  function handleCreateFirstNote() {
    goto('/notes/new');
  }

  function handleGoToNotes() {
    goto('/notes');
  }
</script>

<div class="home-graph-page">
  {#if loading}
    <div class="loading-skeleton">
      <Skeleton size="xl" class="w-full h-full" />
    </div>
  {:else if !hasNotes}
    <div class="welcome-screen">
      <div class="welcome-content">
        <span class="welcome-icon">🌌</span>
        <h1>Knowledge Constellation</h1>
        <p>Ваше созвездие знаний пусто. Создайте первую звезду, чтобы начать.</p>
        <div class="welcome-actions">
          <button class="btn-primary" onclick={handleCreateFirstNote}>
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="12" y1="5" x2="12" y2="19"/>
              <line x1="5" y1="12" x2="19" y2="12"/>
            </svg>
            Создать первую звезду
          </button>
        </div>
      </div>
    </div>
  {:else if graphData.nodes.length === 0}
    <div class="welcome-screen">
      <div class="welcome-content">
        <span class="welcome-icon">⭐</span>
        <h1>Одинокая звезда</h1>
        <p>У вас есть заметки, но связей пока нет. Выберите заметку, чтобы увидеть её связи.</p>
        <div class="welcome-actions">
          <button class="btn-primary" onclick={handleGoToNotes}>
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="8" y1="6" x2="21" y2="6"/>
              <line x1="8" y1="12" x2="21" y2="12"/>
              <line x1="8" y1="18" x2="21" y2="18"/>
              <line x1="3" y1="6" x2="3.01" y2="6"/>
              <line x1="3" y1="12" x2="3.01" y2="12"/>
              <line x1="3" y1="18" x2="3.01" y2="18"/>
            </svg>
            Все заметки
          </button>
          <button class="btn-secondary" onclick={handleCreateFirstNote}>
            <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="12" y1="5" x2="12" y2="19"/>
              <line x1="5" y1="12" x2="19" y2="12"/>
            </svg>
            Новая заметка
          </button>
        </div>
      </div>
    </div>
  {:else}
    <div class="graph-container">
      <Graph3D
        data={graphData}
        onNodeClick={handleNodeClick}
        onNodeRightClick={handleNodeRightClick}
      />
    </div>
    
    <NodeContextMenu
      note={selectedNode}
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
  .home-graph-page {
    height: 100vh;
    width: 100%;
  }

  .loading-skeleton {
    height: 100vh;
    padding: 2rem;
  }

  .welcome-screen {
    display: flex;
    justify-content: center;
    align-items: center;
    height: 100vh;
    background: linear-gradient(145deg, #0a1a3a, #020617);
  }

  .welcome-content {
    text-align: center;
    padding: 2rem;
  }

  .welcome-icon {
    font-size: 4rem;
    margin-bottom: 1rem;
    display: block;
  }

  .welcome-content h1 {
    font-size: 2rem;
    font-weight: 600;
    color: white;
    margin: 0 0 1rem;
    text-shadow: 0 0 20px rgba(255, 200, 100, 0.3);
  }

  .welcome-content p {
    font-size: 1.125rem;
    color: rgba(255, 255, 255, 0.7);
    max-width: 400px;
    margin: 0 auto 1.5rem;
    line-height: 1.6;
  }

  .welcome-actions {
    display: flex;
    gap: 1rem;
    justify-content: center;
    flex-wrap: wrap;
  }

  .btn-primary,
  .btn-secondary {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.875rem 1.5rem;
    border: none;
    border-radius: 0.75rem;
    font-size: 1rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s;
  }

  .btn-primary {
    background: rgba(99, 102, 241, 0.8);
    color: white;
  }

  .btn-primary:hover {
    background: rgba(99, 102, 241, 1);
    transform: translateY(-2px);
  }

  .btn-secondary {
    background: rgba(255, 255, 255, 0.1);
    color: rgba(255, 255, 255, 0.9);
    border: 1px solid rgba(255, 255, 255, 0.2);
  }

  .btn-secondary:hover {
    background: rgba(255, 255, 255, 0.2);
    transform: translateY(-2px);
  }

  .graph-container {
    width: 100%;
    height: 100vh;
  }

  @media (max-width: 768px) {
    .welcome-content h1 {
      font-size: 1.5rem;
    }

    .welcome-content p {
      font-size: 1rem;
    }
  }
</style>