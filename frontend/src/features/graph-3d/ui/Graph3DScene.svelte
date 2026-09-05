<script lang="ts">
  import { onMount, onDestroy } from "svelte";
  import { browser } from "$app/environment";
  import { Graph3DEngine } from "../lib/engine";
  import type { GraphNode, GraphLink, Graph3DCallbacks } from "../model/types";

  interface Props {
    nodes: GraphNode[];
    links: GraphLink[];
    centerNodeId?: string | null;
    selectedNodeId?: string | null;
    onNodeClick?: (node: { id: string; title: string; type?: string }) => void;
    onNodeDoubleClick?: (node: { id: string; title: string; type?: string }) => void;
    onReady?: () => void;
    onLoadingChange?: (loading: boolean) => void;
    onError?: (message: string) => void;
  }

  const {
    nodes,
    links,
    centerNodeId,
    selectedNodeId,
    onNodeClick,
    onNodeDoubleClick,
    onReady,
    onLoadingChange,
    onError,
  }: Props = $props();

  let container: HTMLDivElement;
  let engine: Graph3DEngine | null = null;

  // Vitest still relies on the env flag; Playwright runs in a real browser
  // where `process` does not exist, so the URL parameter is used instead.
  const stableRender =
    (typeof process !== "undefined" && process.env?.VITEST === "true") ||
    (browser && new URL(window.location.href).searchParams.get("stableRender") === "true");

  let sceneStable = $state(false);

  onMount(() => {
    if (!browser || !container) return;

    const callbacks: Graph3DCallbacks = {
      onNodeClick,
      onNodeDoubleClick,
      onReady: () => {
        sceneStable = true;
        onReady?.();
      },
      onLoadingChange,
      onError,
    };

    try {
      engine = new Graph3DEngine(container, callbacks, {
        disableAnimation: stableRender,
      });
      engine.setData(nodes, links);
      if (centerNodeId) engine.centerOnNode(centerNodeId);
      if (selectedNodeId) engine.setSelectedNodeId(selectedNodeId);

      container.addEventListener("click", handleClick);
      container.addEventListener("dblclick", handleDoubleClick);
    } catch (e) {
      if (import.meta.env.DEV) {
        console.error("[Graph3D] Initialization failed:", e);
      }
      onError?.("Failed to initialize 3D visualization");
    }
  });

  onDestroy(() => {
    container?.removeEventListener("click", handleClick);
    container?.removeEventListener("dblclick", handleDoubleClick);
    engine?.dispose();
    engine = null;
  });

  let lastNodes: GraphNode[] = [];
  let lastLinks: GraphLink[] = [];

  $effect(() => {
    const n = nodes;
    const l = links;
    if (engine && (n !== lastNodes || l !== lastLinks)) {
      engine.setData(n, l);
      lastNodes = n;
      lastLinks = l;
    }
  });

  $effect(() => {
    if (engine && centerNodeId) {
      engine.centerOnNode(centerNodeId);
    }
  });

  $effect(() => {
    if (engine && selectedNodeId) {
      engine.setSelectedNodeId(selectedNodeId);
    }
  });

  function handleResize() {
    engine?.handleResize();
  }

  function handleClick(event: MouseEvent) {
    engine?.handleClick(event);
  }

  function handleDoubleClick(event: MouseEvent) {
    engine?.handleDoubleClick(event);
  }
</script>

<svelte:window onresize={handleResize} />

<div
  bind:this={container}
  class="graph-3d-container"
  data-testid="graph-3d-scene"
  data-test-stable={sceneStable ? "true" : "false"}
  role="img"
  aria-label="3D graph"
></div>

<style>
  .graph-3d-container {
    width: 100%;
    height: 100%;
    position: relative;
    overflow: hidden;
    background: #050510;
    cursor: grab;
  }

  .graph-3d-container:active {
    cursor: grabbing;
  }
</style>
