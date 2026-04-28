<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import type { GraphData } from '$lib/api/graph';
  import type { Component } from 'svelte';

  const { data }: { data: GraphData } = $props();

  let Graph3DComponent: Component<{ data: GraphData }> | null = $state(null);
  let isLoading = $state(true);
  let loadError = $state<string | null>(null);
  let webglSupported = $state(true);

  // Check WebGL support
  function checkWebGL(): boolean {
    if (!browser) return false;
    try {
      const canvas = document.createElement('canvas');
      const gl = canvas.getContext('webgl') || canvas.getContext('experimental-webgl');
      return !!gl;
    } catch {
      return false;
    }
  }

  onMount(async () => {
    // Check WebGL support first
    webglSupported = checkWebGL();
    
    if (!webglSupported) {
      loadError = 'WebGL is not supported by your browser or device. Please use a modern browser with WebGL enabled.';
      isLoading = false;
      return;
    }

    try {
      // Dynamic import of Graph3D component
      const module = await import('./Graph3D.svelte');
      // Svelte 5 default export
      Graph3DComponent = (module as { default?: Component<{ data: GraphData }> }).default || null;
    } catch (e) {
      loadError = 'Failed to load 3D visualization';
      if (import.meta.env.DEV) {
        console.error('[LazyGraph3D] Error loading Graph3D:', e);
      }
    } finally {
      isLoading = false;
    }
  });
</script>

{#if isLoading}
  <div class="lazy-loading" role="status" aria-live="polite">
    <div class="spinner" aria-hidden="true"></div>
    <p>Loading 3D engine...</p>
  </div>
{:else if loadError}
  <div class="lazy-error" role="alert" aria-live="assertive">
    <span class="error-icon" aria-hidden="true">⚠️</span>
    <p>{loadError}</p>
  </div>
{:else if Graph3DComponent}
  <Graph3DComponent {data} />
{/if}

<style>
  .lazy-loading, .lazy-error {
    width: 100%;
    height: 100vh;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    background: #050510;
    color: white;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 3px solid rgba(255,255,255,0.3);
    border-top-color: #88aaff;
    border-radius: 50%;
    animation: spin 1s linear infinite;
    margin-bottom: 1rem;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .error-icon {
    font-size: 32px;
    margin-bottom: 0.5rem;
  }
</style>
