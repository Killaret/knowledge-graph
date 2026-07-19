<script lang="ts">
  import GraphTooltip from "$components/molecules/GraphTooltip.svelte";
  import LinkTooltip from "$components/molecules/LinkTooltip.svelte";
  import type { HotkeysState } from "$features/graph-interaction/hotkeys";
  import { CelestialBody } from "$shared/lib/domain";

  const {
    canvas,
    nodes,
    links,
    loading = false,
    hoveredNodeId = null,
    hoveredLink = null,
    tooltipPosition = { x: 0, y: 0 },
    duplicateWarning = null,
    focusMode = false,
    showUndoToast = false,
    undoToastStage = "done",
    hotkeysState,
    onCloseSearch,
    onRestoreDeletedNode,
    onCancelUndo,
    onUpdateSearch,
    onLinkEdit,
    onLinkDelete,
  }: {
    canvas: HTMLCanvasElement | null;
    nodes: Array<{ id: string; title: string; type?: string }>;
    links: Array<{
      source: string;
      target: string;
      link_type?: string;
      weight?: number;
      source_type?: string;
    }>;
    loading?: boolean;
    hoveredNodeId?: string | null;
    hoveredLink?: {
      source: string;
      target: string;
      link_type: string;
      weight: number;
      source_type: string;
    } | null;
    tooltipPosition?: { x: number; y: number };
    duplicateWarning?: {
      message: string;
      x: number;
      y: number;
      linkId: string;
    } | null;
    focusMode?: boolean;
    showUndoToast?: boolean;
    undoToastStage?: "done" | "restore";
    hotkeysState: HotkeysState;
    onCloseSearch?: () => void;
    onRestoreDeletedNode?: () => void;
    onCancelUndo?: () => void;
    onUpdateSearch?: () => void;
    onLinkEdit?: () => void;
    onLinkDelete?: () => void;
  } = $props();

  let graphTooltip: GraphTooltip | null = $state(null);
  let searchInput: HTMLInputElement | null = $state(null);

  function getNodeEmoji(type?: string): string {
    return CelestialBody.fromString(type).emoji;
  }

  $effect(() => {
    if (!graphTooltip || !canvas) return;
    if (hoveredNodeId) {
      const node = nodes.find((n) => n.id === hoveredNodeId);
      if (node) {
        graphTooltip.showNodeTooltip(
          node.title,
          node.type || "Unknown",
          getNodeEmoji(node.type),
        );
      } else {
        graphTooltip.hide();
      }
    } else {
      graphTooltip.hide();
    }
  });

  $effect(() => {
    if (hotkeysState.showSearchBox && searchInput) {
      requestAnimationFrame(() => searchInput?.focus());
    }
  });
</script>

{#if hoveredNodeId && canvas}
  <GraphTooltip bind:this={graphTooltip} target={canvas} />
{/if}

{#if hoveredLink}
  {@const sourceNode = nodes.find((n) => n.id === hoveredLink.source)}
  {@const targetNode = nodes.find((n) => n.id === hoveredLink.target)}
  <LinkTooltip
    visible={true}
    x={tooltipPosition.x}
    y={tooltipPosition.y}
    linkType={hoveredLink.link_type}
    weight={hoveredLink.weight}
    sourceType={hoveredLink.source_type}
    sourceTitle={sourceNode?.title || "Unknown"}
    targetTitle={targetNode?.title || "Unknown"}
    onEdit={onLinkEdit}
    onDelete={onLinkDelete}
  />
{/if}

{#if duplicateWarning}
  <div
    class="duplicate-warning"
    style="position: absolute; left: {duplicateWarning.x}px; top: {duplicateWarning.y}px; background: rgba(255, 204, 0, 0.95); color: #000; padding: 6px 10px; border-radius: 4px; font-size: 12px; font-weight: 600; z-index: 100; pointer-events: none; box-shadow: 0 2px 8px rgba(0,0,0,0.3);"
  >
    {duplicateWarning.message}
  </div>
{/if}

{#if focusMode}
  <div
    class="focus-mode-indicator"
    style="position: absolute; top: 16px; right: 16px; background: rgba(0,0,0,0.6); border: 1px solid rgba(255,255,255,0.2); border-radius: 6px; padding: 6px; color: white; z-index: 50; display: flex; align-items: center; gap: 4px; font-size: 12px;"
    title="Focus mode is active. Press Esc to restore effects."
  >
    <svg
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
    >
      <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
      <circle cx="12" cy="12" r="3" />
    </svg>
    Focus
  </div>
{/if}

{#if showUndoToast}
  <div
    class="undo-toast"
    style="position: fixed; bottom: 20px; right: 20px; background: rgba(10, 26, 58, 0.95); border: 1px solid rgba(138, 43, 226, 0.5); border-radius: 12px; padding: 16px; min-width: 300px; max-width: 400px; box-shadow: 0 4px 20px rgba(0,0,0,0.5); z-index: 1000; display: flex; align-items: center; justify-content: space-between; gap: 12px; color: white; animation: slide-up 0.3s ease;"
  >
    {#if undoToastStage === "done"}
      <span style="font-size: 14px; color: rgba(255,255,255,0.9);"
        >Note deleted.</span
      >
    {:else}
      <span style="font-size: 14px;">Note deleted.</span>
      <div style="display: flex; gap: 8px; align-items: center;">
        <button
          onclick={onRestoreDeletedNode}
          style="padding: 6px 12px; border: none; border-radius: 4px; background: #8b5cf6; color: white; cursor: pointer; font-size: 13px; font-weight: 600;"
        >
          Restore
        </button>
        <button
          onclick={onCancelUndo}
          style="background: none; border: none; color: rgba(255,255,255,0.6); cursor: pointer; font-size: 18px; line-height: 1;"
          aria-label="Close"
        >
          ×
        </button>
      </div>
    {/if}
  </div>
{/if}

{#if hotkeysState.showSearchBox}
  <div
    class="search-box"
    data-testid="search-box"
    style="position: absolute; top: 16px; left: 50%; transform: translateX(-50%); background: rgba(10, 26, 58, 0.9); border: 1px solid rgba(138, 43, 226, 0.5); border-radius: 8px; padding: 8px 12px; display: flex; align-items: center; gap: 8px; z-index: 100; box-shadow: 0 4px 20px rgba(0,0,0,0.4);"
  >
    <span style="color: rgba(255,255,255,0.6); font-size: 14px;">🔍</span>
    <input
      bind:this={searchInput}
      bind:value={hotkeysState.searchQuery}
      oninput={onUpdateSearch}
      placeholder="Search nodes..."
      style="background: transparent; border: none; color: white; outline: none; min-width: 200px; font-size: 14px;"
    />
    {#if hotkeysState.searchMatchIds.length > 0}
      <span style="color: rgba(255,255,255,0.6); font-size: 12px;"
        >{hotkeysState.searchCurrentIndex + 1}/{hotkeysState.searchMatchIds
          .length}</span
      >
    {/if}
    <button
      onclick={onCloseSearch}
      style="background: none; border: none; color: rgba(255,255,255,0.6); cursor: pointer; font-size: 16px;"
      >×</button
    >
  </div>
{/if}

{#if hotkeysState.showHelpTooltip && hotkeysState.helpTooltipPosition.x === -1}
  <div
    class="help-tooltip"
    style="position: fixed; bottom: 32px; left: 50%; transform: translateX(-50%); background: rgba(10, 26, 58, 0.95); border: 1px solid rgba(138, 43, 226, 0.5); border-radius: 10px; padding: 12px 18px; color: rgba(255,255,255,0.92); font-size: 13px; max-width: 360px; z-index: 1000; pointer-events: none; box-shadow: 0 4px 24px rgba(0,0,0,0.6); text-align: center; line-height: 1.5;"
  >
    <div
      style="display: flex; align-items: center; justify-content: center; gap: 6px; margin-bottom: 4px; color: #a78bfa; font-weight: 600; font-size: 11px; letter-spacing: 0.05em; text-transform: uppercase;"
    >
      <svg
        width="12"
        height="12"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
      >
        <circle cx="12" cy="12" r="10" />
        <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
        <line x1="12" y1="17" x2="12.01" y2="17" />
      </svg>
      Tip
    </div>
    {hotkeysState.helpTooltipMessage}
  </div>
{:else if hotkeysState.showHelpTooltip}
  <div
    class="help-tooltip"
    style="position: fixed; left: {hotkeysState.helpTooltipPosition
      .x}px; top: {hotkeysState.helpTooltipPosition
      .y}px; transform: translate(-50%, -100%); background: rgba(10, 26, 58, 0.95); border: 1px solid rgba(138, 43, 226, 0.5); border-radius: 8px; padding: 10px 14px; color: white; font-size: 13px; max-width: 320px; z-index: 1000; pointer-events: none; box-shadow: 0 4px 20px rgba(0,0,0,0.5);"
  >
    {hotkeysState.helpTooltipMessage}
  </div>
{/if}

<div
  data-testid="graph-stats"
  style="position: absolute; top: 16px; left: 16px; background: rgba(0,0,0,0.6); border: 1px solid rgba(255,255,255,0.2); border-radius: 6px; padding: 6px 10px; color: white; z-index: 50; font-size: 12px; display: flex; align-items: center; gap: 8px;"
>
  {#if loading}
    <span
      style="display: inline-block; width: 12px; height: 12px; border: 2px solid rgba(255,255,255,0.3); border-top-color: white; border-radius: 50%; animation: spin 1s linear infinite;"
    ></span>
  {/if}
  <span>{nodes.length} nodes</span>
  <span style="color: rgba(255,255,255,0.4);">·</span>
  <span>{links.length} links</span>
</div>

<style>
  @keyframes slide-up {
    from {
      opacity: 0;
      transform: translateY(20px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
