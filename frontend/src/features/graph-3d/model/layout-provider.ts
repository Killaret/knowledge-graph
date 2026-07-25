import type { GraphData, GraphLink } from "$shared/api/graph";
import { getFullGraphData, getGraphData } from "$shared/api/graph";
import { toSimulationNodes } from "../config";
import type { SimulationNode } from "./types";

/**
 * Abstract adapter for 3D initial layout computation.
 *
 * Implementations may use a client-side force simulation (D3) or call the
 * graph-service to obtain pre-computed 3D coordinates.
 */
export interface GraphLayoutProvider {
  load(params: GraphLayoutParams): Promise<GraphData>;
}

export interface GraphLayoutParams {
  noteId?: string;
  depth?: number;
  userId?: string;
  limit?: number;
}

/**
 * Client-side D3 force layout provider.
 *
 * Returns the input graph as-is; Graph3DEngine will run d3-force-3d to refine
 * positions on the client.
 */
export class D3ForceLayoutProvider implements GraphLayoutProvider {
  async load(params: GraphLayoutParams): Promise<GraphData> {
    if (params.noteId) {
      return getGraphData(params.noteId, params.depth ?? 2, params.userId);
    }
    return getFullGraphData(params.limit, params.userId);
  }
}

/**
 * Graph-service 3D layout provider.
 *
 * Calls graph-service endpoints that return 3D positions. For note graphs it
 * passes `layout=3d`; for full graphs the service already returns 3D.
 */
export class GraphServiceLayoutProvider implements GraphLayoutProvider {
  async load(params: GraphLayoutParams): Promise<GraphData> {
    if (params.noteId) {
      return getGraphData(params.noteId, params.depth ?? 2, params.userId, "3d");
    }
    return getFullGraphData(params.limit, params.userId);
  }
}

/**
 * Convert API graph data into simulation-ready nodes, optionally merging in
 * service-provided coordinates.
 */
export function prepareSimulationData(data: GraphData): {
  nodes: SimulationNode[];
  links: GraphLink[];
} {
  return {
    nodes: toSimulationNodes(data.nodes, data.links),
    links: data.links,
  };
}
