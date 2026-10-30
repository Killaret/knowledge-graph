import { forceSimulation, forceLink, forceManyBody, forceCenter } from 'd3-force-3d';
import type { GraphData, GraphNode, GraphLink } from '$lib/api/graph';
import type { ObjectManager } from '$lib/three/rendering/objectManager';

export function createSimulation(data: GraphData, objectManager: ObjectManager) {
  const nodes = data.nodes.map((n: GraphNode) => ({
    ...n,
    x: (n as any).x ?? (Math.random() - 0.5) * 100,
    y: (n as any).y ?? (Math.random() - 0.5) * 100,
    z: (n as any).z ?? (Math.random() - 0.5) * 100
  }));
  const links = data.links.map((l: GraphLink) => ({
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
      objectManager.updateLinks(links);
    }
  });
  
  sim.on('end', () => {
    objectManager.updateLinks(links);
  });
  
  return sim;
}

/**
 * Add new nodes and links to an existing simulation without restarting completely
 * This allows for progressive loading while keeping existing nodes stable
 */
export function addNodesToSimulation(
  sim: any,
  newData: GraphData,
  existingData: GraphData,
  objectManager: ObjectManager
): void {
  // Get existing node IDs
  const existingNodeIds = new Set(existingData.nodes.map((n: GraphNode) => n.id));
  const existingLinkIds = new Set(existingData.links.map((l: GraphLink) => `${l.source}-${l.target}`));
  
  // Filter out only new nodes and links
  const newNodes = newData.nodes.filter((n: GraphNode) => !existingNodeIds.has(n.id)).map((n: GraphNode) => ({
    ...n,
    x: (n as any).x ?? (Math.random() - 0.5) * 100,
    y: (n as any).y ?? (Math.random() - 0.5) * 100,
    z: (n as any).z ?? (Math.random() - 0.5) * 100
  }));
  
  const newLinks = newData.links.filter((l: GraphLink) => {
    const linkId = `${l.source}-${l.target}`;
    return !existingLinkIds.has(linkId);
  }).map((l: GraphLink) => ({
    ...l,
    source: l.source,
    target: l.target,
    value: l.weight ?? 1
  }));
  
  if (newNodes.length === 0 && newLinks.length === 0) {
    return;
  }
  
  // Add new visual objects
  objectManager.addNodes(newNodes, newLinks);
  
  // Get current nodes and links from simulation
  const currentNodes = sim.nodes();
  const currentLinks = sim.force('link').links();
  
  // Add new nodes to simulation
  currentNodes.push(...newNodes);
  
  // Add new links to simulation
  currentLinks.push(...newLinks);
  
  // "Warm up" the simulation with reduced charge strength for stability
  const chargeForce = sim.force('charge');
  if (chargeForce) {
    // Temporarily reduce repulsion for smoother integration
    (chargeForce as any).strength(-100);
  }
  
  // Restart simulation with controlled alpha target
  sim.alphaTarget(0.1).restart();
  
  // Gradually restore original charge strength
  setTimeout(() => {
    if (chargeForce) {
      (chargeForce as any).strength(-200);
    }
    // Settle to lower alpha
    sim.alphaTarget(0);
  }, 1000);
}
