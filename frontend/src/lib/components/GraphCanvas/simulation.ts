/**
 * D3-force simulation management for GraphCanvas
 */
import * as d3Force from 'd3-force';
import { filterValidLinks } from '$lib/utils/graphUtils';
import type { SimulationNode, SimulationLink, SimulationState, TransformState } from './types';

export type { SimulationNode, SimulationLink, SimulationState, TransformState };

// Easing function for smooth fade animation
function easeOutCubic(t: number): number {
  return 1 - Math.pow(1 - t, 3);
}

// Initialize opacity maps with 0 for all nodes and links
function initializeOpacityMaps(
  nodes: SimulationNode[],
  links: SimulationLink[],
  state: SimulationState
): void {
  state.nodeOpacity = new Map();
  state.linkOpacity = new Map();

  nodes.forEach(node => {
    state.nodeOpacity.set(node.id, 0);
  });

  links.forEach((link, index) => {
    const linkId = `${link.source}-${link.target}-${index}`;
    state.linkOpacity.set(linkId, 0);
  });
}

// Interpolate opacity values towards target
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
  onResetView: () => void
): void {
  if (!d3Force) {
    return;
  }

  // Reset transform when starting new simulation
  transform.x = 0;
  transform.y = 0;
  transform.k = 1;

  // Distribute nodes in a circle instead of single point (prevents extreme coordinates)
  const simulationNodes = nodes.map((n, i) => {
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
  if (validLinks.length !== links.length) {
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

  // Initialize opacity maps for fade effect
  initializeOpacityMaps(nodes, links, state);

  const totalNodes = nodes.length;
  let tickCount = 0;

  state.simulation = d3Force
    .forceSimulation(simulationNodes as any)
    .force(
      'link',
      d3Force
        .forceLink(edges)
        .id((d: any) => d.id)
        .distance(100)
        .strength(0.3)
    )
    .force('charge', d3Force.forceManyBody().strength(-150)) // Reduced repulsion
    .force('center', d3Force.forceCenter(width / 2, height / 2).strength(0.5))
    .force('collision', d3Force.forceCollide().radius(30))
    .alphaDecay(0.01) // Slower cooling for stability
    .on('tick', () => {
      onTick();
      tickCount++;

      // Update fade effect every 5 ticks
      if (tickCount % 5 === 0 && state.simulation) {
        const currentNodes = state.simulation.nodes();
        const nodesWithPosition = currentNodes.filter((n: any) =>
          n.x !== undefined && !isNaN(n.x) &&
          n.y !== undefined && !isNaN(n.y)
        ).length;

        const progress = Math.min(nodesWithPosition / totalNodes, 1);
        const targetOpacity = easeOutCubic(progress);

        interpolateOpacity(state.nodeOpacity, targetOpacity, 0.1);
        interpolateOpacity(state.linkOpacity, targetOpacity, 0.1);
      }
    })
    .on('end', () => {
      // Final fade animation to full opacity
      if (state.fadeAnimationId !== null) {
        cancelAnimationFrame(state.fadeAnimationId);
      }

      const startTime = performance.now();
      const duration = 2400; // 2.4 seconds for final fade

      const animateFinalFade = (currentTime: number) => {
        const elapsed = currentTime - startTime;
        const progress = Math.min(elapsed / duration, 1);
        const targetOpacity = easeOutCubic(progress);

        interpolateOpacity(state.nodeOpacity, targetOpacity, 0.15);
        interpolateOpacity(state.linkOpacity, targetOpacity, 0.15);

        if (progress < 1) {
          state.fadeAnimationId = requestAnimationFrame(animateFinalFade);
        } else {
          state.fadeAnimationId = null;
        }
      };

      state.fadeAnimationId = requestAnimationFrame(animateFinalFade);
    });

  // Warmup: run simulation synchronously for initial positioning
  if (state.simulation) {
    for (let i = 0; i < 50; i++) {
      state.simulation.tick();
    }

    // Compute transform BEFORE first draw
    onResetView();

    // Then start the animation
    state.simulation.alpha(1).restart();
  }
  state.isRunning = true;
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
}
