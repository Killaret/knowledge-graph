<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import { detectDeviceCapabilities, shouldUse3D } from '$lib/utils/deviceCapabilities';
  import GraphCanvas from './GraphCanvas.svelte';

  interface GraphNode {
    id: string;
    title: string;
    type: string;
  }

  interface GraphLink {
    source: string;
    target: string;
    weight: number;
  }

  const { 
    nodes = [] as GraphNode[],
    links = [] as GraphLink[]
  } = $props<{
    nodes: GraphNode[];
    links: GraphLink[];
  }>();

  let use3D = $state(false); // Start with false to avoid flash
  let isLoading = $state(true);
  let _loadError = $state<string | null>(null);
  let Graph3DComponent: any = $state(null);

  onMount(async () => {
    if (!browser) {
      isLoading = false;
      return;
    }

    // Detect device capabilities
    const deviceCaps = detectDeviceCapabilities();
    
    // Check if we should use 3D
    let shouldRender3D = shouldUse3D(deviceCaps);

    // Allow forcing 3D via query param (useful for E2E/testing)
    try {
      const params = new URLSearchParams(window.location.search);
      if (params.get('force3d') === '1') shouldRender3D = true;
    } catch { /* ignore in SSR */ }
    
    use3D = shouldRender3D;

    // Dynamically import 3D component if needed
    if (use3D) {
      try {
        const module = await import('./Graph3D.svelte');
        Graph3DComponent = module.default;
      } catch (e) {
        console.warn('Failed to load 3D graph component, falling back to 2D:', e);
        _loadError = 'Failed to load 3D visualization';
        use3D = false;
      }
    }

    isLoading = false;
  });
</script>

{#if isLoading}
  <div class="graph-loading">
    <div class="spinner"></div>
    <p>Loading visualization...</p>
  </div>
{:else if use3D && Graph3DComponent}
  <div class="graph-wrapper graph-3d">
    <Graph3DComponent data={{ nodes, links }} />
    <div class="performance-hint">3D Mode</div>
  </div>
{:else}
  <div class="graph-wrapper graph-2d">
    <GraphCanvas {nodes} {links} />
    <div class="performance-hint">2D Mode (optimized)</div>
  </div>
{/if}

{#if _loadError}
  <div class="graph-error-banner" role="alert">{_loadError}</div>
{/if}

<style>
  .graph-loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    gap: 1rem;
    color: #88aacc;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 3px solid rgba(100, 150, 200, 0.3);
    border-top-color: #ffdd88;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .graph-wrapper {
    width: 100%;
    height: 100%;
    position: relative;
  }

  .graph-3d,
  .graph-2d {
    width: 100%;
    height: 100%;
  }

  .performance-hint {
    position: absolute;
    bottom: 10px;
    right: 10px;
    font-size: 0.7rem;
    color: rgba(136, 170, 204, 0.5);
    padding: 4px 8px;
    background: rgba(10, 26, 58, 0.6);
    border-radius: 4px;
    pointer-events: none;
    user-select: none;
  }

  .graph-2d .performance-hint {
    color: rgba(255, 200, 100, 0.7);
  }
</style>
