<script lang="ts">
  import { onMount } from "svelte";
  import { browser } from "$app/environment";
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import { GraphPageShell } from "$widgets/graph-page";
  import { CelestialBody } from "$entities";
  import { createLayoutProvider, toRuntimeConfig } from "$features/graph-3d";
  import { type GraphData } from "$shared/api/graph";
  import type { Component } from "svelte";
  import type { Props as Graph3DViewerProps } from "$widgets/graph-3d-viewer/Graph3DViewer.svelte";
  import { graphStore } from "$shared/stores/graph.svelte";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const runtimeConfig = toRuntimeConfig();
  const layoutProvider = createLayoutProvider(runtimeConfig);

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  const graphTypeFilters = [
    { id: "all", label: t("filter.all"), emoji: "🌌", description: t("filter.all.description") },
    ...[
      "star",
      "planet",
      "moon",
      "comet",
      "galaxy",
      "nebula",
      "asteroid",
      "satellite",
      "blackhole",
      "dust",
      "unknown",
    ].map((id) => {
      const body = CelestialBody.fromString(id);
      return { id, label: body.label, emoji: body.emoji, description: body.description };
    }),
  ];

  let graphData: GraphData = $state({ nodes: [], links: [] });
  let loading = $state(true);
  let error = $state("");
  let Graph3DViewer: Component<Graph3DViewerProps> | null = $state(null);

  function getRouteId(): string {
    const id = $page.params.id;
    if (!id) throw new Error("Missing route parameter: id");
    return id;
  }

  onMount(async () => {
    if (!browser) return;
    try {
      const id = getRouteId();
      graphStore.selectedNodeId = id;
      [graphData] = await Promise.all([
        layoutProvider.load({ noteId: id, depth: 3 }),
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

<GraphPageShell
  view="3d"
  layoutProvider={runtimeConfig.layoutProvider}
  searchQuery=""
  selectedType="all"
  typeFilters={graphTypeFilters}
  notes={graphData.nodes.map((n) => ({ id: n.id, title: n.title, type: n.type }))}
  nodeCount={graphData.nodes.length}
  linkCount={graphData.links.length}
  selectedNodeId={graphStore.selectedNodeId}
  onSearch={() => {}}
  onFilter={() => {}}
  onToggleView={(view) => {
    if (view === "graph") goto(`/graph/${$page.params.id}`);
    else if (view === "list") goto("/");
  }}
  onNodeSelect={(id) => (graphStore.selectedNodeId = id)}
  onSignIn={() => goto("/auth/login")}
  onRegister={() => goto("/auth/register")}
>
  <div class="page">
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
        centerNodeId={$page.params.id}
        selectedNodeId={graphStore.selectedNodeId}
        onNodeClick={(node: { id: string }) => (graphStore.selectedNodeId = node.id)}
      />
    {/if}
  </div>
</GraphPageShell>

<style>
  .page {
    position: relative;
    width: 100%;
    height: 100%;
    overflow: hidden;
    background: transparent;
    color: white;
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
