/**
 * D3-force simulation management for GraphCanvas
 */
import * as d3Force from 'd3-force';
import { filterValidLinks } from '$shared/utils/graphUtils';
import type { SimulationNode, SimulationLink, SimulationState, TransformState } from './types';

export type { SimulationNode, SimulationLink, SimulationState, TransformState };

// Easing function for smooth fade animation
function easeOutCubic(t: number): number {
  return 1 - Math.pow(1 - t, 3);
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

// Initialize opacity maps with 0 for all nodes and links
function initializeOpacityMaps(
  nodes: SimulationNode[],
  links: SimulationLink[],
  state: SimulationState
): void {
  state.nodeOpacity = new Map();
  state.linkOpacity = new Map();

  // Initialize link opacity (links fade in after their source/target nodes)
  links.forEach((link, index) => {
    const linkId = `${link.source}-${link.target}-${index}`;
    state.linkOpacity.set(linkId, 0);
  });

  // Initialize node opacity
  nodes.forEach(node => {
    state.nodeOpacity.set(node.id, 0);
  });
}

function startFadeAnimation(state: SimulationState, totalNodes: number, onStable?: () => void): void {
  if (state.fadeAnimationId !== null) {
    cancelAnimationFrame(state.fadeAnimationId);
    state.fadeAnimationId = null;
  }

  const animationStartTime = performance.now();
  const WAVE_DURATION = 100; // ms for full wave animation (instant fade-in)

  const animateFade = (timestamp: number) => {
    if (!state.simulation) {
      state.fadeAnimationId = null;
      return;
    }

    const elapsedTime = timestamp - animationStartTime;
    const waveProgress = Math.min(elapsedTime / WAVE_DURATION, 1);

    // Update node opacity
    state.nodeOpacity.forEach((currentOpacity, nodeId) => {
      const targetOpacity = easeOutCubic(Math.min(waveProgress, 1));
      const newOpacity = currentOpacity + (targetOpacity - currentOpacity) * 0.3; // Faster interpolation
      state.nodeOpacity.set(nodeId, Math.min(Math.max(newOpacity, 0), 1));
    });

    // Update link opacity (links fade in after nodes)
    state.linkOpacity.forEach((currentOpacity, linkId) => {
      const targetOpacity = easeOutCubic(Math.min(waveProgress, 1));
      const newOpacity = currentOpacity + (targetOpacity - currentOpacity) * 0.3; // Faster interpolation
      state.linkOpacity.set(linkId, Math.min(Math.max(newOpacity, 0), 1));
    });

    if (waveProgress < 1 || anyOpacityBelowOne(state)) {
      state.fadeAnimationId = requestAnimationFrame(animateFade);
    } else {
      state.fadeAnimationId = null;
      // Stop simulation after fade-in to prevent jitter
      if (state.simulation) {
        state.simulation.stop();
      }
      state.isRunning = false;
      state.stable = true;
      onStable?.();
    }
  };

  state.fadeAnimationId = requestAnimationFrame(animateFade);
}

/**
 * Start the force-directed graph simulation
 */
export function startSimulation(
  nodes: SimulationNode[],
  links: SimulationLink[],
  width: number,
  height: number,
  state: SimulationState,
  transform: TransformState,
  onTick: () => void,
  onResetView: () => void,
  onStable?: () => void
): void {
  if (!d3Force) {
    return;
  }

  // Reset transform when starting new simulation
  transform.x = 0;
  transform.y = 0;
  transform.k = 1;

  // Distribute nodes in a circle instead of single point (prevents extreme coordinates)
  // Preserve fixed-position nodes (e.g. Knowledge Core) that already have x/y/fx/fy.
  const simulationNodes = nodes.map((n, i) => {
    if (n.fx != null && n.fy != null && n.x != null && n.y != null) {
      return { ...n };
    }
    const angle = (i / nodes.length) * 2 * Math.PI;
    const radius = Math.min(width, height) * 0.3; // 30% of smaller dimension
    return {
      ...n,
      x: width / 2 + Math.cos(angle) * radius,
      y: height / 2 + Math.sin(angle) * radius
    };
  });

  if (import.meta.env.DEV) {
    console.log('[simulation] Initial node coordinates:', simulationNodes.slice(0, 3).map(n => ({ id: n.id, x: n.x, y: n.y })));
  }

  // Filter links to only include those where both source and target nodes exist
  const validLinks = filterValidLinks(nodes, links);
  if (validLinks.length !== links.length && import.meta.env.DEV) {
    console.warn(`[GraphCanvas] Filtered out ${links.length - validLinks.length} orphan links`);
  }

  const edges = validLinks.map((l) => ({
    source: l.source,
    target: l.target,
    weight: l.weight ?? 1,
    link_type: l.link_type
  }));

  state.simLinks = edges;

  // Stop existing simulation and fade animation if any
  if (state.simulation) {
    state.simulation.stop();
  }
  if (state.fadeAnimationId !== null) {
    cancelAnimationFrame(state.fadeAnimationId);
    state.fadeAnimationId = null;
  }

  // Initialize opacity maps for fade effect using filtered links
  initializeOpacityMaps(nodes, edges, state);

  const totalNodes = nodes.length;

  state.simulation = d3Force
    .forceSimulation(simulationNodes as any)
    .force(
      'link',
      d3Force
        .forceLink(edges)
        .id((d: any) => d.id)
        .distance(100)
        .strength(0.2)
    )
    .force('charge', d3Force.forceManyBody().strength(-100)) // Further reduced repulsion
    .force('center', d3Force.forceCenter(width / 2, height / 2).strength(0.3))
    .force('collision', d3Force.forceCollide().radius(25))
    .alphaDecay(0.1) // Even faster cooling
    .on('tick', () => {
      onTick();
    })
    .on('end', () => {
      // Simulation ended - nodes are stable
      // No additional fade animation needed (already handled in startFadeAnimation)
      state.fadeAnimationId = null;
    });

  // Warmup: run simulation synchronously for initial positioning
  if (state.simulation) {
    for (let i = 0; i < 200; i++) {
      state.simulation.tick();
    }

    // Compute transform BEFORE first draw
    onResetView();

    // Then start the animation
    state.simulation.alpha(1).restart();
    startFadeAnimation(state, totalNodes, onStable);
  }
  state.isRunning = true;
  state.stable = false;
}

/**
 * Stop the simulation
 */
export function stopSimulation(state: SimulationState): void {
  if (state.simulation) {
    state.simulation.stop();
    state.isRunning = false;
  }
  // Stop fade animation if running
  if (state.fadeAnimationId !== null) {
    cancelAnimationFrame(state.fadeAnimationId);
    state.fadeAnimationId = null;
  }
}

/**
 * Restart the simulation with alpha heat
 */
export function restartSimulation(state: SimulationState): void {
  if (state.simulation) {
    state.simulation.alpha(1).restart();
    state.isRunning = true;
  }
}

/**
 * Get current simulation nodes
 */
export function getSimulationNodes(state: SimulationState): SimulationNode[] {
  return state.simulation?.nodes() || [];
}

/**
 * Clear simulation state
 */
export function clearSimulation(state: SimulationState): void {
  if (state.simulation) {
    state.simulation.stop();
    state.simulation = null;
  }
  // Stop fade animation if running
  if (state.fadeAnimationId !== null) {
    cancelAnimationFrame(state.fadeAnimationId);
    state.fadeAnimationId = null;
  }
  state.simLinks = [];
  state.nodeOpacity = new Map();
  state.linkOpacity = new Map();
  state.isRunning = false;
  state.stable = false;
}
