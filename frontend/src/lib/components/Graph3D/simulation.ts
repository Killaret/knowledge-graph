/**
 * Simulation module - wraps $lib/three/simulation/forceSimulation
 */
import {
  createSimulation,
  addNodesToSimulation
} from '$lib/three/simulation/forceSimulation';

export {
  createSimulation,
  addNodesToSimulation
};

import type { GraphNode, GraphLink } from './types';

export interface SimulationCallbacks {
  onTick?: () => void;
  onEnd?: () => void;
}

/**
 * Setup simulation with nodes and links
 */
export function setupSimulation(
  simulation: any,
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
export function restartSimulation(simulation: any, alpha: number = 1): void {
  if (!simulation) return;
  simulation.alpha(alpha).restart();
}

/**
 * Get simulation nodes
 */
export function getSimulationNodes(simulation: any): any[] {
  if (!simulation) return [];
  return simulation.nodes() || [];
}

/**
 * Update simulation with new nodes and links
 */
export function updateSimulation(
  simulation: any,
  nodes: GraphNode[],
  links: GraphLink[],
  callbacks?: SimulationCallbacks
): void {
  if (!simulation) return;

  // Update nodes
  simulation.nodes(nodes);

  // Update links
  simulation.force('link', null); // Remove old link force
  simulation.force('link', (simulation: any) => {
    // Create new link force with updated links
    // This is a simplified version - actual implementation depends on forceSimulation
    simulation.links?.(links);
  });

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
export function stopSimulation(simulation: any): void {
  if (!simulation) return;
  simulation.stop();
}

/**
 * Clear simulation data
 */
export function clearSimulation(simulation: any): void {
  if (!simulation) return;
  stopSimulation(simulation);
}
