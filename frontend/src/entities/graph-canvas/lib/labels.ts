/**
 * Selective node labels (UI-GRAPH-1).
 *
 * On a dense graph every node carrying a caption turns into unreadable noise.
 * Labels stay on for nodes that matter right now — hovered, selected, search
 * matches and their neighbors — plus hub nodes (high link degree), and come
 * back for every node once the user zooms in close enough.
 */
import type { SimulationLink, SimulationNode } from "./types";
import { getLinkEndpointId } from "./types";

/** Zoom factor at which every node gets its label back. */
export const FULL_LABEL_ZOOM = 1.5;
/** Link degree that marks a node as a hub — hubs keep their labels. */
export const HUB_MIN_DEGREE = 3;
/** Graphs at or below this size stay fully captioned — density is not a problem yet. */
export const SMALL_GRAPH_ALL_LABELS = 20;

export interface LabelSelection {
  hoveredId?: string | null;
  selectedId?: string | null;
  searchMatchIds?: Iterable<string>;
  /** Current canvas zoom (transform.k). */
  zoomK: number;
  fullLabelZoom?: number;
  hubMinDegree?: number;
}

/**
 * Build the set of node ids whose labels should be drawn this frame.
 * Returns `undefined` when every label should be drawn (zoomed in or
 * deterministic snapshot mode) — callers treat `undefined` as "label all".
 */
export function computeLabeledNodeIds(
  nodes: SimulationNode[],
  links: SimulationLink[],
  sel: LabelSelection
): Set<string> | undefined {
  const fullZoom = sel.fullLabelZoom ?? FULL_LABEL_ZOOM;
  if (sel.zoomK >= fullZoom) return undefined;
  if (nodes.length <= SMALL_GRAPH_ALL_LABELS) return undefined;

  const labeled = new Set<string>();
  const hubDegree = sel.hubMinDegree ?? HUB_MIN_DEGREE;

  const degree = new Map<string, number>();
  for (const link of links) {
    const s = getLinkEndpointId(link.source);
    const t = getLinkEndpointId(link.target);
    degree.set(s, (degree.get(s) ?? 0) + 1);
    degree.set(t, (degree.get(t) ?? 0) + 1);
    const a = sel.hoveredId;
    if (a && (s === a || t === a)) {
      labeled.add(s);
      labeled.add(t);
    }
  }

  for (const node of nodes) {
    if ((degree.get(node.id) ?? 0) >= hubDegree) labeled.add(node.id);
  }

  if (sel.hoveredId) labeled.add(sel.hoveredId);
  if (sel.selectedId) labeled.add(sel.selectedId);
  if (sel.searchMatchIds) {
    for (const id of sel.searchMatchIds) labeled.add(id);
  }

  return labeled;
}
