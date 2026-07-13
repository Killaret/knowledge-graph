<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import type { GraphDelta } from '$lib/api/graph';
  import LinkTooltip from '$lib/components/LinkTooltip.svelte';
  import GraphTooltip from '$lib/components/GraphTooltip.svelte';
  import HelpHotkeysModal from '$lib/components/HelpHotkeysModal.svelte';
  import { ParticleSystem } from './GraphCanvas/particle-system';
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
    findLinkAtPosition,
    startAnimationLoop,
    clearAnimationState,
    type TransformState,
    type DragState,
    type SimulationNode,
    applyDelta as applyDeltaToSimulation,
    createBlackHole,
    updateBlackHolePosition,
    updateBlackHolePulse,
    isNodeOverBlackHole,
    isPointOverBlackHole,
    type BlackHoleState,
    updateGhostNodePosition,
    updateGhostNodePulse,
    type GhostNodeState,
    type GravitySystem
  } from './GraphCanvas';
  import { createGhostNode } from './GraphCanvas/ghost-node';
  import { createGravitySystem } from './GraphCanvas/gravity-system';
  
  // FSD imports
  import {
    createDragDropState,
    getMouseWorldPosition,
    findNodeAtPosition,
    handleMouseDown,
    type DragDropState
  } from '$features/graph-interaction/drag-and-drop';
  import {
    createHotkeysState,
    handleKeyDownEvent,
    updateSearch as updateSearchUtil,
    resetInactivityTimer,
    updateActivity,
    showRandomTip as showRandomTipUtil,
    type HotkeysState
  } from '$features/graph-interaction/hotkeys';
  import {
    createZoomPanState,
    handleZoom as handleZoomPan,
    handleTouchStart,
    type ZoomPanState
  } from '$features/graph-interaction/zoom-pan';
  import {
    createNoteFormState,
    openNoteForm,
    closeNoteForm as closeNoteFormUtil,
    createNote as createNoteUtil,
    type NoteFormState
  } from '$features/graph-forms/note-form';
  import {
    createLinkFormState,
    openLinkForm,
    closeLinkForm as closeLinkFormUtil,
    createLink as createLinkUtil,
    type LinkFormState
  } from '$features/graph-forms/link-form';

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
    disableVariation = false
  }: {
    nodes: Array<{ id: string; title: string; type?: string; createdAt?: string; created_at?: string }>;
    links: Array<{ source: string; target: string; weight?: number; link_type?: string; source_type?: string }>;
    onNodeClick?: (node: { id: string; title: string; type?: string }) => void;
    onLinkEdit?: (link: { source: string; target: string; link_type: string; weight: number }) => void;
    onLinkDelete?: (link: { source: string; target: string; link_type: string }) => void;
    onNoteCreate?: (data: { title: string; content: string; type: string }) => void;
    onLinkCreate?: (link: { source: string; target: string; link_type: string; weight: number }) => void;
    onNoteDelete?: (nodeId: string) => void;
    onNoteRestore?: (nodeId: string) => void;
    helpContent?: string;
    delta?: GraphDelta;
    disableVariation?: boolean;
  } = $props();

  let stableRender = false;
  $effect(() => {
    stableRender = disableVariation || (browser && new URL(window.location.href).searchParams.get('stableRender') === 'true');
  });

  // Debug: проверяем типы узлов при изменении (dev only)
  $effect(() => {
    if (import.meta.env.DEV && nodes.length > 0) {
      const types = nodes.map(n => n.type || 'undefined');
      const uniqueTypes = [...new Set(types)];
      if (import.meta.env.DEV) { console.log('[GraphCanvas] Received nodes types:', uniqueTypes, 'Total:', nodes.length) };
      if (import.meta.env.DEV) { console.log('[GraphCanvas] First node:', nodes[0]) };
    }
  });

  let canvas: HTMLCanvasElement | null = null;
  let ctx: CanvasRenderingContext2D | null = null;
  let width = 800;
  let height = 600;
  let animationLoop: { stop: () => void } | null = null;
  let resizeCleanup: { clear: () => void } | null = null;
  let observerCleanup: { disconnect: () => void } | null = null;
  const angles = new Map<string, number>();
  const speeds = new Map<string, number>();

  // NOTE: transform and dragState need reactivity for Svelte bindings,
  // but simState must NOT be $state — d3-force mutates link objects (source/target become node refs)
  // and Svelte 5 Proxy intercepts those mutations, breaking d3 internals.
  const transform: TransformState = $state({ x: 0, y: 0, k: 1 });
  const dragState: DragState = $state({ dragging: false, dragStart: { x: 0, y: 0 } });
  const simState: SimulationState = {
    simulation: null,
    simLinks: [],
    isRunning: false,
    nodeOpacity: new Map(),
    linkOpacity: new Map(),
    fadeAnimationId: null
  };

  // Для отслеживания изменений данных по содержимому (не по ссылке)
  let lastDataKey = '';
  let mounted = $state(false);

  // Используем утилиты для resize
  const resizeState = { width, height };

  // Double-tap zoom state (FSD)
  const zoomPanState: ZoomPanState = $state(createZoomPanState());

  // Link tooltip state
  let hoveredLink: { source: string; target: string; link_type: string; weight: number; source_type: string } | null = $state(null);
  let tooltipPosition = $state({ x: 0, y: 0 });
  let hoveredNodeId: string | null = $state(null);
  
  // Particle system
  let particleSystem: ParticleSystem | null = $state(null);
  
  // Time for animations
  let animationTime = $state(0);

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

  // Selection state for keyboard delete
  let selectedNodeId: string | null = $state(null);

  // Note creation form state (FSD)
  const noteFormState: NoteFormState = $state(createNoteFormState());

  // Link creation form state (FSD)
  const linkFormState: LinkFormState = $state(createLinkFormState());

  // Focus mode: hides decorative effects when true
  let focusMode = $state(false);

  // Hotkeys state (FSD)
  const hotkeysState: HotkeysState = $state(createHotkeysState());
  let searchInput: HTMLInputElement | null = null;

  // Duplicate link warning state
  let duplicateWarning = $state<{ message: string; x: number; y: number; linkId: string } | null>(null);
  let duplicateWarningTimeout: ReturnType<typeof setTimeout> | null = null;
  let highlightedLinkId: string | null = $state(null);
  let highlightedLinkTimeout: ReturnType<typeof setTimeout> | null = null;

  onMount(() => {
    if (!browser || !canvas) return;
    
    // SSR-safe: получаем контекст canvas
    ctx = canvas.getContext('2d')!;
    
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
          // Update animation time
          animationTime = performance.now();
          
          // Update interactive element positions and pulses
          updateBlackHolePosition(blackHole, width, height);
          updateBlackHolePulse(blackHole, animationTime);
          updateGhostNodePosition(ghostNode, width, height, nodes);
          updateGhostNodePulse(ghostNode, animationTime);

          // Apply subtle gravity attraction
          if (gravitySystem.isEnabled(simNodes.length)) {
            gravitySystem.applyAttraction(simNodes);
          }
          
          // Draw the full graph with all effects
          const linkMousePos = (dragDropState.draggedNodeId && !dragDropState.linkPreviewTarget)
            ? { sourceId: dragDropState.draggedNodeId, x: dragDropState.mouseWorldPosition.x, y: dragDropState.mouseWorldPosition.y }
            : null;
          draw(ctx, width, height, simState.simLinks, simNodes, angles, transform, simState.nodeOpacity, simState.linkOpacity, stableRender, animationTime, hoveredNodeId, particleSystem, blackHole, ghostNode, gravitySystem, focusMode, hotkeysState.searchMatchIds, highlightedLinkId, dragDropState.linkPreviewTarget, linkMousePos);
        }
      },
      stableRender
    );
    
    mounted = true; // triggers $effect re-run since it's $state

    resetInactivityTimer(hotkeysState, () => showRandomTipUtil(hotkeysState, hotkeyLines));

    return () => {
      mounted = false; // $state
      observerCleanup?.disconnect();
      resizeCleanup?.clear();
      animationLoop?.stop();
      clearSimulation(simState);
      particleSystem?.clear();
      clearAnimationState(angles, speeds);
      if (hotkeysState.inactivityTimeout) clearTimeout(hotkeysState.inactivityTimeout);
    };
  });

  // Реактивно перезапускаем симуляцию при изменении данных
  $effect(() => {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    const _ = mounted; // track mounted state
    const nodesCount = nodes.length;
    const linksCount = links.length;
    const dataKey = `${nodesCount}-${linksCount}`;

    if (dataKey === lastDataKey && simState.isRunning) {
      return;
    }
    lastDataKey = dataKey;

    if (!browser || !mounted) return;

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
      links,
      width,
      height,
      simState,
      transform,
      () => {
        const simNodes = getSimulationNodes(simState);
        if (ctx && simNodes.length > 0) {
          const linkMousePos2 = (dragDropState.draggedNodeId && !dragDropState.linkPreviewTarget)
            ? { sourceId: dragDropState.draggedNodeId, x: dragDropState.mouseWorldPosition.x, y: dragDropState.mouseWorldPosition.y }
            : null;
          draw(ctx, width, height, simState.simLinks, simNodes, angles, transform, simState.nodeOpacity, simState.linkOpacity, stableRender, animationTime, hoveredNodeId, particleSystem, blackHole, ghostNode, gravitySystem, focusMode, hotkeysState.searchMatchIds, highlightedLinkId, dragDropState.linkPreviewTarget, linkMousePos2);
        }
      },
      () => {
        const simNodes = getSimulationNodes(simState);
        if (ctx && simNodes.length > 0) {
          resetView(ctx, width, height, simNodes, transform);
        }
      }
    );
  });

  // Применяем дельта-обновления инкрементально
  $effect(() => {
    if (!delta || !browser || !mounted || !simState.isRunning) {
      return;
    }

    console.log('[GraphCanvas] Applying delta:', delta);

    // Используем delta модуль для применения обновлений
    applyDeltaToSimulation(delta, {
      nodes,
      links,
      width,
      height,
      state: simState,
      transform,
      onTick: () => {
        const simNodes = getSimulationNodes(simState);
        if (ctx && simNodes.length > 0) {
          const linkMousePos3 = (dragDropState.draggedNodeId && !dragDropState.linkPreviewTarget)
            ? { sourceId: dragDropState.draggedNodeId, x: dragDropState.mouseWorldPosition.x, y: dragDropState.mouseWorldPosition.y }
            : null;
          draw(ctx, width, height, simState.simLinks, simNodes, angles, transform, simState.nodeOpacity, simState.linkOpacity, stableRender, animationTime, hoveredNodeId, particleSystem, blackHole, ghostNode, gravitySystem, focusMode, hotkeysState.searchMatchIds, highlightedLinkId, dragDropState.linkPreviewTarget, linkMousePos3);
        }
      },
      onResetView: () => {
        const simNodes = getSimulationNodes(simState);
        if (ctx && simNodes.length > 0) {
          resetView(ctx, width, height, simNodes, transform);
        }
      }
    });
  });



  function redraw() {
    const simNodes = getSimulationNodes(simState);
    if (ctx) {
      const linkMousePos = (dragDropState.draggedNodeId && !dragDropState.linkPreviewTarget)
        ? { sourceId: dragDropState.draggedNodeId, x: dragDropState.mouseWorldPosition.x, y: dragDropState.mouseWorldPosition.y }
        : null;
      draw(ctx, width, height, simState.simLinks, simNodes, angles, transform, simState.nodeOpacity, simState.linkOpacity, stableRender, animationTime, hoveredNodeId, particleSystem, blackHole, ghostNode, gravitySystem, focusMode, hotkeysState.searchMatchIds, highlightedLinkId, dragDropState.linkPreviewTarget, linkMousePos);
    }
  }

  function onZoom(e: WheelEvent) {
    updateActivity(hotkeysState, () => showRandomTipUtil(hotkeysState, hotkeyLines));
    handleZoomPan(e, transform, canvas!, redraw);
  }

  function onMouseDown(e: MouseEvent) {
    updateActivity(hotkeysState, () => showRandomTipUtil(hotkeysState, hotkeyLines));
    handleMouseDown(
      e,
      canvas!,
      transform,
      dragState,
      dragDropState,
      getSimulationNodes(simState),
      ghostNode,
      isTechnicalNode,
      {
        onNodeDragStart: (nodeId) => {
          if (nodeId === 'ghost') {
            openNoteForm(noteFormState, e.clientX, e.clientY);
          }
        }
      }
    );
    // Fix dragged node in place so simulation doesn't move it while drawing link preview
    if (dragDropState.draggedNodeId) {
      const node = getSimulationNodes(simState).find((n) => n.id === dragDropState.draggedNodeId);
      if (node && node.x != null && node.y != null) {
        node.fx = node.x;
        node.fy = node.y;
      }
    }
  }

  function onMouseMove(e: MouseEvent) {
    updateActivity(hotkeysState, () => showRandomTipUtil(hotkeysState, hotkeyLines));
    const pos = getMouseWorldPosition(e, canvas!, transform);
    dragDropState.mouseWorldPosition = pos;

    // Update hover states for interactive elements
    blackHole.hovered = isPointOverBlackHole(e.clientX, e.clientY, blackHole, transform);
    // Ghost node is now drawn in screen coords at (60, 60)
    const canvasRect = canvas?.getBoundingClientRect();
    const screenX = canvasRect ? e.clientX - canvasRect.left : e.clientX;
    const screenY = canvasRect ? e.clientY - canvasRect.top : e.clientY;
    const dx = screenX - 60, dy = screenY - 60;
    ghostNode.hovered = Math.sqrt(dx * dx + dy * dy) < ghostNode.radius;

    // Dragging a node — node stays fixed, we draw a preview line to mouse position
    if (dragDropState.draggedNodeId && dragState.dragging) {
      const node = getSimulationNodes(simState).find((n) => n.id === dragDropState.draggedNodeId);
      if (node && node.x != null && node.y != null) {
        // Do NOT move the node — keep it at original position
        // Check if over black hole (use original node position)
        blackHole.hovered = isNodeOverBlackHole(node, blackHole);

        // Check if under cursor there is another node to link to
        const targetNode = findNodeAtPosition(pos.x, pos.y, getSimulationNodes(simState));
        if (targetNode && targetNode.id !== dragDropState.draggedNodeId && !isTechnicalNode(targetNode.id)) {
          dragDropState.isDraggingForLink = true;
          dragDropState.linkTargetNodeId = targetNode.id;
          // Preview from source node to target node position
          dragDropState.linkPreviewTarget = { sourceId: dragDropState.draggedNodeId, targetId: targetNode.id };
        } else {
          dragDropState.isDraggingForLink = false;
          dragDropState.linkTargetNodeId = null;
          // Preview from source node to current mouse position (stored in mouseWorldPosition)
          dragDropState.linkPreviewTarget = null;
        }
      }
      redraw();
      return;
    }

    // Panning
    if (dragState.dragging) {
      transform.x = e.clientX - dragState.dragStart.x;
      transform.y = e.clientY - dragState.dragStart.y;
      redraw();
      return;
    }

    // Hover detection (only when not dragging)
    const hovered = findLinkAtPosition(pos.x, pos.y, simState.simLinks, getSimulationNodes(simState), transform);

    let foundHoveredNode = false;
    let hoveredTechnicalNode = null;
    const simNodes = getSimulationNodes(simState);
    for (const node of simNodes) {
      if (node.x && node.y) {
        const dx = pos.x - node.x;
        const dy = pos.y - node.y;
        if (Math.sqrt(dx * dx + dy * dy) < 30) {
          hoveredNodeId = node.id;
          foundHoveredNode = true;
          if (node.type === 'technical') {
            hoveredTechnicalNode = node;
          }
          break;
        }
      }
    }
    if (!foundHoveredNode) {
      hoveredNodeId = null;
    }

    if (hoveredTechnicalNode) {
      hotkeysState.helpTooltipMessage = 'Click to open help, or press ?';
      hotkeysState.helpTooltipPosition = { x: e.clientX, y: e.clientY - 10 };
      hotkeysState.showHelpTooltip = true;
    } else if (hotkeysState.showHelpTooltip && hotkeysState.helpTooltipMessage === 'Click to open help, or press ?') {
      hotkeysState.showHelpTooltip = false;
    }

    if (hovered) {
      const sourceNode = typeof hovered.source === 'string'
        ? simNodes.find((n) => n.id === hovered.source)
        : hovered.source;
      const targetNode = typeof hovered.target === 'string'
        ? simNodes.find((n) => n.id === hovered.target)
        : hovered.target;

      if (sourceNode && targetNode) {
        hoveredLink = {
          source: typeof hovered.source === 'string' ? hovered.source : (hovered.source as any).id,
          target: typeof hovered.target === 'string' ? hovered.target : (hovered.target as any).id,
          link_type: hovered.link_type || 'related',
          weight: hovered.weight ?? 0.5,
          source_type: (hovered as any).source_type || 'user'
        };
        const centerX = (sourceNode.x! + targetNode.x!) / 2;
        const centerY = (sourceNode.y! + targetNode.y!) / 2;
        tooltipPosition = {
          x: centerX * transform.k + transform.x + 10,
          y: centerY * transform.k + transform.y + 10
        };
      }
    } else {
      hoveredLink = null;
    }
  }

  function onMouseUp(e: MouseEvent) {
    updateActivity(hotkeysState, () => showRandomTipUtil(hotkeysState, hotkeyLines));
    const pos = getMouseWorldPosition(e, canvas!, transform);
    const wasDraggingNode = dragDropState.draggedNodeId !== null;

    if (dragDropState.draggedNodeId) {
      const node = getSimulationNodes(simState).find((n) => n.id === dragDropState.draggedNodeId);
      if (node) {
        // Release the fixed position so simulation can resume
        node.fx = undefined;
        node.fy = undefined;

        // Node never moved — check what's under the cursor on drop
        const targetNode = findNodeAtPosition(pos.x, pos.y, getSimulationNodes(simState));
        if (targetNode && targetNode.id !== dragDropState.draggedNodeId && !isTechnicalNode(targetNode.id)) {
          // Dropped on another node -> open link form
          openLinkForm(linkFormState, dragDropState.draggedNodeId, targetNode.id, e.clientX, e.clientY);
        }
      }
      dragDropState.draggedNodeId = null;
      dragDropState.isDraggingForLink = false;
      dragDropState.linkPreviewTarget = null;
    }

    dragState.dragging = false;
    canvas!.style.cursor = 'grab';
    blackHole.hovered = false;

    // Ghost node click up (without drag) - already handled in mousedown
    if (!wasDraggingNode && !ghostNode.hovered) {
      const clickedNode = findNodeAtPosition(pos.x, pos.y, getSimulationNodes(simState));
      if (!clickedNode) {
        hoveredLink = null;
      }
    }
  }

  function onClick(e: MouseEvent) {
    updateActivity(hotkeysState, () => showRandomTipUtil(hotkeysState, hotkeyLines));
    const pos = getMouseWorldPosition(e, canvas!, transform);
    const clickedNode = findNodeAtPosition(pos.x, pos.y, getSimulationNodes(simState));
    if (clickedNode) {
      if (isTechnicalNode(clickedNode.id)) {
        openHelpModal();
        return;
      }
      selectedNodeId = clickedNode.id;
      if (onNodeClick) {
        onNodeClick({ id: clickedNode.id, title: clickedNode.title, type: clickedNode.type });
      }
    } else {
      // Deselect when clicking on empty space
      selectedNodeId = null;
    }
  }

  // Double-click to zoom in (desktop); double-click on empty space zooms in at cursor
  function onDblClick(e: MouseEvent) {
    if (!canvas) return;
    const rect = canvas.getBoundingClientRect();
    const mouseX = e.clientX - rect.left;
    const mouseY = e.clientY - rect.top;
    const zoomFactor = 1.8;
    const newScale = Math.min(transform.k * zoomFactor, 5);
    const scaleChange = newScale / transform.k;
    transform.x = mouseX - (mouseX - transform.x) * scaleChange;
    transform.y = mouseY - (mouseY - transform.y) * scaleChange;
    transform.k = newScale;
    redraw();
  }

  // Global mouseup — cancels pan/drag even if mouse leaves canvas
  function onWindowMouseUp(e: MouseEvent) {
    if (!dragState.dragging) return;
    onMouseUp(e);
  }

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  function animateNodeDeletion(node: SimulationNode, onComplete: () => void) {
    const startX = node.x!;
    const startY = node.y!;
    const startTime = performance.now();
    const duration = 300;

    function step() {
      const elapsed = performance.now() - startTime;
      const t = Math.min(elapsed / duration, 1);
      const ease = t * t; // accelerate toward center

      node.x = startX + (blackHole.x - startX) * ease;
      node.y = startY + (blackHole.y - startY) * ease;
      node.scale = 1 - ease;
      node.opacity = 1 - ease;

      redraw();

      if (t < 1) {
        requestAnimationFrame(step);
      } else {
        onComplete();
      }
    }

    requestAnimationFrame(step);
  }

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  function animateSpringBack(node: SimulationNode, target: { x: number; y: number }) {
    const startX = node.x!;
    const startY = node.y!;
    const startTime = performance.now();
    const duration = 300;

    function step() {
      const elapsed = performance.now() - startTime;
      const t = Math.min(elapsed / duration, 1);
      const ease = 1 - Math.pow(1 - t, 3);

      node.x = startX + (target.x - startX) * ease;
      node.y = startY + (target.y - startY) * ease;
      node.fx = node.x;
      node.fy = node.y;

      redraw();

      if (t < 1) {
        requestAnimationFrame(step);
      } else {
        node.fx = undefined;
        node.fy = undefined;
      }
    }

    requestAnimationFrame(step);
  }

  function handleCreateNote() {
    createNoteUtil(noteFormState, {
      onNoteCreate: (data) => {
        if (onNoteCreate) {
          onNoteCreate(data);
        }
      },
      onFormClose: () => {
        redraw();
      }
    });
  }

  function handleNoteFormClose() {
    closeNoteFormUtil(noteFormState);
    redraw();
  }

  function handleCreateLink() {
    createLinkUtil(linkFormState, links, {
      onLinkCreate: (link) => {
        if (onLinkCreate) {
          onLinkCreate(link);
        }
      },
      onFormClose: () => {
        redraw();
      },
      onDuplicateWarning: (source, target, linkType, x, y) => {
        showDuplicateWarning(source, target, linkType, x, y);
      }
    });
  }

  function handleLinkFormClose() {
    closeLinkFormUtil(linkFormState);
    redraw();
  }

  function handleLinkEdit() {
    if (hoveredLink && onLinkEdit) {
      onLinkEdit({
        source: hoveredLink.source,
        target: hoveredLink.target,
        link_type: hoveredLink.link_type,
        weight: hoveredLink.weight
      });
    }
    hoveredLink = null;
  }

  function handleLinkDelete() {
    if (hoveredLink && onLinkDelete) {
      onLinkDelete({
        source: hoveredLink.source,
        target: hoveredLink.target,
        link_type: hoveredLink.link_type
      });
    }
    hoveredLink = null;
  }


  // Double-tap zoom handler
  function onTouchStart(e: TouchEvent) {
    if (!browser) return; // Skip in server-side rendering
    updateActivity(hotkeysState, () => showRandomTipUtil(hotkeysState, hotkeyLines));
    handleTouchStart(e, zoomPanState, transform, canvas!, getSimulationNodes(simState), ctx, width, height);
  }

  // Keyboard shortcuts: Esc toggles focus mode, F opens search, ? opens help, N for ghost node, Del/Backspace for delete, Ctrl+Z for undo
  function handleKeyDown(e: KeyboardEvent) {
    handleKeyDownEvent(
      e,
      hotkeysState,
      canvas,
      transform,
      getSimulationNodes(simState),
      ghostNode,
      selectedNodeId,
      noteFormState.showNoteForm,
      linkFormState.showLinkForm,
      searchInput,
      {
        onFocusModeToggle: () => {
          focusMode = !focusMode;
          redraw();
        },
        onSearchOpen: () => {
          hotkeysState.showSearchBox = true;
          hotkeysState.searchQuery = '';
          hotkeysState.searchMatchIds = [];
          hotkeysState.searchCurrentIndex = 0;
        },
        onSearchClose: () => {
          hotkeysState.showSearchBox = false;
          hotkeysState.searchQuery = '';
          hotkeysState.searchMatchIds = [];
          hotkeysState.searchCurrentIndex = 0;
          redraw();
        },
        onHelpToggle: () => {
          hotkeysState.showHelpModal = !hotkeysState.showHelpModal;
        },
        onGhostNodeCreate: () => {
          if (canvas) {
            const rect = canvas.getBoundingClientRect();
            const centerX = (rect.width / 2 - transform.x) / transform.k;
            const centerY = (rect.height / 2 - transform.y) / transform.k;
            ghostNode = createGhostNode(rect.width, rect.height, getSimulationNodes(simState));
            openNoteForm(noteFormState, centerX, centerY);
            redraw();
          }
        },
        onNodeDelete: (nodeId) => {
          if (onNoteDelete) {
            onNoteDelete(nodeId);
            selectedNodeId = null;
            redraw();
          }
        },
        onUndo: () => {
          console.log('[GraphCanvas] Undo not yet implemented');
        }
      }
    );
  }

  function handleUpdateSearch() {
    updateSearchUtil(hotkeysState, getSimulationNodes(simState));
    redraw();
  }

  function showDuplicateWarning(source: string, target: string, linkType: string, x: number, y: number) {
    const stableLinkId = `${source}-${target}-${linkType}`;
    highlightedLinkId = stableLinkId;
    duplicateWarning = { message: 'This link already exists', x, y, linkId: stableLinkId };

    if (duplicateWarningTimeout) clearTimeout(duplicateWarningTimeout);
    if (highlightedLinkTimeout) clearTimeout(highlightedLinkTimeout);

    duplicateWarningTimeout = setTimeout(() => {
      duplicateWarning = null;
    }, 2000);

    highlightedLinkTimeout = setTimeout(() => {
      highlightedLinkId = null;
      redraw();
    }, 1000);
  }

  // Undo deletion integration - two-stage toast
  let lastDeletedNodeId: string | null = $state(null);
  let showUndoToast = $state(false);
  let undoToastStage = $state<'done' | 'restore'>('done');
  let undoToastTimeout: ReturnType<typeof setTimeout> | null = null;

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  function showUndoToastFor(nodeId: string) {
    lastDeletedNodeId = nodeId;
    showUndoToast = true;
    undoToastStage = 'done';
    if (undoToastTimeout) clearTimeout(undoToastTimeout);

    // Stage 1: "Done" for 1.5s
    undoToastTimeout = setTimeout(() => {
      undoToastStage = 'restore';
      // Stage 2: "Restore" for 5s
      undoToastTimeout = setTimeout(() => {
        showUndoToast = false;
        lastDeletedNodeId = null;
        undoToastStage = 'done';
      }, 5000);
    }, 1500);
  }

  function restoreDeletedNode() {
    if (lastDeletedNodeId && onNoteRestore) {
      onNoteRestore(lastDeletedNodeId);
    }
    showUndoToast = false;
    lastDeletedNodeId = null;
    undoToastStage = 'done';
    if (undoToastTimeout) clearTimeout(undoToastTimeout);
  }

  function cancelUndo() {
    showUndoToast = false;
    lastDeletedNodeId = null;
    undoToastStage = 'done';
    if (undoToastTimeout) clearTimeout(undoToastTimeout);
  }

  function isTechnicalNode(nodeId: string): boolean {
    return nodes.some((n) => n.id === nodeId && n.type === 'technical');
  }

  function pinTechnicalNodes(
    nodeList: Array<{ id: string; title: string; type?: string; createdAt?: string; created_at?: string }>
  ): Array<{ id: string; title: string; type?: string; createdAt?: string; created_at?: string; x?: number; y?: number; fx?: number; fy?: number }> {
    return nodeList.map((n) => {
      if (n.type === 'technical') {
        const padding = 60;
        return {
          ...n,
          x: padding,
          y: padding,
          fx: padding,
          fy: padding
        };
      }
      return n;
    });
  }

  // Help content and inactivity tips
  const hotkeyLines = [
    'F — search nodes by name',
    'Esc — toggle focus mode (hide effects)',
    '? — show/hide this help',
    'N — create ghost node at center',
    'Delete/Backspace — delete selected node',
    'Ctrl+Z — undo (coming soon)',
    'Ctrl+Shift+N — quick capture a new note',
    'Drag node to another node — create a link',
    'Drag node to black hole — delete note',
    'Double-click empty space — create new note',
    'Mouse wheel — zoom in/out'
  ];

  function openHelpModal() {
    hotkeysState.showHelpModal = true;
    hotkeysState.showHelpTooltip = false;
  }

  function closeHelpModal() {
    hotkeysState.showHelpModal = false;
  }

  function handleCloseSearch() {
    hotkeysState.showSearchBox = false;
    redraw();
  }</script>

<svelte:window onkeydown={handleKeyDown} onmouseup={onWindowMouseUp} />

<canvas
  bind:this={canvas}
  data-testid="graph-canvas"
  onmousedown={onMouseDown}
  onmousemove={onMouseMove}
  onmouseup={onMouseUp}
  ondblclick={onDblClick}
  onclick={onClick}
  ontouchstart={onTouchStart}
  onwheel={onZoom}
  style="width: 100%; height: 100%; cursor: grab; background: linear-gradient(145deg, #0a1a3a, #020617);"
></canvas>

{#if hoveredNodeId}
  {@const hoveredNode = nodes.find(n => n.id === hoveredNodeId)}
  {#if hoveredNode}
    <GraphTooltip target={canvas} />
  {/if}
{/if}

{#if hoveredLink}
  {@const currentLink = hoveredLink}
  {@const sourceNode = nodes.find(n => n.id === currentLink.source)}
  {@const targetNode = nodes.find(n => n.id === currentLink.target)}
  <LinkTooltip
    visible={true}
    x={tooltipPosition.x}
    y={tooltipPosition.y}
    linkType={currentLink.link_type}
    weight={currentLink.weight}
    sourceType={currentLink.source_type}
    sourceTitle={sourceNode?.title || 'Unknown'}
    targetTitle={targetNode?.title || 'Unknown'}
    onEdit={onLinkEdit ? handleLinkEdit : undefined}
    onDelete={onLinkDelete ? handleLinkDelete : undefined}
  />
{/if}

{#if noteFormState.showNoteForm}
  <div
    class="note-form"
    style="position: absolute; left: {noteFormState.noteFormPosition.x}px; top: {noteFormState.noteFormPosition.y}px; background: rgba(10, 26, 58, 0.98); border: 1px solid rgba(138, 43, 226, 0.6); border-radius: 12px; padding: 20px; min-width: 320px; box-shadow: 0 8px 32px rgba(0, 0, 0, 0.6); z-index: 100; backdrop-filter: blur(12px);"
  >
    <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px;">
      <h3 style="margin: 0; color: #a78bfa; font-size: 16px; font-weight: 600;">Create New Note</h3>
      <button
        onclick={handleNoteFormClose}
        style="background: none; border: none; color: rgba(255,255,255,0.6); font-size: 20px; cursor: pointer; padding: 4px 8px; border-radius: 4px; transition: all 0.2s;"
        aria-label="Close"
      >
        ×
      </button>
    </div>
    <input
      type="text"
      placeholder="Title"
      bind:value={noteFormState.newNoteTitle}
      style="width: 100%; padding: 12px; margin-bottom: 12px; border: 1px solid rgba(255,255,255,0.2); border-radius: 8px; background: rgba(0,0,0,0.4); color: white; box-sizing: border-box; font-size: 14px; transition: border-color 0.2s;"
      onkeydown={(e) => e.key === 'Enter' && handleCreateNote()}
    />
    <textarea
      placeholder="Content (optional)"
      bind:value={noteFormState.newNoteContent}
      style="width: 100%; padding: 12px; margin-bottom: 12px; border: 1px solid rgba(255,255,255,0.2); border-radius: 8px; background: rgba(0,0,0,0.4); color: white; min-height: 100px; box-sizing: border-box; font-size: 14px; resize: vertical; transition: border-color 0.2s;"
    ></textarea>
    <select
      bind:value={noteFormState.newNoteType}
      style="width: 100%; padding: 12px; margin-bottom: 16px; border: 1px solid rgba(255,255,255,0.2); border-radius: 8px; background: rgba(0,0,0,0.4); color: white; font-size: 14px; cursor: pointer; transition: border-color 0.2s;"
    >
      <option value="star">⭐ Star</option>
      <option value="planet">🪐 Planet</option>
      <option value="comet">☄️ Comet</option>
      <option value="galaxy">🌀 Galaxy</option>
      <option value="asteroid">🌑 Asteroid</option>
    </select>
    <div style="display: flex; gap: 12px; justify-content: flex-end;">
      <button 
        onclick={handleNoteFormClose} 
        style="padding: 10px 20px; border: 1px solid rgba(255,255,255,0.3); border-radius: 8px; background: transparent; color: white; cursor: pointer; font-size: 14px; transition: all 0.2s;"
      >
        Cancel
      </button>
      <button 
        onclick={handleCreateNote} 
        style="padding: 10px 20px; border: none; border-radius: 8px; background: linear-gradient(135deg, #8b5cf6, #6366f1); color: white; cursor: pointer; font-size: 14px; font-weight: 500; transition: all 0.2s; box-shadow: 0 4px 12px rgba(139, 92, 246, 0.3);"
      >
        Create
      </button>
    </div>
  </div>
{/if}

{#if linkFormState.showLinkForm}
  <div
    class="link-form"
    style="position: absolute; left: {linkFormState.linkFormPosition.x}px; top: {linkFormState.linkFormPosition.y}px; background: rgba(10, 26, 58, 0.98); border: 1px solid rgba(255, 204, 0, 0.6); border-radius: 12px; padding: 20px; min-width: 300px; box-shadow: 0 8px 32px rgba(0, 0, 0, 0.6); z-index: 100; backdrop-filter: blur(12px);"
  >
    <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px;">
      <h3 style="margin: 0; color: #fbbf24; font-size: 16px; font-weight: 600;">Create Link</h3>
      <button
        onclick={handleLinkFormClose}
        style="background: none; border: none; color: rgba(255,255,255,0.6); font-size: 20px; cursor: pointer; padding: 4px 8px; border-radius: 4px; transition: all 0.2s;"
        aria-label="Close"
      >
        ×
      </button>
    </div>
    <select
      bind:value={linkFormState.newLinkType}
      style="width: 100%; padding: 12px; margin-bottom: 16px; border: 1px solid rgba(255,255,255,0.2); border-radius: 8px; background: rgba(0,0,0,0.4); color: white; font-size: 14px; cursor: pointer; transition: border-color 0.2s;"
    >
      <option value="reference">📖 Reference</option>
      <option value="dependency">🔗 Dependency</option>
      <option value="related">🔀 Related</option>
      <option value="custom">✨ Custom</option>
    </select>
    <label style="display: block; color: rgba(255,255,255,0.8); font-size: 13px; margin-bottom: 8px; font-weight: 500;">Link Strength: {linkFormState.newLinkWeight.toFixed(1)}</label>
    <input
      type="range"
      min="0.1"
      max="1.0"
      step="0.1"
      bind:value={linkFormState.newLinkWeight}
      style="width: 100%; margin-bottom: 16px; accent-color: #fbbf24;"
    />
    <div style="display: flex; gap: 12px; justify-content: flex-end;">
      <button 
        onclick={handleLinkFormClose} 
        style="padding: 10px 20px; border: 1px solid rgba(255,255,255,0.3); border-radius: 8px; background: transparent; color: white; cursor: pointer; font-size: 14px; transition: all 0.2s;"
      >
        Cancel
      </button>
      <button 
        onclick={handleCreateLink} 
        style="padding: 10px 20px; border: none; border-radius: 8px; background: linear-gradient(135deg, #fbbf24, #f59e0b); color: #000; cursor: pointer; font-size: 14px; font-weight: 500; transition: all 0.2s; box-shadow: 0 4px 12px rgba(251, 191, 36, 0.3);"
      >
        Create Link
      </button>
    </div>
  </div>
{/if}

{#if hotkeysState.showSearchBox}
  <div
    class="search-box"
    style="position: absolute; top: 16px; left: 50%; transform: translateX(-50%); background: rgba(10, 26, 58, 0.95); border: 1px solid rgba(138, 43, 226, 0.5); border-radius: 8px; padding: 8px 12px; display: flex; align-items: center; gap: 8px; z-index: 100; box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);"
  >
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="rgba(255,255,255,0.7)" stroke-width="2">
      <circle cx="11" cy="11" r="8" />
      <line x1="21" y1="21" x2="16.65" y2="16.65" />
    </svg>
    <input
      bind:this={searchInput}
      type="text"
      placeholder="Search nodes..."
      bind:value={hotkeysState.searchQuery}
      oninput={handleUpdateSearch}
      style="background: transparent; border: none; color: white; outline: none; min-width: 200px; font-size: 14px;"
    />
    {#if hotkeysState.searchMatchIds.length > 0}
      <span style="color: rgba(255,255,255,0.6); font-size: 12px;">{hotkeysState.searchCurrentIndex + 1}/{hotkeysState.searchMatchIds.length}</span>
    {/if}
    <button onclick={handleCloseSearch} style="background: none; border: none; color: rgba(255,255,255,0.6); cursor: pointer; font-size: 16px;">×</button>
  </div>
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
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
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
    {#if undoToastStage === 'done'}
      <span style="font-size: 14px; color: rgba(255,255,255,0.9);">Note deleted.</span>
    {:else}
      <span style="font-size: 14px;">Note deleted.</span>
      <div style="display: flex; gap: 8px; align-items: center;">
        <button
          onclick={restoreDeletedNode}
          style="padding: 6px 12px; border: none; border-radius: 4px; background: #8b5cf6; color: white; cursor: pointer; font-size: 13px; font-weight: 600;"
        >
          Restore
        </button>
        <button
          onclick={cancelUndo}
          style="background: none; border: none; color: rgba(255,255,255,0.6); cursor: pointer; font-size: 18px; line-height: 1;"
          aria-label="Close"
        >
          ×
        </button>
      </div>
    {/if}
  </div>
{/if}

{#if hotkeysState.showHelpTooltip && hotkeysState.helpTooltipPosition.x === -1}
  <!-- Inactivity tip — shown at bottom-center -->
  <div
    class="help-tooltip"
    style="position: fixed; bottom: 32px; left: 50%; transform: translateX(-50%); background: rgba(10, 26, 58, 0.95); border: 1px solid rgba(138, 43, 226, 0.5); border-radius: 10px; padding: 12px 18px; color: rgba(255,255,255,0.92); font-size: 13px; max-width: 360px; z-index: 1000; pointer-events: none; box-shadow: 0 4px 24px rgba(0,0,0,0.6); text-align: center; line-height: 1.5;"
  >
    <div style="display: flex; align-items: center; justify-content: center; gap: 6px; margin-bottom: 4px; color: #a78bfa; font-weight: 600; font-size: 11px; letter-spacing: 0.05em; text-transform: uppercase;">
      <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="12" cy="12" r="10" />
        <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
        <line x1="12" y1="17" x2="12.01" y2="17" />
      </svg>
      Tip
    </div>
    {hotkeysState.helpTooltipMessage}
  </div>
{:else if hotkeysState.showHelpTooltip}
  <!-- Node hover tooltip — shown near cursor -->
  <div
    class="help-tooltip"
    style="position: fixed; left: {hotkeysState.helpTooltipPosition.x}px; top: {hotkeysState.helpTooltipPosition.y}px; transform: translate(-50%, -100%); background: rgba(10, 26, 58, 0.95); border: 1px solid rgba(138, 43, 226, 0.5); border-radius: 8px; padding: 10px 14px; color: white; font-size: 13px; max-width: 320px; z-index: 1000; pointer-events: none; box-shadow: 0 4px 20px rgba(0,0,0,0.5);"
  >
    {hotkeysState.helpTooltipMessage}
  </div>
{/if}

{#if hotkeysState.showHelpModal}
  <HelpHotkeysModal
    hotkeyLines={hotkeyLines}
    helpContent={helpContent}
    onClose={closeHelpModal}
  />
{/if}

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
</style>
