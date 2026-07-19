/**
 * Delta update management for GraphCanvas
 * Handles incremental graph updates with animations
 */

import type { GraphDeltaData } from "$shared/api/graph";
import { GraphDelta } from "$shared/lib/domain";
import type {
  SimulationNode,
  SimulationLink,
  SimulationState,
  TransformState,
} from "./types";
import * as d3Force from "d3-force";

// Easing function for smooth fade animation
function easeOutCubic(t: number): number {
  return 1 - Math.pow(1 - t, 3);
}

function computeStableProgress(
  currentNodes: any[],
  totalNodes: number,
): number {
  if (totalNodes === 0) return 1;

  const stableNodes = currentNodes.filter(
    (n: any) =>
      n.x !== undefined &&
      !isNaN(n.x) &&
      n.y !== undefined &&
      !isNaN(n.y) &&
      Math.hypot(n.vx ?? 0, n.vy ?? 0) < 0.2,
  ).length;

  return Math.min(stableNodes / totalNodes, 1);
}

function initializeOpacityMaps(
  nodes: SimulationNode[],
  links: SimulationLink[],
  state: SimulationState,
): void {
  state.nodeOpacity = new Map();
  state.linkOpacity = new Map();

  nodes.forEach((node) => {
    state.nodeOpacity.set(node.id, 0);
  });

  links.forEach((link, index) => {
    const linkId = `${link.source}-${link.target}-${index}`;
    state.linkOpacity.set(linkId, 0);
  });
}

function interpolateOpacity(
  opacityMap: Map<string, number>,
  targetOpacity: number,
  factor: number = 0.1,
): void {
  opacityMap.forEach((currentOpacity, key) => {
    const newOpacity =
      currentOpacity + (targetOpacity - currentOpacity) * factor;
    opacityMap.set(key, Math.min(Math.max(newOpacity, 0), 1));
  });
}

function anyOpacityBelowOne(state: SimulationState): boolean {
  for (const value of state.nodeOpacity.values()) {
    if (value < 0.999) return true;
  }
  for (const value of state.linkOpacity.values()) {
    if (value < 0.999) return true;
  }
  return false;
}

function startFadeAnimation(state: SimulationState, totalNodes: number): void {
  if (state.fadeAnimationId !== null) {
    cancelAnimationFrame(state.fadeAnimationId);
    state.fadeAnimationId = null;
  }

  const animateFade = () => {
    if (!state.simulation) {
      state.fadeAnimationId = null;
      return;
    }

    const currentNodes = state.simulation.nodes();
    const progress = computeStableProgress(currentNodes, totalNodes);
    const targetOpacity = easeOutCubic(progress);

    interpolateOpacity(state.nodeOpacity, targetOpacity, 0.12);
    interpolateOpacity(state.linkOpacity, targetOpacity, 0.12);

    if (progress < 1 || anyOpacityBelowOne(state)) {
      state.fadeAnimationId = requestAnimationFrame(animateFade);
    } else {
      state.fadeAnimationId = null;
    }
  };

  state.fadeAnimationId = requestAnimationFrame(animateFade);
}

export interface DeltaUpdateOptions {
  nodes: SimulationNode[];
  links: SimulationLink[];
  width: number;
  height: number;
  state: SimulationState;
  transform: TransformState;
  onTick: () => void;
  onResetView: () => void;
}

/**
 * Apply delta updates to the graph simulation
 * Returns true if simulation was restarted, false otherwise
 */
export function applyDelta(
  delta: GraphDeltaData,
  options: DeltaUpdateOptions,
): boolean {
  const domainDelta = GraphDelta.fromAPI(delta);

  if (import.meta.env.DEV) {
    console.log(
      "[Delta] Applying delta with",
      domainDelta.totalChanges,
      "changes",
    );
  }

  // Если изменений много (>10), перезапускаем симуляцию полностью
  if (domainDelta.requiresFullRestart()) {
    return applyFullRestart(domainDelta, options);
  }

  // Для небольших изменений применяем инкрементальные обновления
  return applyIncremental(domainDelta, options);
}

/**
 * Full simulation restart for large deltas
 */
function applyFullRestart(
  delta: GraphDelta,
  options: DeltaUpdateOptions,
): boolean {
  const { nodes, links, width, height, state, onTick, onResetView } = options;

  if (import.meta.env.DEV) {
    console.log("[Delta] Full simulation restart");
  }

  // Фильтруем удаленные узлы
  const filteredNodes = nodes.filter(
    (n) => !delta.removedNodeIds.includes(n.id),
  );

  // Добавляем новые узлы
  if (delta.addedNodes.length > 0) {
    filteredNodes.push(...delta.addedNodes);
  }

  // Обновляем существующие узлы
  if (delta.updatedNodes.length > 0) {
    delta.updatedNodes.forEach((updated) => {
      const index = filteredNodes.findIndex((n) => n.id === updated.id);
      if (index !== -1) {
        filteredNodes[index] = updated;
      }
    });
  }

  // Фильтруем и обновляем связи
  const filteredLinks = links.filter((l) => {
    if (delta.removedLinks.length > 0) {
      const isRemoved = delta.removedLinks.some(
        (removed) => removed.source === l.source && removed.target === l.target,
      );
      if (isRemoved) return false;
    }
    return true;
  });

  // Добавляем новые связи
  if (delta.addedLinks.length > 0) {
    filteredLinks.push(...delta.addedLinks);
  }

  // Инициализируем прозрачность для новых узлов
  if (delta.addedNodes.length > 0) {
    delta.addedNodes.forEach((node) => {
      state.nodeOpacity.set(node.id, 0);
    });
  }

  // Перезапускаем симуляцию
  if (state.simulation) {
    state.simulation.stop();
  }
  if (state.fadeAnimationId !== null) {
    cancelAnimationFrame(state.fadeAnimationId);
    state.fadeAnimationId = null;
  }

  // Распределяем новые узлы в круге
  const simulationNodes = filteredNodes.map((n, i) => {
    const angle = (i / filteredNodes.length) * 2 * Math.PI;
    const radius = Math.min(width, height) * 0.3;
    return {
      ...n,
      x: width / 2 + Math.cos(angle) * radius,
      y: height / 2 + Math.sin(angle) * radius,
    };
  });

  state.simLinks = filteredLinks.map((l) => ({
    source: l.source,
    target: l.target,
    weight: l.weight ?? 1,
    link_type: l.link_type,
  }));

  // Initialize opacity maps for fade effect using current nodes and links
  initializeOpacityMaps(filteredNodes, state.simLinks, state);

  const totalNodes = filteredNodes.length;
  let tickCount = 0;

  state.simulation = d3Force
    .forceSimulation(simulationNodes as any)
    .force(
      "link",
      d3Force
        .forceLink(state.simLinks)
        .id((d: any) => d.id)
        .distance(100)
        .strength(0.3),
    )
    .force("charge", d3Force.forceManyBody().strength(-150))
    .force("center", d3Force.forceCenter(width / 2, height / 2).strength(0.5))
    .force("collision", d3Force.forceCollide().radius(30))
    .alphaDecay(0.01)
    .on("tick", () => {
      onTick();
      tickCount++;

      // Update opacity for fade effect based on node stabilization
      if (tickCount % 5 === 0 && state.simulation) {
        const currentNodes = state.simulation.nodes();
        const progress = computeStableProgress(currentNodes, totalNodes);
        const targetOpacity = easeOutCubic(progress);

        interpolateOpacity(state.nodeOpacity, targetOpacity, 0.12);
        interpolateOpacity(state.linkOpacity, targetOpacity, 0.12);
      }
    })
    .on("end", () => {
      // Final fade animation
      if (state.fadeAnimationId !== null) {
        cancelAnimationFrame(state.fadeAnimationId);
      }

      const startTime = performance.now();
      const duration = 2400;

      const animateFinalFade = (currentTime: number) => {
        const elapsed = currentTime - startTime;
        const progress = Math.min(elapsed / duration, 1);
        const targetOpacity = 1 - Math.pow(1 - progress, 3);

        state.nodeOpacity.forEach((_, nodeId) => {
          const currentOpacity = state.nodeOpacity.get(nodeId) || 0;
          const newOpacity =
            currentOpacity + (targetOpacity - currentOpacity) * 0.15;
          state.nodeOpacity.set(nodeId, Math.min(Math.max(newOpacity, 0), 1));
        });

        state.linkOpacity.forEach((_, linkId) => {
          const currentOpacity = state.linkOpacity.get(linkId) || 0;
          const newOpacity =
            currentOpacity + (targetOpacity - currentOpacity) * 0.15;
          state.linkOpacity.set(linkId, Math.min(Math.max(newOpacity, 0), 1));
        });

        if (progress < 1) {
          state.fadeAnimationId = requestAnimationFrame(animateFinalFade);
        } else {
          state.fadeAnimationId = null;
        }
      };

      state.fadeAnimationId = requestAnimationFrame(animateFinalFade);
    });

  // Warmup
  for (let i = 0; i < 50; i++) {
    state.simulation.tick();
  }

  onResetView();
  state.simulation.alpha(1).restart();
  startFadeAnimation(state, totalNodes);
  state.isRunning = true;

  return true;
}

/**
 * Incremental delta application for small changes
 */
function applyIncremental(
  delta: GraphDelta,
  options: DeltaUpdateOptions,
): boolean {
  const { state } = options;

  if (import.meta.env.DEV) {
    console.log("[Delta] Incremental update");
  }

  let simulationRestarted = false;

  // Обновляем существующие узлы (без перезапуска симуляции)
  if (delta.updatedNodes.length > 0) {
    const simNodes = state.simulation?.nodes() || [];
    delta.updatedNodes.forEach((updated) => {
      const simNode = simNodes.find((n: any) => n.id === updated.id);
      if (simNode) {
        simNode.title = updated.title;
        simNode.type = updated.type;
        if (updated.x !== undefined) simNode.x = updated.x;
        if (updated.y !== undefined) simNode.y = updated.y;
      }
    });

    // Легкий перезапуск симуляции для применения изменений
    if (state.simulation) {
      state.simulation.alpha(0.3).restart();
      simulationRestarted = true;
    }
  }

  // Для добавления/удаления узлов и связей используем полный перезапуск
  // даже для небольших изменений, так как D3 требует этого
  if (
    delta.addedNodes.length > 0 ||
    delta.removedNodeIds.length > 0 ||
    delta.addedLinks.length > 0 ||
    delta.removedLinks.length > 0
  ) {
    return applyFullRestart(delta, options);
  }

  return simulationRestarted;
}
