import { forceSimulation, forceLink, forceManyBody, forceCenter } from 'd3-force-3d';
import type { GraphData } from '$lib/api/graph';
import type { ObjectManager } from '$lib/three/rendering/objectManager';

export function createSimulation(data: GraphData, objectManager: ObjectManager) {
  const nodes = data.nodes.map(n => ({
    ...n,
    x: (n as any).x ?? (Math.random() - 0.5) * 50,
    y: (n as any).y ?? (Math.random() - 0.5) * 50,
    z: (n as any).z ?? (Math.random() - 0.5) * 50
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
      .distance(30)
      .strength(0.5)
    )
    .force('charge', forceManyBody()
      .strength(-50)
      .distanceMax(150)
    )
    .force('center', forceCenter(0, 0, 0));
  
  (sim as any).alphaDecay(0.02);
  sim.on('tick', () => {
    objectManager.updatePositions(nodes);
    objectManager.updateLinks(links);
  });
  
  return sim;
}
