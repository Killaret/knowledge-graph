<script lang="ts">
  import { onMount } from "svelte";
  import GraphCanvas from "$widgets/graph-canvas/GraphCanvas.svelte";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

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

  const { nodes = [] as GraphNode[], links = [] as GraphLink[] } = $props<{
    nodes: GraphNode[];
    links: GraphLink[];
  }>();

  let isLoading = $state(true);

  onMount(() => {
    isLoading = false;
  });
</script>

{#if isLoading}
  <div class="graph-loading" role="status" aria-live="polite">
    <div class="spinner" aria-hidden="true"></div>
    <p>{t("smartGraph.loading")}</p>
  </div>
{:else}
  <div class="graph-wrapper graph-2d">
    <GraphCanvas {nodes} {links} />
    <div class="performance-hint" aria-hidden="true">
      {t("smartGraph.mode2D")}
    </div>
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
    to {
      transform: rotate(360deg);
    }
  }

  .graph-wrapper {
    width: 100%;
    height: 100%;
    position: relative;
  }

  .graph-2d {
    width: 100%;
    height: 100%;
  }

  .performance-hint {
    position: absolute;
    bottom: 10px;
    right: 10px;
    font-size: 0.7rem;
    color: rgba(255, 200, 100, 0.7);
    padding: 4px 8px;
    background: rgba(10, 26, 58, 0.6);
    border-radius: 4px;
    pointer-events: none;
    user-select: none;
  }
</style>
