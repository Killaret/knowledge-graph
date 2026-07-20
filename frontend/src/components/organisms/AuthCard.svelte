<script lang="ts">
  import { onMount } from "svelte";
  import { fly, fade } from "svelte/transition";
  import { quintOut } from "svelte/easing";
  import GalaxyIcon from "$components/atoms/GalaxyIcon.svelte";
  import GraphCanvas from "$components/organisms/GraphCanvas.svelte";
  import { getGraphWithPreload } from "$shared/hooks/usePreloadedData";
  import type { GraphData } from "$shared/api/graph";

  interface Props {
    title: string;
    subtitle?: string;
    showIcon?: boolean;
    children?: import("svelte").Snippet;
  }

  const { title, subtitle = "", showIcon = true, children }: Props = $props();

  let graphData = $state<GraphData>({ nodes: [], links: [] });

  onMount(() => {
    getGraphWithPreload(100)
      .then((data) => {
        graphData = data;
      })
      .catch((err) => {
        if (import.meta.env.DEV) {
          console.warn("[AuthCard] Failed to load graph background:", err);
        }
      });
  });
</script>

<div class="auth-page">
  <div class="graph-background cosmic-background">
    <GraphCanvas
      nodes={graphData.nodes}
      links={graphData.links}
      readonly={true}
      disableVariation={true}
      className="cosmic-background"
    />
  </div>

  <div
    class="auth-container"
    in:fly={{ y: 30, duration: 800, easing: quintOut }}
    out:fade={{ duration: 300 }}
  >
    <div class="logo-section">
      {#if showIcon}
        <GalaxyIcon size={64} class="logo-icon" />
      {/if}
      <h1 class="title">{title}</h1>
      {#if subtitle}
        <p class="subtitle">{subtitle}</p>
      {/if}
    </div>

    <div class="card">
      {@render children?.()}
    </div>
  </div>
</div>

<style>
  .auth-page {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    position: relative;
    padding: 2rem;
    overflow: hidden;
  }

  .graph-background {
    position: absolute;
    inset: 0;
    z-index: 0;
    pointer-events: none;
  }

  .auth-container {
    position: relative;
    z-index: 10;
    width: 100%;
    max-width: 480px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2rem;
  }

  .logo-section {
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.75rem;
  }

  .logo-section :global(.logo-icon) {
    margin-bottom: 0.5rem;
  }

  .title {
    margin: 0;
    font-size: 1.75rem;
    font-weight: 700;
    color: var(--color-text-dark, #e0e0e0);
    text-shadow: 0 0 20px rgba(255, 204, 0, 0.3);
    letter-spacing: 0.05em;
  }

  .subtitle {
    margin: 0;
    font-size: 1rem;
    color: var(--color-text-secondary, #94a3b8);
    max-width: 300px;
  }

  .card {
    width: 100%;
    padding: 2rem;
    background: rgba(10, 10, 26, 0.7);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border-radius: 16px;
    border: 1px solid rgba(255, 255, 255, 0.1);
    box-shadow:
      0 0 0 1px rgba(255, 204, 0, 0.1),
      0 8px 32px rgba(0, 0, 0, 0.4),
      0 0 40px rgba(64, 169, 255, 0.1);
    transition:
      box-shadow 0.3s ease,
      border-color 0.3s ease;
  }

  .card:hover {
    box-shadow:
      0 0 0 1px rgba(255, 204, 0, 0.2),
      0 12px 40px rgba(0, 0, 0, 0.5),
      0 0 60px rgba(64, 169, 255, 0.15);
    border-color: rgba(255, 204, 0, 0.2);
  }

  @media (max-width: 640px) {
    .auth-page {
      padding: 1rem;
    }

    .title {
      font-size: 1.5rem;
    }

    .card {
      padding: 1.5rem;
    }
  }
</style>
