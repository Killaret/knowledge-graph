/**
 * GRAPH-LIGHT-1: the graph turns into the list and back (decision 83).
 *
 * Notes fly from their places in the graph to the cards of the real list and
 * back. A card outside the visible part of the list takes its note to that
 * edge, where the note fades — the stream pours into the list instead of
 * vanishing off screen. Visible cards move first in both directions, higher
 * rows before lower ones, so the list fills in (and empties) like a wave; notes
 * of cards beyond the edge follow. Notes without a card (filtered out of the
 * list) fade where they are. Threads fade with whichever end has gone further. Positions are
 * interpolated in world space, so the camera transform of the frame converts
 * the screen targets.
 *
 * The fog (decision 85) is not switched off for the flight: its clear circle
 * opens from the middle over the first part of the way and closes over the
 * last part of the way back, so notes it hid come out gradually instead of
 * all at once.
 */
import { easeInOutCubic } from "../camera";
import type { FogRenderParams } from "../fog";
import type { SimulationNode } from "../types";

export type MorphDirection = "to-list" | "to-graph";

/** Duration of the whole morph; the list fades in over its last part. */
export const LIST_MORPH_MS = 1100;

/** A list card: where its note lands, in canvas pixels. Rows come in list order. */
export interface ListRow {
  id: string;
  x: number;
  y: number;
}

/** The visible part of the list, in canvas pixels. */
export interface ListWindow {
  top: number;
  bottom: number;
}

interface MorphTarget {
  x: number;
  y: number;
  /** The card is outside the visible part of the list: the note fades on the edge. */
  beyondEdge: boolean;
}

/** The page's request: list geometry, direction, and on the way back the note to gather around. */
export interface ListMorphRequest {
  rows: readonly ListRow[];
  visible: ListWindow;
  direction: MorphDirection;
  /** On the way back: the graph gathers around this note (a click on its card). */
  focusId?: string | null;
  onDone: () => void;
}

export interface GraphListMorph {
  targets: ReadonlyMap<string, MorphTarget>;
  /** Stagger rank in [0, 1]: visible cards first, each group in list order. */
  rank: ReadonlyMap<string, number>;
  direction: MorphDirection;
  t0: number;
  duration: number;
}

/** Share of the flight that the stagger spreads over. */
const STAGGER = 0.35;

/** How far past the visible edge a note travels before it is gone. */
const EDGE_OVERSHOOT = 36;

/** Share of the way toward the list over which the fog opens. */
const FOG_LIFT_SHARE = 0.35;

export function createMorph(
  rows: readonly ListRow[],
  visible: ListWindow,
  direction: MorphDirection,
  now: number,
  duration = LIST_MORPH_MS
): GraphListMorph {
  const targets = new Map<string, MorphTarget>();
  for (const row of rows) {
    const above = row.y < visible.top;
    const below = row.y > visible.bottom;
    const y = above
      ? visible.top - EDGE_OVERSHOOT
      : below
        ? visible.bottom + EDGE_OVERSHOOT
        : row.y;
    targets.set(row.id, { x: row.x, y, beyondEdge: above || below });
  }
  const order = [
    ...rows.filter((row) => !targets.get(row.id)!.beyondEdge),
    ...rows.filter((row) => targets.get(row.id)!.beyondEdge),
  ];
  const rank = new Map<string, number>();
  const last = Math.max(1, order.length - 1);
  order.forEach((row, i) => rank.set(row.id, i / last));
  return { targets, rank, direction, t0: now, duration };
}

/** How far the whole picture is toward the list: 0 = graph, 1 = list. */
export function morphTowardList(morph: GraphListMorph, now: number): number {
  const p = Math.min(1, Math.max(0, (now - morph.t0) / morph.duration));
  return morph.direction === "to-list" ? p : 1 - p;
}

export function morphFinished(morph: GraphListMorph, now: number): boolean {
  return now - morph.t0 >= morph.duration;
}

/**
 * Eased progress of one note toward its row. The stagger delays notes by rank
 * from the start of the run, whichever way it goes.
 */
export function noteTowardList(morph: GraphListMorph, id: string, towardList: number): number {
  const rank = morph.rank.get(id);
  if (rank === undefined) return towardList;
  const run = morph.direction === "to-list" ? towardList : 1 - towardList;
  const moved = easeInOutCubic(Math.min(1, Math.max(0, (run - STAGGER * rank) / (1 - STAGGER))));
  return morph.direction === "to-list" ? moved : 1 - moved;
}

/** Opacity factor of a note: visible cards keep it, the edge and no card fade it. */
export function morphNoteOpacity(morph: GraphListMorph, id: string, towardList: number): number {
  const target = morph.targets.get(id);
  if (!target) return 1 - towardList;
  return target.beyondEdge ? 1 - noteTowardList(morph, id, towardList) : 1;
}

/** How much of each note's threads is left: a thread fades as either end leaves. */
export function morphThreadFade(
  morph: GraphListMorph,
  nodes: readonly SimulationNode[],
  towardList: number
): Map<string, number> {
  const fade = new Map<string, number>();
  for (const node of nodes) {
    const gone = morph.targets.has(node.id)
      ? noteTowardList(morph, node.id, towardList)
      : towardList;
    fade.set(node.id, 1 - gone);
  }
  return fade;
}

/** How far the fog has lifted: 0 = as it is on the graph, 1 = gone. */
export function morphFogLift(towardList: number): number {
  return easeInOutCubic(Math.min(1, Math.max(0, towardList / FOG_LIFT_SHARE)));
}

/**
 * The fog with its clear circle opened by `lift`, up to past the farthest
 * corner of the canvas. A fog that hides nothing (off, focus mode, or already
 * wider than the canvas) is returned as it is.
 */
export function liftFog<T extends FogRenderParams>(
  fog: T,
  lift: number,
  width: number,
  height: number
): T {
  if (!fog.enabled || fog.mode === "off" || fog.mode === "first-person" || lift <= 0) return fog;
  const clear = Math.hypot(width, height) + fog.feather;
  if (fog.radius >= clear) return fog;
  return { ...fog, radius: fog.radius + (clear - fog.radius) * Math.min(1, lift) };
}

/**
 * Copies of the nodes placed between the graph and their rows. The simulation
 * itself is not touched.
 */
export function morphNodes(
  nodes: readonly SimulationNode[],
  morph: GraphListMorph,
  towardList: number,
  transform: { x: number; y: number; k: number }
): SimulationNode[] {
  return nodes.map((node) => {
    const target = morph.targets.get(node.id);
    if (!target || node.x == null || node.y == null) return node;
    const p = noteTowardList(morph, node.id, towardList);
    const tx = (target.x - transform.x) / transform.k;
    const ty = (target.y - transform.y) / transform.k;
    return { ...node, x: node.x + (tx - node.x) * p, y: node.y + (ty - node.y) * p };
  });
}
