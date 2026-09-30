/**
 * Dependency-chain analysis (LINK-TYPES-1).
 *
 * `dependency` links form directed chains "from the start of the task to its
 * end": edge A→B means A is the prerequisite and B follows it. Hovering a note
 * that has dependency links highlights the whole chain through it — both the
 * prerequisites and the dependents — bounded by a configurable depth. Cycles
 * are detected separately so the UI can render them red and warn in the note
 * panel.
 *
 * The helpers are renderer-agnostic: the same module feeds the 2D canvas, the
 * 3D engine and the note-details panel.
 */

/** Structural shape shared by SimulationLink and GraphLink endpoints. */
export interface DependencyLinkLike {
  source: unknown;
  target: unknown;
  link_type?: string;
}

function endpointId(value: unknown): string | undefined {
  if (typeof value === "string") return value;
  if (typeof value === "number" && Number.isFinite(value)) return String(value);
  if (value && typeof value === "object" && "id" in value) {
    const id = (value as { id: unknown }).id;
    if (typeof id === "string") return id;
  }
  return undefined;
}

export function isDependencyLink(link: DependencyLinkLike): boolean {
  return link.link_type === "dependency";
}

export interface DependencyChain {
  /** Node id → BFS distance from the hovered node (0 = hovered node). */
  nodeDepth: Map<string, number>;
  /** Directed "source|target" key → distance of the link's far endpoint. */
  linkDepth: Map<string, number>;
  /** Directed keys of links that lie on a dependency cycle. */
  cycleLinks: Set<string>;
  /** True when the highlighted chain touches a dependency cycle. */
  hasCycle: boolean;
}

export const DEPENDENCY_HIGHLIGHT_DEPTH_DEFAULT = 10;

/** Directed link key used across renderers. */
export function dependencyLinkKey(sourceId: string, targetId: string): string {
  return `${sourceId}|${targetId}`;
}

/**
 * Nodes that belong to at least one directed dependency cycle.
 *
 * Iterative Tarjan SCC: every node in a strongly connected component of size
 * > 1 — or with a self-loop — lies on a cycle. Dependency subgraphs are small
 * relative to the whole graph, so a straightforward implementation suffices.
 */
export function computeDependencyCycleNodes(links: DependencyLinkLike[]): Set<string> {
  const adjacency = new Map<string, string[]>();
  const selfLoop = new Set<string>();

  for (const link of links) {
    if (!isDependencyLink(link)) continue;
    const s = endpointId(link.source);
    const t = endpointId(link.target);
    if (!s || !t) continue;
    if (s === t) {
      selfLoop.add(s);
      continue;
    }
    if (!adjacency.has(s)) adjacency.set(s, []);
    adjacency.get(s)!.push(t);
    if (!adjacency.has(t)) adjacency.set(t, []);
  }

  // Iterative Tarjan.
  const index = new Map<string, number>();
  const lowlink = new Map<string, number>();
  const onStack = new Set<string>();
  const stack: string[] = [];
  const cyclic = new Set<string>(selfLoop);
  let nextIndex = 0;

  for (const start of adjacency.keys()) {
    if (index.has(start)) continue;

    const work: Array<{ node: string; childIdx: number }> = [{ node: start, childIdx: 0 }];
    index.set(start, nextIndex);
    lowlink.set(start, nextIndex);
    nextIndex++;
    stack.push(start);
    onStack.add(start);

    while (work.length > 0) {
      const frame = work[work.length - 1];
      const children = adjacency.get(frame.node) ?? [];

      if (frame.childIdx < children.length) {
        const child = children[frame.childIdx++];
        if (!index.has(child)) {
          index.set(child, nextIndex);
          lowlink.set(child, nextIndex);
          nextIndex++;
          stack.push(child);
          onStack.add(child);
          work.push({ node: child, childIdx: 0 });
        } else if (onStack.has(child)) {
          lowlink.set(frame.node, Math.min(lowlink.get(frame.node)!, index.get(child)!));
        }
        continue;
      }

      work.pop();
      if (lowlink.get(frame.node) === index.get(frame.node)) {
        const scc: string[] = [];
        let w: string;
        do {
          w = stack.pop()!;
          onStack.delete(w);
          scc.push(w);
        } while (w !== frame.node);
        if (scc.length > 1) {
          for (const member of scc) cyclic.add(member);
        }
      }
      if (work.length > 0) {
        const parent = work[work.length - 1].node;
        lowlink.set(parent, Math.min(lowlink.get(parent)!, lowlink.get(frame.node)!));
      }
    }
  }

  return cyclic;
}

/**
 * The dependency chain through `hoveredNodeId`: nodes and links reachable by
 * walking dependency edges in BOTH directions (prerequisites and dependents),
 * bounded by `maxDepth` hops. Returns null when the hovered node has no
 * dependency links — in that case regular neighbour highlighting applies.
 */
export function computeDependencyChain(
  hoveredNodeId: string | null | undefined,
  links: DependencyLinkLike[],
  maxDepth: number = DEPENDENCY_HIGHLIGHT_DEPTH_DEFAULT
): DependencyChain | null {
  if (!hoveredNodeId) return null;
  const depthLimit = Math.max(0, maxDepth);

  const out = new Map<string, string[]>();
  const into = new Map<string, string[]>();
  const depEdges: Array<{ s: string; t: string }> = [];

  for (const link of links) {
    if (!isDependencyLink(link)) continue;
    const s = endpointId(link.source);
    const t = endpointId(link.target);
    if (!s || !t) continue;
    depEdges.push({ s, t });
    if (!out.has(s)) out.set(s, []);
    out.get(s)!.push(t);
    if (!into.has(t)) into.set(t, []);
    into.get(t)!.push(s);
  }

  if (!out.has(hoveredNodeId) && !into.has(hoveredNodeId)) return null;

  const nodeDepth = new Map<string, number>([[hoveredNodeId, 0]]);
  let frontier = [hoveredNodeId];

  for (let depth = 1; depth <= depthLimit && frontier.length > 0; depth++) {
    const next: string[] = [];
    for (const nodeId of frontier) {
      for (const succ of out.get(nodeId) ?? []) {
        if (!nodeDepth.has(succ)) {
          nodeDepth.set(succ, depth);
          next.push(succ);
        }
      }
      for (const pred of into.get(nodeId) ?? []) {
        if (!nodeDepth.has(pred)) {
          nodeDepth.set(pred, depth);
          next.push(pred);
        }
      }
    }
    frontier = next;
  }

  const linkDepth = new Map<string, number>();
  for (const { s, t } of depEdges) {
    const ds = nodeDepth.get(s);
    const dt = nodeDepth.get(t);
    if (ds === undefined || dt === undefined) continue;
    linkDepth.set(dependencyLinkKey(s, t), Math.max(ds, dt));
  }

  const cycleNodes = computeDependencyCycleNodes(links);
  const cycleLinks = new Set<string>();
  for (const { s, t } of depEdges) {
    if (nodeDepth.has(s) && nodeDepth.has(t) && cycleNodes.has(s) && cycleNodes.has(t)) {
      cycleLinks.add(dependencyLinkKey(s, t));
    }
  }

  return {
    nodeDepth,
    linkDepth,
    cycleLinks,
    hasCycle: cycleLinks.size > 0,
  };
}

/**
 * Opacity factor for a chain element at the given BFS distance: the hovered
 * node is brightest, brightness decays with distance.
 */
export function chainDepthOpacity(depth: number): number {
  return Math.max(0.35, 1 - depth * 0.12);
}
