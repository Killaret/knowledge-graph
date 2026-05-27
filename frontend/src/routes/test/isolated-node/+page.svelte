<svelte:options runes={false} />

<script lang="ts">
  /**
   * Test page for isolated node rendering
   * Used for visual regression testing of individual celestial body types
   */
  import { page } from '$app/stores';
  import { browser } from '$app/environment';
  import type { GraphNode, GraphLink } from '$lib/api/graph';
  
  // Dynamic import for browser-only component
  let GraphCanvas: any;
  
  if (browser) {
    import('$lib/components/GraphCanvas.svelte').then(m => {
      GraphCanvas = m.default;
    });
    // Debug: log variation parameters for the test node in browser
    import('$lib/utils/variation').then(mod => {
      const v = mod.getVariation('test-node', nodeType);
      console.log('[DEBUG][isolated-node] variation for test-node', v);
    }).catch(()=>{});
  }
  
  // Get type from query param
  let nodeType = 'star';
  $: nodeType = $page.url.searchParams.get('type') || 'star';
  
  // Create isolated node (reactive)
  $: isolatedNode = {
    id: 'test-node',
    title: `Test ${nodeType.charAt(0).toUpperCase() + nodeType.slice(1)}`,
    type: nodeType
  } as GraphNode;
  
  $: nodes = [isolatedNode];
  $: links = [] as GraphLink[];
</script>

<div class="test-container" data-testid="isolated-node-test">
  <div class="info">
    <h1>Isolated {nodeType} node</h1>
  </div>
  
  <div class="canvas-wrapper">
    {#if GraphCanvas}
      <svelte:component this={GraphCanvas} {nodes} {links} />
    {:else}
      <div class="loading">Loading canvas...</div>
    {/if}
  </div>
</div>

<style>
  .test-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 20px;
    background: var(--gradient-cosmic-bg, radial-gradient(ellipse at 50% 100%, #1a0505 0%, #000 80%));
    min-height: 100vh;
  }
  
  .info {
    color: white;
    text-align: center;
    margin-bottom: 20px;
  }
  
  .info h1 {
    font-size: 1.5rem;
    margin: 0 0 8px 0;
    text-transform: capitalize;
  }
  
  .info p {
    font-size: 0.875rem;
    opacity: 0.7;
    margin: 0;
  }
  
  .canvas-wrapper {
    width: 800px;
    height: 600px;
    border: 2px solid rgba(255, 255, 255, 0.2);
    border-radius: 8px;
    overflow: hidden;
    background: rgba(0, 0, 0, 0.5);
  }
  
  .loading {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
    color: white;
    font-size: 1.2rem;
  }
</style>
