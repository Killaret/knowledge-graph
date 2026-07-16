/**
 * Gravity system for subtle node attraction and background lens distortion.
 *
 * Tunable parameters are sourced from knowledge-graph.config.json:
 *   frontend.graph.2d.gravity_nodes_threshold  — node count above which gravity is disabled
 *   frontend.graph.2d.gravity_max_distance     — max world-unit attraction radius
 *
 * GRAVITY_COEFFICIENT is kept local — it is a physics constant, not a user-tunable setting.
 */

import type { SimulationNode } from './types';
import { graphConfig2D } from '$shared/config';

/** Physics spring constant — controls attraction force magnitude */
const GRAVITY_COEFFICIENT = 0.0001;
/** Max world-unit distance within which gravity is applied (from config) */
const MAX_GRAVITY_DISTANCE = graphConfig2D.gravity_max_distance;
/** Node count above which gravity system is disabled for performance (from config) */
const PERFORMANCE_THRESHOLD_NODES = graphConfig2D.gravity_nodes_threshold;

export interface GravitySystem {
  applyAttraction: (nodes: SimulationNode[]) => void;
  getDistortion: (
    x: number,
    y: number,
    nodes: SimulationNode[],
    maxDistance?: number
  ) => { dx: number; dy: number };
  isEnabled: (nodeCount: number) => boolean;
}

export function createGravitySystem(): GravitySystem {
  return {
    applyAttraction(nodes: SimulationNode[]) {
      if (!this.isEnabled(nodes.length)) return;

      const n = nodes.length;
      for (let i = 0; i < n; i++) {
        const nodeA = nodes[i];
        if (nodeA.x == null || nodeA.y == null) continue;

        for (let j = i + 1; j < n; j++) {
          const nodeB = nodes[j];
          if (nodeB.x == null || nodeB.y == null) continue;

          const dx = nodeB.x - nodeA.x;
          const dy = nodeB.y - nodeA.y;
          const distanceSq = dx * dx + dy * dy;

          if (distanceSq === 0 || distanceSq > MAX_GRAVITY_DISTANCE * MAX_GRAVITY_DISTANCE) continue;

          const distance = Math.sqrt(distanceSq);
          const force = GRAVITY_COEFFICIENT / (distance * 0.1 + 1);
          const fx = (dx / distance) * force;
          const fy = (dy / distance) * force;

          nodeA.x += fx;
          nodeA.y += fy;
          nodeB.x -= fx;
          nodeB.y -= fy;
        }
      }
    },

    getDistortion(
      x: number,
      y: number,
      nodes: SimulationNode[],
      maxDistance: number = 100
    ): { dx: number; dy: number } {
      if (!this.isEnabled(nodes.length)) return { dx: 0, dy: 0 };

      let totalDx = 0;
      let totalDy = 0;
      let totalWeight = 0;

      for (const node of nodes) {
        if (node.x == null || node.y == null) continue;

        const dx = node.x - x;
        const dy = node.y - y;
        const distance = Math.sqrt(dx * dx + dy * dy);

        if (distance === 0 || distance >= maxDistance) continue;

        const maxOffset = node.type === 'star' ? 20 : node.type === 'planet' ? 15 : 10;
        const weight = 1 - distance / maxDistance;
        const offset = maxOffset * weight;

        totalDx += (dx / distance) * offset;
        totalDy += (dy / distance) * offset;
        totalWeight += weight;
      }

      if (totalWeight === 0) return { dx: 0, dy: 0 };

      return {
        dx: totalDx / totalWeight,
        dy: totalDy / totalWeight
      };
    },

    isEnabled(nodeCount: number) {
      return nodeCount <= PERFORMANCE_THRESHOLD_NODES;
    }
  };
}

export function drawDistortedBackgroundGrid(
  ctx: CanvasRenderingContext2D,
  width: number,
  height: number,
  nodes: SimulationNode[],
  animationTime: number,
  gridSize: number = 100
): void {
  const gravity = createGravitySystem();
  if (!gravity.isEnabled(nodes.length)) return;

  const maxDistance = 100;

  ctx.save();
  ctx.strokeStyle = 'rgba(255, 255, 255, 0.03)';
  ctx.lineWidth = 1;

  // Vertical lines
  for (let x = 0; x <= width; x += gridSize) {
    ctx.beginPath();
    for (let y = 0; y <= height; y += 10) {
      const distortion = gravity.getDistortion(x, y, nodes, maxDistance);
      const px = x + distortion.dx;
      const py = y + distortion.dy;
      if (y === 0) {
        ctx.moveTo(px, py);
      } else {
        ctx.lineTo(px, py);
      }
    }
    ctx.stroke();
  }

  // Horizontal lines
  for (let y = 0; y <= height; y += gridSize) {
    ctx.beginPath();
    for (let x = 0; x <= width; x += 10) {
      const distortion = gravity.getDistortion(x, y, nodes, maxDistance);
      const px = x + distortion.dx;
      const py = y + distortion.dy;
      if (x === 0) {
        ctx.moveTo(px, py);
      } else {
        ctx.lineTo(px, py);
      }
    }
    ctx.stroke();
  }

  ctx.restore();
}
