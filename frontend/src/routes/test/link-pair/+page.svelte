<svelte:options runes={false} />

<script lang="ts">
  /**
   * Test page for link rendering between two nodes
   * Used for visual regression testing of different link types
   *
   * Query parameters:
   * - linkType: reference | dependency | related | custom
   * - sourceType: star | planet | comet | galaxy | asteroid (default: star)
   * - targetType: star | planet | comet | galaxy | asteroid (default: planet)
   */
  import { page } from "$app/stores";
  import { browser } from "$app/environment";
  import type { GraphNode, GraphLink } from "$shared/api/graph";

  // Dynamic import for browser-only component
  let GraphCanvas: typeof import("$components/organisms/GraphCanvas.svelte").default | undefined;

  if (browser) {
    import("$components/organisms/GraphCanvas.svelte").then((m) => {
      GraphCanvas = m.default;
    });
  }

  // Get query params (reactive)
  $: linkType = $page.url.searchParams.get("linkType") || "reference";
  $: sourceType = $page.url.searchParams.get("sourceType") || "star";
  $: targetType = $page.url.searchParams.get("targetType") || "planet";

  // Create two nodes with link (reactive)
  $: sourceNode = {
    id: "source-node",
    title: "Source",
    type: sourceType,
  } as GraphNode;

  $: targetNode = {
    id: "target-node",
    title: "Target",
    type: targetType,
  } as GraphNode;

  $: link = {
    source: "source-node",
    target: "target-node",
    weight: 0.8,
    link_type: linkType,
  } as GraphLink;

  $: nodes = [sourceNode, targetNode];
  $: links = [link];
</script>

<div class="test-container" data-testid="link-pair-test">
  <div class="info">
    <h1>Link test: {linkType}</h1>
    <p>{sourceType} → {targetType}</p>
  </div>

  <div class="canvas-wrapper">
    {#if GraphCanvas}
      <svelte:component this={GraphCanvas} {nodes} {links} />
    {:else}
      <div class="loading">Loading canvas...</div>
    {/if}
  </div>
</div>

<style>
  .test-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 20px;
    background: var(
      --gradient-cosmic-bg,
      radial-gradient(ellipse at 50% 100%, #1a0505 0%, #000 80%)
    );
    min-height: 100vh;
  }

  .info {
    color: white;
    text-align: center;
    margin-bottom: 20px;
  }

  .info h1 {
    font-size: 1.5rem;
    margin: 0 0 8px 0;
    text-transform: capitalize;
  }

  .info p {
    font-size: 0.875rem;
    opacity: 0.7;
    margin: 0;
    text-transform: capitalize;
  }

  .canvas-wrapper {
    width: 600px;
    height: 400px;
    border: 2px solid rgba(255, 255, 255, 0.2);
    border-radius: 8px;
    overflow: hidden;
    background: rgba(0, 0, 0, 0.5);
  }

  .loading {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
    color: white;
    font-size: 1.2rem;
  }
</style>
