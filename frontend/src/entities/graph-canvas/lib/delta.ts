/**
 * Delta update management for GraphCanvas
 * Handles incremental graph updates with animations
 */

import type { GraphDeltaData, GraphLink } from "$shared/api/graph";
import { GraphDelta } from "$entities";
import {
  getLinkEndpointId,
  type SimulationNode,
  type SimulationLink,
  type SimulationState,
  type TransformState,
} from "./types";
import { getLinkId } from "./simulation";
import { addNodesToSimulation } from "./incremental";
import * as d3Force from "d3-force";

// Easing function for smooth fade animation
function easeOutCubic(t: number): number {
  return 1 - Math.pow(1 - t, 3);
}

function computeStableProgress(currentNodes: SimulationNode[], totalNodes: number): number {
  if (totalNodes === 0) return 1;

  const stableNodes = currentNodes.filter(
    (n) =>
      n.x !== undefined &&
      !isNaN(n.x) &&
      n.y !== undefined &&
      !isNaN(n.y) &&
      Math.hypot(n.vx ?? 0, n.vy ?? 0) < 0.2
  ).length;

  return Math.min(stableNodes / totalNodes, 1);
}

function initializeOpacityMaps(
  nodes: SimulationNode[],
  links: SimulationLink[],
  state: SimulationState
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
  factor: number = 0.1
): void {
  opacityMap.forEach((currentOpacity, key) => {
    const newOpacity = currentOpacity + (targetOpacity - currentOpacity) * factor;
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
export function applyDelta(delta: GraphDeltaData, options: DeltaUpdateOptions): boolean {
  const domainDelta = GraphDelta.fromAPI(delta);

  if (import.meta.env.DEV) {
    console.log("[Delta] Applying delta with", domainDelta.totalChanges, "changes");
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
function applyFullRestart(delta: GraphDelta, options: DeltaUpdateOptions): boolean {
  const { nodes, links, width, height, state, onTick, onResetView } = options;

  if (import.meta.env.DEV) {
    console.log("[Delta] Full simulation restart");
  }

  // Фильтруем удаленные узлы
  const filteredNodes = nodes.filter((n) => !delta.removedNodeIds.includes(n.id));

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
        (removed) => removed.source === l.source && removed.target === l.target
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
  const simulationNodes: SimulationNode[] = filteredNodes.map((n, i) => {
    const angle = (i / filteredNodes.length) * 2 * Math.PI;
    const radius = Math.min(width, height) * 0.3;
    return {
      ...n,
      x: width / 2 + Math.cos(angle) * radius,
      y: height / 2 + Math.sin(angle) * radius,
    };
  });

  state.simLinks = filteredLinks.map((l) => ({
    id: l.id,
    source: l.source,
    target: l.target,
    weight: l.weight ?? 1,
    link_type: l.link_type,
    source_type: l.source_type,
    gamma_origin: l.gamma_origin,
    last_weight_update: l.last_weight_update,
  }));

  // Initialize opacity maps for fade effect using current nodes and links
  initializeOpacityMaps(filteredNodes, state.simLinks, state);

  const totalNodes = filteredNodes.length;
  let tickCount = 0;

  state.simulation = d3Force
    .forceSimulation<SimulationNode, SimulationLink>(simulationNodes)
    .force(
      "link",
      d3Force
        .forceLink<SimulationNode, SimulationLink>(state.simLinks)
        .id((d) => d.id)
        .distance(100)
        .strength(0.3)
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
          const newOpacity = currentOpacity + (targetOpacity - currentOpacity) * 0.15;
          state.nodeOpacity.set(nodeId, Math.min(Math.max(newOpacity, 0), 1));
        });

        state.linkOpacity.forEach((_, linkId) => {
          const currentOpacity = state.linkOpacity.get(linkId) || 0;
          const newOpacity = currentOpacity + (targetOpacity - currentOpacity) * 0.15;
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
 * Incremental delta application for small changes — SYNC-1, stage B.
 *
 * Everything lands in place on the live simulation: renamed nodes keep their
 * positions, new nodes grow out of placed neighbours via
 * `addNodesToSimulation`, removed nodes/links drop out of the link force and
 * the render maps. A full re-layout happens only for large deltas
 * (`requiresFullRestart` in `applyDelta`) — a busy import, not a single edit.
 */
interface ForceLinkLike {
  links(): SimulationLink[];
  links(links: SimulationLink[]): ForceLinkLike;
}

/** Identity used by delta matching: id when both sides carry it, else the
 * directed endpoints + type (same key as the client-side merge). */
export function deltaLinkKey(link: {
  id?: string;
  source: unknown;
  target: unknown;
  link_type?: string;
}): string {
  return `${getLinkEndpointId(link.source as SimulationLink["source"])}:${getLinkEndpointId(
    link.target as SimulationLink["target"]
  )}:${link.link_type ?? "related"}`;
}

function isRemovedLink(simLink: SimulationLink, removed: GraphLink[]): boolean {
  return removed.some(
    (r) => (r.id !== undefined && r.id === simLink.id) || deltaLinkKey(r) === deltaLinkKey(simLink)
  );
}

function applyIncremental(delta: GraphDelta, options: DeltaUpdateOptions): boolean {
  const { width, height, state } = options;

  if (import.meta.env.DEV) {
    console.log("[Delta] Incremental update");
  }

  const sim = state.simulation;
  let touched = false;

  // Updated nodes: mutate fields in place — position untouched.
  if (delta.updatedNodes.length > 0 && sim) {
    const simNodes = sim.nodes();
    delta.updatedNodes.forEach((updated) => {
      const simNode = simNodes.find((n) => n.id === updated.id);
      if (simNode) {
        simNode.title = updated.title;
        simNode.type = updated.type;
        if (updated.x !== undefined) simNode.x = updated.x;
        if (updated.y !== undefined) simNode.y = updated.y;
        touched = true;
      }
    });
  }

  // Removed nodes: drop the node plus every incident link.
  if (delta.removedNodeIds.length > 0 && sim) {
    const removedIds = new Set(delta.removedNodeIds);
    const remaining = sim.nodes().filter((n) => !removedIds.has(n.id));
    if (remaining.length !== sim.nodes().length) {
      sim.nodes(remaining);
      const incident = (l: SimulationLink) =>
        removedIds.has(getLinkEndpointId(l.source)) || removedIds.has(getLinkEndpointId(l.target));
      for (const l of state.simLinks.filter(incident)) {
        state.linkOpacity.delete(getLinkId(l));
      }
      state.simLinks = state.simLinks.filter((l) => !incident(l));
      const linkForce = sim.force("link") as ForceLinkLike | undefined;
      if (linkForce && typeof linkForce.links === "function") {
        linkForce.links(linkForce.links().filter((l) => !incident(l)));
      }
      for (const id of removedIds) state.nodeOpacity.delete(id);
      touched = true;
    }
  }

  // Removed links.
  if (delta.removedLinks.length > 0 && sim) {
    for (const l of state.simLinks.filter((l) => isRemovedLink(l, delta.removedLinks))) {
      state.linkOpacity.delete(getLinkId(l));
    }
    state.simLinks = state.simLinks.filter((l) => !isRemovedLink(l, delta.removedLinks));
    const linkForce = sim.force("link") as ForceLinkLike | undefined;
    if (linkForce && typeof linkForce.links === "function") {
      linkForce.links(linkForce.links().filter((l) => !isRemovedLink(l, delta.removedLinks)));
    }
    touched = true;
  }

  // Added links: the server sends changed links as `added` (set semantics),
  // so an existing same-key link is updated, not duplicated. Links arriving
  // alongside added nodes are handled by addNodesToSimulation below.
  const standaloneLinks: SimulationLink[] = [];
  if (delta.addedLinks.length > 0 && sim) {
    const byKey = new Map(state.simLinks.map((l) => [deltaLinkKey(l), l]));
    for (const added of delta.addedLinks) {
      const existing = byKey.get(deltaLinkKey(added));
      if (existing) {
        existing.id = added.id ?? existing.id;
        existing.weight = added.weight;
        existing.link_type = added.link_type;
        existing.source_type = added.source_type;
        existing.gamma_origin = added.gamma_origin;
        existing.last_weight_update = added.last_weight_update;
      } else {
        standaloneLinks.push(added as SimulationLink);
      }
      touched = true;
    }
  }

  // Added nodes — reuse the incremental adder from UI-LOAD-1: new nodes
  // spawn next to a placed neighbour, fade in, and the sim is reheated
  // gently, never rebuilt.
  if (delta.addedNodes.length > 0 && sim) {
    // Skip nodes the simulation already has — the server may resend an "added"
    // record for a node the client merged earlier.
    const existingIds = new Set(sim.nodes().map((n) => n.id));
    const freshNodes = delta.addedNodes.filter((n) => !existingIds.has(n.id));
    if (freshNodes.length > 0) {
      addNodesToSimulation(state, freshNodes as SimulationNode[], standaloneLinks, width, height);
      standaloneLinks.length = 0;
      touched = true;
    }
  }

  // Links between already-placed nodes still need appending.
  if (standaloneLinks.length > 0 && sim) {
    state.simLinks = [...state.simLinks, ...standaloneLinks];
    const linkForce = sim.force("link") as ForceLinkLike | undefined;
    if (linkForce && typeof linkForce.links === "function") {
      linkForce.links([...linkForce.links(), ...standaloneLinks]);
    }
    for (const l of standaloneLinks) state.linkOpacity.set(getLinkId(l), 0);
    touched = true;
  }

  if (touched && sim) {
    // A gentle nudge applies the change without re-layouting placed nodes.
    sim.alpha(0.3).restart();
    state.isRunning = true;
    state.stable = false;
    return true;
  }

  return false;
}
