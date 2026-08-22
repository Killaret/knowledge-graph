/**
 * Simulation module - wraps $lib/three/simulation/forceSimulation
 */
import { createSimulation } from '$lib/three/simulation/forceSimulation';
import type { GraphNode, GraphLink, Simulation, SimulationNode } from './types';

export { createSimulation };

export interface SimulationCallbacks {
  onTick?: () => void;
  onEnd?: () => void;
}

/**
 * Add nodes to existing simulation
 */
export function addNodesToSimulation(
  simulation: Simulation,
  nodes: GraphNode[],
  links?: GraphLink[],
  callbacks?: SimulationCallbacks
): void {
  if (!simulation) return;

  // Update simulation nodes
  const currentNodes = simulation.nodes();
  simulation.nodes([...currentNodes, ...(nodes as SimulationNode[])]);

  // Setup callbacks if provided
  if (callbacks?.onTick) {
    simulation.on('tick', callbacks.onTick);
  }
  if (callbacks?.onEnd) {
    simulation.on('end', callbacks.onEnd);
  }
}

/**
 * Setup simulation with nodes and links
 */
export function setupSimulation(
  simulation: Simulation,
  nodes: GraphNode[],
  links: GraphLink[],
  callbacks?: SimulationCallbacks
): void {
  if (!simulation) return;

  // Add nodes to simulation
  addNodesToSimulation(simulation, nodes);

  // Setup tick callback
  if (callbacks?.onTick) {
    simulation.on('tick', callbacks.onTick);
  }

  // Setup end callback
  if (callbacks?.onEnd) {
    simulation.on('end', callbacks.onEnd);
  }
}

/**
 * Restart simulation with alpha
 */
export function restartSimulation(simulation: Simulation, alpha: number = 1): void {
  if (!simulation) return;
  simulation.alpha(alpha).restart();
}

/**
 * Get simulation nodes
 */
export function getSimulationNodes(simulation: Simulation): SimulationNode[] {
  if (!simulation) return [];
  return simulation.nodes() || [];
}

/**
 * Update simulation with new nodes and links
 */
export function updateSimulation(
  simulation: Simulation,
  nodes: GraphNode[],
  links: GraphLink[],
  callbacks?: SimulationCallbacks
): void {
  if (!simulation) return;

  // Update nodes
  simulation.nodes(nodes as SimulationNode[]);

  // Update links
  simulation.force('link', null); // Remove old link force
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  simulation.force('link', ((sim: unknown) => {
    // Create new link force with updated links
    // This is a simplified version - actual implementation depends on forceSimulation
    (sim as { links?: (links: GraphLink[]) => void }).links?.(links);
  }) as any);

  // Restart simulation
  simulation.alpha(1).restart();

  // Setup callbacks
  if (callbacks?.onTick) {
    simulation.on('tick', callbacks.onTick);
  }
  if (callbacks?.onEnd) {
    simulation.on('end', callbacks.onEnd);
  }
}

/**
 * Stop simulation
 */
export function stopSimulation(simulation: Simulation): void {
  if (!simulation) return;
  simulation.stop();
}

/**
 * Clear simulation data
 */
export function clearSimulation(simulation: Simulation): void {
  if (!simulation) return;
  stopSimulation(simulation);
}
