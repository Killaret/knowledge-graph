<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import { detectDeviceCapabilities, shouldUse3D } from '$lib/utils/deviceCapabilities';
  import GraphCanvas from './GraphCanvas.svelte';

  interface GraphNode {
    id: string;
    title: string;
    type?: string;
    size?: number;
  }

  interface GraphLink {
    source: string;
    target: string;
    weight?: number;
    link_type?: string;
  }

  const { 
    nodes = [] as GraphNode[],
    links = [] as GraphLink[]
  } = $props<{
    nodes: GraphNode[];
    links: GraphLink[];
  }>();

  let use3D = $state(false);
  let isLoading = $state(true);
  let Graph3DComponent: any = $state(null);
  let webglSupported = $state(true);
  // Allow forcing 3D mode via URL param ?force3d=1 (useful for debugging/CI)
  let isForce3D = $state(false);

  $effect(() => {
    if (browser) {
      isForce3D = new URLSearchParams(window.location.search).get('force3d') === '1';
    }
  });

  onMount(async () => {
    if (!browser) {
      isLoading = false;
      return;
    }

    // Detect device capabilities
    const deviceCaps = detectDeviceCapabilities();
    
    // Check if we should use 3D
    const shouldRender3D = shouldUse3D(deviceCaps);
    
    // Also check for WebGL support
    let hasWebGL = false;
    try {
      const canvas = document.createElement('canvas');
      const gl = canvas.getContext('webgl') || canvas.getContext('experimental-webgl');
      hasWebGL = !!gl;
    } catch {
      hasWebGL = false;
    }
    webglSupported = hasWebGL;

    use3D = isForce3D ? true : (shouldRender3D && hasWebGL);

    // Dynamically import 3D component if needed
    if (use3D) {
      try {
        const module = await import('./Graph3D.svelte');
        Graph3DComponent = module.default;
      } catch (e) {
        console.warn('Failed to load 3D graph component, falling back to 2D:', e);
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
{:else if !webglSupported}
  <div class="graph-wrapper graph-2d">
    <div class="webgl-fallback">
      <p>3D visualization requires WebGL, which is not supported by your browser or device.</p>
      <p>Showing 2D view instead.</p>
    </div>
    <GraphCanvas {nodes} {links} />
    <div class="performance-hint">2D Mode (WebGL not available)</div>
  </div>
{:else}
  <div class="graph-wrapper graph-2d">
    <GraphCanvas {nodes} {links} />
    <div class="performance-hint">2D Mode (optimized)</div>
  </div>
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

  .webgl-fallback {
    position: absolute;
    top: 20px;
    left: 50%;
    transform: translateX(-50%);
    background: rgba(255, 100, 100, 0.9);
    color: white;
    padding: 12px 20px;
    border-radius: 8px;
    text-align: center;
    z-index: 100;
    max-width: 80%;
  }

  .webgl-fallback p {
    margin: 0;
    font-size: 0.9rem;
  }

  .webgl-fallback p:first-child {
    font-weight: bold;
    margin-bottom: 4px;
  }
</style>
