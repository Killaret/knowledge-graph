/**
 * Animation loop management for GraphCanvas
 */

import { getVariation } from "$shared/utils/variation";
import { CelestialBody } from "$entities";

export interface AnimationState {
  animationId: number;
  angles: Map<string, number>;
  speeds: Map<string, number>;
}

/**
 * Calculate rotation speed based on the CelestialBody configuration.
 */
export function getBaseSpeed(type: string): number {
  return CelestialBody.fromString(type).baseSpeed;
}

/**
 * Map a node's deterministic phase shift to a broad speed multiplier.
 * Result includes both magnitude (0.3x–1.8x base) and direction,
 * so nodes of the same type quickly desynchronize instead of rotating in lockstep.
 */
function getSpeedFromVariation(id: string, type: string): number {
  const baseSpeed = getBaseSpeed(type);
  const variation = getVariation(id, type);
  const phase = variation.phaseShift;

  // phaseShift is uniform in [0, 2π]. Split it into direction and magnitude.
  const direction = phase > Math.PI ? -1 : 1;
  const normalized = (phase % Math.PI) / Math.PI;
  const multiplier = 0.3 + normalized * 1.5;

  return baseSpeed * multiplier * direction;
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
    const type = node.type ?? "star";

    let speed = speeds.get(id);
    if (speed === undefined) {
      if (disableVariation) {
        // In stable render mode we want deterministic, non-animated snapshots
        // Keep the angle fixed at 0 so stable render snapshots remain
        // consistent with baseline expectations.
        angles.set(id, 0);
        speed = 0;
      } else {
        // Initialize the angle from the node's phase shift so every node starts
        // at a different rotation. The speed is derived from the same variation
        // but with a much wider spread and possible reverse direction.
        const variation = getVariation(id, type);
        if (!angles.has(id)) {
          angles.set(id, variation.phaseShift);
        }
        speed = getSpeedFromVariation(id, type);
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
  onUpdate: (timestamp: number) => void,
  disableVariation: boolean = false
): { stop: () => void } {
  let animationId: number;
  const angles = new Map<string, number>();
  const speeds = new Map<string, number>();

  function animate(timestamp: number) {
    const nodes = getNodes();
    updateNodeAngles(nodes, angles, speeds, disableVariation);
    onUpdate(timestamp);
    animationId = requestAnimationFrame(animate);
  }

  animate(performance.now());

  return {
    stop: () => {
      cancelAnimationFrame(animationId);
    },
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
