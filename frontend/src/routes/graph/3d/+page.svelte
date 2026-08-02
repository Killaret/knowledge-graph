<script lang="ts">
  import { onMount } from "svelte";
  import { browser } from "$app/environment";
  import { goto } from "$app/navigation";
  import BackButton from "$components/atoms/BackButton.svelte";
  import { createLayoutProvider, toRuntimeConfig } from "$features/graph-3d";
  import { type GraphData } from "$shared/api/graph";
  import { graphStore } from "$shared/stores/graph.svelte";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const runtimeConfig = toRuntimeConfig();
  const layoutProvider = createLayoutProvider(runtimeConfig);

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  let graphData: GraphData = $state({ nodes: [], links: [] });
  let loading = $state(true);
  let error = $state("");
  let Graph3DViewer: any = $state(null);

  onMount(async () => {
    if (!browser) return;
    try {
      [graphData] = await Promise.all([
        layoutProvider.load({}),
        import("$widgets/graph-3d-viewer/Graph3DViewer.svelte").then((mod) => {
          Graph3DViewer = mod.default;
        }),
      ]);
    } catch (e) {
      if (import.meta.env.DEV) {
        console.error("Failed to load 3D graph:", e);
      }
      error = t("graph.loadDataError");
    } finally {
      loading = false;
    }
  });
</script>

<div class="page">
  <div class="top-controls">
    <BackButton href="/graph" />
    <button class="home-btn" onclick={() => goto("/")}>{t("graph.goHome")}</button>
  </div>

  {#if loading}
    <div class="center">
      <div class="spinner"></div>
      <p>{t("page.loadingGraph")}</p>
    </div>
  {:else if error}
    <div class="center error">
      <p>{error}</p>
      <button onclick={() => goto("/")}>{t("graph.goHome")}</button>
    </div>
  {:else if graphData.nodes.length === 0}
    <div class="center">
      <h2>{t("graph3d.noDataTitle")}</h2>
      <p>{t("graph3d.noDataMessage")}</p>
    </div>
  {:else if Graph3DViewer}
    <Graph3DViewer
      nodes={graphData.nodes}
      links={graphData.links}
      selectedNodeId={graphStore.selectedNodeId}
      onNodeClick={(node: { id: string }) => (graphStore.selectedNodeId = node.id)}
    />
  {/if}
</div>

<style>
  .page {
    position: relative;
    width: 100%;
    height: 100vh;
    overflow: hidden;
    background: #050510;
    color: white;
  }

  .top-controls {
    position: absolute;
    top: 16px;
    left: 16px;
    z-index: 100;
    display: flex;
    gap: 8px;
  }

  .home-btn {
    padding: 8px 12px;
    border: 1px solid rgba(255, 255, 255, 0.2);
    background: rgba(5, 5, 16, 0.7);
    color: white;
    border-radius: 8px;
    cursor: pointer;
    backdrop-filter: blur(4px);
  }

  .home-btn:hover {
    background: rgba(136, 170, 255, 0.2);
  }

  .center {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 16px;
    text-align: center;
  }

  .error {
    color: #ff6b6b;
  }

  .error button {
    padding: 8px 16px;
    background: #3b82f6;
    color: white;
    border: none;
    border-radius: 6px;
    cursor: pointer;
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
    to {
      transform: rotate(360deg);
    }
  }
</style>
