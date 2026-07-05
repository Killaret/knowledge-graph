<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import type { GraphDelta } from '$lib/api/graph';
  import LinkTooltip from '$lib/components/LinkTooltip.svelte';
  import GraphTooltip from '$lib/components/GraphTooltip.svelte';
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
    handleZoom,
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
    createGhostNode,
    updateGhostNodePosition,
    updateGhostNodePulse,
    isPointOverGhostNode,
    type GhostNodeState,
    createGravitySystem,
    type GravitySystem
  } from './GraphCanvas';

  const {
    nodes,
    links,
    onNodeClick,
    onLinkEdit,
    onLinkDelete,
    onNoteCreate,
    onLinkCreate,
    onNoteDelete,
    delta,
    disableVariation = false
  }: {
    nodes: Array<{ id: string; title: string; type?: string }>;
    links: Array<{ source: string; target: string; weight?: number; link_type?: string; source_type?: string }>;
    onNodeClick?: (node: { id: string; title: string; type?: string }) => void;
    onLinkEdit?: (link: { source: string; target: string; link_type: string; weight: number }) => void;
    onLinkDelete?: (link: { source: string; target: string; link_type: string }) => void;
    onNoteCreate?: (data: { title: string; content: string; type: string }) => void;
    onLinkCreate?: (link: { source: string; target: string; link_type: string; weight: number }) => void;
    onNoteDelete?: (nodeId: string) => void;
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

  let canvas: HTMLCanvasElement;
  let ctx: CanvasRenderingContext2D;
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

  // Double-tap zoom state
  let lastTouchTime = 0;
  let lastTouchPos = { x: 0, y: 0 };
  let tapCount = 0;

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
  let ghostNode: GhostNodeState = $state(createGhostNode(width, height, nodes));
  let gravitySystem: GravitySystem = $state(createGravitySystem());

  // Drag-and-drop state
  let draggedNodeId: string | null = $state(null);
  let dragStartPosition = $state({ x: 0, y: 0 });
  let isDraggingForLink = $state(false);
  let linkSourceNodeId: string | null = $state(null);
  let linkTargetNodeId: string | null = $state(null);
  let mouseWorldPosition = $state({ x: 0, y: 0 });

  // Note creation form state
  let showNoteForm = $state(false);
  let noteFormPosition = $state({ x: 0, y: 0 });
  let newNoteTitle = $state('');
  let newNoteContent = $state('');
  let newNoteType = $state('planet');

  // Link creation form state
  let showLinkForm = $state(false);
  let linkFormPosition = $state({ x: 0, y: 0 });
  let newLinkType = $state('related');
  let newLinkWeight = $state(0.5);

  onMount(() => {
    if (!browser) return;
    
    // SSR-safe: получаем контекст canvas
    ctx = canvas.getContext('2d')!;
    
    // Начальный resize
    resizeCanvas(canvas, resizeState);
    width = resizeState.width;
    height = resizeState.height;
    
    // Initialize interactive systems
    particleSystem = new ParticleSystem(nodes.length);
    blackHole = createBlackHole(width, height);
    ghostNode = createGhostNode(width, height, nodes);
    gravitySystem = createGravitySystem();
    
    // ResizeObserver для отслеживания размера контейнера
    observerCleanup = setupResizeObserver(canvas, () => {
      resizeCanvas(canvas, resizeState);
      width = resizeState.width;
      height = resizeState.height;
    });
    
    // Отложенный resize для стабильных размеров
    resizeCleanup = scheduleDelayedResize(() => {
      resizeCanvas(canvas, resizeState);
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
          draw(ctx, width, height, simState.simLinks, simNodes, angles, transform, simState.nodeOpacity, simState.linkOpacity, stableRender, animationTime, hoveredNodeId, particleSystem, blackHole, ghostNode, gravitySystem);
        }
      },
      stableRender
    );
    
    mounted = true; // triggers $effect re-run since it's $state

    return () => {
      mounted = false; // $state
      observerCleanup?.disconnect();
      resizeCleanup?.clear();
      animationLoop?.stop();
      clearSimulation(simState);
      particleSystem?.clear();
      clearAnimationState(angles, speeds);
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

    // Запускаем новую симуляцию
    startSimulation(
      nodes,
      links,
      width,
      height,
      simState,
      transform,
      () => {
        const simNodes = getSimulationNodes(simState);
        if (ctx && simNodes.length > 0) {
          draw(ctx, width, height, simState.simLinks, simNodes, angles, transform, simState.nodeOpacity, simState.linkOpacity, stableRender, animationTime, hoveredNodeId, particleSystem, blackHole, ghostNode, gravitySystem);
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
          draw(ctx, width, height, simState.simLinks, simNodes, angles, transform, simState.nodeOpacity, simState.linkOpacity, stableRender, animationTime, hoveredNodeId, particleSystem, blackHole, ghostNode, gravitySystem);
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



  function getMouseWorldPosition(e: MouseEvent): { x: number; y: number } {
    const rect = canvas.getBoundingClientRect();
    return {
      x: (e.clientX - rect.left - transform.x) / transform.k,
      y: (e.clientY - rect.top - transform.y) / transform.k
    };
  }

  function findNodeAtPosition(x: number, y: number): SimulationNode | undefined {
    const simNodes = getSimulationNodes(simState);
    for (const node of simNodes) {
      if (node.x == null || node.y == null) continue;
      const dx = node.x - x;
      const dy = node.y - y;
      if (Math.sqrt(dx * dx + dy * dy) < 30) {
        return node;
      }
    }
    return undefined;
  }

  function redraw() {
    const simNodes = getSimulationNodes(simState);
    if (ctx) {
      draw(ctx, width, height, simState.simLinks, simNodes, angles, transform, simState.nodeOpacity, simState.linkOpacity, stableRender, animationTime, hoveredNodeId, particleSystem, blackHole, ghostNode, gravitySystem);
    }
  }

  function onZoom(e: WheelEvent) {
    handleZoom(e, transform, canvas, redraw);
  }

  function onMouseDown(e: MouseEvent) {
    const pos = getMouseWorldPosition(e);
    mouseWorldPosition = pos;

    // Ghost node click -> create note form
    if (isPointOverGhostNode(e.clientX, e.clientY, ghostNode, transform)) {
      showNoteForm = true;
      noteFormPosition = { x: e.clientX, y: e.clientY };
      e.preventDefault();
      return;
    }

    // Node click -> start dragging for link/move/delete
    const node = findNodeAtPosition(pos.x, pos.y);
    if (node) {
      draggedNodeId = node.id;
      dragStartPosition = { x: node.x!, y: node.y! };
      dragState.dragging = true;
      canvas.style.cursor = 'grabbing';
      e.preventDefault();
      return;
    }

    // Empty space -> pan
    dragState.dragStart = { x: e.clientX - transform.x, y: e.clientY - transform.y };
    dragState.dragging = true;
    canvas.style.cursor = 'grabbing';
  }

  function onMouseMove(e: MouseEvent) {
    const pos = getMouseWorldPosition(e);
    mouseWorldPosition = pos;

    // Update hover states for interactive elements
    blackHole.hovered = isPointOverBlackHole(e.clientX, e.clientY, blackHole, transform);
    ghostNode.hovered = isPointOverGhostNode(e.clientX, e.clientY, ghostNode, transform);

    // Dragging a node
    if (draggedNodeId && dragState.dragging) {
      const node = getSimulationNodes(simState).find((n) => n.id === draggedNodeId);
      if (node && node.x != null && node.y != null) {
        node.x = pos.x;
        node.y = pos.y;
        node.fx = pos.x;
        node.fy = pos.y;

        // Check if over black hole
        blackHole.hovered = isNodeOverBlackHole(node, blackHole);
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
    const simNodes = getSimulationNodes(simState);
    for (const node of simNodes) {
      if (node.x && node.y) {
        const dx = pos.x - node.x;
        const dy = pos.y - node.y;
        if (Math.sqrt(dx * dx + dy * dy) < 30) {
          hoveredNodeId = node.id;
          foundHoveredNode = true;
          break;
        }
      }
    }
    if (!foundHoveredNode) {
      hoveredNodeId = null;
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
    const pos = getMouseWorldPosition(e);
    const wasDraggingNode = draggedNodeId !== null;

    if (draggedNodeId) {
      const node = getSimulationNodes(simState).find((n) => n.id === draggedNodeId);
      if (node) {
        // Check if dropped over black hole -> delete
        if (isNodeOverBlackHole(node, blackHole)) {
          animateNodeDeletion(node, () => {
            if (onNoteDelete) {
              onNoteDelete(node.id);
            }
          });
        } else {
          // Check if dropped over another node -> create link
          const targetNode = findNodeAtPosition(pos.x, pos.y);
          if (targetNode && targetNode.id !== draggedNodeId) {
            linkSourceNodeId = draggedNodeId;
            linkTargetNodeId = targetNode.id;
            showLinkForm = true;
            linkFormPosition = { x: e.clientX, y: e.clientY };

            // Spring-back source node to original position
            animateSpringBack(node, dragStartPosition);
          } else {
            // Release fixed position
            node.fx = undefined;
            node.fy = undefined;
          }
        }
      }
      draggedNodeId = null;
    }

    dragState.dragging = false;
    canvas.style.cursor = 'grab';
    blackHole.hovered = false;

    // Ghost node click up (without drag) - already handled in mousedown
    if (!wasDraggingNode && !ghostNode.hovered) {
      const clickedNode = findNodeAtPosition(pos.x, pos.y);
      if (!clickedNode) {
        hoveredLink = null;
      }
    }
  }

  function onClick(e: MouseEvent) {
    const pos = getMouseWorldPosition(e);
    const clickedNode = findNodeAtPosition(pos.x, pos.y);
    if (clickedNode && onNodeClick) {
      onNodeClick({ id: clickedNode.id, title: clickedNode.title, type: clickedNode.type });
    }
  }

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

  function createNote() {
    if (newNoteTitle.trim() && onNoteCreate) {
      onNoteCreate({
        title: newNoteTitle.trim(),
        content: newNoteContent.trim(),
        type: newNoteType
      });
    }
    closeNoteForm();
  }

  function closeNoteForm() {
    showNoteForm = false;
    newNoteTitle = '';
    newNoteContent = '';
    newNoteType = 'planet';
  }

  function createLink() {
    if (linkSourceNodeId && linkTargetNodeId && onLinkCreate) {
      onLinkCreate({
        source: linkSourceNodeId,
        target: linkTargetNodeId,
        link_type: newLinkType,
        weight: newLinkWeight
      });
    }
    closeLinkForm();
  }

  function closeLinkForm() {
    showLinkForm = false;
    linkSourceNodeId = null;
    linkTargetNodeId = null;
    newLinkType = 'related';
    newLinkWeight = 0.5;
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
  function handleTouchStart(e: TouchEvent) {
    if (!browser) return; // Skip in server-side rendering

    if (e.touches.length === 1) {
      const now = Date.now();
      const touch = e.touches[0];
      const dx = Math.abs(touch.clientX - lastTouchPos.x);
      const dy = Math.abs(touch.clientY - lastTouchPos.y);

      if (now - lastTouchTime < 300 && dx < 30 && dy < 30) {
        tapCount++;
        handleDoubleTap(touch.clientX, touch.clientY);
        e.preventDefault();
      } else {
        tapCount = 0;
      }

      lastTouchTime = now;
      lastTouchPos = { x: touch.clientX, y: touch.clientY };
    }
  }

  function handleDoubleTap(clientX: number, clientY: number) {
    const rect = canvas.getBoundingClientRect();
    const x = (clientX - rect.left - transform.x) / transform.k;
    const y = (clientY - rect.top - transform.y) / transform.k;

    if (tapCount === 1) {
      const newScale = transform.k * 2;
      const centerX = x * newScale;
      const centerY = y * newScale;

      transform.x = clientX - rect.left - centerX;
      transform.y = clientY - rect.top - centerY;
      transform.k = newScale;
    } else if (tapCount === 2) {
      const simNodes = getSimulationNodes(simState);
      if (ctx && simNodes.length > 0) {
        resetView(ctx, width, height, simNodes, transform);
        tapCount = 0;
      }
    }
  }</script>

<canvas
  bind:this={canvas}
  data-testid="graph-canvas"
  onmousedown={onMouseDown}
  onmousemove={onMouseMove}
  onmouseup={onMouseUp}
  onclick={onClick}
  ontouchstart={handleTouchStart}
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

{#if showNoteForm}
  <div
    class="note-form"
    style="position: absolute; left: {noteFormPosition.x}px; top: {noteFormPosition.y}px; background: rgba(10, 26, 58, 0.95); border: 1px solid rgba(138, 43, 226, 0.5); border-radius: 8px; padding: 16px; min-width: 280px; box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5); z-index: 100;"
  >
    <h3 style="margin: 0 0 12px 0; color: white; font-size: 14px;">Create New Note</h3>
    <input
      type="text"
      placeholder="Title"
      bind:value={newNoteTitle}
      style="width: 100%; padding: 8px; margin-bottom: 8px; border: 1px solid rgba(255,255,255,0.2); border-radius: 4px; background: rgba(0,0,0,0.3); color: white; box-sizing: border-box;"
    />
    <textarea
      placeholder="Content (optional)"
      bind:value={newNoteContent}
      style="width: 100%; padding: 8px; margin-bottom: 8px; border: 1px solid rgba(255,255,255,0.2); border-radius: 4px; background: rgba(0,0,0,0.3); color: white; min-height: 80px; box-sizing: border-box;"
    ></textarea>
    <select
      bind:value={newNoteType}
      style="width: 100%; padding: 8px; margin-bottom: 12px; border: 1px solid rgba(255,255,255,0.2); border-radius: 4px; background: rgba(0,0,0,0.3); color: white;"
    >
      <option value="star">Star</option>
      <option value="planet">Planet</option>
      <option value="comet">Comet</option>
      <option value="galaxy">Galaxy</option>
      <option value="asteroid">Asteroid</option>
    </select>
    <div style="display: flex; gap: 8px; justify-content: flex-end;">
      <button onclick={closeNoteForm} style="padding: 6px 12px; border: 1px solid rgba(255,255,255,0.3); border-radius: 4px; background: transparent; color: white; cursor: pointer;">Cancel</button>
      <button onclick={createNote} style="padding: 6px 12px; border: none; border-radius: 4px; background: #8b5cf6; color: white; cursor: pointer;">Create</button>
    </div>
  </div>
{/if}

{#if showLinkForm}
  <div
    class="link-form"
    style="position: absolute; left: {linkFormPosition.x}px; top: {linkFormPosition.y}px; background: rgba(10, 26, 58, 0.95); border: 1px solid rgba(255, 204, 0, 0.5); border-radius: 8px; padding: 16px; min-width: 260px; box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5); z-index: 100;"
  >
    <h3 style="margin: 0 0 12px 0; color: white; font-size: 14px;">Create Link</h3>
    <select
      bind:value={newLinkType}
      style="width: 100%; padding: 8px; margin-bottom: 12px; border: 1px solid rgba(255,255,255,0.2); border-radius: 4px; background: rgba(0,0,0,0.3); color: white;"
    >
      <option value="reference">Reference</option>
      <option value="dependency">Dependency</option>
      <option value="related">Related</option>
      <option value="custom">Custom</option>
    </select>
    <label style="display: block; color: rgba(255,255,255,0.7); font-size: 12px; margin-bottom: 4px;">Weight: {newLinkWeight.toFixed(1)}</label>
    <input
      type="range"
      min="0.1"
      max="1.0"
      step="0.1"
      bind:value={newLinkWeight}
      style="width: 100%; margin-bottom: 12px;"
    />
    <div style="display: flex; gap: 8px; justify-content: flex-end;">
      <button onclick={closeLinkForm} style="padding: 6px 12px; border: 1px solid rgba(255,255,255,0.3); border-radius: 4px; background: transparent; color: white; cursor: pointer;">Cancel</button>
      <button onclick={createLink} style="padding: 6px 12px; border: none; border-radius: 4px; background: #ffcc00; color: #000; cursor: pointer;">Create</button>
    </div>
  </div>
{/if}
