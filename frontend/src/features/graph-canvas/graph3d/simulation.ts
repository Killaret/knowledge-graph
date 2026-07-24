import { forceSimulation, forceLink, forceManyBody, forceCenter } from "d3-force-3d";
import type { SimulationLinkDatum } from "d3-force-3d";
import type { GraphLink } from "$shared/api/graph";
import type { SimulationNode } from "./types";

type D3Simulation = ReturnType<typeof forceSimulation<SimulationNode>>;

export function createGraphSimulation(nodes: SimulationNode[], links: GraphLink[]): D3Simulation {
  // Initialize positions in a deterministic golden spiral if missing
  for (let i = 0; i < nodes.length; i++) {
    const node = nodes[i];
    if (node.x == null || Number.isNaN(node.x)) {
      const r = 25 * Math.sqrt(0.5 + i);
      const angle = i * 2.39996;
      node.x = r * Math.cos(angle);
      node.y = r * Math.sin(angle);
      node.z = (Math.random() - 0.5) * 20;
    }
    if (node.y == null || Number.isNaN(node.y)) node.y = 0;
    if (node.z == null || Number.isNaN(node.z)) node.z = 0;
    node.vx = node.vy = node.vz = 0;
  }

  const sim = forceSimulation(nodes)
    .force(
      "link",
      forceLink<SimulationNode, SimulationLinkDatum<SimulationNode>>(links)
        .id((d: SimulationNode) => d.id)
        .distance(40)
        .strength(0.5)
    )
    .force("charge", forceManyBody<SimulationNode>().strength(-150).distanceMax(150))
    .force("center", forceCenter<SimulationNode>(0, 0, 0))
    .alphaDecay(0.02)
    .alphaMin(0.001)
    .velocityDecay(0.4)
    .stop();

  return sim;
}
