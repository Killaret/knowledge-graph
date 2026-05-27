/**
 * Animation loop management for GraphCanvas
 */

import { getVariation } from '$lib/utils/variation';

export interface AnimationState {
  animationId: number;
  angles: Map<string, number>;
  speeds: Map<string, number>;
}

/**
 * Calculate rotation speed based on node type
 */
export function getBaseSpeed(type: string): number {
  switch (type) {
    case 'planet':
      return 0.02;
    case 'comet':
      return 0.03;
    case 'galaxy':
      return 0.01;
    case 'nebula':
      return 0.008;
    case 'star':
    default:
      return 0.005;
  }
}

/**
 * Update rotation angles for all nodes
 */
export function updateNodeAngles(
  nodes: Array<{ id: string; type?: string }>,
  angles: Map<string, number>,
  speeds: Map<string, number>,
  disableVariation: boolean = false
): void {
  for (const node of nodes) {
    const id = node.id;
    const type = node.type ?? 'star';
    const baseSpeed = getBaseSpeed(type);

    let speed = speeds.get(id);
    if (speed === undefined) {
      if (disableVariation) {
        // In stable render mode we want deterministic, non-animated snapshots
        // Keep the angle fixed at 0 so stable render snapshots remain
        // consistent with baseline expectations.
        angles.set(id, 0);
        speed = 0;
      } else {
        // Get variation for this node to determine speed multiplier
        const variation = getVariation(id, type);
        const speedMultiplier = 0.8 + variation.phaseShift * 0.1;
        speed = baseSpeed * speedMultiplier;
      }
      speeds.set(id, speed);
    }

    const current = angles.get(id) || 0;
    angles.set(id, current + speed);
  }
}

/**
 * Start the animation loop
 */
export function startAnimationLoop(
  getNodes: () => Array<{ id: string; type?: string }>,
  onUpdate: () => void,
  disableVariation: boolean = false
): { stop: () => void } {
  let animationId: number;
  const angles = new Map<string, number>();
  const speeds = new Map<string, number>();

  function animate() {
    const nodes = getNodes();
    updateNodeAngles(nodes, angles, speeds, disableVariation);
    onUpdate();
    animationId = requestAnimationFrame(animate);
  }

  animate();

  return {
    stop: () => {
      cancelAnimationFrame(animationId);
    }
  };
}

/**
 * Clear animation state
 */
export function clearAnimationState(
  angles: Map<string, number>,
  speeds: Map<string, number>
): void {
  angles.clear();
  speeds.clear();
}
