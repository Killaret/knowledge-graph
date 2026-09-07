/**
 * Synthetic 2D graph renderer performance benchmark (temporary).
 */
import { describe, it } from "vitest";
import { draw } from "./renderer";
import { createMockCanvasContext } from "./test-canvas-mock";
import type { SimulationNode, SimulationLink } from "./types";

const BODY_TYPES = ["star", "planet", "comet", "galaxy", "moon"];

function makeNodes(count: number): SimulationNode[] {
  return Array.from({ length: count }, (_, i) => ({
    id: `node-${i}`,
    title: `Note ${i}`,
    type: BODY_TYPES[i % BODY_TYPES.length],
    x: (i % 100) * 8 + Math.random() * 4,
    y: Math.floor(i / 100) * 6 + Math.random() * 4,
    vx: 0,
    vy: 0,
    index: i,
  }));
}

function makeLinks(nodes: SimulationNode[]): SimulationLink[] {
  const links: SimulationLink[] = [];
  for (let i = 0; i < nodes.length; i++) {
    const target = (i + 1) % nodes.length;
    links.push({
      source: nodes[i].id,
      target: nodes[target].id,
      weight: 0.5,
      link_type: "related",
    });
    const cross = (i + Math.floor(nodes.length / 3)) % nodes.length;
    if (cross !== i && cross !== target) {
      links.push({
        source: nodes[i].id,
        target: nodes[cross].id,
        weight: 0.7,
        link_type: "related",
      });
    }
  }
  return links;
}

function runBench(nodeCount: number, frames = 60): number {
  const ctx = createMockCanvasContext();
  const nodes = makeNodes(nodeCount);
  const links = makeLinks(nodes);
  const angles = new Map(nodes.map((n) => [n.id, 0]));
  const transform = { x: 0, y: 0, k: 1 };

  // warm-up
  draw(ctx, 800, 600, links, nodes, angles, transform);

  const start = performance.now();
  for (let i = 0; i < frames; i++) {
    draw(
      ctx,
      800,
      600,
      links,
      nodes,
      angles,
      transform,
      undefined,
      undefined,
      undefined,
      undefined,
      false,
      i * 16
    );
  }
  const total = performance.now() - start;
  return total / frames;
}

describe("2D graph renderer responsiveness", () => {
  const counts = [10, 50, 100, 200, 300, 500];
  for (const count of counts) {
    it(`renders ${count} nodes efficiently`, () => {
      const msPerFrame = runBench(count);
      console.log(
        `[PERF] ${count} nodes, ${(count * 1.5).toFixed(0)} links: ${msPerFrame.toFixed(3)} ms/frame`
      );
    });
  }
});
