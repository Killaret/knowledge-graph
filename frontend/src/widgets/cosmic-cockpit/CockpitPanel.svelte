<script lang="ts">
  import { browser } from "$app/environment";
  import {
    COCKPIT_EDGE_SIZE,
    cockpitStore,
    type CockpitPanelPosition,
  } from "$features/cosmic-cockpit";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";

  interface Props {
    position: CockpitPanelPosition;
    size: number;
    handleSize?: number;
    title?: string;
    delay?: number;
    /** Content-based sizing bounds; when set, the panel measures its own
     * content and reports the clamped natural size via onSizeChange instead
     * of always taking up the full `size`. */
    minSize?: number;
    maxSize?: number;
    onSizeChange?: (size: number) => void;
    children?: import("svelte").Snippet;
  }

  const {
    position,
    size,
    handleSize = COCKPIT_EDGE_SIZE,
    title,
    delay,
    minSize,
    maxSize,
    onSizeChange,
    children,
  }: Props = $props();

  let panelRef: HTMLDivElement | null = $state(null);
  let contentRef: HTMLDivElement | null = $state(null);
  let dragStartX = $state(0);
  let dragStartY = $state(0);
  let dragging = $state(false);
  let hoverTimer: ReturnType<typeof setTimeout> | null = $state(null);
  let closeTimer: ReturnType<typeof setTimeout> | null = $state(null);

  const panel = $derived(cockpitStore.panels[position]);
  const isOpen = $derived(panel.open || panel.pinned || panel.hovering);
  const visibleSize = $derived(isOpen ? size : handleSize);
  const hoverDelay = $derived(delay ?? cockpitStore.hoverDelay);
  const isVertical = $derived(position === "left" || position === "right");

  // Content-based sizing: measure the panel's natural content size and
  // report it up so the panel (and the frame insets it drives) only take up
  // as much space as their content actually needs, clamped to [minSize,
  // maxSize] so panels stay usable.
  $effect(() => {
    if (!browser || !contentRef || minSize == null || maxSize == null || !onSizeChange) return;

    const measure = () => {
      if (!contentRef) return;
      const natural = isVertical ? contentRef.scrollWidth : contentRef.scrollHeight;
      const clamped = Math.min(maxSize, Math.max(minSize, natural));
      onSizeChange(clamped);
    };

    measure();
    const observer = new ResizeObserver(measure);
    observer.observe(contentRef);
    return () => observer.disconnect();
  });

  const locale = getCurrentLocale();
  const t = (key: string, params?: Record<string, string>) => formatMessage(key, locale, params);

  function open() {
    if (hoverTimer) return;
    hoverTimer = setTimeout(() => {
      cockpitStore.openPanel(position);
      hoverTimer = null;
    }, hoverDelay);
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
    }, hoverDelay);
  }

  function cancelClose() {
    if (closeTimer) {
      clearTimeout(closeTimer);
      closeTimer = null;
    }
  }

  function handleEnter() {
    if (cockpitStore.firstPerson) return;
    cockpitStore.hoverPanel(position, true);
    cancelClose();
    if (!panel.open && !panel.pinned) {
      open();
    }
  }

  function handleLeave() {
    cockpitStore.hoverPanel(position, false);
    cancelOpen();
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

  function getPanelStyle(): string {
    const base =
      "position:absolute; z-index:150; display:flex; overflow:hidden; pointer-events:auto;";
    const flexDir =
      position === "left" || position === "right" ? "row" : "column";
    const isVertical = flexDir === "row";

    const transform = isVertical
      ? position === "left"
        ? "transform: translateX(calc(var(--panel-visible) - var(--panel-full)));"
        : "transform: translateX(calc(var(--panel-full) - var(--panel-visible)));"
      : position === "top"
        ? "transform: translateY(calc(var(--panel-visible) - var(--panel-full)));"
        : "transform: translateY(calc(var(--panel-full) - var(--panel-visible)));";

    switch (position) {
      case "left":
        return `${base} flex-direction:${flexDir}; top:var(--inset-top,0); bottom:var(--inset-bottom,0); left:0; width:var(--panel-full); ${transform}`;
      case "right":
        return `${base} flex-direction:${flexDir}; top:var(--inset-top,0); bottom:var(--inset-bottom,0); right:0; width:var(--panel-full); ${transform}`;
      case "top":
        return `${base} flex-direction:${flexDir}; left:var(--inset-left,0); right:var(--inset-right,0); top:0; height:var(--panel-full); ${transform}`;
      case "bottom":
        return `${base} flex-direction:${flexDir}; left:var(--inset-left,0); right:var(--inset-right,0); bottom:0; height:var(--panel-full); ${transform}`;
    }
  }

  function getTransition(): string {
    return cockpitStore.reducedMotion ? "none" : "transform 0.3s ease";
  }

  const arrowRotation = $derived(
    ({
      left: 0,
      right: 180,
      top: 90,
      bottom: -90,
    } as const)[position]
  );

  // Desynchronize the gradient-text animation phase per panel position so
  // all four panel titles don't shimmer in lockstep.
  const panelTitleDelay = $derived(
    `${
      (
        {
          top: 0,
          right: -1.5,
          bottom: -3,
          left: -4.5,
        } as const
      )[position]
    }s`
  );
</script>

{#if browser}
  <div
    bind:this={panelRef}
    class="cockpit-panel cockpit-panel--{position}"
    class:first-person={cockpitStore.firstPerson}
    style="--panel-full:{size}px;--panel-visible:{visibleSize}px;--panel-handle-size:{handleSize}px;{getPanelStyle()} transition:{getTransition()};"
    onmouseenter={handleEnter}
    onmouseleave={handleLeave}
    data-testid="cockpit-panel-{position}"
    role="region"
    aria-label="{title ? `${title} — ` : ""}{t('cockpit.panel.ariaLabel', { position })}"
  >
    <div class="panel-glow" aria-hidden="true"></div>

    {#if (position === "right" || position === "bottom") && !isOpen}
      <div
        class="panel-handle panel-handle--{position}"
        onpointerdown={handlePointerDown}
        onpointermove={handlePointerMove}
        onpointerup={handlePointerUp}
        onpointercancel={handlePointerUp}
        data-testid="cockpit-handle-{position}"
        aria-label={t("cockpit.handle.open", { position })}
        role="button"
        tabindex="-1"
      >
        <svg
          class="handle-arrow"
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2.5"
          stroke-linecap="round"
          stroke-linejoin="round"
          style="transform: rotate({arrowRotation}deg);"
          aria-hidden="true"
        >
          <polyline points="9 18 15 12 9 6" />
        </svg>
      </div>
    {/if}

    <div class="panel-body">
      <header class="panel-header">
        {#if title}
          <h3
            class="panel-title cockpit-gradient-text"
            style="--cockpit-text-delay: {panelTitleDelay};"
          >
            {title}
          </h3>
        {/if}
        <div class="panel-actions" class:panel-actions--no-title={!title}>
          <button
            type="button"
            class="pin-btn"
            class:pinned={panel.pinned}
            onclick={togglePin}
            aria-pressed={panel.pinned}
            aria-label={panel.pinned ? t("cockpit.panel.unpin") : t("cockpit.panel.pin")}
            data-testid="cockpit-panel-pin-{position}"
            title={panel.pinned ? t("cockpit.panel.unpin") : t("cockpit.panel.pin")}
          >
            {panel.pinned ? "📌" : "📎"}
          </button>
          <button
            type="button"
            class="close-btn"
            onclick={close}
            aria-label={t("cockpit.panel.close")}
            data-testid="cockpit-panel-close-{position}"
            title={t("cockpit.panel.close")}
          >
            ✕
          </button>
        </div>
      </header>

      <div class="panel-content" bind:this={contentRef}>
        {@render children?.()}
      </div>
    </div>

    {#if (position === "left" || position === "top") && !isOpen}
      <div
        class="panel-handle panel-handle--{position}"
        onpointerdown={handlePointerDown}
        onpointermove={handlePointerMove}
        onpointerup={handlePointerUp}
        onpointercancel={handlePointerUp}
        data-testid="cockpit-handle-{position}"
        aria-label={t("cockpit.handle.open", { position })}
        role="button"
        tabindex="-1"
      >
        <svg
          class="handle-arrow"
          width="14"
          height="14"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2.5"
          stroke-linecap="round"
          stroke-linejoin="round"
          style="transform: rotate({arrowRotation}deg);"
          aria-hidden="true"
        >
          <polyline points="9 18 15 12 9 6" />
        </svg>
      </div>
    {/if}
  </div>
{/if}

<style>
  .cockpit-panel {
    --cockpit-bevel: 10px;
    background: rgba(10, 10, 15, 0.92);
    backdrop-filter: blur(18px);
    border: 1px solid rgba(45, 212, 191, 0.25);
    box-shadow:
      0 0 28px rgba(0, 0, 0, 0.55),
      0 0 12px rgba(45, 212, 191, 0.08);
    color: var(--color-text, #e0e0e0);
    /* Cut all four corners at 45° so the panel reads as a beveled sci-fi
       console surface instead of a flat rectangle — the same visual
       language as the pseudo-3D corner bolts on CockpitFrame, applied to
       the panels themselves. */
    clip-path: polygon(
      var(--cockpit-bevel) 0%,
      calc(100% - var(--cockpit-bevel)) 0%,
      100% var(--cockpit-bevel),
      100% calc(100% - var(--cockpit-bevel)),
      calc(100% - var(--cockpit-bevel)) 100%,
      var(--cockpit-bevel) 100%,
      0% calc(100% - var(--cockpit-bevel)),
      0% var(--cockpit-bevel)
    );
  }

  .cockpit-panel.first-person {
    opacity: 0;
    pointer-events: none;
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
    z-index: 0;
  }

  .panel-body {
    position: relative;
    z-index: 1;
    display: flex;
    flex-direction: column;
    flex: 1;
    min-width: 0;
    min-height: 0;
    overflow: hidden;
  }

  .panel-header {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 10px 18px;
    min-height: 44px;
    border-bottom: 1px solid rgba(45, 212, 191, 0.15);
    user-select: none;
    flex-shrink: 0;
  }

  /* Small diagonal accents at the header's top corners, echoing the cut
     bevel of the panel itself and CockpitFrame's corner bolts, so the
     header visually reads as part of the same beveled console surface. */
  .panel-header::before,
  .panel-header::after {
    content: "";
    position: absolute;
    top: 0;
    width: 14px;
    height: 2px;
    background: linear-gradient(90deg, rgba(45, 212, 191, 0.7), transparent);
    transform-origin: 0 0;
    transform: rotate(45deg);
    pointer-events: none;
  }

  .panel-header::before {
    left: 0;
  }

  .panel-header::after {
    right: 0;
    background: linear-gradient(90deg, transparent, rgba(45, 212, 191, 0.7));
    transform-origin: 100% 0;
    transform: rotate(-45deg);
  }

  .panel-title {
    margin: 0;
    font-size: 13px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.08em;
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
    color: var(--cockpit-accent, #2dd4bf);
  }

  .panel-content {
    flex: 1;
    overflow: auto;
    padding: 12px;
    min-width: 0;
    min-height: 0;
  }

  .panel-handle {
    position: relative;
    z-index: 2;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    cursor: pointer;
    background: linear-gradient(
      180deg,
      rgba(45, 212, 191, 0.15) 0%,
      rgba(10, 10, 15, 0.9) 45%,
      rgba(10, 10, 15, 0.9) 55%,
      rgba(192, 38, 211, 0.12) 100%
    );
    border: 1px solid rgba(45, 212, 191, 0.22);
    box-shadow:
      inset 0 0 10px rgba(45, 212, 191, 0.08),
      0 0 12px rgba(0, 0, 0, 0.5);
    transition:
      background 0.25s ease,
      box-shadow 0.25s ease;
  }

  .panel-handle--left,
  .panel-handle--right {
    width: var(--panel-handle-size, 24px);
    height: auto;
    align-self: stretch;
  }

  .panel-handle--top,
  .panel-handle--bottom {
    height: var(--panel-handle-size, 24px);
    width: auto;
    align-self: stretch;
  }

  .panel-handle:hover {
    background: linear-gradient(
      180deg,
      rgba(45, 212, 191, 0.28) 0%,
      rgba(20, 20, 35, 0.95) 45%,
      rgba(20, 20, 35, 0.95) 55%,
      rgba(192, 38, 211, 0.2) 100%
    );
    box-shadow:
      inset 0 0 16px rgba(45, 212, 191, 0.15),
      0 0 18px rgba(45, 212, 191, 0.15);
  }

  .handle-arrow {
    color: rgba(45, 212, 191, 0.75);
    filter: drop-shadow(0 0 4px rgba(45, 212, 191, 0.4));
    animation: pulse 1.6s ease-in-out infinite;
  }

  .panel-handle.open .handle-arrow {
    animation: none;
    opacity: 0.4;
  }

  @keyframes pulse {
    0%,
    100% {
      opacity: 0.55;
      transform: scale(1);
    }
    50% {
      opacity: 1;
      transform: scale(1.15);
    }
  }

  @media (prefers-reduced-motion: reduce) {
    .cockpit-panel {
      transition: none !important;
    }
    .handle-arrow {
      animation: none;
    }
  }
</style>
