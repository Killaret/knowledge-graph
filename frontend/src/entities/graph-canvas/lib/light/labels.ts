/**
 * GRAPH-LIGHT-1: captions of the light style (decision 83).
 *
 * Captions keep one on-screen size at any zoom, sit to the right of their
 * light (to the left when the right side is taken), use a soft halo instead
 * of a hard outline, are not cut to a dozen characters and never overlap:
 * the more important caption is placed first and a later one that would
 * collide is skipped for this frame.
 */
import { CelestialBody } from "$entities";
import type { SimulationNode } from "../types";
import { lightFrame } from "./style";
import { zoomCompensation } from "./glyphs";
import { lightCoreRadius, lightTypeColor, mixRgb, rgba, WHITE } from "./palette";

export interface LightCaption {
  node: SimulationNode;
  opacity: number;
  /** Larger first: hovered, then its neighbourhood, then by cosmic scale. */
  priority: number;
}

const MAX_CAPTION_CHARS = 40;
const FONT_STACK = '"Segoe UI", Inter, system-ui, sans-serif';
const TEXT: readonly [number, number, number] = [228, 234, 247];

export function captionPriority(
  node: SimulationNode,
  hoveredNodeId: string | null,
  hoveredNeighborIds?: Set<string>
): number {
  if (hoveredNodeId && node.id === hoveredNodeId) return 1000;
  if (hoveredNodeId && hoveredNeighborIds?.has(node.id)) return 500;
  return CelestialBody.fromString(node.type).scaleRank;
}

function captionText(title: string | undefined): string {
  const text = (title || "Untitled").trim();
  return text.length > MAX_CAPTION_CHARS ? text.slice(0, MAX_CAPTION_CHARS - 1) + "…" : text;
}

function overlaps(box: [number, number, number, number]): boolean {
  const [x, y, w, h] = box;
  return lightFrame.labelBoxes.some(
    ([bx, by, bw, bh]) => x < bx + bw && x + w > bx && y < by + bh && y + h > by
  );
}

/**
 * Draw captions in world coordinates (the caller has applied the pan/zoom
 * transform). Collision boxes are kept in screen pixels.
 */
export function drawLightCaptions(
  ctx: CanvasRenderingContext2D,
  captions: LightCaption[],
  iconRadius: number
): void {
  const k = lightFrame.k;
  // The transform is the same for every caption of the frame; a context
  // without one (tests) places captions without the overlap check.
  const m = typeof ctx.getTransform === "function" ? ctx.getTransform() : undefined;
  const sorted = [...captions].sort((a, b) => b.priority - a.priority);

  ctx.save();
  ctx.textBaseline = "middle";
  ctx.shadowColor = "rgba(2,3,9,0.95)";
  ctx.shadowBlur = lightFrame.stable ? 0 : 6;

  for (const { node, opacity } of sorted) {
    if (node.x == null || node.y == null || opacity <= 0.02) continue;
    const body = CelestialBody.fromString(node.type);
    const broad = body.type === "galaxy" || body.type === "star";
    const sizePx = body.type === "galaxy" ? 13 : broad ? 12 : 11.5;
    ctx.font = `${broad ? 500 : 400} ${sizePx / k}px ${FONT_STACK}`;
    const text = captionText(node.title);
    const widthWorld = ctx.measureText(text).width;
    const core = lightCoreRadius(body.type, iconRadius * body.baseRadius) * zoomCompensation();
    const gap = core * (body.type === "galaxy" ? 2.4 : 1.6) + 6 / k;

    let x = node.x + gap;
    const y = node.y;
    if (m) {
      const toScreen = (wx: number, wy: number): [number, number] => [
        m.a * wx + m.c * wy + m.e,
        m.b * wx + m.d * wy + m.f,
      ];
      const heightWorld = (sizePx + 3) / k;
      const box = (left: number): [number, number, number, number] => {
        const [sx, sy] = toScreen(left, y - heightWorld / 2);
        return [sx - 2, sy, widthWorld * m.a + 4, heightWorld * m.d];
      };
      let candidate = box(x);
      if (overlaps(candidate)) {
        x = node.x - gap - widthWorld;
        candidate = box(x);
        if (overlaps(candidate)) continue;
      }
      lightFrame.labelBoxes.push(candidate);
    }

    const color = body.type === "galaxy" ? mixRgb(lightTypeColor(body.type), WHITE, 0.55) : TEXT;
    ctx.fillStyle = rgba(color, (broad ? 0.95 : 0.82) * opacity);
    ctx.fillText(text, x, y);
  }

  ctx.restore();
}
