<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
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
    onNodeClick
  }: { 
    nodes: Array<{ id: string; title: string; type?: string }>;
    links: Array<{ source: string; target: string; weight?: number; link_type?: string }>;
    onNodeClick?: (node: { id: string; title: string; type?: string }) => void;
  } = $props();

  // Debug: проверяем типы узлов при изменении
  $effect(() => {
    if (nodes.length > 0) {
      const types = nodes.map(n => n.type || 'undefined');
      const uniqueTypes = [...new Set(types)];
      console.log('[GraphCanvas] Received nodes types:', uniqueTypes, 'Total:', nodes.length);
      console.log('[GraphCanvas] First node:', nodes[0]);
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

  const transform: TransformState = $state({ x: 0, y: 0, k: 1 });
  const dragState: DragState = $state({ dragging: false, dragStart: { x: 0, y: 0 } });
  const simState: SimulationState = $state({
    simulation: null,
    simLinks: [],
    isRunning: false
  });

  // Для отслеживания изменений данных по содержимому (не по ссылке)
  let lastDataKey = '';

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
          draw(ctx, width, height, simState.simLinks, simNodes, angles, transform);
        }
      }
    );
    
    return () => {
      observerCleanup?.disconnect();
      resizeCleanup?.clear();
      animationLoop?.stop();
      clearSimulation(simState);
      clearAnimationState(angles, speeds);
    };
  });

  // Реактивно перезапускаем симуляцию при изменении данных
  $effect(() => {
    const nodesCount = nodes.length;
    const linksCount = links.length;
    const dataKey = `${nodesCount}-${linksCount}`;

    if (dataKey === lastDataKey && simState.isRunning) {
      return;
    }
    lastDataKey = dataKey;

    if (!browser) return;

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
          draw(ctx, width, height, simState.simLinks, simNodes, angles, transform);
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

  // Обёртки для обработчиков событий
  function onZoom(e: WheelEvent) {
    handleZoom(e, transform, () => {
      const simNodes = getSimulationNodes(simState);
      if (ctx && simNodes.length > 0) {
        draw(ctx, width, height, simState.simLinks, simNodes, angles, transform);
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
        draw(ctx, width, height, simState.simLinks, simNodes, angles, transform);
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