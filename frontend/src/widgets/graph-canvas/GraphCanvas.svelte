<script lang="ts">
  import { onMount } from "svelte";
  import { browser } from "$app/environment";
  import { formatMessage, getCurrentLocale } from "$shared/utils/i18n";
  import { graphConfig2D } from "$shared/config";
  import type { GraphDeltaData } from "$shared/api/graph";
  import { GraphCanvasOverlay, GraphCanvasModals, LinkTypeLegend } from "$features/graph-ui";
  import GraphNodeContextMenu from "$components/molecules/GraphNodeContextMenu.svelte";
  import { graphStore } from "$shared/stores/graph.svelte";
  import HelpHotkeysModal from "$components/organisms/HelpHotkeysModal.svelte";
  import { ParticleSystem } from "$entities/graph-canvas/lib/particle-system";
  import {
    resizeCanvas,
    setupResizeObserver,
    scheduleDelayedResize,
    startSimulation,
    clearSimulation,
    getSimulationNodes,
    type SimulationState,
    type SimulationNode,
    draw,
    resetView,
    startAnimationLoop,
    clearAnimationState,
    updateNodeAngles,
    type TransformState,
    type DragState,
    type BlackHoleState,
    createBlackHole,
    updateBlackHolePosition,
    updateBlackHolePulse,
    updateBlackHoleZoom,
    isPointOverBlackHole,
    type GhostNodeState,
    updateGhostNodePosition,
    updateGhostNodePulse,
    updateGhostNodeZoom,
    type GravitySystem,
    drawFog,
    getHoveredNeighborIds,
    applyDelta as applyDeltaToSimulation,
  } from "$entities/graph-canvas/lib";
  import { createGhostNode } from "$entities/graph-canvas/lib/ghost-node";
  import { createGravitySystem } from "$entities/graph-canvas/lib/gravity-system";

  const locale = getCurrentLocale();
  const t = (key: string) => formatMessage(key, locale);

  // FSD imports
  import {
    createDragDropState,
    type DragDropState,
  } from "$features/graph-interaction/drag-and-drop";
  import {
    createHotkeysState,
    resetInactivityTimer,
    showRandomTip,
    updateSearch,
    type HotkeysState,
  } from "$features/graph-interaction/hotkeys";
  import { createZoomPanState, type ZoomPanState } from "$features/graph-interaction/zoom-pan";
  import {
    attachEvents,
    type GraphCanvasEventContext,
  } from "$features/graph-interaction/event-bridge";
  import {
    createGraphCanvasState,
    isTechnicalNode,
    pinTechnicalNodes,
  } from "$features/graph-canvas/canvas-state.svelte";
  import { createFogState } from "$features/graph-canvas/fog-state.svelte";
  import {
    createFogWarningState,
    updateFogWarning,
  } from "$features/graph-canvas/fog-warning";
  import {
    createNoteFormState,
    createNote,
    closeNoteForm,
    type NoteFormState,
  } from "$features/graph-forms/note-form";
  import {
    createLinkFormState,
    createLink,
    closeLinkForm,
    type LinkFormState,
  } from "$features/graph-forms/link-form";

  /* eslint-disable prefer-const -- Svelte 5 props are destructured from a reactive $props() proxy; they are conventionally declared with `let` even when not reassigned locally, especially for bindable props. */
  let {
    nodes,
    links,
    onNodeClick,
    onLinkEdit,
    onLinkDelete,
    onNoteCreate,
    onLinkCreate,
    onNoteDelete,
    onNoteRestore,
    onCreateChildNote,
    helpContent,
    delta,
    disableVariation = false,
    readonly = false,
    showLinkTypeLegend = true,
    className = "",
    controller = $bindable<
      | {
          focusMode: boolean;
          resetView: () => void;
          openSearch: () => void;
          toggleFocus: () => void;
          fogEnabled: boolean;
          toggleFog: () => void;
        }
      | undefined
    >(undefined),
  }: {
    nodes: Array<{
      id: string;
      title: string;
      type?: string;
      createdAt?: string;
      created_at?: string;
    }>;
    links: Array<{
      id?: string;
      source: string;
      target: string;
      weight?: number;
      link_type?: string;
      source_type?: string;
      last_weight_update?: string;
    }>;
    onNodeClick?: (node: { id: string; title: string; type?: string }) => void;
    onLinkEdit?: (link: {
      id?: string;
      source: string;
      target: string;
      link_type: string;
      weight: number;
    }) => void;
    onLinkDelete?: (link: {
      id?: string;
      source: string;
      target: string;
      link_type: string;
    }) => void;
    onNoteCreate?: (data: { title: string; content: string; type: string }) => void;
    onLinkCreate?: (link: {
      source: string;
      target: string;
      link_type: string;
      weight: number;
    }) => void;
    onNoteDelete?: (nodeId: string) => void;
    onNoteRestore?: (nodeId: string) => void;
    onCreateChildNote?: (node: { id: string; title: string; type?: string }) => void;
    helpContent?: string;
    delta?: GraphDeltaData;
    disableVariation?: boolean;
    readonly?: boolean;
    showLinkTypeLegend?: boolean;
    className?: string;
    controller?: {
      focusMode: boolean;
      resetView: () => void;
      openSearch: () => void;
      toggleFocus: () => void;
      fogEnabled: boolean;
      toggleFog: () => void;
    };
  } = $props();
  /* eslint-enable prefer-const */

  let stableRender = false;
  $effect(() => {
    stableRender =
      disableVariation ||
      (browser && new URL(window.location.href).searchParams.get("stableRender") === "true");
  });

  // Debug: проверяем типы узлов при изменении (dev only)
  $effect(() => {
    if (import.meta.env.DEV && nodes.length > 0) {
      const types = nodes.map((n) => n.type || "undefined");
      const uniqueTypes = [...new Set(types)];
      if (import.meta.env.DEV) {
        console.log("[GraphCanvas] Received nodes types:", uniqueTypes, "Total:", nodes.length);
      }
      if (import.meta.env.DEV) {
        console.log("[GraphCanvas] First node:", nodes[0]);
      }
    }
  });

  let canvas: HTMLCanvasElement | null = $state(null);
  let ctx: CanvasRenderingContext2D | null = null;
  let offscreenCanvas: HTMLCanvasElement | null = null;
  let offscreenCtx: CanvasRenderingContext2D | null = null;
  let lastCacheKey = "";
  let width = 800;
  let height = 600;
  let animationLoop: { stop: () => void } | null = null;
  let resizeCleanup: { clear: () => void } | null = null;
  let observerCleanup: { disconnect: () => void } | null = null;
  let detachEvents: (() => void) | null = null;
  const angles = new Map<string, number>();
  const speeds = new Map<string, number>();

  // NOTE: transform and dragState need reactivity for Svelte bindings,
  // but simState must NOT be $state — d3-force mutates link objects (source/target become node refs)
  // and Svelte 5 Proxy intercepts those mutations, breaking d3 internals.
  const transform: TransformState = $state({ x: 0, y: 0, k: 1 });
  const dragState: DragState = $state({
    dragging: false,
    dragStart: { x: 0, y: 0 },
  });
  const simState: SimulationState = {
    simulation: null,
    simLinks: [],
    isRunning: false,
    stable: false,
    nodeOpacity: new Map(),
    linkOpacity: new Map(),
    dyingLinks: [],
    dyingLinkOpacity: new Map(),
    fadeAnimationId: null,
  };

  // Filter links based on hidden types and minimum weight
  const visibleLinks = $derived(
    links.filter((l) => {
      const typeMatch = !graphStore.hiddenLinkTypes.includes(l.link_type ?? "related");
      const weightMatch = (l.weight ?? 0.5) >= graphStore.minLinkWeight;
      return typeMatch && weightMatch;
    })
  );

  // Для отслеживания изменений данных по содержимому (не по ссылке)
  let lastDataKey = "";
  let mounted = $state(false);

  // Используем утилиты для resize
  const resizeState = { width, height };

  // Double-tap zoom state (FSD)
  const zoomPanState: ZoomPanState = $state(createZoomPanState());

  const canvasState = createGraphCanvasState();
  const fogState = createFogState();

  // Expose simulation stability for visual regression tests
  $effect(() => {
    if (canvas) {
      canvas.dataset.testStable = graphStable ? "true" : "false";
    }
  });

  // Particle system
  let particleSystem: ParticleSystem | null = $state(null);

  // Time for animations
  let animationTime = $state(0);
  let graphStable = $state(false);

  // Fog warning controller: debounced one-shot toast for low-FPS adaptive mode.
  let fogWarningState = $state(createFogWarningState(0));

  // Throttle rendering: only the animation loop does actual drawing;
  // D3 ticks and input events just request a frame.
  let needsRedraw = false;
  let lastDrawTimestamp = 0;
  const IDLE_FPS = graphConfig2D.idle_fps;

  // Interactive canvas elements
  let blackHole: BlackHoleState = $state(createBlackHole(width, height));
  let ghostNode: GhostNodeState = $state(createGhostNode(width, height, []));
  let gravitySystem: GravitySystem = $state(createGravitySystem());

  // Update ghost node when nodes change
  $effect(() => {
    ghostNode = createGhostNode(width, height, nodes);
  });

  // Drag-and-drop state (FSD)
  const dragDropState: DragDropState = $state(createDragDropState());

  // Note creation form state (FSD)
  let noteFormState: NoteFormState = $state(createNoteFormState());

  // Link creation form state (FSD)
  let linkFormState: LinkFormState = $state(createLinkFormState());

  // Context menu for right-clicked node
  let contextMenu = $state<{
    visible: boolean;
    x: number;
    y: number;
    node: { id: string; title: string; type?: string } | null;
  }>({
    visible: false,
    x: 0,
    y: 0,
    node: null,
  });

  // Hotkeys state (FSD)
  const hotkeysState: HotkeysState = $state(createHotkeysState());

  // Expose controller for external control panels (e.g. public graph top bar).
  // We read the reactive values here so the controller object is recreated and
  // the top-bar active states update when focus/fog change.
  $effect(() => {
    const focusMode = canvasState.focusMode;
    const fogEnabled = fogState.enabled;
    controller = {
      focusMode,
      fogEnabled,
      resetView: () => {
        const simNodes = getSimulationNodes(simState);
        if (ctx && simNodes.length > 0) {
          resetView(ctx, width, height, simNodes, transform);
        }
      },
      openSearch: () => canvasState.handleOpenSearch(hotkeysState),
      toggleFocus: () => canvasState.handleToggleFocus(redraw),
      toggleFog: () => fogState.toggle(),
    };
  });

  onMount(() => {
    if (!browser || !canvas) return;

    // Expose for debugging
    window.__graphCanvas = {
      getSimulationNodes: () => getSimulationNodes(simState),
      transform,
    };

    // SSR-safe: получаем контекст canvas
    ctx = canvas.getContext("2d")!;

    // Начальный resize
    resizeCanvas(canvas!, resizeState);
    width = resizeState.width;
    height = resizeState.height;

    // Initialize interactive systems
    particleSystem = new ParticleSystem(nodes.length);
    blackHole = createBlackHole(width, height);
    blackHole.label = t("graph.blackHole.tooltip");
    ghostNode = createGhostNode(width, height, nodes);
    gravitySystem = createGravitySystem();

    // ResizeObserver для отслеживания размера контейнера
    observerCleanup = setupResizeObserver(canvas!, () => {
      resizeCanvas(canvas!, resizeState);
      width = resizeState.width;
      height = resizeState.height;
    });

    // Отложенный resize для стабильных размеров
    resizeCleanup = scheduleDelayedResize(() => {
      resizeCanvas(canvas!, resizeState);
      width = resizeState.width;
      height = resizeState.height;
    }, 100);

    // Start the animation loop. The loop ticks every rAF frame, but the
    // onUpdate callback below skips heavy work and drawing when the graph is
    // stable and the user is not interacting.
    animationLoop = startAnimationLoop((timestamp) => {
      if (!ctx) return;

      // Track FPS every frame; this is cheap and keeps performance metrics
      // accurate for the adaptive fog system.
      fogState.tick(timestamp);

      // Throttle drawing: render at full 60 fps while the graph is moving or
      // the user is interacting; otherwise fall back to idle_fps to save CPU.
      const isInteracting =
        !!canvasState.hoveredNodeId ||
        !!dragDropState.draggedNodeId ||
        dragDropState.isDraggingForLink ||
        !!dragDropState.linkPreviewTarget ||
        !!canvasState.focusMode ||
        dragState.dragging;
      const busy = !graphStable || isInteracting;
      const elapsed = timestamp - lastDrawTimestamp;
      const idleFrameInterval = 1000 / IDLE_FPS;
      const shouldDraw = busy || needsRedraw || elapsed >= idleFrameInterval;
      if (!(shouldDraw && elapsed >= (busy ? 1000 / 60 : idleFrameInterval))) {
        return;
      }

      lastDrawTimestamp = timestamp;

      // In stable render mode keep animation time fixed for deterministic screenshots
      if (stableRender) {
        animationTime = 0;
      } else {
        animationTime = performance.now();
      }

      // Update interactive element positions, zoom scale, and pulses
      updateBlackHoleZoom(blackHole, transform.k);
      updateBlackHolePosition(blackHole, width, height);
      updateBlackHolePulse(blackHole, animationTime);
      updateGhostNodeZoom(ghostNode, transform.k);
      updateGhostNodePosition(ghostNode, width, height, nodes);
      updateGhostNodePulse(ghostNode, animationTime);

      // Apply subtle gravity attraction only when not taking stable screenshots
      const simNodes = getSimulationNodes(simState);
      if (!stableRender && gravitySystem.isEnabled(simNodes.length)) {
        gravitySystem.applyAttraction(simNodes);
      }

      // Update node rotation angles; these are read by draw() during redraw.
      updateNodeAngles(simNodes, angles, speeds, stableRender);

      // Build a node map once per frame for fast neighbor/radius lookups.
      const nodeMap = new Map<string, SimulationNode>();
      for (const node of simNodes) {
        nodeMap.set(node.id, node);
      }

      // Update fog center/radius for adaptive rendering.
      const hoveredNode = canvasState.hoveredNodeId
        ? (nodeMap.get(canvasState.hoveredNodeId) ?? null)
        : null;
      const hoveredNeighborIds = getHoveredNeighborIds(
        canvasState.hoveredNodeId,
        simState.simLinks
      );
      fogState.update(
        width,
        height,
        transform,
        hoveredNode,
        canvasState.focusMode,
        simState.simLinks,
        nodeMap
      );

      // Manage the debounced low-FPS warning toast. It should only fire once per
      // high-load episode, stay visible for 2 seconds, and only trigger while the
      // user is actively interacting with the canvas.
      fogWarningState = updateFogWarning(
        fogWarningState,
        timestamp,
        fogState.showWarning,
        isInteracting
      );

      needsRedraw = true;
      doRedraw(simNodes, hoveredNeighborIds);
    });

    mounted = true; // triggers $effect re-run since it's $state

    resetInactivityTimer(hotkeysState, () => showRandomTip(hotkeysState, canvasState.hotkeyLines));

    detachEvents = attachEvents(canvas!, eventContext, window);

    return () => {
      mounted = false; // $state
      detachEvents?.();
      detachEvents = null;
      observerCleanup?.disconnect();
      resizeCleanup?.clear();
      animationLoop?.stop();
      clearSimulation(simState);
      particleSystem?.clear();
      clearAnimationState(angles, speeds);
      if (hotkeysState.inactivityTimeout) clearTimeout(hotkeysState.inactivityTimeout);
      if (canvasState.duplicateWarningTimeout) clearTimeout(canvasState.duplicateWarningTimeout);
      if (canvasState.highlightedLinkTimeout) clearTimeout(canvasState.highlightedLinkTimeout);
      if (canvasState.undoToastTimeout) clearTimeout(canvasState.undoToastTimeout);
    };
  });

  // Реактивно перезапускаем симуляцию при изменении данных
  $effect(() => {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const _ = mounted; // track mounted state
    const nodesCount = nodes.length;
    const linksCount = visibleLinks.length;
    const hiddenTypesCount = graphStore.hiddenLinkTypes.length;
    const minWeight = graphStore.minLinkWeight;
    const dataKey = `${nodesCount}-${linksCount}-${hiddenTypesCount}-${minWeight}`;

    if (dataKey === lastDataKey && simState.isRunning) {
      return;
    }
    lastDataKey = dataKey;

    if (!browser || !mounted) return;

    // The particle system is created once in onMount, but the node count
    // (and therefore the performance threshold) changes as data loads/filters.
    particleSystem?.updateNodeCount(nodes.length);

    graphStable = false;

    if (nodes.length === 0) {
      clearSimulation(simState);
      if (ctx) {
        ctx.clearRect(0, 0, width, height);
      }
      return;
    }

    // Очищаем для новых данных
    clearAnimationState(angles, speeds);
    clearSimulation(simState);

    // Pin technical nodes (e.g. Knowledge Core) to fixed screen positions
    const pinnedNodes = pinTechnicalNodes(nodes);

    // Запускаем новую симуляцию
    startSimulation(
      pinnedNodes,
      visibleLinks,
      width,
      height,
      simState,
      transform,
      () => {
        redraw();
      },
      () => {
        const simNodes = getSimulationNodes(simState);
        if (ctx && simNodes.length > 0) {
          resetView(ctx, width, height, simNodes, transform);
        }
      },
      () => {
        graphStable = true;
      }
    );
  });

  // Применяем дельта-обновления инкрементально
  $effect(() => {
    if (!delta || !browser || !mounted || !simState.isRunning) {
      return;
    }

    // Apply incremental delta updates to the simulation
    applyDeltaToSimulation(delta, {
      nodes,
      links,
      width,
      height,
      state: simState,
      transform,
      onTick: () => {
        redraw();
      },
      onResetView: () => {
        const simNodes = getSimulationNodes(simState);
        if (ctx && simNodes.length > 0) {
          resetView(ctx, width, height, simNodes, transform);
        }
      },
    });
  });

  function hasFadingOpacity(): boolean {
    for (const value of simState.nodeOpacity.values()) {
      if (value < 0.999) return true;
    }
    for (const value of simState.linkOpacity.values()) {
      if (value < 0.999) return true;
    }
    return false;
  }

  function buildCacheKey(): string {
    const fog = fogState.snapshot;
    return JSON.stringify({
      w: width,
      h: height,
      data: lastDataKey,
      tx: Math.round(transform.x),
      ty: Math.round(transform.y),
      tk: transform.k.toFixed(3),
      hover: canvasState.hoveredNodeId ?? "",
      focus: canvasState.focusMode,
      highlight: canvasState.highlightedLinkId ?? "",
      search: [...(hotkeysState.searchMatchIds ?? [])].sort().join(","),
      fog: fog.enabled
        ? `${fog.mode}:${Math.round(fog.radius / 5)}:${Math.round(fog.centerX / 5)}:${Math.round(fog.centerY / 5)}`
        : "off",
      bh: `${blackHole?.hovered ? 1 : 0}:${Math.round((blackHole?.pulsePhase ?? 0) * 10)}`,
      gh: `${ghostNode?.hovered ? 1 : 0}:${Math.round((ghostNode?.pulsePhase ?? 0) * 10)}`,
      stable: stableRender,
      anim: Math.round(animationTime / 500),
    });
  }

  function canUseCache(): boolean {
    return (
      graphStable &&
      !canvasState.focusMode &&
      !canvasState.hoveredNodeId &&
      !dragDropState.draggedNodeId &&
      !dragDropState.isDraggingForLink &&
      !dragDropState.linkPreviewTarget &&
      simState.dyingLinks.length === 0 &&
      simState.dyingLinkOpacity.size === 0 &&
      !hasFadingOpacity() &&
      !particleSystem?.isEnabled()
    );
  }

  function getOffscreenContext(): CanvasRenderingContext2D {
    if (!offscreenCanvas) {
      offscreenCanvas = document.createElement("canvas");
    }
    if (offscreenCanvas.width !== width || offscreenCanvas.height !== height) {
      offscreenCanvas.width = width;
      offscreenCanvas.height = height;
      // Reset context after resize
      offscreenCtx = null;
    }
    if (!offscreenCtx) {
      offscreenCtx = offscreenCanvas.getContext("2d");
    }
    return offscreenCtx!;
  }

  function doRedraw(simNodes: SimulationNode[], hoveredNeighborIds: Set<string>) {
    if (!needsRedraw || !ctx) return;
    needsRedraw = false;

    const linkMousePos =
      dragDropState.draggedNodeId && !dragDropState.linkPreviewTarget
        ? {
            sourceId: dragDropState.draggedNodeId,
            x: dragDropState.mouseWorldPosition.x,
            y: dragDropState.mouseWorldPosition.y,
          }
        : null;

    const cacheKey = canUseCache() ? buildCacheKey() : "";
    if (cacheKey && offscreenCanvas && offscreenCtx && cacheKey === lastCacheKey) {
      ctx.drawImage(offscreenCanvas, 0, 0);
      return;
    }

    const targetCtx = cacheKey ? getOffscreenContext() : ctx;
    draw(
      targetCtx,
      width,
      height,
      simState.simLinks,
      simNodes,
      angles,
      transform,
      simState.nodeOpacity,
      simState.linkOpacity,
      simState.dyingLinks,
      simState.dyingLinkOpacity,
      stableRender,
      animationTime,
      canvasState.hoveredNodeId,
      particleSystem,
      blackHole,
      ghostNode,
      gravitySystem,
      canvasState.focusMode,
      hotkeysState.searchMatchIds,
      canvasState.highlightedLinkId,
      dragDropState.linkPreviewTarget,
      linkMousePos,
      fogState.snapshot,
      hoveredNeighborIds
    );
    drawFog(targetCtx, width, height, fogState.snapshot);

    if (cacheKey) {
      lastCacheKey = cacheKey;
      ctx.drawImage(offscreenCanvas!, 0, 0);
    }
  }

  function scheduleRedraw() {
    needsRedraw = true;
  }

  // Backwards-compatible alias: outside callers schedule a frame,
  // actual rendering happens once per animation-frame loop.
  function redraw() {
    scheduleRedraw();
  }

  const eventContext: GraphCanvasEventContext = {
    get readonly() {
      return readonly;
    },
    get browser() {
      return browser;
    },
    isTechnicalNode: (nodeId) => isTechnicalNode(nodes, nodeId),
    getCanvas: () => canvas,
    getCtx: () => ctx,
    getWidth: () => width,
    getHeight: () => height,
    transform,
    dragState,
    dragDropState,
    simState,
    hotkeysState,
    get noteFormState() {
      return noteFormState;
    },
    get linkFormState() {
      return linkFormState;
    },
    zoomPanState,
    getGhostNode: () => ghostNode,
    getBlackHole: () => blackHole,
    getGravitySystem: () => gravitySystem,
    getHoveredNodeId: () => canvasState.hoveredNodeId,
    setHoveredNodeId: (id) => {
      canvasState.hoveredNodeId = id;
    },
    getHoveredLink: () => canvasState.hoveredLink,
    setHoveredLink: (link) => {
      canvasState.hoveredLink = link;
    },
    getTooltipPosition: () => canvasState.tooltipPosition,
    setTooltipPosition: (pos) => {
      canvasState.tooltipPosition = pos;
    },
    getFocusMode: () => canvasState.focusMode,
    setFocusMode: (v) => {
      canvasState.focusMode = v;
    },
    getSelectedNodeId: () => canvasState.selectedNodeId,
    setSelectedNodeId: (id) => {
      canvasState.selectedNodeId = id;
    },
    redraw,
    toggleFocus: () => canvasState.handleToggleFocus(redraw),
    openSearch: () => canvasState.handleOpenSearch(hotkeysState),
    closeSearch: () => canvasState.handleCloseSearch(hotkeysState, redraw),
    openHelp: () => canvasState.openHelpModal(hotkeysState),
    toggleHelp: () => {
      hotkeysState.showHelpModal = !hotkeysState.showHelpModal;
      hotkeysState.showHelpTooltip = false;
    },
    setGhostNode: (node) => {
      ghostNode = node;
    },
    get onNodeClick() {
      return onNodeClick;
    },
    get onNodeContextMenu() {
      return (node: { id: string; title: string; type?: string }, x: number, y: number) => {
        contextMenu = { visible: true, x, y, node };
      };
    },
    get onNoteDelete() {
      return onNoteDelete;
    },
    get onSingularityDrop() {
      return onNoteDelete;
    },
    isOverSingularity: (clientX: number, clientY: number) => {
      if (!canvas || !blackHole) return false;
      const rect = canvas.getBoundingClientRect();
      const screenX = clientX - rect.left;
      const screenY = clientY - rect.top;
      return isPointOverBlackHole(screenX, screenY, blackHole);
    },
    setSingularityHovered: (hovered: boolean) => {
      blackHole.hovered = hovered;
    },
    getKeyLines: () => canvasState.hotkeyLines,
  };
</script>

<canvas
  bind:this={canvas}
  data-testid="graph-canvas"
  class={className}
  style="width: 100%; height: 100%; cursor: grab; background: transparent;"
></canvas>

<GraphNodeContextMenu
  x={contextMenu.x}
  y={contextMenu.y}
  visible={contextMenu.visible}
  node={contextMenu.node ?? undefined}
  onClose={() => (contextMenu = { ...contextMenu, visible: false })}
  onCreateChild={() => {
    if (contextMenu.node && onCreateChildNote) {
      onCreateChildNote(contextMenu.node);
    }
    contextMenu = { ...contextMenu, visible: false };
  }}
  onViewDetails={() => {
    if (contextMenu.node && onNodeClick) {
      onNodeClick(contextMenu.node);
      canvasState.selectedNodeId = contextMenu.node.id;
    }
    contextMenu = { ...contextMenu, visible: false };
  }}
/>

{#if showLinkTypeLegend}
  <LinkTypeLegend
    hiddenTypes={graphStore.hiddenLinkTypes}
    minWeight={graphStore.minLinkWeight}
    showMinWeight={true}
    onToggle={(type) => graphStore.toggleLinkType(type)}
    onMinWeightChange={(value) => (graphStore.minLinkWeight = value)}
  />
{/if}

<GraphCanvasModals
  activeForm={noteFormState.showNoteForm ? "note" : linkFormState.showLinkForm ? "link" : null}
  bind:noteFormState
  bind:linkFormState
  onSave={(form) =>
    form === "note"
      ? createNote(noteFormState, {
          onNoteCreate,
          onFormClose: redraw,
        })
      : createLink(linkFormState, links, {
          onLinkCreate,
          onFormClose: redraw,
          onDuplicateWarning: (source, target, linkType, x, y) =>
            canvasState.showDuplicateWarning(source, target, linkType, x, y),
        })}
  onCancel={(form) =>
    form === "note"
      ? (closeNoteForm(noteFormState), redraw())
      : (closeLinkForm(linkFormState), redraw())}
/>

<GraphCanvasOverlay
  {canvas}
  {nodes}
  hoveredNodeId={canvasState.hoveredNodeId}
  hoveredLink={canvasState.hoveredLink}
  tooltipPosition={canvasState.tooltipPosition}
  duplicateWarning={canvasState.duplicateWarning}
  focusMode={canvasState.focusMode}
  fogWarning={fogWarningState.kind}
  showUndoToast={canvasState.showUndoToast}
  undoToastStage={canvasState.undoToastStage}
  {hotkeysState}
  onCloseSearch={() => canvasState.handleCloseSearch(hotkeysState, redraw)}
  onRestoreDeletedNode={() => canvasState.restoreDeletedNode(onNoteRestore)}
  onCancelUndo={canvasState.cancelUndo}
  onUpdateSearch={() => {
    updateSearch(hotkeysState, getSimulationNodes(simState));
    redraw();
  }}
  onLinkEdit={onLinkEdit ? () => canvasState.handleLinkEdit(onLinkEdit) : undefined}
  onLinkDelete={onLinkDelete ? () => canvasState.handleLinkDelete(onLinkDelete) : undefined}
/>
{#if hotkeysState.showHelpModal}
  <HelpHotkeysModal
    hotkeyLines={canvasState.hotkeyLines}
    {helpContent}
    onClose={() => canvasState.closeHelpModal(hotkeysState)}
  />
{/if}
