<script lang="ts">
  import { browser } from "$app/environment";
  import {
    COCKPIT_EDGE_SIZE,
    cockpitStore,
    type CockpitPanelPosition,
  } from "$features/cosmic-cockpit";

  interface Props {
    position: CockpitPanelPosition;
    size: number;
    title?: string;
    children?: import("svelte").Snippet;
  }

  const { position, size, title, children }: Props = $props();

  let edgeRef: HTMLDivElement | null = $state(null);
  let panelRef: HTMLDivElement | null = $state(null);
  let dragStartX = $state(0);
  let dragStartY = $state(0);
  let dragging = $state(false);
  let hoverTimer: ReturnType<typeof setTimeout> | null = $state(null);
  let closeTimer: ReturnType<typeof setTimeout> | null = $state(null);

  const panel = $derived(cockpitStore.panels[position]);
  const isOpen = $derived(panel.open || panel.pinned || panel.hovering);

  function open() {
    if (hoverTimer) return; // already queued
    hoverTimer = setTimeout(() => {
      cockpitStore.openPanel(position);
      hoverTimer = null;
    }, cockpitStore.hoverDelay);
  }

  function cancelOpen() {
    if (hoverTimer) {
      clearTimeout(hoverTimer);
      hoverTimer = null;
    }
  }

  function scheduleClose() {
    if (!cockpitStore.autoCollapse || panel.pinned) return;
    if (closeTimer) clearTimeout(closeTimer);
    closeTimer = setTimeout(() => {
      if (!panel.pinned && !panel.hovering) {
        cockpitStore.closePanel(position);
      }
      closeTimer = null;
    }, cockpitStore.hoverDelay);
  }

  function cancelClose() {
    if (closeTimer) {
      clearTimeout(closeTimer);
      closeTimer = null;
    }
  }

  function handleEdgeEnter() {
    if (cockpitStore.firstPerson) return;
    cockpitStore.hoverPanel(position, true);
    cancelClose();
    if (!panel.open && !panel.pinned) {
      open();
    }
  }

  function handleEdgeLeave() {
    cockpitStore.hoverPanel(position, false);
    cancelOpen();
    if (panel.open && !panel.pinned) {
      scheduleClose();
    }
  }

  function handlePanelEnter() {
    if (cockpitStore.firstPerson) return;
    cockpitStore.hoverPanel(position, true);
    cancelClose();
  }

  function handlePanelLeave() {
    cockpitStore.hoverPanel(position, false);
    if (panel.open && !panel.pinned) {
      scheduleClose();
    }
  }

  function handlePointerDown(e: PointerEvent) {
    if (cockpitStore.firstPerson) return;
    dragging = true;
    dragStartX = e.clientX;
    dragStartY = e.clientY;
    (e.currentTarget as Element | null)?.setPointerCapture?.(e.pointerId);
  }

  function handlePointerMove(e: PointerEvent) {
    if (!dragging || cockpitStore.firstPerson) return;
    const dx = e.clientX - dragStartX;
    const dy = e.clientY - dragStartY;

    const threshold = cockpitStore.edgeSensitivity;
    let pulled = false;

    if (position === "left" && dx > threshold) pulled = true;
    if (position === "right" && -dx > threshold) pulled = true;
    if (position === "top" && dy > threshold) pulled = true;
    if (position === "bottom" && -dy > threshold) pulled = true;

    if (pulled && !panel.open) {
      cockpitStore.openPanel(position);
      dragging = false;
    }
  }

  function handlePointerUp() {
    dragging = false;
  }

  function togglePin() {
    cockpitStore.togglePin(position);
  }

  function close() {
    cockpitStore.setPanel(position, { open: false, pinned: false });
  }

  function getEdgeStyle(): string {
    switch (position) {
      case "top":
        return `top:0;left:0;width:100vw;height:${COCKPIT_EDGE_SIZE}px;`;
      case "bottom":
        return `bottom:0;left:0;width:100vw;height:${COCKPIT_EDGE_SIZE}px;`;
      case "left":
        return `top:0;left:0;width:${COCKPIT_EDGE_SIZE}px;height:100vh;`;
      case "right":
        return `top:0;right:0;width:${COCKPIT_EDGE_SIZE}px;height:100vh;`;
    }
  }

  function getPanelStyle(): string {
    switch (position) {
      case "top":
        return `top:0;left:0;width:100vw;height:${size}px;transform:translateY(${isOpen ? 0 : -size}px);`;
      case "bottom":
        return `bottom:0;left:0;width:100vw;height:${size}px;transform:translateY(${isOpen ? 0 : size}px);`;
      case "left":
        return `top:0;left:0;width:${size}px;height:100vh;transform:translateX(${isOpen ? 0 : -size}px);`;
      case "right":
        return `top:0;right:0;width:${size}px;height:100vh;transform:translateX(${isOpen ? 0 : size}px);`;
    }
  }

  function getTransition(): string {
    return cockpitStore.reducedMotion ? "none" : "transform 0.3s ease";
  }
</script>

{#if browser}
  <!-- Invisible edge trigger for hover and drag-to-open -->
  <div
    bind:this={edgeRef}
    class="cockpit-edge cockpit-edge--{position}"
    class:first-person={cockpitStore.firstPerson}
    style={getEdgeStyle()}
    onmouseenter={handleEdgeEnter}
    onmouseleave={handleEdgeLeave}
    onpointerdown={handlePointerDown}
    onpointermove={handlePointerMove}
    onpointerup={handlePointerUp}
    onpointercancel={handlePointerUp}
    role="button"
    aria-label="Open {position} panel"
    tabindex="-1"
  ></div>
{/if}

<!-- Slide-out panel -->
<div
  bind:this={panelRef}
  class="cockpit-panel cockpit-panel--{position}"
  class:first-person={cockpitStore.firstPerson}
  style="{getPanelStyle()} transition: {getTransition()};"
  onmouseenter={handlePanelEnter}
  onmouseleave={handlePanelLeave}
  data-testid="cockpit-panel-{position}"
  role="region"
  aria-label="{position} cockpit panel"
>
  <div class="panel-glow"></div>
  <header class="panel-header">
    {#if title}
      <h3 class="panel-title">{title}</h3>
    {/if}
    <div class="panel-actions" class:panel-actions--no-title={!title}>
      <button
        type="button"
        class="pin-btn"
        class:pinned={panel.pinned}
        onclick={togglePin}
        aria-pressed={panel.pinned}
        aria-label={panel.pinned ? "Unpin panel" : "Pin panel"}
        data-testid="cockpit-panel-pin-{position}"
      >
        {panel.pinned ? "📌" : "📎"}
      </button>
      <button
        type="button"
        class="close-btn"
        onclick={close}
        aria-label="Close panel"
        data-testid="cockpit-panel-close-{position}"
      >
        ✕
      </button>
    </div>
  </header>

  <div class="panel-content">
    {@render children?.()}
  </div>
</div>

<style>
  .cockpit-edge {
    position: fixed;
    z-index: 200;
    background: transparent;
    cursor: pointer;
  }

  .cockpit-edge::after {
    content: "";
    position: absolute;
    background: rgba(45, 212, 191, 0.25);
    box-shadow: 0 0 12px rgba(45, 212, 191, 0.3);
    opacity: 0;
    transition: opacity 0.2s ease;
  }

  .cockpit-edge--top::after,
  .cockpit-edge--bottom::after {
    left: 50%;
    transform: translateX(-50%);
    width: 120px;
    height: 4px;
    border-radius: 2px;
  }

  .cockpit-edge--top::after {
    top: 10px;
  }

  .cockpit-edge--bottom::after {
    bottom: 10px;
  }

  .cockpit-edge--left::after,
  .cockpit-edge--right::after {
    top: 50%;
    transform: translateY(-50%);
    width: 4px;
    height: 120px;
    border-radius: 2px;
  }

  .cockpit-edge--left::after {
    left: 10px;
  }

  .cockpit-edge--right::after {
    right: 10px;
  }

  .cockpit-edge:hover::after {
    opacity: 1;
  }

  .cockpit-edge.first-person {
    opacity: 0;
    pointer-events: none;
  }

  .cockpit-panel {
    position: fixed;
    z-index: 150;
    display: flex;
    flex-direction: column;
    background: rgba(10, 10, 15, 0.88);
    backdrop-filter: blur(16px);
    border: 1px solid rgba(45, 212, 191, 0.25);
    box-shadow: 0 0 24px rgba(45, 212, 191, 0.12);
    color: var(--color-text, #e0e0e0);
    overflow: hidden;
    pointer-events: auto;
  }

  .cockpit-panel.first-person {
    opacity: 0;
    pointer-events: none;
  }

  .cockpit-panel--top {
    border-bottom: 1px solid rgba(45, 212, 191, 0.4);
  }

  .cockpit-panel--bottom {
    border-top: 1px solid rgba(45, 212, 191, 0.4);
  }

  .cockpit-panel--left {
    border-right: 1px solid rgba(45, 212, 191, 0.4);
  }

  .cockpit-panel--right {
    border-left: 1px solid rgba(45, 212, 191, 0.4);
  }

  .panel-glow {
    position: absolute;
    inset: 0;
    background: linear-gradient(
      135deg,
      rgba(45, 212, 191, 0.04) 0%,
      transparent 50%,
      rgba(192, 38, 211, 0.04) 100%
    );
    pointer-events: none;
  }

  .panel-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 14px;
    min-height: 44px;
    border-bottom: 1px solid rgba(45, 212, 191, 0.15);
    user-select: none;
  }

  .panel-title {
    margin: 0;
    font-size: 13px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: rgba(45, 212, 191, 0.9);
  }

  .panel-actions {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-left: auto;
  }

  .panel-actions--no-title {
    margin-left: auto;
  }

  .pin-btn,
  .close-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    border: none;
    border-radius: 6px;
    background: rgba(255, 255, 255, 0.06);
    color: var(--color-text, #e0e0e0);
    cursor: pointer;
    font-size: 14px;
    transition: background 0.2s ease;
  }

  .pin-btn:hover,
  .close-btn:hover {
    background: rgba(45, 212, 191, 0.2);
  }

  .pin-btn.pinned {
    background: rgba(45, 212, 191, 0.25);
    color: #2dd4bf;
  }

  .panel-content {
    flex: 1;
    overflow: auto;
    padding: 12px;
  }

  @media (max-width: 768px) {
    .cockpit-edge::after {
      opacity: 1;
    }
  }
</style>
