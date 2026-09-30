/**
 * GRAPH-LIGHT-1 (decision 83): which look the 2D graph uses.
 *
 * "classic" keeps the icon renderers; "light" draws notes as sources of
 * light. The value comes from `frontend.graph.style` and can be switched at
 * runtime (tests, a future settings toggle) without reloading the page.
 */
import { graphStyle as configuredStyle } from "$shared/config";

export type GraphStyle = "classic" | "light";

let activeStyle: GraphStyle = configuredStyle;

export function getGraphStyle(): GraphStyle {
  return activeStyle;
}

export function setGraphStyle(style: GraphStyle): void {
  activeStyle = style;
}

export function isLightStyle(): boolean {
  return activeStyle === "light";
}

/**
 * Per-frame values the light renderers need that the classic call signatures
 * do not carry: the current zoom (to keep lines and captions at a constant
 * on-screen size), the animation clock and the deterministic snapshot flag.
 * `labelBoxes` collects caption rectangles in screen pixels so captions do not
 * overlap within one frame.
 */
export interface LightFrame {
  k: number;
  time: number;
  stable: boolean;
  labelBoxes: Array<[number, number, number, number]>;
  /**
   * How far the hover focus has faded in, 0..1. The canvas animates it so the
   * neighbourhood lights up and the rest dims smoothly; callers that do not
   * animate leave it at 1 (instant focus).
   */
  focusMix: number;
  /** Selected note: drawn with a ring and a slow pulse. */
  selectedId: string | null;
  /** Recommendations of the hovered note (decision 81), loaded by the canvas. */
  recommendationsFor: string | null;
  recommendations: ReadonlyArray<{ id: string; score: number }>;
  /**
   * While the graph turns into the list: how much of each note's threads is
   * left, by note id (1 = fully shown). Null outside the morph.
   */
  threadFade: ReadonlyMap<string, number> | null;
  /**
   * Background motion — breathing halos, twinkling stars, the pulse around
   * the selected note. Off under prefers-reduced-motion, in snapshot mode and
   * on large graphs (see lightAmbient).
   */
  ambient: boolean;
}

export const lightFrame: LightFrame = {
  k: 1,
  time: 0,
  stable: false,
  labelBoxes: [],
  focusMix: 1,
  selectedId: null,
  recommendationsFor: null,
  recommendations: [],
  threadFade: null,
  ambient: true,
};

export function setLightThreadFade(fade: ReadonlyMap<string, number> | null): void {
  lightFrame.threadFade = fade;
}

export function setLightSelection(id: string | null): void {
  lightFrame.selectedId = id;
}

export function setLightRecommendations(
  forId: string | null,
  items: ReadonlyArray<{ id: string; score: number }>
): void {
  lightFrame.recommendationsFor = forId;
  lightFrame.recommendations = items;
}

export function setLightAmbient(on: boolean): void {
  lightFrame.ambient = on;
}

/**
 * Whether background motion runs: not when the system asks to reduce motion,
 * not in snapshot mode, and not on graphs larger than the configured limit
 * (`frontend.graph.ambient_max_nodes`), where every frame costs too much.
 */
export function lightAmbient(options: {
  reducedMotion: boolean;
  snapshot: boolean;
  nodeCount: number;
  maxNodes: number;
}): boolean {
  return !options.reducedMotion && !options.snapshot && options.nodeCount <= options.maxNodes;
}

export function setLightFocusMix(mix: number): void {
  lightFrame.focusMix = Math.min(1, Math.max(0, mix));
}

export function beginLightFrame(k: number, time: number, stable: boolean): void {
  lightFrame.k = k > 0 && Number.isFinite(k) ? k : 1;
  lightFrame.time = time;
  lightFrame.stable = stable;
  lightFrame.labelBoxes.length = 0;
}
