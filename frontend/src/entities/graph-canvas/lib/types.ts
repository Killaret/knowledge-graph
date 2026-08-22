/**
 * Shared types for GraphCanvas modules
 */
import type { Simulation, SimulationNodeDatum, SimulationLinkDatum } from "d3-force";

export interface SimulationNode extends SimulationNodeDatum {
  id: string;
  title: string;
  type?: string;
  scale?: number;
  opacity?: number;
  createdAt?: string;
  /** Optional custom fill color; overrides the palette if set. */
  color?: string;
  /** Optional custom glow color; computed from fill if omitted. */
  glowColor?: string;
}

export interface SimulationLink extends SimulationLinkDatum<SimulationNode> {
  id?: string;
  source: SimulationNode | string | number;
  target: SimulationNode | string | number;
  weight?: number;
  link_type?: string;
  source_type?: string; // 'user' or 'gamma'
  last_weight_update?: string;
}

// Re-export node/link datum base properties explicitly for clarity
export type { SimulationNodeDatum, SimulationLinkDatum };

export interface SimulationState {
  simulation: Simulation<SimulationNode, SimulationLink> | null;
  simLinks: SimulationLink[];
  isRunning: boolean;
  stable: boolean;
  nodeOpacity: Map<string, number>;
  linkOpacity: Map<string, number>;
  dyingLinks: SimulationLink[];
  dyingLinkOpacity: Map<string, number>;
  fadeAnimationId: number | null;
}

export interface TransformState {
  x: number;
  y: number;
  k: number;
}

export interface DragState {
  dragging: boolean;
  dragStart: { x: number; y: number };
}

export interface ResizeState {
  width: number;
  height: number;
}

/**
 * Resolve a link endpoint reference to the actual simulation node.
 * d3-force allows `source`/`target` to be a node id string, an array index
 * number, or the resolved node object itself.
 */
export function resolveLinkEndpoint(
  ref: string | number | SimulationNode,
  nodes: SimulationNode[],
  nodeMap?: Map<string, SimulationNode>
): SimulationNode | undefined {
  if (typeof ref === "object" && ref !== null) {
    return ref;
  }
  if (typeof ref === "number") {
    return nodes[ref];
  }
  if (nodeMap) {
    return nodeMap.get(ref);
  }
  return nodes.find((n) => n.id === ref);
}

/** Extract the node id from a link endpoint reference. */
export function getLinkEndpointId(ref: string | number | { id: string }): string {
  if (typeof ref === "string") return ref;
  if (typeof ref === "number") return String(ref);
  return ref.id;
}
