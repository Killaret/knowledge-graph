<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import type { GraphDelta } from '$lib/api/graph';
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
    startAnimationLoop,
    clearAnimationState,
    type TransformState,
    type DragState
  } from './GraphCanvas';

  const {
    nodes,
    links,
    onNodeClick,
    delta
  }: {
    nodes: Array<{ id: string; title: string; type?: string }>;
    links: Array<{ source: string; target: string; weight?: number; link_type?: string }>;
    onNodeClick?: (node: { id: string; title: string; type?: string }) => void;
    delta?: GraphDelta;
  } = $props();

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

  onMount(() => {
    if (!browser) return;
    
    // SSR-safe: получаем контекст canvas
    ctx = canvas.getContext('2d')!;
    
    // Начальный resize
    resizeCanvas(canvas, resizeState);
    width = resizeState.width;
    height = resizeState.height;
    
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
          draw(ctx, width, height, simState.simLinks, simNodes, angles, transform, simState.nodeOpacity, simState.linkOpacity);
        }
      }
    );
    
    mounted = true; // triggers $effect re-run since it's $state

    return () => {
      mounted = false; // $state
      observerCleanup?.disconnect();
      resizeCleanup?.clear();
      animationLoop?.stop();
      clearSimulation(simState);
      clearAnimationState(angles, speeds);
    };
  });

  // Реактивно перезапускаем симуляцию при изменении данных
  $effect(() => {
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
          draw(ctx, width, height, simState.simLinks, simNodes, angles, transform, simState.nodeOpacity, simState.linkOpacity);
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
          draw(ctx, width, height, simState.simLinks, simNodes, angles, transform, simState.nodeOpacity, simState.linkOpacity);
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
        draw(ctx, width, height, simState.simLinks, simNodes, angles, transform, simState.nodeOpacity, simState.linkOpacity);
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
        draw(ctx, width, height, simState.simLinks, simNodes, angles, transform, simState.nodeOpacity, simState.linkOpacity);
      }
    });
  }

  function onPanEnd() {
    handlePanEnd(dragState, canvas);
  }

  function onClick(e: MouseEvent) {
    handleClick(e, canvas, transform, getSimulationNodes(simState), onNodeClick);
  }
</script>

<canvas
  bind:this={canvas}
  onmousedown={onPanStart}
  onmousemove={onPanMove}
  onmouseup={onPanEnd}
  onclick={onClick}
  onwheel={onZoom}
  style="width: 100%; height: 100%; cursor: grab; background: linear-gradient(145deg, #0a1a3a, #020617);"
></canvas>
