<script lang="ts">
  import { onMount } from "svelte";
  import { browser } from "$app/environment";
  import { fade } from "svelte/transition";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
  import type { GraphNode, GraphLink } from "$shared/api/graph";
  import type { Component } from "svelte";

  interface Props {
    nodes: GraphNode[];
    links: GraphLink[];
    centerNodeId?: string | null;
    selectedNodeId?: string | null;
    onNodeClick?: (node: { id: string; title: string; type?: string }) => void;
    onNodeDoubleClick?: (node: { id: string; title: string; type?: string }) => void;
  }

  const { nodes, links, centerNodeId, selectedNodeId, onNodeClick, onNodeDoubleClick }: Props =
    $props();

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string | number>) =>
    formatMessage(key, locale, params);

  let Graph3DComponent: Component<{
    nodes: GraphNode[];
    links: GraphLink[];
    centerNodeId?: string | null;
    selectedNodeId?: string | null;
    onNodeClick?: (node: { id: string; title: string; type?: string }) => void;
    onNodeDoubleClick?: (node: { id: string; title: string; type?: string }) => void;
    onReady?: () => void;
    onLoadingChange?: (loading: boolean) => void;
    onError?: (message: string) => void;
  }> | null = $state(null);

  let isLoading = $state(true);
  let error = $state<string | null>(null);

  onMount(async () => {
    if (!browser) return;
    try {
      const module = await import("$components/organisms/Graph3D.svelte");
      Graph3DComponent = module.default;
    } catch (e) {
      if (import.meta.env.DEV) {
        console.error("[GraphCanvas3D] Failed to load 3D component:", e);
      }
      error = t("graph3d.errorLoad");
      isLoading = false;
    }
  });

  function handleReady() {
    isLoading = false;
  }

  function handleLoadingChange(loading: boolean) {
    isLoading = loading;
  }

  function handleError(message: string) {
    error = message;
    isLoading = false;
  }
</script>

<div class="graph-canvas-3d" data-testid="graph-canvas-3d">
  {#if error}
    <div class="error-overlay" data-testid="graph-3d-error">
      <div class="error-content">
        <span class="error-icon">🌌</span>
        <h2>{error}</h2>
        <p>{t("graph3d.errorHint")}</p>
      </div>
    </div>
  {:else if isLoading}
    <div class="loading-overlay" transition:fade={{ duration: 400 }} data-testid="graph-3d-loading">
      <div class="spinner"></div>
      <p class="loading-title">{t("graph3d.loadingTitle")}</p>
      <p class="loading-subtitle">{t("graph3d.loadingSubtitle")}</p>
    </div>
  {/if}

  {#if Graph3DComponent}
    <Graph3DComponent
      {nodes}
      {links}
      {centerNodeId}
      {selectedNodeId}
      {onNodeClick}
      {onNodeDoubleClick}
      onReady={handleReady}
      onLoadingChange={handleLoadingChange}
      onError={handleError}
    />
  {/if}
</div>

<style>
  .graph-canvas-3d {
    width: 100%;
    height: 100%;
    position: relative;
    overflow: hidden;
    background: #050510;
  }

  .loading-overlay,
  .error-overlay {
    position: absolute;
    inset: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 1rem;
    color: white;
    background: rgba(5, 5, 16, 0.9);
    z-index: 10;
    backdrop-filter: blur(4px);
    pointer-events: none;
  }

  .loading-title {
    font-size: 1.5rem;
    font-weight: 500;
    color: rgba(136, 170, 255, 0.9);
    text-align: center;
  }

  .loading-subtitle {
    color: rgba(148, 163, 184, 0.8);
    font-style: italic;
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

  .error-content {
    text-align: center;
    padding: 2rem;
  }

  .error-icon {
    font-size: 48px;
    display: block;
    margin-bottom: 1rem;
  }

  .error-content h2 {
    color: #ff6b6b;
    margin-bottom: 0.5rem;
  }

  .error-content p {
    color: #94a3b8;
  }
</style>
