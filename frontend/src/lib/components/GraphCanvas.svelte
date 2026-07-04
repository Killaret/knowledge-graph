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
    handlePanStart,
    handlePanMove,
    handlePanEnd,
    handleClick,
    findLinkAtPosition,
    startAnimationLoop,
    clearAnimationState,
    type TransformState,
    type DragState,
    applyDelta as applyDeltaToSimulation
  } from './GraphCanvas';
  import { drawBackground, drawAnimatedLink } from './GraphCanvas/renderer';

  const {
    nodes,
    links,
    onNodeClick,
    onLinkEdit,
    onLinkDelete,
    delta,
    disableVariation = false
  }: {
    nodes: Array<{ id: string; title: string; type?: string }>;
    links: Array<{ source: string; target: string; weight?: number; link_type?: string; source_type?: string }>;
    onNodeClick?: (node: { id: string; title: string; type?: string }) => void;
    onLinkEdit?: (link: { source: string; target: string; link_type: string; weight: number }) => void;
    onLinkDelete?: (link: { source: string; target: string; link_type: string }) => void;
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

  onMount(() => {
    if (!browser) return;
    
    // SSR-safe: получаем контекст canvas
    ctx = canvas.getContext('2d')!;
    
    // Начальный resize
    resizeCanvas(canvas, resizeState);
    width = resizeState.width;
    height = resizeState.height;
    
    // Initialize particle system
    particleSystem = new ParticleSystem(nodes.length);
    
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
        if (ctx && simNodes.length > 0) {
          // Update animation time
          animationTime = performance.now();
          
          // Draw background
          drawBackground(ctx, width, height, simNodes, animationTime);
          
          // Draw with new effects
          draw(ctx, width, height, simState.simLinks, simNodes, angles, transform, simState.nodeOpacity, simState.linkOpacity, stableRender, animationTime, hoveredNodeId, particleSystem);
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
          draw(ctx, width, height, simState.simLinks, simNodes, angles, transform, simState.nodeOpacity, simState.linkOpacity, stableRender);
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
          draw(ctx, width, height, simState.simLinks, simNodes, angles, transform, simState.nodeOpacity, simState.linkOpacity, stableRender);
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

  // Обёртки для обработчиков событий
  function onZoom(e: WheelEvent) {
    handleZoom(e, transform, canvas, () => {
      const simNodes = getSimulationNodes(simState);
      if (ctx && simNodes.length > 0) {
        draw(ctx, width, height, simState.simLinks, simNodes, angles, transform, simState.nodeOpacity, simState.linkOpacity, stableRender);
      }
    });
  }

  function onPanStart(e: MouseEvent) {
    handlePanStart(e, dragState, transform, canvas);
  }

  function onPanMove(e: MouseEvent) {
    handlePanMove(e, dragState, transform, () => {
      const simNodes = getSimulationNodes(simState);
      if (ctx && simNodes.length > 0) {
        draw(ctx, width, height, simState.simLinks, simNodes, angles, transform, simState.nodeOpacity, simState.linkOpacity, stableRender);
      }
    });

    // Link hover detection (only when not dragging)
    if (!dragState.dragging) {
      const rect = canvas.getBoundingClientRect();
      const mouseX = (e.clientX - rect.left - transform.x) / transform.k;
      const mouseY = (e.clientY - rect.top - transform.y) / transform.k;

      const simNodes = getSimulationNodes(simState);
      const hovered = findLinkAtPosition(mouseX, mouseY, simState.simLinks, simNodes, transform);

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

          // Position tooltip near the center of the link
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
  }

  function onPanEnd() {
    handlePanEnd(dragState, canvas);
    hoveredLink = null;
  }

  function onClick(e: MouseEvent) {
    handleClick(e, canvas, transform, getSimulationNodes(simState), onNodeClick);
    hoveredLink = null;
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
  onmousedown={onPanStart}
  onmousemove={onPanMove}
  onmouseup={onPanEnd}
  onclick={onClick}
  ontouchstart={handleTouchStart}
  onwheel={onZoom}
  style="width: 100%; height: 100%; cursor: grab; background: linear-gradient(145deg, #0a1a3a, #020617);"
></canvas>

{#if hoveredLink}
  {@const sourceNode = nodes.find(n => n.id === hoveredLink.source)}
  {@const targetNode = nodes.find(n => n.id === hoveredLink.target)}
  <LinkTooltip
    visible={true}
    x={tooltipPosition.x}
    y={tooltipPosition.y}
    linkType={hoveredLink.link_type}
    weight={hoveredLink.weight}
    sourceType={hoveredLink.source_type}
    sourceTitle={sourceNode?.title || 'Unknown'}
    targetTitle={targetNode?.title || 'Unknown'}
    onEdit={onLinkEdit ? handleLinkEdit : undefined}
    onDelete={onLinkDelete ? handleLinkDelete : undefined}
  />
{/if}
