import type { CockpitPanelPosition } from "../model/cockpit.types";

export interface Point {
  x: number;
  y: number;
}

/**
 * Determine whether a pointer is inside a panel's edge trigger zone.
 * Edge size is measured from the corresponding screen border inwards.
 */
export function isInsideEdge(
  position: CockpitPanelPosition,
  pointer: Point,
  edgeSize: number,
  rect: { width: number; height: number }
): boolean {
  switch (position) {
    case "top":
      return pointer.y >= 0 && pointer.y <= edgeSize && pointer.x >= 0 && pointer.x <= rect.width;
    case "bottom":
      return (
        pointer.y >= rect.height - edgeSize &&
        pointer.y <= rect.height &&
        pointer.x >= 0 &&
        pointer.x <= rect.width
      );
    case "left":
      return pointer.x >= 0 && pointer.x <= edgeSize && pointer.y >= 0 && pointer.y <= rect.height;
    case "right":
      return (
        pointer.x >= rect.width - edgeSize &&
        pointer.x <= rect.width &&
        pointer.y >= 0 &&
        pointer.y <= rect.height
      );
    default:
      return false;
  }
}

/**
 * Calculate how far the user has dragged from the screen edge.
 * Positive values mean dragging the panel open, negative mean closing.
 */
export function dragDistance(
  position: CockpitPanelPosition,
  start: Point,
  current: Point,
  rect: { width: number; height: number }
): number {
  switch (position) {
    case "top":
      return current.y - start.y;
    case "bottom":
      return start.y - current.y;
    case "left":
      return current.x - start.x;
    case "right":
      return start.x - current.x;
    default:
      return 0;
  }
}

/**
 * Compute a panel's CSS transform/position values based on its state.
 * Returns the offset in px when closed (negative into the corresponding edge).
 */
export function getPanelOffset(
  position: CockpitPanelPosition,
  size: number,
  open: boolean
): { transform: string } {
  if (open) {
    return { transform: "translate(0, 0)" };
  }

  switch (position) {
    case "top":
      return { transform: `translateY(-${size}px)` };
    case "bottom":
      return { transform: `translateY(${size}px)` };
    case "left":
      return { transform: `translateX(-${size}px)` };
    case "right":
      return { transform: `translateX(${size}px)` };
  }
}
