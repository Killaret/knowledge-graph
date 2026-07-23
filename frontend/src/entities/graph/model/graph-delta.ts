import type { GraphDeltaData, GraphLink, GraphNode } from "./graph-node";

export interface GraphDeltaProps {
  addedNodes: GraphNode[];
  removedNodeIds: string[];
  updatedNodes: GraphNode[];
  addedLinks: GraphLink[];
  removedLinks: GraphLink[];
  timestamp: Date;
  version: number;
}

/**
 * GraphDelta — Entity representing an incremental update to a graph.
 *
 * Keeps the decision logic (empty check, total change count, restart threshold)
 * separate from the D3-specific application code in `delta.ts`.
 */
export class GraphDelta {
  constructor(public readonly props: Readonly<GraphDeltaProps>) {}

  get addedNodes(): GraphNode[] {
    return this.props.addedNodes;
  }

  get removedNodeIds(): string[] {
    return this.props.removedNodeIds;
  }

  get updatedNodes(): GraphNode[] {
    return this.props.updatedNodes;
  }

  get addedLinks(): GraphLink[] {
    return this.props.addedLinks;
  }

  get removedLinks(): GraphLink[] {
    return this.props.removedLinks;
  }

  get timestamp(): Date {
    return this.props.timestamp;
  }

  get version(): number {
    return this.props.version;
  }

  get totalChanges(): number {
    return (
      this.addedNodes.length +
      this.removedNodeIds.length +
      this.updatedNodes.length +
      this.addedLinks.length +
      this.removedLinks.length
    );
  }

  isEmpty(): boolean {
    return this.totalChanges === 0;
  }

  requiresFullRestart(threshold: number = 10): boolean {
    return this.totalChanges > threshold;
  }

  merge(other: GraphDelta): GraphDelta {
    return new GraphDelta({
      addedNodes: [...this.addedNodes, ...other.addedNodes],
      removedNodeIds: [
        ...new Set([...this.removedNodeIds, ...other.removedNodeIds]),
      ],
      updatedNodes: [...this.updatedNodes, ...other.updatedNodes],
      addedLinks: [...this.addedLinks, ...other.addedLinks],
      removedLinks: [...this.removedLinks, ...other.removedLinks],
      timestamp: new Date(),
      version: this.version + 1,
    });
  }

  static fromAPI(data: GraphDeltaData, version: number = 0): GraphDelta {
    return new GraphDelta({
      addedNodes: data.added_nodes ?? [],
      removedNodeIds: data.removed_nodes ?? [],
      updatedNodes: data.updated_nodes ?? [],
      addedLinks: data.added_links ?? [],
      removedLinks: data.removed_links ?? [],
      timestamp: new Date(),
      version,
    });
  }

  static empty(version: number = 0): GraphDelta {
    return GraphDelta.fromAPI({}, version);
  }
}
