/**
 * Selective node labels (UI-GRAPH-1).
 *
 * On a dense graph every node carrying a caption turns into unreadable noise.
 * Labels stay on for nodes that matter right now — hovered, selected, search
 * matches and their neighbors — plus the top-N hubs by link degree, and come
 * back for every node once the user zooms in close enough.
 *
 * The hub count is relative to the graph, not an absolute degree threshold:
 * an absolute "degree >= 3" labels almost every node on a graph whose average
 * degree is ~5 (owner complaint). N comes from `frontend.graph.label_hub_count`
 * in knowledge-graph.config.json.
 */
import { graphLabelHubCount } from "$shared/config";

/** Zoom factor at which every node gets its label back. */
export const FULL_LABEL_ZOOM = 1.5;
/** Graphs at or below this size stay fully captioned — density is not a problem yet. */
export const SMALL_GRAPH_ALL_LABELS = 20;

export interface LabelSelection {
  hoveredId?: string | null;
  selectedId?: string | null;
  searchMatchIds?: Iterable<string>;
  /** Current canvas zoom (transform.k). */
  zoomK: number;
  fullLabelZoom?: number;
  /** Override for the configured hub-label limit (tests). */
  maxHubLabels?: number;
}

interface LabelNode {
  id: string;
}

interface LabelLink {
  source: unknown;
  target: unknown;
}

function endpointId(ref: unknown): string {
  if (typeof ref === "string") return ref;
  if (typeof ref === "number") return String(ref);
  return (ref as { id: string }).id;
}

/**
 * Build the set of node ids whose labels should be drawn this frame.
 * Returns `undefined` when every label should be drawn (zoomed in, small
 * graph) — callers treat `undefined` as "label all".
 */
export function computeLabeledNodeIds<N extends LabelNode>(
  nodes: N[],
  links: LabelLink[],
  sel: LabelSelection
): Set<string> | undefined {
  const fullZoom = sel.fullLabelZoom ?? FULL_LABEL_ZOOM;
  if (sel.zoomK >= fullZoom) return undefined;
  if (nodes.length <= SMALL_GRAPH_ALL_LABELS) return undefined;

  const labeled = new Set<string>();
  const maxHubs = sel.maxHubLabels ?? graphLabelHubCount;

  const degree = new Map<string, number>();
  for (const link of links) {
    const s = endpointId(link.source);
    const t = endpointId(link.target);
    degree.set(s, (degree.get(s) ?? 0) + 1);
    degree.set(t, (degree.get(t) ?? 0) + 1);
    const a = sel.hoveredId;
    if (a && (s === a || t === a)) {
      labeled.add(s);
      labeled.add(t);
    }
  }

  // Top-N hubs by degree, deterministic tie-break by id (stable snapshots
  // depend on the same nodes winning every run).
  const hubs = nodes
    .map((n) => ({ id: n.id, d: degree.get(n.id) ?? 0 }))
    .filter((e) => e.d > 0)
    .sort((a, b) => b.d - a.d || (a.id < b.id ? -1 : a.id > b.id ? 1 : 0))
    .slice(0, maxHubs);
  for (const e of hubs) labeled.add(e.id);

  if (sel.hoveredId) labeled.add(sel.hoveredId);
  if (sel.selectedId) labeled.add(sel.selectedId);
  if (sel.searchMatchIds) {
    for (const id of sel.searchMatchIds) labeled.add(id);
  }

  return labeled;
}
