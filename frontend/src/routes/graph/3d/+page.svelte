<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import { getFullGraphData } from '$lib/api/graph';
  import LazyGraph3D from '$lib/components/LazyGraph3D.svelte';
  import type { GraphData } from '$lib/api/graph';

  let graphData: GraphData | null = $state(null);
  let loading = $state(true);
  let error = $state('');

  onMount(async () => {
    if (!browser) return;
    try {
      graphData = await getFullGraphData();
    } catch (e) {
      error = 'Failed to load full graph';
      console.error(e);
    } finally {
      loading = false;
    }
  });
</script>

<div class="page">
  {#if loading}
    <div class="center">
      <div class="spinner"></div>
      <p>Loading full knowledge graph...</p>
    </div>
  {:else if error}
    <div class="center error">{error}</div>
  {:else if graphData}
    <div class="stats-bar">
      <span><strong>{graphData.nodes.length}</strong> nodes</span>
      <span><strong>{graphData.links.length}</strong> links</span>
    </div>
    <LazyGraph3D data={graphData} />
  {/if}
</div>

<style>
  .page {
    position: relative;
    width: 100%;
    height: 100vh;
    overflow: hidden;
    background: #050510;
  }

  .center {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: white;
    gap: 16px;
  }

  .error {
    color: #ff6666;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 3px solid rgba(255, 255, 255, 0.3);
    border-top-color: #88aaff;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .stats-bar {
    position: absolute;
    top: 20px;
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

  .stats-bar strong {
    color: #88aaff;
    font-weight: 600;
  }
</style>
