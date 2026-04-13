<script lang="ts">
  import { onMount } from 'svelte';
  import type { GraphData } from '$lib/api/graph';
  import type { Component } from 'svelte';

  const { data }: { data: GraphData } = $props();

  let Graph3DComponent: Component<{ data: GraphData }> | null = $state(null);
  let isLoading = $state(true);
  let loadError = $state<string | null>(null);

  onMount(async () => {
    try {
      // Dynamic import of Graph3D component
      const module = await import('./Graph3D.svelte');
      // Svelte 5 default export
      Graph3DComponent = (module as { default?: Component<{ data: GraphData }> }).default || null;
    } catch (e) {
      loadError = 'Failed to load 3D visualization';
      console.error('Error loading Graph3D:', e);
    } finally {
      isLoading = false;
    }
  });
</script>

{#if isLoading}
  <div class="lazy-loading">
    <div class="spinner"></div>
    <p>Loading 3D engine...</p>
  </div>
{:else if loadError}
  <div class="lazy-error">
    <span class="error-icon">⚠️</span>
    <p>{loadError}</p>
  </div>
{:else if Graph3DComponent}
  {#key data.nodes.length + data.links.length}
    <Graph3DComponent {data} />
  {/key}
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