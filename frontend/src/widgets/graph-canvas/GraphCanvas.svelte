<script lang="ts">
  import { onMount } from "svelte";
  import { browser } from "$app/environment";
  import type { GraphDeltaData } from "$shared/api/graph";
  import { GraphMode } from "$entities";
  import {
    GraphCanvasOverlay,
    GraphCanvasModals,
    GraphCanvasControls,
    LinkTypeLegend,
  } from "$features/graph-ui";
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
    draw,
    resetView,
    startAnimationLoop,
    clearAnimationState,
    type TransformState,
    type DragState,
    type BlackHoleState,
    createBlackHole,
    updateBlackHolePosition,
    updateBlackHolePulse,
    type GhostNodeState,
    updateGhostNodePosition,
    updateGhostNodePulse,
    type GravitySystem,
    applyDelta as applyDeltaToSimulation,
  } from "$entities/graph-canvas/lib";
  import { createGhostNode } from "$entities/graph-canvas/lib/ghost-node";
  import { createGravitySystem } from "$entities/graph-canvas/lib/gravity-system";

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

  const {
    nodes,
    links,
    onNodeClick,
    onLinkEdit,
    onLinkDelete,
    onNoteCreate,
    onLinkCreate,
    onNoteDelete,
    onNoteRestore,
    helpContent,
    delta,
    disableVariation = false,
    readonly = false,
    className = "",
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
    helpContent?: string;
    delta?: GraphDeltaData;
    disableVariation?: boolean;
    readonly?: boolean;
    className?: string;
  } = $props();

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

  // Filter links based on selected types and minimum weight
  const visibleLinks = $derived(
    links.filter((l) => {
      const typeMatch =
        graphStore.selectedLinkTypes.length === 0 ||
        graphStore.selectedLinkTypes.includes(l.link_type ?? "related");
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

  // Hotkeys state (FSD)
  const hotkeysState: HotkeysState = $state(createHotkeysState());

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

    // Запускаем анимацию
    animationLoop = startAnimationLoop(
      () => getSimulationNodes(simState),
      () => {
        const simNodes = getSimulationNodes(simState);
        if (ctx) {
          // In stable render mode keep animation time fixed for deterministic screenshots
          if (stableRender) {
            animationTime = 0;
          } else {
            animationTime = performance.now();
          }

          // Update interactive element positions and pulses
          updateBlackHolePosition(blackHole, width, height);
          updateBlackHolePulse(blackHole, animationTime);
          updateGhostNodePosition(ghostNode, width, height, nodes);
          updateGhostNodePulse(ghostNode, animationTime);

          // Apply subtle gravity attraction only when not taking stable screenshots
          if (!stableRender && gravitySystem.isEnabled(simNodes.length)) {
            gravitySystem.applyAttraction(simNodes);
          }

          // Draw the full graph with all effects
          redraw();
        }
      },
      stableRender
    );

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
    const selectedTypesCount = graphStore.selectedLinkTypes.length;
    const minWeight = graphStore.minLinkWeight;
    const dataKey = `${nodesCount}-${linksCount}-${selectedTypesCount}-${minWeight}`;

    if (dataKey === lastDataKey && simState.isRunning) {
      return;
    }
    lastDataKey = dataKey;

    if (!browser || !mounted) return;

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

  function redraw() {
    const simNodes = getSimulationNodes(simState);
    if (ctx) {
      const linkMousePos =
        dragDropState.draggedNodeId && !dragDropState.linkPreviewTarget
          ? {
              sourceId: dragDropState.draggedNodeId,
              x: dragDropState.mouseWorldPosition.x,
              y: dragDropState.mouseWorldPosition.y,
            }
          : null;
      draw(
        ctx,
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
        linkMousePos
      );
    }
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
    get onNoteDelete() {
      return onNoteDelete;
    },
    getKeyLines: () => canvasState.hotkeyLines,
  };
</script>

<canvas
  bind:this={canvas}
  data-testid="graph-canvas"
  class={className}
  style="width: 100%; height: 100%; cursor: grab; background: linear-gradient(145deg, #0a1a3a, #020617);"
></canvas>

<GraphCanvasControls
  mode={GraphMode.fromFocus(canvasState.focusMode)}
  onReset={() => {
    const simNodes = getSimulationNodes(simState);
    if (ctx && simNodes.length > 0) {
      resetView(ctx, width, height, simNodes, transform);
    }
  }}
  onSearch={() => canvasState.handleOpenSearch(hotkeysState)}
  onToggleMode={() => canvasState.handleToggleFocus(redraw)}
  onToggleFocus={() => canvasState.handleToggleFocus(redraw)}
/>

<LinkTypeLegend
  selectedTypes={graphStore.selectedLinkTypes}
  minWeight={graphStore.minLinkWeight}
  showMinWeight={true}
  onToggle={(type) => graphStore.toggleLinkType(type)}
  onMinWeightChange={(value) => (graphStore.minLinkWeight = value)}
/>

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
  {links}
  loading={!graphStable && nodes.length > 0}
  hoveredNodeId={canvasState.hoveredNodeId}
  hoveredLink={canvasState.hoveredLink}
  tooltipPosition={canvasState.tooltipPosition}
  duplicateWarning={canvasState.duplicateWarning}
  focusMode={canvasState.focusMode}
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
