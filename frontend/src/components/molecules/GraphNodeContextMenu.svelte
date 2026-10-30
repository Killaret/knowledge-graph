<script lang="ts">
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
  import { onMount } from "svelte";

  interface Props {
    x: number;
    y: number;
    visible: boolean;
    node?: { id: string; title: string; type?: string };
    onClose: () => void;
    onCreateChild: () => void;
    onViewDetails?: () => void;
  }

  const { x, y, visible, node, onClose, onCreateChild, onViewDetails }: Props = $props();

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  let menuEl: HTMLDivElement | null = $state(null);

  function handleClickOutside(e: MouseEvent) {
    if (menuEl && !menuEl.contains(e.target as Node)) {
      onClose();
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === "Escape") {
      onClose();
    }
  }

  onMount(() => {
    window.addEventListener("click", handleClickOutside);
    window.addEventListener("keydown", handleKeyDown);
    return () => {
      window.removeEventListener("click", handleClickOutside);
      window.removeEventListener("keydown", handleKeyDown);
    };
  });

  function adjustPosition(node: HTMLDivElement | null): { x: number; y: number } {
    if (!node || typeof window === "undefined") return { x, y };
    const rect = node.getBoundingClientRect();
    const winW = window.innerWidth;
    const winH = window.innerHeight;
    let adjustedX = x;
    let adjustedY = y;
    if (adjustedX + rect.width > winW) {
      adjustedX = Math.max(8, winW - rect.width - 8);
    }
    if (adjustedY + rect.height > winH) {
      adjustedY = Math.max(8, winH - rect.height - 8);
    }
    return { x: adjustedX, y: adjustedY };
  }

  $effect(() => {
    if (visible && menuEl) {
      const pos = adjustPosition(menuEl);
      menuEl.style.left = `${pos.x}px`;
      menuEl.style.top = `${pos.y}px`;
    }
  });
</script>

{#if visible && node}
  <div
    bind:this={menuEl}
    class="graph-node-context-menu"
    role="menu"
    aria-label={t("graph.contextMenu.ariaLabel")}
    style="position: fixed; left: {x}px; top: {y}px;"
  >
    <div class="context-header">
      <span class="node-emoji">{node.type ? "●" : "●"}</span>
      <span class="node-title" title={node.title}>{node.title}</span>
    </div>
    <button
      type="button"
      class="context-item"
      role="menuitem"
      onclick={() => {
        onCreateChild();
        onClose();
      }}
      data-testid="context-menu-create-child"
    >
      {t("graph.contextMenu.createChildNote")}
    </button>
    {#if onViewDetails}
      <button
        type="button"
        class="context-item"
        role="menuitem"
        onclick={() => {
          onViewDetails();
          onClose();
        }}
        data-testid="context-menu-view-details"
      >
        {t("graph.contextMenu.viewDetails")}
      </button>
    {/if}
  </div>
{/if}

<style>
  .graph-node-context-menu {
    min-width: 180px;
    background: rgba(10, 15, 30, 0.96);
    border: 1px solid rgba(45, 212, 191, 0.3);
    border-radius: 10px;
    padding: 8px 0;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.5);
    z-index: 200;
    backdrop-filter: blur(12px);
    color: #e0e0e0;
  }

  .context-header {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 14px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    margin-bottom: 4px;
  }

  .node-emoji {
    font-size: 12px;
    color: #2dd4bf;
  }

  .node-title {
    font-size: 13px;
    font-weight: 600;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    max-width: 200px;
  }

  .context-item {
    display: block;
    width: 100%;
    text-align: left;
    background: transparent;
    border: none;
    color: inherit;
    padding: 10px 14px;
    font-size: 13px;
    cursor: pointer;
    transition: background 0.15s ease;
  }

  .context-item:hover,
  .context-item:focus {
    background: rgba(45, 212, 191, 0.12);
    outline: none;
  }
</style>
