/**
 * Animation loop management for GraphCanvas
 */

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
  speeds: Map<string, number>
): void {
  for (const node of nodes) {
    const id = node.id;
    const type = node.type ?? 'star';
    const baseSpeed = getBaseSpeed(type);

    let speed = speeds.get(id);
    if (speed === undefined) {
      speed = baseSpeed * (0.7 + Math.random() * 0.6);
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
  onUpdate: () => void
): { stop: () => void } {
  let animationId: number;
  const angles = new Map<string, number>();
  const speeds = new Map<string, number>();

  function animate() {
    const nodes = getNodes();
    updateNodeAngles(nodes, angles, speeds);
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
