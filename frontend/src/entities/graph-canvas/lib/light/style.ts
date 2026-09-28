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
}

export const lightFrame: LightFrame = { k: 1, time: 0, stable: false, labelBoxes: [] };

export function beginLightFrame(k: number, time: number, stable: boolean): void {
  lightFrame.k = k > 0 && Number.isFinite(k) ? k : 1;
  lightFrame.time = time;
  lightFrame.stable = stable;
  lightFrame.labelBoxes.length = 0;
}
