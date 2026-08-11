<script lang="ts">
  import { onMount } from "svelte";
  import { page } from "$app/stores";
  import { goto } from "$app/navigation";
  import SmartGraph from "$widgets/graph-canvas/SmartGraph.svelte";
  import { getGraphData } from "$shared/api/graph";
  import { GraphPageShell } from "$widgets/graph-page";
  import { CelestialBody } from "$entities";
  import type { GraphNode, GraphLink } from "$shared/api/graph";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

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

  let nodes: GraphNode[] = $state([]);
  let links: GraphLink[] = $state([]);
  let loading = $state(true);
  let error = $state("");

  function getRouteId(): string {
    const id = $page.params.id;
    if (!id) throw new Error("Missing route parameter: id");
    return id;
  }

  onMount(async () => {
    try {
      const id = getRouteId();
      const data = await getGraphData(id);
      // Map nodes with defaults for SmartGraph compatibility
      nodes = data.nodes.map((n: GraphNode) => ({
        ...n,
        type: n.type ?? "default",
        size: n.size ?? 5,
      }));
      links = data.links.map((l: GraphLink) => ({
        ...l,
        weight: l.weight ?? 1,
      }));
    } catch (e) {
      error = t("graph.loadDataError");
      if (import.meta.env.DEV) {
        console.error(e);
      }
    } finally {
      loading = false;
    }
  });
</script>

<GraphPageShell
  view="graph"
  layoutProvider="d3"
  searchQuery=""
  selectedType="all"
  typeFilters={graphTypeFilters}
  {nodes}
  {links}
  onSearch={() => {}}
  onFilter={() => {}}
  onToggleView={(view) => {
    if (view === "3d") goto(`/graph/3d/${$page.params.id}`);
    else if (view === "list") goto("/");
  }}
  onSignIn={() => goto("/auth/login")}
  onRegister={() => goto("/auth/register")}
>
  <div class="graph-page">
    {#if loading}
      <div class="loading-state">
        <div class="spinner"></div>
        <p>{t("graph.loadingConstellation")}</p>
      </div>
    {:else if error}
      <div class="error-state">
        <p class="error">{error}</p>
      </div>
    {:else}
      <div class="graph-container graph-3d-container" data-testid="graph-container">
        <SmartGraph {nodes} {links} />
      </div>
    {/if}
  </div>
</GraphPageShell>

<style>
  .graph-page {
    position: relative;
    width: 100%;
    height: 100%;
    background: transparent;
    color: #e0e0e0;
  }

  .graph-container {
    position: absolute;
    inset: 0;
    overflow: hidden;
  }

  .loading-state,
  .error-state {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 1rem;
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
    to {
      transform: rotate(360deg);
    }
  }

  .error {
    color: #ff6b6b;
    padding: 1rem;
    background: rgba(255, 100, 100, 0.1);
    border-radius: 8px;
    border: 1px solid rgba(255, 100, 100, 0.3);
  }
</style>
