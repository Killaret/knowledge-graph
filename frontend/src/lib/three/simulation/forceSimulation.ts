import { forceSimulation, forceLink, forceManyBody, forceCenter } from 'd3-force-3d';
import type { GraphData } from '$lib/api/graph';
import type { ObjectManager } from '$lib/three/rendering/objectManager';

export function createSimulation(data: GraphData, objectManager: ObjectManager) {
  const nodes = data.nodes.map(n => ({
    ...n,
    x: (n as any).x ?? (Math.random() - 0.5) * 100,
    y: (n as any).y ?? (Math.random() - 0.5) * 100,
    z: (n as any).z ?? (Math.random() - 0.5) * 100
  }));
  const links = data.links.map(l => ({
    ...l,
    source: l.source,
    target: l.target,
    value: l.weight ?? 1
  }));
  objectManager.createAll(nodes, links);

  const sim = forceSimulation(nodes)
    .force('link', forceLink(links)
      .id((d: any) => d.id)
      .distance(50)
      .strength(0.8)
    )
    .force('charge', forceManyBody()
      .strength(-200)
      .distanceMax(200)
    )
    .force('center', forceCenter(0, 0, 0));
  
  (sim as any).alphaDecay(0.015);
  
  let tickCount = 0;
  sim.on('tick', () => {
    objectManager.updatePositions(nodes);
    // Only update links every 10 ticks for performance, and log first few
    if (++tickCount % 10 === 0 || tickCount <= 3) {
      console.log(`[forceSimulation] Tick ${tickCount}, updating ${links.length} links`);
      objectManager.updateLinks(links);
    }
  });
  
  sim.on('end', () => {
    console.log(`[forceSimulation] Simulation ended after ${tickCount} ticks, final link update`);
    objectManager.updateLinks(links);
  });
  
  return sim;
}
