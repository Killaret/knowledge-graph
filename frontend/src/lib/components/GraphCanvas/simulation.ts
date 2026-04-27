/**
 * D3-force simulation management for GraphCanvas
 */
import * as d3Force from 'd3-force';
import { filterValidLinks } from '$lib/utils/graphUtils';
import type { SimulationNode, SimulationLink, SimulationState, TransformState } from './types';

export type { SimulationNode, SimulationLink, SimulationState, TransformState };

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

  // Stop existing simulation if any
  if (state.simulation) {
    state.simulation.stop();
  }

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
  state.simLinks = [];
  state.isRunning = false;
}
